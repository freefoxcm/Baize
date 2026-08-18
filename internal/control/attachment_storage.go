package control

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"reasonix/internal/filelock"
)

const maxFileAttachmentBytes = 100 * 1024 * 1024
const defaultAttachmentQuotaBytes = 1024 * 1024 * 1024

var (
	ErrAttachmentTooLarge = errors.New("attachment exceeds per-file limit")
	ErrAttachmentQuota    = errors.New("workspace attachment quota exceeded")
	ErrAttachmentInvalid  = errors.New("invalid attachment")
)

type AttachmentPolicy struct {
	MaxFileBytes        int64
	WorkspaceQuotaBytes int64
}

func DefaultAttachmentPolicy() AttachmentPolicy {
	return AttachmentPolicy{MaxFileBytes: maxFileAttachmentBytes, WorkspaceQuotaBytes: defaultAttachmentQuotaBytes}
}

func (p AttachmentPolicy) normalized() AttachmentPolicy {
	if p.MaxFileBytes <= 0 {
		p.MaxFileBytes = maxFileAttachmentBytes
	}
	if p.WorkspaceQuotaBytes <= 0 {
		p.WorkspaceQuotaBytes = defaultAttachmentQuotaBytes
	}
	return p
}

type AttachmentInfo struct {
	Path       string    `json:"path"`
	Name       string    `json:"name"`
	Size       int64     `json:"size"`
	SHA256     string    `json:"sha256,omitempty"`
	ModifiedAt time.Time `json:"modified_at,omitempty"`
}

type AttachmentUsage struct {
	Count          int   `json:"count"`
	TotalBytes     int64 `json:"total_bytes"`
	MaxFileBytes   int64 `json:"max_file_bytes"`
	QuotaBytes     int64 `json:"quota_bytes"`
	RemainingBytes int64 `json:"remaining_bytes"`
}

type attachmentContextReader struct {
	ctx context.Context
	src io.Reader
}

func (r attachmentContextReader) Read(p []byte) (int, error) {
	if err := r.ctx.Err(); err != nil {
		return 0, err
	}
	return r.src.Read(p)
}

// SaveAttachmentReaderInRoot streams one attachment to a same-directory
// temporary file and atomically publishes it after hash and quota checks.
func SaveAttachmentReaderInRoot(ctx context.Context, root, origName string, src io.Reader, knownSize int64, policy AttachmentPolicy) (info AttachmentInfo, err error) {
	if src == nil {
		return info, fmt.Errorf("%w: missing file data", ErrAttachmentInvalid)
	}
	name, err := validateAttachmentName(origName)
	if err != nil {
		return info, err
	}
	policy = policy.normalized()
	if policy.WorkspaceQuotaBytes < policy.MaxFileBytes {
		return info, fmt.Errorf("%w: workspace quota is smaller than the per-file limit", ErrAttachmentInvalid)
	}
	if knownSize == 0 {
		return info, fmt.Errorf("%w: empty files are not allowed", ErrAttachmentInvalid)
	}
	if knownSize > policy.MaxFileBytes {
		return info, fmt.Errorf("%w: maximum is %d bytes", ErrAttachmentTooLarge, policy.MaxFileBytes)
	}
	if strings.TrimSpace(root) == "" {
		root = "."
	}
	absRoot, err := filepath.Abs(root)
	if err != nil {
		return info, err
	}
	if err := ensureAttachmentRootIn(absRoot); err != nil {
		return info, err
	}
	dir := filepath.Join(absRoot, ".reasonix", "attachments")
	release, err := filelock.Acquire(ctx, filepath.Join(dir, ".quota.lock"))
	if err != nil {
		return info, err
	}
	defer release()
	used, _, err := attachmentUsageLocked(dir)
	if err != nil {
		return info, err
	}
	if knownSize > 0 && used+knownSize > policy.WorkspaceQuotaBytes {
		return info, fmt.Errorf("%w: %d bytes remaining", ErrAttachmentQuota, max(policy.WorkspaceQuotaBytes-used, 0))
	}
	tmp, err := os.CreateTemp(dir, ".upload-*")
	if err != nil {
		return info, err
	}
	tmpPath := tmp.Name()
	defer func() {
		_ = tmp.Close()
		if err != nil {
			_ = os.Remove(tmpPath)
		}
	}()
	if err = tmp.Chmod(0o600); err != nil {
		return info, err
	}
	h := sha256.New()
	limited := io.LimitReader(attachmentContextReader{ctx: ctx, src: src}, policy.MaxFileBytes+1)
	written, copyErr := io.CopyBuffer(io.MultiWriter(tmp, h), limited, make([]byte, 128*1024))
	if copyErr != nil {
		return info, copyErr
	}
	if written == 0 {
		return info, fmt.Errorf("%w: empty files are not allowed", ErrAttachmentInvalid)
	}
	if written > policy.MaxFileBytes {
		return info, fmt.Errorf("%w: maximum is %d bytes", ErrAttachmentTooLarge, policy.MaxFileBytes)
	}
	if used+written > policy.WorkspaceQuotaBytes {
		return info, fmt.Errorf("%w: %d bytes remaining", ErrAttachmentQuota, max(policy.WorkspaceQuotaBytes-used, 0))
	}
	if err = tmp.Sync(); err != nil {
		return info, err
	}
	if err = tmp.Close(); err != nil {
		return info, err
	}
	finalPath, rel, err := uniqueAttachmentDestination(absRoot, name)
	if err != nil {
		return info, err
	}
	if err = os.Rename(tmpPath, finalPath); err != nil {
		return info, err
	}
	info = AttachmentInfo{Path: filepath.ToSlash(rel), Name: name, Size: written, SHA256: hex.EncodeToString(h.Sum(nil))}
	return info, nil
}

func validateAttachmentName(origName string) (string, error) {
	name := strings.TrimSpace(origName)
	if name == "" || name == "." || name == ".." || strings.ContainsAny(name, `/\`) {
		return "", fmt.Errorf("%w: unsafe filename", ErrAttachmentInvalid)
	}
	name = strings.Map(func(r rune) rune {
		if r < 32 || strings.ContainsRune(`<>:"|?*`, r) {
			return '-'
		}
		return r
	}, name)
	runes := []rune(strings.Trim(name, " ."))
	if len(runes) > 160 {
		runes = runes[:160]
	}
	name = string(runes)
	if name == "" {
		return "", fmt.Errorf("%w: unsafe filename", ErrAttachmentInvalid)
	}
	return name, nil
}

func uniqueAttachmentDestination(absRoot, name string) (string, string, error) {
	storedName := storedAttachmentName(name)
	for range maxAttachmentCreateAttempts {
		seq := attachmentPathSeq.Add(1)
		stored := fmt.Sprintf("upload-%s-%06d-%s", attachmentNow().Format("20060102-150405.000000"), seq, storedName)
		rel := filepath.Join(".reasonix", "attachments", stored)
		final := filepath.Join(absRoot, rel)
		if _, err := os.Lstat(final); os.IsNotExist(err) {
			return final, rel, nil
		} else if err != nil {
			return "", "", err
		}
	}
	return "", "", fmt.Errorf("create unique attachment path")
}

func storedAttachmentName(name string) string {
	ext := strings.ToLower(filepath.Ext(name))
	base := strings.TrimSuffix(name, filepath.Ext(name))
	if !safeAttachmentExt.MatchString(ext) {
		ext = ".bin"
	}
	base = strings.Map(func(r rune) rune {
		if r < 32 || r == ' ' || strings.ContainsRune(`<>:"/\|?*`, r) {
			return '-'
		}
		return r
	}, strings.Trim(base, " .-"))
	if base == "" {
		base = "attachment"
	}
	return base + ext
}

func attachmentUsageLocked(dir string) (int64, int, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return 0, 0, err
	}
	var total int64
	count := 0
	for _, entry := range entries {
		if strings.HasPrefix(entry.Name(), ".") {
			continue
		}
		fi, err := entry.Info()
		if err != nil {
			return 0, 0, err
		}
		if fi.Mode().IsRegular() && fi.Mode()&os.ModeSymlink == 0 {
			total += fi.Size()
			count++
		}
	}
	return total, count, nil
}

func (c *Controller) SaveAttachment(ctx context.Context, name string, src io.Reader, knownSize int64) (AttachmentInfo, error) {
	root := strings.TrimSpace(c.WorkspaceRoot())
	if root == "" {
		return AttachmentInfo{}, fmt.Errorf("%w: no workspace", ErrAttachmentInvalid)
	}
	return SaveAttachmentReaderInRoot(ctx, root, name, src, knownSize, c.attachmentPolicy)
}

func (c *Controller) ListAttachments(ctx context.Context) ([]AttachmentInfo, AttachmentUsage, error) {
	root := strings.TrimSpace(c.WorkspaceRoot())
	if root == "" {
		return nil, AttachmentUsage{}, fmt.Errorf("%w: no workspace", ErrAttachmentInvalid)
	}
	absRoot, err := filepath.Abs(root)
	if err != nil {
		return nil, AttachmentUsage{}, err
	}
	if err := ensureAttachmentRootIn(absRoot); err != nil {
		return nil, AttachmentUsage{}, err
	}
	dir := filepath.Join(absRoot, ".reasonix", "attachments")
	release, err := filelock.Acquire(ctx, filepath.Join(dir, ".quota.lock"))
	if err != nil {
		return nil, AttachmentUsage{}, err
	}
	defer release()
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, AttachmentUsage{}, err
	}
	items := make([]AttachmentInfo, 0, len(entries))
	var total int64
	for _, entry := range entries {
		if strings.HasPrefix(entry.Name(), ".") {
			continue
		}
		fi, err := entry.Info()
		if err != nil {
			return nil, AttachmentUsage{}, err
		}
		if !fi.Mode().IsRegular() || fi.Mode()&os.ModeSymlink != 0 {
			continue
		}
		total += fi.Size()
		items = append(items, AttachmentInfo{Path: filepath.ToSlash(filepath.Join(".reasonix", "attachments", entry.Name())), Name: displayAttachmentName(entry.Name()), Size: fi.Size(), ModifiedAt: fi.ModTime()})
	}
	sort.Slice(items, func(i, j int) bool { return items[i].ModifiedAt.After(items[j].ModifiedAt) })
	policy := c.attachmentPolicy.normalized()
	usage := AttachmentUsage{Count: len(items), TotalBytes: total, MaxFileBytes: policy.MaxFileBytes, QuotaBytes: policy.WorkspaceQuotaBytes, RemainingBytes: max(policy.WorkspaceQuotaBytes-total, 0)}
	return items, usage, nil
}

func displayAttachmentName(stored string) string {
	parts := strings.SplitN(stored, "-", 5)
	if len(parts) == 5 && parts[0] == "upload" {
		return parts[4]
	}
	return stored
}

func (c *Controller) DeleteAttachment(ctx context.Context, rel string) error {
	root := strings.TrimSpace(c.WorkspaceRoot())
	if root == "" {
		return fmt.Errorf("%w: no workspace", ErrAttachmentInvalid)
	}
	absRoot, err := filepath.Abs(root)
	if err != nil {
		return err
	}
	if err := ensureAttachmentRootIn(absRoot); err != nil {
		return err
	}
	dir := filepath.Join(absRoot, ".reasonix", "attachments")
	target, err := confinedAttachmentPath(absRoot, rel)
	if err != nil {
		return err
	}
	release, err := filelock.Acquire(ctx, filepath.Join(dir, ".quota.lock"))
	if err != nil {
		return err
	}
	defer release()
	fi, err := os.Lstat(target)
	if err != nil {
		return err
	}
	if fi.Mode()&os.ModeSymlink != 0 || !fi.Mode().IsRegular() {
		return fmt.Errorf("%w: attachment is not a regular file", ErrAttachmentInvalid)
	}
	if err := os.Remove(target); err != nil {
		return err
	}
	c.rotateSessionTemp()
	return nil
}

func (c *Controller) ClearAttachments(ctx context.Context) (int, error) {
	root := strings.TrimSpace(c.WorkspaceRoot())
	if root == "" {
		return 0, fmt.Errorf("%w: no workspace", ErrAttachmentInvalid)
	}
	absRoot, err := filepath.Abs(root)
	if err != nil {
		return 0, err
	}
	if err := ensureAttachmentRootIn(absRoot); err != nil {
		return 0, err
	}
	dir := filepath.Join(absRoot, ".reasonix", "attachments")
	release, err := filelock.Acquire(ctx, filepath.Join(dir, ".quota.lock"))
	if err != nil {
		return 0, err
	}
	defer release()
	entries, err := os.ReadDir(dir)
	if err != nil {
		return 0, err
	}
	removed := 0
	for _, entry := range entries {
		if entry.Name() == ".quota.lock" {
			continue
		}
		path := filepath.Join(dir, entry.Name())
		fi, statErr := os.Lstat(path)
		if statErr != nil {
			return removed, statErr
		}
		if fi.IsDir() {
			continue
		}
		if err := os.Remove(path); err != nil {
			return removed, err
		}
		if !strings.HasPrefix(entry.Name(), ".") {
			removed++
		}
	}
	if removed > 0 {
		c.rotateSessionTemp()
	}
	return removed, nil
}

func confinedAttachmentPath(absRoot, rel string) (string, error) {
	if filepath.IsAbs(rel) {
		return "", fmt.Errorf("%w: absolute paths are not allowed", ErrAttachmentInvalid)
	}
	clean := filepath.Clean(filepath.FromSlash(strings.TrimSpace(rel)))
	expected := filepath.Join(".reasonix", "attachments")
	parent, name := filepath.Split(clean)
	if filepath.Clean(parent) != expected || name == "" || name == ".quota.lock" || strings.HasPrefix(name, ".upload-") {
		return "", fmt.Errorf("%w: path is outside .reasonix/attachments", ErrAttachmentInvalid)
	}
	return filepath.Join(absRoot, clean), nil
}

func SaveAttachmentFileInRoot(root, path string, policy AttachmentPolicy) (string, error) {
	info, err := os.Lstat(path)
	if err != nil {
		return "", err
	}
	if info.Mode()&os.ModeSymlink != 0 {
		return "", fmt.Errorf("attachment path must not be a symlink")
	}
	policy = policy.normalized()
	if info.IsDir() || !info.Mode().IsRegular() || info.Size() <= 0 {
		return "", fmt.Errorf("%w: attachment must be a non-empty regular file", ErrAttachmentInvalid)
	}
	if info.Size() > policy.MaxFileBytes {
		return "", fmt.Errorf("%w: maximum is %d bytes", ErrAttachmentTooLarge, policy.MaxFileBytes)
	}
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	opened, err := f.Stat()
	if err != nil {
		return "", err
	}
	if !os.SameFile(info, opened) {
		return "", fmt.Errorf("attachment changed while opening")
	}
	saved, err := SaveAttachmentReaderInRoot(context.Background(), root, filepath.Base(path), f, opened.Size(), policy)
	if err != nil {
		return "", err
	}
	if after, err := f.Stat(); err != nil {
		_ = os.Remove(filepath.Join(root, filepath.FromSlash(saved.Path)))
		return "", err
	} else if !os.SameFile(opened, after) || after.Size() != opened.Size() {
		_ = os.Remove(filepath.Join(root, filepath.FromSlash(saved.Path)))
		return "", fmt.Errorf("attachment changed while reading")
	}
	return saved.Path, nil
}

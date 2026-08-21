package control

import (
	"bytes"
	"context"
	"encoding/base64"
	"errors"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"

	"reasonix/internal/sessiontemp"
)

type zeroAttachmentReader struct{}

func (zeroAttachmentReader) Read(p []byte) (int, error) {
	clear(p)
	return len(p), nil
}

type cancelingAttachmentReader struct {
	cancel context.CancelFunc
	read   bool
}

func (r *cancelingAttachmentReader) Read(p []byte) (int, error) {
	if r.read {
		return 0, io.EOF
	}
	r.read = true
	n := copy(p, "partial")
	r.cancel()
	return n, nil
}

const tinyPNG = "iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mNk+M9QDwADhgGAWjR9awAAAABJRU5ErkJggg=="

func TestSaveImageDataURL(t *testing.T) {
	t.Chdir(t.TempDir())
	got, err := SaveImageDataURL("data:image/png;base64," + tinyPNG)
	if err != nil {
		t.Fatalf("SaveImageDataURL: %v", err)
	}
	if !strings.HasPrefix(got, ".reasonix/attachments/clipboard-") || !strings.HasSuffix(got, ".png") {
		t.Fatalf("path = %q, want attachment png path", got)
	}
}

func TestSaveImageDataURLRejectsSpoofedMime(t *testing.T) {
	t.Chdir(t.TempDir())
	if _, err := SaveImageDataURL("data:image/png;base64,aGk="); err == nil {
		t.Fatal("spoofed image mime should fail")
	}
}

func TestCreateAttachmentFileSkipsExistingPath(t *testing.T) {
	t.Chdir(t.TempDir())
	if err := ensureAttachmentRoot(); err != nil {
		t.Fatal(err)
	}

	first := attachmentPath(".png")
	if err := os.WriteFile(first, []byte("keep"), 0o644); err != nil {
		t.Fatal(err)
	}

	rel, f, err := createAttachmentFile(".png")
	if err != nil {
		t.Fatalf("createAttachmentFile: %v", err)
	}
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}
	if rel == first {
		t.Fatalf("createAttachmentFile reused existing path %q", rel)
	}
	if got, err := os.ReadFile(first); err != nil {
		t.Fatal(err)
	} else if string(got) != "keep" {
		t.Fatalf("existing attachment was overwritten: %q", got)
	}
}

func TestSaveImageBytesUsesUniquePathsWithinSameTimestamp(t *testing.T) {
	t.Chdir(t.TempDir())
	oldNow := attachmentNow
	attachmentNow = func() time.Time {
		return time.Date(2026, 6, 1, 10, 20, 30, 123456000, time.UTC)
	}
	defer func() {
		attachmentNow = oldNow
	}()

	raw := mustBase64(t, tinyPNG)
	first, err := SaveImageBytes("image/png", raw)
	if err != nil {
		t.Fatalf("first SaveImageBytes: %v", err)
	}
	second, err := SaveImageBytes("image/png", raw)
	if err != nil {
		t.Fatalf("second SaveImageBytes: %v", err)
	}
	if first == second {
		t.Fatalf("paths collided: %q", first)
	}
	for _, path := range []string{first, second} {
		if got, err := os.ReadFile(path); err != nil {
			t.Fatalf("read %s: %v", path, err)
		} else if string(got) != string(raw) {
			t.Fatalf("content for %s changed", path)
		}
	}
}

func TestSaveImageFile(t *testing.T) {
	t.Chdir(t.TempDir())
	if err := os.WriteFile("source.png", mustBase64(t, tinyPNG), 0o644); err != nil {
		t.Fatal(err)
	}
	got, err := SaveImageFile("source.png")
	if err != nil {
		t.Fatalf("SaveImageFile: %v", err)
	}
	if !strings.HasPrefix(got, ".reasonix/attachments/clipboard-") || !strings.HasSuffix(got, ".png") {
		t.Fatalf("path = %q, want attachment png path", got)
	}
}

func TestSaveAttachmentFile(t *testing.T) {
	t.Chdir(t.TempDir())
	if err := os.WriteFile("notes.pdf", []byte("%PDF-1.4 body"), 0o644); err != nil {
		t.Fatal(err)
	}
	got, err := SaveAttachmentFile("notes.pdf")
	if err != nil {
		t.Fatalf("SaveAttachmentFile: %v", err)
	}
	if !strings.HasPrefix(got, ".reasonix/attachments/upload-") || !strings.HasSuffix(got, ".pdf") {
		t.Fatalf("path = %q, want attachment pdf path", got)
	}
	if data, err := os.ReadFile(got); err != nil || string(data) != "%PDF-1.4 body" {
		t.Fatalf("stored bytes = %q (err %v), want original", data, err)
	}
}

func TestSaveAttachmentFileRejectsEmptyAndDir(t *testing.T) {
	t.Chdir(t.TempDir())
	if err := os.WriteFile("empty.txt", nil, 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := SaveAttachmentFile("empty.txt"); err == nil {
		t.Fatal("empty file should fail")
	}
	if err := os.Mkdir("adir", 0o755); err != nil {
		t.Fatal(err)
	}
	if _, err := SaveAttachmentFile("adir"); err == nil {
		t.Fatal("directory should fail")
	}
}

func TestSaveAttachmentFileSanitizesExtension(t *testing.T) {
	t.Chdir(t.TempDir())
	if err := os.WriteFile("payload.weird-ext-here", []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	got, err := SaveAttachmentFile("payload.weird-ext-here")
	if err != nil {
		t.Fatalf("SaveAttachmentFile: %v", err)
	}
	if !strings.HasSuffix(got, ".bin") {
		t.Fatalf("path = %q, want .bin fallback for unsafe extension", got)
	}
}

func TestSaveAttachmentFileRejectsSymlink(t *testing.T) {
	t.Chdir(t.TempDir())
	if err := os.WriteFile("source.bin", []byte("payload"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink("source.bin", "link.bin"); err != nil {
		t.Skipf("symlink unsupported: %v", err)
	}
	if _, err := SaveAttachmentFile("link.bin"); err == nil {
		t.Fatal("symlink attachment path should fail")
	}
}

func TestSaveAttachmentReaderStreamsHashesAndUsesPrivatePermissions(t *testing.T) {
	root := t.TempDir()
	raw := []byte("month,amount\n2026-01,10\n")
	info, err := SaveAttachmentReaderInRoot(context.Background(), root, "sales.csv", bytes.NewReader(raw), int64(len(raw)), AttachmentPolicy{MaxFileBytes: 1024, WorkspaceQuotaBytes: 2048})
	if err != nil {
		t.Fatal(err)
	}
	if info.Name != "sales.csv" || info.Size != int64(len(raw)) || len(info.SHA256) != 64 {
		t.Fatalf("unexpected attachment info: %+v", info)
	}
	fi, err := os.Stat(filepath.Join(root, filepath.FromSlash(info.Path)))
	if err != nil {
		t.Fatal(err)
	}
	if runtime.GOOS != "windows" && fi.Mode().Perm() != 0o600 {
		t.Fatalf("attachment mode = %o, want 600", fi.Mode().Perm())
	}
}

func TestAttachmentDefaultHundredMiBBoundary(t *testing.T) {
	if os.Getenv("REASONIX_LARGE_ATTACHMENT_TESTS") != "1" {
		t.Skip("set REASONIX_LARGE_ATTACHMENT_TESTS=1 for the 100 MiB acceptance boundary")
	}
	policy := DefaultAttachmentPolicy()
	exactRoot := t.TempDir()
	info, err := SaveAttachmentReaderInRoot(context.Background(), exactRoot, "exact.csv", io.LimitReader(zeroAttachmentReader{}, policy.MaxFileBytes), policy.MaxFileBytes, policy)
	if err != nil {
		t.Fatalf("100 MiB upload: %v", err)
	}
	if info.Size != policy.MaxFileBytes {
		t.Fatalf("exact upload size=%d, want %d", info.Size, policy.MaxFileBytes)
	}
	overRoot := t.TempDir()
	_, err = SaveAttachmentReaderInRoot(context.Background(), overRoot, "over.csv", io.LimitReader(zeroAttachmentReader{}, policy.MaxFileBytes+1), -1, policy)
	if !errors.Is(err, ErrAttachmentTooLarge) {
		t.Fatalf("100 MiB + 1 upload error=%v, want ErrAttachmentTooLarge", err)
	}
	entries, err := os.ReadDir(filepath.Join(overRoot, ".reasonix", "attachments"))
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range entries {
		if !strings.HasPrefix(entry.Name(), ".") {
			t.Fatalf("over-limit upload left attachment %q", entry.Name())
		}
	}
}

func TestSaveAttachmentReaderCleansTempFilesOnLimitAndQuotaFailures(t *testing.T) {
	root := t.TempDir()
	policy := AttachmentPolicy{MaxFileBytes: 4, WorkspaceQuotaBytes: 5}
	if _, err := SaveAttachmentReaderInRoot(context.Background(), root, "large.csv", strings.NewReader("12345"), -1, policy); !errors.Is(err, ErrAttachmentTooLarge) {
		t.Fatalf("large upload error = %v", err)
	}
	if _, err := SaveAttachmentReaderInRoot(context.Background(), root, "first.csv", strings.NewReader("1234"), 4, policy); err != nil {
		t.Fatal(err)
	}
	if _, err := SaveAttachmentReaderInRoot(context.Background(), root, "quota.csv", strings.NewReader("12"), 2, policy); !errors.Is(err, ErrAttachmentQuota) {
		t.Fatalf("quota upload error = %v", err)
	}
	entries, err := os.ReadDir(filepath.Join(root, ".reasonix", "attachments"))
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range entries {
		if strings.HasPrefix(entry.Name(), ".upload-") {
			t.Fatalf("failed upload left temporary file %q", entry.Name())
		}
	}
}

func TestSaveAttachmentReaderCleansTempFileWhenCanceled(t *testing.T) {
	root := t.TempDir()
	ctx, cancel := context.WithCancel(context.Background())
	_, err := SaveAttachmentReaderInRoot(ctx, root, "canceled.csv", &cancelingAttachmentReader{cancel: cancel}, -1, DefaultAttachmentPolicy())
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("canceled upload error = %v, want context.Canceled", err)
	}
	entries, err := os.ReadDir(filepath.Join(root, ".reasonix", "attachments"))
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range entries {
		if strings.HasPrefix(entry.Name(), ".upload-") {
			t.Fatalf("canceled upload left temporary file %q", entry.Name())
		}
	}
}

func TestSaveAttachmentReaderRejectsSymlinkedReasonixDirectory(t *testing.T) {
	root := t.TempDir()
	outside := t.TempDir()
	if err := os.Symlink(outside, filepath.Join(root, ".reasonix")); err != nil {
		t.Skipf("symlinks unavailable: %v", err)
	}
	_, err := SaveAttachmentReaderInRoot(context.Background(), root, "sales.csv", strings.NewReader("a\n1\n"), 4, DefaultAttachmentPolicy())
	if err == nil {
		t.Fatal("symlinked .reasonix directory should be rejected")
	}
}

func TestConcurrentAttachmentUploadsCannotExceedQuota(t *testing.T) {
	root := t.TempDir()
	policy := AttachmentPolicy{MaxFileBytes: 8, WorkspaceQuotaBytes: 10}
	var wg sync.WaitGroup
	results := make(chan error, 2)
	for _, name := range []string{"a.csv", "b.csv"} {
		wg.Go(func() {
			_, err := SaveAttachmentReaderInRoot(context.Background(), root, name, strings.NewReader("123456"), 6, policy)
			results <- err
		})
	}
	wg.Wait()
	close(results)
	success, quota := 0, 0
	for err := range results {
		if err == nil {
			success++
		} else if errors.Is(err, ErrAttachmentQuota) {
			quota++
		} else {
			t.Fatalf("unexpected error: %v", err)
		}
	}
	if success != 1 || quota != 1 {
		t.Fatalf("success=%d quota=%d, want 1/1", success, quota)
	}
}

func TestAttachmentManagementDeletesAndRotatesSessionTemp(t *testing.T) {
	root := t.TempDir()
	tempManager := sessiontemp.NewWithRoot(t.TempDir())
	ctrl := New(Options{WorkspaceRoot: root, SessionTemp: tempManager, AttachmentPolicy: AttachmentPolicy{MaxFileBytes: 1024, WorkspaceQuotaBytes: 4096}})
	defer ctrl.Close()
	lease, err := tempManager.Acquire()
	if err != nil {
		t.Fatal(err)
	}
	oldTemp := lease.Dir()
	lease.Release()
	info, err := ctrl.SaveAttachment(context.Background(), "sales.csv", strings.NewReader("a\n1\n"), 4)
	if err != nil {
		t.Fatal(err)
	}
	items, usage, err := ctrl.ListAttachments(context.Background())
	if err != nil || len(items) != 1 || usage.Count != 1 {
		t.Fatalf("list=%+v usage=%+v err=%v", items, usage, err)
	}
	if err := ctrl.DeleteAttachment(context.Background(), info.Path); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(oldTemp); !os.IsNotExist(err) {
		t.Fatalf("old session temp still exists after attachment delete: %v", err)
	}
	_, usage, err = ctrl.ListAttachments(context.Background())
	if err != nil || usage.Count != 0 || usage.TotalBytes != 0 {
		t.Fatalf("usage after delete = %+v, err=%v", usage, err)
	}
}

func TestSaveImageFileRejectsSymlink(t *testing.T) {
	t.Chdir(t.TempDir())
	if err := os.WriteFile("source.png", mustBase64(t, tinyPNG), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink("source.png", "link.png"); err != nil {
		t.Skipf("symlink unsupported: %v", err)
	}
	if _, err := SaveImageFile("link.png"); err == nil {
		t.Fatal("symlink image path should fail")
	}
}

func TestImageDataURLRejectsOutsideAttachmentDir(t *testing.T) {
	t.Chdir(t.TempDir())
	if err := os.WriteFile("x.png", []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := ImageDataURL("x.png"); err == nil {
		t.Fatal("outside attachment dir should fail")
	}
	if _, err := ImageDataURL("../.reasonix/attachments/x.png"); err == nil {
		t.Fatal("traversal path should fail")
	}
}

func TestImageDataURLRejectsSymlinkFile(t *testing.T) {
	t.Chdir(t.TempDir())
	if err := ensureAttachmentRoot(); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile("secret.png", []byte("secret"), 0o644); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(".reasonix", "attachments", "link.png")
	if err := os.Symlink(filepath.Join("..", "..", "secret.png"), link); err != nil {
		t.Skipf("symlink unsupported: %v", err)
	}
	if _, err := ImageDataURL(link); err == nil {
		t.Fatal("symlink attachment file should fail")
	}
}

func TestImageDataURLRejectsSymlinkAttachmentDir(t *testing.T) {
	t.Chdir(t.TempDir())
	if err := os.Mkdir(".reasonix", 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir("elsewhere", 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink("../elsewhere", filepath.Join(".reasonix", "attachments")); err != nil {
		t.Skipf("symlink unsupported: %v", err)
	}
	if _, err := ImageDataURL(".reasonix/attachments/x.png"); err == nil {
		t.Fatal("symlink attachment directory should fail")
	}
}

func TestImageDataURLRejectsSymlinkSubdirectory(t *testing.T) {
	t.Chdir(t.TempDir())
	if err := ensureAttachmentRoot(); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir("outside", 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join("outside", "x.png"), mustBase64(t, tinyPNG), 0o644); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(".reasonix", "attachments", "link")
	if err := os.Symlink(filepath.Join("..", "..", "outside"), link); err != nil {
		t.Skipf("symlink unsupported: %v", err)
	}
	if _, err := ImageDataURL(filepath.Join(link, "x.png")); err == nil {
		t.Fatal("symlink attachment subdirectory should fail")
	}
}

func mustBase64(t *testing.T, s string) []byte {
	t.Helper()
	raw, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		t.Fatal(err)
	}
	return raw
}

func stubClipboardTools(t *testing.T, look func(string) (string, error), run func(string, ...string) ([]byte, []byte, error)) {
	t.Helper()
	previousLook := lookClipboardTool
	previousRun := runClipboardTool
	t.Cleanup(func() {
		lookClipboardTool = previousLook
		runClipboardTool = previousRun
	})
	lookClipboardTool = look
	runClipboardTool = run
}

func TestSaveLinuxClipboardImageSeparatesNoImageFromMissingTools(t *testing.T) {
	t.Chdir(t.TempDir())
	stubClipboardTools(t,
		func(string) (string, error) { return "", exec.ErrNotFound },
		func(string, ...string) ([]byte, []byte, error) {
			t.Fatal("missing tools must not run")
			return nil, nil, nil
		},
	)
	_, err := saveLinuxClipboardImage()
	if err == nil || errors.Is(err, ErrNoClipboardImage) {
		t.Fatalf("missing tools reported as an empty clipboard: %v", err)
	}
	if !strings.Contains(err.Error(), "needs wl-paste") {
		t.Fatalf("missing tools lost their actionable message: %v", err)
	}

	stubClipboardTools(t,
		func(name string) (string, error) {
			if name == "wl-paste" {
				return name, nil
			}
			return "", exec.ErrNotFound
		},
		func(_ string, args ...string) ([]byte, []byte, error) {
			if len(args) != 1 || args[0] != "--list-types" {
				t.Fatalf("text clipboard unexpectedly read as an image: %v", args)
			}
			return []byte("text/plain\nUTF8_STRING\n"), nil, nil
		},
	)
	if _, err := saveLinuxClipboardImage(); !errors.Is(err, ErrNoClipboardImage) {
		t.Fatalf("text-only clipboard reported as a broken setup: %v", err)
	}
}

func TestSaveLinuxClipboardImagePreservesProbeFailure(t *testing.T) {
	stubClipboardTools(t,
		func(name string) (string, error) {
			if name == "wl-paste" {
				return name, nil
			}
			return "", exec.ErrNotFound
		},
		func(string, ...string) ([]byte, []byte, error) {
			return nil, []byte("failed to connect to display"), errors.New("display unavailable")
		},
	)

	_, err := saveLinuxClipboardImage()
	if err == nil || errors.Is(err, ErrNoClipboardImage) {
		t.Fatalf("clipboard probe failure reported as no image: %v", err)
	}
	if !strings.Contains(err.Error(), "probe wl-paste clipboard types") {
		t.Fatalf("clipboard probe failure lost its operation: %v", err)
	}
}

func TestSaveLinuxClipboardImagePreservesImageReadFailure(t *testing.T) {
	stubClipboardTools(t,
		func(name string) (string, error) {
			if name == "wl-paste" {
				return name, nil
			}
			return "", exec.ErrNotFound
		},
		func(_ string, args ...string) ([]byte, []byte, error) {
			if len(args) == 1 && args[0] == "--list-types" {
				return []byte("text/plain\nimage/png\n"), nil, nil
			}
			return nil, nil, errors.New("selection changed")
		},
	)

	_, err := saveLinuxClipboardImage()
	if err == nil || errors.Is(err, ErrNoClipboardImage) {
		t.Fatalf("clipboard image read failure reported as no image: %v", err)
	}
	if !strings.Contains(err.Error(), "read clipboard image with wl-paste") {
		t.Fatalf("clipboard image read failure lost its operation: %v", err)
	}
}

func TestSaveLinuxClipboardImageTreatsEmptySelectionAsNoImage(t *testing.T) {
	stubClipboardTools(t,
		func(name string) (string, error) {
			if name == "wl-paste" {
				return name, nil
			}
			return "", exec.ErrNotFound
		},
		func(string, ...string) ([]byte, []byte, error) {
			return nil, []byte("Nothing is copied\n"), errors.New("exit status 1")
		},
	)

	if _, err := saveLinuxClipboardImage(); !errors.Is(err, ErrNoClipboardImage) {
		t.Fatalf("empty clipboard = %v, want ErrNoClipboardImage", err)
	}
}

func TestSaveDarwinClipboardImagePreservesOperationalFailure(t *testing.T) {
	want := errors.New("attachment directory unavailable")
	_, err := saveDarwinClipboardImageWith(func(string) (string, error) {
		return "", want
	})
	if !errors.Is(err, want) {
		t.Fatalf("darwin clipboard operational failure = %v, want %v", err, want)
	}
}

func TestSaveDarwinClipboardImageReturnsNoImageOnlyAfterBothTypesMiss(t *testing.T) {
	var classes []string
	_, err := saveDarwinClipboardImageWith(func(class string) (string, error) {
		classes = append(classes, class)
		return "", ErrNoClipboardImage
	})
	if !errors.Is(err, ErrNoClipboardImage) {
		t.Fatalf("darwin empty clipboard = %v, want ErrNoClipboardImage", err)
	}
	if got, want := strings.Join(classes, ","), "PNGf,JPEG"; got != want {
		t.Fatalf("darwin clipboard classes = %q, want %q", got, want)
	}
}

func TestClassifyDarwinClipboardResultDistinguishesMissingTypeFromFailure(t *testing.T) {
	const marker = "__NO_IMAGE__"
	if err := classifyDarwinClipboardResult([]byte(marker+"\n"), nil, marker); !errors.Is(err, ErrNoClipboardImage) {
		t.Fatalf("darwin no-image marker = %v, want ErrNoClipboardImage", err)
	}

	want := errors.New("osascript failed")
	err := classifyDarwinClipboardResult([]byte("clipboard service unavailable\n"), want, marker)
	if !errors.Is(err, want) || errors.Is(err, ErrNoClipboardImage) {
		t.Fatalf("darwin operational failure = %v, want wrapped %v", err, want)
	}
}

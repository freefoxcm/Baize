package serve

import (
	"bytes"
	"errors"
	"io"
	"mime"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"unicode/utf8"

	"reasonix/internal/fileref"
	fileenc "reasonix/internal/fileutil/encoding"
)

const (
	workspacePreviewLimit = 2 * 1024 * 1024
	workspaceContentLimit = 10 * 1024 * 1024
	workspaceSearchLimit  = 20
)

type workspaceEntry struct {
	Name       string `json:"name"`
	Path       string `json:"path"`
	IsDir      bool   `json:"isDir"`
	Size       int64  `json:"size,omitempty"`
	ModifiedAt string `json:"modifiedAt,omitempty"`
}

type workspacePreview struct {
	Path       string `json:"path"`
	Name       string `json:"name"`
	Kind       string `json:"kind"`
	MIME       string `json:"mime"`
	Size       int64  `json:"size"`
	ModifiedAt string `json:"modifiedAt"`
	Body       string `json:"body,omitempty"`
	Truncated  bool   `json:"truncated"`
	ContentURL string `json:"contentUrl,omitempty"`
}

var workspaceMediaTypes = map[string]string{
	".bmp":  "image/bmp",
	".gif":  "image/gif",
	".jpeg": "image/jpeg",
	".jpg":  "image/jpeg",
	".pdf":  "application/pdf",
	".png":  "image/png",
	".svg":  "image/svg+xml",
	".webp": "image/webp",
}

var workspaceMarkdownExtensions = map[string]bool{
	".md": true, ".markdown": true, ".mdown": true,
}

var workspaceCodeExtensions = map[string]bool{
	".bat": true, ".c": true, ".cc": true, ".cpp": true, ".cs": true,
	".css": true, ".go": true, ".h": true, ".hpp": true, ".ini": true,
	".java": true, ".js": true, ".json": true, ".jsx": true, ".log": true,
	".ps1": true, ".py": true, ".rb": true, ".rs": true, ".sh": true,
	".sql": true, ".toml": true, ".ts": true, ".tsx": true, ".xml": true,
	".yaml": true, ".yml": true,
}

func (s *Server) workspaceEntries(w http.ResponseWriter, r *http.Request) {
	root := s.ctl().WorkspaceRoot()
	if root == "" {
		http.Error(w, "no workspace", http.StatusNotFound)
		return
	}
	raw := strings.TrimSpace(r.URL.Query().Get("path"))
	if raw == "" {
		raw = "."
	}
	dir, rel, status, err := resolveWorkspaceViewPath(root, raw)
	if err != nil {
		http.Error(w, http.StatusText(status), status)
		return
	}
	info, err := os.Stat(dir)
	if err != nil || !info.IsDir() {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		http.Error(w, "unable to read directory", http.StatusInternalServerError)
		return
	}
	dirs := make([]workspaceEntry, 0, len(entries))
	files := make([]workspaceEntry, 0, len(entries))
	for _, entry := range entries {
		path := workspaceChildPath(rel, entry.Name())
		if protectedWorkspacePath(path) || fileref.SkipEntry(path, entry.Name(), entry.IsDir()) {
			continue
		}
		item := workspaceEntry{Name: entry.Name(), Path: path, IsDir: entry.IsDir()}
		entryInfo, err := entry.Info()
		if err != nil {
			continue
		}
		if !item.IsDir && !entryInfo.Mode().IsRegular() {
			continue
		}
		item.ModifiedAt = entryInfo.ModTime().UTC().Format(timeLayout)
		if !item.IsDir {
			item.Size = entryInfo.Size()
			files = append(files, item)
		} else {
			dirs = append(dirs, item)
		}
	}
	sort.Slice(dirs, func(i, j int) bool { return strings.ToLower(dirs[i].Name) < strings.ToLower(dirs[j].Name) })
	sort.Slice(files, func(i, j int) bool { return strings.ToLower(files[i].Name) < strings.ToLower(files[j].Name) })
	writeJSON(w, map[string]any{"path": rel, "entries": append(dirs, files...)})
}

func (s *Server) workspaceSearch(w http.ResponseWriter, r *http.Request) {
	root := s.ctl().WorkspaceRoot()
	if root == "" {
		http.Error(w, "no workspace", http.StatusNotFound)
		return
	}
	query := strings.TrimSpace(r.URL.Query().Get("q"))
	if len(query) < 2 || strings.ContainsAny(query, `/\`) {
		http.Error(w, "query must contain at least two filename characters", http.StatusBadRequest)
		return
	}
	base, err := filepath.Abs(root)
	if err != nil {
		http.Error(w, "invalid workspace", http.StatusInternalServerError)
		return
	}
	results := fileref.Search(base, query, workspaceSearchLimit)
	entries := make([]workspaceEntry, 0, len(results))
	for _, result := range results {
		if protectedWorkspacePath(result.Path) {
			continue
		}
		name := filepath.Base(filepath.FromSlash(result.Path))
		entries = append(entries, workspaceEntry{Name: name, Path: result.Path, IsDir: result.IsDir})
	}
	writeJSON(w, map[string]any{"entries": entries})
}

func (s *Server) workspacePreview(w http.ResponseWriter, r *http.Request) {
	root := s.ctl().WorkspaceRoot()
	if root == "" {
		http.Error(w, "no workspace", http.StatusNotFound)
		return
	}
	raw := strings.TrimSpace(r.URL.Query().Get("path"))
	if raw == "" {
		http.Error(w, "missing path", http.StatusBadRequest)
		return
	}
	path, rel, status, err := resolveWorkspaceViewPath(root, raw)
	if err != nil {
		http.Error(w, http.StatusText(status), status)
		return
	}
	info, err := os.Stat(path)
	if err != nil || info.IsDir() || !info.Mode().IsRegular() {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	kind, contentType := workspacePreviewKind(path)
	preview := workspacePreview{
		Path:       rel,
		Name:       info.Name(),
		Kind:       kind,
		MIME:       contentType,
		Size:       info.Size(),
		ModifiedAt: info.ModTime().UTC().Format(timeLayout),
	}
	if kind == "image" || kind == "pdf" {
		if info.Size() > workspaceContentLimit {
			http.Error(w, "preview too large", http.StatusRequestEntityTooLarge)
			return
		}
		preview.ContentURL = "/workspace/content?path=" + url.QueryEscape(rel)
		writeJSON(w, preview)
		return
	}
	body, truncated, binary, err := readWorkspaceText(path)
	if err != nil {
		http.Error(w, "unable to read file", http.StatusInternalServerError)
		return
	}
	preview.Truncated = truncated
	if binary {
		preview.Kind = "binary"
		preview.MIME = "application/octet-stream"
	} else {
		preview.Body = body
	}
	writeJSON(w, preview)
}

func (s *Server) workspaceContent(w http.ResponseWriter, r *http.Request) {
	root := s.ctl().WorkspaceRoot()
	if root == "" {
		http.Error(w, "no workspace", http.StatusNotFound)
		return
	}
	raw := strings.TrimSpace(r.URL.Query().Get("path"))
	if raw == "" {
		http.Error(w, "missing path", http.StatusBadRequest)
		return
	}
	path, _, status, err := resolveWorkspaceViewPath(root, raw)
	if err != nil {
		http.Error(w, http.StatusText(status), status)
		return
	}
	info, err := os.Stat(path)
	if err != nil || info.IsDir() || !info.Mode().IsRegular() {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	contentType := workspaceMediaTypes[strings.ToLower(filepath.Ext(path))]
	if contentType == "" {
		http.Error(w, "unsupported type", http.StatusUnsupportedMediaType)
		return
	}
	if info.Size() > workspaceContentLimit {
		http.Error(w, "too large", http.StatusRequestEntityTooLarge)
		return
	}
	file, err := os.Open(path)
	if err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	defer file.Close()
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Content-Disposition", mime.FormatMediaType("inline", map[string]string{"filename": info.Name()}))
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("Cache-Control", "no-store")
	if contentType == "image/svg+xml" {
		w.Header().Set("Content-Security-Policy", "sandbox; default-src 'none'; style-src 'unsafe-inline'; img-src data:; script-src 'none'")
	}
	http.ServeContent(w, r, info.Name(), info.ModTime(), file)
}

const timeLayout = "2006-01-02T15:04:05Z"

func resolveWorkspaceViewPath(root, raw string) (string, string, int, error) {
	if protectedWorkspacePath(raw) {
		return "", "", http.StatusForbidden, os.ErrPermission
	}
	path, err := securePathJoin(root, raw)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return "", "", http.StatusNotFound, err
		}
		return "", "", http.StatusBadRequest, err
	}
	rootAbs, err := filepath.Abs(root)
	if err != nil {
		return "", "", http.StatusInternalServerError, err
	}
	if resolved, resolveErr := filepath.EvalSymlinks(rootAbs); resolveErr == nil {
		rootAbs = resolved
	}
	rel, err := filepath.Rel(rootAbs, path)
	if err != nil {
		return "", "", http.StatusBadRequest, err
	}
	rel = filepath.ToSlash(rel)
	if rel == "." {
		rel = ""
	}
	if protectedWorkspacePath(rel) || protectedWorkspacePath(path) {
		return "", "", http.StatusForbidden, os.ErrPermission
	}
	return path, rel, http.StatusOK, nil
}

func protectedWorkspacePath(path string) bool {
	path = strings.ReplaceAll(path, `\`, "/")
	for part := range strings.SplitSeq(path, "/") {
		name := strings.ToLower(strings.TrimSpace(part))
		if name == "" || name == "." || name == ".." {
			continue
		}
		if len(name) == 2 && name[1] == ':' && name[0] >= 'a' && name[0] <= 'z' {
			continue
		}
		if strings.Contains(name, ":") {
			return true
		}
		name = strings.TrimRight(name, " .")
		if name == ".env.example" {
			continue
		}
		if name == ".env" || strings.HasPrefix(name, ".env.") {
			return true
		}
		switch name {
		case ".npmrc", ".pypirc", ".netrc", "_netrc", ".mcp.json", "id_rsa", "id_dsa", "id_ecdsa", "id_ed25519":
			return true
		}
		switch filepath.Ext(name) {
		case ".key", ".pem", ".p12", ".pfx":
			return true
		}
	}
	return false
}

func workspaceChildPath(parent, name string) string {
	parent = strings.Trim(filepath.ToSlash(parent), "/")
	if parent == "" || parent == "." {
		return name
	}
	return parent + "/" + name
}

func workspacePreviewKind(path string) (string, string) {
	ext := strings.ToLower(filepath.Ext(path))
	if contentType := workspaceMediaTypes[ext]; contentType != "" {
		if contentType == "application/pdf" {
			return "pdf", contentType
		}
		return "image", contentType
	}
	if workspaceMarkdownExtensions[ext] {
		return "markdown", "text/markdown; charset=utf-8"
	}
	if ext == ".html" || ext == ".htm" {
		return "html", "text/html; charset=utf-8"
	}
	if workspaceCodeExtensions[ext] {
		contentType := mime.TypeByExtension(ext)
		if contentType == "" {
			contentType = "text/plain; charset=utf-8"
		}
		return "code", contentType
	}
	return "text", "text/plain; charset=utf-8"
}

func readWorkspaceText(path string) (string, bool, bool, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", false, false, err
	}
	defer file.Close()
	buf := make([]byte, workspacePreviewLimit+1)
	n, err := file.Read(buf)
	if err != nil && !errors.Is(err, io.EOF) {
		return "", false, false, err
	}
	data := buf[:n]
	truncated := len(data) > workspacePreviewLimit
	if truncated {
		data = data[:workspacePreviewLimit]
		data = trimWorkspaceUTF8Suffix(data)
	}
	encoding, _ := fileenc.Detect(data)
	utf16 := encoding == fileenc.UTF16LE || encoding == fileenc.UTF16BE || encoding == fileenc.UTF16LENoBOM || encoding == fileenc.UTF16BENoBOM
	if bytes.Contains(data, []byte{0}) && !utf16 {
		return "", truncated, true, nil
	}
	if encoding == fileenc.LossyUTF8 {
		return "", truncated, true, nil
	}
	return string(fileenc.Decode(data, encoding)), truncated, false, nil
}

func trimWorkspaceUTF8Suffix(data []byte) []byte {
	if utf8.Valid(data) {
		return data
	}
	for i := len(data) - 1; i >= 0 && len(data)-i <= utf8.UTFMax; i-- {
		if !utf8.RuneStart(data[i]) {
			continue
		}
		if !utf8.Valid(data[:i]) || utf8.FullRune(data[i:]) {
			return data
		}
		return data[:i]
	}
	return data
}

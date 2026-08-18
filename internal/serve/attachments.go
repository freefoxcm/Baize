package serve

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"reasonix/internal/control"
)

type attachmentTurn struct {
	display string
	input   string
	refLine string
}

func (s *Server) prepareAttachmentTurn(r *http.Request, input string, requested []string) (attachmentTurn, error) {
	items, _, err := s.ctl().ListAttachments(r.Context())
	if err != nil {
		return attachmentTurn{}, err
	}
	available := make(map[string]control.AttachmentInfo, len(items))
	for _, item := range items {
		available[item.Path] = item
	}

	selected := make([]control.AttachmentInfo, 0, len(requested))
	seen := make(map[string]struct{}, len(requested))
	for _, raw := range requested {
		path := filepath.ToSlash(strings.TrimSpace(raw))
		if path == "" {
			return attachmentTurn{}, fmt.Errorf("attachment path is empty")
		}
		if _, ok := seen[path]; ok {
			continue
		}
		item, ok := available[path]
		if !ok {
			return attachmentTurn{}, fmt.Errorf("attachment does not exist in this workspace: %s", path)
		}
		if directImageAttachment(item.Path) && item.Size > 10<<20 {
			return attachmentTurn{}, fmt.Errorf("image attachment exceeds the 10 MiB limit: %s", item.Name)
		}
		seen[path] = struct{}{}
		selected = append(selected, item)
	}
	if len(selected) == 0 {
		return attachmentTurn{}, fmt.Errorf("missing attachments")
	}

	displayRefs := make([]string, 0, len(selected))
	rawRefs := make([]string, 0, len(selected))
	imageCount := 0
	for _, item := range selected {
		rawRef := "@" + control.EscapeRefPath(item.Path)
		rawRefs = append(rawRefs, rawRef)
		displayRefs = append(displayRefs, attachmentDisplayRef(item, rawRef))
		if directImageAttachment(item.Path) {
			imageCount++
		}
	}

	visibleInput := strings.TrimSpace(input)
	modelInput := visibleInput
	if modelInput == "" {
		switch imageCount {
		case len(selected):
			modelInput = "Analyze the attached image or images and respond with the relevant findings."
		case 0:
			modelInput = "Review the attached file or files and respond with the relevant findings."
		default:
			modelInput = "Review the attached images and files and respond with the relevant findings."
		}
	}
	return attachmentTurn{
		display: joinAttachmentText(visibleInput, strings.Join(displayRefs, " ")),
		input:   modelInput,
		refLine: joinAttachmentText(visibleInput, strings.Join(rawRefs, " ")),
	}, nil
}

func attachmentDisplayRef(item control.AttachmentInfo, fallback string) string {
	name := strings.TrimSpace(strings.NewReplacer(
		"[", "", "]", "", "(", "", ")", "", "\r", " ", "\n", " ", "\t", " ",
	).Replace(item.Name))
	name = strings.Join(strings.Fields(name), " ")
	if name == "" {
		name = filepath.Base(item.Path)
	}
	if strings.ContainsAny(item.Path, "\r\n") {
		return fallback
	}
	displayPath := strings.NewReplacer("%", "%25", ")", "%29").Replace(item.Path)
	return "@[" + name + "](" + displayPath + ")"
}

func joinAttachmentText(left, right string) string {
	left = strings.TrimSpace(left)
	right = strings.TrimSpace(right)
	if left == "" {
		return right
	}
	if right == "" {
		return left
	}
	return left + "\n\n" + right
}

func directImageAttachment(path string) bool {
	switch strings.ToLower(filepath.Ext(path)) {
	case ".png", ".jpg", ".jpeg", ".gif", ".webp":
		return true
	default:
		return false
	}
}

func requestIsCrossSite(r *http.Request) bool {
	if strings.EqualFold(strings.TrimSpace(r.Header.Get("Sec-Fetch-Site")), "cross-site") {
		return true
	}
	origin := strings.TrimSpace(r.Header.Get("Origin"))
	if origin == "" {
		return false
	}
	u, err := url.Parse(origin)
	return err != nil || u.Host == "" || !strings.EqualFold(u.Host, r.Host)
}

func (s *Server) attachMultipart(w http.ResponseWriter, r *http.Request) {
	reader, err := r.MultipartReader()
	if err != nil {
		http.Error(w, "invalid multipart body", http.StatusBadRequest)
		return
	}
	var saved *control.AttachmentInfo
	for {
		part, nextErr := reader.NextPart()
		if errors.Is(nextErr, io.EOF) {
			break
		}
		if nextErr != nil {
			http.Error(w, "invalid multipart body", http.StatusBadRequest)
			return
		}
		if part.FileName() == "" {
			_ = part.Close()
			continue
		}
		if saved != nil {
			_ = part.Close()
			_ = s.ctl().DeleteAttachment(r.Context(), saved.Path)
			http.Error(w, "exactly one file is required", http.StatusBadRequest)
			return
		}
		name := sanitizeFilename(part.FileName())
		if name == "" {
			_ = part.Close()
			http.Error(w, "invalid filename", http.StatusBadRequest)
			return
		}
		info, saveErr := s.ctl().SaveAttachment(r.Context(), name, part, -1)
		_ = part.Close()
		if saveErr != nil {
			writeAttachmentError(w, saveErr)
			return
		}
		saved = &info
	}
	if saved == nil {
		http.Error(w, "missing file", http.StatusBadRequest)
		return
	}
	writeJSON(w, saved)
}

func (s *Server) attachments(w http.ResponseWriter, r *http.Request) {
	items, usage, err := s.ctl().ListAttachments(r.Context())
	if err != nil {
		writeAttachmentError(w, err)
		return
	}
	writeJSON(w, map[string]any{"attachments": items, "usage": usage})
}

func (s *Server) deleteAttachment(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Path string `json:"path"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || strings.TrimSpace(body.Path) == "" {
		http.Error(w, "missing path", http.StatusBadRequest)
		return
	}
	if err := s.ctl().DeleteAttachment(r.Context(), body.Path); err != nil {
		writeAttachmentError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) clearAttachments(w http.ResponseWriter, r *http.Request) {
	removed, err := s.ctl().ClearAttachments(r.Context())
	if err != nil {
		writeAttachmentError(w, err)
		return
	}
	writeJSON(w, map[string]int{"removed": removed})
}

func writeAttachmentError(w http.ResponseWriter, err error) {
	status := http.StatusInternalServerError
	switch {
	case errors.Is(err, control.ErrAttachmentTooLarge):
		status = http.StatusRequestEntityTooLarge
	case errors.Is(err, control.ErrAttachmentQuota):
		status = http.StatusInsufficientStorage
	case errors.Is(err, control.ErrAttachmentInvalid), errors.Is(err, os.ErrNotExist):
		status = http.StatusBadRequest
	}
	http.Error(w, err.Error(), status)
}

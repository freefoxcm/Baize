package serve

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"

	"reasonix/internal/control"
)

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

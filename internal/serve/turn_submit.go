package serve

import (
	"encoding/json"
	"net/http"
	"strings"
)

type submitRequest struct {
	Input       string   `json:"input"`
	Format      string   `json:"format"`
	Action      string   `json:"action"`
	Attachments []string `json:"attachments"`
}

// submit starts an HTTP turn. Intercepted management commands return 204;
// accepted agent turns return 202 and stream output through /events.
func (s *Server) submit(w http.ResponseWriter, r *http.Request) {
	var body submitRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || strings.TrimSpace(body.Input) == "" && len(body.Attachments) == 0 {
		http.Error(w, "missing input", http.StatusBadRequest)
		return
	}
	body.Format = strings.TrimSpace(body.Format)
	body.Action = strings.TrimSpace(body.Action)
	if body.Format != "" && body.Format != "json_object" {
		http.Error(w, `unsupported format (supported: "json_object")`, http.StatusBadRequest)
		return
	}
	if err := validateSubmitAction(body.Format, body.Action); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	trimmed := strings.TrimSpace(body.Input)
	if strings.HasPrefix(trimmed, "!") {
		http.Error(w, "shell commands are unavailable over HTTP", http.StatusForbidden)
		return
	}
	if len(body.Attachments) > 0 && (body.Action != "" || strings.HasPrefix(trimmed, "/")) {
		http.Error(w, "attachments cannot be combined with actions or slash commands", http.StatusBadRequest)
		return
	}
	if s.interceptSubmitCommand(w, r, trimmed) {
		return
	}

	s.bindMu.Lock()
	ctrl := s.ctl()
	if len(body.Attachments) > 0 && strings.TrimSpace(ctrl.Goal()) != "" {
		s.bindMu.Unlock()
		http.Error(w, "attachments are unavailable while a Goal is active", http.StatusConflict)
		return
	}
	var attachmentInput attachmentTurn
	if len(body.Attachments) > 0 {
		var err error
		attachmentInput, err = s.prepareAttachmentTurn(r, body.Input, body.Attachments)
		if err != nil {
			s.bindMu.Unlock()
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	}
	if ctrl.Running() {
		s.bindMu.Unlock()
		http.Error(w, "session is busy; use POST /inbox/items for durable follow-up", http.StatusConflict)
		return
	}
	if len(body.Attachments) > 0 {
		ctrl.SubmitHTTPAttachmentTurn(attachmentInput.display, attachmentInput.input, attachmentInput.refLine, body.Format)
	} else {
		submitWithAction(ctrl, body.Input, body.Format, body.Action)
	}
	if !ctrl.Running() && !ctrl.RuntimeStatus().PendingPrompt {
		s.bindMu.Unlock()
		http.Error(w, "input was not admitted; session is rotating, closed, or finishing — use POST /inbox/items", http.StatusConflict)
		return
	}
	s.bindMu.Unlock()
	w.WriteHeader(http.StatusAccepted)
}

func (s *Server) interceptSubmitCommand(w http.ResponseWriter, r *http.Request, input string) bool {
	if strings.HasPrefix(input, "/model ") {
		ref := strings.TrimSpace(strings.TrimPrefix(input, "/model"))
		if ref == "" {
			return false
		}
		if err := s.switchModel(r.Context(), ref); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return true
		}
		w.WriteHeader(http.StatusNoContent)
		return true
	}
	if strings.HasPrefix(input, "/switch ") {
		if err := s.switchSession(input); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return true
		}
		w.WriteHeader(http.StatusNoContent)
		return true
	}
	if input != "/effort" && !strings.HasPrefix(input, "/effort ") && input != "/thinking" && !strings.HasPrefix(input, "/thinking ") {
		return false
	}
	level := strings.TrimSpace(strings.TrimPrefix(input, "/effort"))
	if rest, ok := strings.CutPrefix(input, "/thinking"); ok {
		level = strings.TrimSpace(rest)
	}
	if level == "" {
		s.effort(w, r)
		return true
	}
	if err := s.switchEffort(r.Context(), level); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return true
	}
	w.WriteHeader(http.StatusNoContent)
	return true
}

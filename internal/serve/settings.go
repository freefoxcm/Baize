package serve

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sort"
	"strings"
	"sync"

	"reasonix/internal/agent"
	"reasonix/internal/config"
)

type settingsRuntimeState struct {
	mu       sync.Mutex
	pending  bool
	applying bool
	lastErr  string
}

type safeSettings struct {
	DefaultModel           string   `json:"defaultModel"`
	PlannerModel           string   `json:"plannerModel"`
	SubagentModel          string   `json:"subagentModel"`
	SubagentEffort         string   `json:"subagentEffort"`
	DefaultApprovalMode    string   `json:"defaultApprovalMode"`
	MaxSubagentDepth       int      `json:"maxSubagentDepth"`
	MaxSubagentConcurrency int      `json:"maxSubagentConcurrency"`
	MaxParallelWriters     int      `json:"maxParallelWriters"`
	CompactRatio           float64  `json:"compactRatio"`
	ReasoningLanguage      string   `json:"reasoningLanguage"`
	Models                 []string `json:"models"`
}

type settingsResponse struct {
	Revision   string       `json:"revision"`
	Global     safeSettings `json:"global"`
	Effective  safeSettings `json:"effective"`
	Overridden []string     `json:"overridden"`
	Apply      string       `json:"apply"`
	ApplyError string       `json:"applyError,omitempty"`
}

type settingsPatch struct {
	Revision               string   `json:"revision"`
	DefaultModel           *string  `json:"defaultModel,omitempty"`
	PlannerModel           *string  `json:"plannerModel,omitempty"`
	SubagentModel          *string  `json:"subagentModel,omitempty"`
	SubagentEffort         *string  `json:"subagentEffort,omitempty"`
	DefaultApprovalMode    *string  `json:"defaultApprovalMode,omitempty"`
	MaxSubagentDepth       *int     `json:"maxSubagentDepth,omitempty"`
	MaxSubagentConcurrency *int     `json:"maxSubagentConcurrency,omitempty"`
	MaxParallelWriters     *int     `json:"maxParallelWriters,omitempty"`
	CompactRatio           *float64 `json:"compactRatio,omitempty"`
	ReasoningLanguage      *string  `json:"reasoningLanguage,omitempty"`
}

func (s *Server) settingsView(w http.ResponseWriter, _ *http.Request) {
	view, err := s.safeSettingsView()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, view)
}

func (s *Server) patchSettings(w http.ResponseWriter, r *http.Request) {
	var patch settingsPatch
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 64<<10)).Decode(&patch); err != nil {
		http.Error(w, "invalid settings payload", http.StatusBadRequest)
		return
	}
	path := config.UserConfigPath()
	if path == "" {
		http.Error(w, "no user config path", http.StatusInternalServerError)
		return
	}
	unlocked := config.LockUserConfigEdits()
	locked := true
	defer func() {
		if locked {
			unlocked()
		}
	}()
	currentRevision, err := settingsRevision(path)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if patch.Revision == "" || patch.Revision != currentRevision {
		unlocked()
		locked = false
		view, viewErr := s.safeSettingsView()
		if viewErr != nil {
			http.Error(w, viewErr.Error(), http.StatusConflict)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusConflict)
		_ = json.NewEncoder(w).Encode(view)
		return
	}
	cfg := config.LoadForEdit(path)
	if err := applySafeSettingsPatch(cfg, patch); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err := cfg.SaveTo(path); err != nil {
		http.Error(w, fmt.Sprintf("save settings: %v", err), http.StatusInternalServerError)
		return
	}
	unlocked()
	locked = false
	s.queueOrApplySettings()
	view, err := s.safeSettingsView()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, view)
}

func (s *Server) applySettings(w http.ResponseWriter, r *http.Request) {
	if controllerHasActiveRuntimeWork(s.ctl()) {
		s.markSettingsPending("")
		w.WriteHeader(http.StatusAccepted)
		return
	}
	if err := s.applySettingsRuntime(r.Context()); err != nil {
		http.Error(w, err.Error(), http.StatusConflict)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func applySafeSettingsPatch(cfg *config.Config, patch settingsPatch) error {
	if err := applySafeModelSettings(cfg, patch); err != nil {
		return err
	}
	if patch.DefaultApprovalMode != nil {
		if err := cfg.SetDesktopDefaultToolApprovalMode(*patch.DefaultApprovalMode); err != nil {
			return err
		}
	}
	if err := applySafeConcurrencySettings(cfg, patch); err != nil {
		return err
	}
	if patch.CompactRatio != nil {
		if err := cfg.SetCompactRatio(*patch.CompactRatio); err != nil {
			return err
		}
	}
	if patch.ReasoningLanguage != nil {
		if err := cfg.SetReasoningLanguage(*patch.ReasoningLanguage); err != nil {
			return err
		}
	}
	return nil
}

func applySafeModelSettings(cfg *config.Config, patch settingsPatch) error {
	if patch.DefaultModel != nil {
		model := strings.TrimSpace(*patch.DefaultModel)
		entry, ok := cfg.ResolveModel(model)
		if !ok || !entry.Configured() {
			return fmt.Errorf("default model %q is unavailable", model)
		}
		if err := cfg.SetDefaultModel(model); err != nil {
			return err
		}
	}
	setModel := func(target *string, raw string, label string) error {
		raw = strings.TrimSpace(raw)
		if raw == "" {
			*target = ""
			return nil
		}
		entry, ok := cfg.ResolveModel(raw)
		if !ok || !entry.Configured() {
			return fmt.Errorf("%s model %q is unavailable", label, raw)
		}
		*target = entry.Name + "/" + entry.Model
		return nil
	}
	if patch.PlannerModel != nil {
		if err := setModel(&cfg.Agent.PlannerModel, *patch.PlannerModel, "planner"); err != nil {
			return err
		}
	}
	if patch.SubagentModel != nil {
		if err := setModel(&cfg.Agent.SubagentModel, *patch.SubagentModel, "subagent"); err != nil {
			return err
		}
	}
	if patch.SubagentEffort != nil {
		level := strings.TrimSpace(*patch.SubagentEffort)
		if level == "" || level == "auto" {
			cfg.Agent.SubagentEffort = ""
		} else {
			model := cfg.Agent.SubagentModel
			if model == "" {
				model = cfg.DefaultModel
			}
			entry, ok := cfg.ResolveModel(model)
			if !ok {
				return fmt.Errorf("unknown subagent model %q", model)
			}
			effort, err := config.NormalizeEffort(entry, level)
			if err != nil {
				return err
			}
			cfg.Agent.SubagentEffort = effort
		}
	}
	return nil
}

func applySafeConcurrencySettings(cfg *config.Config, patch settingsPatch) error {
	if patch.MaxSubagentDepth != nil {
		if *patch.MaxSubagentDepth < 1 || *patch.MaxSubagentDepth > agent.DefaultMaxSubagentDepth {
			return fmt.Errorf("maxSubagentDepth must be between 1 and %d", agent.DefaultMaxSubagentDepth)
		}
		cfg.Agent.MaxSubagentDepth = *patch.MaxSubagentDepth
	}
	total, writers := agent.NormalizeConcurrencyLimits(cfg.Agent.MaxSubagentConcurrency, cfg.Agent.MaxParallelWriters)
	if patch.MaxSubagentConcurrency != nil {
		if *patch.MaxSubagentConcurrency < 1 || *patch.MaxSubagentConcurrency > 32 {
			return fmt.Errorf("maxSubagentConcurrency must be between 1 and 32")
		}
		total = *patch.MaxSubagentConcurrency
	}
	if patch.MaxParallelWriters != nil {
		if *patch.MaxParallelWriters < 1 || *patch.MaxParallelWriters > 32 {
			return fmt.Errorf("maxParallelWriters must be between 1 and 32")
		}
		writers = *patch.MaxParallelWriters
	}
	if writers > total {
		return fmt.Errorf("maxParallelWriters cannot exceed maxSubagentConcurrency")
	}
	cfg.Agent.MaxSubagentConcurrency, cfg.Agent.MaxParallelWriters = total, writers
	return nil
}

func (s *Server) safeSettingsView() (settingsResponse, error) {
	global, err := config.LoadUserConfigReadOnly()
	if err != nil {
		return settingsResponse{}, err
	}
	effective, err := config.LoadForRootReadOnly(s.ctl().WorkspaceRoot())
	if err != nil {
		return settingsResponse{}, err
	}
	revision, err := settingsRevision(config.UserConfigPath())
	if err != nil {
		return settingsResponse{}, err
	}
	globalView := safeSettingsFromConfig(global)
	effectiveView := safeSettingsFromConfig(effective)
	state, lastErr := s.settingsState()
	return settingsResponse{Revision: revision, Global: globalView, Effective: effectiveView, Overridden: settingsOverrides(globalView, effectiveView), Apply: state, ApplyError: lastErr}, nil
}

func safeSettingsFromConfig(cfg *config.Config) safeSettings {
	total, writers := agent.NormalizeConcurrencyLimits(cfg.Agent.MaxSubagentConcurrency, cfg.Agent.MaxParallelWriters)
	models := make([]string, 0)
	for i := range cfg.Providers {
		entry := &cfg.Providers[i]
		if !entry.Configured() {
			continue
		}
		for _, model := range entry.ModelList() {
			models = append(models, entry.Name+"/"+model)
		}
	}
	sort.Strings(models)
	reasoningLanguage := cfg.Agent.ReasoningLanguage
	if reasoningLanguage == "" {
		reasoningLanguage = "auto"
	}
	effort := cfg.Agent.SubagentEffort
	if effort == "" {
		effort = "auto"
	}
	return safeSettings{
		DefaultModel: cfg.DefaultModel, PlannerModel: cfg.Agent.PlannerModel,
		SubagentModel: cfg.Agent.SubagentModel, SubagentEffort: effort,
		DefaultApprovalMode:    cfg.DesktopDefaultToolApprovalMode(),
		MaxSubagentDepth:       agent.NormalizeMaxSubagentDepth(cfg.Agent.MaxSubagentDepth),
		MaxSubagentConcurrency: total, MaxParallelWriters: writers,
		CompactRatio: cfg.Agent.CompactRatio, ReasoningLanguage: reasoningLanguage,
		Models: models,
	}
}

func settingsOverrides(global, effective safeSettings) []string {
	var out []string
	if global.DefaultModel != effective.DefaultModel {
		out = append(out, "defaultModel")
	}
	if global.PlannerModel != effective.PlannerModel {
		out = append(out, "plannerModel")
	}
	if global.SubagentModel != effective.SubagentModel {
		out = append(out, "subagentModel")
	}
	if global.SubagentEffort != effective.SubagentEffort {
		out = append(out, "subagentEffort")
	}
	if global.MaxSubagentDepth != effective.MaxSubagentDepth {
		out = append(out, "maxSubagentDepth")
	}
	if global.MaxSubagentConcurrency != effective.MaxSubagentConcurrency {
		out = append(out, "maxSubagentConcurrency")
	}
	if global.MaxParallelWriters != effective.MaxParallelWriters {
		out = append(out, "maxParallelWriters")
	}
	if global.CompactRatio != effective.CompactRatio {
		out = append(out, "compactRatio")
	}
	if global.ReasoningLanguage != effective.ReasoningLanguage {
		out = append(out, "reasoningLanguage")
	}
	return out
}

func settingsRevision(path string) (string, error) {
	raw, err := os.ReadFile(path)
	if err != nil && !os.IsNotExist(err) {
		return "", err
	}
	hash := sha256.Sum256(raw)
	return hex.EncodeToString(hash[:]), nil
}

func (s *Server) queueOrApplySettings() {
	if controllerHasActiveRuntimeWork(s.ctl()) {
		s.markSettingsPending("")
		return
	}
	go func() { _ = s.applySettingsRuntime(context.Background()) }()
}

func (s *Server) onSettingsTurnDone() {
	if s.settings == nil {
		return
	}
	s.settings.mu.Lock()
	pending := s.settings.pending && !s.settings.applying
	s.settings.mu.Unlock()
	if pending {
		go func() { _ = s.applySettingsRuntime(context.Background()) }()
	}
}

func (s *Server) applySettingsRuntime(ctx context.Context) error {
	if s.settings == nil {
		s.settings = &settingsRuntimeState{}
	}
	s.settings.mu.Lock()
	if s.settings.applying {
		s.settings.mu.Unlock()
		return fmt.Errorf("settings are already applying")
	}
	s.settings.applying = true
	s.settings.pending = false
	s.settings.lastErr = ""
	s.settings.mu.Unlock()
	err := s.reloadExtensions(ctx)
	s.settings.mu.Lock()
	s.settings.applying = false
	if err != nil {
		s.settings.pending = true
		s.settings.lastErr = err.Error()
	}
	s.settings.mu.Unlock()
	return err
}

func (s *Server) markSettingsPending(errText string) {
	if s.settings == nil {
		s.settings = &settingsRuntimeState{}
	}
	s.settings.mu.Lock()
	s.settings.pending = true
	if errText != "" {
		s.settings.lastErr = errText
	}
	s.settings.mu.Unlock()
}

func (s *Server) settingsState() (string, string) {
	if s.settings == nil {
		return "applied", ""
	}
	s.settings.mu.Lock()
	defer s.settings.mu.Unlock()
	state := "applied"
	if s.settings.applying {
		state = "applying"
	} else if s.settings.pending {
		state = "pending"
	}
	return state, s.settings.lastErr
}

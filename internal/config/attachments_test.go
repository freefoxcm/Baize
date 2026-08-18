package config

import (
	"strings"
	"testing"

	"github.com/BurntSushi/toml"
)

func TestDefaultAttachmentLimits(t *testing.T) {
	cfg := Default()
	if cfg.Attachments.MaxFileMiB != 100 || cfg.Attachments.WorkspaceQuotaMiB != 1024 {
		t.Fatalf("default attachment limits = %+v, want 100/1024 MiB", cfg.Attachments)
	}
	if err := cfg.Attachments.Validate(); err != nil {
		t.Fatalf("default attachment limits are invalid: %v", err)
	}
}

func TestZeroAttachmentConfigUsesDefaultsForProgrammaticCallers(t *testing.T) {
	if got := (AttachmentsConfig{}).Effective(); got != defaultAttachmentsConfig() {
		t.Fatalf("zero-value effective limits = %+v", got)
	}
}

func TestAttachmentLimitsValidateBoundaries(t *testing.T) {
	tests := []struct {
		name    string
		config  AttachmentsConfig
		wantErr bool
	}{
		{name: "minimum", config: AttachmentsConfig{MaxFileMiB: 1, WorkspaceQuotaMiB: 1}},
		{name: "maximum file", config: AttachmentsConfig{MaxFileMiB: 1024, WorkspaceQuotaMiB: 1024}},
		{name: "zero", config: AttachmentsConfig{MaxFileMiB: 0, WorkspaceQuotaMiB: 1024}, wantErr: true},
		{name: "over maximum", config: AttachmentsConfig{MaxFileMiB: 1025, WorkspaceQuotaMiB: 2048}, wantErr: true},
		{name: "quota below file", config: AttachmentsConfig{MaxFileMiB: 100, WorkspaceQuotaMiB: 99}, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Fatalf("Validate() error = %v, wantErr=%v", err, tt.wantErr)
			}
		})
	}
}

func TestAttachmentLimitsRenderAndDecodeRoundTrip(t *testing.T) {
	cfg := Default()
	cfg.Attachments = AttachmentsConfig{MaxFileMiB: 256, WorkspaceQuotaMiB: 2048}
	rendered := RenderTOMLForScope(cfg, RenderScopeUser)
	for _, want := range []string{"[attachments]", "max_file_mib = 256", "workspace_quota_mib = 2048"} {
		if !strings.Contains(rendered, want) {
			t.Fatalf("rendered config missing %q:\n%s", want, rendered)
		}
	}
	got := Default()
	if _, err := toml.Decode(rendered, got); err != nil {
		t.Fatalf("decode rendered TOML: %v", err)
	}
	if got.Attachments != cfg.Attachments {
		t.Fatalf("round trip limits = %+v, want %+v", got.Attachments, cfg.Attachments)
	}
}

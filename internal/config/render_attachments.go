package config

import (
	"fmt"
	"strings"
)

func renderConfigHeader(b *strings.Builder, c *Config) {
	b.WriteString("# Reasonix configuration.\n")
	fmt.Fprintf(b, "# Resolution order: flag > ./reasonix.toml > %s > built-in defaults.\n", userConfigDisplayPath())
	b.WriteString("# Fields marked user/global only are not overridden by ./reasonix.toml.\n")
	b.WriteString("# Secrets are named via api_key_env and stored in Reasonix's global .env; never put keys here.\n\n")
	fmt.Fprintf(b, "config_version = %d   # schema marker for diagnostics; old versions may ignore it\n", configVersion(c))
}

func renderAttachmentConfig(b *strings.Builder, c, defaults *Config, scope RenderScope) {
	if scope == RenderScopeProject && c.Attachments == defaults.Attachments {
		return
	}
	b.WriteString("[attachments]\n")
	fmt.Fprintf(b, "max_file_mib = %d   # per-file attachment limit, 1-1024 MiB\n", c.Attachments.MaxFileMiB)
	fmt.Fprintf(b, "workspace_quota_mib = %d   # total .reasonix/attachments quota; must be >= max_file_mib\n\n", c.Attachments.WorkspaceQuotaMiB)
}

func renderAttachmentDelta(b *strings.Builder, c, defaults *Config) {
	if c.Attachments == defaults.Attachments {
		return
	}
	b.WriteString("[attachments]\n")
	fmt.Fprintf(b, "max_file_mib = %d\n", c.Attachments.MaxFileMiB)
	fmt.Fprintf(b, "workspace_quota_mib = %d\n\n", c.Attachments.WorkspaceQuotaMiB)
}

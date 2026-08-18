package config

import "fmt"

const (
	DefaultAttachmentMaxFileMiB        = 100
	DefaultAttachmentWorkspaceQuotaMiB = 1024
)

// AttachmentsConfig controls workspace-local attachment storage. Limits are
// expressed in MiB so reasonix.toml stays readable and platform independent.
type AttachmentsConfig struct {
	MaxFileMiB        int `toml:"max_file_mib"`
	WorkspaceQuotaMiB int `toml:"workspace_quota_mib"`
}

func (c AttachmentsConfig) MaxFileBytes() int64 {
	return int64(c.MaxFileMiB) * 1024 * 1024
}

func (c AttachmentsConfig) WorkspaceQuotaBytes() int64 {
	return int64(c.WorkspaceQuotaMiB) * 1024 * 1024
}

func (c AttachmentsConfig) Effective() AttachmentsConfig {
	if c == (AttachmentsConfig{}) {
		return defaultAttachmentsConfig()
	}
	return c
}

func (c AttachmentsConfig) Validate() error {
	if c.MaxFileMiB < 1 || c.MaxFileMiB > 1024 {
		return fmt.Errorf("attachments.max_file_mib must be between 1 and 1024")
	}
	if c.WorkspaceQuotaMiB < c.MaxFileMiB {
		return fmt.Errorf("attachments.workspace_quota_mib must be greater than or equal to attachments.max_file_mib")
	}
	return nil
}

func defaultAttachmentsConfig() AttachmentsConfig {
	return AttachmentsConfig{
		MaxFileMiB:        DefaultAttachmentMaxFileMiB,
		WorkspaceQuotaMiB: DefaultAttachmentWorkspaceQuotaMiB,
	}
}

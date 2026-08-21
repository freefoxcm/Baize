package config

import (
	"fmt"
	"strings"
)

func renderServeConfig(b *strings.Builder, cfg ServeConfig) {
	b.WriteString("[serve]\n")
	fmt.Fprintf(b, "show_task_errors = %v   # Serve: show failed tool and subagent cards; task failure notices always remain visible\n", cfg.ShowTaskErrors)
	fmt.Fprintf(b, "auth_mode = %q   # none|token|password\n", cfg.AuthMode)
	fmt.Fprintf(b, "token = %q\n", cfg.Token)
	fmt.Fprintf(b, "password_hash = %q\n", cfg.PasswordHash)
	fmt.Fprintf(b, "behind_proxy = %v\n\n[desktop]\n", cfg.BehindProxy)
}

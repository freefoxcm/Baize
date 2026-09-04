package cli

import (
	"errors"
	"fmt"
	"os"

	"reasonix/internal/config"
	"reasonix/internal/i18n"
	"reasonix/internal/netclient"
)

func providerManagerItems(s *providerSetupSession) []menuItem {
	cfg := s.cfg
	items := make([]menuItem, 0, len(cfg.Providers)+5)
	for _, p := range cfg.Providers {
		models := p.ModelList()
		keyStatus := i18n.M.SetupKeyMissing
		if p.APIKeyEnv == "" || config.CredentialIsSet(p.APIKeyEnv) || s.pendingCredentials[p.APIKeyEnv] != "" {
			keyStatus = i18n.M.SetupKeySet
		}
		desc := fmt.Sprintf("%s · %d %s · %s", p.Kind, len(models), i18n.M.SetupModelsUnit, keyStatus)
		if cfg.DefaultModel == p.Name || config.ModelRefsProvider(cfg.DefaultModel, p.Name) {
			desc += " · " + i18n.M.SetupDefaultBadge
		}
		items = append(items, menuItem{name: p.Name, desc: desc})
	}
	return append(items,
		menuItem{name: i18n.M.SetupAddOpenAI, desc: i18n.M.CustomProviderDesc},
		menuItem{name: i18n.M.SetupAddAnthropic, desc: i18n.M.AnthropicProviderDesc},
		menuItem{name: i18n.M.SetupAddResponses, desc: i18n.M.ResponsesProviderDesc},
		menuItem{name: i18n.M.SetupSaveExit, desc: i18n.M.SetupSaveExitDesc},
		menuItem{name: i18n.M.SetupCancel, desc: i18n.M.SetupCancelDesc},
	)
}

func addProviderToSession(s *providerSetupSession, kind string) bool {
	var result providerPromptResult
	var err error
	var proxy netclient.ProxySpec
	if s != nil && s.cfg != nil {
		proxy = s.cfg.NetworkProxySpec()
	}
	switch kind {
	case "anthropic":
		result, err = promptAnthropicProvider(proxy)
	case "responses":
		result, err = promptResponsesProvider(proxy)
	default:
		result, err = promptCustomProvider(proxy)
	}
	if err != nil {
		if !errors.Is(err, errCancelled) {
			fmt.Fprintln(os.Stderr, err)
		}
		return false
	}
	for i := range result.entries {
		result.entries[i], err = promptProviderVision(result.entries[i])
		if err != nil {
			return false
		}
	}
	for _, entry := range result.entries {
		if !confirmSharedCredential(s.cfg, entry, "") {
			return false
		}
	}
	if err := s.add(result.entries); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return false
	}
	s.addProviderAccess(result.entries)
	for key, value := range result.credentials {
		if err := s.setCredential(key, value); err != nil {
			fmt.Fprintln(os.Stderr, err)
			return false
		}
	}
	// After the new keys are staged, so usability sees them.
	s.promoteDefaultToNewProviders(result.entries)
	return true
}

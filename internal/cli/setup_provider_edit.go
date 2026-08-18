package cli

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"reasonix/internal/config"
	"reasonix/internal/i18n"
)

func editProvider(s *providerSetupSession, current config.ProviderEntry) {
	in := bufio.NewScanner(os.Stdin)
	edited := current
	edited.BaseURL = ask(in, os.Stdout, i18n.M.CustomPromptBaseURL, current.BaseURL)
	models := ask(in, os.Stdout, i18n.M.SetupPromptModels, strings.Join(current.ModelList(), ","))
	edited.Models = splitModels(models)
	if len(edited.Models) == 1 {
		edited.Model = edited.Models[0]
	} else {
		edited.Model = ""
	}
	if len(edited.Models) > 0 && !containsString(edited.Models, edited.Default) {
		edited.Default = edited.Models[0]
	}
	edited.APIKeyEnv = promptOptionalAPIKeyEnvName(in, os.Stdout, i18n.M.CustomPromptKeyEnv, current.APIKeyEnv)
	var err error
	edited, err = promptProviderVision(edited)
	if err != nil {
		return
	}
	if !confirmSharedCredential(s.cfg, edited, current.Name) {
		return
	}
	if err := s.upsert([]config.ProviderEntry{edited}); err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
}

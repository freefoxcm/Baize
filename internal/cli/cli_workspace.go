package cli

import (
	"fmt"
	"os"

	"reasonix/internal/i18n"
)

func prepareCLIWorkspace(dir string) (string, int) {
	if code := chdirTo(dir); code != 0 {
		return "", code
	}
	root, err := workspaceRootForDir(dir)
	if err != nil {
		fmt.Fprintln(os.Stderr, i18n.M.ErrorPrefix, err)
		return "", 1
	}
	return root, 0
}

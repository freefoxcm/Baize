package main

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestBaizeForkArtifactNames(t *testing.T) {
	root := filepath.Join("..", "..")
	wants := map[string][]string{
		"Makefile": {
			`-o bin/baize$(GOEXE) ./cmd/reasonix`,
			`-o bin/baize-plugin-example$(GOEXE) ./cmd/reasonix-plugin-example`,
			`-o dist/baize-$$os-$$arch$$ext ./cmd/reasonix`,
		},
		"Dockerfile": {
			`-o /out/baize ./cmd/reasonix`,
			`COPY --from=builder /out/baize /usr/local/bin/baize`,
			`ENTRYPOINT ["baize-entrypoint"]`,
		},
		filepath.Join("docker", "baize-entrypoint.sh"): {
			`exec /usr/local/bin/baize "$@"`,
		},
		filepath.Join(".github", "workflows", "baize-ci.yml"): {
			`go build -o "$RUNNER_TEMP/baize" ./cmd/reasonix`,
			`go build -o "$env:RUNNER_TEMP\baize.exe" ./cmd/reasonix`,
		},
		".goreleaser.baize.yaml": {
			`project_name: baize`,
			`binary: baize`,
			`name_template: "baize-{{ .Os }}-{{ .Arch }}"`,
		},
	}
	for name, markers := range wants {
		data, err := os.ReadFile(filepath.Join(root, name))
		if err != nil {
			t.Fatalf("read %s: %v", name, err)
		}
		text := string(data)
		for _, marker := range markers {
			if !strings.Contains(text, marker) {
				t.Errorf("%s missing Baize artifact contract %q", name, marker)
			}
		}
	}

	dockerfile, err := os.ReadFile(filepath.Join(root, "Dockerfile"))
	if err != nil {
		t.Fatal(err)
	}
	for _, forbidden := range []string{`/usr/local/bin/reasonix`, `reasonix-entrypoint`} {
		if strings.Contains(string(dockerfile), forbidden) {
			t.Errorf("Dockerfile still installs legacy command %q", forbidden)
		}
	}
	if _, err := os.Stat(filepath.Join(root, "docker", "reasonix-entrypoint.sh")); !errors.Is(err, os.ErrNotExist) {
		t.Errorf("legacy docker entrypoint still exists: %v", err)
	}
}

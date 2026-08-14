package cli

import (
	"fmt"
	"io"
	"os"
	"runtime"
	"strconv"
	"strings"

	"golang.org/x/crypto/bcrypt"
	"golang.org/x/term"

	"reasonix/internal/fileutil"
)

// readServeTokenFile loads the auth=token pre-shared token from a file so the
// secret never appears in argv (visible via ps). The file must hold a single
// non-empty line and, on POSIX systems, must not be group/world accessible.
func readServeTokenFile(path string) (string, error) {
	return readServeSecretFile(path, "token")
}

func readServePasswordHashFile(path string) (string, error) {
	hash, err := readServeSecretFile(path, "password hash")
	if err != nil {
		return "", err
	}
	if _, err := bcrypt.Cost([]byte(hash)); err != nil {
		return "", fmt.Errorf("password hash file %s does not contain a valid bcrypt hash: %w", path, err)
	}
	return hash, nil
}

func readServeSecretFile(path, kind string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	fi, err := f.Stat()
	if err != nil {
		return "", err
	}
	if !fi.Mode().IsRegular() {
		return "", fmt.Errorf("%s file %s must be a regular file", kind, path)
	}
	if runtime.GOOS != "windows" && fi.Mode().Perm()&0o077 != 0 {
		return "", fmt.Errorf("%s file %s must not be group/world accessible (chmod 600)", kind, path)
	}
	b, err := io.ReadAll(io.LimitReader(f, (64<<10)+1))
	if err != nil {
		return "", err
	}
	if len(b) > 64<<10 {
		return "", fmt.Errorf("%s file %s is too large", kind, path)
	}
	secret := strings.TrimSpace(string(b))
	if secret == "" {
		return "", fmt.Errorf("%s file %s is empty", kind, path)
	}
	if strings.ContainsAny(secret, "\r\n") {
		return "", fmt.Errorf("%s file %s must hold a single line", kind, path)
	}
	return secret, nil
}

func promptServePassword(in *os.File, out io.Writer) (string, error) {
	fd := int(in.Fd())
	if !term.IsTerminal(fd) {
		return "", fmt.Errorf("--hash-password requires --password when stdin is not a terminal")
	}
	if _, err := fmt.Fprint(out, "Password: "); err != nil {
		return "", err
	}
	b, err := term.ReadPassword(fd)
	_, _ = fmt.Fprintln(out)
	if err != nil {
		return "", err
	}
	password := string(b)
	if password == "" {
		return "", fmt.Errorf("password must not be empty")
	}
	return password, nil
}

// writeServeAddrFile records the actual bound listen address (host:port) so a
// supervisor that started serve with --addr 127.0.0.1:0 can discover the real
// port. Written atomically with owner-only permissions.
func writeServeAddrFile(path, addr string) error {
	return fileutil.AtomicWriteFile(path, []byte(addr+"\n"), 0o600)
}

// writeServePidFile records the server's pid for supervisors that cannot
// capture the shell's $! (or want a belt-and-braces check).
func writeServePidFile(path string) error {
	return fileutil.AtomicWriteFile(path, []byte(strconv.Itoa(os.Getpid())+"\n"), 0o600)
}

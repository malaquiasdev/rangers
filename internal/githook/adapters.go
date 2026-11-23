package githook

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const executeBits = 0o111

type OSFiles struct{}

type GitCLI struct {
	Dir string
}

func (OSFiles) ReadFile(path string) ([]byte, error) {
	return os.ReadFile(path)
}

func (OSFiles) Glob(pattern string) ([]string, error) {
	return filepath.Glob(pattern)
}

func (OSFiles) MakeExecutable(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		return err
	}
	return os.Chmod(path, info.Mode().Perm()|executeBits)
}

func (g GitCLI) SetConfig(ctx context.Context, key, value string) error {
	out, err := exec.CommandContext(ctx, "git", "-C", g.Dir, "config", key, value).CombinedOutput()
	if err != nil {
		return fmt.Errorf("git config %s: %w: %s", key, err, strings.TrimSpace(string(out)))
	}
	return nil
}

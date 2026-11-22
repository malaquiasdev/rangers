package install

import (
	"context"
	"fmt"
	"io"
	"os/exec"
	"strings"
)

type ExecRunner struct {
	Out io.Writer
}

func (ExecRunner) Has(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}

func (ExecRunner) Output(ctx context.Context, name string, args ...string) (string, error) {
	out, err := exec.CommandContext(ctx, name, args...).CombinedOutput()
	if err != nil {
		detail := strings.TrimSpace(string(out))
		return string(out), fmt.Errorf("%s: %w: %s", words(name, args), err, detail)
	}
	return string(out), nil
}

func (e ExecRunner) Stream(ctx context.Context, name string, args ...string) error {
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Stdout = e.Out
	cmd.Stderr = e.Out
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("%s: %w", words(name, args), err)
	}
	return nil
}

func words(name string, args []string) string {
	return strings.Join(append([]string{name}, args...), " ")
}

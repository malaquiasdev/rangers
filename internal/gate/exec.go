package gate

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strings"
)

type ExecRunner struct{}

func (ExecRunner) Has(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}

func (ExecRunner) Output(ctx context.Context, name string, args ...string) (string, error) {
	out, err := exec.CommandContext(ctx, name, args...).Output()
	if err == nil {
		return string(out), nil
	}
	detail := ""
	if exit := new(exec.ExitError); errors.As(err, &exit) {
		detail = strings.TrimSpace(string(exit.Stderr))
	}
	command := strings.Join(append([]string{name}, args...), " ")
	return string(out), fmt.Errorf("%s: %w: %s", command, err, detail)
}

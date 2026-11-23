package check

import (
	"context"
	"io"
	"os"
	"os/exec"
)

type ExecRunner struct {
	Dir      string
	Out, Err io.Writer
}

func (r ExecRunner) Run(ctx context.Context, call Invocation) error {
	cmd := exec.CommandContext(ctx, call.Name, call.Args...)
	cmd.Dir = r.Dir
	cmd.Env = append(os.Environ(), call.Env...)
	if !call.Quiet {
		cmd.Stdout, cmd.Stderr = r.Out, r.Err
	}
	return cmd.Run()
}

func (r ExecRunner) Output(ctx context.Context, name string, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Dir = r.Dir
	cmd.Stderr = r.Err
	out, err := cmd.Output()
	return string(out), err
}

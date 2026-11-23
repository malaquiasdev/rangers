package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
	_ "time/tzdata"

	"github.com/malaquiasdev/rangers/internal/audit"
	"github.com/malaquiasdev/rangers/internal/check"
	"github.com/malaquiasdev/rangers/internal/githook"
)

func runAudit(ctx context.Context, args []string) int {
	dep := &audit.Dependencies{Store: audit.OSStore{}, Now: time.Now}
	return reportAudit(audit.Command(ctx, args, os.Getenv, dep))
}

func runCheck(ctx context.Context, args []string) int {
	root, err := repoRoot(ctx)
	if err != nil {
		return reportCheck(err)
	}
	return reportCheck(check.Command(ctx, args, os.Getenv, checkDependencies(root)))
}

func runHooks(ctx context.Context, args []string) int {
	root, err := repoRoot(ctx)
	if err != nil {
		return reportHooks(err)
	}
	dep := &githook.Dependencies{
		Check: checkDependencies(root),
		Files: githook.OSFiles{},
		Git:   githook.GitCLI{Dir: root},
		Out:   os.Stdout,
		Dir:   root,
	}
	return reportHooks(githook.Command(ctx, args, os.Getenv, dep))
}

func repoRoot(ctx context.Context) (string, error) {
	cmd := exec.CommandContext(ctx, "git", "rev-parse", "--show-toplevel")
	cmd.Stderr = os.Stderr
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("git rev-parse --show-toplevel: %w", err)
	}
	return strings.TrimSpace(string(out)), nil
}

func checkDependencies(root string) *check.Dependencies {
	return &check.Dependencies{
		Dir:      root,
		Root:     os.DirFS(root),
		Runner:   check.ExecRunner{Dir: root, Out: os.Stdout, Err: os.Stderr},
		ReadHost: os.ReadFile,
		Out:      os.Stdout,
	}
}

func reportAudit(err error) int {
	switch {
	case err == nil:
		return 0
	case errors.Is(err, audit.ErrUsage):
		fmt.Fprintln(os.Stderr, err)
		return 2
	}
	fmt.Fprintln(os.Stderr, "rangers audit:", err)
	return 1
}

func reportCheck(err error) int {
	switch {
	case err == nil:
		return 0
	case errors.Is(err, check.ErrFailed):
		return 1
	}
	fmt.Fprintln(os.Stderr, "rangers check:", err)
	return 1
}

func reportHooks(err error) int {
	switch {
	case err == nil:
		return 0
	case errors.Is(err, githook.ErrRefused), errors.Is(err, check.ErrFailed):
		return 1
	case errors.Is(err, githook.ErrUsage):
		fmt.Fprintln(os.Stderr, "rangers hooks:", err)
		return 2
	}
	fmt.Fprintln(os.Stderr, "rangers hooks:", err)
	return 1
}

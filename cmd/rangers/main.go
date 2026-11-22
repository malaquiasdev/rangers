package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"time"

	"github.com/malaquiasdev/rangers/internal/gate"
	"github.com/malaquiasdev/rangers/internal/install"
	"github.com/malaquiasdev/rangers/internal/learn"
	"github.com/malaquiasdev/rangers/internal/schedule"
	"github.com/malaquiasdev/rangers/internal/transcript"
)

const usage = "uso: rangers learn [metrics] [flags] | rangers schedule install|uninstall|status" +
	" | rangers install [--check] | rangers gate [ask] --repo owner/name --pr N [--merge]"

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	code := run(ctx, os.Args[1:])
	stop()
	os.Exit(code)
}

func run(ctx context.Context, args []string) int {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, usage)
		return 2
	}
	switch args[0] {
	case "learn":
		return report(args[0], learn.Command(ctx, args[1:], os.Getenv, learnDependencies()))
	case "install":
		return report(args[0], install.Command(ctx, args[1:], os.Getenv, installDependencies()))
	case "gate":
		return reportGate(gate.Command(ctx, args[1:], os.Getenv, gateDependencies()))
	case "schedule":
		return report(args[0], schedule.Command(ctx, args[1:], os.Getenv, scheduleDependencies()))
	}
	fmt.Fprintln(os.Stderr, usage)
	return 2
}

func report(name string, err error) int {
	if err == nil {
		return 0
	}
	if errors.Is(err, install.ErrIncomplete) {
		return 1
	}
	fmt.Fprintln(os.Stderr, "rangers "+name+":", err)
	return 1
}

func reportGate(err error) int {
	switch {
	case err == nil:
		return 0
	case errors.Is(err, gate.ErrRefused):
		return 1
	}
	fmt.Fprintln(os.Stderr, "rangers gate:", err)
	return 2
}

func gateDependencies() *gate.Dependencies {
	return &gate.Dependencies{
		Runner: gate.ExecRunner{},
		Lines:  transcript.Conversation,
		Out:    os.Stdout,
		Now:    time.Now,
	}
}

func learnDependencies() *learn.Dependencies {
	return &learn.Dependencies{
		Claude:   learn.ClaudeCLI{},
		Git:      learn.GitCLI{},
		Reminder: learn.OsascriptReminder{Timeout: 10 * time.Second},
		Now:      time.Now,
		Out:      os.Stdout,
		Err:      os.Stderr,
	}
}

func scheduleDependencies() *schedule.Dependencies {
	return &schedule.Dependencies{
		Launchctl: schedule.ExecLaunchctl{},
		UID:       os.Getuid(),
		Out:       os.Stdout,
	}
}

func installDependencies() *install.Dependencies {
	dir, err := os.Getwd()
	if err != nil {
		dir = "."
	}
	return &install.Dependencies{
		Runner: install.ExecRunner{Out: os.Stdout},
		Files:  install.OSFiles{},
		Out:    os.Stdout,
		Dir:    dir,
	}
}

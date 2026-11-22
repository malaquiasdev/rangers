package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"time"

	"github.com/malaquiasdev/rangers/internal/learn"
	"github.com/malaquiasdev/rangers/internal/schedule"
)

const usage = "uso: rangers learn [metrics] [flags] | rangers schedule install|uninstall|status"

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
	fmt.Fprintln(os.Stderr, "rangers "+name+":", err)
	return 1
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

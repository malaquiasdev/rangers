package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"slices"
	"time"

	"github.com/malaquiasdev/rangers/internal/agentwatch"
	"github.com/malaquiasdev/rangers/internal/gate"
	"github.com/malaquiasdev/rangers/internal/hook"
	"github.com/malaquiasdev/rangers/internal/install"
	"github.com/malaquiasdev/rangers/internal/learn"
	"github.com/malaquiasdev/rangers/internal/schedule"
	"github.com/malaquiasdev/rangers/internal/transcript"
)

const (
	usage = "uso: rangers learn [metrics] [flags] | rangers schedule install|uninstall|status" +
		" | rangers install [--check] | rangers gate [ask] --repo owner/name --pr N [--merge]" +
		" | rangers hook stop-push-check|complete-reminders | rangers watch [limit]" +
		" | rangers audit WHERE ACTION LINK WHY | rangers check [quick|full]" +
		" | rangers hooks install | rangers hooks run pre-commit|commit-msg FILE|pre-push"

	gateName     = "gate"
	hookName     = "hook"
	watchName    = "watch"
	auditName    = "audit"
	checkName    = "check"
	hooksName    = "hooks"
	stopHook     = "stop-push-check"
	completeHook = "complete-reminders"
)

var (
	binAliases = map[string][]string{
		stopHook:      {hookName, stopHook},
		completeHook:  {hookName, completeHook},
		"agent-watch": {watchName},
		auditName:     {auditName},
	}

	commands = map[string]func(context.Context, []string) int{
		"learn": func(ctx context.Context, args []string) int {
			return report("learn", learn.Command(ctx, args, os.Getenv, learnDependencies()))
		},
		"install": func(ctx context.Context, args []string) int {
			return report("install", install.Command(ctx, args, os.Getenv, installDependencies()))
		},
		gateName: func(ctx context.Context, args []string) int {
			return reportGate(gate.Command(ctx, args, os.Getenv, gateDependencies()))
		},
		"schedule": func(ctx context.Context, args []string) int {
			return report("schedule", schedule.Command(ctx, args, os.Getenv, scheduleDependencies()))
		},
		auditName: runAudit,
		checkName: runCheck,
		hooksName: runHooks,
		hookName: func(ctx context.Context, args []string) int {
			return reportHook(hook.Command(ctx, args, os.Getenv, hookDependencies()))
		},
		watchName: func(ctx context.Context, args []string) int {
			return reportWatch(agentwatch.Command(ctx, args, os.Getenv, watchDependencies()))
		},
	}
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	code := run(ctx, expand(os.Args))
	stop()
	os.Exit(code)
}

func expand(argv []string) []string {
	if len(argv) == 0 {
		return nil
	}
	prefix := binAliases[filepath.Base(argv[0])]
	return append(slices.Clone(prefix), argv[1:]...)
}

func run(ctx context.Context, args []string) int {
	if len(args) > 0 {
		if command, ok := commands[args[0]]; ok {
			return command(ctx, args[1:])
		}
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

func reportHook(err error) int {
	if err == nil {
		return 0
	}
	fmt.Fprintln(os.Stderr, "rangers hook:", err)
	return 2
}

func reportWatch(err error) int {
	switch {
	case err == nil:
		return 0
	case errors.Is(err, context.Canceled):
		return 130
	case errors.Is(err, agentwatch.ErrUsage):
		fmt.Fprintln(os.Stderr, "rangers watch:", err)
		return 2
	}
	fmt.Fprintln(os.Stderr, "rangers watch:", err)
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

func hookDependencies() *hook.Dependencies {
	return &hook.Dependencies{In: os.Stdin, Out: os.Stdout, Scripter: hook.OsascriptCLI{}}
}

func watchDependencies() *agentwatch.Dependencies {
	return &agentwatch.Dependencies{
		Git:   agentwatch.GitCLI{},
		Out:   os.Stdout,
		Now:   time.Now,
		Sleep: agentwatch.Sleep,
	}
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
	binary, err := os.Executable()
	if err != nil {
		binary = "rangers"
	}
	home, err := os.UserHomeDir()
	if err != nil {
		home = os.Getenv("HOME")
	}
	return &learn.Dependencies{
		Brain:    learn.DirBrain{Timeout: 10 * time.Second},
		Binary:   binary,
		Home:     home,
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
	binary, _ := os.Executable()
	return &install.Dependencies{
		Runner: install.ExecRunner{Out: os.Stdout},
		Files:  install.OSFiles{},
		Out:    os.Stdout,
		Dir:    dir,
		Binary: binary,
	}
}

package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"slices"
	"strings"
	"testing"

	"github.com/malaquiasdev/rangers/internal/agentwatch"
	"github.com/malaquiasdev/rangers/internal/gate"
	"github.com/malaquiasdev/rangers/internal/hook"
)

func TestMain(m *testing.M) {
	for _, pair := range os.Environ() {
		name, _, _ := strings.Cut(pair, "=")
		if strings.HasPrefix(name, "GIT_") {
			_ = os.Unsetenv(name)
		}
	}
	os.Exit(m.Run())
}

func TestReportGate(t *testing.T) {
	t.Run("given an allowed gate, when reporting, then the exit code is 0", func(t *testing.T) {
		if got := reportGate(nil); got != 0 {
			t.Fatalf("got %d", got)
		}
	})

	t.Run("given a refused gate, when reporting, then the exit code is 1", func(t *testing.T) {
		if got := reportGate(fmt.Errorf("verdict: %w", gate.ErrRefused)); got != 1 {
			t.Fatalf("got %d", got)
		}
	})

	t.Run("given an error, when reporting, then the exit code is 2", func(t *testing.T) {
		if got := reportGate(errors.New("gh: not found")); got != 2 {
			t.Fatalf("got %d", got)
		}
	})
}

func TestRunGate(t *testing.T) {
	t.Run("given gate with bad flags, when running, then the exit code is 2", func(t *testing.T) {
		if got := run(t.Context(), []string{"gate", "--pr", "0"}); got != 2 {
			t.Fatalf("got %d", got)
		}
	})
}

func TestExpand(t *testing.T) {
	tests := []struct {
		name string
		argv []string
		want []string
	}{
		{"given the program called as rangers, when expanding, then the arguments stay as they are", []string{"/usr/bin/rangers", "gate", "--pr", "1"}, []string{"gate", "--pr", "1"}},
		{"given the program called as stop-push-check, when expanding, then it runs the hook of that name", []string{"/home/x/.claude/bin/stop-push-check"}, []string{"hook", "stop-push-check"}},
		{"given the program called as complete-reminders, when expanding, then it runs the hook of that name", []string{"/home/x/.claude/bin/complete-reminders"}, []string{"hook", "complete-reminders"}},
		{"given the program called as agent-watch with a limit, when expanding, then it runs watch with that limit", []string{"/home/x/.claude/bin/agent-watch", "900"}, []string{"watch", "900"}},
		{"given the program called as agent-watch with no arguments, when expanding, then it runs watch", []string{"agent-watch"}, []string{"watch"}},
		{"given no arguments at all, when expanding, then there is nothing to run", nil, nil},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := expand(tc.argv); !slices.Equal(got, tc.want) {
				t.Fatalf("got %q, want %q", got, tc.want)
			}
		})
	}
}

func TestReportHook(t *testing.T) {
	t.Run("given a hook that ran, when reporting, then the exit code is 0", func(t *testing.T) {
		if got := reportHook(nil); got != 0 {
			t.Fatalf("got %d", got)
		}
	})

	t.Run("given a usage error, when reporting, then the exit code is 2", func(t *testing.T) {
		if got := reportHook(hook.ErrUsage); got != 2 {
			t.Fatalf("got %d", got)
		}
	})
}

func TestReportWatch(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want int
	}{
		{"given a watch that ended, when reporting, then the exit code is 0", nil, 0},
		{"given an interrupt, when reporting, then the exit code is 130", fmt.Errorf("sleep: %w", context.Canceled), 130},
		{"given a usage error, when reporting, then the exit code is 2", agentwatch.ErrUsage, 2},
		{"given any other error, when reporting, then the exit code is 1", errors.New("boom"), 1},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := reportWatch(tc.err); got != tc.want {
				t.Fatalf("got %d, want %d", got, tc.want)
			}
		})
	}
}

func TestRunHookAndWatch(t *testing.T) {
	t.Run("given hook with no name, when running, then the exit code is 2", func(t *testing.T) {
		if got := run(t.Context(), []string{"hook"}); got != 2 {
			t.Fatalf("got %d", got)
		}
	})

	t.Run("given watch with a bad limit, when running, then the exit code is 2", func(t *testing.T) {
		if got := run(t.Context(), []string{"watch", "soon"}); got != 2 {
			t.Fatalf("got %d", got)
		}
	})
}

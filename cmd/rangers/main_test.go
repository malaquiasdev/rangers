package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/malaquiasdev/rangers/internal/agentwatch"
	"github.com/malaquiasdev/rangers/internal/audit"
	"github.com/malaquiasdev/rangers/internal/check"
	"github.com/malaquiasdev/rangers/internal/gate"
	"github.com/malaquiasdev/rangers/internal/githook"
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
		{"given the program called as audit with its fields, when expanding, then it runs audit with them", []string{"/home/x/.claude/bin/audit", "repo", "merge", "link", "why"}, []string{"audit", "repo", "merge", "link", "why"}},
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

func TestReportAudit(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want int
	}{
		{"given an audit that wrote the row, when reporting, then the exit code is 0", nil, 0},
		{"given a usage error, when reporting, then the exit code is 2", audit.ErrUsage, 2},
		{"given any other error, when reporting, then the exit code is 1", errors.New("disk full"), 1},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := reportAudit(tc.err); got != tc.want {
				t.Fatalf("got %d, want %d", got, tc.want)
			}
		})
	}
}

func TestReportCheck(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want int
	}{
		{"given checks that passed, when reporting, then the exit code is 0", nil, 0},
		{"given a failed check, when reporting, then the exit code is 1", check.ErrFailed, 1},
		{"given any other error, when reporting, then the exit code is 1", errors.New("no names file"), 1},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := reportCheck(tc.err); got != tc.want {
				t.Fatalf("got %d, want %d", got, tc.want)
			}
		})
	}
}

func TestReportHooks(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want int
	}{
		{"given a hook that passed, when reporting, then the exit code is 0", nil, 0},
		{"given a refused message, when reporting, then the exit code is 1", githook.ErrRefused, 1},
		{"given a failed check, when reporting, then the exit code is 1", fmt.Errorf("pre-push: %w", check.ErrFailed), 1},
		{"given a usage error, when reporting, then the exit code is 2", githook.ErrUsage, 2},
		{"given any other error, when reporting, then the exit code is 1", errors.New("git: not found"), 1},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := reportHooks(tc.err); got != tc.want {
				t.Fatalf("got %d, want %d", got, tc.want)
			}
		})
	}
}

func TestRunAudit(t *testing.T) {
	t.Run("given four fields and a brain folder, when running audit, then the row is written and the exit code is 0", func(t *testing.T) {
		brain := t.TempDir()
		dir := filepath.Join(brain, "claude", "auditoria")
		if err := os.MkdirAll(dir, 0o750); err != nil {
			t.Fatal(err)
		}
		t.Setenv("BRAIN_DIR", brain)
		t.Setenv("CLAUDE_SESSION_ID", "session-1234")
		if got := run(t.Context(), []string{"audit", "repo", "merge", "https://example.test/1", "green"}); got != 0 {
			t.Fatalf("got %d", got)
		}
		notes, err := filepath.Glob(filepath.Join(dir, "20*.md"))
		if err != nil || len(notes) != 1 {
			t.Fatalf("got %v, %v", notes, err)
		}
		raw, err := os.ReadFile(notes[0])
		if err != nil || !strings.Contains(string(raw), " | session- | repo | merge | https://example.test/1 | green |\n") {
			t.Fatalf("got %q, %v", raw, err)
		}
	})

	t.Run("given too few fields, when running audit, then the exit code is 2", func(t *testing.T) {
		if got := run(t.Context(), []string{"audit", "repo"}); got != 2 {
			t.Fatalf("got %d", got)
		}
	})
}

func TestRunHooks(t *testing.T) {
	t.Run("given hooks with no name, when running, then the exit code is 2", func(t *testing.T) {
		if got := run(t.Context(), []string{"hooks"}); got != 2 {
			t.Fatalf("got %d", got)
		}
	})

	t.Run("given a good commit message, when running commit-msg, then the exit code is 0", func(t *testing.T) {
		message := filepath.Join(t.TempDir(), "COMMIT_EDITMSG")
		if err := os.WriteFile(message, []byte("feat: add a thing\n"), 0o600); err != nil {
			t.Fatal(err)
		}
		if got := run(t.Context(), []string{"hooks", "run", "commit-msg", message}); got != 0 {
			t.Fatalf("got %d", got)
		}
	})

	t.Run("given a bad commit message, when running commit-msg, then the exit code is 1", func(t *testing.T) {
		message := filepath.Join(t.TempDir(), "COMMIT_EDITMSG")
		if err := os.WriteFile(message, []byte("added a thing\n"), 0o600); err != nil {
			t.Fatal(err)
		}
		if got := run(t.Context(), []string{"hooks", "run", "commit-msg", message}); got != 1 {
			t.Fatalf("got %d", got)
		}
	})
}

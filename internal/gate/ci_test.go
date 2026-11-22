package gate

import (
	"errors"
	"strings"
	"testing"
)

func approvedScenario(t *testing.T) *scenario {
	t.Helper()
	s := newScenario(t, openPR())
	s.ask("s1", token)
	s.answer("s1", "sim")
	return s
}

func TestCI(t *testing.T) {
	t.Run("given every check passing or skipped, when running the dry run, then the gate allows and merges nothing", func(t *testing.T) {
		s := approvedScenario(t)
		if err := s.run(); err != nil {
			t.Fatal(err)
		}
		want := "pr: ok open, ready, base main\n" +
			"production: ok not a release\n" +
			"approval: ok the human said yes to " + token + "\n" +
			"ci: ok 3 checks passed or skipped\n" +
			"allowed: dry run, nothing merged\n"
		if s.printed() != want {
			t.Fatalf("got\n%s", s.printed())
		}
	})

	t.Run("given a failing check, when running the gate, then it refuses and lists the name", func(t *testing.T) {
		s := approvedScenario(t)
		s.runner.replies[checkCmd] = reply{out: `[{"name":"build","bucket":"pass"},{"name":"unit tests","bucket":"fail"}]`}
		if err := s.run(); !errors.Is(err, ErrRefused) {
			t.Fatalf("got %v", err)
		}
		if got := s.lastLine(); got != "refused: ci not green: unit tests (fail)" {
			t.Fatalf("got %q", got)
		}
	})

	t.Run("given a pending check and gh exiting with 8, when running the gate, then it refuses instead of failing", func(t *testing.T) {
		s := approvedScenario(t)
		s.runner.replies[checkCmd] = reply{out: `[{"name":"deploy","bucket":"pending"}]`, err: errors.New("exit status 8")}
		if err := s.run(); !errors.Is(err, ErrRefused) {
			t.Fatalf("got %v", err)
		}
		if got := s.lastLine(); got != "refused: ci not green: deploy (pending)" {
			t.Fatalf("got %q", got)
		}
	})

	t.Run("given a cancelled check, when running the gate, then it refuses", func(t *testing.T) {
		s := approvedScenario(t)
		s.runner.replies[checkCmd] = reply{out: `[{"name":"build","bucket":"cancel"}]`}
		if err := s.run(); !errors.Is(err, ErrRefused) {
			t.Fatalf("got %v", err)
		}
	})

	t.Run("given no checks, when running the gate, then it refuses", func(t *testing.T) {
		s := approvedScenario(t)
		s.runner.replies[checkCmd] = reply{out: `[]`}
		if err := s.run(); !errors.Is(err, ErrRefused) {
			t.Fatalf("got %v", err)
		}
		if got := s.lastLine(); got != "refused: ci has no checks for this head" {
			t.Fatalf("got %q", got)
		}
	})

	t.Run("given gh failing with no checks output, when running the gate, then it returns the error and does not refuse", func(t *testing.T) {
		s := approvedScenario(t)
		s.runner.replies[checkCmd] = reply{err: errors.New("gh: HTTP 502")}
		err := s.run()
		if err == nil || errors.Is(err, ErrRefused) || !strings.Contains(err.Error(), "HTTP 502") {
			t.Fatalf("got %v", err)
		}
	})

	t.Run("given no approval, when running the gate, then the checks are not even asked", func(t *testing.T) {
		s := newScenario(t, openPR())
		if err := s.run(); !errors.Is(err, ErrRefused) {
			t.Fatalf("got %v", err)
		}
		for _, call := range s.runner.calls {
			if call == checkCmd {
				t.Fatal("the checks were asked before the approval")
			}
		}
	})
}

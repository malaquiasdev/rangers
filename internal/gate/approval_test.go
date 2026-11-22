package gate

import (
	"errors"
	"strings"
	"testing"
	"time"
)

const refusedForToken = "refused: no yes from the human for " + token + "; ask with rangers gate ask"

func TestApproval(t *testing.T) {
	t.Run("given the human said sim to the question for this head, when running the dry run, then the approval passes", func(t *testing.T) {
		s := newScenario(t, openPR())
		s.ask("s1", token)
		s.answer("s1", "sim")
		if err := s.run(); err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(s.printed(), "approval: ok the human said yes to "+token) {
			t.Fatalf("got\n%s", s.printed())
		}
	})

	t.Run("given an approval for a different SHA, when running the gate, then it refuses", func(t *testing.T) {
		s := newScenario(t, openPR())
		s.ask("s1", other)
		s.answer("s1", "sim")
		if err := s.run(); !errors.Is(err, ErrRefused) {
			t.Fatalf("got %v", err)
		}
		if got := s.lastLine(); got != refusedForToken {
			t.Fatalf("got %q", got)
		}
	})

	t.Run("given a push after the question, when running the gate, then the old approval stops matching", func(t *testing.T) {
		s := newScenario(t, openPR())
		s.ask("s1", token)
		s.answer("s1", "sim")
		pushed := openPR()
		pushed.head = "ffffffffffffffffffffffffffffffffffffffff"
		s.runner.replies[viewCmd] = reply{out: pushed.json(t)}
		if err := s.run(); !errors.Is(err, ErrRefused) {
			t.Fatalf("got %v", err)
		}
		if !strings.Contains(s.lastLine(), "gate:2cd50664") {
			t.Fatalf("got %q", s.lastLine())
		}
	})

	t.Run("given a yes that came inside a system notification, when running the gate, then it is not an approval", func(t *testing.T) {
		s := newScenario(t, openPR())
		s.ask("s1", token)
		s.answer("s1", "<task-notification>sim</task-notification>")
		if err := s.run(); !errors.Is(err, ErrRefused) {
			t.Fatalf("got %v", err)
		}
		if got := s.lastLine(); got != refusedForToken {
			t.Fatalf("got %q", got)
		}
	})

	t.Run("given a yes that came inside a tool result, when running the gate, then it is not an approval", func(t *testing.T) {
		s := newScenario(t, openPR())
		s.ask("s1", token)
		s.toolResult("s1", "sim")
		if err := s.run(); !errors.Is(err, ErrRefused) {
			t.Fatalf("got %v", err)
		}
	})

	t.Run("given a system notification between the question and the yes, when running the gate, then the human's yes still counts", func(t *testing.T) {
		s := newScenario(t, openPR())
		s.ask("s1", token)
		s.answer("s1", "<system-reminder>background task finished</system-reminder>")
		s.answer("s1", "Sim!")
		if err := s.run(); err != nil {
			t.Fatalf("got %v", err)
		}
	})

	t.Run("given a no after the question and a yes later, when running the gate, then it refuses", func(t *testing.T) {
		s := newScenario(t, openPR())
		s.ask("s1", token)
		s.answer("s1", "não")
		s.answer("s1", "sim")
		if err := s.run(); !errors.Is(err, ErrRefused) {
			t.Fatalf("got %v", err)
		}
	})

	t.Run("given a yes in another session than the question, when running the gate, then it refuses", func(t *testing.T) {
		s := newScenario(t, openPR())
		s.ask("s1", token)
		s.answer("s2", "sim")
		if err := s.run(); !errors.Is(err, ErrRefused) {
			t.Fatalf("got %v", err)
		}
	})

	t.Run("given no question at all, when running the gate, then it refuses and points to rangers gate ask", func(t *testing.T) {
		s := newScenario(t, openPR())
		if err := s.run(); !errors.Is(err, ErrRefused) {
			t.Fatalf("got %v", err)
		}
		if got := s.lastLine(); got != refusedForToken {
			t.Fatalf("got %q", got)
		}
	})

	t.Run("given a valid approval on a release pull request, when running the gate, then it still stops at production", func(t *testing.T) {
		pr := openPR()
		pr.title = "chore(main): release 1.4.0"
		s := newScenario(t, pr)
		s.ask("s1", token)
		s.answer("s1", "sim")
		if err := s.run(); !errors.Is(err, ErrRefused) {
			t.Fatalf("got %v", err)
		}
		if got := s.lastLine(); got != "refused: "+productionVerdict {
			t.Fatalf("got %q", got)
		}
		if s.reads != 0 {
			t.Fatal("the transcripts were read before the production stop")
		}
	})

	t.Run("given a run, when reading the transcripts, then it looks back 48 hours", func(t *testing.T) {
		s := newScenario(t, openPR())
		_ = s.run()
		if want := clock().Add(-48 * time.Hour); !s.since.Equal(want) {
			t.Fatalf("got %v, want %v", s.since, want)
		}
	})
}

func TestAffirmative(t *testing.T) {
	yes := []string{"yes", "YES", "Sim", "sim.", "ok!", " pode ", "Pode, ", "OK ."}
	for _, text := range yes {
		t.Run("given "+text+", when checking the answer, then it is a yes", func(t *testing.T) {
			if !affirmative(text) {
				t.Fatal("expected a yes")
			}
		})
	}
	no := []string{"", "no", "não", "yes please", "sim, mas espere", "okay", "...sim", "talvez"}
	for _, text := range no {
		t.Run("given "+text+", when checking the answer, then it is not a yes", func(t *testing.T) {
			if affirmative(text) {
				t.Fatal("expected no")
			}
		})
	}
}

func TestApprovalHash(t *testing.T) {
	t.Run("given a repository, a number and a head, when hashing, then it returns the first 8 hex of the sha256", func(t *testing.T) {
		if got := approvalHash(repo, 7, headSHA); got != "4a7010e4" {
			t.Fatalf("got %q", got)
		}
	})

	t.Run("given another head, when hashing, then the hash changes", func(t *testing.T) {
		if got := approvalHash(repo, 7, "ffffffffffffffffffffffffffffffffffffffff"); got != "2cd50664" {
			t.Fatalf("got %q", got)
		}
	})
}

func TestAsk(t *testing.T) {
	t.Run("given an open pull request, when asking, then it prints the hash and the line to paste", func(t *testing.T) {
		s := newScenario(t, openPR())
		if err := s.runAsk(); err != nil {
			t.Fatal(err)
		}
		want := "4a7010e4\nMerge acme/api#7 at 0123456? Answer sim. gate:4a7010e4\n"
		if s.printed() != want {
			t.Fatalf("got %q", s.printed())
		}
	})
}

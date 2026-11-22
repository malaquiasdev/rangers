package gate

import (
	"errors"
	"strings"
	"testing"
)

func TestMerge(t *testing.T) {
	t.Run("given a valid approval and green CI, when running the dry run, then merge is never called", func(t *testing.T) {
		s := approvedScenario(t)
		if err := s.run(); err != nil {
			t.Fatal(err)
		}
		for _, call := range s.runner.calls {
			if strings.HasPrefix(call, "gh pr merge") {
				t.Fatalf("merge was called: %s", call)
			}
		}
		if got := s.lastLine(); got != "allowed: dry run, nothing merged" {
			t.Fatalf("got %q", got)
		}
	})

	t.Run("given a valid approval and green CI, when running with --merge, then it merges once pinned to the head SHA", func(t *testing.T) {
		s := approvedScenario(t)
		if err := s.run("--merge"); err != nil {
			t.Fatal(err)
		}
		if n := s.runner.called(mergeCmd); n != 1 {
			t.Fatalf("merge called %d times: %v", n, s.runner.calls)
		}
		if !strings.Contains(s.printed(), "merge: ok squash at 0123456\n") || s.lastLine() != "allowed: merged" {
			t.Fatalf("got\n%s", s.printed())
		}
	})

	t.Run("given the audit script is missing, when merging, then it does not call it", func(t *testing.T) {
		s := approvedScenario(t)
		if err := s.run("--merge"); err != nil {
			t.Fatal(err)
		}
		for _, call := range s.runner.calls {
			if strings.HasPrefix(call, auditBin) {
				t.Fatalf("audit was called: %s", call)
			}
		}
	})

	t.Run("given the audit script exists, when merging, then it appends one audit line after the merge", func(t *testing.T) {
		s := approvedScenario(t)
		s.runner.present[auditBin] = true
		s.runner.replies[auditCmd] = reply{}
		if err := s.run("--merge"); err != nil {
			t.Fatal(err)
		}
		if n := s.runner.called(auditCmd); n != 1 {
			t.Fatalf("audit called %d times: %v", n, s.runner.calls)
		}
		calls := s.runner.calls
		if calls[len(calls)-2] != mergeCmd || calls[len(calls)-1] != auditCmd {
			t.Fatalf("got %v", calls)
		}
		if !strings.Contains(s.printed(), "merge: ok squash at 0123456; audit recorded\n") {
			t.Fatalf("got\n%s", s.printed())
		}
	})

	t.Run("given the audit script failing, when merging, then the merge still counts and the line says it was not recorded", func(t *testing.T) {
		s := approvedScenario(t)
		s.runner.present[auditBin] = true
		s.runner.replies[auditCmd] = reply{err: errors.New("disk full")}
		if err := s.run("--merge"); err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(s.printed(), "merge: ok squash at 0123456; audit not recorded\n") {
			t.Fatalf("got\n%s", s.printed())
		}
	})

	t.Run("given gh refusing the merge because the head moved, when merging, then it returns the error and does not audit", func(t *testing.T) {
		s := approvedScenario(t)
		s.runner.present[auditBin] = true
		s.runner.replies[mergeCmd] = reply{err: errors.New("head branch was modified")}
		err := s.run("--merge")
		if err == nil || errors.Is(err, ErrRefused) {
			t.Fatalf("got %v", err)
		}
		if n := s.runner.called(auditCmd); n != 0 {
			t.Fatal("the audit ran after a failed merge")
		}
	})

	t.Run("given failing CI, when running with --merge, then merge is never called", func(t *testing.T) {
		s := approvedScenario(t)
		s.runner.replies[checkCmd] = reply{out: `[{"name":"build","bucket":"fail"}]`}
		if err := s.run("--merge"); !errors.Is(err, ErrRefused) {
			t.Fatalf("got %v", err)
		}
		if n := s.runner.called(mergeCmd); n != 0 {
			t.Fatal("merge was called")
		}
	})

	t.Run("given no approval, when running with --merge, then merge is never called", func(t *testing.T) {
		s := newScenario(t, openPR())
		if err := s.run("--merge"); !errors.Is(err, ErrRefused) {
			t.Fatalf("got %v", err)
		}
		if n := s.runner.called(mergeCmd); n != 0 {
			t.Fatal("merge was called")
		}
	})

	t.Run("given a release pull request with approval and green CI, when running with --merge, then merge is never called", func(t *testing.T) {
		pr := openPR()
		pr.labels = []string{"production"}
		s := newScenario(t, pr)
		s.ask("s1", token)
		s.answer("s1", "sim")
		if err := s.run("--merge"); !errors.Is(err, ErrRefused) {
			t.Fatalf("got %v", err)
		}
		if n := s.runner.called(mergeCmd); n != 0 {
			t.Fatal("merge was called")
		}
	})
}

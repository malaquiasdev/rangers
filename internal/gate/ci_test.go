package gate

import (
	"errors"
	"strings"
	"testing"
)

const (
	pullRequestWorkflow = "name: ci\non:\n  pull_request:\n    branches: [main]\njobs:\n  test:\n    runs-on: ubuntu-latest\n    steps:\n      - run: make check\n"
	releaseWorkflow     = "name: release\non:\n  push:\n    branches: [main]\njobs:\n  release:\n    runs-on: ubuntu-latest\n    steps:\n      - run: make release\n"
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
		s.sameWorkflows(map[string]string{"ci.yml": pullRequestWorkflow})
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

	t.Run("given gh reporting no checks with empty output and exit 1, when running the gate, then it refuses with exit 1 and no merge call even with --merge", func(t *testing.T) {
		s := approvedScenario(t)
		s.runner.replies[checkCmd] = reply{err: errors.New(checkCmd + ": exit status 1: " + noChecks)}
		s.sameWorkflows(map[string]string{"ci.yml": pullRequestWorkflow})
		if err := s.run("--merge"); !errors.Is(err, ErrRefused) {
			t.Fatalf("got %v", err)
		}
		if got := s.lastLine(); got != "refused: ci has no checks for this head" {
			t.Fatalf("got %q", got)
		}
		if n := s.runner.called(mergeCmd); n != 0 {
			t.Fatalf("merge called %d times: %v", n, s.runner.calls)
		}
	})

	t.Run("given an unrelated gh failure with empty output, when running the gate, then it returns the error and does not refuse", func(t *testing.T) {
		s := approvedScenario(t)
		s.runner.replies[checkCmd] = reply{err: errors.New(checkCmd + ": exit status 1: could not resolve to a PullRequest with the number of 7")}
		err := s.run("--merge")
		if err == nil || errors.Is(err, ErrRefused) || !strings.Contains(err.Error(), "could not resolve") {
			t.Fatalf("got %v", err)
		}
		if n := s.runner.called(mergeCmd); n != 0 {
			t.Fatalf("merge called %d times: %v", n, s.runner.calls)
		}
	})

	t.Run("given malformed JSON from gh, when running the gate, then it returns the error and does not refuse", func(t *testing.T) {
		s := approvedScenario(t)
		s.runner.replies[checkCmd] = reply{out: `{"name":`}
		err := s.run("--merge")
		if err == nil || errors.Is(err, ErrRefused) {
			t.Fatalf("got %v", err)
		}
		if n := s.runner.called(mergeCmd); n != 0 {
			t.Fatalf("merge called %d times: %v", n, s.runner.calls)
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

	t.Run("given no workflow runs on pull requests and gh reporting no checks, when running the gate, then the ci is not applicable and the gate allows", func(t *testing.T) {
		s := approvedScenario(t)
		s.runner.replies[checkCmd] = reply{err: errors.New(checkCmd + ": exit status 1: " + noChecks)}
		s.sameWorkflows(map[string]string{"release.yml": releaseWorkflow, "notes.txt": "not a workflow"})
		if err := s.run(); err != nil {
			t.Fatal(err)
		}
		want := "pr: ok open, ready, base main\n" +
			"production: ok not a release\n" +
			"approval: ok the human said yes to " + token + "\n" +
			"ci: ok not applicable, no workflow runs on pull requests\n" +
			"allowed: dry run, nothing merged\n"
		if s.printed() != want {
			t.Fatalf("got\n%s", s.printed())
		}
	})

	t.Run("given no workflows folder at either ref and an empty check list, when running the gate, then the ci is not applicable", func(t *testing.T) {
		s := approvedScenario(t)
		s.runner.replies[checkCmd] = reply{out: `[]`}
		s.noWorkflowsAt("main")
		s.noWorkflowsAt(headSHA)
		if err := s.run(); err != nil {
			t.Fatal(err)
		}
		if got := strings.Split(strings.TrimSpace(s.printed()), "\n")[3]; got != "ci: ok not applicable, no workflow runs on pull requests" {
			t.Fatalf("got %q", got)
		}
	})

	t.Run("given a workflow that runs on pull requests and no checks, when running the gate, then it refuses even with --merge", func(t *testing.T) {
		triggers := map[string]string{
			"pull_request":         "on:\n  pull_request:\n",
			"pull_request_target":  "on: pull_request_target\n",
			"a list of events":     "on: [workflow_dispatch, pull_request]\n",
			"push without filter":  "on: push\n",
			"push with null value": "on:\n  push:\n",
			"push ignoring":        "on:\n  push:\n    branches-ignore: [docs]\n",
			"push on a glob":       "on:\n  push:\n    branches: ['feat/**']\n",
			"push on the head":     "on:\n  push:\n    branches: [feat/endpoint]\n",
			"unparseable yaml":     "on: [unclosed\n",
		}
		for name, body := range triggers {
			t.Run(name, func(t *testing.T) {
				s := approvedScenario(t)
				s.runner.replies[checkCmd] = reply{err: errors.New(checkCmd + ": exit status 1: " + noChecks)}
				s.sameWorkflows(map[string]string{"ci.yaml": body})
				if err := s.run("--merge"); !errors.Is(err, ErrRefused) {
					t.Fatalf("got %v", err)
				}
				if got := s.lastLine(); got != "refused: ci has no checks for this head" {
					t.Fatalf("got %q", got)
				}
				if n := s.runner.called(mergeCmd); n != 0 {
					t.Fatalf("merge called %d times: %v", n, s.runner.calls)
				}
			})
		}
	})

	t.Run("given workflows that only run on other events or other branches and no checks, when running the gate, then the ci is not applicable", func(t *testing.T) {
		quiet := map[string]string{
			"push to main":      "on:\n  push:\n    branches: [main]\n",
			"schedule":          "on:\n  schedule:\n    - cron: '0 3 * * *'\n",
			"manual":            "on: workflow_dispatch\n",
			"release published": "on:\n  release:\n    types: [published]\n",
		}
		for name, body := range quiet {
			t.Run(name, func(t *testing.T) {
				s := approvedScenario(t)
				s.runner.replies[checkCmd] = reply{out: `[]`}
				s.sameWorkflows(map[string]string{"quiet.yml": body})
				if err := s.run(); err != nil {
					t.Fatal(err)
				}
				if got := s.lastLine(); !strings.HasPrefix(got, "allowed:") {
					t.Fatalf("got %q", got)
				}
			})
		}
	})

	t.Run("given the head deleted the only workflow that runs on pull requests, when running the gate, then it refuses", func(t *testing.T) {
		s := approvedScenario(t)
		s.runner.replies[checkCmd] = reply{err: errors.New(checkCmd + ": exit status 1: " + noChecks)}
		s.workflowsAt("main", map[string]string{"ci.yml": pullRequestWorkflow})
		s.noWorkflowsAt(headSHA)
		if err := s.run("--merge"); !errors.Is(err, ErrRefused) {
			t.Fatalf("got %v", err)
		}
		if got := s.lastLine(); got != "refused: ci has no checks for this head" {
			t.Fatalf("got %q", got)
		}
		if n := s.runner.called(mergeCmd); n != 0 {
			t.Fatalf("merge called %d times: %v", n, s.runner.calls)
		}
	})

	t.Run("given the head added a workflow that runs on pull requests, when running the gate, then it refuses", func(t *testing.T) {
		s := approvedScenario(t)
		s.runner.replies[checkCmd] = reply{out: `[]`}
		s.noWorkflowsAt("main")
		s.workflowsAt(headSHA, map[string]string{"ci.yml": pullRequestWorkflow})
		if err := s.run(); !errors.Is(err, ErrRefused) {
			t.Fatalf("got %v", err)
		}
	})

	t.Run("given the workflows listing failing, when running the gate with --merge, then it returns the error and makes no merge call", func(t *testing.T) {
		for _, failing := range []string{"main", headSHA} {
			t.Run(failing, func(t *testing.T) {
				s := approvedScenario(t)
				s.runner.replies[checkCmd] = reply{err: errors.New(checkCmd + ": exit status 1: " + noChecks)}
				s.sameWorkflows(map[string]string{"release.yml": releaseWorkflow})
				s.workflowsFail(failing)
				err := s.run("--merge")
				if err == nil || errors.Is(err, ErrRefused) || !strings.Contains(err.Error(), "HTTP 502") {
					t.Fatalf("got %v", err)
				}
				if n := s.runner.called(mergeCmd); n != 0 {
					t.Fatalf("merge called %d times: %v", n, s.runner.calls)
				}
			})
		}
	})

	t.Run("given a workflow file that cannot be read, when running the gate with --merge, then it returns the error and makes no merge call", func(t *testing.T) {
		s := approvedScenario(t)
		s.runner.replies[checkCmd] = reply{out: `[]`}
		s.sameWorkflows(map[string]string{"release.yml": releaseWorkflow})
		s.runner.replies[fileCmd("main", "release.yml")] = reply{err: errors.New("gh: Bad Gateway (HTTP 502)")}
		err := s.run("--merge")
		if err == nil || errors.Is(err, ErrRefused) {
			t.Fatalf("got %v", err)
		}
		if n := s.runner.called(mergeCmd); n != 0 {
			t.Fatalf("merge called %d times: %v", n, s.runner.calls)
		}
	})

	t.Run("given checks that exist, when running the gate, then the workflows are not read", func(t *testing.T) {
		s := approvedScenario(t)
		if err := s.run(); err != nil {
			t.Fatal(err)
		}
		for _, call := range s.runner.calls {
			if strings.Contains(call, workflowsPath) {
				t.Fatalf("the workflows were read: %s", call)
			}
		}
	})

	t.Run("given no checks and no workflow on pull requests, when merging, then the audit reason says the ci was not applicable", func(t *testing.T) {
		s := approvedScenario(t)
		s.runner.replies[checkCmd] = reply{out: `[]`}
		s.sameWorkflows(map[string]string{"release.yml": releaseWorkflow})
		s.runner.present[auditBin] = true
		notApplicable := strings.Replace(auditCmd, "checks green", "ci not applicable", 1)
		s.runner.replies[notApplicable] = reply{}
		if err := s.run("--merge"); err != nil {
			t.Fatal(err)
		}
		if n := s.runner.called(notApplicable); n != 1 {
			t.Fatalf("audit called %d times: %v", n, s.runner.calls)
		}
		if n := s.runner.called(mergeCmd); n != 1 {
			t.Fatalf("merge called %d times: %v", n, s.runner.calls)
		}
	})
}

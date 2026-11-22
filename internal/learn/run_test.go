package learn

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

const claudeJSON = `{"type":"result","is_error":false,"duration_ms":5400,"num_turns":3,"total_cost_usd":0.12345,"usage":{"input_tokens":100,"cache_creation_input_tokens":500,"cache_read_input_tokens":1000,"output_tokens":50},"result":"proposed rule"}`

const learningAgent = "---\nname: learner\ndescription: Learning. Reads the transcripts and proposes rules.\nmodel: opus\n---\n\nbody\n"

const otherAgent = "---\nname: other\ndescription: Author. Implements one unit.\nmodel: sonnet\n---\n\nbody\n"

type fakeReminder struct {
	calls [][2]string
	fail  error
}

func (f *fakeReminder) Create(title, body string) error {
	f.calls = append(f.calls, [2]string{title, body})
	return f.fail
}

type scenario struct {
	t        *testing.T
	cfg      Config
	now      time.Time
	log      string
	reminder *fakeReminder
	out      *bytes.Buffer
}

func testGitEnv() []string {
	var kept []string
	for _, entry := range os.Environ() {
		if !strings.HasPrefix(entry, "GIT_") {
			kept = append(kept, entry)
		}
	}
	return append(kept, "GIT_CONFIG_NOSYSTEM=1")
}

func execute(t *testing.T, dir, name string, args ...string) {
	t.Helper()
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	cmd.Env = testGitEnv()
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("%s %v: %v\n%s", name, args, err, out)
	}
}

func newScenario(t *testing.T) *scenario {
	t.Helper()
	root := t.TempDir()
	t.Setenv("HOME", filepath.Join(root, "home"))
	t.Setenv("GIT_CONFIG_NOSYSTEM", "1")
	repo := filepath.Join(root, "rangers")
	if err := os.MkdirAll(filepath.Join(repo, "agents"), 0o750); err != nil {
		t.Fatal(err)
	}
	for name, body := range map[string]string{"learner.md": learningAgent, "other.md": otherAgent} {
		if err := os.WriteFile(filepath.Join(repo, "agents", name), []byte(body), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	execute(t, repo, "git", "init", "-q", "-b", "main")
	execute(t, repo, "git", "add", "-A")
	execute(t, repo, "git", "-c", "user.name=t", "-c", "user.email=t@t.t", "commit", "-q", "-m", "initial")

	s := &scenario{
		t:        t,
		now:      time.Date(2026, 11, 22, 10, 30, 0, 0, time.Local),
		log:      filepath.Join(root, "claude.log"),
		reminder: &fakeReminder{},
		out:      &bytes.Buffer{},
		cfg: Config{
			IntervalHours:  2,
			MinLines:       10,
			BudgetUSD:      3,
			DailyBudgetUSD: 6,
			Model:          "opus",
			WindowHours:    24,
			State:          filepath.Join(root, "state"),
			Repo:           repo,
			Projects:       filepath.Join(root, "projects"),
		},
	}
	s.fakeClaude(claudeJSON, 0)
	return s
}

func (s *scenario) fakeClaude(reply string, code int) {
	s.t.Helper()
	bin := filepath.Join(s.t.TempDir(), "bin")
	if err := os.MkdirAll(bin, 0o750); err != nil {
		s.t.Fatal(err)
	}
	script := fmt.Sprintf("#!/bin/sh\npwd -P > %q\nfor a in \"$@\"; do printf '%%s\\n' \"$a\" >> %q; done\nprintf '%%s' '%s'\nexit %d\n", s.log, s.log, reply, code)
	if err := os.WriteFile(filepath.Join(bin, "claude"), []byte(script), 0o500); err != nil {
		s.t.Fatal(err)
	}
	s.t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))
}

func (s *scenario) lines(file string, n int, at, mtime time.Time) {
	s.t.Helper()
	target := filepath.Join(s.cfg.Projects, "proj", file+".jsonl")
	if err := os.MkdirAll(filepath.Dir(target), 0o750); err != nil {
		s.t.Fatal(err)
	}
	var rows []string
	for i := range n {
		raw, err := json.Marshal(map[string]any{
			"type":      "user",
			"timestamp": at.UTC().Format("2006-01-02T15:04:05.000Z"),
			"message":   map[string]any{"role": "user", "content": fmt.Sprintf("line %d of %s", i, file)},
		})
		if err != nil {
			s.t.Fatal(err)
		}
		rows = append(rows, string(raw))
	}
	if err := os.WriteFile(target, []byte(strings.Join(rows, "\n")+"\n"), 0o600); err != nil {
		s.t.Fatal(err)
	}
	if err := os.Chtimes(target, mtime, mtime); err != nil {
		s.t.Fatal(err)
	}
}

func (s *scenario) checkpointAgo(ago time.Duration) {
	s.t.Helper()
	if err := os.MkdirAll(s.cfg.State, 0o750); err != nil {
		s.t.Fatal(err)
	}
	path := filepath.Join(s.cfg.State, "checkpoint")
	if err := os.WriteFile(path, nil, 0o600); err != nil {
		s.t.Fatal(err)
	}
	at := s.now.Add(-ago)
	if err := os.Chtimes(path, at, at); err != nil {
		s.t.Fatal(err)
	}
}

func (s *scenario) run() error {
	return Run(context.Background(), &s.cfg, Dependencies{
		Claude:   ClaudeCLI{},
		Git:      GitCLI{},
		Reminder: s.reminder,
		Now:      func() time.Time { return s.now },
		Out:      s.out,
	})
}

func (s *scenario) claudeArguments() []string {
	s.t.Helper()
	raw, err := os.ReadFile(s.log)
	if err != nil {
		s.t.Fatalf("claude was not called: %v", err)
	}
	return strings.Split(strings.TrimSpace(string(raw)), "\n")
}

func (s *scenario) claudeNotCalled() bool {
	_, err := os.Stat(s.log)
	return errors.Is(err, os.ErrNotExist)
}

func (s *scenario) branches() string {
	s.t.Helper()
	cmd := exec.Command("git", "-C", s.cfg.Repo, "branch", "--list", "learn/*")
	cmd.Env = testGitEnv()
	out, err := cmd.Output()
	if err != nil {
		s.t.Fatal(err)
	}
	return strings.TrimSpace(string(out))
}

func (s *scenario) stateFile(name string) string {
	s.t.Helper()
	raw, err := os.ReadFile(filepath.Join(s.cfg.State, name))
	if err != nil {
		s.t.Fatal(err)
	}
	return string(raw)
}

func (s *scenario) stateHas(name string) bool {
	_, err := os.Stat(filepath.Join(s.cfg.State, name))
	return err == nil
}

func (s *scenario) metricsRows(rows ...string) {
	s.t.Helper()
	if err := os.MkdirAll(s.cfg.State, 0o750); err != nil {
		s.t.Fatal(err)
	}
	content := strings.Join(rows, "\n")
	if content != "" {
		content += "\n"
	}
	if err := os.WriteFile(filepath.Join(s.cfg.State, "metrics.tsv"), []byte(content), 0o600); err != nil {
		s.t.Fatal(err)
	}
}

func costRow(stamp, cost string) string {
	return stamp + "\t12\t5\t3\t" + cost + "\t1600\t50"
}

func (s *scenario) fresh(n int) {
	s.lines("new-session", n, s.now.Add(-time.Hour), s.now.Add(-time.Hour))
}

func TestRunExecutesTheModel(t *testing.T) {
	t.Run("given enough lines and no checkpoint, when running, then it creates the branch and worktree, calls claude inside it and records the result", func(t *testing.T) {
		s := newScenario(t)
		s.fresh(12)
		if err := s.run(); err != nil {
			t.Fatal(err)
		}

		worktree := filepath.Join(filepath.Dir(s.cfg.Repo), "rangers-worktrees", "learn-2026-11-22-1030")
		if _, err := os.Stat(worktree); err != nil {
			t.Fatalf("the worktree does not exist: %v", err)
		}
		if branches := s.branches(); !strings.Contains(branches, "learn/2026-11-22-1030") {
			t.Fatalf("the branch does not exist: %q", branches)
		}

		input := filepath.Join(s.cfg.State, "input-2026-11-22-1030.txt")
		arguments := s.claudeArguments()
		resolved, err := filepath.EvalSymlinks(worktree)
		if err != nil {
			t.Fatal(err)
		}
		if arguments[0] != resolved {
			t.Fatalf("claude ran in %q, expected %q", arguments[0], resolved)
		}
		args := arguments[1:]
		if args[0] != "-p" || !strings.Contains(args[1], input) || !strings.Contains(args[1], "learn/2026-11-22-1030") {
			t.Fatalf("unexpected prompt: %q", args[:2])
		}
		rest := strings.Join(args[2:], "|")
		for _, piece := range []string{
			"--agent|learner",
			"--permission-mode|acceptEdits",
			"--allowedTools|Read,Edit,Write,Bash(git add:*),Bash(git commit:*),Bash(git status:*),Bash(git diff:*)",
			"--model|opus",
			"--max-budget-usd|3",
			"--output-format|json",
		} {
			if !strings.Contains(rest, piece) {
				t.Errorf("missing %q in %q", piece, rest)
			}
		}

		if got := s.out.String(); !strings.Contains(got, "proposed rule\n") || !strings.Contains(got, "ramo learn/2026-11-22-1030 em "+worktree) {
			t.Fatalf("unexpected output: %q", got)
		}
		if rows := strings.Count(s.stateFile("input-2026-11-22-1030.txt"), "\n"); rows != 12 {
			t.Fatalf("the input has %d rows", rows)
		}
		if raw := s.stateFile("output-2026-11-22-1030.json"); raw != claudeJSON {
			t.Fatalf("the stored output differs: %q", raw)
		}
	})

	t.Run("given a finished round, when it ends, then it appends the metrics row and updates the checkpoint", func(t *testing.T) {
		s := newScenario(t)
		s.fresh(12)
		if err := s.run(); err != nil {
			t.Fatal(err)
		}
		expected := "2026-11-22-1030\t12\t5\t3\t0.123\t1600\t50\n"
		if got := s.stateFile("metrics.tsv"); got != expected {
			t.Fatalf("expected %q, got %q", expected, got)
		}
		info, err := os.Stat(filepath.Join(s.cfg.State, "checkpoint"))
		if err != nil {
			t.Fatal(err)
		}
		if !info.ModTime().Equal(s.now) {
			t.Fatalf("checkpoint at %v, expected %v", info.ModTime(), s.now)
		}
	})

	t.Run("given a finished round, when it ends, then it creates a reminder with the branch and the worktree", func(t *testing.T) {
		s := newScenario(t)
		s.fresh(12)
		if err := s.run(); err != nil {
			t.Fatal(err)
		}
		if len(s.reminder.calls) != 1 {
			t.Fatalf("expected 1 reminder, got %d", len(s.reminder.calls))
		}
		title, body := s.reminder.calls[0][0], s.reminder.calls[0][1]
		if title != "Revisar as regras propostas" || !strings.Contains(body, "learn/2026-11-22-1030") || !strings.Contains(body, "learn-2026-11-22-1030") {
			t.Fatalf("unexpected reminder: %q / %q", title, body)
		}
	})

	t.Run("given a configured budget and model, when running, then it passes both to claude", func(t *testing.T) {
		s := newScenario(t)
		s.cfg.BudgetUSD = 1.5
		s.cfg.Model = "sonnet"
		s.fresh(12)
		if err := s.run(); err != nil {
			t.Fatal(err)
		}
		rest := strings.Join(s.claudeArguments(), "|")
		if !strings.Contains(rest, "--max-budget-usd|1.5") || !strings.Contains(rest, "--model|sonnet") {
			t.Fatalf("arguments: %q", rest)
		}
	})

	t.Run("given several agents, when running, then it uses the one whose description starts with Learning.", func(t *testing.T) {
		s := newScenario(t)
		s.fresh(12)
		if err := s.run(); err != nil {
			t.Fatal(err)
		}
		if rest := strings.Join(s.claudeArguments(), "|"); !strings.Contains(rest, "--agent|learner|") {
			t.Fatalf("arguments: %q", rest)
		}
	})
}

func TestRunGates(t *testing.T) {
	t.Run("given a checkpoint 30 minutes ago, when running, then it prints the reason and does not call the model", func(t *testing.T) {
		s := newScenario(t)
		s.fresh(12)
		s.checkpointAgo(30 * time.Minute)
		if err := s.run(); err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(s.out.String(), "a última rodada foi há 30 min, menos que o intervalo de 2h") {
			t.Fatalf("output: %q", s.out.String())
		}
		if !s.claudeNotCalled() || s.branches() != "" || len(s.reminder.calls) != 0 || s.stateHas("metrics.tsv") {
			t.Fatal("the round should have stopped before any effect")
		}
	})

	t.Run("given fewer lines than the minimum, when running, then it prints the reason, drops the input and does not call the model", func(t *testing.T) {
		s := newScenario(t)
		s.fresh(9)
		if err := s.run(); err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(s.out.String(), "só 9 falas novas, abaixo do mínimo de 10") {
			t.Fatalf("output: %q", s.out.String())
		}
		if !s.claudeNotCalled() || s.branches() != "" || s.stateHas("input-2026-11-22-1030.txt") {
			t.Fatal("the round should have stopped without leaving the input")
		}
	})

	t.Run("given a checkpoint 3 hours ago, when running, then only the lines after it enter", func(t *testing.T) {
		s := newScenario(t)
		s.lines("mixed", 10, s.now.Add(-4*time.Hour), s.now.Add(-time.Hour))
		s.lines("later", 10, s.now.Add(-time.Hour), s.now.Add(-time.Hour))
		s.checkpointAgo(3 * time.Hour)
		if err := s.run(); err != nil {
			t.Fatal(err)
		}
		input := s.stateFile("input-2026-11-22-1030.txt")
		if strings.Count(input, "\n") != 10 || strings.Contains(input, "mixed") {
			t.Fatalf("input: %q", input)
		}
	})

	t.Run("given lines from 30 hours ago and a 24 hour window with no checkpoint, when running, then the window drops them and the gate blocks", func(t *testing.T) {
		s := newScenario(t)
		s.lines("old", 12, s.now.Add(-30*time.Hour), s.now.Add(-time.Hour))
		if err := s.run(); err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(s.out.String(), "só 0 falas novas") || !s.claudeNotCalled() {
			t.Fatalf("output: %q", s.out.String())
		}
	})

	t.Run("given lines from 30 hours ago and a 48 hour window with no checkpoint, when running, then the window keeps them and the model runs", func(t *testing.T) {
		s := newScenario(t)
		s.cfg.WindowHours = 48
		s.lines("old", 12, s.now.Add(-30*time.Hour), s.now.Add(-time.Hour))
		if err := s.run(); err != nil {
			t.Fatal(err)
		}
		if s.claudeNotCalled() {
			t.Fatal("claude should have run")
		}
	})
}

func TestRunDailyBudget(t *testing.T) {
	budgetArgument := func(s *scenario) string {
		for i, argument := range s.claudeArguments() {
			if argument == "--max-budget-usd" {
				return s.claudeArguments()[i+1]
			}
		}
		return ""
	}

	t.Run("given no metrics file, when running, then it runs with the per-round budget", func(t *testing.T) {
		s := newScenario(t)
		s.fresh(12)
		if err := s.run(); err != nil {
			t.Fatal(err)
		}
		if got := budgetArgument(s); got != "3" {
			t.Fatalf("expected 3, got %q", got)
		}
	})

	t.Run("given an empty metrics file, when running, then it runs with the per-round budget", func(t *testing.T) {
		s := newScenario(t)
		s.metricsRows()
		s.fresh(12)
		if err := s.run(); err != nil {
			t.Fatal(err)
		}
		if got := budgetArgument(s); got != "3" {
			t.Fatalf("expected 3, got %q", got)
		}
	})

	t.Run("given rows from today below the daily cap with room left, when running, then it keeps the per-round budget", func(t *testing.T) {
		s := newScenario(t)
		s.metricsRows(costRow("2026-11-22-0800", "0.4"), costRow("2026-11-22-0900", "0.6"))
		s.fresh(12)
		if err := s.run(); err != nil {
			t.Fatal(err)
		}
		if got := budgetArgument(s); got != "3" {
			t.Fatalf("expected 3, got %q", got)
		}
	})

	t.Run("given rows from today that reach the daily cap, when running, then it prints the reason and does not call the model", func(t *testing.T) {
		s := newScenario(t)
		s.metricsRows(costRow("2026-11-22-0700", "3"), costRow("2026-11-22-0900", "3"))
		s.fresh(12)
		if err := s.run(); err != nil {
			t.Fatal(err)
		}
		expected := "teto diário de US$ 6.00 atingido (US$ 6.00 gastos hoje): não rodo o modelo\n"
		if s.out.String() != expected {
			t.Fatalf("expected %q, got %q", expected, s.out.String())
		}
		if !s.claudeNotCalled() || s.branches() != "" || len(s.reminder.calls) != 0 || s.stateHas("input-2026-11-22-1030.txt") {
			t.Fatal("the round should have stopped before any effect")
		}
	})

	t.Run("given rows from today above the daily cap, when running, then it blocks the round", func(t *testing.T) {
		s := newScenario(t)
		s.metricsRows(costRow("2026-11-22-0700", "7.25"))
		s.fresh(12)
		if err := s.run(); err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(s.out.String(), "teto diário de US$ 6.00 atingido (US$ 7.25 gastos hoje)") || !s.claudeNotCalled() {
			t.Fatalf("output: %q", s.out.String())
		}
	})

	t.Run("given rows only from other days, when running, then it ignores them and keeps the per-round budget", func(t *testing.T) {
		s := newScenario(t)
		s.metricsRows(costRow("2026-11-21-2359", "9"), costRow("2026-11-23-0001", "9"), costRow("2025-11-22-0900", "9"))
		s.fresh(12)
		if err := s.run(); err != nil {
			t.Fatal(err)
		}
		if got := budgetArgument(s); got != "3" {
			t.Fatalf("expected 3, got %q", got)
		}
	})

	t.Run("given a remaining balance smaller than the per-round budget, when running, then it passes the balance to claude", func(t *testing.T) {
		s := newScenario(t)
		s.metricsRows(costRow("2026-11-21-1000", "5"), costRow("2026-11-22-0700", "4.1"))
		s.fresh(12)
		if err := s.run(); err != nil {
			t.Fatal(err)
		}
		if got := budgetArgument(s); got != "1.9" {
			t.Fatalf("expected 1.9, got %q", got)
		}
	})

	t.Run("given a finished round, when the next one starts the same day, then the cost of the first counts against the cap", func(t *testing.T) {
		s := newScenario(t)
		s.cfg.DailyBudgetUSD = 0.1
		s.fresh(12)
		if err := s.run(); err != nil {
			t.Fatal(err)
		}
		s.now = s.now.Add(3 * time.Hour)
		s.fresh(12)
		s.out.Reset()
		if err := s.run(); err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(s.out.String(), "teto diário de US$ 0.10 atingido (US$ 0.12 gastos hoje)") {
			t.Fatalf("output: %q", s.out.String())
		}
	})
}

func TestRunFailures(t *testing.T) {
	t.Run("given a claude that fails, when running, then it returns the error with no metrics, no checkpoint and no reminder", func(t *testing.T) {
		s := newScenario(t)
		s.fakeClaude("", 1)
		s.fresh(12)
		if err := s.run(); err == nil {
			t.Fatal("expected an error")
		}
		if s.stateHas("metrics.tsv") || s.stateHas("checkpoint") || len(s.reminder.calls) != 0 {
			t.Fatal("a failed round must leave no trace of success")
		}
	})

	t.Run("given a repo without a learning agent, when running, then it returns an error before creating the worktree and before calling the model", func(t *testing.T) {
		s := newScenario(t)
		if err := os.Remove(filepath.Join(s.cfg.Repo, "agents", "learner.md")); err != nil {
			t.Fatal(err)
		}
		s.fresh(12)
		err := s.run()
		if err == nil || !strings.Contains(err.Error(), "Learning.") {
			t.Fatalf("got %v", err)
		}
		if !s.claudeNotCalled() || s.branches() != "" {
			t.Fatal("nothing should have been created")
		}
	})

	t.Run("given a reminder that fails, when running, then it returns the error but the metrics and the checkpoint are already written", func(t *testing.T) {
		s := newScenario(t)
		s.reminder.fail = errors.New("no permission")
		s.fresh(12)
		err := s.run()
		if err == nil || !strings.Contains(err.Error(), "no permission") {
			t.Fatalf("got %v", err)
		}
		if !s.stateHas("metrics.tsv") || !s.stateHas("checkpoint") {
			t.Fatal("the state should be written")
		}
	})
}

func gitSnapshot(t *testing.T, repo string) string {
	t.Helper()
	var snapshot strings.Builder
	for _, args := range [][]string{
		{"for-each-ref"},
		{"config", "core.bare"},
		{"worktree", "list"},
		{"rev-parse", "HEAD"},
		{"status", "--porcelain"},
		{"ls-files", "--stage"},
	} {
		cmd := exec.Command("git", append([]string{"-C", repo}, args...)...)
		cmd.Env = testGitEnv()
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
		snapshot.Write(out)
	}
	return snapshot.String()
}

func TestRunIgnoresGitVariablesFromTheEnvironment(t *testing.T) {
	t.Run("given GIT_DIR, GIT_INDEX_FILE and GIT_WORK_TREE pointing at another repo, when running, then that repo does not change and the round uses its own repo", func(t *testing.T) {
		s := newScenario(t)
		decoy := filepath.Join(t.TempDir(), "decoy")
		if err := os.MkdirAll(decoy, 0o750); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(decoy, "keep.txt"), []byte("keep\n"), 0o600); err != nil {
			t.Fatal(err)
		}
		execute(t, decoy, "git", "init", "-q", "-b", "main")
		execute(t, decoy, "git", "add", "-A")
		execute(t, decoy, "git", "-c", "user.name=t", "-c", "user.email=t@t.t", "commit", "-q", "-m", "decoy")
		before := gitSnapshot(t, decoy)

		t.Setenv("GIT_DIR", filepath.Join(decoy, ".git"))
		t.Setenv("GIT_INDEX_FILE", filepath.Join(decoy, ".git", "index"))
		t.Setenv("GIT_WORK_TREE", decoy)
		s.fresh(12)
		if err := s.run(); err != nil {
			t.Fatal(err)
		}

		if after := gitSnapshot(t, decoy); after != before {
			t.Fatalf("the repo from the environment changed:\nbefore\n%s\nafter\n%s", before, after)
		}
		if branches := s.branches(); !strings.Contains(branches, "learn/2026-11-22-1030") {
			t.Fatalf("the round did not use its own repo: %q", branches)
		}
	})
}

func TestReminderScript(t *testing.T) {
	t.Run("given a title and body with quotes and backslashes, when building the AppleScript, then it escapes both", func(t *testing.T) {
		script := reminderScript(`a "title"`, `body \ with "quotes"`)
		if !strings.Contains(script, `name:"a \"title\""`) || !strings.Contains(script, `body:"body \\ with \"quotes\""`) {
			t.Fatalf("script: %s", script)
		}
		if !strings.Contains(script, `tell list "Tarefas Claude"`) {
			t.Fatalf("missing the list: %s", script)
		}
	})
}

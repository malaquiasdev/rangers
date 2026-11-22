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

const (
	claudeJSON = `{"type":"result","is_error":false,"duration_ms":5400,"num_turns":3,` +
		`"total_cost_usd":0.12345,"usage":{"input_tokens":100,"cache_creation_input_tokens":500,` +
		`"cache_read_input_tokens":1000,"output_tokens":50},"result":"proposed rule"}`
	learningAgent = "---\nname: learner\ndescription: Learning. Reads the transcripts and " +
		"proposes rules.\nmodel: opus\n---\n\nbody\n"
	otherAgent = "---\nname: other\ndescription: Author. Implements one unit.\n" +
		"model: sonnet\n---\n\nbody\n"
)

type fakeReminder struct {
	calls [][2]string
	fail  error
}

type fakeBrain struct {
	dirs []string
	fail error
}

type blockingClaude struct{}

type slowAfterFirst struct{ calls int }

type brokenBranchLookup struct {
	GitCLI
	fail error
}

type scenario struct {
	t        *testing.T
	cfg      Config
	home     string
	now      time.Time
	log      string
	reminder *fakeReminder
	out      *bytes.Buffer
	errOut   *bytes.Buffer
	claude   Claude
	brain    Brain
	git      Git
}

func (f *fakeBrain) Probe(_ context.Context, dir string) error {
	f.dirs = append(f.dirs, dir)
	return f.fail
}

func (blockingClaude) Run(ctx context.Context, _ string, _ []string) ([]byte, error) {
	<-ctx.Done()
	return nil, ctx.Err()
}

func (f *fakeReminder) Create(_ context.Context, title, body string) error {
	f.calls = append(f.calls, [2]string{title, body})
	return f.fail
}

func (b brokenBranchLookup) BranchExists(_ context.Context, _, _ string) (bool, error) {
	return false, b.fail
}

func writeExecutable(t *testing.T, path, content string) {
	t.Helper()
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o600)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = file.WriteString(content); err != nil {
		t.Fatal(err)
	}
	if err = file.Chmod(0o500); err != nil {
		t.Fatal(err)
	}
	if err = file.Close(); err != nil {
		t.Fatal(err)
	}
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
	cmd := exec.CommandContext(t.Context(), name, args...)
	cmd.Dir = dir
	cmd.Env = testGitEnv()
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("%s %v: %v\n%s", name, args, err, out)
	}
}

func newScenario(t *testing.T) *scenario {
	t.Helper()
	root := t.TempDir()
	home := filepath.Join(root, "home")
	t.Setenv("HOME", home)
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
	if err := os.MkdirAll(filepath.Join(root, "brain"), 0o750); err != nil {
		t.Fatal(err)
	}
	execute(t, repo, "git", "init", "-q", "-b", "main")
	execute(t, repo, "git", "add", "-A")
	execute(t, repo, "git", "-c", "user.name=t", "-c", "user.email=t@t.t", "commit", "-q", "-m", "initial")

	s := &scenario{
		t:        t,
		home:     home,
		now:      time.Date(2026, 11, 22, 10, 30, 0, 0, time.Local),
		log:      filepath.Join(root, "claude.log"),
		reminder: &fakeReminder{},
		out:      &bytes.Buffer{},
		errOut:   &bytes.Buffer{},
		cfg: Config{
			IntervalHours:      2,
			MinLines:           10,
			BudgetUSD:          3,
			DailyBudgetUSD:     6,
			Model:              "opus",
			WindowHours:        24,
			State:              filepath.Join(root, "state"),
			Repo:               repo,
			Projects:           filepath.Join(root, "projects"),
			Brain:              filepath.Join(root, "brain"),
			NotesModel:         "sonnet",
			NotesMaxBytes:      60000,
			SkipNotes:          true,
			CallTimeoutMinutes: 20,
		},
	}
	s.fakeClaude(claudeJSON, 0)
	return s
}

func (s *scenario) fakeClaude(reply string, code int) {
	s.t.Helper()
	s.fakeClaudeFailingOn(reply, code, "none")
}

func (s *scenario) fakeClaudeFailingOn(reply string, code int, failModel string) {
	s.t.Helper()
	bin := filepath.Join(s.t.TempDir(), "bin")
	if err := os.MkdirAll(bin, 0o750); err != nil {
		s.t.Fatal(err)
	}
	script := fmt.Sprintf("#!/bin/sh\nprintf -- '---\\n' >> %q\npwd -P >> %q\nprev=\nfail=0\nfor a in \"$@\"; do\n  printf '%%s\\n' \"$a\" >> %q\n  if [ \"$prev\" = --model ] && [ \"$a\" = %q ]; then fail=1; fi\n  prev=$a\ndone\n[ $fail = 1 ] && exit 1\nprintf '%%s' '%s'\nexit %d\n", s.log, s.log, s.log, failModel, reply, code)
	writeExecutable(s.t, filepath.Join(bin, "claude"), script)
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
	var claude Claude = ClaudeCLI{}
	if s.claude != nil {
		claude = s.claude
	}
	var brain Brain = DirBrain{Timeout: time.Second}
	if s.brain != nil {
		brain = s.brain
	}
	var git Git = GitCLI{}
	if s.git != nil {
		git = s.git
	}
	return Run(context.Background(), &s.cfg, &Dependencies{
		Claude:   claude,
		Brain:    brain,
		Binary:   "/opt/rangers/bin/rangers",
		Home:     s.home,
		Git:      git,
		Reminder: s.reminder,
		Now:      func() time.Time { return s.now },
		Out:      s.out,
		Err:      s.errOut,
	})
}

func (s *scenario) claudeCalls() [][]string {
	s.t.Helper()
	raw, err := os.ReadFile(s.log)
	if err != nil {
		s.t.Fatalf("claude was not called: %v", err)
	}
	var calls [][]string
	for _, block := range strings.Split(string(raw), "---\n")[1:] {
		calls = append(calls, strings.Split(strings.TrimSpace(block), "\n"))
	}
	return calls
}

func (s *scenario) claudeArguments() []string {
	s.t.Helper()
	return s.claudeCalls()[0]
}

func (s *scenario) claudeNotCalled() bool {
	_, err := os.Stat(s.log)
	return errors.Is(err, os.ErrNotExist)
}

func (s *scenario) branches() string {
	s.t.Helper()
	cmd := exec.CommandContext(s.t.Context(), "git", "-C", s.cfg.Repo, "branch", "--list", "learn/*")
	cmd.Env = testGitEnv()
	out, err := cmd.Output()
	if err != nil {
		s.t.Fatal(err)
	}
	return strings.TrimSpace(string(out))
}

func (s *scenario) branch(name string) {
	s.t.Helper()
	execute(s.t, s.cfg.Repo, "git", "branch", name, "main")
}

func (s *scenario) pending(content string) {
	s.t.Helper()
	if err := os.MkdirAll(s.cfg.State, 0o750); err != nil {
		s.t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(s.cfg.State, "pending-reminders.txt"), []byte(content), 0o600); err != nil {
		s.t.Fatal(err)
	}
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

func (s *scenario) assistant(file, text string, at time.Time) {
	s.t.Helper()
	target := filepath.Join(s.cfg.Projects, "proj", file+".jsonl")
	raw, err := json.Marshal(map[string]any{
		"type":      "assistant",
		"timestamp": at.UTC().Format("2006-01-02T15:04:05.000Z"),
		"message":   map[string]any{"role": "assistant", "content": []map[string]any{{"type": "text", "text": text}}},
	})
	if err != nil {
		s.t.Fatal(err)
	}
	handle, err := os.OpenFile(target, os.O_APPEND|os.O_WRONLY, 0o600)
	if err != nil {
		s.t.Fatal(err)
	}
	if _, err := handle.WriteString(string(raw) + "\n"); err != nil {
		s.t.Fatal(err)
	}
	if err := handle.Close(); err != nil {
		s.t.Fatal(err)
	}
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
		expected := "2026-11-22-1030\t12\t5\t3\t0.123\t1600\t50\trules\n"
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

	t.Run("given a rules round, when running, then the prompt forbids quotes and session ids in the files and asks for them under Origin: in the final message", func(t *testing.T) {
		s := newScenario(t)
		s.fresh(12)
		if err := s.run(); err != nil {
			t.Fatal(err)
		}
		prompt := argumentAfter(s.claudeArguments(), "-p")
		for _, piece := range []string{
			"Do not write quotes, session ids or Origin in the rules files or in the commit messages",
			"final message",
			"`Origin:`",
			"one bullet per rule",
		} {
			if !strings.Contains(prompt, piece) {
				t.Errorf("missing %q in %q", piece, prompt)
			}
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

func argumentAfter(call []string, flag string) string {
	for i, argument := range call {
		if argument == flag && i+1 < len(call) {
			return call[i+1]
		}
	}
	return ""
}

func (s *scenario) withNotes() {
	s.cfg.SkipNotes = false
	s.fresh(12)
	s.assistant("new-session", "decided to use the cache", s.now.Add(-30*time.Minute))
}

func (s *scenario) resolvedHome() {
	s.t.Helper()
	if err := os.MkdirAll(filepath.Join(s.home, ".claude", "bin"), 0o750); err != nil {
		s.t.Fatal(err)
	}
	resolved, err := filepath.EvalSymlinks(s.home)
	if err != nil {
		s.t.Fatal(err)
	}
	s.home = resolved
}

func (s *scenario) auditFile(link bool) string {
	s.t.Helper()
	s.resolvedHome()
	path := filepath.Join(s.home, ".claude", "bin", "audit")
	if !link {
		writeExecutable(s.t, path, "#!/bin/sh\n")
		return path
	}
	target := filepath.Join(s.t.TempDir(), "audit")
	writeExecutable(s.t, target, "#!/bin/sh\n")
	if err := os.Symlink(target, path); err != nil {
		s.t.Fatal(err)
	}
	resolved, err := filepath.EvalSymlinks(target)
	if err != nil {
		s.t.Fatal(err)
	}
	return resolved
}

func (s *scenario) notesAllowedTools() string {
	s.t.Helper()
	if err := s.run(); err != nil {
		s.t.Fatal(err)
	}
	calls := s.claudeCalls()
	if len(calls) != 2 {
		s.t.Fatalf("expected 2 calls, got %d", len(calls))
	}
	return argumentAfter(calls[1], "--allowedTools")
}

func TestRunNotesPass(t *testing.T) {
	t.Run("given human and assistant lines, when running, then the notes pass runs in the brain folder with the notes model and the right tools", func(t *testing.T) {
		s := newScenario(t)
		s.withNotes()
		if err := s.run(); err != nil {
			t.Fatal(err)
		}
		calls := s.claudeCalls()
		if len(calls) != 2 {
			t.Fatalf("expected 2 calls, got %d", len(calls))
		}
		notes := calls[1]
		brain, err := filepath.EvalSymlinks(s.cfg.Brain)
		if err != nil {
			t.Fatal(err)
		}
		if notes[0] != brain {
			t.Fatalf("notes pass ran in %q, expected %q", notes[0], brain)
		}
		input := filepath.Join(s.cfg.State, "notes-input-2026-11-22-1030.txt")
		if prompt := argumentAfter(notes, "-p"); !strings.Contains(prompt, input) || !strings.Contains(prompt, "notes pass") {
			t.Fatalf("prompt: %q", prompt)
		}
		for flag, expected := range map[string]string{
			"--agent":           "learner",
			"--model":           "sonnet",
			"--permission-mode": "acceptEdits",
			"--allowedTools":    "Read,Write,Edit,Glob,Grep,Bash(~/.claude/bin/audit:*),Bash(" + s.home + "/.claude/bin/audit:*)",
			"--output-format":   "json",
			"--max-budget-usd":  "3",
		} {
			if got := argumentAfter(notes, flag); got != expected {
				t.Errorf("%s: expected %q, got %q", flag, expected, got)
			}
		}
	})

	t.Run("given the audit script is a link to a file elsewhere, when running, then the notes pass also allows the resolved path", func(t *testing.T) {
		s := newScenario(t)
		s.withNotes()
		resolved := s.auditFile(true)
		expected := "Read,Write,Edit,Glob,Grep,Bash(~/.claude/bin/audit:*),Bash(" + s.home + "/.claude/bin/audit:*),Bash(" + resolved + ":*)"
		if got := s.notesAllowedTools(); got != expected {
			t.Fatalf("expected %q, got %q", expected, got)
		}
	})

	t.Run("given the audit script is a regular file, when running, then the allowed path is not repeated", func(t *testing.T) {
		s := newScenario(t)
		s.withNotes()
		s.auditFile(false)
		expected := "Read,Write,Edit,Glob,Grep,Bash(~/.claude/bin/audit:*),Bash(" + s.home + "/.claude/bin/audit:*)"
		if got := s.notesAllowedTools(); got != expected {
			t.Fatalf("expected %q, got %q", expected, got)
		}
	})

	t.Run("given the audit script does not exist, when running, then the round does not fail and allows only the two home forms", func(t *testing.T) {
		s := newScenario(t)
		s.withNotes()
		s.resolvedHome()
		expected := "Read,Write,Edit,Glob,Grep,Bash(~/.claude/bin/audit:*),Bash(" + s.home + "/.claude/bin/audit:*)"
		if got := s.notesAllowedTools(); got != expected {
			t.Fatalf("expected %q, got %q", expected, got)
		}
	})

	t.Run("given human and assistant lines, when running, then the notes input marks the project and the assistant text", func(t *testing.T) {
		s := newScenario(t)
		s.withNotes()
		if err := s.run(); err != nil {
			t.Fatal(err)
		}
		input := s.stateFile("notes-input-2026-11-22-1030.txt")
		if strings.Count(input, "\n") != 13 || !strings.Contains(input, "proj\tnew-session\t") || !strings.Contains(input, "[assistant] decided to use the cache") {
			t.Fatalf("input: %q", input)
		}
		if raw := s.stateFile("notes-output-2026-11-22-1030.json"); raw != claudeJSON {
			t.Fatalf("stored output: %q", raw)
		}
	})

	t.Run("given a small byte limit, when running, then the notes input keeps the most recent lines", func(t *testing.T) {
		s := newScenario(t)
		s.withNotes()
		s.cfg.NotesMaxBytes = 200
		if err := s.run(); err != nil {
			t.Fatal(err)
		}
		input := s.stateFile("notes-input-2026-11-22-1030.txt")
		if len(input) > 200 || !strings.HasSuffix(input, "[assistant] decided to use the cache\n") {
			t.Fatalf("input (%d bytes): %q", len(input), input)
		}
	})

	t.Run("given spend earlier today, when running, then the notes pass gets the remaining balance after the rules pass", func(t *testing.T) {
		s := newScenario(t)
		s.withNotes()
		s.metricsRows(costRow("2026-11-22-0700", "4.1"))
		if err := s.run(); err != nil {
			t.Fatal(err)
		}
		calls := s.claudeCalls()
		if got := argumentAfter(calls[0], "--max-budget-usd"); got != "1.9" {
			t.Fatalf("rules pass budget: %q", got)
		}
		if got := argumentAfter(calls[1], "--max-budget-usd"); got != "1.777" {
			t.Fatalf("notes pass budget: %q", got)
		}
	})

	t.Run("given a daily cap, when both passes have run, then the next round is blocked in both passes because the cost of both counts", func(t *testing.T) {
		s := newScenario(t)
		s.withNotes()
		s.cfg.DailyBudgetUSD = 0.2
		if err := s.run(); err != nil {
			t.Fatal(err)
		}
		if got := argumentAfter(s.claudeCalls()[1], "--max-budget-usd"); got != "0.077" {
			t.Fatalf("notes pass budget: %q", got)
		}
		s.now = s.now.Add(3 * time.Hour)
		s.fresh(12)
		s.out.Reset()
		if err := s.run(); err != nil {
			t.Fatal(err)
		}
		expected := "teto diário de US$ 0.20 atingido (US$ 0.25 gastos hoje): não rodo o modelo\n"
		if s.out.String() != expected+"[notas] "+expected {
			t.Fatalf("output: %q", s.out.String())
		}
		if len(s.claudeCalls()) != 2 {
			t.Fatal("no new call was expected")
		}
	})

	t.Run("given a notes pass that fails, when running, then the rules pass stays done and only the notes pass reports the error", func(t *testing.T) {
		s := newScenario(t)
		s.withNotes()
		s.fakeClaudeFailingOn(claudeJSON, 0, "sonnet")
		if err := s.run(); err == nil {
			t.Fatal("expected an error")
		}
		if !s.stateHas("checkpoint") || len(s.reminder.calls) != 1 || !strings.Contains(s.stateFile("metrics.tsv"), "\trules\n") {
			t.Fatal("the rules pass should be done")
		}
		if s.stateHas("checkpoint-notes") || strings.Contains(s.stateFile("metrics.tsv"), "\tnotes\n") {
			t.Fatal("the failed notes pass must leave no trace of success")
		}
	})

	t.Run("given a rules pass that fails, when running, then the notes pass still runs and the rules checkpoint stays untouched", func(t *testing.T) {
		s := newScenario(t)
		s.withNotes()
		s.fakeClaudeFailingOn(claudeJSON, 0, "opus")
		if err := s.run(); err == nil {
			t.Fatal("expected an error")
		}
		if s.stateHas("checkpoint") || len(s.reminder.calls) != 0 {
			t.Fatal("the failed rules pass must leave no trace of success")
		}
		if !s.stateHas("checkpoint-notes") || !strings.Contains(s.stateFile("metrics.tsv"), "\tnotes\n") {
			t.Fatal("the notes pass should be done")
		}
	})

	t.Run("given skip notes, when running, then only the rules pass calls the model", func(t *testing.T) {
		s := newScenario(t)
		s.withNotes()
		s.cfg.SkipNotes = true
		if err := s.run(); err != nil {
			t.Fatal(err)
		}
		if len(s.claudeCalls()) != 1 || s.stateHas("notes-input-2026-11-22-1030.txt") || s.stateHas("checkpoint-notes") {
			t.Fatal("the notes pass should not have run")
		}
	})

	t.Run("given a finished notes pass, when it ends, then it records a notes row, updates its own checkpoint and creates no reminder", func(t *testing.T) {
		s := newScenario(t)
		s.withNotes()
		if err := s.run(); err != nil {
			t.Fatal(err)
		}
		metrics := s.stateFile("metrics.tsv")
		if !strings.Contains(metrics, "2026-11-22-1030\t12\t5\t3\t0.123\t1600\t50\tnotes\n") {
			t.Fatalf("metrics: %q", metrics)
		}
		info, err := os.Stat(filepath.Join(s.cfg.State, "checkpoint-notes"))
		if err != nil || !info.ModTime().Equal(s.now) {
			t.Fatalf("checkpoint-notes: %v, %v", info, err)
		}
		if len(s.reminder.calls) != 1 {
			t.Fatalf("expected only the rules reminder, got %d", len(s.reminder.calls))
		}
	})

	t.Run("given a notes checkpoint 30 minutes ago, when running, then only the notes pass is blocked by the interval", func(t *testing.T) {
		s := newScenario(t)
		s.withNotes()
		if err := os.MkdirAll(s.cfg.State, 0o750); err != nil {
			t.Fatal(err)
		}
		path := filepath.Join(s.cfg.State, "checkpoint-notes")
		if err := os.WriteFile(path, nil, 0o600); err != nil {
			t.Fatal(err)
		}
		recent := s.now.Add(-30 * time.Minute)
		if err := os.Chtimes(path, recent, recent); err != nil {
			t.Fatal(err)
		}
		if err := s.run(); err != nil {
			t.Fatal(err)
		}
		if len(s.claudeCalls()) != 1 || !strings.Contains(s.out.String(), "[notas] a última rodada foi há 30 min, menos que o intervalo de 2h") {
			t.Fatalf("calls %d, output %q", len(s.claudeCalls()), s.out.String())
		}
	})

	t.Run("given nine human lines and many assistant lines, when running, then the notes pass counts only the human lines for the minimum", func(t *testing.T) {
		s := newScenario(t)
		s.cfg.SkipNotes = false
		s.fresh(9)
		for i := range 5 {
			s.assistant("new-session", fmt.Sprintf("answer %d", i), s.now.Add(-30*time.Minute))
		}
		if err := s.run(); err != nil {
			t.Fatal(err)
		}
		if !s.claudeNotCalled() || !strings.Contains(s.out.String(), "[notas] só 9 falas novas, abaixo do mínimo de 10") {
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
		cmd := exec.CommandContext(t.Context(), "git", append([]string{"-C", repo}, args...)...)
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

func TestRunReminderFailure(t *testing.T) {
	failing := func(t *testing.T) *scenario {
		t.Helper()
		s := newScenario(t)
		s.reminder.fail = errors.New("AppleEvent timed out (-1712)")
		s.fresh(12)
		return s
	}
	pendingLine := "Revisar as regras propostas\tRamo learn/2026-11-22-1030 em %s/rangers-worktrees/learn-2026-11-22-1030\n"

	t.Run("given a reminder that fails, when running, then it succeeds and the metrics, the checkpoint and the branch line are written", func(t *testing.T) {
		s := failing(t)
		if err := s.run(); err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if !s.stateHas("metrics.tsv") || !s.stateHas("checkpoint") {
			t.Fatal("the state should be written")
		}
		if !strings.Contains(s.out.String(), "ramo learn/2026-11-22-1030 em ") {
			t.Fatalf("output: %q", s.out.String())
		}
	})

	t.Run("given a reminder that fails, when running, then it prints a warning in Portuguese on the error output", func(t *testing.T) {
		s := failing(t)
		if err := s.run(); err != nil {
			t.Fatal(err)
		}
		warning := s.errOut.String()
		if !strings.HasPrefix(warning, "aviso: não consegui criar o lembrete") || !strings.Contains(warning, "AppleEvent timed out (-1712)") || !strings.Contains(warning, "pending-reminders.txt") {
			t.Fatalf("warning: %q", warning)
		}
		if strings.Contains(s.out.String(), "aviso") {
			t.Fatalf("the warning leaked to the output: %q", s.out.String())
		}
	})

	t.Run("given a reminder that fails, when running, then it appends the title and the branch to pending-reminders.txt", func(t *testing.T) {
		s := failing(t)
		if err := s.run(); err != nil {
			t.Fatal(err)
		}
		expected := fmt.Sprintf(pendingLine, filepath.Dir(s.cfg.Repo))
		if got := s.stateFile("pending-reminders.txt"); got != expected {
			t.Fatalf("expected %q, got %q", expected, got)
		}
	})

	t.Run("given two rounds whose reminders fail, when running both, then the file keeps one line per round", func(t *testing.T) {
		s := failing(t)
		if err := s.run(); err != nil {
			t.Fatal(err)
		}
		s.now = s.now.Add(3 * time.Hour)
		s.fresh(12)
		if err := s.run(); err != nil {
			t.Fatal(err)
		}
		lines := strings.Split(strings.TrimSpace(s.stateFile("pending-reminders.txt")), "\n")
		if len(lines) != 2 || !strings.Contains(lines[1], "learn/2026-11-22-1330") {
			t.Fatalf("lines: %q", lines)
		}
	})

	t.Run("given a reminder that works, when running, then there is no warning and no pending file", func(t *testing.T) {
		s := newScenario(t)
		s.fresh(12)
		if err := s.run(); err != nil {
			t.Fatal(err)
		}
		if s.errOut.Len() != 0 || s.stateHas("pending-reminders.txt") {
			t.Fatalf("warning %q, pending file %v", s.errOut.String(), s.stateHas("pending-reminders.txt"))
		}
	})
}

func TestRunPrunesPendingReminders(t *testing.T) {
	const title = "Revisar as regras propostas"
	gone := title + "\tRamo learn/2026-11-20-0900 em /work/learn-2026-11-20-0900\n"
	kept := title + "\tRamo learn/2026-11-21-0900 em /work/learn-2026-11-21-0900\n"
	free := "Dar acesso ao brain ao rangers\tAcesso Total ao Disco para /opt/rangers/bin/rangers em Ajustes\n"
	blocked := func(t *testing.T) *scenario {
		t.Helper()
		s := newScenario(t)
		s.checkpointAgo(30 * time.Minute)
		s.branch("learn/2026-11-21-0900")
		return s
	}

	t.Run("given a line for a branch that no longer exists, when a round starts, then the line is removed and no reminder is created for it", func(t *testing.T) {
		s := blocked(t)
		s.pending(gone + kept)
		if err := s.run(); err != nil {
			t.Fatal(err)
		}
		if got := s.stateFile("pending-reminders.txt"); got != kept {
			t.Fatalf("expected %q, got %q", kept, got)
		}
		if len(s.reminder.calls) != 0 {
			t.Fatalf("reminders: %v", s.reminder.calls)
		}
	})

	t.Run("given a line for a branch that still exists, when a round starts, then the file stays untouched", func(t *testing.T) {
		s := blocked(t)
		s.pending(kept)
		if err := s.run(); err != nil {
			t.Fatal(err)
		}
		if got := s.stateFile("pending-reminders.txt"); got != kept {
			t.Fatalf("expected %q, got %q", kept, got)
		}
	})

	t.Run("given a free-text line with no branch name, when a round starts, then the line stays untouched", func(t *testing.T) {
		s := blocked(t)
		s.pending(free + gone)
		if err := s.run(); err != nil {
			t.Fatal(err)
		}
		if got := s.stateFile("pending-reminders.txt"); got != free {
			t.Fatalf("expected %q, got %q", free, got)
		}
	})

	t.Run("given no pending file, when a round starts, then the round runs as before and creates no file", func(t *testing.T) {
		s := newScenario(t)
		s.fresh(12)
		if err := s.run(); err != nil {
			t.Fatal(err)
		}
		if s.stateHas("pending-reminders.txt") || len(s.claudeCalls()) != 1 {
			t.Fatalf("pending file %v, claude calls %d", s.stateHas("pending-reminders.txt"), len(s.claudeCalls()))
		}
	})

	t.Run("given an empty pending file, when a round starts, then the file stays empty", func(t *testing.T) {
		s := blocked(t)
		s.pending("")
		if err := s.run(); err != nil {
			t.Fatal(err)
		}
		if got := s.stateFile("pending-reminders.txt"); got != "" {
			t.Fatalf("expected an empty file, got %q", got)
		}
	})

	t.Run("given a git that fails, when a round starts, then the file stays intact, a warning goes to the error output and the round continues", func(t *testing.T) {
		s := newScenario(t)
		s.fresh(12)
		s.git = brokenBranchLookup{fail: errors.New("fatal: not a git repository")}
		s.pending(gone + kept)
		if err := s.run(); err != nil {
			t.Fatal(err)
		}
		if got := s.stateFile("pending-reminders.txt"); got != gone+kept {
			t.Fatalf("file changed: %q", got)
		}
		if warning := s.errOut.String(); !strings.HasPrefix(warning, "aviso: ") || !strings.Contains(warning, "not a git repository") {
			t.Fatalf("warning: %q", warning)
		}
		if len(s.claudeCalls()) != 1 {
			t.Fatal("the round should have run")
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

func TestRunCallTimeout(t *testing.T) {
	hanging := func(t *testing.T) *scenario {
		t.Helper()
		s := newScenario(t)
		s.cfg.CallTimeoutMinutes = 0.001
		s.claude = blockingClaude{}
		s.fresh(12)
		return s
	}

	t.Run("given a claude that never answers, when running, then it returns the timeout error within the configured limit", func(t *testing.T) {
		s := hanging(t)
		started := time.Now()
		err := s.run()
		if err == nil || !strings.Contains(err.Error(), "passou de") || !strings.Contains(err.Error(), "encerrado") {
			t.Fatalf("got %v", err)
		}
		if elapsed := time.Since(started); elapsed > 10*time.Second {
			t.Fatalf("took %v", elapsed)
		}
	})

	t.Run("given a claude that never answers, when running, then it leaves no metrics, no checkpoint and no reminder", func(t *testing.T) {
		s := hanging(t)
		if err := s.run(); err == nil {
			t.Fatal("expected an error")
		}
		if s.stateHas("metrics.tsv") || s.stateHas("checkpoint") || len(s.reminder.calls) != 0 {
			t.Fatal("a timed out round must leave no trace of success")
		}
	})

	t.Run("given a claude that never answers in the notes pass, when running, then the rules pass still finishes and the notes error is returned", func(t *testing.T) {
		s := hanging(t)
		s.cfg.SkipNotes = false
		s.claude = &slowAfterFirst{}
		s.assistant("new-session", "decided to use the cache", s.now.Add(-30*time.Minute))
		err := s.run()
		if err == nil || !strings.Contains(err.Error(), "passou de") {
			t.Fatalf("got %v", err)
		}
		if !s.stateHas("checkpoint") || s.stateHas("checkpoint-notes") {
			t.Fatal("only the rules checkpoint should exist")
		}
	})
}

func (c *slowAfterFirst) Run(ctx context.Context, dir string, args []string) ([]byte, error) {
	c.calls++
	if c.calls == 1 {
		return []byte(claudeJSON), nil
	}
	return blockingClaude{}.Run(ctx, dir, args)
}

func TestRunBrainAccess(t *testing.T) {
	const message = "[notas] sem acesso ao brain (operation not permitted): dê Acesso Total ao Disco a " +
		"/opt/rangers/bin/rangers em Ajustes > Privacidade e Segurança; pulo as notas\n"
	denied := func(t *testing.T) *scenario {
		t.Helper()
		s := newScenario(t)
		s.withNotes()
		s.brain = &fakeBrain{fail: errors.New("operation not permitted")}
		return s
	}

	t.Run("given a brain folder the job cannot use, when running, then claude is called only for the rules pass and no notes input is written", func(t *testing.T) {
		s := denied(t)
		if err := s.run(); err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		calls := s.claudeCalls()
		if len(calls) != 1 || argumentAfter(calls[0], "--model") != "opus" {
			t.Fatalf("calls: %q", calls)
		}
		if s.stateHas("notes-input-2026-11-22-1030.txt") || s.stateHas("notes-output-2026-11-22-1030.json") {
			t.Fatal("the notes pass must not write its input or output")
		}
		if dirs := s.brain.(*fakeBrain).dirs; len(dirs) != 1 || dirs[0] != s.cfg.Brain {
			t.Fatalf("probed %q", dirs)
		}
	})

	t.Run("given a brain folder the job cannot use, when running, then it prints the notice once, with the binary path, on the output", func(t *testing.T) {
		s := denied(t)
		if err := s.run(); err != nil {
			t.Fatal(err)
		}
		if got := strings.Count(s.out.String(), message); got != 1 {
			t.Fatalf("notice appeared %d times in %q", got, s.out.String())
		}
	})

	t.Run("given a brain folder the job cannot use, when running, then it appends a reminder line to pending-reminders.txt", func(t *testing.T) {
		s := denied(t)
		if err := s.run(); err != nil {
			t.Fatal(err)
		}
		expected := "Dar acesso ao brain ao rangers\tAcesso Total ao Disco para /opt/rangers/bin/rangers em Ajustes > Privacidade e Segurança\n"
		if got := s.stateFile("pending-reminders.txt"); got != expected {
			t.Fatalf("expected %q, got %q", expected, got)
		}
	})

	t.Run("given a brain folder the job cannot use, when running, then the rules pass is unaffected and the notes checkpoint stays where it was", func(t *testing.T) {
		s := denied(t)
		if err := s.run(); err != nil {
			t.Fatal(err)
		}
		if !s.stateHas("checkpoint") || !s.stateHas("metrics.tsv") || s.branches() == "" {
			t.Fatal("the rules pass should have finished")
		}
		if s.stateHas("checkpoint-notes") {
			t.Fatal("the notes checkpoint must not advance")
		}
		if strings.Contains(s.stateFile("metrics.tsv"), "notes") {
			t.Fatalf("metrics: %q", s.stateFile("metrics.tsv"))
		}
	})

	t.Run("given a brain folder that works, when running, then the notes pass runs as before and no notice or pending line appears", func(t *testing.T) {
		s := newScenario(t)
		s.withNotes()
		s.brain = &fakeBrain{}
		if err := s.run(); err != nil {
			t.Fatal(err)
		}
		if len(s.claudeCalls()) != 2 || !s.stateHas("checkpoint-notes") || s.stateHas("pending-reminders.txt") {
			t.Fatal("the notes pass should have run")
		}
		if strings.Contains(s.out.String(), "sem acesso ao brain") {
			t.Fatalf("output: %q", s.out.String())
		}
	})

	t.Run("given a skipped notes pass, when running, then the brain folder is not probed", func(t *testing.T) {
		s := newScenario(t)
		s.fresh(12)
		s.brain = &fakeBrain{fail: errors.New("operation not permitted")}
		if err := s.run(); err != nil {
			t.Fatal(err)
		}
		if len(s.brain.(*fakeBrain).dirs) != 0 {
			t.Fatal("nothing should be probed")
		}
	})
}

func TestRunOriginNote(t *testing.T) {
	const withOrigin = `{"type":"result","is_error":false,"duration_ms":5400,"num_turns":3,` +
		`"total_cost_usd":0.1,"usage":{"input_tokens":1,"cache_creation_input_tokens":1,` +
		`"cache_read_input_tokens":1,"output_tokens":1},` +
		`"result":"1 rule: ask before merging\nOrigin:\n- ask before merging, session abc12345, \"nao mergeia sem eu mandar\"\n"}`
	const withoutOrigin = `{"type":"result","is_error":false,"duration_ms":5400,"num_turns":3,` +
		`"total_cost_usd":0.1,"usage":{"input_tokens":1,"cache_creation_input_tokens":1,` +
		`"cache_read_input_tokens":1,"output_tokens":1},"result":"1 rule: ask before merging"}`
	note := func(s *scenario) string {
		return filepath.Join(s.cfg.Brain, "claude", "estudos", "origem-das-regras-rangers.md")
	}

	t.Run("given a final message with an Origin: section and a writable brain, when the rules round ends, then the section is appended once under the round stamp", func(t *testing.T) {
		s := newScenario(t)
		s.fakeClaude(withOrigin, 0)
		s.fresh(12)
		if err := os.MkdirAll(filepath.Dir(note(s)), 0o750); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(note(s), []byte("# Origem\n"), 0o600); err != nil {
			t.Fatal(err)
		}
		if err := s.run(); err != nil {
			t.Fatal(err)
		}
		raw, err := os.ReadFile(note(s))
		if err != nil {
			t.Fatal(err)
		}
		expected := "# Origem\n\n## 2026-11-22-1030\n\n" +
			"- ask before merging, session abc12345, \"nao mergeia sem eu mandar\"\n"
		if string(raw) != expected {
			t.Fatalf("expected %q, got %q", expected, raw)
		}
		if strings.Count(string(raw), "## 2026-11-22-1030") != 1 {
			t.Fatalf("stamp heading repeated: %q", raw)
		}
	})

	t.Run("given a brain without the notes folder, when the rules round ends, then it creates the note", func(t *testing.T) {
		s := newScenario(t)
		s.fakeClaude(withOrigin, 0)
		s.fresh(12)
		if err := s.run(); err != nil {
			t.Fatal(err)
		}
		raw, err := os.ReadFile(note(s))
		if err != nil {
			t.Fatal(err)
		}
		if !strings.HasPrefix(string(raw), "## 2026-11-22-1030\n\n- ask before merging") {
			t.Fatalf("note: %q", raw)
		}
	})

	t.Run("given a final message without an Origin: section, when the rules round ends, then nothing is written to the brain", func(t *testing.T) {
		s := newScenario(t)
		s.fakeClaude(withoutOrigin, 0)
		s.fresh(12)
		if err := s.run(); err != nil {
			t.Fatal(err)
		}
		if _, err := os.Stat(note(s)); !errors.Is(err, os.ErrNotExist) {
			t.Fatalf("the note must not exist: %v", err)
		}
	})

	t.Run("given a brain the job cannot write, when the rules round ends, then nothing is written and the round does not fail", func(t *testing.T) {
		s := newScenario(t)
		s.fakeClaude(withOrigin, 0)
		s.fresh(12)
		s.brain = &fakeBrain{fail: errors.New("operation not permitted")}
		if err := s.run(); err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if _, err := os.Stat(note(s)); !errors.Is(err, os.ErrNotExist) {
			t.Fatalf("the note must not exist: %v", err)
		}
		if dirs := s.brain.(*fakeBrain).dirs; len(dirs) != 1 || dirs[0] != s.cfg.Brain {
			t.Fatalf("probed %q", dirs)
		}
		if s.errOut.Len() != 0 || len(s.reminder.calls) != 1 {
			t.Fatalf("err output %q, reminders %d", s.errOut.String(), len(s.reminder.calls))
		}
	})
}

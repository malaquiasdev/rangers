package gate

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/malaquiasdev/rangers/internal/transcript"
)

const (
	repo     = "acme/api"
	number   = "7"
	headSHA  = "0123456789abcdef0123456789abcdef01234567"
	homeDir  = "/home/x"
	viewCmd  = "gh pr view 7 --repo acme/api --json " + prFields
	repoCmd  = "gh repo view acme/api --json defaultBranchRef"
	checkCmd = "gh pr checks 7 --repo acme/api --json name,state,bucket"
	green    = `[{"name":"build","state":"SUCCESS","bucket":"pass"},` +
		`{"name":"lint","state":"SUCCESS","bucket":"pass"},` +
		`{"name":"docs","state":"SKIPPED","bucket":"skipping"}]`
	workflowsPath = "repos/acme/api/contents/.github/workflows"
	rawHeader     = "Accept: application/vnd.github.raw+json"
	notFound      = "exit status 1: gh: Not Found (HTTP 404)"
	noCommit      = "exit status 1: gh: No commit found for the ref 0123456 (HTTP 404)"
	noChecks      = "no checks reported on the 'feat/endpoint' branch"
	token         = "gate:4a7010e4"
	other         = "gate:2cd50664"
	mergeCmd      = "gh pr merge 7 --repo acme/api --squash --match-head-commit " + headSHA
	auditBin      = homeDir + "/.claude/bin/audit"
	auditCmd      = auditBin + " acme/api merge https://github.com/acme/api/pull/7 " +
		"rangers gate: " + token + ", checks green, squash at 0123456"
)

type reply struct {
	out string
	err error
}

type fakeRunner struct {
	present map[string]bool
	replies map[string]reply
	calls   []string
}

type prFixture struct {
	state  string
	base   string
	title  string
	branch string
	head   string
	draft  bool
	labels []string
}

type scenario struct {
	t        *testing.T
	runner   *fakeRunner
	out      *bytes.Buffer
	projects string
	sessions map[string][]string
	ticks    int
	reads    int
	since    time.Time
}

func (f *fakeRunner) Has(name string) bool {
	return f.present[name]
}

func (f *fakeRunner) Output(_ context.Context, name string, args ...string) (string, error) {
	cmd := strings.Join(append([]string{name}, args...), " ")
	f.calls = append(f.calls, cmd)
	if r, ok := f.replies[cmd]; ok {
		return r.out, r.err
	}
	return "", errors.New("no reply for " + cmd)
}

func (f *fakeRunner) called(cmd string) int {
	count := 0
	for _, call := range f.calls {
		if call == cmd {
			count++
		}
	}
	return count
}

func openPR() prFixture {
	return prFixture{
		state:  "OPEN",
		base:   "main",
		title:  "feat(api): add the endpoint",
		branch: "feat/endpoint",
		head:   headSHA,
	}
}

func (p prFixture) json(t *testing.T) string {
	t.Helper()
	labels := make([]map[string]string, len(p.labels))
	for i, name := range p.labels {
		labels[i] = map[string]string{"name": name}
	}
	raw, err := json.Marshal(map[string]any{
		"state":       p.state,
		"isDraft":     p.draft,
		"baseRefName": p.base,
		"headRefOid":  p.head,
		"title":       p.title,
		"headRefName": p.branch,
		"labels":      labels,
	})
	if err != nil {
		t.Fatal(err)
	}
	return string(raw)
}

func newScenario(t *testing.T, pr prFixture) *scenario {
	t.Helper()
	runner := &fakeRunner{
		present: map[string]bool{},
		replies: map[string]reply{
			viewCmd:  {out: pr.json(t)},
			repoCmd:  {out: `{"defaultBranchRef":{"name":"main"}}`},
			checkCmd: {out: green},
			mergeCmd: {},
		},
	}
	return &scenario{
		t:        t,
		runner:   runner,
		out:      &bytes.Buffer{},
		projects: t.TempDir(),
		sessions: map[string][]string{},
	}
}

func clock() time.Time {
	return time.Date(2026, 11, 22, 12, 0, 0, 0, time.UTC)
}

func (s *scenario) run(extra ...string) error {
	s.t.Helper()
	args := append([]string{"--repo", repo, "--pr", number, "--transcripts", s.projects}, extra...)
	return s.command(args...)
}

func (s *scenario) runAsk() error {
	s.t.Helper()
	return s.command("ask", "--repo", repo, "--pr", number)
}

func (s *scenario) command(args ...string) error {
	s.t.Helper()
	s.flush()
	dep := &Dependencies{Runner: s.runner, Out: s.out, Lines: s.lines, Now: clock}
	env := func(string) string { return homeDir }
	return Command(s.t.Context(), args, env, dep)
}

func (s *scenario) lines(dir string, since time.Time) ([]transcript.Line, error) {
	s.reads++
	s.since = since
	return transcript.Conversation(dir, since)
}

func (s *scenario) say(session, role string, content any) {
	s.t.Helper()
	s.ticks++
	at := clock().Add(-2 * time.Hour).Add(time.Duration(s.ticks) * time.Minute)
	raw, err := json.Marshal(map[string]any{
		"type":      role,
		"timestamp": at.Format("2006-01-02T15:04:05.000Z"),
		"message":   map[string]any{"role": role, "content": content},
	})
	if err != nil {
		s.t.Fatal(err)
	}
	s.sessions[session] = append(s.sessions[session], string(raw))
}

func (s *scenario) ask(session, hash string) {
	s.t.Helper()
	s.say(session, "assistant", []map[string]string{
		{"type": "text", "text": "Merge acme/api#7 at 0123456? Answer sim. " + hash},
	})
}

func (s *scenario) answer(session, text string) {
	s.t.Helper()
	s.say(session, "user", text)
}

func (s *scenario) toolResult(session, text string) {
	s.t.Helper()
	s.say(session, "user", []map[string]string{
		{"type": "tool_result", "tool_use_id": "t1", "content": text},
	})
}

func (s *scenario) flush() {
	s.t.Helper()
	for session, rows := range s.sessions {
		file := filepath.Join(s.projects, "proj", session+".jsonl")
		if err := os.MkdirAll(filepath.Dir(file), 0o750); err != nil {
			s.t.Fatal(err)
		}
		if err := os.WriteFile(file, []byte(strings.Join(rows, "\n")+"\n"), 0o600); err != nil {
			s.t.Fatal(err)
		}
		if err := os.Chtimes(file, clock(), clock()); err != nil {
			s.t.Fatal(err)
		}
	}
}

func (s *scenario) printed() string {
	return s.out.String()
}

func (s *scenario) lastLine() string {
	lines := strings.Split(strings.TrimSpace(s.out.String()), "\n")
	return lines[len(lines)-1]
}

func listCmd(ref string) string {
	return "gh api " + workflowsPath + "?ref=" + ref
}

func fileCmd(ref, name string) string {
	return "gh api -H " + rawHeader + " " + workflowsPath + "/" + name + "?ref=" + ref
}

func (s *scenario) workflowsAt(ref string, files map[string]string) {
	s.t.Helper()
	entries := make([]map[string]string, 0, len(files))
	for name, body := range files {
		entries = append(entries, map[string]string{"name": name, "type": "file"})
		s.runner.replies[fileCmd(ref, name)] = reply{out: body}
	}
	raw, err := json.Marshal(entries)
	if err != nil {
		s.t.Fatal(err)
	}
	s.runner.replies[listCmd(ref)] = reply{out: string(raw)}
}

func (s *scenario) noWorkflowsAt(ref string) {
	s.t.Helper()
	s.runner.replies[listCmd(ref)] = reply{err: errors.New(listCmd(ref) + ": " + notFound)}
}

func (s *scenario) workflowsFail(ref string) {
	s.t.Helper()
	s.runner.replies[listCmd(ref)] = reply{err: errors.New(listCmd(ref) + ": exit status 1: gh: Bad Gateway (HTTP 502)")}
}

func (s *scenario) sameWorkflows(files map[string]string) {
	s.t.Helper()
	s.workflowsAt("main", files)
	s.workflowsAt(headSHA, files)
}

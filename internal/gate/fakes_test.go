package gate

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
)

const (
	repo    = "acme/api"
	number  = "7"
	headSHA = "0123456789abcdef0123456789abcdef01234567"
	homeDir = "/home/x"
	viewCmd = "gh pr view 7 --repo acme/api --json " + prFields
	repoCmd = "gh repo view acme/api --json defaultBranchRef"
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
	t      *testing.T
	runner *fakeRunner
	out    *bytes.Buffer
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
			viewCmd: {out: pr.json(t)},
			repoCmd: {out: `{"defaultBranchRef":{"name":"main"}}`},
		},
	}
	return &scenario{t: t, runner: runner, out: &bytes.Buffer{}}
}

func (s *scenario) run(extra ...string) error {
	s.t.Helper()
	args := append([]string{"--repo", repo, "--pr", number}, extra...)
	dep := &Dependencies{Runner: s.runner, Out: s.out}
	env := func(string) string { return homeDir }
	return Command(s.t.Context(), args, env, dep)
}

func (s *scenario) printed() string {
	return s.out.String()
}

func (s *scenario) lastLine() string {
	lines := strings.Split(strings.TrimSpace(s.out.String()), "\n")
	return lines[len(lines)-1]
}

package install

import (
	"bytes"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

const (
	sessionsSource    = repoDir + "/skills/claude-sessions"
	sessionsInstalled = claudeDir + "/skills/claude-sessions"
	sessionsBlocked   = "skills/claude-sessions exists and is not a link, left alone\n"
)

type realRepo struct {
	t      *testing.T
	repo   string
	claude string
	out    *bytes.Buffer
}

func withSessions(t *testing.T) *setup {
	t.Helper()
	s := newSetup(t)
	s.allInstalled()
	s.files.files[sessionsSource+"/SKILL.md"] = "---\nname: claude-sessions\n---\n"
	return s
}

func newRealRepo(t *testing.T) *realRepo {
	t.Helper()
	r := &realRepo{t: t, repo: t.TempDir(), claude: t.TempDir(), out: &bytes.Buffer{}}
	write(t, filepath.Join(r.repo, "deps.toml"), "format_version = 1\n")
	for _, name := range []string{"claude-sessions", "pr-description"} {
		write(t, filepath.Join(r.repo, "skills", name, "SKILL.md"), "name: "+name+"\n")
	}
	return r
}

func (r *realRepo) run(args ...string) error {
	runner := &fakeRunner{
		present: map[string]bool{},
		replies: map[string]reply{"docker info": {out: "Server: ok"}},
		streams: map[string]error{},
	}
	r.out.Reset()
	dep := &Dependencies{Runner: runner, Files: OSFiles{}, Out: r.out, Dir: r.repo}
	env := func(string) string { return "" }
	return Command(r.t.Context(), append([]string{"--claude-dir", r.claude}, args...), env, dep)
}

func (r *realRepo) installed(name string) string {
	return filepath.Join(r.claude, "skills", name)
}

func TestInstallSkillLinks(t *testing.T) {
	t.Run("given skill folders in the repository and none installed, when installing, then each one is linked into the claude dir skills", func(t *testing.T) {
		r := newRealRepo(t)
		if err := r.run(); err != nil {
			t.Fatal(err)
		}
		for _, name := range []string{"claude-sessions", "pr-description"} {
			target, err := os.Readlink(r.installed(name))
			if err != nil || target != filepath.Join(r.repo, "skills", name) {
				t.Fatalf("%s: got %q, %v", name, target, err)
			}
		}
		if !strings.Contains(r.out.String(), "2 linked into "+r.claude) {
			t.Fatalf("got\n%s", r.out.String())
		}
	})

	t.Run("given a real folder already in the place of a skill, when installing, then it is left alone, says so and the other skill is linked", func(t *testing.T) {
		r := newRealRepo(t)
		write(t, filepath.Join(r.installed("claude-sessions"), "SKILL.md"), "mine\n")
		if err := r.run(); err != nil {
			t.Fatal(err)
		}
		kept, err := os.ReadFile(filepath.Join(r.installed("claude-sessions"), "SKILL.md"))
		if err != nil || string(kept) != "mine\n" {
			t.Fatalf("got %q, %v", kept, err)
		}
		if _, err = os.Readlink(r.installed("pr-description")); err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(r.out.String(), sessionsBlocked) {
			t.Fatalf("got\n%s", r.out.String())
		}
		if !strings.Contains(r.out.String(), "1 linked into "+r.claude) {
			t.Fatalf("got\n%s", r.out.String())
		}
	})

	t.Run("given a regular file in the place of a skill, when installing, then it is left alone and says so", func(t *testing.T) {
		r := newRealRepo(t)
		write(t, r.installed("pr-description"), "mine\n")
		if err := r.run(); err != nil {
			t.Fatal(err)
		}
		kept, err := os.ReadFile(r.installed("pr-description"))
		if err != nil || string(kept) != "mine\n" {
			t.Fatalf("got %q, %v", kept, err)
		}
		if !strings.Contains(r.out.String(), "skills/pr-description exists and is not a link, left alone\n") {
			t.Fatalf("got\n%s", r.out.String())
		}
	})

	t.Run("given a link to another place, when installing, then it is pointed at the repository", func(t *testing.T) {
		r := newRealRepo(t)
		if err := os.MkdirAll(filepath.Dir(r.installed("claude-sessions")), 0o750); err != nil {
			t.Fatal(err)
		}
		if err := os.Symlink("/elsewhere", r.installed("claude-sessions")); err != nil {
			t.Fatal(err)
		}
		if err := r.run(); err != nil {
			t.Fatal(err)
		}
		target, err := os.Readlink(r.installed("claude-sessions"))
		if err != nil || target != filepath.Join(r.repo, "skills", "claude-sessions") {
			t.Fatalf("got %q, %v", target, err)
		}
	})

	t.Run("given the skill links already made, when installing again, then nothing changes", func(t *testing.T) {
		s := withSessions(t)
		if err := s.run(); err != nil {
			t.Fatal(err)
		}
		made := slices.Clone(s.files.symlinked)
		if err := s.run(); err != nil {
			t.Fatal(err)
		}
		if !slices.Equal(s.files.symlinked, made) || s.files.links[sessionsInstalled] != sessionsSource {
			t.Fatalf("symlinked %v, links %v", s.files.symlinked, s.files.links)
		}
		if strings.Contains(s.typed(), "left alone") {
			t.Fatalf("got\n%s", s.typed())
		}
	})

	t.Run("given a folder in skills without SKILL.md, when installing, then it is not linked", func(t *testing.T) {
		s := withSessions(t)
		s.files.files[repoDir+"/skills/draft/notes.md"] = "x"
		if err := s.run(); err != nil {
			t.Fatal(err)
		}
		if _, ok := s.files.links[claudeDir+"/skills/draft"]; ok {
			t.Fatalf("got %v", s.files.links)
		}
	})
}

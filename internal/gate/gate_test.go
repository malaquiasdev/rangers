package gate

import (
	"errors"
	"os"
	"strings"
	"testing"
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

func TestOpenPullRequest(t *testing.T) {
	cases := map[string]struct {
		change func(*prFixture)
		reason string
	}{
		"a draft":         {func(p *prFixture) { p.draft = true }, "the pull request is a draft"},
		"closed":          {func(p *prFixture) { p.state = "CLOSED" }, "the pull request is closed"},
		"merged":          {func(p *prFixture) { p.state = "MERGED" }, "the pull request is merged"},
		"on another base": {func(p *prFixture) { p.base = "develop" }, "base is develop, not the default branch main"},
	}
	for name, c := range cases {
		t.Run("given a pull request that is "+name+", when running the gate, then it refuses with the reason", func(t *testing.T) {
			pr := openPR()
			c.change(&pr)
			s := newScenario(t, pr)
			err := s.run()
			if !errors.Is(err, ErrRefused) {
				t.Fatalf("got %v", err)
			}
			if got, want := s.lastLine(), "refused: "+c.reason; got != want {
				t.Fatalf("got %q, want %q", got, want)
			}
		})
	}

	t.Run("given gh failing, when running the gate, then it returns the error and does not refuse", func(t *testing.T) {
		s := newScenario(t, openPR())
		s.runner.replies[viewCmd] = reply{err: errors.New("gh: not found")}
		err := s.run()
		if err == nil || errors.Is(err, ErrRefused) {
			t.Fatalf("got %v", err)
		}
		if s.printed() != "" {
			t.Fatalf("got %q", s.printed())
		}
	})
}

func TestProductionStop(t *testing.T) {
	const verdict = "refused: production needs the human's explicit yes; the gate stops here"
	cases := map[string]func(*prFixture){
		"a title that starts with chore(main): release": func(p *prFixture) { p.title = "chore(main): release 1.4.0" },
		"a release-please head branch":                  func(p *prFixture) { p.branch = "release-please--branches--main" },
		"a production label":                            func(p *prFixture) { p.labels = []string{"bug", "Production"} },
	}
	for name, change := range cases {
		t.Run("given "+name+", when running the gate, then it stops at production", func(t *testing.T) {
			pr := openPR()
			change(&pr)
			s := newScenario(t, pr)
			if err := s.run(); !errors.Is(err, ErrRefused) {
				t.Fatalf("got %v", err)
			}
			if got := s.lastLine(); got != verdict {
				t.Fatalf("got %q", got)
			}
			if !strings.Contains(s.printed(), "production: refused") {
				t.Fatalf("got\n%s", s.printed())
			}
		})
	}
}

func TestOptions(t *testing.T) {
	cases := map[string][]string{
		"no repo":              {"--pr", "7"},
		"a repo without owner": {"--repo", "api", "--pr", "7"},
		"no pull request":      {"--repo", repo},
		"a stray word":         {"--repo", repo, "--pr", "7", "now"},
		"an unknown flag":      {"--repo", repo, "--pr", "7", "--force"},
	}
	for name, args := range cases {
		t.Run("given "+name+", when running the gate, then it fails with an error and calls nothing", func(t *testing.T) {
			s := newScenario(t, openPR())
			dep := &Dependencies{Runner: s.runner, Out: s.out}
			err := Command(t.Context(), args, func(string) string { return homeDir }, dep)
			if err == nil || errors.Is(err, ErrRefused) {
				t.Fatalf("got %v", err)
			}
			if len(s.runner.calls) != 0 {
				t.Fatalf("got %v", s.runner.calls)
			}
		})
	}
}

func TestOutputLimit(t *testing.T) {
	t.Run("given lines longer than the limit, when fitting, then the text has at most 560 characters and keeps every line", func(t *testing.T) {
		lines := []string{"pr: ok", "ci: " + strings.Repeat("x", 900), "refused: ci"}
		text := fit(lines, outputLimit)
		if n := len([]rune(text)); n > outputLimit {
			t.Fatalf("got %d characters", n)
		}
		if got := strings.Split(text, "\n"); len(got) != 3 || got[2] != "refused: ci" {
			t.Fatalf("got %q", text)
		}
	})

	t.Run("given lines inside the limit, when fitting, then it changes nothing", func(t *testing.T) {
		if got := fit([]string{"a", "b"}, outputLimit); got != "a\nb" {
			t.Fatalf("got %q", got)
		}
	})
}

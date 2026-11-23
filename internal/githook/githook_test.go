package githook

import (
	"bytes"
	"context"
	"errors"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/malaquiasdev/rangers/internal/check"
)

const (
	repo          = "/repo"
	namesText     = "Foo|Bar Baz\n"
	formatFailure = "a mensagem precisa seguir o Conventional Commits\n"
	nameFailure   = "nome de personagem na mensagem do commit\n"
	installed     = "ganchos ligados: pre-commit, commit-msg e pre-push\n"
)

type memFiles struct {
	files      map[string]string
	executable []string
	chmodErr   error
}

type recordingGit struct {
	calls []string
	err   error
}

type quietRunner struct {
	calls []string
}

type harness struct {
	files  *memFiles
	git    *recordingGit
	runner *quietRunner
	out    bytes.Buffer
	root   fstest.MapFS
}

func TestMain(m *testing.M) {
	for _, pair := range os.Environ() {
		name, _, _ := strings.Cut(pair, "=")
		if strings.HasPrefix(name, "GIT_") {
			_ = os.Unsetenv(name)
		}
	}
	os.Exit(m.Run())
}

func (m *memFiles) ReadFile(path string) ([]byte, error) {
	if text, ok := m.files[path]; ok {
		return []byte(text), nil
	}
	return nil, fs.ErrNotExist
}

func (m *memFiles) Glob(pattern string) ([]string, error) {
	var matches []string
	for path := range m.files {
		if ok, _ := filepath.Match(pattern, path); ok {
			matches = append(matches, path)
		}
	}
	return matches, nil
}

func (m *memFiles) MakeExecutable(path string) error {
	if m.chmodErr != nil {
		return m.chmodErr
	}
	m.executable = append(m.executable, path)
	return nil
}

func (g *recordingGit) SetConfig(_ context.Context, key, value string) error {
	g.calls = append(g.calls, key+"="+value)
	return g.err
}

func (r *quietRunner) Run(_ context.Context, call check.Invocation) error {
	r.calls = append(r.calls, call.Name)
	return nil
}

func (r *quietRunner) Output(_ context.Context, name string, _ ...string) (string, error) {
	r.calls = append(r.calls, name)
	return "", nil
}

func newHarness() *harness {
	h := &harness{
		files:  &memFiles{files: map[string]string{repo + "/scripts/characters.txt": namesText}},
		git:    &recordingGit{},
		runner: &quietRunner{},
		root: fstest.MapFS{
			"deps.toml":              {Data: []byte("format_version = 1\n")},
			"scripts/characters.txt": {Data: []byte(namesText)},
		},
	}
	return h
}

func (h *harness) run(t *testing.T, args ...string) error {
	t.Helper()
	dep := &Dependencies{
		Check: &check.Dependencies{
			Dir:      repo,
			Root:     h.root,
			Runner:   h.runner,
			ReadHost: func(string) ([]byte, error) { return nil, fs.ErrNotExist },
			Out:      &h.out,
		},
		Files: h.files,
		Git:   h.git,
		Out:   &h.out,
		Dir:   repo,
	}
	return Command(t.Context(), args, func(string) string { return "" }, dep)
}

func (h *harness) message(text string) string {
	h.files.files["COMMIT_EDITMSG"] = text
	return "COMMIT_EDITMSG"
}

func TestCommitMessageFormat(t *testing.T) {
	accepted := []string{
		"feat: add a thing",
		"fix(parser): handle empty input",
		"docs(read-me2)!: rewrite",
		"chore!: drop the old flag",
		"refactor: split it\n\nbody line",
		"test: cover it",
		"build: bump",
		"ci: pin",
		"perf: faster",
		"style: format",
		"revert: undo",
		"feat: a subject\nthe second line is free",
	}
	for _, text := range accepted {
		t.Run("given the message "+strings.SplitN(text, "\n", 2)[0]+", when committing, then it is accepted silently", func(t *testing.T) {
			h := newHarness()
			if err := h.run(t, "run", "commit-msg", h.message(text)); err != nil {
				t.Fatal(err)
			}
			if h.out.Len() != 0 {
				t.Fatalf("printed %q", h.out.String())
			}
		})
	}

	refused := map[string]string{
		"an unknown type":           "wip: something",
		"an upper case type":        "Feat: something",
		"no space after the colon":  "feat:something",
		"nothing after the colon":   "feat: ",
		"an upper case scope":       "feat(Scope): something",
		"an empty scope":            "feat(): something",
		"no type":                   "update stuff",
		"an empty file":             "",
		"a blank first line":        "\nfeat: something",
		"the subject on line two":   "wip\nfeat: something",
		"a comment line first":      "# explain\nfeat: something",
		"a scope with a underscore": "feat(a_b): something",
	}
	for name, text := range refused {
		t.Run("given "+name+", when committing, then it is refused with the Conventional Commits message", func(t *testing.T) {
			h := newHarness()
			err := h.run(t, "run", "commit-msg", h.message(text))
			if !errors.Is(err, ErrRefused) {
				t.Fatalf("got %v", err)
			}
			if h.out.String() != formatFailure {
				t.Fatalf("got %q", h.out.String())
			}
		})
	}
}

func TestCommitMessageNames(t *testing.T) {
	refused := map[string]string{
		"the subject":            "feat: thank Foo",
		"the body":               "feat: fine\n\nreviewed by Bar Baz",
		"a comment line":         "feat: fine\n# Foo was here",
		"punctuation around it":  "feat: fine (Foo)",
		"the last line, no EOL":  "feat: fine\nFoo",
		"the name by itself":     "feat: Foo",
		"the name in the middle": "feat: a Foo b",
	}
	for name, text := range refused {
		t.Run("given a name in "+name+", when committing, then it is refused with the names message", func(t *testing.T) {
			h := newHarness()
			err := h.run(t, "run", "commit-msg", h.message(text))
			if !errors.Is(err, ErrRefused) {
				t.Fatalf("got %v", err)
			}
			if h.out.String() != nameFailure {
				t.Fatalf("got %q", h.out.String())
			}
		})
	}

	t.Run("given a name inside a longer word, when committing, then it is accepted", func(t *testing.T) {
		h := newHarness()
		if err := h.run(t, "run", "commit-msg", h.message("feat: Foolish and Food")); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("given a bad format and a name, when committing, then only the format message is printed", func(t *testing.T) {
		h := newHarness()
		err := h.run(t, "run", "commit-msg", h.message("Foo did it"))
		if !errors.Is(err, ErrRefused) || h.out.String() != formatFailure {
			t.Fatalf("got %v and %q", err, h.out.String())
		}
	})

	t.Run("given no names file, when committing, then it fails with an error and prints nothing", func(t *testing.T) {
		h := newHarness()
		delete(h.files.files, repo+"/scripts/characters.txt")
		err := h.run(t, "run", "commit-msg", h.message("feat: fine"))
		if err == nil || errors.Is(err, ErrRefused) || h.out.Len() != 0 {
			t.Fatalf("got %v and %q", err, h.out.String())
		}
	})

	t.Run("given a message file that does not exist, when committing, then it fails with an error", func(t *testing.T) {
		h := newHarness()
		if err := h.run(t, "run", "commit-msg", "missing"); err == nil || errors.Is(err, ErrRefused) {
			t.Fatalf("got %v", err)
		}
	})

	t.Run("given commit-msg with no file, when running, then it is a usage error", func(t *testing.T) {
		if err := newHarness().run(t, "run", "commit-msg"); !errors.Is(err, ErrUsage) {
			t.Fatalf("got %v", err)
		}
	})
}

func TestRunChecks(t *testing.T) {
	t.Run("given pre-commit, when running, then it runs the quick checks", func(t *testing.T) {
		h := newHarness()
		if err := h.run(t, "run", "pre-commit"); err != nil {
			t.Fatal(err)
		}
		if got := h.out.String(); !strings.HasSuffix(got, "tudo certo (quick)\n") {
			t.Fatalf("got %q", got)
		}
		if len(h.runner.calls) != 0 {
			t.Fatalf("ran %v", h.runner.calls)
		}
	})

	t.Run("given pre-push, when running, then it runs the full checks", func(t *testing.T) {
		h := newHarness()
		if err := h.run(t, "run", "pre-push"); err != nil {
			t.Fatal(err)
		}
		if got := h.out.String(); !strings.HasSuffix(got, "tudo certo (full)\n") {
			t.Fatalf("got %q", got)
		}
		if len(h.runner.calls) == 0 {
			t.Fatal("the slow tools did not run")
		}
	})

	t.Run("given a failing check, when running pre-commit, then the failure error comes back", func(t *testing.T) {
		h := newHarness()
		h.root["skills/pilot/SKILL.md"] = &fstest.MapFile{Data: []byte("name: other\n")}
		if err := h.run(t, "run", "pre-commit"); !errors.Is(err, check.ErrFailed) {
			t.Fatalf("got %v", err)
		}
	})
}

func TestUsage(t *testing.T) {
	tests := map[string][]string{
		"no arguments":        nil,
		"an unknown command":  {"remove"},
		"run without a name":  {"run"},
		"an unknown hook":     {"run", "post-merge"},
		"install with extras": {"install", "now"},
	}
	for name, args := range tests {
		t.Run("given "+name+", when running, then it is a usage error", func(t *testing.T) {
			if err := newHarness().run(t, args...); !errors.Is(err, ErrUsage) {
				t.Fatalf("got %v", err)
			}
		})
	}
}

func TestInstall(t *testing.T) {
	t.Run("given the hooks of the repository, when installing, then they become executable, Git is pointed to them and the line is printed", func(t *testing.T) {
		h := newHarness()
		for _, name := range []string{"pre-commit", "commit-msg", "pre-push"} {
			h.files.files[repo+"/.githooks/"+name] = "#!/usr/bin/env bash\n"
		}
		if err := h.run(t, "install"); err != nil {
			t.Fatal(err)
		}
		if len(h.files.executable) != 3 {
			t.Fatalf("got %v", h.files.executable)
		}
		if len(h.git.calls) != 1 || h.git.calls[0] != "core.hooksPath=.githooks" {
			t.Fatalf("got %v", h.git.calls)
		}
		if h.out.String() != installed {
			t.Fatalf("got %q", h.out.String())
		}
	})

	t.Run("given a hook that cannot be made executable, when installing, then it fails before touching the Git config", func(t *testing.T) {
		h := newHarness()
		h.files.files[repo+"/.githooks/pre-commit"] = "x"
		h.files.chmodErr = fs.ErrPermission
		if err := h.run(t, "install"); !errors.Is(err, fs.ErrPermission) {
			t.Fatalf("got %v", err)
		}
		if len(h.git.calls) != 0 || h.out.Len() != 0 {
			t.Fatalf("got %v and %q", h.git.calls, h.out.String())
		}
	})

	t.Run("given Git failing to set the config, when installing, then it fails without the success line", func(t *testing.T) {
		h := newHarness()
		h.git.err = errors.New("not a repository")
		if err := h.run(t, "install"); err == nil || h.out.Len() != 0 {
			t.Fatalf("got %v and %q", err, h.out.String())
		}
	})
}

func TestOSFiles(t *testing.T) {
	t.Run("given a file without the execute bits, when making it executable, then the owner can run it", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "hook")
		if err := os.WriteFile(path, []byte("x"), 0o600); err != nil {
			t.Fatal(err)
		}
		if err := (OSFiles{}).MakeExecutable(path); err != nil {
			t.Fatal(err)
		}
		info, err := os.Stat(path)
		if err != nil || info.Mode().Perm()&0o100 == 0 {
			t.Fatalf("got %v, %v", info, err)
		}
	})

	t.Run("given a missing file, when making it executable, then it fails", func(t *testing.T) {
		if err := (OSFiles{}).MakeExecutable(filepath.Join(t.TempDir(), "missing")); err == nil {
			t.Fatal("expected an error")
		}
	})

	t.Run("given files in a folder, when reading and globbing, then it returns them", func(t *testing.T) {
		dir := t.TempDir()
		if err := os.WriteFile(filepath.Join(dir, "a"), []byte("one"), 0o600); err != nil {
			t.Fatal(err)
		}
		files := OSFiles{}
		matches, err := files.Glob(filepath.Join(dir, "*"))
		if err != nil || len(matches) != 1 {
			t.Fatalf("got %v, %v", matches, err)
		}
		if raw, err := files.ReadFile(matches[0]); err != nil || string(raw) != "one" {
			t.Fatalf("got %q, %v", raw, err)
		}
	})
}

func TestGitCLI(t *testing.T) {
	t.Run("given a repository, when setting a config key, then Git stores it", func(t *testing.T) {
		dir := t.TempDir()
		if out, err := exec.CommandContext(t.Context(), "git", "-C", dir, "init", "-q").CombinedOutput(); err != nil {
			t.Fatalf("git init: %v: %s", err, out)
		}
		if err := (GitCLI{Dir: dir}).SetConfig(t.Context(), "core.hooksPath", ".githooks"); err != nil {
			t.Fatal(err)
		}
		out, err := exec.CommandContext(t.Context(), "git", "-C", dir, "config", "core.hooksPath").Output()
		if err != nil || strings.TrimSpace(string(out)) != ".githooks" {
			t.Fatalf("got %q, %v", out, err)
		}
	})

	t.Run("given a folder that is not a repository, when setting a config key, then it fails", func(t *testing.T) {
		if err := (GitCLI{Dir: t.TempDir()}).SetConfig(t.Context(), "core.hooksPath", ".githooks"); err == nil {
			t.Fatal("expected an error")
		}
	})
}

package agentwatch

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

var base = time.Unix(1_800_000_000, 0)

type fakeGit struct {
	commits map[string]int64
}

func (f fakeGit) LastCommit(_ context.Context, dir string) (int64, bool) {
	stamp, ok := f.commits[dir]
	return stamp, ok
}

type harness struct {
	t       *testing.T
	root    string
	list    string
	git     fakeGit
	env     map[string]string
	sleeps  []time.Duration
	passes  int
	onSleep func(pass int)
	out     bytes.Buffer
}

func newHarness(t *testing.T) *harness {
	t.Helper()
	h := &harness{t: t, root: t.TempDir(), git: fakeGit{commits: map[string]int64{}}, passes: 2, env: map[string]string{}}
	h.list = filepath.Join(h.root, "agent-watch.txt")
	return h
}

func (h *harness) file(name string, age time.Duration) string {
	h.t.Helper()
	path := filepath.Join(h.root, name)
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		h.t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("x"), 0o600); err != nil {
		h.t.Fatal(err)
	}
	when := base.Add(-age)
	if err := os.Chtimes(path, when, when); err != nil {
		h.t.Fatal(err)
	}
	return path
}

func (h *harness) link(name, target string, age time.Duration) string {
	h.t.Helper()
	path := filepath.Join(h.root, name)
	if err := os.Symlink(target, path); err != nil {
		h.t.Fatal(err)
	}
	stamp := base.Add(-age).Format("200601021504")
	if out, err := exec.CommandContext(context.Background(), "touch", "-h", "-t", stamp, path).CombinedOutput(); err != nil {
		h.t.Fatalf("touch -h: %v: %s", err, out)
	}
	return path
}

func (h *harness) dir(name string) string {
	h.t.Helper()
	path := filepath.Join(h.root, name)
	if err := os.MkdirAll(path, 0o700); err != nil {
		h.t.Fatal(err)
	}
	return path
}

func (h *harness) write(content string) {
	h.t.Helper()
	if err := os.WriteFile(h.list, []byte(content), 0o600); err != nil {
		h.t.Fatal(err)
	}
}

func (h *harness) entries(lines ...string) {
	h.t.Helper()
	h.write(strings.Join(lines, "\n") + "\n")
}

func (h *harness) run(args ...string) (string, error) {
	h.t.Helper()
	ctx, cancel := context.WithCancel(h.t.Context())
	defer cancel()
	env := map[string]string{"AGENT_WATCH_LIST": h.list}
	for name, value := range h.env {
		env[name] = value
	}
	dep := &Dependencies{
		Git: h.git,
		Out: &h.out,
		Now: func() time.Time { return base },
		Sleep: func(_ context.Context, d time.Duration) error {
			h.sleeps = append(h.sleeps, d)
			if h.onSleep != nil {
				h.onSleep(len(h.sleeps))
			}
			if len(h.sleeps) >= h.passes {
				cancel()
			}
			return ctx.Err()
		},
	}
	err := Command(ctx, args, func(name string) string { return env[name] }, dep)
	return h.out.String(), err
}

func (h *harness) reports(want string, args ...string) {
	h.t.Helper()
	out, err := h.run(args...)
	if err != nil {
		h.t.Fatalf("error %v, output %q", err, out)
	}
	if out != want {
		h.t.Fatalf("got %q, want %q", out, want)
	}
	if len(h.sleeps) != 0 {
		h.t.Fatalf("slept %d times after reporting", len(h.sleeps))
	}
}

func (h *harness) keepsWatching(args ...string) {
	h.t.Helper()
	out, err := h.run(args...)
	if !errors.Is(err, context.Canceled) {
		h.t.Fatalf("error %v, output %q", err, out)
	}
	if out != "" {
		h.t.Fatalf("reported %q", out)
	}
}

func stopped(name string, age time.Duration, last int64) string {
	return fmt.Sprintf("PARADO %s há %d min (última atividade: %s)\n", name, int(age.Minutes()), time.Unix(last, 0).Format("15:04"))
}

func TestList(t *testing.T) {
	t.Run("given a missing list, when watching, then it says the list is empty and exits", func(t *testing.T) {
		h := newHarness(t)
		h.reports("lista vazia\n")
	})

	t.Run("given an empty list, when watching, then it says the list is empty and exits", func(t *testing.T) {
		h := newHarness(t)
		h.write("")
		h.reports("lista vazia\n")
	})

	t.Run("given a list with only comments and blank lines, when watching, then it keeps watching", func(t *testing.T) {
		h := newHarness(t)
		h.write("# zack /tmp/x\n\n   \n")
		h.keepsWatching()
	})

	t.Run("given a commented entry before a stale one, when watching, then it skips the comment", func(t *testing.T) {
		h := newHarness(t)
		stale := h.file("stale.jsonl", time.Hour)
		h.entries("# billy /nowhere", "zack "+stale)
		h.reports(stopped("zack", time.Hour, base.Add(-time.Hour).Unix()))
	})

	t.Run("given the last line without a newline, when watching, then it still checks it", func(t *testing.T) {
		h := newHarness(t)
		stale := h.file("stale.jsonl", time.Hour)
		h.write("zack " + stale)
		h.reports(stopped("zack", time.Hour, base.Add(-time.Hour).Unix()))
	})

	t.Run("given a list path in the environment, when watching, then it reads that list", func(t *testing.T) {
		h := newHarness(t)
		stale := h.file("stale.jsonl", time.Hour)
		other := filepath.Join(h.root, "other.txt")
		if err := os.WriteFile(other, []byte("zack "+stale+"\n"), 0o600); err != nil {
			t.Fatal(err)
		}
		h.env["AGENT_WATCH_LIST"] = other
		h.reports(stopped("zack", time.Hour, base.Add(-time.Hour).Unix()))
	})

	t.Run("given no list path in the environment, when watching, then it reads agent-watch.txt in the Claude folder of the home", func(t *testing.T) {
		home := t.TempDir()
		folder := filepath.Join(home, ".claude")
		if err := os.MkdirAll(folder, 0o700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(folder, "agent-watch.txt"), []byte("zack /nowhere\n"), 0o600); err != nil {
			t.Fatal(err)
		}
		var out bytes.Buffer
		dep := &Dependencies{Git: fakeGit{}, Out: &out, Now: func() time.Time { return base }}
		err := Command(t.Context(), nil, func(name string) string { return map[string]string{"HOME": home}[name] }, dep)
		if err != nil || out.String() != "SUMIU zack\n" {
			t.Fatalf("error %v, output %q", err, out.String())
		}
	})

	t.Run("given a changed list, when it loops again, then it reads the list again", func(t *testing.T) {
		h := newHarness(t)
		fresh := h.file("fresh.jsonl", time.Minute)
		stale := h.file("stale.jsonl", time.Hour)
		h.entries("zack " + fresh)
		h.onSleep = func(int) { h.entries("zack "+fresh, "billy "+stale) }
		h.passes = 5
		out, err := h.run()
		if err != nil || out != stopped("billy", time.Hour, base.Add(-time.Hour).Unix()) {
			t.Fatalf("error %v, output %q", err, out)
		}
		if len(h.sleeps) != 1 {
			t.Fatalf("slept %d times", len(h.sleeps))
		}
	})
}

func TestVerdict(t *testing.T) {
	tests := []struct {
		name  string
		age   time.Duration
		args  []string
		check func(h *harness, file string)
	}{
		{"given a file written a minute ago, when watching, then it keeps watching", time.Minute, nil, func(h *harness, _ string) { h.keepsWatching() }},
		{"given a file idle for exactly the limit, when watching, then it keeps watching", 900 * time.Second, nil, func(h *harness, _ string) { h.keepsWatching() }},
		{"given a file idle one second over the limit, when watching, then it reports it idle", 901 * time.Second, nil, func(h *harness, _ string) {
			h.reports(stopped("zack", 901*time.Second, base.Add(-901*time.Second).Unix()))
		}},
		{"given a file idle for an hour, when watching, then it reports the idle minutes", time.Hour, nil, func(h *harness, _ string) {
			h.reports(stopped("zack", time.Hour, base.Add(-time.Hour).Unix()))
		}},
		{"given a limit of 60 seconds and a file idle for 2 minutes, when watching, then it reports it idle", 2 * time.Minute, []string{"60"}, func(h *harness, _ string) {
			h.reports(stopped("zack", 2*time.Minute, base.Add(-2*time.Minute).Unix()), "60")
		}},
		{"given a limit of 0 seconds and a file idle for a minute, when watching, then it reports it idle", time.Minute, []string{"0"}, func(h *harness, _ string) {
			h.reports(stopped("zack", time.Minute, base.Add(-time.Minute).Unix()), "0")
		}},
		{"given a limit of 7200 seconds and a file idle for an hour, when watching, then it keeps watching", time.Hour, []string{"7200"}, func(h *harness, _ string) { h.keepsWatching("7200") }},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			h := newHarness(t)
			file := h.file("transcript.jsonl", tc.age)
			h.entries("zack " + file)
			tc.check(h, file)
		})
	}
}

func TestUsage(t *testing.T) {
	for _, args := range [][]string{{"soon"}, {"-5"}, {"60", "extra"}} {
		t.Run(fmt.Sprintf("given the arguments %q, when watching, then it returns the usage error", args), func(t *testing.T) {
			h := newHarness(t)
			h.entries("zack " + h.file("transcript.jsonl", time.Minute))
			out, err := h.run(args...)
			if !errors.Is(err, ErrUsage) || out != "" {
				t.Fatalf("error %v, output %q", err, out)
			}
		})
	}
}

func TestInterval(t *testing.T) {
	tests := []struct {
		name  string
		value string
		want  time.Duration
	}{
		{"given no interval variable, when it loops, then it sleeps 120 seconds", "", 120 * time.Second},
		{"given an interval of 5 seconds, when it loops, then it sleeps 5 seconds", "5", 5 * time.Second},
		{"given a fractional interval, when it loops, then it sleeps that fraction", "0.5", 500 * time.Millisecond},
		{"given an interval that is not a number, when it loops, then it sleeps 120 seconds", "soon", 120 * time.Second},
		{"given an interval of zero, when it loops, then it sleeps 120 seconds", "0", 120 * time.Second},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			h := newHarness(t)
			h.entries("zack " + h.file("transcript.jsonl", time.Minute))
			h.env["AGENT_WATCH_INTERVAL"] = tc.value
			h.keepsWatching()
			if len(h.sleeps) == 0 || h.sleeps[0] != tc.want {
				t.Fatalf("slept %v, want %s", h.sleeps, tc.want)
			}
		})
	}
}

func TestGone(t *testing.T) {
	t.Run("given a file that does not exist and no directory, when watching, then it reports it gone", func(t *testing.T) {
		h := newHarness(t)
		h.entries("zack " + filepath.Join(h.root, "missing.jsonl"))
		h.reports("SUMIU zack\n")
	})

	t.Run("given an entry with only a name, when watching, then it reports it gone", func(t *testing.T) {
		h := newHarness(t)
		h.entries("zack")
		h.reports("SUMIU zack\n")
	})

	t.Run("given a missing file and a directory with no files, when watching, then it reports it gone", func(t *testing.T) {
		h := newHarness(t)
		dir := h.dir("worktree")
		h.entries("zack " + filepath.Join(h.root, "missing.jsonl") + " " + dir)
		h.reports("SUMIU zack\n")
	})

	t.Run("given a missing file and a missing directory, when watching, then it reports it gone", func(t *testing.T) {
		h := newHarness(t)
		h.entries("zack " + filepath.Join(h.root, "missing.jsonl") + " " + filepath.Join(h.root, "nowhere"))
		h.reports("SUMIU zack\n")
	})

	t.Run("given a directory path in the file field and no directory list, when watching, then it reports it gone", func(t *testing.T) {
		h := newHarness(t)
		h.entries("zack " + h.dir("worktree"))
		h.reports("SUMIU zack\n")
	})
}

func TestSymlinks(t *testing.T) {
	t.Run("given a symlink with an old stamp to a fresh transcript, when watching, then it keeps watching", func(t *testing.T) {
		h := newHarness(t)
		target := h.file("transcript.jsonl", time.Minute)
		link := h.link("task.output", target, 2*time.Hour)
		h.entries("zack " + link)
		h.keepsWatching()
	})

	t.Run("given a symlink with a fresh stamp to a stale transcript, when watching, then it reports the transcript age", func(t *testing.T) {
		h := newHarness(t)
		target := h.file("transcript.jsonl", time.Hour)
		link := h.link("task.output", target, time.Minute)
		h.entries("zack " + link)
		h.reports(stopped("zack", time.Hour, base.Add(-time.Hour).Unix()))
	})

	t.Run("given a symlink to a missing transcript and no directory, when watching, then it reports it gone", func(t *testing.T) {
		h := newHarness(t)
		link := h.link("task.output", filepath.Join(h.root, "missing.jsonl"), time.Minute)
		h.entries("zack " + link)
		h.reports("SUMIU zack\n")
	})

	t.Run("given a symlink to a missing transcript and a directory with no files, when watching, then it reports it gone instead of an idle time", func(t *testing.T) {
		h := newHarness(t)
		link := h.link("task.output", filepath.Join(h.root, "missing.jsonl"), time.Minute)
		h.entries("zack " + link + " " + h.dir("worktree"))
		h.reports("SUMIU zack\n")
	})
}

func TestDirectories(t *testing.T) {
	t.Run("given a stale file and a directory with a fresh file, when watching, then it keeps watching", func(t *testing.T) {
		h := newHarness(t)
		stale := h.file("transcript.jsonl", time.Hour)
		h.file("worktree/src/code.go", time.Minute)
		h.entries("zack " + stale + " " + filepath.Join(h.root, "worktree"))
		h.keepsWatching()
	})

	t.Run("given a fresh file and a directory with only stale files, when watching, then it keeps watching", func(t *testing.T) {
		h := newHarness(t)
		fresh := h.file("transcript.jsonl", time.Minute)
		h.file("worktree/code.go", 3*time.Hour)
		h.entries("zack " + fresh + " " + filepath.Join(h.root, "worktree"))
		h.keepsWatching()
	})

	t.Run("given a missing file and a directory with a stale file, when watching, then it reports the directory age", func(t *testing.T) {
		h := newHarness(t)
		h.file("worktree/code.go", time.Hour)
		h.entries("zack " + filepath.Join(h.root, "missing.jsonl") + " " + filepath.Join(h.root, "worktree"))
		h.reports(stopped("zack", time.Hour, base.Add(-time.Hour).Unix()))
	})

	t.Run("given several directories, when watching, then it uses the newest file of all of them", func(t *testing.T) {
		h := newHarness(t)
		stale := h.file("transcript.jsonl", 3*time.Hour)
		h.file("a/code.go", 2*time.Hour)
		h.file("b/deep/er/code.go", time.Minute)
		h.entries("zack " + stale + " " + filepath.Join(h.root, "a") + "   " + filepath.Join(h.root, "b"))
		h.keepsWatching()
	})

	for _, ignored := range []string{"node_modules", ".git", "vendor"} {
		t.Run("given fresh files only inside "+ignored+", when watching, then it ignores them", func(t *testing.T) {
			h := newHarness(t)
			h.file("worktree/code.go", time.Hour)
			h.file("worktree/"+ignored+"/pkg/fresh.js", time.Minute)
			h.entries("zack " + filepath.Join(h.root, "missing.jsonl") + " " + filepath.Join(h.root, "worktree"))
			h.reports(stopped("zack", time.Hour, base.Add(-time.Hour).Unix()))
		})
	}

	t.Run("given a fresh file that is called .git, when watching, then it ignores it as a worktree pointer", func(t *testing.T) {
		h := newHarness(t)
		h.file("worktree/code.go", time.Hour)
		h.file("worktree/.git", time.Minute)
		h.entries("zack " + filepath.Join(h.root, "missing.jsonl") + " " + filepath.Join(h.root, "worktree"))
		h.reports(stopped("zack", time.Hour, base.Add(-time.Hour).Unix()))
	})

	t.Run("given a fresh symlink inside the directory, when watching, then it does not count as a file", func(t *testing.T) {
		h := newHarness(t)
		h.file("worktree/code.go", time.Hour)
		h.link("worktree/latest", h.file("outside.txt", time.Hour), time.Minute)
		h.entries("zack " + filepath.Join(h.root, "missing.jsonl") + " " + filepath.Join(h.root, "worktree"))
		h.reports(stopped("zack", time.Hour, base.Add(-time.Hour).Unix()))
	})
}

func TestGitActivity(t *testing.T) {
	t.Run("given a commit newer than the files, when watching, then the commit counts as activity", func(t *testing.T) {
		h := newHarness(t)
		dir := h.dir("worktree")
		h.file("worktree/code.go", 3*time.Hour)
		h.git.commits[dir] = base.Add(-time.Minute).Unix()
		h.entries("zack " + filepath.Join(h.root, "missing.jsonl") + " " + dir)
		h.keepsWatching()
	})

	t.Run("given a commit older than the files, when watching, then the files decide", func(t *testing.T) {
		h := newHarness(t)
		dir := h.dir("worktree")
		h.file("worktree/code.go", time.Hour)
		h.git.commits[dir] = base.Add(-5 * time.Hour).Unix()
		h.entries("zack " + filepath.Join(h.root, "missing.jsonl") + " " + dir)
		h.reports(stopped("zack", time.Hour, base.Add(-time.Hour).Unix()))
	})

	t.Run("given a directory that is not a repository, when watching, then only the files count", func(t *testing.T) {
		h := newHarness(t)
		dir := h.dir("worktree")
		h.file("worktree/code.go", time.Hour)
		h.entries("zack " + filepath.Join(h.root, "missing.jsonl") + " " + dir)
		h.reports(stopped("zack", time.Hour, base.Add(-time.Hour).Unix()))
	})

	t.Run("given a commit in a directory with no files, when watching, then the commit alone keeps the agent alive", func(t *testing.T) {
		h := newHarness(t)
		dir := h.dir("worktree")
		h.git.commits[dir] = base.Add(-time.Minute).Unix()
		h.entries("zack " + filepath.Join(h.root, "missing.jsonl") + " " + dir)
		h.keepsWatching()
	})
}

func TestSeveralEntries(t *testing.T) {
	t.Run("given a fresh entry before a stale one, when watching, then it reports the stale one", func(t *testing.T) {
		h := newHarness(t)
		h.entries("zack "+h.file("a.jsonl", time.Minute), "billy "+h.file("b.jsonl", time.Hour))
		h.reports(stopped("billy", time.Hour, base.Add(-time.Hour).Unix()))
	})

	t.Run("given two stale entries, when watching, then it reports only the first", func(t *testing.T) {
		h := newHarness(t)
		h.entries("zack "+h.file("a.jsonl", 2*time.Hour), "billy "+h.file("b.jsonl", time.Hour))
		h.reports(stopped("zack", 2*time.Hour, base.Add(-2*time.Hour).Unix()))
	})

	t.Run("given a gone entry before a stale one, when watching, then it reports the gone one", func(t *testing.T) {
		h := newHarness(t)
		h.entries("zack /nowhere/at/all", "billy "+h.file("b.jsonl", time.Hour))
		h.reports("SUMIU zack\n")
	})
}

func TestGitCLI(t *testing.T) {
	for _, pair := range os.Environ() {
		if name, _, _ := strings.Cut(pair, "="); strings.HasPrefix(name, "GIT_") {
			t.Setenv(name, "")
			if err := os.Unsetenv(name); err != nil {
				t.Fatal(err)
			}
		}
	}

	t.Run("given a repository with a commit, when asking, then it returns the commit time", func(t *testing.T) {
		dir := t.TempDir()
		stamp := "2020-01-02T03:04:05Z"
		for _, args := range [][]string{{"init", "-q"}, {"-c", "user.name=t", "-c", "user.email=t@t", "commit", "-q", "--allow-empty", "-m", "x"}} {
			cmd := exec.CommandContext(t.Context(), "git", append([]string{"-C", dir}, args...)...)
			cmd.Env = append(os.Environ(), "GIT_AUTHOR_DATE="+stamp, "GIT_COMMITTER_DATE="+stamp)
			if out, err := cmd.CombinedOutput(); err != nil {
				t.Fatalf("git %v: %v: %s", args, err, out)
			}
		}
		got, ok := GitCLI{}.LastCommit(t.Context(), dir)
		want, _ := time.Parse(time.RFC3339, stamp)
		if !ok || got != want.Unix() {
			t.Fatalf("got %d, %v", got, ok)
		}
	})

	t.Run("given a repository with no commit, when asking, then it reports nothing", func(t *testing.T) {
		dir := t.TempDir()
		if out, err := exec.CommandContext(t.Context(), "git", "-C", dir, "init", "-q").CombinedOutput(); err != nil {
			t.Fatalf("%v: %s", err, out)
		}
		if _, ok := (GitCLI{}).LastCommit(t.Context(), dir); ok {
			t.Fatal("found a commit")
		}
	})

	t.Run("given a directory outside any repository, when asking, then it reports nothing", func(t *testing.T) {
		if _, ok := (GitCLI{}).LastCommit(t.Context(), t.TempDir()); ok {
			t.Fatal("found a commit")
		}
	})
}

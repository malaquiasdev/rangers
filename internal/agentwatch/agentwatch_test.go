package agentwatch

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"testing"
	"time"
)

const watchWindow = 3 * time.Second

type watch struct {
	t    *testing.T
	root string
}

func newWatch(t *testing.T) *watch {
	t.Helper()
	if runtime.GOOS != "darwin" {
		t.Skip("agent-watch reads modification times with the BSD stat")
	}
	return &watch{t: t, root: t.TempDir()}
}

func (w *watch) file(name string, age time.Duration) string {
	w.t.Helper()
	path := filepath.Join(w.root, name)
	if err := os.WriteFile(path, []byte("x"), 0o600); err != nil {
		w.t.Fatal(err)
	}
	when := time.Now().Add(-age)
	if err := os.Chtimes(path, when, when); err != nil {
		w.t.Fatal(err)
	}
	return path
}

func (w *watch) link(name, target string, age time.Duration) string {
	w.t.Helper()
	path := filepath.Join(w.root, name)
	if err := os.Symlink(target, path); err != nil {
		w.t.Fatal(err)
	}
	stamp := time.Now().Add(-age).Format("200601021504")
	if out, err := exec.Command("touch", "-h", "-t", stamp, path).CombinedOutput(); err != nil {
		w.t.Fatalf("touch -h: %v: %s", err, out)
	}
	return path
}

func (w *watch) dir(name string) string {
	w.t.Helper()
	path := filepath.Join(w.root, name)
	if err := os.Mkdir(path, 0o700); err != nil {
		w.t.Fatal(err)
	}
	return path
}

func (w *watch) run(entry string) (string, bool) {
	w.t.Helper()
	list := filepath.Join(w.root, "agent-watch.txt")
	if err := os.WriteFile(list, []byte(entry+"\n"), 0o600); err != nil {
		w.t.Fatal(err)
	}
	script, err := filepath.Abs(filepath.Join("..", "..", "bin", "agent-watch"))
	if err != nil {
		w.t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), watchWindow)
	defer cancel()
	cmd := exec.CommandContext(ctx, "bash", script, "900")
	cmd.Env = append(os.Environ(), "AGENT_WATCH_LIST="+list, "AGENT_WATCH_INTERVAL=1")
	out, runErr := cmd.Output()
	exit := new(exec.ExitError)
	if errors.As(runErr, &exit) || ctx.Err() != nil {
		return strings.TrimSpace(string(out)), false
	}
	if runErr != nil {
		w.t.Fatal(runErr)
	}
	return strings.TrimSpace(string(out)), true
}

func TestAgentWatchSymlink(t *testing.T) {
	t.Run("given a symlink with an old stamp to a fresh transcript, when watching, then it keeps watching", func(t *testing.T) {
		w := newWatch(t)
		target := w.file("transcript.jsonl", time.Minute)
		link := w.link("task.output", target, 2*time.Hour)
		if out, exited := w.run("zack " + link); exited {
			t.Fatalf("reported %q for a transcript written a minute ago", out)
		}
	})

	t.Run("given a symlink with a fresh stamp to a stale transcript, when watching, then it reports the transcript age", func(t *testing.T) {
		w := newWatch(t)
		target := w.file("transcript.jsonl", time.Hour)
		link := w.link("task.output", target, time.Minute)
		out, exited := w.run("zack " + link)
		if !exited {
			t.Fatal("kept watching a transcript idle for an hour")
		}
		if !regexp.MustCompile(`^PARADO zack há 6[01] min`).MatchString(out) {
			t.Fatalf("got %q", out)
		}
	})

	t.Run("given a symlink to a missing transcript and no directory, when watching, then it reports it gone", func(t *testing.T) {
		w := newWatch(t)
		link := w.link("task.output", filepath.Join(w.root, "missing.jsonl"), time.Minute)
		out, exited := w.run("zack " + link)
		if !exited || out != "SUMIU zack" {
			t.Fatalf("got %q, exited %v", out, exited)
		}
	})

	t.Run("given a symlink to a missing transcript and a directory with no files, when watching, then it reports it gone instead of an idle time", func(t *testing.T) {
		w := newWatch(t)
		link := w.link("task.output", filepath.Join(w.root, "missing.jsonl"), time.Minute)
		dir := w.dir("worktree")
		out, exited := w.run("zack " + link + " " + dir)
		if !exited || out != "SUMIU zack" {
			t.Fatalf("got %q, exited %v", out, exited)
		}
	})
}

package hook

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

const completeScript = `on run argv
  set listName to item 1 of argv
  set waitingPrefix to item 2 of argv
  tell application "Reminders"
    if not (exists list listName) then return
    set waiting to (every reminder of list listName whose completed is false and name starts with waitingPrefix)
    repeat with entry in waiting
      set completed of entry to true
    end repeat
  end tell
end run
`

type fakeScripter struct {
	absent   bool
	err      error
	hang     bool
	calls    int
	script   string
	args     []string
	deadline time.Time
}

func (f *fakeScripter) Available() bool {
	return !f.absent
}

func (f *fakeScripter) Run(ctx context.Context, script string, args []string) error {
	f.calls++
	f.script = script
	f.args = args
	f.deadline, _ = ctx.Deadline()
	if f.hang {
		<-ctx.Done()
		return ctx.Err()
	}
	return f.err
}

type unreadable struct {
	t *testing.T
}

func (u unreadable) Read([]byte) (int, error) {
	u.t.Error("the hook must not read standard input")
	return 0, errors.New("read")
}

func envOf(values map[string]string) func(string) string {
	return func(name string) string { return values[name] }
}

func runComplete(t *testing.T, f *fakeScripter, env map[string]string) (string, error) {
	t.Helper()
	var out bytes.Buffer
	dep := &Dependencies{In: unreadable{t}, Out: &out, Scripter: f}
	err := Command(t.Context(), []string{"complete-reminders"}, envOf(env), dep)
	return out.String(), err
}

func TestCompleteReminders(t *testing.T) {
	t.Run("given the human sends a message, when the hook runs, then it asks osascript about the Tarefas Claude list and the prefix", func(t *testing.T) {
		f := &fakeScripter{}
		out, err := runComplete(t, f, nil)
		if err != nil || out != "" {
			t.Fatalf("err %v, out %q", err, out)
		}
		if f.calls != 1 || strings.Join(f.args, "|") != "Tarefas Claude|Claude espera: " {
			t.Fatalf("calls %d, args %q", f.calls, f.args)
		}
	})

	t.Run("given the human sends a message, when the hook runs, then the script completes only open reminders that start with the prefix", func(t *testing.T) {
		f := &fakeScripter{}
		if _, err := runComplete(t, f, nil); err != nil {
			t.Fatal(err)
		}
		if f.script != completeScript {
			t.Fatalf("got %q", f.script)
		}
	})

	t.Run("given osascript is not installed, when the hook runs, then it does nothing and stays silent", func(t *testing.T) {
		f := &fakeScripter{absent: true}
		out, err := runComplete(t, f, nil)
		if err != nil || out != "" || f.calls != 0 {
			t.Fatalf("err %v, out %q, calls %d", err, out, f.calls)
		}
	})

	t.Run("given osascript fails, when the hook runs, then it stays silent and reports no error", func(t *testing.T) {
		f := &fakeScripter{err: errors.New("exit status 1")}
		out, err := runComplete(t, f, nil)
		if err != nil || out != "" {
			t.Fatalf("err %v, out %q", err, out)
		}
	})

	t.Run("given osascript hangs, when the hook runs, then it gives up after the timeout", func(t *testing.T) {
		f := &fakeScripter{hang: true}
		started := time.Now()
		out, err := runComplete(t, f, map[string]string{"COMPLETE_REMINDERS_TIMEOUT": "0.05"})
		if err != nil || out != "" {
			t.Fatalf("err %v, out %q", err, out)
		}
		if time.Since(started) > 5*time.Second {
			t.Fatal("kept waiting for osascript")
		}
	})
}

func TestCompleteRemindersTimeout(t *testing.T) {
	tests := []struct {
		name  string
		value string
		want  time.Duration
	}{
		{"given no timeout variable, when the hook runs, then the limit is 10 seconds", "", 10 * time.Second},
		{"given a timeout in seconds, when the hook runs, then the limit is that many seconds", "3", 3 * time.Second},
		{"given a fractional timeout, when the hook runs, then the limit is that fraction", "0.5", 500 * time.Millisecond},
		{"given a timeout that is not a number, when the hook runs, then the limit is 10 seconds", "soon", 10 * time.Second},
		{"given a timeout of zero, when the hook runs, then the limit is 10 seconds", "0", 10 * time.Second},
		{"given a negative timeout, when the hook runs, then the limit is 10 seconds", "-4", 10 * time.Second},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			f := &fakeScripter{}
			before := time.Now()
			if _, err := runComplete(t, f, map[string]string{"COMPLETE_REMINDERS_TIMEOUT": tc.value}); err != nil {
				t.Fatal(err)
			}
			got := f.deadline.Sub(before)
			if got > tc.want+time.Second/10 || got < tc.want-time.Second/2 {
				t.Fatalf("limit %s, want %s", got, tc.want)
			}
		})
	}
}

func TestPrefixIsShared(t *testing.T) {
	t.Run("given the prefix the hook closes, when the stop hook blocks, then the reminder it asks for uses that same prefix", func(t *testing.T) {
		f := &fakeScripter{}
		if _, err := runComplete(t, f, nil); err != nil {
			t.Fatal(err)
		}
		path := writeTranscript(t, encode(t, []entry{user("go"), said("Which one?")}))
		out, err := runStop(t, payloadFor(t, path, false))
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(out, f.args[1]) || !strings.Contains(out, f.args[0]) {
			t.Fatalf("reason %q does not use %q", out, f.args)
		}
	})
}

func writeStub(t *testing.T, body string) {
	t.Helper()
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "osascript"), []byte("#!/bin/sh\n"+body), 0o700); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", dir)
}

func TestOsascriptCLI(t *testing.T) {
	t.Run("given no osascript on the path, when asking, then it is not available", func(t *testing.T) {
		t.Setenv("PATH", t.TempDir())
		if (OsascriptCLI{}).Available() {
			t.Fatal("available")
		}
	})

	t.Run("given osascript on the path, when asking, then it is available", func(t *testing.T) {
		writeStub(t, "exit 0\n")
		if !(OsascriptCLI{}).Available() {
			t.Fatal("not available")
		}
	})

	t.Run("given a script, when running, then osascript gets a dash, the arguments and the script on standard input", func(t *testing.T) {
		record := filepath.Join(t.TempDir(), "record")
		writeStub(t, "for a in \"$@\"; do printf '%s\\n' \"$a\"; done > "+record+".args\n/bin/cat > "+record+".stdin\n")
		if err := (OsascriptCLI{}).Run(t.Context(), "the script", []string{"list", "prefix: "}); err != nil {
			t.Fatal(err)
		}
		args, _ := os.ReadFile(record + ".args")
		stdin, _ := os.ReadFile(record + ".stdin")
		if string(args) != "-\nlist\nprefix: \n" || string(stdin) != "the script" {
			t.Fatalf("args %q, stdin %q", args, stdin)
		}
	})

	t.Run("given osascript fails, when running, then it returns the error", func(t *testing.T) {
		writeStub(t, "echo noise\necho warning >&2\nexit 1\n")
		if err := (OsascriptCLI{}).Run(t.Context(), "x", nil); err == nil {
			t.Fatal("no error")
		}
	})

	t.Run("given osascript hangs, when the context expires, then it stops the process and returns", func(t *testing.T) {
		writeStub(t, "exec /bin/sleep 30\n")
		ctx, cancel := context.WithTimeout(t.Context(), 100*time.Millisecond)
		defer cancel()
		started := time.Now()
		if err := (OsascriptCLI{}).Run(ctx, "x", nil); err == nil {
			t.Fatal("no error")
		}
		if time.Since(started) > 10*time.Second {
			t.Fatal("kept waiting for osascript")
		}
	})
}

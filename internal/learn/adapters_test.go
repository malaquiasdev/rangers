package learn

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestOsascriptReminderTimeout(t *testing.T) {
	t.Run("given an osascript that hangs, when creating a reminder with a short timeout, then it returns an error long before the script ends", func(t *testing.T) {
		bin := t.TempDir()
		writeExecutable(t, filepath.Join(bin, "osascript"), "#!/bin/sh\nexec sleep 30\n")
		t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))
		started := time.Now()
		err := OsascriptReminder{Timeout: 200 * time.Millisecond}.Create(t.Context(), "title", "body")
		if err == nil {
			t.Fatal("expected an error")
		}
		if elapsed := time.Since(started); elapsed > 5*time.Second {
			t.Fatalf("took %v", elapsed)
		}
	})
}

func TestClaudeCLITimeout(t *testing.T) {
	t.Run("given a claude that leaves a child holding its output open, when the context expires, then Run returns long before the child ends", func(t *testing.T) {
		bin := t.TempDir()
		writeExecutable(t, filepath.Join(bin, "claude"), "#!/bin/sh\nsleep 30\nsleep 30\n")
		t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))
		ctx, cancel := context.WithTimeout(t.Context(), 200*time.Millisecond)
		defer cancel()
		started := time.Now()
		_, err := ClaudeCLI{}.Run(ctx, t.TempDir(), nil)
		if err == nil {
			t.Fatal("expected an error")
		}
		if elapsed := time.Since(started); elapsed > 10*time.Second {
			t.Fatalf("took %v", elapsed)
		}
	})
}

func TestDirBrainProbe(t *testing.T) {
	probe := DirBrain{Timeout: 2 * time.Second}

	t.Run("given a folder that can be listed and written, when probing, then it succeeds and leaves no file behind", func(t *testing.T) {
		dir := t.TempDir()
		if err := probe.Probe(t.Context(), dir); err != nil {
			t.Fatal(err)
		}
		entries, err := os.ReadDir(dir)
		if err != nil || len(entries) != 0 {
			t.Fatalf("entries %v, err %v", entries, err)
		}
	})

	t.Run("given a folder that does not exist, when probing, then it returns an error naming the folder", func(t *testing.T) {
		dir := filepath.Join(t.TempDir(), "missing")
		err := probe.Probe(t.Context(), dir)
		if err == nil || !strings.Contains(err.Error(), dir) {
			t.Fatalf("got %v", err)
		}
	})

	t.Run("given a path that is a file, when probing, then it returns an error", func(t *testing.T) {
		file := filepath.Join(t.TempDir(), "notes")
		if err := os.WriteFile(file, nil, 0o600); err != nil {
			t.Fatal(err)
		}
		if err := probe.Probe(t.Context(), file); err == nil {
			t.Fatal("expected an error")
		}
	})
}

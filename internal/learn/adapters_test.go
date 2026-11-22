package learn

import (
	"context"
	"os"
	"path/filepath"
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

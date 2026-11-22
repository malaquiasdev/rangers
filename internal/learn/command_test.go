package learn

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func command(t *testing.T, args ...string) (string, error) {
	t.Helper()
	home := t.TempDir()
	out := &bytes.Buffer{}
	dep := &Dependencies{Now: func() time.Time { return time.Date(2026, 11, 22, 10, 30, 0, 0, time.UTC) }, Out: out}
	err := Command(context.Background(), args, func(name string) string {
		if name == "HOME" {
			return home
		}
		return ""
	}, dep)
	return out.String(), err
}

func TestCommand(t *testing.T) {
	t.Run("given the metrics subcommand and a state folder, when running, then it shows the metrics from it", func(t *testing.T) {
		state := t.TempDir()
		row := "2026-11-22-1030\t12\t5\t3\t0.123\t1600\t50\n"
		if err := os.WriteFile(filepath.Join(state, "metrics.tsv"), []byte(row), 0o600); err != nil {
			t.Fatal(err)
		}
		out, err := command(t, "metrics", "--state", state)
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(out, strings.TrimSuffix(row, "\n")+"\trules\n") || !strings.Contains(out, "1 rodadas") {
			t.Fatalf("got %q", out)
		}
	})

	t.Run("given no subcommand and no new lines, when running, then the gate blocks and it exits without error", func(t *testing.T) {
		root := t.TempDir()
		out, err := command(t, "--state", filepath.Join(root, "state"), "--projects", filepath.Join(root, "projects"))
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(out, "só 0 falas novas, abaixo do mínimo de 10") {
			t.Fatalf("got %q", out)
		}
	})

	t.Run("given an unknown flag, when running, then it returns an error", func(t *testing.T) {
		if _, err := command(t, "--does-not-exist"); err == nil {
			t.Fatal("expected an error")
		}
	})
}

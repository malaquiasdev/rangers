package transcript

import (
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestConversation(t *testing.T) {
	t.Run("given a long assistant message, when reading the conversation, then nothing is cut", func(t *testing.T) {
		projects := t.TempDir()
		long := strings.Repeat("a", 900) + " gate:abcd1234"
		write(t, filepath.Join(projects, "proj/s.jsonl"), after(), assistantLine(t, after(), long))
		lines, err := Conversation(projects, cutoff)
		if err != nil {
			t.Fatal(err)
		}
		if len(lines) != 1 || !strings.HasSuffix(lines[0].Text, "gate:abcd1234") || !lines[0].Assistant {
			t.Fatalf("got %+v", lines)
		}
	})

	t.Run("given a long human message, when reading the conversation, then nothing is cut", func(t *testing.T) {
		projects := t.TempDir()
		write(t, filepath.Join(projects, "proj/s.jsonl"), after(), humanLine(t, after(), strings.Repeat("b", 900)))
		lines, err := Conversation(projects, cutoff)
		if err != nil {
			t.Fatal(err)
		}
		if len(lines) != 1 || len(lines[0].Text) != 900 {
			t.Fatalf("got %+v", lines)
		}
	})

	t.Run("given two sessions, when reading the conversation, then each session keeps the order of its file", func(t *testing.T) {
		projects := t.TempDir()
		write(t, filepath.Join(projects, "proj/a.jsonl"), after(),
			assistantLine(t, after().Add(2*time.Minute), "a question"),
			humanLine(t, after().Add(3*time.Minute), "sim"),
		)
		write(t, filepath.Join(projects, "proj/b.jsonl"), after(), humanLine(t, after().Add(time.Minute), "other"))
		lines, err := Conversation(projects, cutoff)
		if err != nil {
			t.Fatal(err)
		}
		got := make([]string, len(lines))
		for i, line := range lines {
			got[i] = line.Session + ":" + line.Text
		}
		if want := "a:a question,a:sim,b:other"; strings.Join(got, ",") != want {
			t.Fatalf("got %v", got)
		}
	})

	t.Run("given a system notification, when reading the conversation, then it is dropped like a human line", func(t *testing.T) {
		projects := t.TempDir()
		write(t, filepath.Join(projects, "proj/s.jsonl"), after(), humanLine(t, after(), "<task-notification>sim</task-notification>"))
		lines, err := Conversation(projects, cutoff)
		if err != nil || len(lines) != 0 {
			t.Fatalf("got %+v, %v", lines, err)
		}
	})

	t.Run("given a projects folder that does not exist, when reading the conversation, then it returns nothing and no error", func(t *testing.T) {
		lines, err := Conversation(filepath.Join(t.TempDir(), "missing"), cutoff)
		if err != nil || len(lines) != 0 {
			t.Fatalf("got %+v, %v", lines, err)
		}
	})
}

package transcript

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

var cutoff = time.Date(2026, 11, 20, 12, 0, 0, 0, time.UTC)

func entryLine(t *testing.T, kind, role string, at time.Time, content any) string {
	t.Helper()
	raw, err := json.Marshal(map[string]any{
		"type":      kind,
		"timestamp": at.Format("2006-01-02T15:04:05.000Z"),
		"message":   map[string]any{"role": role, "content": content},
	})
	if err != nil {
		t.Fatal(err)
	}
	return string(raw)
}

func after() time.Time { return cutoff.Add(time.Hour) }

func write(t *testing.T, projects, path string, mtime time.Time, lines ...string) {
	t.Helper()
	target := filepath.Join(projects, path)
	if err := os.MkdirAll(filepath.Dir(target), 0o750); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(target, []byte(strings.Join(lines, "\n")+"\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Chtimes(target, mtime, mtime); err != nil {
		t.Fatal(err)
	}
}

func extract(t *testing.T, projects string) []Line {
	t.Helper()
	lines, err := Extract(projects, cutoff)
	if err != nil {
		t.Fatal(err)
	}
	return lines
}

func TestExtract(t *testing.T) {
	t.Run("given a human text line after the cutoff, when extracting, then it returns session, time and text", func(t *testing.T) {
		projects := t.TempDir()
		write(t, projects, "proj/abc-123.jsonl", after(), entryLine(t, "user", "user", after(), "fix the test"))
		lines := extract(t, projects)
		if len(lines) != 1 {
			t.Fatalf("expected 1 line, got %d", len(lines))
		}
		expected := "abc-123\t2026-11-20T13:00:00.000Z\tfix the test"
		if lines[0].String() != expected {
			t.Fatalf("expected %q, got %q", expected, lines[0].String())
		}
	})

	t.Run("given content as a list of parts, when extracting, then it joins only the text parts with a space", func(t *testing.T) {
		projects := t.TempDir()
		parts := []map[string]any{
			{"type": "text", "text": "first"},
			{"type": "image", "source": "x"},
			{"type": "text", "text": "second"},
		}
		write(t, projects, "proj/s.jsonl", after(), entryLine(t, "user", "user", after(), parts))
		lines := extract(t, projects)
		if len(lines) != 1 || lines[0].Text != "first second" {
			t.Fatalf("got %+v", lines)
		}
	})

	prefixes := []string{"<command-name>", "Base directory for this skill", "Another Claude asked", "[Request interrupted by user]", "Caveat: the messages", "[Cross-session notice]", "[Subagent finished]"}
	for _, prefix := range prefixes {
		t.Run("given a line starting with "+prefix[:5]+", when extracting, then it drops the line", func(t *testing.T) {
			projects := t.TempDir()
			write(t, projects, "proj/s.jsonl", after(), entryLine(t, "user", "user", after(), prefix+" rest"), entryLine(t, "user", "user", after(), "kept"))
			lines := extract(t, projects)
			if len(lines) != 1 || lines[0].Text != "kept" {
				t.Fatalf("got %+v", lines)
			}
		})
	}

	t.Run("given lines of another type or role, when extracting, then it drops them", func(t *testing.T) {
		projects := t.TempDir()
		write(t, projects, "proj/s.jsonl", after(),
			entryLine(t, "assistant", "assistant", after(), "answer"),
			entryLine(t, "user", "assistant", after(), "wrong role"),
			entryLine(t, "system", "user", after(), "wrong type"),
		)
		if lines := extract(t, projects); len(lines) != 0 {
			t.Fatalf("got %+v", lines)
		}
	})

	t.Run("given empty content, when extracting, then it drops the line", func(t *testing.T) {
		projects := t.TempDir()
		write(t, projects, "proj/s.jsonl", after(), entryLine(t, "user", "user", after(), ""), entryLine(t, "user", "user", after(), []map[string]any{}))
		if lines := extract(t, projects); len(lines) != 0 {
			t.Fatalf("got %+v", lines)
		}
	})

	t.Run("given line breaks and a long text, when extracting, then it turns breaks into spaces and cuts at 600 characters", func(t *testing.T) {
		projects := t.TempDir()
		long := "a\nb" + strings.Repeat("ç", 700)
		write(t, projects, "proj/s.jsonl", after(), entryLine(t, "user", "user", after(), long))
		lines := extract(t, projects)
		if len(lines) != 1 {
			t.Fatalf("got %+v", lines)
		}
		text := lines[0].Text
		if strings.Contains(text, "\n") || !strings.HasPrefix(text, "a b") {
			t.Fatalf("line break did not become a space: %q", text)
		}
		if n := len([]rune(text)); n != 600 {
			t.Fatalf("expected 600 characters, got %d", n)
		}
	})

	t.Run("given lines before and after the cutoff, when extracting, then only the later one stays", func(t *testing.T) {
		projects := t.TempDir()
		write(t, projects, "proj/s.jsonl", after(),
			entryLine(t, "user", "user", cutoff.Add(-time.Minute), "old"),
			entryLine(t, "user", "user", cutoff.Add(time.Minute), "new"),
		)
		lines := extract(t, projects)
		if len(lines) != 1 || lines[0].Text != "new" {
			t.Fatalf("got %+v", lines)
		}
	})

	t.Run("given a file not changed since the cutoff, when extracting, then it skips the file", func(t *testing.T) {
		projects := t.TempDir()
		write(t, projects, "proj/stale.jsonl", cutoff.Add(-time.Minute), entryLine(t, "user", "user", after(), "not read"))
		if lines := extract(t, projects); len(lines) != 0 {
			t.Fatalf("got %+v", lines)
		}
	})

	t.Run("given files at depth 1, 2 and 3, when extracting, then it reads only up to depth 2", func(t *testing.T) {
		projects := t.TempDir()
		write(t, projects, "shallow.jsonl", after(), entryLine(t, "user", "user", after(), "one"))
		write(t, projects, "proj/middle.jsonl", after(), entryLine(t, "user", "user", after(), "two"))
		write(t, projects, "proj/sub/deep.jsonl", after(), entryLine(t, "user", "user", after(), "three"))
		write(t, projects, "proj/note.txt", after(), entryLine(t, "user", "user", after(), "four"))
		texts := []string{}
		for _, line := range extract(t, projects) {
			texts = append(texts, line.Text)
		}
		if strings.Join(texts, ",") != "two,one" {
			t.Fatalf("got %v", texts)
		}
	})

	t.Run("given a corrupted line, when extracting, then it returns what it read before and does not fail", func(t *testing.T) {
		projects := t.TempDir()
		write(t, projects, "proj/s.jsonl", after(), entryLine(t, "user", "user", after(), "good"), "{broken")
		lines := extract(t, projects)
		if len(lines) != 1 || lines[0].Text != "good" {
			t.Fatalf("got %+v", lines)
		}
	})

	t.Run("given a projects folder that does not exist, when extracting, then it returns nothing and no error", func(t *testing.T) {
		lines, err := Extract(filepath.Join(t.TempDir(), "missing"), cutoff)
		if err != nil || len(lines) != 0 {
			t.Fatalf("got %+v, %v", lines, err)
		}
	})
}

func TestExtractWithAssistant(t *testing.T) {
	t.Run("given an assistant text message, when extracting with the assistant, then it returns the line marked and cut at 400 characters", func(t *testing.T) {
		projects := t.TempDir()
		long := strings.Repeat("ç", 500)
		parts := []map[string]any{{"type": "text", "text": long}, {"type": "tool_use", "name": "Bash"}}
		write(t, projects, "-Users-x-proj/s1.jsonl", after(), entryLine(t, "assistant", "assistant", after(), parts))
		lines, err := ExtractWithAssistant(projects, cutoff)
		if err != nil {
			t.Fatal(err)
		}
		if len(lines) != 1 || !lines[0].Assistant || len([]rune(lines[0].Text)) != 400 {
			t.Fatalf("got %+v", lines)
		}
		if !strings.HasPrefix(lines[0].Notes(), "-Users-x-proj\ts1\t2026-11-20T13:00:00.000Z\t[assistant] çç") {
			t.Fatalf("got %q", lines[0].Notes())
		}
	})

	t.Run("given an assistant message with no text part, when extracting with the assistant, then it drops the message", func(t *testing.T) {
		projects := t.TempDir()
		parts := []map[string]any{{"type": "tool_use", "name": "Bash"}}
		write(t, projects, "proj/s.jsonl", after(), entryLine(t, "assistant", "assistant", after(), parts))
		lines, err := ExtractWithAssistant(projects, cutoff)
		if err != nil || len(lines) != 0 {
			t.Fatalf("got %+v, %v", lines, err)
		}
	})

	t.Run("given human and assistant lines in two sessions, when extracting with the assistant, then it returns them in time order with the project of each", func(t *testing.T) {
		projects := t.TempDir()
		write(t, projects, "-p-one/a.jsonl", after(),
			entryLine(t, "user", "user", cutoff.Add(10*time.Minute), "human first"),
			entryLine(t, "assistant", "assistant", cutoff.Add(30*time.Minute), "assistant third"),
		)
		write(t, projects, "-p-two/b.jsonl", after(), entryLine(t, "user", "user", cutoff.Add(20*time.Minute), "human second"))
		lines, err := ExtractWithAssistant(projects, cutoff)
		if err != nil {
			t.Fatal(err)
		}
		var got []string
		for _, line := range lines {
			got = append(got, line.Project+":"+line.Text)
		}
		if strings.Join(got, ",") != "-p-one:human first,-p-two:human second,-p-one:assistant third" {
			t.Fatalf("got %v", got)
		}
	})

	t.Run("given assistant messages, when extracting only the human lines, then they stay out", func(t *testing.T) {
		projects := t.TempDir()
		write(t, projects, "proj/s.jsonl", after(), entryLine(t, "assistant", "assistant", after(), "answer"))
		if lines := extract(t, projects); len(lines) != 0 {
			t.Fatalf("got %+v", lines)
		}
	})
}

func TestFormat(t *testing.T) {
	lines := []Line{
		{Project: "p", Session: "s", Time: "t1", Text: "first"},
		{Project: "p", Session: "s", Time: "t2", Text: "second", Assistant: true},
		{Project: "p", Session: "s", Time: "t3", Text: "third"},
	}

	t.Run("given a limit that fits everything, when formatting, then it renders one row per line", func(t *testing.T) {
		expected := "p\ts\tt1\tfirst\np\ts\tt2\t[assistant] second\np\ts\tt3\tthird\n"
		if got := Format(lines, 1000); got != expected {
			t.Fatalf("expected %q, got %q", expected, got)
		}
	})

	t.Run("given a limit smaller than the whole text, when formatting, then it cuts from the start and keeps the most recent lines", func(t *testing.T) {
		limit := len("p\ts\tt2\t[assistant] second\np\ts\tt3\tthird\n")
		expected := "p\ts\tt2\t[assistant] second\np\ts\tt3\tthird\n"
		if got := Format(lines, limit); got != expected {
			t.Fatalf("expected %q, got %q", expected, got)
		}
		if got := Format(lines, limit-1); got != "p\ts\tt3\tthird\n" {
			t.Fatalf("got %q", got)
		}
	})

	t.Run("given a limit smaller than the last line, when formatting, then it returns nothing", func(t *testing.T) {
		if got := Format(lines, 5); got != "" {
			t.Fatalf("got %q", got)
		}
	})
}

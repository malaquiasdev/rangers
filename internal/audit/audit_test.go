package audit

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
	_ "time/tzdata"
)

const (
	brain      = "/brain"
	auditDir   = "/brain/claude/auditoria"
	noteHeader = "---\n" +
		"type: Log\n" +
		"title: Auditoria 2026-11-22\n" +
		"description: \"O que os agentes fizeram sem aprovação em 2026-11-22: merge, push, deploy em dev, escrita no brain\"\n" +
		"timestamp: 2026-11-22\n" +
		"---\n" +
		"# Auditoria 2026-11-22\n\n" +
		"| hora | sessão | onde | ação | link | por quê |\n" +
		"|---|---|---|---|---|---|\n"
	indexLine = "- [[claude/auditoria/2026-11-22|Auditoria 2026-11-22]] — o que os agentes fizeram sozinhos\n"
)

var fourFields = []string{"repo", "merge", "https://example.test/pull/1", "CI green"}

type memStore struct {
	files map[string]string
	dirs  map[string]bool
}

func newStore() *memStore {
	return &memStore{files: map[string]string{}, dirs: map[string]bool{auditDir: true}}
}

func (m *memStore) Exists(path string) bool {
	_, ok := m.files[path]
	return ok
}

func (m *memStore) Create(path string, data []byte) error {
	if !m.dirs[filepath.Dir(path)] {
		return fs.ErrNotExist
	}
	m.files[path] = string(data)
	return nil
}

func (m *memStore) Append(path string, data []byte) error {
	if !m.dirs[filepath.Dir(path)] {
		return fs.ErrNotExist
	}
	m.files[path] += string(data)
	return nil
}

func environment(values map[string]string) func(string) string {
	return func(name string) string { return values[name] }
}

func at(t *testing.T, utc string) func() time.Time {
	t.Helper()
	moment, err := time.Parse(time.RFC3339, utc)
	if err != nil {
		t.Fatal(err)
	}
	return func() time.Time { return moment }
}

func execute(t *testing.T, store Store, env map[string]string, args ...string) error {
	t.Helper()
	dep := &Dependencies{Store: store, Now: at(t, "2026-11-22T15:04:00Z")}
	return Command(t.Context(), args, environment(env), dep)
}

func executeAt(t *testing.T, store Store, utc string, env map[string]string) error {
	t.Helper()
	dep := &Dependencies{Store: store, Now: at(t, utc)}
	return Command(t.Context(), fourFields, environment(env), dep)
}

func TestUsage(t *testing.T) {
	for count := range 4 {
		t.Run(fmt.Sprintf("given %d arguments, when auditing, then it fails with the usage and writes nothing", count), func(t *testing.T) {
			store := newStore()
			err := execute(t, store, map[string]string{"BRAIN_DIR": brain}, fourFields[:count]...)
			if !errors.Is(err, ErrUsage) {
				t.Fatalf("got %v", err)
			}
			if len(store.files) != 0 {
				t.Fatalf("wrote %v", store.files)
			}
		})
	}

	t.Run("given the usage error, when printing it, then it is the message of the Bash script", func(t *testing.T) {
		if got := ErrUsage.Error(); got != "uso: audit <repo|lugar> <ação> <link> <por quê>" {
			t.Fatalf("got %q", got)
		}
	})
}

func TestNewDay(t *testing.T) {
	t.Run("given a day with no note, when auditing, then it creates the note, indexes it and appends the row", func(t *testing.T) {
		store := newStore()
		env := map[string]string{"BRAIN_DIR": brain, "CLAUDE_SESSION_ID": "0123456789abcdef"}
		if err := execute(t, store, env, fourFields...); err != nil {
			t.Fatal(err)
		}
		row := "| 12:04 | 01234567 | repo | merge | https://example.test/pull/1 | CI green |\n"
		if got := store.files[auditDir+"/2026-11-22.md"]; got != noteHeader+row {
			t.Fatalf("got %q", got)
		}
		if got := store.files[auditDir+"/index.md"]; got != indexLine {
			t.Fatalf("got %q", got)
		}
	})
}

func TestExistingDay(t *testing.T) {
	t.Run("given a day that has a note, when auditing, then it only appends the row", func(t *testing.T) {
		store := newStore()
		store.files[auditDir+"/2026-11-22.md"] = "note\n"
		store.files[auditDir+"/index.md"] = "index\n"
		env := map[string]string{"BRAIN_DIR": brain, "AUDIT_SESSION": "abc"}
		if err := execute(t, store, env, fourFields...); err != nil {
			t.Fatal(err)
		}
		want := "note\n| 12:04 | abc | repo | merge | https://example.test/pull/1 | CI green |\n"
		if got := store.files[auditDir+"/2026-11-22.md"]; got != want {
			t.Fatalf("got %q", got)
		}
		if got := store.files[auditDir+"/index.md"]; got != "index\n" {
			t.Fatalf("index changed: %q", got)
		}
	})

	t.Run("given two audits on the same day, when the second runs, then the header and the index line appear once", func(t *testing.T) {
		store := newStore()
		env := map[string]string{"BRAIN_DIR": brain}
		for range 2 {
			if err := execute(t, store, env, fourFields...); err != nil {
				t.Fatal(err)
			}
		}
		if got := store.files[auditDir+"/index.md"]; got != indexLine {
			t.Fatalf("got %q", got)
		}
		if got := len(store.files[auditDir+"/2026-11-22.md"]); got <= len(noteHeader) {
			t.Fatalf("rows missing: %d", got)
		}
	})
}

func TestSaoPauloTime(t *testing.T) {
	tests := []struct {
		name, utc, day, hour string
	}{
		{"given an instant after midnight UTC, when auditing, then the day and hour are the ones in Sao Paulo", "2026-11-23T01:30:00Z", "2026-11-22", "22:30"},
		{"given an instant at noon UTC, when auditing, then the hour is three hours earlier", "2026-11-22T12:00:00Z", "2026-11-22", "09:00"},
		{"given an instant just after midnight in Sao Paulo, when auditing, then the new day starts", "2026-11-23T03:05:00Z", "2026-11-23", "00:05"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			store := newStore()
			if err := executeAt(t, store, tc.utc, map[string]string{"BRAIN_DIR": brain}); err != nil {
				t.Fatal(err)
			}
			note, ok := store.files[auditDir+"/"+tc.day+".md"]
			if !ok {
				t.Fatalf("no note for %s: %v", tc.day, store.files)
			}
			if want := "| " + tc.hour + " | ? | repo |"; !strings.Contains(note, want) {
				t.Fatalf("missing %q in %q", want, note)
			}
		})
	}
}

func TestSession(t *testing.T) {
	tests := []struct {
		name string
		env  map[string]string
		want string
	}{
		{"given a long session id, when auditing, then only its first 8 characters are written", map[string]string{"CLAUDE_SESSION_ID": "abcdefghijkl"}, "abcdefgh"},
		{"given a short session id, when auditing, then it is written whole", map[string]string{"CLAUDE_SESSION_ID": "abc"}, "abc"},
		{"given only the audit session variable, when auditing, then it is used", map[string]string{"AUDIT_SESSION": "manual-run-1"}, "manual-r"},
		{"given both variables, when auditing, then the Claude session id wins", map[string]string{"CLAUDE_SESSION_ID": "claude", "AUDIT_SESSION": "audit"}, "claude"},
		{"given an empty Claude session id, when auditing, then the audit session variable is used", map[string]string{"CLAUDE_SESSION_ID": "", "AUDIT_SESSION": "audit"}, "audit"},
		{"given no session variable, when auditing, then a question mark is written", map[string]string{}, "?"},
		{"given a session id with accents, when auditing, then it is cut by characters and not by bytes", map[string]string{"CLAUDE_SESSION_ID": "ééééééééé"}, "éééééééé"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			store := newStore()
			tc.env["BRAIN_DIR"] = brain
			if err := execute(t, store, tc.env, fourFields...); err != nil {
				t.Fatal(err)
			}
			want := "| 12:04 | " + tc.want + " | repo | merge |"
			if note := store.files[auditDir+"/2026-11-22.md"]; !strings.Contains(note, want) {
				t.Fatalf("missing %q in %q", want, note)
			}
		})
	}
}

func TestBrainDirectory(t *testing.T) {
	t.Run("given no BRAIN_DIR, when auditing, then the note goes under the home folder", func(t *testing.T) {
		store := newStore()
		store.dirs["/home/x/Documents/brain/claude/auditoria"] = true
		if err := execute(t, store, map[string]string{"HOME": "/home/x"}, fourFields...); err != nil {
			t.Fatal(err)
		}
		if !store.Exists("/home/x/Documents/brain/claude/auditoria/2026-11-22.md") {
			t.Fatalf("got %v", store.files)
		}
	})

	t.Run("given an empty BRAIN_DIR, when auditing, then it falls back to the home folder", func(t *testing.T) {
		store := newStore()
		store.dirs["/home/x/Documents/brain/claude/auditoria"] = true
		env := map[string]string{"HOME": "/home/x", "BRAIN_DIR": ""}
		if err := execute(t, store, env, fourFields...); err != nil {
			t.Fatal(err)
		}
		if !store.Exists("/home/x/Documents/brain/claude/auditoria/2026-11-22.md") {
			t.Fatalf("got %v", store.files)
		}
	})

	t.Run("given a BRAIN_DIR, when auditing, then it wins over the home folder", func(t *testing.T) {
		store := newStore()
		env := map[string]string{"HOME": "/home/x", "BRAIN_DIR": brain}
		if err := execute(t, store, env, fourFields...); err != nil {
			t.Fatal(err)
		}
		if !store.Exists(auditDir + "/2026-11-22.md") {
			t.Fatalf("got %v", store.files)
		}
	})

	t.Run("given a folder that does not exist, when auditing, then it fails and creates nothing", func(t *testing.T) {
		store := &memStore{files: map[string]string{}, dirs: map[string]bool{}}
		err := execute(t, store, map[string]string{"BRAIN_DIR": brain}, fourFields...)
		if !errors.Is(err, fs.ErrNotExist) {
			t.Fatalf("got %v", err)
		}
		if len(store.files) != 0 {
			t.Fatalf("wrote %v", store.files)
		}
	})
}

func TestFields(t *testing.T) {
	t.Run("given more than four arguments, when auditing, then the extra ones are ignored", func(t *testing.T) {
		store := newStore()
		args := append(append([]string{}, fourFields...), "extra", "more")
		if err := execute(t, store, map[string]string{"BRAIN_DIR": brain}, args...); err != nil {
			t.Fatal(err)
		}
		if note := store.files[auditDir+"/2026-11-22.md"]; strings.Contains(note, "extra") {
			t.Fatalf("got %q", note)
		}
	})

	t.Run("given fields with percent signs and backslashes, when auditing, then they are written as they are", func(t *testing.T) {
		store := newStore()
		args := []string{"100%s", `a\nb`, "%d", "50% done"}
		if err := execute(t, store, map[string]string{"BRAIN_DIR": brain}, args...); err != nil {
			t.Fatal(err)
		}
		want := `| 12:04 | ? | 100%s | a\nb | %d | 50% done |` + "\n"
		if note := store.files[auditDir+"/2026-11-22.md"]; !strings.Contains(note, want) {
			t.Fatalf("missing %q in %q", want, note)
		}
	})
}

func TestOSStore(t *testing.T) {
	t.Run("given a real folder, when creating and appending, then the file holds both writes", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "note.md")
		store := OSStore{}
		if store.Exists(path) {
			t.Fatal("should not exist yet")
		}
		if err := store.Create(path, []byte("one\n")); err != nil {
			t.Fatal(err)
		}
		if err := store.Append(path, []byte("two\n")); err != nil {
			t.Fatal(err)
		}
		raw, err := os.ReadFile(path)
		if err != nil || string(raw) != "one\ntwo\n" || !store.Exists(path) {
			t.Fatalf("got %q, %v", raw, err)
		}
	})

	t.Run("given a file that does not exist, when appending, then it is created", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "index.md")
		if err := (OSStore{}).Append(path, []byte("line\n")); err != nil {
			t.Fatal(err)
		}
		if raw, err := os.ReadFile(path); err != nil || string(raw) != "line\n" {
			t.Fatalf("got %q, %v", raw, err)
		}
	})

	t.Run("given a folder that does not exist, when creating, then it fails", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "missing", "note.md")
		if err := (OSStore{}).Create(path, []byte("x")); err == nil {
			t.Fatal("expected an error")
		}
	})
}

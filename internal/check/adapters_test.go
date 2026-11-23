package check

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"
)

func TestExecRunner(t *testing.T) {
	t.Run("given a command, when running it, then it runs in the folder and streams both outputs", func(t *testing.T) {
		dir := t.TempDir()
		var out, errs bytes.Buffer
		runner := ExecRunner{Dir: dir, Out: &out, Err: &errs}
		err := runner.Run(t.Context(), Invocation{Name: "sh", Args: []string{"-c", "pwd; echo oops >&2"}})
		if err != nil {
			t.Fatal(err)
		}
		resolved, _ := filepath.EvalSymlinks(dir)
		if got := strings.TrimSpace(out.String()); got != resolved {
			t.Fatalf("got %q, want %q", got, resolved)
		}
		if errs.String() != "oops\n" {
			t.Fatalf("got %q", errs.String())
		}
	})

	t.Run("given a quiet command, when running it, then nothing is printed", func(t *testing.T) {
		var out, errs bytes.Buffer
		runner := ExecRunner{Dir: t.TempDir(), Out: &out, Err: &errs}
		err := runner.Run(t.Context(), Invocation{Name: "sh", Args: []string{"-c", "echo a; echo b >&2"}, Quiet: true})
		if err != nil || out.Len() != 0 || errs.Len() != 0 {
			t.Fatalf("got %v, %q, %q", err, out.String(), errs.String())
		}
	})

	t.Run("given extra variables, when running, then the command sees them and the inherited ones", func(t *testing.T) {
		t.Setenv("RANGERS_INHERITED", "yes")
		var out bytes.Buffer
		runner := ExecRunner{Dir: t.TempDir(), Out: &out, Err: &out}
		call := Invocation{Name: "sh", Args: []string{"-c", "echo $RANGERS_EXTRA $RANGERS_INHERITED"}, Env: []string{"RANGERS_EXTRA=1"}}
		if err := runner.Run(t.Context(), call); err != nil || out.String() != "1 yes\n" {
			t.Fatalf("got %v, %q", err, out.String())
		}
	})

	t.Run("given a command that fails, when running it, then the error is returned", func(t *testing.T) {
		runner := ExecRunner{Dir: t.TempDir(), Out: &bytes.Buffer{}, Err: &bytes.Buffer{}}
		if err := runner.Run(t.Context(), Invocation{Name: "sh", Args: []string{"-c", "exit 3"}}); err == nil {
			t.Fatal("expected an error")
		}
	})

	t.Run("given a command that does not exist, when running it, then the error is returned", func(t *testing.T) {
		runner := ExecRunner{Dir: t.TempDir(), Out: &bytes.Buffer{}, Err: &bytes.Buffer{}}
		if err := runner.Run(t.Context(), Invocation{Name: "rangers-no-such-command"}); err == nil {
			t.Fatal("expected an error")
		}
	})

	t.Run("given a command that prints, when reading its output, then only standard output comes back", func(t *testing.T) {
		var errs bytes.Buffer
		runner := ExecRunner{Dir: t.TempDir(), Out: &bytes.Buffer{}, Err: &errs}
		got, err := runner.Output(t.Context(), "sh", "-c", "echo a.go; echo warn >&2")
		if err != nil || got != "a.go\n" || errs.String() != "warn\n" {
			t.Fatalf("got %q, %v, %q", got, err, errs.String())
		}
	})
}

func TestWholeWord(t *testing.T) {
	tests := []struct {
		name, pattern, text string
		want                bool
	}{
		{"given a word between spaces, when matching, then it matches", "Foo|Bar", "a Foo b", true},
		{"given the word at the line edges, when matching, then it matches", "Foo", "Foo", true},
		{"given the word inside a longer one, when matching, then it does not match", "Foo", "Foolish", false},
		{"given the word after a letter, when matching, then it does not match", "Foo", "xFoo", false},
		{"given the word next to a digit or underscore, when matching, then it does not match", "Foo", "Foo_1", false},
		{"given a name with a space, when matching, then it matches", "Bar Baz", "see Bar Baz.", true},
		{"given a pattern on several lines, when matching, then each line is an alternative", "Foo\nBar", "x Bar y", true},
		{"given another case, when matching, then it does not match", "Foo", "foo", false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			re, err := WholeWord(tc.pattern)
			if err != nil {
				t.Fatal(err)
			}
			if got := re.MatchString(tc.text); got != tc.want {
				t.Fatalf("got %v", got)
			}
		})
	}

	t.Run("given a pattern with a trailing newline, when compiling, then the newline is dropped", func(t *testing.T) {
		re, err := WholeWord("Foo\n")
		if err != nil || !re.MatchString("a Foo") || re.MatchString("a Fo") {
			t.Fatalf("got %v", err)
		}
	})

	t.Run("given an invalid pattern, when compiling, then it returns an error", func(t *testing.T) {
		if _, err := WholeWord("(open"); err == nil {
			t.Fatal("expected an error")
		}
	})
}

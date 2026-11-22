package gate

import (
	"strings"
	"testing"
)

func TestExecRunner(t *testing.T) {
	t.Run("given a command on the PATH, when asking if it exists, then it says yes", func(t *testing.T) {
		if !(ExecRunner{}).Has("sh") {
			t.Fatal("sh should exist")
		}
	})

	t.Run("given a command that is not on the PATH, when asking if it exists, then it says no", func(t *testing.T) {
		if (ExecRunner{}).Has("rangers-no-such-command") {
			t.Fatal("the command should not exist")
		}
	})

	t.Run("given a command that prints, when reading the output, then it returns what it printed", func(t *testing.T) {
		out, err := ExecRunner{}.Output(t.Context(), "echo", "hello")
		if err != nil || out != "hello\n" {
			t.Fatalf("got %q, %v", out, err)
		}
	})

	t.Run("given a command that fails, when reading the output, then the error carries the command and its stderr and the stdout stays", func(t *testing.T) {
		out, err := ExecRunner{}.Output(t.Context(), "sh", "-c", "echo [] ; echo boom >&2; exit 8")
		if err == nil || !strings.Contains(err.Error(), "sh -c") || !strings.Contains(err.Error(), "boom") {
			t.Fatalf("got %v", err)
		}
		if out != "[]\n" {
			t.Fatalf("got output %q", out)
		}
	})
}

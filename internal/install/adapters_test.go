package install

import (
	"bytes"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

func TestMain(m *testing.M) {
	for _, pair := range os.Environ() {
		name, _, _ := strings.Cut(pair, "=")
		if strings.HasPrefix(name, "GIT_") {
			_ = os.Unsetenv(name)
		}
	}
	os.Exit(m.Run())
}

func write(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o750); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
}

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

	t.Run("given a command that fails, when reading the output, then the error carries the command and its output", func(t *testing.T) {
		out, err := ExecRunner{}.Output(t.Context(), "sh", "-c", "echo boom; exit 3")
		if err == nil || !strings.Contains(err.Error(), "sh -c") || !strings.Contains(err.Error(), "boom") {
			t.Fatalf("got %v", err)
		}
		if out != "boom\n" {
			t.Fatalf("got output %q", out)
		}
	})

	t.Run("given a command, when streaming, then its output goes to the writer", func(t *testing.T) {
		var buf bytes.Buffer
		if err := (ExecRunner{Out: &buf}).Stream(t.Context(), "echo", "progress"); err != nil {
			t.Fatal(err)
		}
		if buf.String() != "progress\n" {
			t.Fatalf("got %q", buf.String())
		}
	})

	t.Run("given a command that fails, when streaming, then it returns an error naming the command", func(t *testing.T) {
		err := (ExecRunner{Out: &bytes.Buffer{}}).Stream(t.Context(), "sh", "-c", "exit 4")
		if err == nil || !strings.Contains(err.Error(), "sh -c") {
			t.Fatalf("got %v", err)
		}
	})
}

func TestOSFilesSymlink(t *testing.T) {
	t.Run("given a missing folder for the link, when linking, then it creates the folder and the link", func(t *testing.T) {
		dir := t.TempDir()
		write(t, filepath.Join(dir, "src", "a.md"), "a")
		link := filepath.Join(dir, "claude", "agents", "a.md")
		if err := (OSFiles{}).Symlink(filepath.Join(dir, "src", "a.md"), link); err != nil {
			t.Fatal(err)
		}
		got, err := os.Readlink(link)
		if err != nil || got != filepath.Join(dir, "src", "a.md") {
			t.Fatalf("got %q, %v", got, err)
		}
	})

	t.Run("given a link that points elsewhere, when linking, then it points to the new target", func(t *testing.T) {
		dir := t.TempDir()
		link := filepath.Join(dir, "link")
		if err := os.Symlink(filepath.Join(dir, "old"), link); err != nil {
			t.Fatal(err)
		}
		if err := (OSFiles{}).Symlink(filepath.Join(dir, "new"), link); err != nil {
			t.Fatal(err)
		}
		if got, _ := os.Readlink(link); got != filepath.Join(dir, "new") {
			t.Fatalf("got %q", got)
		}
	})

	t.Run("given a link that already points to the target, when linking twice, then it stays the same", func(t *testing.T) {
		dir := t.TempDir()
		link := filepath.Join(dir, "link")
		for range 2 {
			if err := (OSFiles{}).Symlink(filepath.Join(dir, "t"), link); err != nil {
				t.Fatal(err)
			}
		}
		if got, _ := os.Readlink(link); got != filepath.Join(dir, "t") {
			t.Fatalf("got %q", got)
		}
	})

	t.Run("given a regular file in the place of the link, when linking, then the file becomes the link", func(t *testing.T) {
		dir := t.TempDir()
		link := filepath.Join(dir, "x.md")
		write(t, link, "old")
		if err := (OSFiles{}).Symlink(filepath.Join(dir, "t"), link); err != nil {
			t.Fatal(err)
		}
		if got, _ := os.Readlink(link); got != filepath.Join(dir, "t") {
			t.Fatalf("got %q", got)
		}
	})

	t.Run("given a real folder in the place of the link, when linking, then it refuses and keeps the folder", func(t *testing.T) {
		dir := t.TempDir()
		link := filepath.Join(dir, "rangers")
		write(t, filepath.Join(link, "SKILL.md"), "mine")
		err := (OSFiles{}).Symlink(filepath.Join(dir, "t"), link)
		if err == nil || !strings.Contains(err.Error(), "is a directory") {
			t.Fatalf("got %v", err)
		}
		if _, statErr := os.Stat(filepath.Join(link, "SKILL.md")); statErr != nil {
			t.Fatal("the folder was changed")
		}
	})
}

func TestOSFilesCopyDir(t *testing.T) {
	t.Run("given a tree with a .git folder, when copying, then it copies everything but .git", func(t *testing.T) {
		dir := t.TempDir()
		src, dst := filepath.Join(dir, "src"), filepath.Join(dir, "dst")
		write(t, filepath.Join(src, "SKILL.md"), "skill")
		write(t, filepath.Join(src, "sub", "deep.txt"), "deep")
		write(t, filepath.Join(src, ".git", "HEAD"), "ref")
		if err := (OSFiles{}).CopyDir(src, dst); err != nil {
			t.Fatal(err)
		}
		for path, want := range map[string]string{"SKILL.md": "skill", "sub/deep.txt": "deep"} {
			raw, err := os.ReadFile(filepath.Join(dst, path))
			if err != nil || string(raw) != want {
				t.Errorf("%s: got %q, %v", path, raw, err)
			}
		}
		if _, err := os.Stat(filepath.Join(dst, ".git")); err == nil {
			t.Error(".git was copied")
		}
	})

	t.Run("given a read-only file and a link, when copying, then the mode and the link are kept", func(t *testing.T) {
		dir := t.TempDir()
		src, dst := filepath.Join(dir, "src"), filepath.Join(dir, "dst")
		write(t, filepath.Join(src, "run.sh"), "x")
		if err := os.Chmod(filepath.Join(src, "run.sh"), 0o400); err != nil {
			t.Fatal(err)
		}
		if err := os.Symlink("run.sh", filepath.Join(src, "alias")); err != nil {
			t.Fatal(err)
		}
		if err := (OSFiles{}).CopyDir(src, dst); err != nil {
			t.Fatal(err)
		}
		info, err := os.Stat(filepath.Join(dst, "run.sh"))
		if err != nil || info.Mode().Perm() != 0o400 {
			t.Fatalf("got %v, %v", info, err)
		}
		if got, _ := os.Readlink(filepath.Join(dst, "alias")); got != "run.sh" {
			t.Fatalf("got link %q", got)
		}
	})

	t.Run("given a source that does not exist, when copying, then it returns the error", func(t *testing.T) {
		if err := (OSFiles{}).CopyDir(filepath.Join(t.TempDir(), "none"), t.TempDir()); err == nil {
			t.Fatal("expected an error")
		}
	})
}

func TestOSFilesBasics(t *testing.T) {
	t.Run("given a dangling link, when asking if it exists, then it says yes", func(t *testing.T) {
		link := filepath.Join(t.TempDir(), "dangling")
		if err := os.Symlink("/nowhere", link); err != nil {
			t.Fatal(err)
		}
		if !(OSFiles{}).Exists(link) {
			t.Fatal("the link should count as existing")
		}
	})

	t.Run("given a file, when writing and reading it, then the content comes back", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "marker")
		if err := (OSFiles{}).WriteFile(path, []byte("abc\n")); err != nil {
			t.Fatal(err)
		}
		raw, err := (OSFiles{}).ReadFile(path)
		if err != nil || string(raw) != "abc\n" {
			t.Fatalf("got %q, %v", raw, err)
		}
	})

	t.Run("given a pattern, when globbing, then it lists the matches", func(t *testing.T) {
		dir := t.TempDir()
		write(t, filepath.Join(dir, "a.md"), "a")
		write(t, filepath.Join(dir, "b.txt"), "b")
		got, err := (OSFiles{}).Glob(filepath.Join(dir, "*.md"))
		if err != nil || len(got) != 1 || got[0] != filepath.Join(dir, "a.md") {
			t.Fatalf("got %v, %v", got, err)
		}
	})

	t.Run("given a temp folder, when removing it, then it is gone", func(t *testing.T) {
		tmp, err := (OSFiles{}).MkdirTemp()
		if err != nil {
			t.Fatal(err)
		}
		if err = (OSFiles{}).RemoveAll(tmp); err != nil {
			t.Fatal(err)
		}
		if (OSFiles{}).Exists(tmp) {
			t.Fatal("the folder is still there")
		}
	})
}

func TestOSFilesKinds(t *testing.T) {
	t.Run("given a link, when reading it, then it returns the target", func(t *testing.T) {
		link := filepath.Join(t.TempDir(), "link")
		if err := os.Symlink("/some/target", link); err != nil {
			t.Fatal(err)
		}
		got, err := (OSFiles{}).Readlink(link)
		if err != nil || got != "/some/target" {
			t.Fatalf("got %q, %v", got, err)
		}
	})

	t.Run("given a regular file, when reading it as a link, then it fails", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "file")
		write(t, path, "x")
		if _, err := (OSFiles{}).Readlink(path); err == nil {
			t.Fatal("a regular file has no link target")
		}
	})

	t.Run("given a file, a link, a folder and nothing, when asking if each is a regular file, then only the file is", func(t *testing.T) {
		dir := t.TempDir()
		file := filepath.Join(dir, "file")
		write(t, file, "x")
		link := filepath.Join(dir, "link")
		if err := os.Symlink(file, link); err != nil {
			t.Fatal(err)
		}
		got := []bool{
			(OSFiles{}).IsRegular(file),
			(OSFiles{}).IsRegular(link),
			(OSFiles{}).IsRegular(dir),
			(OSFiles{}).IsRegular(filepath.Join(dir, "missing")),
		}
		want := []bool{true, false, false, false}
		if !slices.Equal(got, want) {
			t.Fatalf("got %v, want %v", got, want)
		}
	})
}

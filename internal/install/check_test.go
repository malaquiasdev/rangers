package install

import (
	"errors"
	"fmt"
	"slices"
	"strings"
	"testing"
)

func row(name, status string) string {
	return fmt.Sprintf("%-13s  %s\n", name, status)
}

func TestCheck(t *testing.T) {
	t.Run("given every piece in place, when checking, then it prints one line per piece and All in place", func(t *testing.T) {
		s := newSetup(t)
		s.allInstalled()
		if err := s.run("--check"); err != nil {
			t.Fatal(err)
		}
		want := row("pstack", "0.15.46") +
			row("openspec", "1.7.0") +
			row("container-use", "container-use version 0.4.2") +
			row("lychee", "0.24.2") +
			row("grilling", "abeeaf95660b") +
			row("docker", "responding") +
			"\nAll in place.\n"
		if s.typed() != want {
			t.Fatalf("got\n%s\nwant\n%s", s.typed(), want)
		}
	})

	t.Run("given pieces missing, when checking, then it marks them, counts them and fails", func(t *testing.T) {
		s := newSetup(t)
		err := s.run("--check")
		if !errors.Is(err, ErrIncomplete) {
			t.Fatalf("got %v", err)
		}
		want := row("pstack", "MISSING") +
			row("openspec", "not installed (optional)") +
			row("container-use", "MISSING") +
			row("lychee", "MISSING") +
			row("grilling", "MISSING") +
			row("docker", "NOT RESPONDING (container-use needs it)") +
			"\n5 item(s) missing. Run rangers install\n"
		if s.typed() != want {
			t.Fatalf("got\n%s\nwant\n%s", s.typed(), want)
		}
	})

	t.Run("given the optional flag, when the optional piece is missing, then it counts as missing", func(t *testing.T) {
		s := newSetup(t)
		s.allInstalled()
		s.runner.present["openspec"] = false
		err := s.run("--check", "--optional")
		if !errors.Is(err, ErrIncomplete) || !strings.Contains(s.typed(), row("openspec", "MISSING")) {
			t.Fatalf("got %v\n%s", err, s.typed())
		}
		if !strings.HasSuffix(s.typed(), "\n1 item(s) missing. Run rangers install\n") {
			t.Fatalf("got\n%s", s.typed())
		}
	})

	t.Run("given a skill folder without the marker, when checking, then it shows present, unpinned and counts it in place", func(t *testing.T) {
		s := newSetup(t)
		s.allInstalled()
		delete(s.files.files, claudeDir+"/skills/grilling/.rangers-commit")
		s.files.files[claudeDir+"/skills/grilling/SKILL.md"] = "x"
		if err := s.run("--check"); err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(s.typed(), row("grilling", "present, unpinned")) {
			t.Fatalf("got\n%s", s.typed())
		}
	})

	t.Run("given a claude dir flag, when checking, then it looks for the skills there", func(t *testing.T) {
		s := newSetup(t)
		s.allInstalled()
		s.files.files["/other/skills/grilling/.rangers-commit"] = "1234567890abcdef\n"
		if err := s.run("--check", "--claude-dir", "/other"); err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(s.typed(), row("grilling", "1234567890ab")) {
			t.Fatalf("got\n%s", s.typed())
		}
	})

	t.Run("given two plugins, when checking, then each shows the version listed under its own name", func(t *testing.T) {
		s := newSetup(t)
		s.allInstalled()
		s.runner.replies["claude plugin list"] = reply{out: "  ❯ pstack@cursor-plugins\n    Scope: user\n\n  ❯ other@x\n    Version: 9.9.9\n"}
		if err := s.run("--check"); err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(s.typed(), row("pstack", "9.9.9")) {
			t.Fatalf("got\n%s", s.typed())
		}
	})

	t.Run("given a plugin listed without a version, when checking, then it shows installed", func(t *testing.T) {
		s := newSetup(t)
		s.allInstalled()
		s.runner.replies["claude plugin list"] = reply{out: "  ❯ pstack@cursor-plugins\n"}
		if err := s.run("--check"); err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(s.typed(), row("pstack", "installed")) {
			t.Fatalf("got\n%s", s.typed())
		}
	})

	t.Run("given a command that exists but gives no version, when checking, then it shows installed", func(t *testing.T) {
		s := newSetup(t)
		s.allInstalled()
		s.runner.replies["openspec --version"] = reply{err: errors.New("boom")}
		if err := s.run("--check"); err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(s.typed(), row("openspec", "installed")) {
			t.Fatalf("got\n%s", s.typed())
		}
	})
}

func TestCheckChangesNothing(t *testing.T) {
	t.Run("given a machine with nothing installed, when checking, then it only runs read commands and writes no file", func(t *testing.T) {
		s := newSetup(t)
		_ = s.run("--check")
		want := []string{
			"claude plugin list",
			"podman image exists docker.io/lycheeverse/lychee:0.24.2",
			"docker info",
		}
		if !slices.Equal(s.runner.calls, want) {
			t.Fatalf("got %v", s.runner.calls)
		}
		if len(s.files.written) != 0 || len(s.files.links) != 0 || s.files.tempDirs != 0 {
			t.Fatalf("changed files: %v %v", s.files.written, s.files.links)
		}
	})
}

func TestCommandArguments(t *testing.T) {
	t.Run("given a folder inside the repository, when running, then it finds deps.toml above it", func(t *testing.T) {
		s := newSetup(t)
		s.allInstalled()
		s.dir = repoDir + "/docs/how-to"
		if err := s.run("--check"); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("given the repo flag, when running from elsewhere, then it reads that repository", func(t *testing.T) {
		s := newSetup(t)
		s.allInstalled()
		s.dir = "/elsewhere"
		if err := s.run("--check", "--repo", repoDir); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("given no deps.toml above the folder, when running, then the error says where it looked", func(t *testing.T) {
		s := newSetup(t)
		s.dir = "/elsewhere"
		err := s.run("--check")
		if err == nil || !strings.Contains(err.Error(), "deps.toml") || !strings.Contains(err.Error(), "--repo") {
			t.Fatalf("got %v", err)
		}
	})

	t.Run("given an unknown flag, when running, then it returns the error", func(t *testing.T) {
		s := newSetup(t)
		if err := s.run("--nope"); err == nil {
			t.Fatal("expected an error")
		}
	})

	t.Run("given a stray argument, when running, then it prints the usage", func(t *testing.T) {
		s := newSetup(t)
		err := s.run("--check", "extra")
		if err == nil || !strings.Contains(err.Error(), "usage: rangers install") {
			t.Fatalf("got %v", err)
		}
	})

	t.Run("given an invalid deps.toml, when running, then it returns the parse error", func(t *testing.T) {
		s := newSetup(t)
		s.files.files[repoDir+"/deps.toml"] = "format_version = 2\n"
		err := s.run("--check")
		if err == nil || !strings.Contains(err.Error(), "format_version") {
			t.Fatalf("got %v", err)
		}
	})
}

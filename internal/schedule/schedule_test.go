package schedule

import (
	"bytes"
	"context"
	"encoding/xml"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const launchctlPrint = `com.rangers.learn = {
	active count = 0
	path = /Users/x/Library/LaunchAgents/com.rangers.learn.plist
	state = not running

	program = /Users/x/.local/bin/rangers
	arguments = {
		/Users/x/.local/bin/rangers
		learn
	}

	runs = 3
	last exit code = 0
	run interval = 7200 seconds

	event triggers = {
		com.rangers.learn.timer => {
			state = active
		}
	}
}
`

type fakeLaunchctl struct {
	calls        []string
	bootoutErr   error
	bootstrapErr error
	printOut     string
	printErr     error
}

type setup struct {
	t       *testing.T
	home    string
	launchd *fakeLaunchctl
	out     *bytes.Buffer
}

func (f *fakeLaunchctl) Bootout(_ context.Context, target string) error {
	f.calls = append(f.calls, "bootout "+target)
	return f.bootoutErr
}

func (f *fakeLaunchctl) Bootstrap(_ context.Context, domain, plist string) error {
	f.calls = append(f.calls, "bootstrap "+domain+" "+plist)
	return f.bootstrapErr
}

func (f *fakeLaunchctl) Print(_ context.Context, target string) (string, error) {
	f.calls = append(f.calls, "print "+target)
	return f.printOut, f.printErr
}

func newSetup(t *testing.T) *setup {
	t.Helper()
	return &setup{t: t, home: t.TempDir(), launchd: &fakeLaunchctl{}, out: &bytes.Buffer{}}
}

func (s *setup) run(args ...string) error {
	env := func(name string) string {
		if name == "HOME" {
			return s.home
		}
		return ""
	}
	dep := &Dependencies{Launchctl: s.launchd, UID: 501, Out: s.out}
	return Command(s.t.Context(), args, env, dep)
}

func (s *setup) plistPath() string {
	return filepath.Join(s.home, "Library", "LaunchAgents", "com.rangers.learn.plist")
}

func readIn(t *testing.T, dir, name string) string {
	t.Helper()
	root, err := os.OpenRoot(dir)
	if err != nil {
		t.Fatal(err)
	}
	raw, err := root.ReadFile(name)
	if err != nil {
		t.Fatal(err)
	}
	if err = root.Close(); err != nil {
		t.Fatal(err)
	}
	return string(raw)
}

func (s *setup) plist() string {
	s.t.Helper()
	return readIn(s.t, s.home, "Library/LaunchAgents/com.rangers.learn.plist")
}

func TestInstallPlist(t *testing.T) {
	t.Run("given a home, when installing, then the plist has the label, the program, the interval and the run at load", func(t *testing.T) {
		s := newSetup(t)
		if err := s.run("install"); err != nil {
			t.Fatal(err)
		}
		plist := s.plist()
		for _, expected := range []string{
			"<key>Label</key>\n  <string>com.rangers.learn</string>",
			"<key>ProgramArguments</key>\n  <array>\n    <string>" + s.home + "/.local/bin/rangers</string>\n    <string>learn</string>\n  </array>",
			"<key>StartInterval</key>\n  <integer>7200</integer>",
			"<key>RunAtLoad</key>\n  <true/>",
			"<key>HOME</key>\n    <string>" + s.home + "</string>",
		} {
			if !strings.Contains(plist, expected) {
				t.Errorf("missing %q in\n%s", expected, plist)
			}
		}
	})

	t.Run("given a home, when installing, then the PATH lists the six folders in order", func(t *testing.T) {
		s := newSetup(t)
		if err := s.run("install"); err != nil {
			t.Fatal(err)
		}
		expected := "<key>PATH</key>\n    <string>" + s.home +
			"/.local/bin:/opt/homebrew/bin:/usr/local/go/bin:/usr/local/bin:/usr/bin:/bin</string>"
		if !strings.Contains(s.plist(), expected) {
			t.Fatalf("missing the PATH in\n%s", s.plist())
		}
	})

	t.Run("given a home, when installing, then standard output and error go to launchd.log in the state folder", func(t *testing.T) {
		s := newSetup(t)
		if err := s.run("install"); err != nil {
			t.Fatal(err)
		}
		log := s.home + "/.local/state/rangers/learn/launchd.log"
		for _, key := range []string{"StandardOutPath", "StandardErrorPath"} {
			if !strings.Contains(s.plist(), "<key>"+key+"</key>\n  <string>"+log+"</string>") {
				t.Errorf("missing %s in\n%s", key, s.plist())
			}
		}
		if _, err := os.Stat(filepath.Dir(log)); err != nil {
			t.Fatalf("the state folder was not created: %v", err)
		}
	})

	t.Run("given the binary flag, when installing, then the plist runs that binary", func(t *testing.T) {
		s := newSetup(t)
		if err := s.run("install", "--binary", "/opt/tools/rangers"); err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(s.plist(), "<string>/opt/tools/rangers</string>\n    <string>learn</string>") {
			t.Fatalf("plist:\n%s", s.plist())
		}
	})

	t.Run("given a path with XML characters, when installing, then the plist stays well formed and escaped", func(t *testing.T) {
		s := newSetup(t)
		if err := s.run("install", "--binary", "/opt/a&b/<rangers>"); err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(s.plist(), "/opt/a&amp;b/&lt;rangers&gt;") {
			t.Fatalf("plist:\n%s", s.plist())
		}
		decoder := xml.NewDecoder(strings.NewReader(s.plist()))
		decoder.Strict = true
		for {
			if _, err := decoder.Token(); err != nil {
				if errors.Is(err, io.EOF) {
					break
				}
				t.Fatalf("the plist is not well formed: %v", err)
			}
		}
	})
}

func TestInstallLaunchctl(t *testing.T) {
	t.Run("given a user, when installing, then it calls bootout and then bootstrap on the plist", func(t *testing.T) {
		s := newSetup(t)
		if err := s.run("install"); err != nil {
			t.Fatal(err)
		}
		expected := []string{
			"bootout gui/501/com.rangers.learn",
			"bootstrap gui/501 " + s.plistPath(),
		}
		if strings.Join(s.launchd.calls, "|") != strings.Join(expected, "|") {
			t.Fatalf("expected %v, got %v", expected, s.launchd.calls)
		}
		if !strings.Contains(s.out.String(), "agendado: a cada 2 horas, log em ") {
			t.Fatalf("output: %q", s.out.String())
		}
	})

	t.Run("given a job that is not loaded, when installing, then the failed bootout is ignored", func(t *testing.T) {
		s := newSetup(t)
		s.launchd.bootoutErr = errors.New("Boot-out failed: 5: Input/output error")
		if err := s.run("install"); err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if len(s.launchd.calls) != 2 {
			t.Fatalf("calls: %v", s.launchd.calls)
		}
	})

	t.Run("given a bootstrap that fails, when installing, then it returns the error", func(t *testing.T) {
		s := newSetup(t)
		s.launchd.bootstrapErr = errors.New("Bootstrap failed: 5")
		err := s.run("install")
		if err == nil || !strings.Contains(err.Error(), "Bootstrap failed: 5") {
			t.Fatalf("got %v", err)
		}
	})

	t.Run("given a job already installed, when installing again, then the plist is the same and the job is reloaded", func(t *testing.T) {
		s := newSetup(t)
		if err := s.run("install"); err != nil {
			t.Fatal(err)
		}
		first := s.plist()
		if err := s.run("install"); err != nil {
			t.Fatal(err)
		}
		if s.plist() != first {
			t.Fatal("the plist changed on reinstall")
		}
		if len(s.launchd.calls) != 4 || s.launchd.calls[2] != s.launchd.calls[0] || s.launchd.calls[3] != s.launchd.calls[1] {
			t.Fatalf("calls: %v", s.launchd.calls)
		}
	})
}

func TestUninstall(t *testing.T) {
	t.Run("given an installed job, when uninstalling, then it calls bootout and deletes the plist", func(t *testing.T) {
		s := newSetup(t)
		if err := s.run("install"); err != nil {
			t.Fatal(err)
		}
		s.launchd.calls = nil
		if err := s.run("uninstall"); err != nil {
			t.Fatal(err)
		}
		if strings.Join(s.launchd.calls, "|") != "bootout gui/501/com.rangers.learn" {
			t.Fatalf("calls: %v", s.launchd.calls)
		}
		if _, err := os.Stat(s.plistPath()); !errors.Is(err, os.ErrNotExist) {
			t.Fatalf("the plist should be gone: %v", err)
		}
		if !strings.HasSuffix(s.out.String(), "agendamento removido\n") {
			t.Fatalf("output: %q", s.out.String())
		}
	})

	t.Run("given no plist and a failing bootout, when uninstalling, then it still succeeds", func(t *testing.T) {
		s := newSetup(t)
		s.launchd.bootoutErr = errors.New("not loaded")
		if err := s.run("uninstall"); err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
	})
}

func TestStatus(t *testing.T) {
	t.Run("given no job, when asking the status, then it says it is not scheduled", func(t *testing.T) {
		s := newSetup(t)
		s.launchd.printErr = errors.New("Could not find service")
		if err := s.run("status"); err != nil {
			t.Fatal(err)
		}
		if s.out.String() != "não agendado\n" {
			t.Fatalf("output: %q", s.out.String())
		}
		if strings.Join(s.launchd.calls, "|") != "print gui/501/com.rangers.learn" {
			t.Fatalf("calls: %v", s.launchd.calls)
		}
	})

	t.Run("given a scheduled job, when asking the status, then it prints state, runs, last exit code and interval from the top level", func(t *testing.T) {
		s := newSetup(t)
		s.launchd.printOut = launchctlPrint
		if err := s.run("status"); err != nil {
			t.Fatal(err)
		}
		expected := "estado: not running\nrodadas: 3\núltimo código de saída: 0\nintervalo: 7200 seconds\n"
		if s.out.String() != expected {
			t.Fatalf("expected %q, got %q", expected, s.out.String())
		}
	})

	t.Run("given no subcommand, when running, then it shows the status", func(t *testing.T) {
		s := newSetup(t)
		s.launchd.printOut = launchctlPrint
		if err := s.run(); err != nil {
			t.Fatal(err)
		}
		if !strings.HasPrefix(s.out.String(), "estado: not running") {
			t.Fatalf("output: %q", s.out.String())
		}
	})

	t.Run("given an output with other spacing and a job that never exited, when asking the status, then it reads what is there", func(t *testing.T) {
		s := newSetup(t)
		s.launchd.printOut = "\r\n  state=running\r\n\t\tlast exit code = (never exited)\r\n  unrelated line\r\n"
		if err := s.run("status"); err != nil {
			t.Fatal(err)
		}
		expected := "estado: running\núltimo código de saída: (never exited)\n"
		if s.out.String() != expected {
			t.Fatalf("expected %q, got %q", expected, s.out.String())
		}
	})

	t.Run("given an output with none of the known lines, when asking the status, then it says it could not read the state", func(t *testing.T) {
		s := newSetup(t)
		s.launchd.printOut = "something else entirely\n"
		if err := s.run("status"); err != nil {
			t.Fatal(err)
		}
		if s.out.String() != "agendado, mas não consegui ler o estado\n" {
			t.Fatalf("output: %q", s.out.String())
		}
	})
}

func TestCommandArguments(t *testing.T) {
	t.Run("given an unknown subcommand, when running, then it returns the usage and calls nothing", func(t *testing.T) {
		s := newSetup(t)
		err := s.run("restart")
		if err == nil || !strings.Contains(err.Error(), "uso: rangers schedule install|uninstall|status") {
			t.Fatalf("got %v", err)
		}
		if len(s.launchd.calls) != 0 {
			t.Fatalf("calls: %v", s.launchd.calls)
		}
	})

	t.Run("given an unknown flag, when installing, then it returns an error and writes no plist", func(t *testing.T) {
		s := newSetup(t)
		if err := s.run("install", "--nope"); err == nil {
			t.Fatal("expected an error")
		}
		if _, err := os.Stat(s.plistPath()); !errors.Is(err, os.ErrNotExist) {
			t.Fatalf("no plist expected: %v", err)
		}
	})
}

func TestExecLaunchctl(t *testing.T) {
	t.Run("given a launchctl on the PATH, when calling each operation, then it runs the right command and returns its output", func(t *testing.T) {
		bin := t.TempDir()
		log := filepath.Join(bin, "calls.log")
		script := "#!/bin/sh\nprintf '%s\\n' \"$*\" >> " + log + "\necho printed\n"
		root, err := os.OpenRoot(bin)
		if err != nil {
			t.Fatal(err)
		}
		file, err := root.OpenFile("launchctl", os.O_WRONLY|os.O_CREATE, 0o600)
		if err != nil {
			t.Fatal(err)
		}
		if _, err = file.WriteString(script); err != nil {
			t.Fatal(err)
		}
		if err = errors.Join(file.Chmod(0o500), file.Close(), root.Close()); err != nil {
			t.Fatal(err)
		}
		t.Setenv("PATH", bin)
		runner := ExecLaunchctl{}
		out, err := runner.Print(t.Context(), "gui/501/x")
		if err != nil || strings.TrimSpace(out) != "printed" {
			t.Fatalf("got %q, %v", out, err)
		}
		if err = errors.Join(runner.Bootout(t.Context(), "gui/501/x"), runner.Bootstrap(t.Context(), "gui/501", "/p.plist")); err != nil {
			t.Fatal(err)
		}
		expected := "print gui/501/x\nbootout gui/501/x\nbootstrap gui/501 /p.plist\n"
		if raw := readIn(t, bin, "calls.log"); raw != expected {
			t.Fatalf("calls: %q", raw)
		}
	})
}

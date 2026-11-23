package check

import (
	"errors"
	"strings"
	"testing"
)

const (
	slashes     = "/" + "/"
	commentFail = "FALHOU: comentário em arquivo Go\n"
	lychee      = "podman run --rm -v /repo:/w:ro -w /w docker.io/lycheeverse/lychee:0.24.2 " +
		"--offline --include-fragments README.md docs/**/*.md"
	python = "python3 -I -m unittest discover -s assets/tests -t assets"
)

func goRepo() map[string]string {
	return map[string]string{"go.mod": "module x\n", "main.go": "package main\n"}
}

func withGo(extra map[string]string) *harness {
	files := clean()
	for path, content := range goRepo() {
		files[path] = file(content)
	}
	for path, content := range extra {
		files[path] = file(content)
	}
	return newHarness(files)
}

func TestGoComments(t *testing.T) {
	t.Run("given a comment line in a Go file, when running, then the line is printed and the check fails", func(t *testing.T) {
		h := withGo(map[string]string{"internal/x/y.go": "package x\n" + slashes + " note\n"})
		if err := h.run(t, "quick"); !errors.Is(err, ErrFailed) {
			t.Fatalf("got %v", err)
		}
		want := "./internal/x/y.go:2:" + slashes + " note\n" + commentFail
		if !strings.Contains(h.out.String(), want) {
			t.Fatalf("missing %q in %q", want, h.out.String())
		}
	})

	t.Run("given a trailing comment or a block comment opening, when running, then they are found", func(t *testing.T) {
		for _, line := range []string{"x := 1 " + slashes + " why", "\t/* block", "/* block"} {
			h := withGo(map[string]string{"a.go": "package a\n" + line + "\n"})
			if err := h.run(t, "quick"); !errors.Is(err, ErrFailed) {
				t.Fatalf("%q: got %v", line, err)
			}
		}
	})

	t.Run("given a URL or an inline block opening, when running, then they are not comments", func(t *testing.T) {
		content := "package a\n\nvar u = \"https:" + slashes + "example.test\"\nvar v = 1 /* not at the start\n"
		h := withGo(map[string]string{"a.go": content})
		if err := h.run(t, "quick"); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("given a comment outside Go files and in the serena folder, when running, then it is ignored", func(t *testing.T) {
		h := withGo(map[string]string{"docs/x.md": slashes + " note\n", ".serena/a.go": slashes + " note\n"})
		if err := h.run(t, "quick"); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("given a comment but no go.mod, when running, then no Go check runs", func(t *testing.T) {
		files := clean()
		files["a.go"] = file(slashes + " note\n")
		h := newHarness(files)
		if err := h.run(t, "quick"); err != nil {
			t.Fatal(err)
		}
		if h.called("gofmt") {
			t.Fatalf("got %q", h.ranked())
		}
	})
}

func TestGoFormat(t *testing.T) {
	t.Run("given files that gofmt lists, when running, then the failure lists them as the Bash script did", func(t *testing.T) {
		h := withGo(nil)
		h.runner.outputs["gofmt -l ."] = "a.go\nb.go\n"
		if err := h.run(t, "quick"); !errors.Is(err, ErrFailed) {
			t.Fatalf("got %v", err)
		}
		if !strings.Contains(h.out.String(), "FALHOU: gofmt: a.go b.go \n") {
			t.Fatalf("got %q", h.out.String())
		}
	})

	t.Run("given nothing listed by gofmt, when running quickly, then only gofmt runs", func(t *testing.T) {
		h := withGo(nil)
		if err := h.run(t, "quick"); err != nil {
			t.Fatal(err)
		}
		if got := h.ranked(); got != "gofmt -l ." {
			t.Fatalf("got %q", got)
		}
	})
}

func TestFullMode(t *testing.T) {
	t.Run("given a Go repository, when running in full mode, then every tool runs in the original order", func(t *testing.T) {
		h := withGo(nil)
		if err := h.run(t, "full"); err != nil {
			t.Fatal(err)
		}
		want := strings.Join([]string{"gofmt -l .", "go vet ./...", "go test ./...", "golangci-lint run", python, lychee}, "\n")
		if got := h.ranked(); got != want {
			t.Fatalf("got\n%s\nwant\n%s", got, want)
		}
	})

	t.Run("given a repository without Go, when running in full mode, then the assets tests and the links still run", func(t *testing.T) {
		h := newHarness(clean())
		if err := h.run(t, "full"); err != nil {
			t.Fatal(err)
		}
		if got, want := h.ranked(), python+"\n"+lychee; got != want {
			t.Fatalf("got %q, want %q", got, want)
		}
	})

	t.Run("given the slow tools, when they run, then the Python tests and the links are quiet and the others are not", func(t *testing.T) {
		h := withGo(nil)
		if err := h.run(t, "full"); err != nil {
			t.Fatal(err)
		}
		for call, quiet := range map[string]bool{
			"go vet ./...": false, "go test ./...": false, "golangci-lint run": false, python: true, lychee: true,
		} {
			if h.runner.quiet[call] != quiet {
				t.Fatalf("%s: quiet is %v", call, h.runner.quiet[call])
			}
		}
	})

	t.Run("given the Python tests, when they run, then they do not write bytecode", func(t *testing.T) {
		h := newHarness(clean())
		if err := h.run(t, "full"); err != nil {
			t.Fatal(err)
		}
		if env := h.runner.envs[python]; len(env) != 1 || env[0] != "PYTHONDONTWRITEBYTECODE=1" {
			t.Fatalf("got %v", env)
		}
	})

	failures := map[string]string{
		"go vet ./...":      "go vet",
		"go test ./...":     "go test",
		"golangci-lint run": "golangci-lint",
		python:              "testes dos assets",
		lychee:              "links da documentação",
	}
	for call, message := range failures {
		t.Run("given "+message+" failing, when running in full mode, then it fails with that name and goes on", func(t *testing.T) {
			h := withGo(nil)
			h.runner.failing[call] = true
			if err := h.run(t, "full"); !errors.Is(err, ErrFailed) {
				t.Fatalf("got %v", err)
			}
			if !strings.Contains(h.out.String(), "FALHOU: "+message+"\n") {
				t.Fatalf("got %q", h.out.String())
			}
			if !h.called("podman") {
				t.Fatalf("stopped early: %q", h.ranked())
			}
		})
	}

	t.Run("given every slow tool failing, when running in full mode, then all five are counted", func(t *testing.T) {
		h := withGo(nil)
		for call := range failures {
			h.runner.failing[call] = true
		}
		if err := h.run(t, "full"); !errors.Is(err, ErrFailed) {
			t.Fatalf("got %v", err)
		}
		if !strings.HasSuffix(h.out.String(), "\n5 checagem(ns) falharam\n") {
			t.Fatalf("got %q", h.out.String())
		}
	})
}

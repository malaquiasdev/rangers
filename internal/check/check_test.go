package check

import (
	"errors"
	"strings"
	"testing"
)

func TestSummary(t *testing.T) {
	t.Run("given a clean repository, when running the full checks, then it prints the warning and the success line and returns nil", func(t *testing.T) {
		h := newHarness(clean())
		if err := h.run(t, "full"); err != nil {
			t.Fatal(err)
		}
		if got, want := h.out.String(), privacyWarning+"tudo certo (full)\n"; got != want {
			t.Fatalf("got %q, want %q", got, want)
		}
	})

	t.Run("given no argument, when running, then the mode is full", func(t *testing.T) {
		h := newHarness(clean())
		if err := h.run(t); err != nil {
			t.Fatal(err)
		}
		if !strings.HasSuffix(h.out.String(), "tudo certo (full)\n") || !h.called("podman run") {
			t.Fatalf("got %q after %q", h.out.String(), h.ranked())
		}
	})

	t.Run("given the quick mode, when running, then the success line says quick and no slow check runs", func(t *testing.T) {
		files := clean()
		files["go.mod"] = file("module x\n")
		h := newHarness(files)
		if err := h.run(t, "quick"); err != nil {
			t.Fatal(err)
		}
		if !strings.HasSuffix(h.out.String(), "tudo certo (quick)\n") {
			t.Fatalf("got %q", h.out.String())
		}
		for _, slow := range []string{"go vet", "go test", "golangci-lint", "python3", "podman"} {
			if h.called(slow) {
				t.Fatalf("%s ran in quick mode: %s", slow, h.ranked())
			}
		}
	})

	t.Run("given one failing check, when running, then it prints one FALHOU line, the count and returns the failure error", func(t *testing.T) {
		files := clean()
		files["skills/pilot/SKILL.md"] = file("---\nname: other\n---\n")
		h := newHarness(files)
		err := h.run(t, "quick")
		if !errors.Is(err, ErrFailed) {
			t.Fatalf("got %v", err)
		}
		want := privacyWarning + "FALHOU: skills/pilot/SKILL.md: name diferente da pasta\n\n1 checagem(ns) falharam\n"
		if got := h.out.String(); got != want {
			t.Fatalf("got %q, want %q", got, want)
		}
	})

	t.Run("given two failing checks, when running, then the count is two", func(t *testing.T) {
		files := clean()
		files["skills/pilot/SKILL.md"] = file("---\nname: other\n---\n")
		files["agents/alpha.md"] = file(strings.Replace(agent("alpha"), "sonnet", "gpt", 1))
		h := newHarness(files)
		if err := h.run(t, "quick"); !errors.Is(err, ErrFailed) {
			t.Fatalf("got %v", err)
		}
		if !strings.HasSuffix(h.out.String(), "\n2 checagem(ns) falharam\n") {
			t.Fatalf("got %q", h.out.String())
		}
	})

	t.Run("given a repository without the names file, when running, then it stops with an error and no verdict", func(t *testing.T) {
		files := clean()
		delete(files, "scripts/characters.txt")
		h := newHarness(files)
		err := h.run(t, "quick")
		if err == nil || errors.Is(err, ErrFailed) || strings.Contains(h.out.String(), "tudo certo") {
			t.Fatalf("got %v and %q", err, h.out.String())
		}
	})
}

func TestShellFiles(t *testing.T) {
	t.Run("given a hook, a script and a bin file, when running, then each is checked with bash -n in that order", func(t *testing.T) {
		files := clean()
		files["scripts/x.sh"] = file("#!/usr/bin/env bash\ntrue\n")
		files[".githooks/pre-commit"] = file("#!/usr/bin/env bash\ntrue\n")
		files["bin/tool"] = file("#!/usr/bin/env bash\ntrue\n")
		h := newHarness(files)
		if err := h.run(t, "quick"); err != nil {
			t.Fatal(err)
		}
		want := "bash -n scripts/x.sh\nbash -n .githooks/pre-commit\nbash -n bin/tool"
		if got := h.ranked(); got != want {
			t.Fatalf("got %q, want %q", got, want)
		}
		if h.runner.quiet["bash -n bin/tool"] {
			t.Fatal("bash -n must not be quiet")
		}
	})

	t.Run("given a file with a syntax error, when running, then it fails naming the file", func(t *testing.T) {
		files := clean()
		files[".githooks/pre-push"] = file("#!/usr/bin/env bash\n")
		h := newHarness(files)
		h.runner.failing["bash -n .githooks/pre-push"] = true
		if err := h.run(t, "quick"); !errors.Is(err, ErrFailed) {
			t.Fatalf("got %v", err)
		}
		if !strings.Contains(h.out.String(), "FALHOU: sintaxe de .githooks/pre-push\n") {
			t.Fatalf("got %q", h.out.String())
		}
	})

	t.Run("given a directory in the hooks folder, when running, then it is skipped", func(t *testing.T) {
		files := clean()
		files[".githooks/nested/inner"] = file("#!/usr/bin/env bash\n")
		h := newHarness(files)
		if err := h.run(t, "quick"); err != nil {
			t.Fatal(err)
		}
		if h.called("bash") {
			t.Fatalf("got %q", h.ranked())
		}
	})

	t.Run("given a comment after the first line, when running, then it fails naming the file", func(t *testing.T) {
		for _, line := range []string{"# note", "   # indented", "\t#tab"} {
			files := clean()
			files[".githooks/pre-commit"] = file("#!/usr/bin/env bash\n" + line + "\ntrue\n")
			h := newHarness(files)
			if err := h.run(t, "quick"); !errors.Is(err, ErrFailed) {
				t.Fatalf("%q: got %v", line, err)
			}
			if !strings.Contains(h.out.String(), "FALHOU: comentário em .githooks/pre-commit\n") {
				t.Fatalf("%q: got %q", line, h.out.String())
			}
		}
	})

	t.Run("given only the shebang and a hash inside a command, when running, then there is no comment", func(t *testing.T) {
		files := clean()
		files[".githooks/pre-commit"] = file("#!/usr/bin/env bash\necho \"a # b\"\n")
		h := newHarness(files)
		if err := h.run(t, "quick"); err != nil {
			t.Fatal(err)
		}
	})
}

func TestAgentFiles(t *testing.T) {
	tests := []struct {
		name, content, failure string
	}{
		{
			"given an agent whose name differs from the file, when running, then it fails",
			strings.Replace(agent("alpha"), "name: alpha", "name: beta", 1),
			"agents/alpha.md: name diferente do nome do arquivo",
		},
		{
			"given an agent with an unknown model, when running, then it fails",
			strings.Replace(agent("alpha"), "model: sonnet", "model: gpt", 1),
			"agents/alpha.md: model inválido",
		},
		{
			"given an agent with an unknown effort, when running, then it fails",
			strings.Replace(agent("alpha"), "effort: high", "effort: huge", 1),
			"agents/alpha.md: effort inválido",
		},
		{
			"given an agent with an empty description, when running, then it fails",
			strings.Replace(agent("alpha"), "description: A role", "description: ", 1),
			"agents/alpha.md: sem description",
		},
		{
			"given an agent without the Character line, when running, then it fails",
			strings.Replace(agent("alpha"), "Character: Someone\n", "", 1),
			"agents/alpha.md: sem a linha Character",
		},
		{
			"given an agent whose model is only in the body, when running, then the header still lacks it",
			strings.Replace(agent("alpha"), "model: sonnet\n", "", 1) + "model: opus\n",
			"agents/alpha.md: model inválido",
		},
		{
			"given an agent with two name lines, when running, then it fails",
			strings.Replace(agent("alpha"), "name: alpha\n", "name: alpha\nname: alpha\n", 1),
			"agents/alpha.md: name diferente do nome do arquivo",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			files := clean()
			files["agents/alpha.md"] = file(tc.content)
			h := newHarness(files)
			if err := h.run(t, "quick"); !errors.Is(err, ErrFailed) {
				t.Fatalf("got %v", err)
			}
			if want := "FALHOU: " + tc.failure + "\n"; !strings.Contains(h.out.String(), want) {
				t.Fatalf("missing %q in %q", want, h.out.String())
			}
		})
	}

	t.Run("given every accepted model and effort, when running, then nothing fails", func(t *testing.T) {
		for _, model := range []string{"opus", "sonnet", "haiku"} {
			for _, effort := range []string{"low", "medium", "high", "xhigh", "max"} {
				content := strings.Replace(strings.Replace(agent("alpha"), "sonnet", model, 1), "high", effort, 1)
				files := clean()
				files["agents/alpha.md"] = file(content)
				if err := newHarness(files).run(t, "quick"); err != nil {
					t.Fatalf("%s %s: %v", model, effort, err)
				}
			}
		}
	})

	t.Run("given an agent file with no header, when running, then every header check fails but the Character line passes", func(t *testing.T) {
		files := clean()
		files["agents/alpha.md"] = file("Character: Someone\n")
		h := newHarness(files)
		if err := h.run(t, "quick"); !errors.Is(err, ErrFailed) {
			t.Fatalf("got %v", err)
		}
		if got := strings.Count(h.out.String(), "FALHOU: "); got != 4 {
			t.Fatalf("got %d failures in %q", got, h.out.String())
		}
	})
}

func TestSkillFiles(t *testing.T) {
	t.Run("given a skill whose name differs from the folder, when running, then it fails", func(t *testing.T) {
		files := clean()
		files["skills/pilot/SKILL.md"] = file("---\nname: pilot-two\n---\n")
		h := newHarness(files)
		if err := h.run(t, "quick"); !errors.Is(err, ErrFailed) {
			t.Fatalf("got %v", err)
		}
		if !strings.Contains(h.out.String(), "FALHOU: skills/pilot/SKILL.md: name diferente da pasta\n") {
			t.Fatalf("got %q", h.out.String())
		}
	})

	t.Run("given a skill whose name line is anywhere in the file, when running, then it passes", func(t *testing.T) {
		files := clean()
		files["skills/pilot/SKILL.md"] = file("title\nname: pilot\n")
		if err := newHarness(files).run(t, "quick"); err != nil {
			t.Fatal(err)
		}
	})
}

func TestDependencyManifest(t *testing.T) {
	tests := []struct {
		name, content, printed string
	}{
		{
			"given a dependency without a function, when running, then it prints what is missing and fails",
			"[[dependency]]\nname = \"a\"\nkind = \"npm\"\nlicense = \"MIT\"\n",
			"dependência sem function a\n",
		},
		{
			"given a dependency without a name, when running, then the missing name is shown as None",
			"[[dependency]]\nkind = \"npm\"\nlicense = \"MIT\"\nfunction = \"f\"\n",
			"dependência sem name None\n",
		},
		{
			"given two broken dependencies, when running, then only the first is reported",
			"[[dependency]]\nname = \"a\"\nkind = \"npm\"\nlicense = \"MIT\"\n[[dependency]]\nname = \"b\"\n",
			"dependência sem function a\n",
		},
		{
			"given a file that is not TOML, when running, then it fails",
			"this is = = not toml",
			"",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			files := clean()
			files["deps.toml"] = file(tc.content)
			h := newHarness(files)
			if err := h.run(t, "quick"); !errors.Is(err, ErrFailed) {
				t.Fatalf("got %v", err)
			}
			out := h.out.String()
			if !strings.Contains(out, tc.printed) || !strings.Contains(out, "FALHOU: deps.toml inválido\n") {
				t.Fatalf("got %q", out)
			}
			if strings.Count(out, "dependência sem") > 1 {
				t.Fatalf("reported more than the first: %q", out)
			}
		})
	}

	t.Run("given no deps.toml, when running, then it fails", func(t *testing.T) {
		files := clean()
		delete(files, "deps.toml")
		h := newHarness(files)
		if err := h.run(t, "quick"); !errors.Is(err, ErrFailed) {
			t.Fatalf("got %v", err)
		}
		if !strings.Contains(h.out.String(), "FALHOU: deps.toml inválido\n") {
			t.Fatalf("got %q", h.out.String())
		}
	})

	t.Run("given a manifest with no dependency table, when running, then it passes", func(t *testing.T) {
		files := clean()
		files["deps.toml"] = file("format_version = 1\n")
		if err := newHarness(files).run(t, "quick"); err != nil {
			t.Fatal(err)
		}
	})
}

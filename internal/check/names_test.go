package check

import (
	"errors"
	"strings"
	"testing"
)

const nameFailure = "FALHOU: nome de personagem fora de agents/ e skills/\n"

func TestCharacterNames(t *testing.T) {
	t.Run("given a name as a whole word in a doc, when running, then the line is printed with its path and number and the check fails", func(t *testing.T) {
		files := clean()
		files["docs/page.md"] = file("one\ntwo\nask Foo today\n")
		h := newHarness(files)
		if err := h.run(t, "quick"); !errors.Is(err, ErrFailed) {
			t.Fatalf("got %v", err)
		}
		want := "./docs/page.md:3:ask Foo today\n" + nameFailure
		if !strings.Contains(h.out.String(), want) {
			t.Fatalf("missing %q in %q", want, h.out.String())
		}
	})

	t.Run("given a name with a space in the list, when it appears in a file, then it is found", func(t *testing.T) {
		files := clean()
		files["docs/page.md"] = file("meet Bar Baz.\n")
		h := newHarness(files)
		if err := h.run(t, "quick"); !errors.Is(err, ErrFailed) {
			t.Fatalf("got %v", err)
		}
	})

	t.Run("given a name inside a longer word, when running, then it is not a match", func(t *testing.T) {
		files := clean()
		files["docs/page.md"] = file("Foolish and Food and xFoo and Foo9 and Foo_ and Fooé\n")
		if err := newHarness(files).run(t, "quick"); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("given a name next to punctuation, when running, then it is a match", func(t *testing.T) {
		for _, line := range []string{"(Foo)", "Foo, then", "a-Foo-b", "Foo", "\"Foo\""} {
			files := clean()
			files["docs/page.md"] = file(line + "\n")
			if err := newHarness(files).run(t, "quick"); !errors.Is(err, ErrFailed) {
				t.Fatalf("%q: got %v", line, err)
			}
		}
	})

	t.Run("given a lower case spelling, when running, then it is not a match", func(t *testing.T) {
		files := clean()
		files["docs/page.md"] = file("foo\n")
		if err := newHarness(files).run(t, "quick"); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("given the name in exempt places, when running, then nothing fails", func(t *testing.T) {
		exempt := []string{
			"agents/alpha.md", "skills/pilot/SKILL.md", "docs/agents/page.md", "x/skills/page.md",
			"README.md", "docs/README.md", "scripts/characters.txt", "deep/dir/characters.txt",
			".git/config", ".serena/memory.md", "docs/.git/x",
		}
		for _, path := range exempt {
			files := clean()
			existing := ""
			if current, ok := files[path]; ok {
				existing = string(current.Data)
			}
			files[path] = file(existing + "Foo\nCharacter: Foo\n")
			if err := newHarness(files).run(t, "quick"); err != nil {
				t.Fatalf("%s: %v", path, err)
			}
		}
	})

	t.Run("given a line with the --agent flag, when running, then it is ignored", func(t *testing.T) {
		files := clean()
		files["docs/page.md"] = file("run it with --agent Foo\n")
		if err := newHarness(files).run(t, "quick"); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("given a binary file with the name, when running, then it is skipped", func(t *testing.T) {
		files := clean()
		files["assets/image.png"] = file("\x89PNG\x00\x00 Foo \x00")
		if err := newHarness(files).run(t, "quick"); err != nil {
			t.Fatal(err)
		}
	})
}

func TestRuleOrigins(t *testing.T) {
	failure := "FALHOU: origem de regra (Origin: session) em rules/ ou agents/: sessão e frase ficam só nas notas privadas\n"

	t.Run("given a rule with an origin session, when running, then the line is printed and the check fails", func(t *testing.T) {
		files := clean()
		files["rules/one.md"] = file("text\nOrigin: session 123\n")
		h := newHarness(files)
		if err := h.run(t, "quick"); !errors.Is(err, ErrFailed) {
			t.Fatalf("got %v", err)
		}
		if want := "rules/one.md:2:Origin: session 123\n" + failure; !strings.Contains(h.out.String(), want) {
			t.Fatalf("missing %q in %q", want, h.out.String())
		}
	})

	t.Run("given spacing variations and the agents folder, when running, then they are found", func(t *testing.T) {
		for path, line := range map[string]string{
			"rules/a.md":        "Origin:session",
			"rules/nested/b.md": "Origin:   \t session",
			"agents/beta.md":    "Origin: session",
		} {
			files := clean()
			files[path] = file(line + "\n")
			if err := newHarness(files).run(t, "quick"); !errors.Is(err, ErrFailed) {
				t.Fatalf("%s: got %v", path, err)
			}
		}
	})

	t.Run("given the same text outside rules and agents, when running, then it is not a match", func(t *testing.T) {
		files := clean()
		files["docs/page.md"] = file("Origin: session\n")
		if err := newHarness(files).run(t, "quick"); err != nil {
			t.Fatal(err)
		}
	})
}

func TestPrivacy(t *testing.T) {
	failure := "FALHOU: nome de empresa no repositório\n"

	t.Run("given no privacy file, when running, then it warns with the path and does not fail", func(t *testing.T) {
		h := newHarness(clean())
		h.env["PRIVACY_FILE"] = "/somewhere/else.txt"
		if err := h.run(t, "quick"); err != nil {
			t.Fatal(err)
		}
		want := "aviso: sem /somewhere/else.txt, a checagem de nomes de empresa foi pulada\n"
		if !strings.Contains(h.out.String(), want) {
			t.Fatalf("got %q", h.out.String())
		}
	})

	t.Run("given a pattern that matches regardless of case, when running, then every file is searched and the check fails", func(t *testing.T) {
		files := clean()
		files["README.md"] = file("one\nWe work with ACME corp\n")
		files["agents/alpha.md"] = file(agent("alpha") + "acme\n")
		h := newHarness(files)
		h.host[privacyDefault] = "acme\n"
		if err := h.run(t, "quick"); !errors.Is(err, ErrFailed) {
			t.Fatalf("got %v", err)
		}
		out := h.out.String()
		for _, want := range []string{"./README.md:2:We work with ACME corp\n", "./agents/alpha.md:", failure} {
			if !strings.Contains(out, want) {
				t.Fatalf("missing %q in %q", want, out)
			}
		}
		if strings.Contains(out, "aviso") {
			t.Fatalf("warned although the file exists: %q", out)
		}
	})

	t.Run("given the PRIVACY_FILE variable, when running, then it wins over the default path", func(t *testing.T) {
		files := clean()
		files["docs/page.md"] = file("secret\n")
		h := newHarness(files)
		h.env["PRIVACY_FILE"] = "/custom.txt"
		h.host["/custom.txt"] = "sec+ret\n"
		h.host[privacyDefault] = "nothing-here\n"
		if err := h.run(t, "quick"); !errors.Is(err, ErrFailed) {
			t.Fatalf("got %v", err)
		}
	})

	t.Run("given several patterns, when one of them matches, then the check fails", func(t *testing.T) {
		files := clean()
		files["docs/page.md"] = file("the second one\n")
		h := newHarness(files)
		h.host[privacyDefault] = "first\nsecond\n"
		if err := h.run(t, "quick"); !errors.Is(err, ErrFailed) {
			t.Fatalf("got %v", err)
		}
	})

	t.Run("given blank lines in the pattern file, when nothing else matches, then the check passes", func(t *testing.T) {
		h := newHarness(clean())
		h.host[privacyDefault] = "\nnothing-here\n\n"
		if err := h.run(t, "quick"); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("given matches only in the git and serena folders, when running, then they are skipped", func(t *testing.T) {
		files := clean()
		files[".git/config"] = file("acme\n")
		files[".serena/memory.md"] = file("acme\n")
		h := newHarness(files)
		h.host[privacyDefault] = "acme\n"
		if err := h.run(t, "quick"); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("given an invalid pattern, when running, then the check fails and says which file", func(t *testing.T) {
		h := newHarness(clean())
		h.host[privacyDefault] = "(unclosed\n"
		if err := h.run(t, "quick"); !errors.Is(err, ErrFailed) {
			t.Fatalf("got %v", err)
		}
		if !strings.Contains(h.out.String(), "padrão inválido em "+privacyDefault) {
			t.Fatalf("got %q", h.out.String())
		}
	})
}

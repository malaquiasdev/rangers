package install

import (
	"os"
	"strings"
	"testing"
)

const sampleManifest = `format_version = 1

[[dependency]]
name = "pstack"
kind = "claude-plugin"
marketplace = "cursor/plugins"
plugin = "pstack@cursor-plugins"
version = "0.15.46"

[[dependency]]
name = "openspec"
optional = true
kind = "npm"
package = "@fission-ai/openspec"
version = "1.7.0"

[[dependency]]
name = "container-use"
kind = "brew"
formula = "dagger/tap/container-use"
version = "0.4.2"

[[dependency]]
name = "lychee"
kind = "container-image"
image = "docker.io/lycheeverse/lychee:0.24.2"

[[dependency]]
name = "grilling"
kind = "skill"
repository = "mattpocock/skills"
path = "skills/productivity/grilling"
commit = "abeeaf95660ba9e8b7c753eb38f441ca707a8823"
`

func TestParseManifest(t *testing.T) {
	t.Run("given a manifest, when parsing, then every dependency keeps its fields in order", func(t *testing.T) {
		deps, err := parseManifest([]byte(sampleManifest))
		if err != nil {
			t.Fatal(err)
		}
		if len(deps) != 5 {
			t.Fatalf("got %d dependencies, want 5", len(deps))
		}
		plugin := deps[0]
		if plugin.Name != "pstack" || plugin.Kind != "claude-plugin" ||
			plugin.Marketplace != "cursor/plugins" || plugin.Plugin != "pstack@cursor-plugins" ||
			plugin.Version != "0.15.46" || plugin.Optional {
			t.Errorf("unexpected plugin %+v", plugin)
		}
		if !deps[1].Optional || deps[1].Package != "@fission-ai/openspec" {
			t.Errorf("unexpected npm %+v", deps[1])
		}
		if deps[2].Formula != "dagger/tap/container-use" || deps[3].Image == "" {
			t.Errorf("unexpected brew or image %+v %+v", deps[2], deps[3])
		}
		skill := deps[4]
		if skill.Repository != "mattpocock/skills" || skill.Path != "skills/productivity/grilling" ||
			skill.Commit != "abeeaf95660ba9e8b7c753eb38f441ca707a8823" {
			t.Errorf("unexpected skill %+v", skill)
		}
	})

	t.Run("given a format_version other than 1, when parsing, then it refuses", func(t *testing.T) {
		_, err := parseManifest([]byte("format_version = 2\n"))
		if err == nil || !strings.Contains(err.Error(), "format_version 2") {
			t.Fatalf("got %v", err)
		}
	})

	t.Run("given text that is not toml, when parsing, then it returns the error", func(t *testing.T) {
		if _, err := parseManifest([]byte("format_version = ")); err == nil {
			t.Fatal("expected an error")
		}
	})

	t.Run("given an unknown kind, when parsing, then the error names the dependency and the kind", func(t *testing.T) {
		raw := "format_version = 1\n[[dependency]]\nname = \"x\"\nkind = \"pip\"\n"
		_, err := parseManifest([]byte(raw))
		if err == nil || !strings.Contains(err.Error(), `x: unknown kind "pip"`) {
			t.Fatalf("got %v", err)
		}
	})

	t.Run("given a dependency without a required field, when parsing, then the error names the field", func(t *testing.T) {
		raw := "format_version = 1\n[[dependency]]\nname = \"x\"\nkind = \"brew\"\n"
		_, err := parseManifest([]byte(raw))
		if err == nil || !strings.Contains(err.Error(), "x: missing formula") {
			t.Fatalf("got %v", err)
		}
	})

	t.Run("given a dependency without a name, when parsing, then it refuses", func(t *testing.T) {
		raw := "format_version = 1\n[[dependency]]\nkind = \"brew\"\nformula = \"a/b/c\"\n"
		if _, err := parseManifest([]byte(raw)); err == nil {
			t.Fatal("expected an error")
		}
	})

	t.Run("given the deps.toml of the repository, when parsing, then it is valid", func(t *testing.T) {
		raw, err := os.ReadFile("../../deps.toml")
		if err != nil {
			t.Fatal(err)
		}
		if _, err = parseManifest(raw); err != nil {
			t.Fatal(err)
		}
	})
}

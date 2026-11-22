package install

import (
	"errors"
	"fmt"
	"strings"

	"github.com/BurntSushi/toml"
)

const (
	formatVersion = 1
	kindPlugin    = "claude-plugin"
	kindNPM       = "npm"
	kindBrew      = "brew"
	kindImage     = "container-image"
	kindSkill     = "skill"
)

type Dependency struct {
	Name        string `toml:"name"`
	Kind        string `toml:"kind"`
	Optional    bool   `toml:"optional"`
	Marketplace string `toml:"marketplace"`
	Plugin      string `toml:"plugin"`
	Version     string `toml:"version"`
	Package     string `toml:"package"`
	Formula     string `toml:"formula"`
	Image       string `toml:"image"`
	Repository  string `toml:"repository"`
	Path        string `toml:"path"`
	Commit      string `toml:"commit"`
}

type manifest struct {
	FormatVersion int          `toml:"format_version"`
	Dependencies  []Dependency `toml:"dependency"`
}

type field struct{ name, value string }

func parseManifest(raw []byte) ([]Dependency, error) {
	var m manifest
	if err := toml.Unmarshal(raw, &m); err != nil {
		return nil, err
	}
	if m.FormatVersion != formatVersion {
		return nil, fmt.Errorf("unsupported format_version %d", m.FormatVersion)
	}
	for _, d := range m.Dependencies {
		if err := d.validate(); err != nil {
			return nil, err
		}
	}
	return m.Dependencies, nil
}

func (d Dependency) validate() error {
	if d.Name == "" {
		return errors.New("a dependency has no name")
	}
	required := d.required()
	if required == nil {
		return fmt.Errorf("%s: unknown kind %q", d.Name, d.Kind)
	}
	var missing []string
	for _, f := range required {
		if f.value == "" {
			missing = append(missing, f.name)
		}
	}
	if len(missing) > 0 {
		return fmt.Errorf("%s: missing %s", d.Name, strings.Join(missing, ", "))
	}
	return nil
}

func (d Dependency) required() []field {
	switch d.Kind {
	case kindPlugin:
		return []field{{"marketplace", d.Marketplace}, {"plugin", d.Plugin}}
	case kindNPM:
		return []field{{"package", d.Package}, {"version", d.Version}}
	case kindBrew:
		return []field{{"formula", d.Formula}}
	case kindImage:
		return []field{{"image", d.Image}}
	case kindSkill:
		return []field{{"repository", d.Repository}, {"path", d.Path}, {"commit", d.Commit}}
	}
	return nil
}

package check

import (
	"fmt"
	"io/fs"
	"path"
	"regexp"
	"slices"
	"strings"

	"github.com/BurntSushi/toml"
)

const (
	namePrefix        = "name: "
	descriptionPrefix = "description: "
	characterPrefix   = "Character: "
	frontmatterFence  = "---"
	missingName       = "None"
)

var (
	modelLine  = regexp.MustCompile(`^model: (opus|sonnet|haiku)$`)
	effortLine = regexp.MustCompile(`^effort: (low|medium|high|xhigh|max)$`)

	dependencyFields = []string{"name", "kind", "license", "function"}
)

func (c *checker) agents() {
	paths, _ := fs.Glob(c.dep.Root, "agents/*.md")
	for _, file := range paths {
		c.agent(file)
	}
}

func (c *checker) agent(file string) {
	raw, err := fs.ReadFile(c.dep.Root, file)
	if err != nil {
		return
	}
	lines := strings.Split(string(raw), "\n")
	head := header(lines)
	if strings.Join(valuesOf(head, namePrefix), "\n") != strings.TrimSuffix(path.Base(file), ".md") {
		c.fail("%s: name diferente do nome do arquivo", file)
	}
	if !slices.ContainsFunc(head, modelLine.MatchString) {
		c.fail("%s: model inválido", file)
	}
	if !slices.ContainsFunc(head, effortLine.MatchString) {
		c.fail("%s: effort inválido", file)
	}
	if !slices.ContainsFunc(head, hasDescription) {
		c.fail("%s: sem description", file)
	}
	if len(valuesOf(lines, characterPrefix)) == 0 {
		c.fail("%s: sem a linha Character", file)
	}
}

func (c *checker) skills() {
	paths, _ := fs.Glob(c.dep.Root, "skills/*/SKILL.md")
	for _, file := range paths {
		raw, err := fs.ReadFile(c.dep.Root, file)
		skill := path.Base(path.Dir(file))
		if err == nil && !slices.Contains(strings.Split(string(raw), "\n"), namePrefix+skill) {
			c.fail("%s: name diferente da pasta", file)
		}
	}
}

func (c *checker) manifest() {
	raw, err := fs.ReadFile(c.dep.Root, "deps.toml")
	if err != nil {
		c.fail("deps.toml inválido")
		return
	}
	var data map[string]any
	if err = toml.Unmarshal(raw, &data); err != nil {
		c.say("deps.toml: %v\n", err)
		c.fail("deps.toml inválido")
		return
	}
	items, _ := data["dependency"].([]map[string]any)
	if problem := firstMissingField(items); problem != "" {
		c.say("%s\n", problem)
		c.fail("deps.toml inválido")
	}
}

func firstMissingField(items []map[string]any) string {
	for _, item := range items {
		for _, field := range dependencyFields {
			if _, ok := item[field]; ok {
				continue
			}
			name := missingName
			if value, ok := item["name"]; ok {
				name = fmt.Sprint(value)
			}
			return fmt.Sprintf("dependência sem %s %s", field, name)
		}
	}
	return ""
}

func header(lines []string) []string {
	if lines[0] != frontmatterFence {
		return nil
	}
	body := lines[1:]
	if end := slices.Index(body, frontmatterFence); end >= 0 {
		return body[:end]
	}
	return body
}

func valuesOf(lines []string, prefix string) []string {
	var values []string
	for _, line := range lines {
		if value, ok := strings.CutPrefix(line, prefix); ok {
			values = append(values, value)
		}
	}
	return values
}

func hasDescription(line string) bool {
	return strings.HasPrefix(line, descriptionPrefix) && len(line) > len(descriptionPrefix)
}

package check

import (
	"io/fs"
	"path/filepath"
	"regexp"
	"strings"
)

const (
	agentFlag    = "--agent "
	wordBoundary = `[^\p{L}\p{N}_]`
	privacyUnder = ".config/rangers/privacy.txt"
)

var originLine = regexp.MustCompile(`Origin:[[:space:]]*session`)

func WholeWord(pattern string) (*regexp.Regexp, error) {
	alternatives := strings.ReplaceAll(strings.TrimRight(pattern, "\n"), "\n", "|")
	return regexp.Compile("(?:^|" + wordBoundary + ")(?:" + alternatives + ")" +
		"(?:" + wordBoundary + "|$)")
}

func (c *checker) characters() error {
	raw, err := fs.ReadFile(c.dep.Root, namesFile)
	if err != nil {
		return err
	}
	names, err := WholeWord(string(raw))
	if err != nil {
		return err
	}
	found := c.search(scan{
		roots:     []string{"."},
		skipDirs:  set(gitDir, serenaDir, "agents", "skills"),
		skipFiles: set("characters.txt", readmeFile),
		matches: func(line string) bool {
			return names.MatchString(line) && !strings.Contains(line, agentFlag)
		},
	})
	c.report(found, "nome de personagem fora de agents/ e skills/")
	return nil
}

func (c *checker) origins() {
	found := c.search(scan{roots: []string{"rules", "agents"}, matches: originLine.MatchString})
	c.report(found, "origem de regra (Origin: session) em rules/ ou agents/: "+
		"sessão e frase ficam só nas notas privadas")
}

func (c *checker) privacyPath() string {
	if path := c.env("PRIVACY_FILE"); path != "" {
		return path
	}
	return filepath.Join(c.env("HOME"), privacyUnder)
}

func (c *checker) privacy() {
	path := c.privacyPath()
	raw, err := c.dep.ReadHost(path)
	if err != nil {
		c.say("aviso: sem %s, a checagem de nomes de empresa foi pulada\n", path)
		return
	}
	patterns := privacyPatterns(string(raw))
	if len(patterns) == 0 {
		return
	}
	names, err := regexp.Compile("(?i)" + strings.Join(patterns, "|"))
	if err != nil {
		c.fail("padrão inválido em %s: %v", path, err)
		return
	}
	found := c.search(scan{roots: []string{"."}, skipDirs: vcsDirs(), matches: names.MatchString})
	c.report(found, "nome de empresa no repositório")
}

func privacyPatterns(text string) []string {
	var patterns []string
	for line := range strings.Lines(text) {
		if line = strings.TrimRight(line, "\r\n"); line != "" {
			patterns = append(patterns, "(?:"+line+")")
		}
	}
	return patterns
}

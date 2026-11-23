package check

import (
	"context"
	"io/fs"
	"regexp"
	"strings"
)

var shellComment = regexp.MustCompile(`^[[:space:]]*#`)

func (c *checker) shell(ctx context.Context) {
	for _, pattern := range []string{"scripts/*.sh", ".githooks/*", "bin/*"} {
		paths, _ := fs.Glob(c.dep.Root, pattern)
		for _, path := range paths {
			c.shellFile(ctx, path)
		}
	}
}

func (c *checker) shellFile(ctx context.Context, path string) {
	info, err := fs.Stat(c.dep.Root, path)
	if err != nil || !info.Mode().IsRegular() {
		return
	}
	if c.dep.Runner.Run(ctx, Invocation{Name: "bash", Args: []string{"-n", path}}) != nil {
		c.fail("sintaxe de %s", path)
	}
	raw, err := fs.ReadFile(c.dep.Root, path)
	if err == nil && hasComment(string(raw)) {
		c.fail("comentário em %s", path)
	}
}

func hasComment(script string) bool {
	_, rest, _ := strings.Cut(script, "\n")
	for line := range strings.Lines(rest) {
		if shellComment.MatchString(line) {
			return true
		}
	}
	return false
}

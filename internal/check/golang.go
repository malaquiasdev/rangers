package check

import (
	"context"
	"io/fs"
	"regexp"
	"strings"
)

const (
	lycheeImage = "docker.io/lycheeverse/lychee:0.24.2"
	lintTool    = "golangci-lint"
)

var goComment = regexp.MustCompile(`(^|[[:space:]])//|^[[:space:]]*/\*`)

func (c *checker) golang(ctx context.Context) {
	if _, err := fs.Stat(c.dep.Root, "go.mod"); err != nil {
		return
	}
	found := c.search(scan{
		roots:    []string{"."},
		skipDirs: vcsDirs(),
		suffix:   ".go",
		matches:  goComment.MatchString,
	})
	c.report(found, "comentário em arquivo Go")
	listed, _ := c.dep.Runner.Output(ctx, "gofmt", "-l", ".")
	if files := strings.Split(strings.TrimRight(listed, "\n"), "\n"); files[0] != "" {
		c.fail("gofmt: %s ", strings.Join(files, " "))
	}
	if c.full() {
		c.tool(ctx, "go vet", Invocation{Name: "go", Args: []string{"vet", "./..."}})
		c.tool(ctx, "go test", Invocation{Name: "go", Args: []string{"test", "./..."}})
		c.tool(ctx, lintTool, Invocation{Name: lintTool, Args: []string{"run"}})
	}
}

func (c *checker) slow(ctx context.Context) {
	if !c.full() {
		return
	}
	c.tool(ctx, "testes dos assets", Invocation{
		Name:  "python3",
		Args:  []string{"-I", "-m", "unittest", "discover", "-s", "assets/tests", "-t", "assets"},
		Env:   []string{"PYTHONDONTWRITEBYTECODE=1"},
		Quiet: true,
	})
	c.tool(ctx, "links da documentação", Invocation{
		Name: "podman",
		Args: []string{
			"run", "--rm", "-v", c.dep.Dir + ":/w:ro", "-w", "/w", lycheeImage,
			"--offline", "--include-fragments", readmeFile, "docs/**/*.md",
		},
		Quiet: true,
	})
}

func (c *checker) tool(ctx context.Context, label string, call Invocation) {
	if c.dep.Runner.Run(ctx, call) != nil {
		c.fail("%s", label)
	}
}

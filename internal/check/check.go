package check

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
)

const (
	modeFull  = "full"
	namesFile = "scripts/characters.txt"
)

var ErrFailed = errors.New("checks failed")

type Invocation struct {
	Name  string
	Args  []string
	Env   []string
	Quiet bool
}

type Runner interface {
	Run(ctx context.Context, call Invocation) error
	Output(ctx context.Context, name string, args ...string) (string, error)
}

type Dependencies struct {
	Dir      string
	Root     fs.FS
	Runner   Runner
	ReadHost func(path string) ([]byte, error)
	Out      io.Writer
}

type checker struct {
	dep      *Dependencies
	env      func(string) string
	mode     string
	failures int
}

func Command(ctx context.Context, args []string, env func(string) string, dep *Dependencies) error {
	c := &checker{dep: dep, env: env, mode: modeFull}
	if len(args) > 0 {
		c.mode = args[0]
	}
	if err := c.run(ctx); err != nil {
		return err
	}
	return c.verdict()
}

func (c *checker) run(ctx context.Context) error {
	c.shell(ctx)
	c.agents()
	c.skills()
	c.manifest()
	if err := c.characters(); err != nil {
		return err
	}
	c.origins()
	c.privacy()
	c.golang(ctx)
	c.slow(ctx)
	return nil
}

func (c *checker) verdict() error {
	if c.failures == 0 {
		c.say("tudo certo (%s)\n", c.mode)
		return nil
	}
	c.say("\n%d checagem(ns) falharam\n", c.failures)
	return ErrFailed
}

func (c *checker) say(format string, args ...any) {
	_, _ = fmt.Fprintf(c.dep.Out, format, args...)
}

func (c *checker) fail(format string, args ...any) {
	c.say("FALHOU: "+format+"\n", args...)
	c.failures++
}

func (c *checker) full() bool {
	return c.mode == modeFull
}

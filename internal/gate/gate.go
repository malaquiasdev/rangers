package gate

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"path/filepath"
	"regexp"
)

const (
	usage = "usage: rangers gate [--merge] [--transcripts DIR] --repo owner/name --pr N" +
		" | rangers gate ask --repo owner/name --pr N"
	outputLimit = 560
)

var (
	ErrRefused = errors.New("refused")

	repoPattern = regexp.MustCompile(`^[\w.-]+/[\w.-]+$`)
)

type Runner interface {
	Has(name string) bool
	Output(ctx context.Context, name string, args ...string) (string, error)
}

type Dependencies struct {
	Runner Runner
	Out    io.Writer
}

type options struct {
	repo        string
	pr          int
	merge       bool
	ask         bool
	transcripts string
}

type outcome struct {
	line    string
	refusal string
}

type step func(ctx context.Context) (outcome, error)

type gate struct {
	dep  *Dependencies
	opts options
	pr   pullRequest
}

func Command(ctx context.Context, args []string, env func(string) string, dep *Dependencies) error {
	opts, err := parseOptions(args, env)
	if err != nil {
		return err
	}
	g := &gate{dep: dep, opts: opts}
	return g.execute(ctx)
}

func parseOptions(args []string, env func(string) string) (options, error) {
	opts := options{transcripts: filepath.Join(env("HOME"), ".claude", "projects")}
	rest := args
	if len(args) > 0 && args[0] == "ask" {
		opts.ask, rest = true, args[1:]
	}
	set := flag.NewFlagSet("gate", flag.ContinueOnError)
	set.StringVar(&opts.repo, "repo", "", "repository as owner/name")
	set.IntVar(&opts.pr, "pr", 0, "pull request number")
	set.BoolVar(&opts.merge, "merge", false, "merge when every check passes")
	set.StringVar(&opts.transcripts, "transcripts", opts.transcripts, "Claude Code projects folder")
	if err := set.Parse(rest); err != nil {
		return options{}, err
	}
	if set.NArg() > 0 || !opts.valid() {
		return options{}, errors.New(usage)
	}
	return opts, nil
}

func (o options) valid() bool {
	return o.pr > 0 && repoPattern.MatchString(o.repo) && (!o.ask || !o.merge)
}

func pass(name, detail string) outcome {
	return outcome{line: name + ": ok " + detail}
}

func refuse(name, reason string) outcome {
	return outcome{line: name + ": refused", refusal: reason}
}

func (g *gate) execute(ctx context.Context) error {
	var lines []string
	for _, run := range g.steps() {
		result, err := run(ctx)
		if err != nil {
			return err
		}
		lines = append(lines, result.line)
		if result.refusal != "" {
			return g.verdict(lines, "refused: "+result.refusal, ErrRefused)
		}
	}
	return g.verdict(lines, g.allowed(), nil)
}

func (g *gate) steps() []step {
	return []step{g.open, g.production}
}

func (g *gate) allowed() string {
	if g.opts.merge {
		return "allowed: merged"
	}
	return "allowed: dry run, nothing merged"
}

func (g *gate) verdict(lines []string, last string, result error) error {
	text := fit(append(lines, last), outputLimit)
	if _, err := fmt.Fprintln(g.dep.Out, text); err != nil {
		return err
	}
	return result
}

package install

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"path/filepath"
)

const (
	usage        = "usage: rangers install [--check] [--optional] [--repo DIR] [--claude-dir DIR]"
	manifestFile = "deps.toml"
	runtimeName  = "docker"
)

var ErrIncomplete = errors.New("pieces missing")

type Dependencies struct {
	Runner Runner
	Files  Files
	Out    io.Writer
	Dir    string
}

type options struct {
	check     bool
	optional  bool
	repo      string
	claudeDir string
}

type installer struct {
	runner Runner
	files  Files
	opts   options
	out    printer
}

func Command(ctx context.Context, args []string, env func(string) string, dep *Dependencies) error {
	opts, err := parseOptions(args, env)
	if err != nil {
		return err
	}
	deps, err := resolve(dep, &opts)
	if err != nil {
		return err
	}
	in := &installer{runner: dep.Runner, files: dep.Files, opts: opts, out: newPrinter(dep.Out, deps)}
	if opts.check {
		return in.check(ctx, deps)
	}
	return in.install(ctx, deps)
}

func resolve(dep *Dependencies, opts *options) ([]Dependency, error) {
	if opts.repo == "" {
		root, err := findRoot(dep.Files, dep.Dir)
		if err != nil {
			return nil, err
		}
		opts.repo = root
	}
	raw, err := dep.Files.ReadFile(filepath.Join(opts.repo, manifestFile))
	if err != nil {
		return nil, err
	}
	deps, err := parseManifest(raw)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", manifestFile, err)
	}
	return deps, nil
}

func parseOptions(args []string, env func(string) string) (options, error) {
	opts := options{claudeDir: filepath.Join(env("HOME"), ".claude")}
	set := flag.NewFlagSet("install", flag.ContinueOnError)
	set.BoolVar(&opts.check, "check", false, "show the state of each piece and change nothing")
	set.BoolVar(&opts.optional, "optional", false, "include the optional pieces")
	set.StringVar(&opts.repo, "repo", "", "root of the rangers repository")
	set.StringVar(&opts.claudeDir, "claude-dir", opts.claudeDir, "Claude Code folder")
	if err := set.Parse(args); err != nil {
		return options{}, err
	}
	if set.NArg() > 0 {
		return options{}, errors.New(usage)
	}
	return opts, nil
}

func findRoot(files Files, dir string) (string, error) {
	for current := dir; ; current = filepath.Dir(current) {
		if files.Exists(filepath.Join(current, manifestFile)) {
			return current, nil
		}
		if filepath.Dir(current) == current {
			return "", fmt.Errorf("no %s from %s up: pass --repo", manifestFile, dir)
		}
	}
}

func (in *installer) check(ctx context.Context, deps []Dependency) error {
	missing := 0
	for _, d := range deps {
		status, ok := in.checkStatus(ctx, d)
		in.out.row(d.Name, status)
		if !ok {
			missing++
		}
	}
	if !in.checkRuntime(ctx) {
		missing++
	}
	return in.out.summary(missing)
}

func (in *installer) checkStatus(ctx context.Context, d Dependency) (string, bool) {
	if version := in.version(ctx, d); version != "" {
		return version, true
	}
	if d.Optional && !in.opts.optional {
		return "not installed (optional)", true
	}
	return "MISSING", false
}

func (in *installer) checkRuntime(ctx context.Context) bool {
	if _, err := in.runner.Output(ctx, runtimeName, "info"); err != nil {
		in.out.row(runtimeName, "NOT RESPONDING (container-use needs it)")
		return false
	}
	in.out.row(runtimeName, "responding")
	return true
}

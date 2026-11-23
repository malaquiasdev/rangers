package githook

import (
	"context"
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/malaquiasdev/rangers/internal/check"
)

const (
	usageText = "usage: rangers hooks install" +
		" | rangers hooks run pre-commit|commit-msg FILE|pre-push"
	runCommand       = "run"
	installCommand   = "install"
	hooksDir         = ".githooks"
	namesFile        = "scripts/characters.txt"
	formatMessage    = "a mensagem precisa seguir o Conventional Commits"
	nameMessage      = "nome de personagem na mensagem do commit"
	installedMessage = "ganchos ligados: pre-commit, commit-msg e pre-push"
)

var (
	ErrRefused = errors.New("commit message refused")
	ErrUsage   = errors.New(usageText)

	conventional = regexp.MustCompile(
		`^(feat|fix|docs|chore|refactor|test|build|ci|perf|style|revert)(\([a-z0-9-]+\))?!?: .+`)
)

type Files interface {
	ReadFile(path string) ([]byte, error)
	Glob(pattern string) ([]string, error)
	MakeExecutable(path string) error
}

type Git interface {
	SetConfig(ctx context.Context, key, value string) error
}

type Dependencies struct {
	Check *check.Dependencies
	Files Files
	Git   Git
	Out   io.Writer
	Dir   string
}

func Command(ctx context.Context, args []string, env func(string) string, dep *Dependencies) error {
	switch {
	case len(args) == 1 && args[0] == installCommand:
		return install(ctx, dep)
	case len(args) > 1 && args[0] == runCommand:
		return run(ctx, args[1:], env, dep)
	}
	return ErrUsage
}

func run(ctx context.Context, args []string, env func(string) string, dep *Dependencies) error {
	switch args[0] {
	case "pre-commit":
		return check.Command(ctx, []string{"quick"}, env, dep.Check)
	case "pre-push":
		return check.Command(ctx, []string{"full"}, env, dep.Check)
	case "commit-msg":
		if len(args) == 2 {
			return commitMessage(args[1], dep)
		}
	}
	return ErrUsage
}

func forbiddenNames(dep *Dependencies) (*regexp.Regexp, error) {
	raw, err := dep.Files.ReadFile(filepath.Join(dep.Dir, namesFile))
	if err != nil {
		return nil, err
	}
	return check.WholeWord(string(raw))
}

func commitMessage(path string, dep *Dependencies) error {
	names, err := forbiddenNames(dep)
	if err != nil {
		return err
	}
	message, err := dep.Files.ReadFile(path)
	if err != nil {
		return err
	}
	subject, _, _ := strings.Cut(string(message), "\n")
	switch {
	case !conventional.MatchString(subject):
		return refuse(dep, formatMessage)
	case names.Match(message):
		return refuse(dep, nameMessage)
	}
	return nil
}

func refuse(dep *Dependencies, message string) error {
	if _, err := fmt.Fprintln(dep.Out, message); err != nil {
		return err
	}
	return ErrRefused
}

func install(ctx context.Context, dep *Dependencies) error {
	hooks, err := dep.Files.Glob(filepath.Join(dep.Dir, hooksDir, "*"))
	if err != nil {
		return err
	}
	for _, hook := range hooks {
		if err = dep.Files.MakeExecutable(hook); err != nil {
			return err
		}
	}
	if err = dep.Git.SetConfig(ctx, "core.hooksPath", hooksDir); err != nil {
		return err
	}
	_, err = fmt.Fprintln(dep.Out, installedMessage)
	return err
}

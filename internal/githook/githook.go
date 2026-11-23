package githook

import (
	"context"
	"errors"
	"io"

	"github.com/malaquiasdev/rangers/internal/check"
)

const usageText = "usage: rangers hooks install | rangers hooks run pre-commit|commit-msg FILE|pre-push"

var (
	ErrRefused = errors.New("commit message refused")
	ErrUsage   = errors.New(usageText)

	errNotImplemented = errors.New("not implemented")
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

type OSFiles struct{}

type GitCLI struct {
	Dir string
}

func (OSFiles) ReadFile(string) ([]byte, error) {
	return nil, errNotImplemented
}

func (OSFiles) Glob(string) ([]string, error) {
	return nil, errNotImplemented
}

func (OSFiles) MakeExecutable(string) error {
	return errNotImplemented
}

func (GitCLI) SetConfig(context.Context, string, string) error {
	return errNotImplemented
}

func Command(_ context.Context, _ []string, _ func(string) string, _ *Dependencies) error {
	return errNotImplemented
}

package check

import (
	"context"
	"errors"
	"io"
	"io/fs"
	"regexp"
)

var (
	ErrFailed = errors.New("checks failed")

	errNotImplemented = errors.New("not implemented")
)

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

type ExecRunner struct {
	Dir      string
	Out, Err io.Writer
}

func (ExecRunner) Run(context.Context, Invocation) error {
	return errNotImplemented
}

func (ExecRunner) Output(context.Context, string, ...string) (string, error) {
	return "", errNotImplemented
}

func WholeWord(string) (*regexp.Regexp, error) {
	return nil, errNotImplemented
}

func Command(_ context.Context, _ []string, _ func(string) string, _ *Dependencies) error {
	return errNotImplemented
}

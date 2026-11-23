package hook

import (
	"context"
	"errors"
	"io"
)

const (
	ListName = "Tarefas Claude"
	Prefix   = "Claude espera: "
)

var ErrUsage = errors.New("usage: rangers hook stop-push-check|complete-reminders")

type Scripter interface {
	Available() bool
	Run(ctx context.Context, script string, args []string) error
}

type Dependencies struct {
	In       io.Reader
	Out      io.Writer
	Scripter Scripter
}

func Command(_ context.Context, _ []string, _ func(string) string, _ *Dependencies) error {
	return nil
}

type OsascriptCLI struct{}

func (OsascriptCLI) Available() bool {
	return false
}

func (OsascriptCLI) Run(_ context.Context, _ string, _ []string) error {
	return nil
}

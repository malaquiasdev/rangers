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

func Command(ctx context.Context, args []string, env func(string) string, dep *Dependencies) error {
	if len(args) == 0 {
		return ErrUsage
	}
	switch args[0] {
	case "stop-push-check":
		stopPushCheck(dep.In, dep.Out)
		return nil
	case "complete-reminders":
		completeReminders(ctx, env, dep.Scripter)
		return nil
	}
	return ErrUsage
}

package agentwatch

import (
	"context"
	"errors"
	"io"
	"time"
)

var ErrUsage = errors.New("usage: rangers watch [limit-seconds]")

type Git interface {
	LastCommit(ctx context.Context, dir string) (int64, bool)
}

type Dependencies struct {
	Git   Git
	Out   io.Writer
	Now   func() time.Time
	Sleep func(ctx context.Context, d time.Duration) error
}

type GitCLI struct{}

func (GitCLI) LastCommit(_ context.Context, _ string) (int64, bool) {
	return 0, false
}

func Command(_ context.Context, _ []string, _ func(string) string, _ *Dependencies) error {
	return nil
}

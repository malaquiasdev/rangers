package schedule

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
)

type ExecLaunchctl struct{}

func (e ExecLaunchctl) Bootout(ctx context.Context, target string) error {
	_, err := e.run(ctx, "bootout", target)
	return err
}

func (e ExecLaunchctl) Bootstrap(ctx context.Context, domain, plist string) error {
	_, err := e.run(ctx, "bootstrap", domain, plist)
	return err
}

func (e ExecLaunchctl) Print(ctx context.Context, target string) (string, error) {
	return e.run(ctx, "print", target)
}

func (ExecLaunchctl) run(ctx context.Context, args ...string) (string, error) {
	out, err := exec.CommandContext(ctx, "launchctl", args...).CombinedOutput()
	if err != nil {
		detail := strings.TrimSpace(string(out))
		return string(out), fmt.Errorf("launchctl %s: %w: %s", args[0], err, detail)
	}
	return string(out), nil
}

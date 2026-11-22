package learn

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

type Claude interface {
	Run(ctx context.Context, dir string, args []string) ([]byte, error)
}

type Git interface {
	NewWorktree(repo, branch, worktree string) error
}

type Reminder interface {
	Create(title, body string) error
}

type ClaudeCLI struct{}

func (ClaudeCLI) Run(ctx context.Context, dir string, args []string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, "claude", args...)
	cmd.Dir = dir
	cmd.Env = withoutGitVariables()
	out, err := cmd.Output()
	if err != nil {
		return nil, withStderr("claude", err)
	}
	return out, nil
}

type GitCLI struct{}

func (GitCLI) NewWorktree(repo, branch, worktree string) error {
	cmd := exec.Command("git", "-C", repo, "worktree", "add", "-q", "-b", branch, worktree, "main")
	cmd.Env = withoutGitVariables()
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git worktree add: %w: %s", err, strings.TrimSpace(string(out)))
	}
	return nil
}

type OsascriptReminder struct {
	Timeout time.Duration
}

func (r OsascriptReminder) Create(title, body string) error {
	ctx, cancel := context.WithTimeout(context.Background(), r.Timeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, "osascript", "-")
	cmd.WaitDelay = time.Second
	cmd.Stdin = strings.NewReader(reminderScript(title, body))
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("osascript: %w: %s", err, strings.TrimSpace(string(out)))
	}
	return nil
}

func reminderScript(title, body string) string {
	return fmt.Sprintf("tell application \"Reminders\"\n  tell list \"Tarefas Claude\"\n    make new reminder with properties {name:%s, body:%s}\n  end tell\nend tell\n", quote(title), quote(body))
}

func quote(text string) string {
	escaped := strings.NewReplacer(`\`, `\\`, `"`, `\"`).Replace(text)
	return `"` + escaped + `"`
}

func withoutGitVariables() []string {
	var kept []string
	for _, entry := range os.Environ() {
		if !strings.HasPrefix(entry, "GIT_") {
			kept = append(kept, entry)
		}
	}
	return kept
}

func withStderr(name string, err error) error {
	var exited *exec.ExitError
	if errors.As(err, &exited) {
		return fmt.Errorf("%s: %w: %s", name, err, strings.TrimSpace(string(exited.Stderr)))
	}
	return fmt.Errorf("%s: %w", name, err)
}

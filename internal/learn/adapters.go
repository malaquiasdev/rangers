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
	NewWorktree(ctx context.Context, repo, branch, worktree string) error
}

type Reminder interface {
	Create(ctx context.Context, title, body string) error
}

type Brain interface {
	Probe(ctx context.Context, dir string) error
}

type ClaudeCLI struct{}

type DirBrain struct {
	Timeout time.Duration
}

type GitCLI struct{}

type OsascriptReminder struct {
	Timeout time.Duration
}

func (ClaudeCLI) Run(ctx context.Context, dir string, args []string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, "claude", args...)
	cmd.WaitDelay = time.Second
	cmd.Dir = dir
	cmd.Env = withoutGitVariables()
	out, err := cmd.Output()
	if err != nil {
		return nil, withStderr("claude", err)
	}
	return out, nil
}

func (b DirBrain) Probe(ctx context.Context, dir string) error {
	limited, cancel := context.WithTimeout(ctx, b.Timeout)
	defer cancel()
	done := make(chan error, 1)
	go func() { done <- probeDir(dir) }()
	select {
	case err := <-done:
		return err
	case <-limited.Done():
		return fmt.Errorf("sem resposta de %s em %s", dir, b.Timeout)
	}
}

func probeDir(dir string) error {
	if _, err := os.ReadDir(dir); err != nil {
		return err
	}
	file, err := os.CreateTemp(dir, ".rangers-probe-*")
	if err != nil {
		return err
	}
	return errors.Join(file.Close(), os.Remove(file.Name()))
}

func (GitCLI) NewWorktree(ctx context.Context, repo, branch, worktree string) error {
	args := []string{"-C", repo, "worktree", "add", "-q", "-b", branch, worktree, "main"}
	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Env = withoutGitVariables()
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git worktree add: %w: %s", err, strings.TrimSpace(string(out)))
	}
	return nil
}

func (r OsascriptReminder) Create(ctx context.Context, title, body string) error {
	limited, cancel := context.WithTimeout(ctx, r.Timeout)
	defer cancel()
	cmd := exec.CommandContext(limited, "osascript", "-")
	cmd.WaitDelay = time.Second
	cmd.Stdin = strings.NewReader(reminderScript(title, body))
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("osascript: %w: %s", err, strings.TrimSpace(string(out)))
	}
	return nil
}

func reminderScript(title, body string) string {
	const template = "tell application \"Reminders\"\n  tell list \"Tarefas Claude\"\n" +
		"    make new reminder with properties {name:%s, body:%s}\n  end tell\nend tell\n"
	return fmt.Sprintf(template, quote(title), quote(body))
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

package hook

import (
	"context"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

const (
	timeoutVariable = "COMPLETE_REMINDERS_TIMEOUT"
	defaultTimeout  = 10 * time.Second

	closeScript = `on run argv
  set listName to item 1 of argv
  set waitingPrefix to item 2 of argv
  tell application "Reminders"
    if not (exists list listName) then return
    set waiting to (every reminder of list listName whose completed is false and ` +
		`name starts with waitingPrefix)
    repeat with entry in waiting
      set completed of entry to true
    end repeat
  end tell
end run
`
)

type OsascriptCLI struct{}

func (OsascriptCLI) Available() bool {
	_, err := exec.LookPath("osascript")
	return err == nil
}

func (OsascriptCLI) Run(ctx context.Context, script string, args []string) error {
	cmd := exec.CommandContext(ctx, "osascript", append([]string{"-"}, args...)...)
	cmd.Stdin = strings.NewReader(script)
	cmd.WaitDelay = time.Second
	return cmd.Run()
}

func completeReminders(ctx context.Context, env func(string) string, scripter Scripter) {
	if !scripter.Available() {
		return
	}
	limited, cancel := context.WithTimeout(ctx, timeout(env(timeoutVariable)))
	defer cancel()
	_ = scripter.Run(limited, closeScript, []string{ListName, Prefix})
}

func timeout(value string) time.Duration {
	seconds, err := strconv.ParseFloat(value, 64)
	if err != nil || seconds <= 0 {
		return defaultTimeout
	}
	return time.Duration(seconds * float64(time.Second))
}

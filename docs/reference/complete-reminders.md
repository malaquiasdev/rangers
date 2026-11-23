# The `complete-reminders` hook

`rangers hook complete-reminders` is a Claude Code `UserPromptSubmit` hook. When the human sends a message, it marks as completed every open reminder of the Reminders list `Tarefas Claude` whose name starts with `Claude espera: `. These are the reminders that the [`stop-push-check`](stop-push-check.md) hook requires at the end of a turn that waits for the human, so the human never completes them by hand. The source is in [`internal/hook/`](../../internal/hook/). The hook is not active until it is registered in `~/.claude/settings.json`.

It needs macOS, `osascript` and the Reminders app.

## Register it

`rangers install` links `~/.claude/bin/complete-reminders` to the `rangers` program, which runs this hook when it is called by that name. Add this to the `hooks` object of `~/.claude/settings.json`, next to the `Stop` entry of [`stop-push-check`](stop-push-check.md#register-it):

```json
{
  "hooks": {
    "UserPromptSubmit": [
      {
        "hooks": [
          { "type": "command", "command": "~/.claude/bin/complete-reminders", "timeout": 15 }
        ]
      }
    ]
  }
}
```

To skip the link, use `"command": "~/.local/bin/rangers hook complete-reminders"` instead.

## What it does

It runs one AppleScript through `osascript -`, with the script on standard input and the list name and the prefix as arguments. The script does nothing when the list does not exist. Otherwise it sets `completed` to `true` on each reminder of the list that is not completed and whose name starts with the prefix. A reminder with another name, or one that is already completed, is not touched.

It does not read the input that Claude Code writes on standard input.

## Output and exit codes

The hook never writes to standard output or standard error, because the standard output of a `UserPromptSubmit` hook is added to the conversation. It always exits `0`, so it never blocks the human's message.

| Situation | Standard output | Exit code |
| --- | --- | --- |
| Reminders completed | nothing | `0` |
| No reminder to complete, or no list | nothing | `0` |
| `osascript` is not installed, or fails | nothing | `0` |
| `osascript` does not finish in time | nothing, `osascript` is stopped | `0` |

## Settings

| Variable | Default | What it sets |
| --- | --- | --- |
| `COMPLETE_REMINDERS_TIMEOUT` | `10` | Seconds before `osascript` is stopped. It can have a fraction. A value that is not a number, or is not above zero, is replaced by `10` |

The list and the prefix are constants in [`internal/hook/hook.go`](../../internal/hook/hook.go), shared with [`stop-push-check`](stop-push-check.md).

## Tests

[`internal/hook/reminders_test.go`](../../internal/hook/reminders_test.go) has one test per scenario. The hook tests use a fake that records what it receives, and the `osascript` adapter tests use a stand-in `osascript` on the `PATH`, so the tests never touch the Reminders app. `rangers check full` runs them with `go test`.

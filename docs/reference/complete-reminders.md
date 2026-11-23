# The `complete-reminders` hook

[`bin/complete-reminders`](../../bin/complete-reminders) is a Claude Code `UserPromptSubmit` hook. When the human sends a message, it marks as completed every open reminder of the Reminders list `Tarefas Claude` whose name starts with `Claude espera: `. These are the reminders that the [`stop-push-check`](stop-push-check.md) hook requires at the end of a turn that waits for the human, so the human never completes them by hand. `rangers install` links it into `~/.claude/bin/` with the other [`bin/`](install-command.md#what-the-install-does) helpers. The hook is not active until it is registered in `~/.claude/settings.json`.

It needs macOS, `osascript` and the Reminders app. It does not need `jq`.

## Register it

Add this to the `hooks` object of `~/.claude/settings.json`, next to the `Stop` entry of [`stop-push-check`](stop-push-check.md#register-it):

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

## What it does

It runs one AppleScript through `osascript`, passing the list name and the prefix as arguments. The script does nothing when the list does not exist. Otherwise it sets `completed` to `true` on each reminder of the list that is not completed and whose name starts with the prefix. A reminder with another name, or one that is already completed, is not touched.

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
| `COMPLETE_REMINDERS_TIMEOUT` | `10` | Seconds before `osascript` is stopped |

The list and the prefix are set at the top of the file. They must match the ones in [`bin/stop-push-check`](../../bin/stop-push-check).

## Tests

[`tests/test_complete_reminders.py`](../../tests/test_complete_reminders.py) has one test per scenario. It runs the hook with a stand-in `osascript` that records what it receives, so the tests never touch the Reminders app. `scripts/check.sh full` runs it.

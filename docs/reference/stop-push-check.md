# The `stop-push-check` hook

[`bin/stop-push-check`](../../bin/stop-push-check) is a Claude Code `Stop` hook. It enforces one rule: a turn that ends waiting for the human must also call `PushNotification` and create a reminder in the Reminders list `Tarefas Claude`. The reminder is needed because the push does not reach the human's iPhone and a reminder does. The [`complete-reminders`](complete-reminders.md) hook closes it on the next human message. `rangers install` links it into `~/.claude/bin/` with the other [`bin/`](install-command.md#what-the-install-does) helpers. The hook is not active until it is registered in `~/.claude/settings.json`.

It needs `jq` on the `PATH`. The reminder needs macOS, the Reminders app and a list named `Tarefas Claude`.

## Register it

Add this to the `hooks` object of `~/.claude/settings.json`:

```json
{
  "hooks": {
    "Stop": [
      {
        "hooks": [
          { "type": "command", "command": "~/.claude/bin/stop-push-check" }
        ]
      }
    ]
  }
}
```

## Input

Claude Code writes the `Stop` event as JSON on standard input. The hook reads two fields.

| Field | Use |
| --- | --- |
| `stop_hook_active` | When `true`, the hook exits `0` without output, so a block cannot loop |
| `transcript_path` | The session transcript, a JSON Lines file |

## What it decides

The turn is every transcript entry after the last user entry that is not only tool results. The last assistant message of the turn is the last one that has text.

The hook blocks when both are true:

- that message asks the human something: its text ends with `?` (trailing whitespace is ignored) or contains `Responda` or `Answer sim`;
- an assistant entry of the turn does not call a tool whose name ends in `PushNotification`, or no assistant entry of the turn runs the reminder command.

The reminder command is a `Bash` tool call whose `command` contains all four of `osascript`, `Reminders`, `Tarefas Claude` and the prefix `Claude espera: `. A command that does not contain all four does not count, for example a reminder in another list or one named without the prefix.

A push call or a reminder in an earlier turn does not count.

The prefix `Claude espera: ` and the list `Tarefas Claude` are set at the top of [`bin/stop-push-check`](../../bin/stop-push-check) and of [`bin/complete-reminders`](../../bin/complete-reminders). The two files must hold the same values, and a test checks the prefix.

## Output and exit codes

| Situation | Standard output | Exit code |
| --- | --- | --- |
| Blocked | `{"decision":"block","reason":"..."}`, telling the assistant what is missing and to end the turn | `0` |
| Anything else | nothing | `0` |

The reason names only what is missing: `PushNotification`, the reminder, or both. For the reminder it gives the `osascript` command to run:

```sh
osascript -e 'tell application "Reminders" to make new reminder at end of reminders of list "Tarefas Claude" with properties {name:"Claude espera: <question>"}'
```

Anything else includes `stop_hook_active` set to `true`, input that is not JSON, a missing or unreadable transcript, an empty turn and a message that asks nothing. A broken line in the transcript is skipped.

## Tests

[`tests/test_stop_push_check.py`](../../tests/test_stop_push_check.py) has one test per scenario. `scripts/check.sh full` runs it.

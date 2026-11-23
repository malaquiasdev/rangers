# The `stop-push-check` hook

[`bin/stop-push-check`](../../bin/stop-push-check) is a Claude Code `Stop` hook. It enforces one rule: a turn that ends waiting for the human must also call `PushNotification`. `rangers install` links it into `~/.claude/bin/` with the other [`bin/`](install-command.md#what-the-install-does) helpers. The hook is not active until it is registered in `~/.claude/settings.json`.

It needs `jq` on the `PATH`.

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
- no assistant entry of the turn calls a tool whose name ends in `PushNotification`.

A push call in an earlier turn does not count.

## Output and exit codes

| Situation | Standard output | Exit code |
| --- | --- | --- |
| Blocked | `{"decision":"block","reason":"..."}`, telling the assistant to call `PushNotification` and end the turn | `0` |
| Anything else | nothing | `0` |

Anything else includes `stop_hook_active` set to `true`, input that is not JSON, a missing or unreadable transcript, an empty turn and a message that asks nothing. A broken line in the transcript is skipped.

## Tests

[`tests/test_stop_push_check.py`](../../tests/test_stop_push_check.py) has one test per scenario. `scripts/check.sh full` runs it.

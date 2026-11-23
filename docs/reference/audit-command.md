# The `rangers audit` command

`rangers audit` adds one row to the audit note of the day, in the notes folder. Agents run it after an action they took without asking, such as a merge, so the human can read what was done. The source is in [`cmd/rangers/`](../../cmd/rangers/) and [`internal/audit/`](../../internal/audit/). The messages are in Portuguese; this page quotes them as they are printed.

## Usage

```sh
rangers audit <where> <action> <link> <why>
```

| Argument | What it is |
| --- | --- |
| `where` | The repository or place, such as `owner/name` or `brain` |
| `action` | What was done, such as `merge` |
| `link` | A link that proves it, such as the pull request |
| `why` | The reason, in a few words |

More than four arguments are accepted and the extra ones are ignored. Each argument is written as it is, with no escape sequences.

`rangers install` also links `~/.claude/bin/audit` to the `rangers` program, which runs `rangers audit` when it is called by that name, with the same arguments. A command that already calls `~/.claude/bin/audit` keeps working.

## Where the note is

The note is `<notes>/claude/auditoria/<day>.md`, where `<day>` is the date in the `America/Sao_Paulo` time zone, as `YYYY-MM-DD`. The folder must already exist; the command does not create it.

| Variable | Default | What it sets |
| --- | --- | --- |
| `BRAIN_DIR` | `~/Documents/brain` | The notes folder. An empty value counts as not set |
| `CLAUDE_SESSION_ID` | none | The session written in the row |
| `AUDIT_SESSION` | none | The session, when `CLAUDE_SESSION_ID` is not set or is empty |

## What it writes

The first run of a day creates the note with a header and a table head, and adds a line that links to the note in `<notes>/claude/auditoria/index.md`. Every run then appends one row:

```text
| <HH:MM> | <session> | <where> | <action> | <link> | <why> |
```

`HH:MM` is the time in São Paulo. `<session>` is the first 8 characters of the session variable, or `?` when neither variable is set. The note and the index are created readable by the owner only.

## Output and exit codes

The command prints nothing when it works.

| Situation | Standard error | Exit code |
| --- | --- | --- |
| The row was written | nothing | `0` |
| Fewer than four arguments | `uso: audit <repo\|lugar> <ação> <link> <por quê>` | `2` |
| The note or the index cannot be written, such as when the folder does not exist | `rangers audit: <error>` | `1` |

## Tests

[`internal/audit/audit_test.go`](../../internal/audit/audit_test.go) has one test per scenario, with a fake note store and a fixed clock. One group writes real files in a temporary folder. `rangers check full` runs them with `go test`.

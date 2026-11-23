# The `rangers watch` command

`rangers watch` watches the agents listed in a file and ends when one of them is idle or gone, so a coordinator that runs it in the background is woken up. The source is in [`cmd/rangers/`](../../cmd/rangers/) and [`internal/agentwatch/`](../../internal/agentwatch/).

## Usage

```sh
rangers watch [limit]
```

`rangers install` also links `~/.claude/bin/agent-watch` to the `rangers` program, which runs `rangers watch` when it is called by that name, with the same arguments.

| Argument | Default | What it sets |
| --- | --- | --- |
| `limit` | `900` | Seconds without activity before an agent is reported idle. It must be a whole number, zero or more |

## The list

The list is a text file, read again on each round, with one agent per line:

```text
<name> <file> [<directory> ...]
```

| Part | What it is |
| --- | --- |
| `name` | How the agent is called in the output |
| `file` | A file the agent writes to, such as its transcript. A link is followed, so the time of the target counts |
| `directory` | Zero or more folders the agent works in |

Blank lines and lines that start with `#` are skipped. Fields are separated by spaces or tabs.

## What counts as activity

The last activity of an agent is the newest of:

- the modification time of `file`, when it is a regular file;
- the modification time of each regular file under each `directory`, leaving out any file or folder called `node_modules`, `.git` or `vendor`, and links;
- the time of the last Git commit of each `directory` that is a repository.

A `directory` that does not exist adds nothing.

## Output and exit codes

Each round reads the list and checks the agents in order. The command prints one line and ends at the first agent that matches:

| Situation | Standard output | Exit code |
| --- | --- | --- |
| The list is missing or empty | `lista vazia` | `0` |
| `file` is not a regular file and no `directory` is given, or no activity is found | `SUMIU <name>` | `0` |
| The last activity is older than `limit` seconds | `PARADO <name> há <N> min (última atividade: <HH:MM>)`, with the idle minutes rounded down and the local time | `0` |
| Every agent is active | nothing; it sleeps and reads the list again | keeps running |
| `limit` is not a whole number of zero or more, or there is more than one argument | the usage on standard error | `2` |
| Interrupted | nothing | `130` |

## Settings

| Variable | Default | What it sets |
| --- | --- | --- |
| `AGENT_WATCH_LIST` | `~/.claude/agent-watch.txt` | Path of the list |
| `AGENT_WATCH_INTERVAL` | `120` | Seconds to sleep between rounds. It can have a fraction. A value that is not a number, or is not above zero, is replaced by `120` |

## Tests

[`internal/agentwatch/agentwatch_test.go`](../../internal/agentwatch/agentwatch_test.go) has one test per scenario. The tests use real files with set modification times and a fake clock, sleep and Git, so none of them waits. One group runs `git` in a temporary repository. `rangers check full` runs them with `go test`.

# Dependencies

The list of what Rangers uses, with a pinned version or commit, the license and the source, is in [`deps.toml`](../../deps.toml). This page describes the format of that file, what `rangers install` covers and what the machine needs. For the version of a piece, read the file.

## Format of `deps.toml`

The file starts with `format_version = 1`. Each dependency is a `[[dependency]]` table.

Fields every dependency has. `scripts/check.sh` fails when one is missing:

| Field | What it holds |
| --- | --- |
| `name` | The short name of the piece |
| `kind` | The kind of piece, which sets the other fields |
| `license` | The license declared by the project |
| `function` | What Rangers uses the piece for, in one sentence |

Optional field:

| Field | What it holds |
| --- | --- |
| `optional` | `true` when Rangers works without the piece. `rangers install --check` does not count it as missing, and `rangers install` installs it only with `--optional` |

Fields by kind:

| `kind` | Its own fields |
| --- | --- |
| `claude-plugin` | `marketplace`, `plugin`, `version`, `commit`, `source` |
| `npm` | `package`, `version`, `source` |
| `brew` | `formula`, `version`, `source`, `requires` |
| `container-image` | `image` (with the tag), `source` |
| `skill` | `repository`, `path`, `commit` |

`requires` says what must exist on the machine for the piece to work.

## What `rangers install` covers

`deps.toml` and the command do not cover the same pieces:

| Piece | Pinned in `deps.toml` | Installed by `rangers install` | Checked by `rangers install --check` |
| --- | --- | --- | --- |
| pstack, i-have-adhd, container-use | Yes | Yes | Yes |
| openspec (`optional = true`) | Yes | Only with `--optional` | Shows `not installed (optional)` without counting it as missing |
| lychee (container image) | Yes | Yes | Yes |
| Skills (`kind = "skill"`) | Yes | Yes | Yes |
| Container runtime (`docker`) | No | No | Yes |

The skills live in the user's skills folder, `~/.claude/skills/`. `deps.toml` gives the repository, folder and commit each one comes from. `rangers install` copies each one and writes a `.rangers-commit` file with the commit. The steps to do it by hand are in [how to install the skills](../how-to/install-skills.md).

## How to install and remove each piece

| Piece | Install | Remove |
| --- | --- | --- |
| pstack | `claude plugin marketplace add cursor/plugins` and `claude plugin install pstack@cursor-plugins` | `claude plugin uninstall pstack@cursor-plugins` and `claude plugin marketplace remove cursor-plugins` |
| i-have-adhd | `claude plugin marketplace add ayghri/i-have-adhd` and `claude plugin install i-have-adhd@i-have-adhd` | `claude plugin uninstall i-have-adhd@i-have-adhd` and `claude plugin marketplace remove i-have-adhd` |
| container-use | `brew install dagger/tap/container-use` | `brew uninstall container-use` |
| openspec | `npm install -g @fission-ai/openspec` | `npm uninstall -g @fission-ai/openspec` |
| lychee | `podman pull <image>`, with the `image` from `deps.toml` | `podman rmi <image>` |

`rangers install` runs the same commands, and `npm install -g` uses the pinned `version`.

## What the machine needs

To build the `rangers` program and run `rangers install`:

- Go, at the version in [`go.mod`](../../go.mod);
- Git, because `rangers install` clones each skill;
- Podman, because `rangers install` pulls the lychee image and `--check` looks for it with `podman image exists`.

To use the agents:

- Claude Code, with the `claude` command in the terminal;
- Homebrew, for container-use;
- a container runtime that answers the `docker` command. Podman works when `docker` calls Podman, for example through a `docker` link that points to `podman`;
- `gh`, logged in, because the gate agent runs `rangers gate`, which reads and merges the pull request with `gh`;
- npm, only to install the optional openspec.

To schedule `rangers learn`:

- macOS, because the schedule uses launchd and the reminder uses `osascript` and the Reminders app, with a list named `Tarefas Claude`;
- Git, because the rules pass creates a worktree.

To run the repository checks:

- Bash and Git;
- Go with `gofmt`, and `golangci-lint`;
- `python3` 3.11 or newer, which has the `tomllib` module;
- Podman, with its machine running, for the link check of `make check`.

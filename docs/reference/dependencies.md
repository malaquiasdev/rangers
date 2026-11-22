# Dependencies

The list of what Rangers uses, with a pinned version or commit, the license and the source, is in [`deps.toml`](../../deps.toml). This page describes the format of that file, what the scripts cover and what the machine needs. For the version of a piece, read the file.

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
| `optional` | `true` when Rangers works without the piece. `verify.sh` does not count it as missing, and `install.sh` installs it only when asked |

Fields by kind:

| `kind` | Its own fields |
| --- | --- |
| `claude-plugin` | `marketplace`, `plugin`, `version`, `commit`, `source` |
| `npm` | `package`, `version`, `source` |
| `brew` | `formula`, `version`, `source`, `requires` |
| `container-image` | `image` (with the tag), `source` |
| `skill` | `repository`, `path`, `commit` |

`requires` says what must exist on the machine for the piece to work.

## What the scripts cover

`deps.toml` and the scripts do not list the same pieces:

| Piece | Pinned in `deps.toml` | Installed by `install.sh` | Checked by `verify.sh` |
| --- | --- | --- | --- |
| pstack, i-have-adhd, container-use | Yes | Yes | Yes |
| openspec (`optional = true`) | Yes | Only with `INSTALL_OPTIONAL=1` | Shows `não instalado (opcional)` without counting it as missing |
| lychee (container image) | Yes | No | No |
| Skills (`kind = "skill"`) | Yes | No | No |
| Container runtime (`docker`) | No | No | Yes |

The skills live in the user's skills folder, `~/.claude/skills/`. `deps.toml` gives the repository, folder and commit each one comes from. The steps to copy one are in [how to install the skills](../how-to/install-skills.md).

## How to install and remove each piece

| Piece | Install | Remove |
| --- | --- | --- |
| pstack | `claude plugin marketplace add cursor/plugins` and `claude plugin install pstack@cursor-plugins` | `claude plugin uninstall pstack@cursor-plugins` and `claude plugin marketplace remove cursor-plugins` |
| i-have-adhd | `claude plugin marketplace add ayghri/i-have-adhd` and `claude plugin install i-have-adhd@i-have-adhd` | `claude plugin uninstall i-have-adhd@i-have-adhd` and `claude plugin marketplace remove i-have-adhd` |
| container-use | `brew install dagger/tap/container-use` | `brew uninstall container-use` |
| openspec | `npm install -g @fission-ai/openspec` | `npm uninstall -g @fission-ai/openspec` |

The install commands are the same as in [`scripts/install.sh`](../../scripts/install.sh).

## What the machine needs

To use the agents:

- Claude Code, with the `claude` command in the terminal;
- Homebrew, for container-use;
- a container runtime that answers the `docker` command. Podman works when `docker` calls Podman, for example through a `docker` link that points to `podman`;
- npm, only to install the optional openspec.

To build and schedule `rangers learn`:

- Go, at the version in [`go.mod`](../../go.mod);
- macOS, because the schedule uses launchd and the reminder uses `osascript` and the Reminders app, with a list named `Tarefas Claude`;
- Git, because the rules pass creates a worktree.

To run the repository checks:

- Bash and Git;
- Go with `gofmt`, and `golangci-lint`;
- `python3` 3.11 or newer, which has the `tomllib` module;
- Podman, with its machine running, for the link check of `make check`.

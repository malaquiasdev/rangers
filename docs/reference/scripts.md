# Scripts

The scripts in [`scripts/`](../../scripts/) are Bash. They install the pieces, run the repository checks and turn on the Git hooks. Their messages are in Portuguese; this page quotes them as they are printed.

| Script | What it does |
| --- | --- |
| [`install.sh`](#scriptsinstallsh) | Installs the pieces that are missing |
| [`check.sh`](#scriptschecksh) | Runs the repository checks |
| [`install-hooks.sh`](#scriptsinstall-hookssh) | Points Git to the hooks in `.githooks/` |

`scripts/characters.txt` is not a script. It holds the names that the checks forbid outside `agents/` and `skills/`, as one extended regular expression.

## `scripts/install.sh`

Installs each piece that is missing and skips the ones already there. Running it again does not touch the pieces already installed.

| Piece | How it knows the piece exists | What it runs when the piece is missing |
| --- | --- | --- |
| pstack | `claude plugin list` contains `pstack@cursor-plugins` | `claude plugin marketplace add cursor/plugins` and `claude plugin install pstack@cursor-plugins` |
| i-have-adhd | `claude plugin list` contains `i-have-adhd@i-have-adhd` | `claude plugin marketplace add ayghri/i-have-adhd` and `claude plugin install i-have-adhd@i-have-adhd` |
| container-use | The `container-use` command exists | `brew install dagger/tap/container-use` |
| openspec | The `openspec` command exists. Only checked when `INSTALL_OPTIONAL=1` | `npm install -g @fission-ai/openspec` |

At the end it prints `Feito. Reinicie o Claude Code e rode /pstack:setup-pstack uma vez.` and the hint `Para conferir: rangers install --check`. It stops at the first command that fails (`set -eu`). To see the state of each piece, use [`rangers install --check`](install-command.md).

It does not:

- install or start the container runtime;
- pin versions: it installs the current version of each source;
- install the skills or the lychee image listed in `deps.toml`;
- link the agents or the skill into Claude Code (see [how to connect the agents](../how-to/connect-agents.md));
- set up container-use as an MCP server;
- build the `rangers` program.

## `scripts/check.sh`

Runs the repository checks from the root of the clone. It takes one argument, the mode: `quick` or `full`. Without an argument it runs `full`.

What each check looks at and which mode runs it is in the [checks and release reference](checks-and-release.md#what-scriptschecksh-checks).

| Situation | Last line | Exit code |
| --- | --- | --- |
| Every check passed | `tudo certo (<mode>)` | `0` |
| One or more checks failed | `N checagem(ns) falharam`, after one `FALHOU:` line per failure | `1` |

## `scripts/install-hooks.sh`

Makes the files in `.githooks/` and `scripts/*.sh` executable and runs `git config core.hooksPath .githooks`. Prints `ganchos ligados: pre-commit, commit-msg e pre-push`. `make hooks` runs it.

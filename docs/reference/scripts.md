# Scripts

The scripts in [`scripts/`](../../scripts/) are Bash. They install and check the pieces, run the repository checks, turn on the Git hooks and schedule the learning job. Their messages are in Portuguese; this page quotes them as they are printed.

| Script | What it does |
| --- | --- |
| [`verify.sh`](#scriptsverifysh) | Shows the state of each piece and changes nothing |
| [`install.sh`](#scriptsinstallsh) | Installs the pieces that are missing |
| [`check.sh`](#scriptschecksh) | Runs the repository checks |
| [`install-hooks.sh`](#scriptsinstall-hookssh) | Points Git to the hooks in `.githooks/` |
| [`learn-schedule.sh`](#scriptslearn-schedulesh) | Installs, removes or shows the launchd job of `rangers learn` |

`scripts/characters.txt` is not a script. It holds the names that the checks forbid outside `agents/` and `skills/`, as one extended regular expression.

## `scripts/verify.sh`

Shows the state of each piece, one per line, and changes nothing on the machine. It runs from any folder.

Each line has the name of the piece and the version the program itself reports. A piece that is not installed shows `FALTANDO`. The last piece line checks whether the `docker` command answers `docker info`: it shows `respondendo` or `SEM RESPOSTA (o container-use precisa dele)`.

| Piece | How the script finds the version |
| --- | --- |
| Claude Code plugins (pstack, i-have-adhd) | Finds the plugin in the output of `claude plugin list` and reads the next `Version:` line |
| container-use | Checks that the command exists and reads the first line of `container-use version` |
| openspec | Same as above with `openspec --version`. When it is missing, shows `não instalado (opcional)` and does not count it as missing |
| Container runtime | `docker info` |

| Situation | Last line | Exit code |
| --- | --- | --- |
| Nothing missing | `Tudo no lugar.` | `0` |
| One or more pieces missing, or `docker` not answering | `N item(ns) faltando. Rode scripts/install.sh` | `1` |

The script does not compare versions with `deps.toml`. It only says whether the piece exists and which version is installed.

## `scripts/install.sh`

Installs each piece that is missing and skips the ones already there. Running it again does not touch the pieces already installed.

| Piece | How it knows the piece exists | What it runs when the piece is missing |
| --- | --- | --- |
| pstack | `claude plugin list` contains `pstack@cursor-plugins` | `claude plugin marketplace add cursor/plugins` and `claude plugin install pstack@cursor-plugins` |
| i-have-adhd | `claude plugin list` contains `i-have-adhd@i-have-adhd` | `claude plugin marketplace add ayghri/i-have-adhd` and `claude plugin install i-have-adhd@i-have-adhd` |
| container-use | The `container-use` command exists | `brew install dagger/tap/container-use` |
| openspec | The `openspec` command exists. Only checked when `INSTALL_OPTIONAL=1` | `npm install -g @fission-ai/openspec` |

At the end it prints `Feito. Reinicie o Claude Code e rode /pstack:setup-pstack uma vez.` and runs `verify.sh`. Its exit code is the one from `verify.sh`. It stops at the first command that fails (`set -eu`).

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

## `scripts/learn-schedule.sh`

Manages the launchd job that runs `rangers learn` on macOS. It takes one argument.

| Argument | What it does |
| --- | --- |
| `install` | Writes `~/Library/LaunchAgents/com.rangers.learn.plist`, creates the state folder `~/.local/state/rangers/learn` and loads the job. Prints `agendado: a cada 2 horas, log em <state>/launchd.log` |
| `uninstall` | Unloads the job and deletes the file. Prints `agendamento removido` |
| `status` (default) | Shows the state, number of runs, last exit code and interval of the job, or `não agendado` |
| anything else | Prints the usage and exits with code `2` |

The job:

- runs `~/.local/bin/rangers learn`, with no flags, every 7200 seconds and once when it loads;
- sets only `HOME` and a `PATH` with `~/.local/bin`, `/opt/homebrew/bin`, `/usr/local/go/bin`, `/usr/local/bin`, `/usr/bin` and `/bin`;
- writes standard output and standard error to `~/.local/state/rangers/learn/launchd.log`.

`install` does not build the program. It expects the binary at `~/.local/bin/rangers`.

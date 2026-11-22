# Scripts

The scripts in [`scripts/`](../../scripts/) are Bash. They run the repository checks and turn on the Git hooks. Installing the pieces is done by [`rangers install`](install-command.md). Their messages are in Portuguese; this page quotes them as they are printed.

| Script | What it does |
| --- | --- |
| [`check.sh`](#scriptschecksh) | Runs the repository checks |
| [`install-hooks.sh`](#scriptsinstall-hookssh) | Points Git to the hooks in `.githooks/` |

`scripts/characters.txt` is not a script. It holds the names that the checks forbid, as one extended regular expression. Which files are exempt is in the [checks and release reference](checks-and-release.md#what-scriptschecksh-checks).

## `scripts/check.sh`

Runs the repository checks from the root of the clone. It takes one argument, the mode: `quick` or `full`. Without an argument it runs `full`.

What each check looks at and which mode runs it is in the [checks and release reference](checks-and-release.md#what-scriptschecksh-checks).

| Situation | Last line | Exit code |
| --- | --- | --- |
| Every check passed | `tudo certo (<mode>)` | `0` |
| One or more checks failed | `N checagem(ns) falharam`, after one `FALHOU:` line per failure | `1` |

## `scripts/install-hooks.sh`

Makes the files in `.githooks/` and `scripts/*.sh` executable and runs `git config core.hooksPath .githooks`. Prints `ganchos ligados: pre-commit, commit-msg e pre-push`. `make hooks` runs it.

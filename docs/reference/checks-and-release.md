# Checks and release

The repository runs its checks on the machine, through `make` and Git hooks. GitHub Actions runs only the release. How to turn the checks on is in [how to run the checks](../how-to/run-checks.md).

## `make` targets

The targets are in the [`Makefile`](../../Makefile).

| Target | What it runs |
| --- | --- |
| `make check` | `go run ./cmd/rangers check full` |
| `make quick` | `go run ./cmd/rangers check quick` |
| `make hooks` | `go run ./cmd/rangers hooks install` |

## Git hooks

The hooks are in [`.githooks/`](../../.githooks/). Git uses them after `make hooks` sets `core.hooksPath`. Each file is one command that runs [`rangers hooks run`](../reference/hooks-command.md), so the `rangers` program must be in your `PATH`.

| Hook | What it does |
| --- | --- |
| `pre-commit` | Runs `rangers check quick` |
| `commit-msg` | Refuses a message whose first line does not follow Conventional Commits, and a message that has a name from `scripts/characters.txt` |
| `pre-push` | Runs `rangers check full` |

The types that `commit-msg` accepts are `feat`, `fix`, `docs`, `chore`, `refactor`, `test`, `build`, `ci`, `perf`, `style` and `revert`, with an optional scope in lower case and an optional `!`.

## What `rangers check` checks

| Check | Fails when | Mode |
| --- | --- | --- |
| Shell syntax | `bash -n` fails on a file in `scripts/*.sh`, `.githooks/` or `bin/` | both |
| Shell comments | One of those files has a comment line after the first line | both |
| Agent files | In a file of `agents/`, `name` differs from the file name, `model` is not `opus`, `sonnet` or `haiku`, `effort` is not `low`, `medium`, `high`, `xhigh` or `max`, `description` is empty, or the `Character:` line is missing | both |
| Skill files | In `skills/<folder>/SKILL.md`, `name` differs from the folder | both |
| `deps.toml` | The file cannot be read as TOML, or a dependency lacks `name`, `kind`, `license` or `function`. Only the first missing field is shown | both |
| Character names | A name from `scripts/characters.txt` appears outside `agents/`, `skills/`, `.git/`, `.serena/`, any file named `README.md` and `scripts/characters.txt` itself. Lines with `--agent ` are ignored | both |
| Privacy | A pattern from the privacy file appears anywhere outside `.git/` | both |
| Go comments | A `.go` file has `//` or a line starting with `/*` | both |
| Go format | `gofmt -l .` lists a file | both |
| Go vet and tests | `go vet ./...` or `go test ./...` fails | `full` |
| Go lint | `golangci-lint run` fails, with the rules in [`.golangci.yml`](../../.golangci.yml) | `full` |
| Pixel art tests | `python3 -I -m unittest discover` fails in `assets/tests` | `full` |
| Documentation links | lychee, run with Podman, finds a broken link or anchor in `README.md` or `docs/` | `full` |

The Go checks run only when `go.mod` exists. The checks skip the `.serena/` folder, binary files and links. The findings of a check come sorted by path.

`scripts/characters.txt` holds the forbidden names as one regular expression, with `|` between the names. A name matches only as a whole word and with the same case, so a longer word that contains it does not match. The same file feeds the `commit-msg` hook.

## The privacy file

The privacy check reads a file of patterns that never go into the repository, such as company names.

- Path: `~/.config/rangers/privacy.txt`, or the path in the `PRIVACY_FILE` variable.
- Format: one regular expression per line, in the syntax of Go's `regexp` package, which has no back-references. Blank lines are ignored. The match ignores case.
- The file lives only on the machine. It is not in the repository.

When the file does not exist, the check is skipped and the command prints `aviso: sem <path>, a checagem de nomes de empresa foi pulada`. The run does not fail.

## Release

[`.github/workflows/release-please.yml`](../../.github/workflows/release-please.yml) is the only workflow. It runs release-please on every push to `main`, with a 5-minute timeout. The action is pinned to a commit.

release-please reads the Conventional Commits on `main`, keeps a release pull request open with the next version and the changelog, and creates the release when that pull request is merged.

| File | What it holds |
| --- | --- |
| [`release-please-config.json`](../../release-please-config.json) | Release type `simple`, changelog at `CHANGELOG.md`, the commit release-please starts from and the version of the first release |
| [`.release-please-manifest.json`](../../.release-please-manifest.json) | The current version |

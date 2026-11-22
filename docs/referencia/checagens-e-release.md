# Checks and release

The repository runs its checks on the machine, through `make` and Git hooks. GitHub Actions runs only the release. How to turn the checks on is in [how to run the checks](../como-fazer/rodar-as-checagens.md).

## `make` targets

The targets are in the [`Makefile`](../../Makefile).

| Target | What it runs |
| --- | --- |
| `make check` | `scripts/checar.sh completo` |
| `make rapido` | `scripts/checar.sh rapido` |
| `make hooks` | `scripts/hooks-instalar.sh` |

## Git hooks

The hooks are in [`.githooks/`](../../.githooks/). Git uses them after `make hooks` sets `core.hooksPath`.

| Hook | What it does |
| --- | --- |
| `pre-commit` | Runs `scripts/checar.sh rapido` |
| `commit-msg` | Refuses a message whose first line does not follow Conventional Commits, and a message that has a name from `scripts/personagens.txt` |
| `pre-push` | Runs `scripts/checar.sh completo` |

The types that `commit-msg` accepts are `feat`, `fix`, `docs`, `chore`, `refactor`, `test`, `build`, `ci`, `perf`, `style` and `revert`, with an optional scope in lower case and an optional `!`.

## What `scripts/checar.sh` checks

| Check | Fails when | Mode |
| --- | --- | --- |
| Shell syntax | `bash -n` fails on a file in `scripts/*.sh` or `.githooks/` | both |
| Shell comments | One of those files has a comment line after the first line | both |
| Agent files | In a file of `agents/`, `name` differs from the file name, `model` is not `opus`, `sonnet` or `haiku`, `effort` is not `low`, `medium`, `high`, `xhigh` or `max`, `description` is empty, or the `Character:` line is missing | both |
| Skill files | In `skills/<folder>/SKILL.md`, `name` differs from the folder | both |
| `deps.toml` | Python's `tomllib` cannot read it, or a dependency lacks `name`, `kind`, `license` or `function` | both |
| Character names | A name from `scripts/personagens.txt` appears outside `agents/`, `skills/` and `.git/`. Lines with `--agent ` are ignored | both |
| Privacy | A pattern from the privacy file appears anywhere outside `.git/` | both |
| Go comments | A `.go` file has `//` or a line starting with `/*` | both |
| Go format | `gofmt -l .` lists a file | both |
| Go vet and tests | `go vet ./...` or `go test ./...` fails | `completo` |
| Go lint | `golangci-lint run` fails, with the rules in [`.golangci.yml`](../../.golangci.yml) | `completo` |
| Documentation links | lychee, run with Podman, finds a broken link or anchor in `README.md` or `docs/` | `completo` |

The Go checks run only when `go.mod` exists. The checks skip the `.serena/` folder.

## The privacy file

The privacy check reads a file of patterns that never go into the repository, such as company names.

- Path: `~/.config/rangers/privacidade.txt`, or the path in the `PRIVACIDADE_ARQUIVO` variable.
- Format: one extended regular expression per line. The match ignores case.
- The file lives only on the machine. It is not in the repository.

When the file does not exist, the check is skipped and the script prints `aviso: sem <path>, a checagem de privacidade foi pulada`. The run does not fail.

## Release

[`.github/workflows/release-please.yml`](../../.github/workflows/release-please.yml) is the only workflow. It runs release-please on every push to `main`, with a 5-minute timeout. The action is pinned to a commit.

release-please reads the Conventional Commits on `main`, keeps a release pull request open with the next version and the changelog, and creates the release when that pull request is merged.

| File | What it holds |
| --- | --- |
| [`release-please-config.json`](../../release-please-config.json) | Release type `simple`, changelog at `CHANGELOG.md`, the commit release-please starts from and the version of the first release |
| [`.release-please-manifest.json`](../../.release-please-manifest.json) | The current version |

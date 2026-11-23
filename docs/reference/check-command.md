# The `rangers check` command

`rangers check` runs the repository checks, in the order of the table in the [checks and release reference](checks-and-release.md#what-rangers-check-checks). The source is in [`cmd/rangers/`](../../cmd/rangers/) and [`internal/check/`](../../internal/check/). The messages are in Portuguese; this page quotes them as they are printed.

## Usage

```sh
rangers check [quick|full]
```

Run it from inside a clone of the repository. It works from the root of the clone, found with `git rev-parse --show-toplevel`. Without an argument it runs `full`. `quick` skips the Go vet, tests and lint, the tests of the pixel art and the link check; any other value skips them too, and shows up in the last line.

`make check` and `make quick` run it through `go run`, so they always use the source in front of you. The `pre-commit` and `pre-push` hooks run it through [`rangers hooks run`](hooks-command.md).

## Settings

| Variable | Default | What it sets |
| --- | --- | --- |
| `PRIVACY_FILE` | `~/.config/rangers/privacy.txt` | The [privacy file](checks-and-release.md#the-privacy-file) |

## Output and exit codes

Each failed check prints its findings, if it has any, and then one `FALHOU:` line. The tools that run on the terminal, such as `go test`, print as they go.

| Situation | Last line | Exit code |
| --- | --- | --- |
| Every check passed | `tudo certo (<mode>)` | `0` |
| One or more checks failed | `N checagem(ns) falharam`, after one `FALHOU:` line per failure | `1` |
| `scripts/characters.txt` or the repository root cannot be read | the error on standard error, with no summary | `1` |

A line that lists a finding looks like `./docs/page.md:3:the text of the line`, with the path from the root and the line number. The findings of a check come sorted by path.

## Tests

[`internal/check/`](../../internal/check/) has one test per scenario of each check, in both modes. The tests use an in-memory file system and a fake that records each tool call, so none of them runs `go`, `podman` or `python3`. One group runs real commands in a temporary folder to test the runner. `rangers check full` runs them with `go test`.

# The `rangers hooks` command

`rangers hooks` is the Git side of the repository checks: it turns the hooks on and runs them. It is not the [`rangers hook`](stop-push-check.md) command, which is for Claude Code hooks. The source is in [`cmd/rangers/`](../../cmd/rangers/) and [`internal/githook/`](../../internal/githook/). The messages are in Portuguese; this page quotes them as they are printed.

## Usage

```sh
rangers hooks install
rangers hooks run pre-commit
rangers hooks run commit-msg <file>
rangers hooks run pre-push
```

Run it from inside a clone of the repository. `make hooks` runs `rangers hooks install` through `go run`.

## What `install` does

It makes each file in `.githooks/` executable and runs `git config core.hooksPath .githooks`. It prints `ganchos ligados: pre-commit, commit-msg e pre-push`.

## What `run` does

The three files in [`.githooks/`](../../.githooks/) have one command each, such as `exec rangers hooks run pre-commit`. Git runs them, and they need `rangers` in the `PATH`. If you changed the program, build it again, as in [the install steps](../tutorials/prepare-the-machine.md#1-see-what-is-already-there).

| Hook | What it does |
| --- | --- |
| `pre-commit` | Runs [`rangers check quick`](check-command.md) |
| `pre-push` | Runs [`rangers check full`](check-command.md) |
| `commit-msg` | Reads the message from `<file>` and refuses it, in this order, when the first line does not follow Conventional Commits, and when any line has a name from `scripts/characters.txt` as a whole word |

The types that `commit-msg` accepts are `feat`, `fix`, `docs`, `chore`, `refactor`, `test`, `build`, `ci`, `perf`, `style` and `revert`, with an optional scope in lower case, digits and `-`, an optional `!` and then `: ` and a subject.

## Output and exit codes

| Situation | Output | Exit code |
| --- | --- | --- |
| The hook passed | what `rangers check` prints, or nothing for `commit-msg` | `0` |
| `commit-msg` refuses the format | `a mensagem precisa seguir o Conventional Commits` on standard output | `1` |
| `commit-msg` refuses a name | `nome de personagem na mensagem do commit` on standard output | `1` |
| A check failed | see [the check output](check-command.md#output-and-exit-codes) | `1` |
| Unknown command or hook, or `commit-msg` without a file | the usage on standard error | `2` |
| A file cannot be read or `git config` fails | the error on standard error | `1` |

## Tests

[`internal/githook/githook_test.go`](../../internal/githook/githook_test.go) has one test per scenario, with a fake file store and a fake Git. The adapters are tested with real files and a real repository in a temporary folder. `rangers check full` runs them with `go test`.

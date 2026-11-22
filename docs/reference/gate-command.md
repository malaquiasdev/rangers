# The `rangers gate` command

`rangers gate` decides, with no model, whether a pull request may be merged. It checks the pull request, stops at production, looks for the human's yes bound to the head commit, and reads the CI. With `--merge` it also squashes the pull request when every check passes. The source is in [`cmd/rangers/`](../../cmd/rangers/) and [`internal/gate/`](../../internal/gate/). Where the gate fits among the agents is in [how the agents split the work](../explanation/how-agents-split-work.md).

## Usage

```sh
rangers gate --repo owner/name --pr N [--merge] [--transcripts DIR]
rangers gate ask --repo owner/name --pr N
```

| Flag | Default | What it sets |
| --- | --- | --- |
| `--repo` | none, required | Repository, as `owner/name` |
| `--pr` | none, required | Pull request number |
| `--merge` | off | Squashes the pull request when every check passes. Without it the command is a dry run that only prints the verdict |
| `--transcripts` | `~/.claude/projects` | Folder of the Claude Code session transcripts, where the human's yes is looked up |

`ask` takes only `--repo` and `--pr`. The command runs `gh`, so `gh` must be installed and logged in.

## The checks

The checks run in this order and the gate stops at the first one that refuses.

| # | Check | Passes when |
| --- | --- | --- |
| 1 | `pr` | The pull request is open, is not a draft and its base is the default branch of the repository. The head SHA is read here |
| 2 | `production` | The title does not start with `chore(main): release`, the head branch does not start with `release-please--` and no label is `production` (any case) |
| 3 | `approval` | The human said yes to the question for this head SHA (see below) |
| 4 | `ci` | `gh pr checks` lists at least one check and every one has the bucket `pass` or `skipping` |
| 5 | `merge` | Only with `--merge`: `gh pr merge` squashes the pull request, pinned with `--match-head-commit` to the head SHA read in check 1 |

The production stop comes before the approval and nothing overrides it: a release pull request is refused even when the human said yes. Its verdict is `production needs the human's explicit yes; the gate stops here`.

A check that is pending, failing or cancelled refuses, and the verdict lists each one as `name (bucket)`. A pull request with no checks is refused too.

## The approval

The approval is bound to the head commit, so a push after the question makes the old yes useless.

1. The hash is the first 8 hex characters of the SHA-256 of `owner/name#N@headSHA`.
2. `rangers gate ask` prints the hash on the first line and, on the second, the line the coordinator pastes into its question to the human: `Merge owner/name#N at <sha7>? Answer sim. gate:<hash>`.
3. The gate reads the transcripts under `--transcripts`, in files changed in the last 48 hours. In one session, it looks for an assistant message that contains `gate:<hash>` and takes the next genuine human message.
4. That message must be exactly `yes`, `sim`, `ok` or `pode`, in any case, ignoring spaces around it and punctuation at its end. Anything else, such as `no` or `yes please`, cancels the question; a later yes needs a new question.

Tool results and messages that Claude Code injects, such as system notifications and task notifications, are not human messages. They are skipped, so they neither approve nor cancel. Only a yes from the human in the same session as the question counts.

## Output and exit codes

The output has at most 560 characters: one line per check that ran, then the verdict as the last line. The line of a check is `<name>: ok <detail>`, or `<name>: refused` when it stopped the gate.

```text
pr: ok open, ready, base main
production: ok not a release
approval: ok the human said yes to gate:4a7010e4
ci: ok 3 checks passed or skipped
allowed: dry run, nothing merged
```

| Verdict | Exit code |
| --- | --- |
| `allowed: dry run, nothing merged`, or `allowed: merged` with `--merge` | `0` |
| `refused: <reason>` | `1` |
| `gh` failed, the output could not be read or the flags are wrong. The error goes to standard error and nothing is merged | `2` |

`gh pr checks` exits with a non-zero code while a check is pending or failing. The gate reads its output anyway and refuses, instead of reporting an error.

## The audit line

After a merge, when `~/.claude/bin/audit` exists, the gate runs it once with the repository, `merge`, the link of the pull request and the reason. The `merge` line then ends with `; audit recorded`, or `; audit not recorded` when the script failed. A failed audit does not undo the merge and does not change the exit code. Without the script nothing is added.

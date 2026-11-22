# The `rangers learn` command

`rangers learn` reads the Claude Code session transcripts that changed and runs the learning agent on them. The source is in [`cmd/rangers/`](../../cmd/rangers/) and [`internal/learn/`](../../internal/learn/). Why it works this way is in [how the learning loop works](../explanation/how-learning-works.md).

## Usage

```sh
rangers learn [flags]
rangers learn metrics [flags]
```

| Call | What it does |
| --- | --- |
| `rangers learn` | Runs the rules pass and then the notes pass |
| `rangers learn metrics` | Prints the table of past rounds from the state folder and runs no model |
| `rangers` with no word, or a first word other than `learn`, `schedule`, `install` or `gate` | Prints the usage line below to standard error and exits with code `2` |

```text
uso: rangers learn [metrics] [flags] | rangers schedule install|uninstall|status | rangers install [--check] | rangers gate [ask] --repo owner/name --pr N [--merge]
```

When a pass fails, the command prints `rangers learn: <error>` to standard error and exits with code `1`. A pass that a gate stops is not a failure: it prints the reason and exits with `0`.

## Flags

Each flag also reads an environment variable. The flag wins over the variable, and the variable wins over the default.

| Flag | Variable | Default | What it sets |
| --- | --- | --- | --- |
| `--interval-hours` | `LEARN_INTERVAL_HOURS` | `2` | Minimum time since the last checkpoint of the pass, in hours |
| `--min-lines` | `LEARN_MIN_LINES` | `10` | Minimum number of new lines typed by the human |
| `--budget-usd` | `LEARN_BUDGET_USD` | `3` | Dollar cap for one round of the model |
| `--daily-budget-usd` | `LEARN_DAILY_BUDGET_USD` | `6` | Dollar cap for all rounds of the day, both passes together |
| `--model` | `LEARN_MODEL` | `opus` | Model of the rules pass |
| `--notes-model` | `LEARN_NOTES_MODEL` | `sonnet` | Model of the notes pass |
| `--notes-max-bytes` | `LEARN_NOTES_MAX_BYTES` | `60000` | Size limit of the notes pass input file |
| `--skip-notes` | `LEARN_NOTES=off` | off | Skips the notes pass |
| `--call-timeout-minutes` | `LEARN_CALL_TIMEOUT_MINUTES` | `20` | Time limit of each `claude` call, in minutes. When it runs out, the process is killed and the pass fails with an error that says so |
| `--window-hours` | `LEARN_WINDOW_HOURS` | `24` | How far back to read when a pass has no checkpoint yet |
| `--state` | `LEARN_STATE` | `~/.local/state/rangers/learn` | State folder |
| `--repo` | `LEARN_REPO` | `~/Developer/me/rangers`, a fixed path | Clone of this repository, where the agents and the rules live |
| `--projects` | `LEARN_PROJECTS` | `~/.claude/projects` | Folder of the Claude Code transcripts |
| `--brain` | `LEARN_BRAIN` | `~/Documents/brain`, a fixed path | Notes folder where the notes pass writes |

The defaults of `--repo` and `--brain` are fixed paths under the home folder. The command does not look for the clone from the current folder, as `rangers install` does. When your clone or your notes folder is somewhere else, pass `--repo` and `--brain`, or set `LEARN_REPO` and `LEARN_BRAIN`. The scheduled job sets neither, so it only works with the defaults (see [Schedule](#schedule)).

The command refuses to start when `--budget-usd`, `--daily-budget-usd`, `--notes-max-bytes` or `--call-timeout-minutes` is zero or less, or when a variable holds a value that is not a number.

## The two passes

The rules pass runs first. The notes pass runs after it, even when the rules pass failed. Both errors, if any, are reported together.

| | Rules pass | Notes pass |
| --- | --- | --- |
| Reads | Only what the human typed | What the human typed and what the assistant answered, in time order |
| Runs in | A new worktree of `--repo` | The `--brain` folder |
| Model | `--model` | `--notes-model` |
| Tools the model may use | `Read`, `Edit`, `Write`, and `git add`, `git commit`, `git status`, `git diff` | `Read`, `Write`, `Edit`, `Glob`, `Grep`, and `audit`, allowed as `~/.claude/bin/audit`, as the same path with the home folder written out, and, when that file is a link, as the path it points to. A call that goes through a shell variable or is chained with other commands still needs approval |
| Leaves | A local commit on branch `learn/<stamp>` and a reminder | Notes in the notes folder |
| Checkpoint file | `checkpoint` | `checkpoint-notes` |
| Prefix of its messages | none | `[notas] ` |

Both passes run the same agent: the first file in `<repo>/agents/` whose `description` starts with `Learning.`. If no file matches, the pass fails. The model is started with `claude -p`, permission mode `acceptEdits`, the cap of the round as `--max-budget-usd` and JSON output. Variables that start with `GIT_` are removed from its environment.

`<stamp>` is the local date and time of the run, as `YYYY-MM-DD-HHMM`.

## When the notes folder is not accessible

Before the notes pass writes its input or calls the model, it lists `--brain` and creates and removes a temporary file in it, giving up after 10 seconds. If that fails, the pass does not call the model, does not write its input and does not touch `checkpoint-notes`. It prints this to standard output, with the path of the running program, and returns no error:

```text
[notas] sem acesso ao brain (<error>): dê Acesso Total ao Disco a <path of the rangers binary> em Ajustes > Privacidade e Segurança; pulo as notas
```

It also appends a line titled `Dar acesso ao brain ao rangers` to `pending-reminders.txt`. The rules pass is not affected. A scheduled job gets this error on a folder under `~/Documents` until the program has Full Disk Access (see [Give the job access to the notes folder](../how-to/schedule-learning.md#give-the-job-access-to-the-notes-folder)).

## Gates

Each pass checks its gates in this order and stops at the first one that blocks:

1. **Interval.** If the checkpoint of the pass is younger than `--interval-hours`, the pass stops. Without a checkpoint, the pass reads the last `--window-hours`.
2. **Lines.** If the human typed fewer than `--min-lines` lines since the checkpoint, the pass stops without calling the model.
3. **Daily cap.** If the rounds of today (by the local date of the stamp), both passes, already cost `--daily-budget-usd` or more, the pass stops without calling the model.

The cap of a round is the smaller of `--budget-usd` and what is left of the daily cap.

A pass that stops at a gate does not touch its checkpoint, so the lines wait for the next run.

## What the input holds

The command reads the `*.jsonl` files in `--projects` and one folder below it, when the file changed after the checkpoint. From each file it keeps the messages written after the checkpoint.

- Messages that Claude Code injects, such as command output and caveats, are dropped. The list of prefixes is `ignoredPrefixes` in [`internal/transcript/transcript.go`](../../internal/transcript/transcript.go).
- Line breaks become spaces. A human message is cut at 600 characters, an assistant message at 400.
- The rules pass input has one line per message: session, time and text.
- The notes pass input has one line per message: project folder, session, time and text, with `[assistant] ` before the assistant's text. When the file would pass `--notes-max-bytes`, the oldest lines are dropped.

## What a rules round does

1. Writes the input to `<state>/input-<stamp>.txt`.
2. Creates the branch `learn/<stamp>` from `main` in a new worktree at `<repo>-worktrees/learn-<stamp>`.
3. Runs the model in that worktree. The agent edits the agent files or `rules/` and commits on the branch.
4. Saves the answer, appends a row to `metrics.tsv` and touches the checkpoint.
5. Prints the answer and `ramo learn/<stamp> em <worktree>`.
6. Creates a reminder titled `Revisar as regras propostas` in the Reminders list `Tarefas Claude`, with the branch and the worktree in its body. It uses `osascript` and gives up after 10 seconds.

If the reminder fails, the command prints a warning to standard error and appends the reminder to `pending-reminders.txt`. The round still counts as done.

Each call to `claude` is killed after `--call-timeout-minutes`. The round then fails with `o claude passou de <time> sem terminar; processo encerrado` and writes no metrics row and no checkpoint.

The command never pushes, never opens a pull request and never merges.

## The state folder

| File | What it holds |
| --- | --- |
| `checkpoint`, `checkpoint-notes` | Empty files. Their modification time is the start of the last round of each pass |
| `metrics.tsv` | One row per round of the model |
| `input-<stamp>.txt`, `notes-input-<stamp>.txt` | The input given to the model |
| `output-<stamp>.json`, `notes-output-<stamp>.json` | The raw answer of `claude` |
| `pending-reminders.txt` | Reminders that could not be created, and the notice that the notes folder is not accessible: title and body, one per line |
| `launchd.log` | Output of the scheduled job, when it was installed by `rangers schedule install` |

## `metrics.tsv` columns

The file has no header. The columns, separated by tabs:

| Column | What it holds |
| --- | --- |
| `rodada` | The stamp of the round |
| `falas` | Lines given to the model |
| `segundos` | Duration of the model, in seconds |
| `turnos` | Turns the model took |
| `custo_usd` | Cost in dollars, rounded down to three decimals |
| `tokens_entrada` | Input tokens, counting cache writes and reads |
| `tokens_saida` | Output tokens |
| `passada` | `rules` or `notes`. A row without this column counts as `rules` |

`rangers learn metrics` prints the header, every row, the totals of time and cost with their averages, the token totals and one line of totals per pass. With no rows, it prints `nenhuma rodada registrada`.

## Schedule

`rangers schedule` manages a launchd job, `com.rangers.learn`, that runs `~/.local/bin/rangers learn` every 2 hours and once when it loads. The job passes no flags and sets only `HOME` and `PATH`, so it runs with the defaults above.

| Subcommand | What it does |
| --- | --- |
| `install` | Writes `~/Library/LaunchAgents/com.rangers.learn.plist`, creates the state folder and loads the job. `--binary` sets the program the job runs; the default is `~/.local/bin/rangers`. Prints `agendado: a cada 2 horas, log em <state>/launchd.log` |
| `uninstall` | Unloads the job and deletes the file. Prints `agendamento removido` |
| `status` (default) | Prints `estado`, `rodadas`, `último código de saída` and `intervalo` of the job, or `não agendado` |
| anything else | Returns `uso: rangers schedule install|uninstall|status` and exits with code `1` |

The job sets a `PATH` with `~/.local/bin`, `/opt/homebrew/bin`, `/usr/local/go/bin`, `/usr/local/bin`, `/usr/bin` and `/bin`, and writes standard output and standard error to `~/.local/state/rangers/learn/launchd.log`. `install` does not build the program.

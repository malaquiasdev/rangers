---
name: claude-sessions
description: Orchestrates Claude Code agents without tmux, as a background subagent (Agent tool) or a headless process (claude -p) with a log and --resume; includes the epic coordinator mode (worktree per card, review, QA, watchdog). Use when the human asks to delegate a task to another agent, to run an epic with several agents, to follow or correct a running agent; or when /claude-sessions runs.
---

# claude-sessions

A session the human opens by hand (claude.ai or phone) is started with `claude rc --permission-mode bypassPermissions --name <name>`. This skill is for when I delegate and coordinate from here.

Read the project's notes in the notes folder for local traps before you start.

tmux is out of the flow: a stuck Enter and an epic idle for hours with nobody seeing it.

## Model

- The strongest model only in the coordinator. Every agent runs on Sonnet (money code, front end, infrastructure, review, QA), with the model explicit in the call (`model: sonnet` / `--model sonnet`).
- Haiku only for a checklist or a mechanical task, always with `effort: high` (at the default it fails too often).
- An open session does not change model: recreate it with an empty context.

## Task up to ~1 h: background subagent

`Agent` with `run_in_background`, an explicit `model` and the complete task message (blocks below). The result arrives as a notification: do not poll. To correct or continue: `SendMessage` to the agent id.

## Long task, or one that must outlive the session: `claude -p`

```bash
claude -p --model sonnet --output-format json "<task message>" > <log>.json 2> <log>.err   # Bash with run_in_background
jq -r .session_id <log>.json > <name>.session_id
claude -p --resume "$(cat <name>.session_id)" --model sonnet --output-format json "<correction>" > <log2>.json
```

- One log per process, in a file. Keep the `session_id` as soon as the output exists.
- A correction is `--resume <id> "<msg>"`: an agent that stopped does not wake up by itself.
- Card finished: end the process right away (`kill` the pid), never leave it open.
- Before ending an agent that worked on a card, ask it to save its notes to the notes folder and wait for the paths.

## Talking to the human while agents run

1. **On dispatch:** one line. What was dispatched and when it comes back ("3 agents reviewing 54 issues, ~5 min").
2. **While it runs:** nothing. At most one partial, one line, and only if something changed that the human needs to know now ("batch 2 of 3: 2 security issues stopped"). Never "still running", never explain the process.
3. **Before the end:** confirm with a command (`gh`, the tracker, the cloud CLI) every finding that will become a decision. A finding not confirmed goes out as "not confirmed" or stays out.
4. **At the end:** one message. First line = the most important decision (yes/no). Up to 5 items, each with a link. The other decisions go in a numbered "for later" list, with no answer asked now. Up to ~560 characters. The detail goes to the notes folder and the human gets the link.

## Watchdog

A watchdog runs in the background, checks every 5 minutes and alerts after 20 minutes without progress (no new commit, no new log line). An agent that says "waiting for CI/the review bot" and goes idle: tell it to follow in the background (`gh pr checks --watch`, Monitor) and report only at the end.

## Coordinator mode (epic with several agents)

I coordinate and do not write code. The contract is approved first, one sub-issue per PR.

1. **Worktree per card:** `git -C <repo> worktree add -b <branch> <repo>-worktrees/<card> origin/main`. Use the branch pattern of the repository, not the tracker's default name if the repository refuses it.
2. **One agent per card:** its name is the card id, running in its worktree.
3. **Task message** (always with these blocks):
   - who it is (agent X of the epic) and the card to implement; read the card, the contract and the design;
   - which worktree and branch it is on; what NOT to touch (files of another agent in the same repository);
   - rules: nothing in production, no merge, no deploy, zero inline comments, `null` and not `undefined`, PR in the repository template, move the card (In Progress to Code Review), talk only to me;
   - the end: send me the link, the fakes and the pending items in up to 5 lines and WAIT for my review; only after I approve, request the automated review on the PR, fix until it is clean, **at most 3 rounds**;
   - "confirm in one line and start".
4. **Size and commits:** a PR above 500 lines of code (tests and fixtures do not count) becomes a stack with `gh stack`, each PR reviewed on its own. Inside a PR, one commit per logical step (with its test), each commit compiles and passes alone; a review fix goes in its own commit citing the finding. Put this in the author's task message.
   1. **When to call the review bot:** only on a PR about money, a contract, infrastructure, permissions or new logic above ~50 lines. A PR of 1 or 2 lines, text, only tests or a small screen adjustment: only me and the human. **The bot never blocks the PR:** if it publishes findings, the author adjusts; if it cannot (publishes nothing, does not answer in ~30 minutes, or errors), the PR goes straight to the human, with no insisting and no fixing the bot. Wait for its verdict in the background, never in a shell loop. A verdict with no detail means the finding is unproven: ask the author to reproduce it.
5. **Independent review (1st barrier, before the bot):** for a big PR or one about money, a Sonnet subagent (or a `claude -p` process) with `/code-review high`, read only, focused on paying twice, not paying without notice, a wrong status and fakes that hide a bug. The findings go back to the author with "confirm each one in the code before touching it". Order: me, the bot (at most 3 rounds per PR), the human.
6. **QA:** a separate agent `qa-<feature>`, in the dev environment, through the screen (Playwright, `channel: 'chrome'`) and the database (read only). Complete evidence in the notes folder at `<context>/<slug>/evidence/<qa>-<date>/`: unmasked screenshots plus the database query and result, per step, with the script in Gherkin (`script.feature`, one Scenario per step, `@pending` on what did not run).
7. **An agent's claim about state (card, PR, deploy) is confirmed before it is passed on** (`gh pr view`, the tracker). Agents are wrong with confidence.
8. **Scoreboard for the human:** the first line is the next action; one decision per message; when asking about pending items, repeat the numbered list.

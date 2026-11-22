# How the learning loop works

The agents get better when the human's corrections become written rules. The learning loop does that without the human asking: it reads the sessions, proposes rules and waits for the human to accept them. The facts of each step are in the [`rangers learn` reference](../referencia/comando-learn.md).

## Corrections live in the transcripts

When the human corrects an agent ("that was not a review, do not rewrite it"), the correction is in the Claude Code transcript of that session. If nobody writes it down, the next session makes the same mistake. The learning agent reads what the human typed since the last round and keeps what is worth keeping six months from now: a correction, a repeated mistake, a rule that was missing. Each rule cites its origin, the session and the human's own words, so the human can judge it.

A rule for one agent goes into that agent's file. A rule for every agent goes into [`rules/`](../../rules/).

## A program decides when to spend, not the model

Calling a large model every two hours on every session would cost money for nothing most of the time. So the decision to run belongs to a small Go program with no model, and the model only starts when three gates agree:

- enough time has passed since the last round;
- the human typed enough new lines to be worth reading;
- the day still has room under the dollar cap.

Each round also has its own dollar cap, which never goes above what is left of the day. A round stopped by a gate leaves its lines for the next one, so nothing is lost by waiting.

launchd runs the program every two hours. A run that stops at a gate calls no model and costs nothing.

## Rules are proposed, never applied

The rules pass works in a new worktree, on a local branch `learn/<stamp>`, and commits there. It never pushes, never opens a pull request and never merges. A reminder tells the human that a branch is waiting. The human reads the diff and keeps all, some or none of the rules (see [how to review the proposed rules](../como-fazer/revisar-as-regras-propostas.md)).

This keeps the human in charge of what the agents become. A wrong rule would change every future session, so it costs more than a missed one.

## Two passes, two jobs

The same agent runs twice per round, with different inputs and different places to write:

- The **rules pass** reads only what the human typed, because rules come from corrections, and commits on a local branch of this repository, in its own worktree.
- The **notes pass** also reads what the assistant answered, because decisions and results are often in the answer, and writes notes in a separate notes folder. It never writes a rule and never touches this repository.

Each pass has its own checkpoint, so one can run while the other waits at a gate. The notes pass uses a smaller model and a size limit on its input, because it reads much more text. It can be turned off with `--skip-notes`.

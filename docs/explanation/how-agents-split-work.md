# How the agents split the work

Each agent has one role and a clear limit on what it does not do. The split follows a simple idea: whoever writes the code does not review it, whoever reviews does not edit, and no agent takes anything to production without the human's explicit yes.

Only the gate agent can merge, and only through `rangers gate --merge`, when its task says to merge. The program refuses a production pull request, requires the human's yes bound to the head commit and requires a green CI: at least one check, each one passed or skipped. The gate never merges by hand. The coordinator, the author and the swarm are forbidden to merge, and no agent deploys.

The roles are in the files of [`agents/`](../../agents/). The `description` of each one starts with the role in English: Coordinator, Author, Reviewer and so on. The file format is in the [agent file reference](../reference/agent-file.md).

## The path of an activity

1. The human types `/rangers` with an issue or a short description (see [how to start an activity](../how-to/start-activity.md)). The skill hands the text to the coordinator as it is.
2. The **coordinator** finds the repository, reads the issue and refines the activity with the human using the `grilling` skill, one question at a time. It picks the pstack playbook that fits and writes the plan for a reader who is not technical: a three-line summary, the scenarios as "given, when, then", a flowchart when there is a decision or a flow, and the technical detail in a closed block at the end.
3. The **courier** publishes the plan as a private page and sends the link with the question. The question shows the summary and a 12-character hash of the plan and asks for one yes. Without the yes, nothing starts.
4. After the yes, the coordinator starts the **author** for one unit, or the **swarm** when the plan has several units that touch different files. The author works in its own worktree, writes the failing test before the code, makes small commits and opens its own pull request.
5. When the pull request is open, the coordinator calls the other agents only when they are needed (see below). On a Linear issue it moves the card to code review; on a GitHub issue, the pull request closes the issue. Then it stops and tells the human what is left: read the pull request, merge it and deploy it.

## Who comes in only when needed

| Role | What it does | When the coordinator calls it |
| --- | --- | --- |
| Explorer | Answers questions about code and data without changing anything, with `path:line` for each claim. It has two modes: `find`, to locate something, and `explain`, to say how it works and what breaks if it changes | Before planning, when it needs to understand the code |
| Reviewer | Reads the pull request in a fresh session, without the author's context. Each finding needs proof and a severity: BLOQUEIA, MÉDIO or NIT. Only a BLOQUEIA with proof blocks the merge. At most two rounds | Only when the human asks for a review. By default, the CI and the review bot of the repository are the review |
| Adversarial reviewer | Starts from an empty context and tries to break the change with concrete inputs. Only a finding proved by running something counts | Only for a unit marked high risk: money, production, security or a guard |
| QA | Tests in dev what was merged, through the entry point of the service, and records complete evidence. It only reads the database, inside a read-only transaction | Only when the plan has scenarios that need a real test in dev after the merge |

## The roles outside the path

The **gate** runs the `rangers gate` program, which checks the human's yes, the CI and production before a merge. The program decides, not the model. The gate asks the human with `rangers gate ask`, which prints a question tied to the head commit, and it always stops at production. Its checks and exit codes are in the [reference](../reference/gate-command.md).

The **documentation** agent organizes the `docs/` folder of a repository by the Diataxis method. Only the human calls it, and it never opens a pull request or changes code.

The **learning** agent turns the human's corrections into proposed rules. The `rangers learn` program runs it on a schedule; no agent calls it. It commits on a local branch and never pushes or opens a pull request. How it works is in [how the learning loop works](how-learning-works.md).

## Who tells the human

The human follows the agents from the Claude app. Messages go out through the coordinator and the courier:

- The coordinator decides when the human must know: the plan waits for the yes, a pull request opened, an agent stopped, got blocked or finished without a result, the reviewer found a BLOQUEIA, or something is ready to test.
- The courier sends the push notification. When the push asks the human to answer, it also creates a reminder, because a push does not show while the app is open. When the human answers, the coordinator marks the reminder as done.
- If the courier is not available, the coordinator starts its report with a `PUSH:` line, and the session that called it sends the push. The `/rangers` skill does the same with an `ARTIFACT:` line: it publishes that file as a page, exactly as it is.

Agents with limited tools, such as the reviewers and the explorer, do not notify. They answer the coordinator.

## Why the limits are written as prohibitions

Each file has a "What you do not do" section. It exists because the most expensive mistake of an agent is doing more than it was asked: a merge nobody approved, an edit during a review, a guess without proof. Writing the limit in the agent itself keeps the rule close to whoever could break it.

## Models and effort

Each agent declares its model and effort in its own file. Roles that need broad judgment, such as coordinating, learning and documenting, run on a larger model. Execution roles, such as author, swarm, reviewers and QA, run on a middle model. The gate and the courier run on the smallest one, because they run a program or deliver a text without changing it. The exact values are in the files of `agents/`.

## The voice of each agent

Each file has a `Character:` line that gives the agent its own voice. The voice is for the conversation between agents. Most files forbid the character name in what the agent writes, such as code, commits, pull requests and comments. The repository checks refuse those names outside `agents/`, `skills/` and the README, where the cast is presented, and the commit hook refuses them in commit messages.

# Rangers documentation

Rangers is a repository of agents and rules for Claude Code, on top of open source projects that already coordinate agents. A human starts an activity with `/rangers`, a coordinator plans it, and other agents build, review and test it until a pull request is ready. The pages are split by what you need right now.

## To learn

- [Set up the machine for the first time](tutorials/prepare-the-machine.md): install the pieces, set up pstack and see the agents in Claude Code.

## To get a task done

- [How to connect the agents, the rules and the skill to Claude Code](how-to/connect-agents.md)
- [How to install the skills the agents use](how-to/install-skills.md)
- [How to start an activity](how-to/start-activity.md)
- [How to use Rangers from a tablet or a phone](how-to/use-from-a-tablet.md)
- [How to run the test with a real activity](how-to/run-test.md)
- [How to schedule the learning job](how-to/schedule-learning.md)
- [How to review the rules a learning round proposed](how-to/review-proposed-rules.md)
- [How to run the checks before a commit](how-to/run-checks.md)
- [How to check the documentation links](how-to/check-doc-links.md)
- [How to uninstall](how-to/uninstall.md)

## To look something up

- [Dependencies](reference/dependencies.md): format of `deps.toml`, what `rangers install` covers, install and remove commands, what the machine needs.
- [The `rangers install` command](reference/install-command.md): the check of each piece, flags and exit codes.
- [Scripts](reference/scripts.md): what each script in `scripts/` does and returns.
- [The `rangers learn` command](reference/learn-command.md): flags, passes, gates, state folder and schedule.
- [The `rangers gate` command](reference/gate-command.md): checks, the approval bound to the head commit, the production stop, output and exit codes.
- [The `stop-push-check` hook](reference/stop-push-check.md): the Stop hook that requires a push notification and a reminder, its settings snippet and what it blocks.
- [The `complete-reminders` hook](reference/complete-reminders.md): the UserPromptSubmit hook that completes the waiting reminders, its settings snippet and its limits.
- [Checks and release](reference/checks-and-release.md): `make` targets, Git hooks, what the checks look at, the privacy file and the release workflow.
- [Agent file](reference/agent-file.md): header fields and body sections.
- [Test plan](reference/test-plan.md): measures and decision criteria.

## To understand

- [Why a thin layer on top of open source](explanation/why-a-thin-layer.md)
- [How the agents split the work](explanation/how-agents-split-work.md)
- [How the learning loop works](explanation/how-learning-works.md)

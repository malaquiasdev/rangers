# Set up the machine for the first time

In this tutorial you will get a machine ready to use Rangers: install the open source pieces, set up pstack and see the agents of this repository in the Claude Code list.

## Before you start

You need:

- a clone of this repository;
- Claude Code installed, with the `claude` command in the terminal;
- Homebrew;
- a container runtime that answers the `docker` command. Podman works as long as `docker` calls Podman (for example, a `docker` link that points to `podman`) and the Podman machine is up with `podman machine start`.

Open a terminal at the root of the clone. Every command below runs from there.

## 1. See what is already there

```sh
scripts/verify.sh
```

You will see one line per piece, with its version next to it. The ones not installed yet show `FALTANDO`. The last line says how many items are missing, or `Tudo no lugar.`.

Do not worry about the `FALTANDO` lines: the next step takes care of them.

## 2. Install what is missing

```sh
scripts/install.sh
```

The script installs only what was missing and, at the end, runs the check again. Look at the last line: now it says `Tudo no lugar.`.

If it still says `item(ns) faltando`, look in the lines above for the piece marked `FALTANDO` or `SEM RESPOSTA`. The most common one is the `docker` line: the script does not start the container runtime. Start Podman with `podman machine start` and run `scripts/verify.sh` again.

## 3. Set up pstack

Close Claude Code, if it is open, and open a new session. In the session, run:

```
/pstack:setup-pstack
```

pstack finds the models you have and writes a rule with the model of each role. Follow what it asks. You only do this once.

## 4. Connect the agents and the skill

Back in the terminal, at the root of the clone:

```sh
mkdir -p ~/.claude/agents ~/.claude/skills
ln -s "$PWD"/agents/*.md ~/.claude/agents/
ln -s "$PWD/skills/rangers" ~/.claude/skills/rangers
```

Open another new Claude Code session and run:

```
/agents
```

The agents of this repository are in the list, each with its role in the description: Coordinator, Author, Reviewer and the others. Type `/rangers` and notice that the skill shows up too.

## What you did

The machine has the open source pieces installed, pstack set up, and the agents and the `/rangers` skill connected. The agents also call skills from other repositories, such as `grilling`; install them with [how to install the skills](../how-to/install-skills.md). Then start a real activity: see [how to start an activity](../how-to/start-activity.md).

To understand why the repository is so small, read [why a thin layer](../explanation/why-a-thin-layer.md).

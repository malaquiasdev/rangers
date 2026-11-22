# Set up the machine for the first time

In this tutorial you will get a machine ready to use Rangers: install the open source pieces, set up pstack and see the agents of this repository in the Claude Code list.

## Before you start

You need:

- a clone of this repository;
- Claude Code installed, with the `claude` command in the terminal;
- Go 1.26, to build the `rangers` program;
- Git and Podman, because `rangers install` clones the skills with Git and pulls an image with Podman;
- `~/.local/bin` in your `PATH`, because step 1 builds the program there;
- Homebrew;
- a container runtime that answers the `docker` command. Podman works as long as `docker` calls Podman (for example, a `docker` link that points to `podman`) and the Podman machine is up with `podman machine start`.

Open a terminal at the root of the clone. Every command below runs from there.

## 1. See what is already there

```sh
mkdir -p ~/.local/bin
go build -o ~/.local/bin/rangers ./cmd/rangers
rangers install --check
```

You will see one line per piece, with its version next to it. The ones not installed yet show `MISSING`. The last line says how many items are missing, or `All in place.`.

Do not worry about the `MISSING` lines: the next step takes care of them.

## 2. Install what is missing

```sh
rangers install
```

The command installs only what was missing and links the agents, the rules and the skill into `~/.claude`. Then run `rangers install --check` again. Look at the last line: now it says `All in place.`.

If it still says `item(s) missing`, look in the lines above for the piece marked `MISSING` or `NOT RESPONDING`. The most common one is the `docker` line: the command does not start the container runtime. Start Podman with `podman machine start` and run `rangers install --check` again.

## 3. Set up pstack

Close Claude Code, if it is open, and open a new session. In the session, run:

```
/pstack:setup-pstack
```

pstack finds the models you have and writes a rule with the model of each role. Follow what it asks. You only do this once.

## 4. See the agents and the skill

`rangers install` already linked them. Open another new Claude Code session and run:

```
/agents
```

The agents of this repository are in the list, each with its role in the description: Coordinator, Author, Reviewer and the others. Type `/rangers` and notice that the skill shows up too.

## What you did

The machine has the open source pieces and the skills the agents call installed, pstack set up, and the agents and the `/rangers` skill connected. Then start a real activity: see [how to start an activity](../how-to/start-activity.md).

To understand why the repository is so small, read [why a thin layer](../explanation/why-a-thin-layer.md).

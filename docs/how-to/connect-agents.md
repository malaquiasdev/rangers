# How to connect the agents, the rules and the skill to Claude Code

Claude Code loads the user's subagents from `~/.claude/agents/`, the user's rules from `~/.claude/rules/` and the user's skills from `~/.claude/skills/`. `rangers install` links the files of this repository into these folders. This guide does the same by hand, so a change in the repository applies in the next session without copying anything.

## Connect the agents

1. From the root of the repository, create the folder if it does not exist, and the links:

   ```sh
   mkdir -p ~/.claude/agents
   ln -s "$PWD"/agents/*.md ~/.claude/agents/
   ```

   If `ln` says `File exists` for a file, an agent with that name is already in the folder. Check what it is with `ls -l ~/.claude/agents` before you decide. Do not delete a regular file in that folder without knowing where it came from.

2. Open a new Claude Code session. Agents load when the session starts.

3. Run `/agents` in the session and check that the agents are in the list.

## Connect the rules

The files in [`rules/`](../../rules/) apply to every agent. From the root of the repository:

```sh
mkdir -p ~/.claude/rules
ln -s "$PWD"/rules/*.md ~/.claude/rules/
```

As with the agents, `File exists` means a rule with that name is already there. Rules load when a session starts.

## Connect the `/rangers` skill

The skill in [`skills/rangers/`](../../skills/rangers/) starts an activity with the coordinator. From the root of the repository:

```sh
mkdir -p ~/.claude/skills
ln -s "$PWD/skills/rangers" ~/.claude/skills/rangers
```

If `~/.claude/skills/rangers` already exists, `ln` puts the link inside it instead. Check it with `ls -l ~/.claude/skills/rangers` first.

Open a new session. Typing `/rangers` shows the skill. How to use it is in [how to start an activity](start-activity.md).

## Check what is connected

```sh
ls -l ~/.claude/agents ~/.claude/rules ~/.claude/skills
```

The files from this repository show up as links (`->`) that point to the clone.

## Disconnect the agents

From the root of the repository, delete only the links that point here, in the agents, rules and skills folders:

```sh
find ~/.claude/agents ~/.claude/rules ~/.claude/skills -maxdepth 1 -type l -lname "$PWD/*" -print -delete
```

The command prints each link it deleted. Regular files and links to other places stay where they are.

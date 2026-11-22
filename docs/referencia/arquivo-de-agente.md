# Agent file

Each agent is a Markdown file in [`agents/`](../../agents/), with a YAML header and a body of instructions in English. Claude Code loads these files as subagents when they are in the user's agents folder, `~/.claude/agents/` (see [how to connect the agents](../como-fazer/ligar-os-agentes.md)).

## Header

| Field | Required | Values | What it sets |
| --- | --- | --- | --- |
| `name` | Yes | The file name without `.md` | How the agent is called in Claude Code |
| `description` | Yes | Starts with the role in English and a period, then one sentence | When to use the agent. Claude Code shows this text in the list of agents |
| `model` | Yes | `opus`, `sonnet`, `haiku` | The model the agent runs on |
| `effort` | Yes | `low`, `medium`, `high`, `xhigh`, `max` | The reasoning effort |
| `tools` | No | Comma-separated list, such as `Read, Grep, Glob, Bash` | The only tools the agent can use. Without the field, the agent gets all of them |

The agents that only read or only run one command declare `tools`. The ones that write files do not.

`scripts/checar.sh` refuses a file whose `name`, `model`, `effort` or `description` breaks the rules above (see the [checks reference](checagens-e-release.md#what-scriptschecarsh-checks)).

`rangers learn` finds the learning agent by its `description`, which must start with `Learning.` (see the [`rangers learn` reference](comando-learn.md#the-two-passes)).

## Body

The body follows the same order in every file:

1. A `Character:` line with the voice of the agent. The checks require it.
2. A short paragraph that states the role.
3. `## What you do`, with the steps of the role.
4. `## What you do not do`, with the limits of the role.

Some files have more sections. The coordinator has `## Who you call and when`, `## How you talk to the human` and `## Push notifications`. The learning agent has `## Notes pass`.

## Rules that repeat

Each file carries its own rules. Some appear in several of them:

- the name of the character does not appear in code, commits, pull requests, comments or documents;
- the answer to whoever called the agent has at most 560 characters;
- production waits for the human's explicit yes.

A rule only applies to the agent whose file carries it. Rules for every agent go in [`rules/`](../../rules/).

## Where to see each agent

The files in `agents/` are the source. To list the role, model and effort of all of them at once, from the root of the repository:

```sh
grep -H -E '^(description|model|effort|tools):' agents/*.md
```

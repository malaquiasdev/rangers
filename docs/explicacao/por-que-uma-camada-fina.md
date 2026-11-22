# Why a thin layer on top of open source

This repository uses open source projects that already coordinate Claude Code agents, and writes only what none of them covers.

## What the open source projects do

The installed plugins and tools do most of the work. Their names show up across the documentation:

| Piece | What it does for Rangers |
| --- | --- |
| pstack | Claude Code plugin with playbooks by kind of task (bug, feature and others), review by another agent, verification and reflection. The agents here call its playbooks and skills |
| i-have-adhd | Plugin that changes the answer format: action first, numbered steps. Turned on with `/i-have-adhd` |
| `grilling` | Skill that refines an activity one question at a time |
| OpenSpec | Contract and spec before code. Optional |
| container-use | An isolated environment per agent, with no power to merge |

The exact list, with version and source, is in [`deps.toml`](../../deps.toml). What each field means is in the [dependencies reference](../referencia/dependencias.md).

## What stays in this repository

Little is left, on purpose:

- the agents, which tell each open source project how to work the way the owner of the repository wants (see [how the agents split the work](como-os-agentes-dividem-o-trabalho.md));
- the `/rangers` skill, the single entry point that hands an activity to the coordinator;
- the rules for every agent, in [`rules/`](../../rules/);
- the list of what is installed and the scripts that install and check it;
- the `rangers learn` program, which turns the human's corrections into proposed rules (see [how the learning loop works](como-o-aprendizado-funciona.md));
- the local checks and the release workflow;
- the test plan that decides whether the idea works.

The upside is little code to keep. The cost is depending on the pace and the choices of other projects. That is why each dependency has a pinned version or commit in `deps.toml`, and each one has a remove command in [how to uninstall](../como-fazer/desfazer-a-instalacao.md).

## The checks run on the machine

The checks run in Git hooks and `make check`, not in GitHub Actions. The only workflow is release-please. This keeps the Actions quota for the release and gives the same answer before a commit as before a push. The list is in the [checks and release reference](../referencia/checagens-e-release.md).

## The test decides

The test answers one question: does a thin layer on top of open source projects take an activity to a pull request with little friction for a human who reads short answers and does not want to open files?

The test uses a real work activity, small, the size of one pull request. A made-up activity does not count, because it does not show the real friction. The measures and what each result decides are in the [test plan reference](../referencia/plano-de-teste.md). The steps are in [how to run the test](../como-fazer/rodar-o-teste.md).

## What does not exist yet

The gate agent runs a `rangers gate` command that checks the human's yes, the CI and production before a merge. That command does not exist yet: the `rangers` program has only `learn`. The coordinator does not call the gate today, and it stops when the pull request is open.

The gate comes in only if the test shows it is missing.

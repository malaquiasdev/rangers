# Rangers documentation

Rangers is a repository of agents and rules for Claude Code, on top of open source projects that already coordinate agents. A human starts an activity with `/rangers`, a coordinator plans it, and other agents build, review and test it until a pull request is ready. The pages are split by what you need right now.

## To learn

- [Set up the machine for the first time](tutorial/preparar-a-maquina.md): install the pieces, set up pstack and see the agents in Claude Code.

## To get a task done

- [How to connect the agents and the skill to Claude Code](como-fazer/ligar-os-agentes.md)
- [How to install the skills the agents use](como-fazer/instalar-as-skills.md)
- [How to start an activity](como-fazer/iniciar-uma-atividade.md)
- [How to run the test with a real activity](como-fazer/rodar-o-teste.md)
- [How to schedule the learning job](como-fazer/agendar-o-aprendizado.md)
- [How to review the rules a learning round proposed](como-fazer/revisar-as-regras-propostas.md)
- [How to run the checks before a commit](como-fazer/rodar-as-checagens.md)
- [How to check the documentation links](como-fazer/conferir-links-da-documentacao.md)
- [How to uninstall](como-fazer/desfazer-a-instalacao.md)

## To look something up

- [Dependencies](referencia/dependencias.md): format of `deps.toml`, what the scripts cover, install and remove commands, what the machine needs.
- [Scripts](referencia/scripts.md): what each script in `scripts/` does and returns.
- [The `rangers learn` command](referencia/comando-learn.md): flags, passes, gates, state folder and schedule.
- [Checks and release](referencia/checagens-e-release.md): `make` targets, Git hooks, what the checks look at, the privacy file and the release workflow.
- [Agent file](referencia/arquivo-de-agente.md): header fields and body sections.
- [Test plan](referencia/plano-de-teste.md): measures and decision criteria.

## To understand

- [Why a thin layer on top of open source](explicacao/por-que-uma-camada-fina.md)
- [How the agents split the work](explicacao/como-os-agentes-dividem-o-trabalho.md)
- [How the learning loop works](explicacao/como-o-aprendizado-funciona.md)

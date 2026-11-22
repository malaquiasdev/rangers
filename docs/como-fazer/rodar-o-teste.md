# How to run the test with a real activity

This guide runs the test that decides whether the thin layer works. The measures and what each result decides are in the [test plan reference](../referencia/plano-de-teste.md).

Before you start, pick a real, small activity that fits in one pull request.

1. Run `scripts/conferir.sh` and check that the last line says `Tudo no lugar.`. If something is missing, run `scripts/instalar.sh` (see the [scripts reference](../referencia/scripts.md)).
2. If you just installed pstack, restart Claude Code and run `/pstack:setup-pstack` once.
3. Check that the agents and the skill are connected with `ls -l ~/.claude/agents ~/.claude/skills`. If they are not, follow [how to connect the agents and the skill](ligar-os-agentes.md).
4. Open a new Claude Code session in the repository of the activity, not in this one.
5. Run `/i-have-adhd` to turn on the short answer format.
6. Start the activity with `/rangers`, as in [how to start an activity](iniciar-uma-atividade.md), and follow it until the pull request is ready.
7. Come back to this repository and record the measures in `resultados/YYYY-MM-DD.md`, at the root, with the date of the test. Use the measures table of the reference as a model. Create the `resultados/` folder if it does not exist yet.

To know what to do with the result, use the criteria table of the [test plan reference](../referencia/plano-de-teste.md#decision-criteria).

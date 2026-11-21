# Como rodar o teste com uma atividade real

Este guia mostra como rodar o teste que decide se a camada fina funciona. As medidas e o que cada resultado decide estão na [referência do plano de teste](../referencia/plano-de-teste.md).

Antes de começar, escolha uma atividade real e pequena, que caiba numa PR.

1. Rode `scripts/conferir.sh` e confira se a última linha diz `Tudo no lugar.`. Se faltar algo, rode `scripts/instalar.sh` (veja a [referência dos scripts](../referencia/scripts.md)).
2. Se acabou de instalar o pstack, reinicie o Claude Code e rode `/pstack:setup-pstack` uma vez.
3. Confira se os agentes estão ligados com `ls -l ~/.claude/agents`. Se não estiverem, siga [como ligar os agentes](ligar-os-agentes.md).
4. Abra uma sessão nova do Claude Code no repositório da atividade, não neste.
5. Rode `/i-have-adhd` e refine a atividade com a skill `grilling`.
6. Toque a atividade com o roteiro do pstack que couber, até a PR ficar pronta.
7. Volte a este repositório e anote as medidas em `resultados/AAAA-MM-DD.md`, na raiz, com a data do teste. Use a tabela de medidas da referência como modelo. Crie a pasta `resultados/` se ela ainda não existir.

Para saber o que fazer com o resultado, use a tabela de critérios da [referência do plano de teste](../referencia/plano-de-teste.md#critérios-de-decisão).

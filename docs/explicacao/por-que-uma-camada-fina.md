# Por que uma camada fina em cima de open source

O Jaiminho começou como uma plataforma própria, a `jaiminho-suite`. Ela coordenava agentes do Claude Code com código escrito do zero. A suíte foi arquivada e este repositório tomou o lugar dela com outra aposta: usar projetos open source que já coordenam agentes e escrever só o que nenhum deles cobre.

## O que fica com os projetos open source

Os plugins e ferramentas instalados fazem o trabalho pesado. Os nomes aparecem em toda a documentação:

| Peça | O que faz para o Jaiminho |
| --- | --- |
| pstack | Plugin do Claude Code com roteiros prontos por tipo de tarefa (bug, feature e outros), revisão por outro agente, verificação e reflexão. Os agentes daqui chamam os roteiros e as skills dele |
| i-have-adhd | Plugin que muda o formato da resposta para quem tem TDAH: ação primeiro, passos numerados. Liga com `/i-have-adhd` |
| `grilling` | Skill que refina uma atividade com uma pergunta por vez |
| OpenSpec | Contrato e spec antes de escrever código |
| container-use | Ambiente isolado por agente, sem poder de merge |
| ccmanager | Avisa quando um agente muda de estado |
| basic-memory | Memória em Markdown entre sessões |

A lista exata, com versão e origem, está em [`deps.toml`](../../deps.toml). O que cada campo quer dizer está na [referência das dependências](../referencia/dependencias.md).

## O que fica neste repositório

Sobra pouco, e de propósito:

- os agentes, que dizem a cada projeto open source como trabalhar do jeito do dono do repo (veja [como os agentes dividem o trabalho](como-os-agentes-dividem-o-trabalho.md));
- a lista do que está instalado e os scripts que instalam e conferem;
- o plano de teste que decide se a ideia funciona;
- as regras próprias que nenhum projeto cobre, que só entram depois do teste.

A vantagem é manter pouco código. O custo é depender do ritmo e das decisões de outros projetos. Por isso cada dependência tem versão ou commit fixo em `deps.toml`, e cada uma tem um comando de remoção em [como desfazer a instalação](../como-fazer/desfazer-a-instalacao.md).

## O teste decide

A pergunta que o teste responde é esta: uma camada fina em cima de projetos open source resolve o que a plataforma antiga resolvia, com menos atrito para quem usa só iPad e tem TDAH?

O teste usa uma atividade real de trabalho, pequena, do tamanho de uma PR. Atividade inventada não vale, porque não mostra o atrito de verdade. As medidas e o que cada resultado decide estão na [referência do plano de teste](../referencia/plano-de-teste.md). O passo a passo está em [como rodar o teste](../como-fazer/rodar-o-teste.md).

Se o teste falhar em duas ou mais medidas, a decisão volta para o desenho, e a plataforma antiga pode voltar a valer a pena. Por isso ela continua guardada, e não apagada.

## O que ainda não tem dono em open source

Três peças não existem prontas em nenhum projeto:

1. O sim do humano com prova no transcript da sessão.
2. O portão que confere a CI antes do merge.
3. As regras de resposta curta para TDAH.

Elas só entram se o teste mostrar que fazem falta. Os agentes já citam algumas delas. O agente de portão chama um comando `jaiminho gate`, e o agente de aprendizado grava regras numa pasta `rules/`. Nenhum dos dois existe ainda neste repositório.

O mesmo vale para o instalador em Go. Até ele existir, os scripts em `scripts/` fazem esse papel (veja a [referência dos scripts](../referencia/scripts.md)).

## O que ficou de fora

| Projeto | Por que ficou de fora |
| --- | --- |
| orchestrate (repositório de plugins do Cursor) | O script depende dos agentes de nuvem do Cursor, e o marketplace do Claude Code só lista o pstack |
| agent-teams (wshobson/agents) | Recusado pelo dono do repo em 21/11/2026 |
| claude-flow (ruflo) | Um estudo de 20/11/2026 descartou: mais complexidade do que ganho |

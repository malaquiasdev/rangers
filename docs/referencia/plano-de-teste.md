# Plano de teste

As medidas e os critérios que decidem se a camada fina funciona. Por que existe o teste está em [por que uma camada fina](../explicacao/por-que-uma-camada-fina.md#o-teste-decide). O passo a passo está em [como rodar o teste](../como-fazer/rodar-o-teste.md).

## A atividade

- Real, de trabalho, escolhida pelo dono do repo.
- Pequena: cabe numa PR.
- O contexto dela fica no repositório da atividade, não neste.

## Medidas

| Medida | Como anotar | Passou se |
| --- | --- | --- |
| Quantos "sim" o humano precisou dar | Contar na sessão | Até 2 |
| Viu os agentes rodando no app | Sim ou não | Sim |
| Leu o texto da sessão sem abrir outro arquivo | Sim ou não, e onde travou | Sim |
| Tempo da atividade até a PR pronta | Minutos | Sem meta fixa; anotar |
| Gasto da cota | Percentual antes e depois | Anotar |
| Algo saiu errado e ninguém avisou | Descrever | Nenhum caso |

## Onde anotar

Em `resultados/AAAA-MM-DD.md`, na raiz deste repo, com a data do teste. A pasta ainda não existe; o primeiro teste a cria.

## Critérios de decisão

| Resultado | Condição | O que fazer |
| --- | --- | --- |
| Funciona | Passou nas três primeiras medidas e nenhum erro ficou sem aviso | Acrescentar só o que faltou |
| Funciona em parte | Uma medida falhou | Anotar qual e corrigir só ela, com um teste novo |
| Não funciona | Duas ou mais medidas falharam | Voltar ao desenho e decidir se vale a plataforma antiga |

## Depois do teste

Se funcionou, decidir se entram as três peças sem dono em open source: o sim com prova, o portão e as regras de resposta curta. Cada uma só entra se o teste mostrou que fez falta.

# Plano de teste

Pergunta: uma camada fina em cima de projetos open source resolve o que a plataforma antiga resolvia, com menos atrito para quem usa só iPad e tem TDAH?

## A atividade

Uma atividade real de trabalho, escolhida pelo Mateus, de tamanho pequeno (uma PR). Não vale atividade inventada. O contexto da atividade fica no repositório dela, não aqui.

## Como rodar

1. Preparar: `scripts/conferir.sh` sem peça faltando, Claude Code reiniciado, `/pstack:setup-pstack` feito.
2. Abrir uma sessão nova no repositório da atividade.
3. Usar `/i-have-adhd` e refinar a atividade com a skill `grilling`.
4. Tocar a atividade com o roteiro do pstack que couber, até a PR ficar pronta.
5. Anotar as medidas abaixo em `resultados/AAAA-MM-DD.md`.

## O que medir

| Medida | Como anotar | Passou se |
| --- | --- | --- |
| Quantos "sim" o Mateus precisou dar | Contar na sessão | Até 2 |
| Viu os agentes rodando no app | Sim ou não | Sim |
| Leu o texto da sessão sem abrir outro arquivo | Sim ou não, e onde travou | Sim |
| Tempo da atividade até a PR pronta | Minutos | Sem comparação fixa; anotar |
| Gasto da cota | Percentual antes e depois | Anotar |
| Algo saiu errado e ninguém avisou | Descrever | Nenhum caso |

## O que decide

- **Funciona:** passou nas três primeiras medidas e nenhum erro ficou sem aviso. Então só se acrescenta o que faltou.
- **Funciona em parte:** uma medida falhou. Anotar qual e corrigir só ela, com um teste novo.
- **Não funciona:** duas ou mais falharam. Voltar ao desenho e decidir se vale a plataforma antiga.

## Depois do teste

Se funcionou, decidir se entram as três peças sem dono em open source: o sim com prova, o portão e as regras de resposta curta. Cada uma só entra se o teste mostrou que fez falta.

## Voltar atrás

Remover as peças pelos comandos de `DEPENDENCIAS.md`.

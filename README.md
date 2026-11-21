# jaiminho

Um repositório de agentes e regras em cima de projetos open source que já coordenam agentes. Substitui a plataforma `jaiminho-suite`, que foi arquivada.

## A ideia

O Jaiminho vira um repositório de agentes que usam projetos open source que já existem, em vez de reconstruir a plataforma. O que sobra do código da suíte antiga é só o necessário para conduzir essas dependências.

Os plugins instalados fazem o trabalho pesado. Este repo guarda:

- os agentes e as skills que chamam esses projetos do jeito que o Mateus trabalha;
- a lista exata do que está instalado e como instalar e remover cada peça;
- o plano de teste que decide se a ideia funciona;
- as regras próprias que nenhum projeto cobre (decididas depois do teste).

## O que tem aqui

| Caminho | Para quê |
| --- | --- |
| `agents/` | Um arquivo por agente, com nome, função, modelo e o projeto que ele usa |
| `deps.toml` | Cada dependência, com a versão e o commit fixos, a licença e a origem |
| `DEPENDENCIAS.md` | Como instalar e remover cada peça, até o instalador em Go existir |
| `TESTE.md` | Como testar com uma atividade real, o que medir e o que decide |
| `scripts/` | Instala e confere o que faltar, até o instalador em Go existir |

## Como usar

1. `scripts/conferir.sh` mostra o estado.
2. `scripts/instalar.sh` completa o que faltar.
3. Reinicie o Claude Code e rode `/pstack:setup-pstack` uma vez.
4. Siga o `TESTE.md`.

## Sem dono em open source

Três coisas não existem prontas e só entram depois do teste, se ele mostrar que fazem falta: o sim do humano com prova no transcript, o portão que confere a CI antes do merge e as regras de resposta curta para TDAH.

## Voltar atrás

`DEPENDENCIAS.md` traz o comando de remoção de cada peça. A plataforma antiga está arquivada em `~/.jaiminho-arquivado-20261121` e o repo `jaiminho-suite` continua no GitHub.

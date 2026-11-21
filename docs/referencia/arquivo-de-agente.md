# Arquivo de agente

Cada agente é um arquivo Markdown em [`agents/`](../../agents/), com um cabeçalho YAML e um corpo de instruções em inglês. O Claude Code lê esses arquivos como subagentes quando eles estão na pasta de agentes do usuário, `~/.claude/agents/` (veja [como ligar os agentes no Claude Code](../como-fazer/ligar-os-agentes.md)).

## Cabeçalho

| Campo | Obrigatório | Valores usados no repo | O que define |
| --- | --- | --- | --- |
| `name` | Sim | O nome do arquivo sem `.md` | Como o agente é chamado no Claude Code |
| `description` | Sim | Começa pelo papel em inglês, seguido de uma frase | Quando usar o agente. O Claude Code mostra esse texto na lista de agentes |
| `model` | Sim | `opus`, `sonnet`, `haiku` | O modelo em que o agente roda |
| `effort` | Sim | `high`, `medium` | O esforço de raciocínio |
| `tools` | Não | Lista separada por vírgula, como `Read, Grep, Glob, Bash` | As únicas ferramentas que o agente pode usar. Sem o campo, o agente recebe todas |

Os agentes que só leem ou só executam um comando declaram `tools`. Os que escrevem arquivos não declaram.

## Corpo

O corpo segue a mesma ordem em todos os arquivos:

1. Uma linha `Character:`, com a voz do agente.
2. Um parágrafo curto que diz o papel.
3. `## What you do`, com os passos do papel.
4. `## What you do not do`, com os limites do papel.

O coordenador tem ainda duas seções: `## How you talk to the human`, com as regras de mensagem para o humano, e `## Push notifications`, com os momentos em que ele avisa o humano por notificação.

## Regras que se repetem

Cada arquivo traz as próprias regras. Algumas aparecem em vários deles:

- o nome do personagem não aparece em código, commit, PR, comentário ou documento;
- a resposta para quem chamou o agente tem no máximo 560 caracteres;
- produção espera o sim explícito do humano.

Uma regra só vale para o agente cujo arquivo a traz.

## Onde ver cada agente

Os arquivos em `agents/` são a fonte. Para listar papel, modelo e esforço de todos de uma vez, na raiz do repo:

```sh
grep -H -E '^(description|model|effort|tools):' agents/*.md
```

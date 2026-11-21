# Dependências

Versões conferidas em 21/11/2026 na máquina onde o teste roda. Todas são open source. Nada aqui foi lido a fundo além dos READMEs e do `SKILL.md` do i-have-adhd.

## Instaladas

| Projeto | Função | Versão | Licença | Origem |
| --- | --- | --- | --- | --- |
| pstack | Espinha: roteiros prontos, revisão por outro agente, PR até o verde | 0.15.46 | MIT | https://github.com/cursor/plugins/tree/main/pstack |
| i-have-adhd | Formato de resposta para TDAH: ação primeiro, passos numerados | 0.4.1 | MIT | https://github.com/ayghri/i-have-adhd |
| ccmanager | Aviso quando um agente muda de estado | 4.4.11 | MIT | https://github.com/kbwo/ccmanager |
| basic-memory | Memória em Markdown entre sessões | 1.0.0 | AGPL-3.0 | https://github.com/basicmachines-co/basic-memory |
| container-use | Ambiente isolado por agente, sem poder de merge | 0.4.2 | Apache-2.0 | https://github.com/dagger/container-use |
| OpenSpec | Contrato e spec antes de codar | 1.7.0 | MIT | https://github.com/Fission-AI/OpenSpec |

A skill `grilling`, que faz perguntas para refinar a atividade, já estava instalada como skill local. A origem dela não foi confirmada.

## Como instalar e remover

| Projeto | Instalar | Remover |
| --- | --- | --- |
| pstack | `claude plugin marketplace add cursor/plugins` e `claude plugin install pstack@cursor-plugins` | `claude plugin uninstall pstack@cursor-plugins` e `claude plugin marketplace remove cursor-plugins` |
| i-have-adhd | `claude plugin marketplace add ayghri/i-have-adhd` e `claude plugin install i-have-adhd@i-have-adhd` | `claude plugin uninstall i-have-adhd` e `claude plugin marketplace remove i-have-adhd` |
| ccmanager | `npm install -g ccmanager` | `npm uninstall -g ccmanager` |
| basic-memory | `uv tool install basic-memory --prerelease=allow` | `uv tool uninstall basic-memory` |
| container-use | `brew install dagger/tap/container-use` | `brew uninstall container-use` |
| OpenSpec | `npm install -g @fission-ai/openspec` | `npm uninstall -g @fission-ai/openspec` |

## Instaladas mas ainda não ligadas

- pstack: falta reiniciar o Claude Code e rodar `/pstack:setup-pstack`.
- basic-memory: não aponta para nenhuma pasta de notas e não tem servidor MCP configurado.
- container-use: não está ligado ao Claude Code.

## Fora, e por quê

| Projeto | Motivo |
| --- | --- |
| orchestrate (repo de plugins do Cursor) | O script depende dos agentes de nuvem do Cursor, e o marketplace do Claude Code só lista o pstack |
| agent-teams (wshobson/agents) | Recusado pelo Mateus em 21/11/2026 |
| claude-flow (ruflo) | O estudo de 20/11 descartou: complexidade acima do ganho |

## Requisitos da máquina

Node 22, npm, uv, Homebrew, e um runtime de container com o comando `docker` respondendo (Podman serve).

# Dependências

A lista do que o Megazord usa, com versão ou commit fixo, licença e origem, está em [`deps.toml`](../../deps.toml). Esta página descreve o formato desse arquivo e diz o que os scripts cobrem. Para a versão de uma peça, leia o arquivo.

## Formato do `deps.toml`

O arquivo começa com `format_version = 1`. Cada dependência é uma tabela `[[dependency]]`.

Campos que toda dependência tem:

| Campo | O que guarda |
| --- | --- |
| `name` | O nome curto da peça |
| `kind` | O tipo de peça, que define os outros campos |
| `license` | A licença declarada pelo projeto |
| `function` | Para que o Megazord usa a peça, em uma frase |

Campos por tipo:

| `kind` | Campos próprios |
| --- | --- |
| `claude-plugin` | `marketplace`, `plugin`, `version`, `commit`, `source` |
| `npm` | `package`, `version`, `source` |
| `brew` | `formula`, `version`, `source`, `requires` |
| `container-image` | `image` (com a tag), `source` |
| `skill` | `repository`, `path`, `commit` |

`requires` diz o que precisa existir na máquina para a peça funcionar.

## O que os scripts cobrem

`deps.toml` e os scripts não listam as mesmas peças:

| Peça | Fixada em `deps.toml` | Instalada por `instalar.sh` | Conferida por `conferir.sh` |
| --- | --- | --- | --- |
| pstack, i-have-adhd, container-use | Sim | Sim | Sim |
| openspec (opcional) | Sim | Só com `INSTALAR_OPCIONAIS=1` | Mostra "não instalado (opcional)" sem acusar falta |
| lychee (imagem de container) | Sim | Não | Não |
| Skills (`kind = "skill"`) | Sim | Não | Não |
| Runtime de container (`docker`) | Não | Não | Sim |

As skills ficam na pasta de skills do usuário, `~/.claude/skills/`. `deps.toml` traz o repositório, a pasta e o commit de origem de cada uma.

## Como instalar e remover cada peça

| Peça | Instalar | Remover |
| --- | --- | --- |
| pstack | `claude plugin marketplace add cursor/plugins` e `claude plugin install pstack@cursor-plugins` | `claude plugin uninstall pstack@cursor-plugins` e `claude plugin marketplace remove cursor-plugins` |
| i-have-adhd | `claude plugin marketplace add ayghri/i-have-adhd` e `claude plugin install i-have-adhd@i-have-adhd` | `claude plugin uninstall i-have-adhd` e `claude plugin marketplace remove i-have-adhd` |
| container-use | `brew install dagger/tap/container-use` | `brew uninstall container-use` |
| openspec | `npm install -g @fission-ai/openspec` | `npm uninstall -g @fission-ai/openspec` |

Os comandos de instalação são os mesmos de [`scripts/instalar.sh`](../../scripts/instalar.sh).

## Requisitos da máquina

- Claude Code, com o comando `claude` no terminal.
- Node e npm. A máquina do teste usa Node 22.
- uv.
- Homebrew.
- Um runtime de container que responda ao comando `docker`. Podman serve quando o comando `docker` chama o Podman, por exemplo por um link `docker` apontando para `podman`.

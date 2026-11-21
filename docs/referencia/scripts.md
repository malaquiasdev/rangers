# Scripts

Os scripts em [`scripts/`](../../scripts/) instalam e conferem as peças até o instalador em Go existir. Os dois são Bash e rodam da raiz do repo ou de qualquer outra pasta.

## `scripts/conferir.sh`

Mostra o estado de cada peça, uma por linha, e não muda nada na máquina.

Cada linha tem o nome da peça e a versão que o próprio programa informa. Uma peça que não está instalada aparece como `FALTANDO`. A última linha confere se o comando `docker` responde a `docker info`: aparece `respondendo` ou `SEM RESPOSTA (o container-use precisa dele)`.

| Peça | Como o script descobre a versão |
| --- | --- |
| Plugins do Claude Code (pstack, i-have-adhd) | Procura o plugin na saída de `claude plugin list` e lê a linha `Version:` seguinte |
| Comandos de terminal (container-use, openspec) | Confere se o comando existe e lê a primeira linha da saída de versão dele |
| Runtime de container | `docker info` |

Saída e código de retorno:

| Situação | Última linha | Código |
| --- | --- | --- |
| Nada faltando | `Tudo no lugar.` | `0` |
| Uma ou mais peças faltando, ou `docker` sem resposta | `N item(ns) faltando. Rode scripts/instalar.sh` | `1` |

O script não confere a versão contra `deps.toml`. Ele só diz se a peça existe e qual versão está instalada.

## `scripts/instalar.sh`

Instala cada peça que estiver faltando e pula as que já existem. Pode ser rodado de novo sem efeito colateral nas peças já instaladas.

| Peça | Como sabe que já existe | O que roda se faltar |
| --- | --- | --- |
| pstack | `claude plugin list` contém `pstack@cursor-plugins` | `claude plugin marketplace add cursor/plugins` e `claude plugin install pstack@cursor-plugins` |
| i-have-adhd | `claude plugin list` contém `i-have-adhd@i-have-adhd` | `claude plugin marketplace add ayghri/i-have-adhd` e `claude plugin install i-have-adhd@i-have-adhd` |
| container-use | O comando `container-use` existe | `brew install dagger/tap/container-use` |
| openspec | O comando `openspec` existe | `npm install -g @fission-ai/openspec` |

Ao final, imprime `Feito. Reinicie o Claude Code e rode /pstack:setup-pstack uma vez.` e roda `conferir.sh`. O código de retorno é o do `conferir.sh`.

O script para no primeiro comando que falhar (`set -eu`).

O que o script não faz:

- não instala nem liga o runtime de container;
- não fixa versão: instala a versão atual de cada fonte;
- não instala as skills nem a imagem do lychee listadas em `deps.toml`;
- não liga os agentes no Claude Code (veja [como ligar os agentes](../como-fazer/ligar-os-agentes.md));
- não configura o container-use como servidor MCP.

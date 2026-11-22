# megazord

Um repositório de agentes e regras em cima de projetos open source que já coordenam agentes. Substitui a plataforma `jaiminho-suite`, que foi arquivada.

Os plugins instalados fazem o trabalho pesado. Este repo guarda os agentes, a lista exata do que está instalado, os scripts que instalam e conferem, e o plano de teste que decide se a ideia funciona. O porquê está em [por que uma camada fina](docs/explicacao/por-que-uma-camada-fina.md).

## O que tem aqui

| Caminho | Para quê |
| --- | --- |
| `agents/` | Um arquivo por agente, com nome, papel, modelo e limites |
| `deps.toml` | Cada dependência, com a versão ou o commit fixo, a licença e a origem |
| `scripts/` | Instala e confere o que faltar, até o instalador em Go existir |
| `docs/` | A documentação, começando pelo [índice](docs/index.md) |

## Começo rápido

1. `scripts/conferir.sh` mostra o estado.
2. `scripts/instalar.sh` completa o que faltar.
3. Reinicie o Claude Code e rode `/pstack:setup-pstack` uma vez.
4. Ligue os agentes: [como ligar os agentes](docs/como-fazer/ligar-os-agentes.md).
5. Rode o teste: [como rodar o teste](docs/como-fazer/rodar-o-teste.md).

Primeira vez? Siga o tutorial [preparar a máquina](docs/tutorial/preparar-a-maquina.md).

## Voltar atrás

[Como desfazer a instalação](docs/como-fazer/desfazer-a-instalacao.md) traz o comando de remoção de cada peça e onde está a plataforma antiga.

# Como desfazer a instalação

Este guia mostra como tirar o Megazord da máquina, por inteiro ou peça por peça.

## Tirar uma peça só

Rode o comando de remoção da peça, da tabela em [como instalar e remover cada peça](../referencia/dependencias.md#como-instalar-e-remover-cada-peça). Se a peça for um plugin do Claude Code, reinicie o Claude Code depois.

Rode `scripts/conferir.sh` em seguida. A peça removida aparece como `FALTANDO`, e o script termina com código `1`. Isso é esperado.

## Tirar tudo

1. Desligue os agentes, como em [desligar os agentes](ligar-os-agentes.md#desligar-os-agentes).
2. Rode o comando de remoção de cada peça da tabela de dependências.
3. Reinicie o Claude Code.
4. Rode `scripts/conferir.sh`. Todas as peças aparecem como `FALTANDO`.

As skills listadas em `deps.toml` ficam em `~/.claude/skills/`. Os scripts não as instalaram, então elas não saem com os comandos acima. Apague a pasta de cada skill se quiser tirá-las também.

## Voltar para a plataforma antiga

A plataforma anterior está guardada em `~/.jaiminho-arquivado-20261121`. O repositório `jaiminho-suite` está arquivado no GitHub. Este repo não traz o passo a passo para religá-la.

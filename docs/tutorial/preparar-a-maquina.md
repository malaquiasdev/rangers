# Preparar a máquina pela primeira vez

Neste tutorial você vai deixar uma máquina pronta para usar o Jaiminho: instalar as peças open source, configurar o pstack e ver os agentes deste repo na lista do Claude Code.

## Antes de começar

Você precisa de:

- o clone deste repositório;
- Claude Code instalado, com o comando `claude` no terminal;
- Node com npm, uv e Homebrew;
- um runtime de container que responda ao comando `docker`. Podman serve, desde que o comando `docker` chame o Podman (por exemplo, um link `docker` apontando para `podman`) e a máquina do Podman esteja ligada com `podman machine start`.

Abra um terminal na raiz do clone. Todos os comandos abaixo rodam dali.

## 1. Veja o que já existe

```sh
scripts/conferir.sh
```

Você vai ver uma linha por peça, com a versão ao lado. As que ainda não estão instaladas aparecem como `FALTANDO`. A última linha diz quantos itens faltam ou `Tudo no lugar.`.

Não se preocupe com os `FALTANDO`: o próximo passo cuida deles.

## 2. Instale o que falta

```sh
scripts/instalar.sh
```

O script instala só o que faltava e, no fim, roda a conferência de novo. Repare na última linha: agora ela diz `Tudo no lugar.`.

Se ainda aparecer `item(ns) faltando`, procure nas linhas de cima qual peça está como `FALTANDO` ou `SEM RESPOSTA`. O caso mais comum é a linha do `docker`: o script não liga o runtime de container. Ligue o Podman com `podman machine start` e rode `scripts/conferir.sh` outra vez.

## 3. Configure o pstack

Feche o Claude Code, se estiver aberto, e abra uma sessão nova. Na sessão, rode:

```
/pstack:setup-pstack
```

O pstack detecta os modelos que você tem e grava uma regra com o modelo de cada papel. Siga o que ele pedir. Isso só precisa ser feito uma vez.

## 4. Ligue os agentes

De volta ao terminal, na raiz do clone:

```sh
mkdir -p ~/.claude/agents
ln -s "$PWD"/agents/*.md ~/.claude/agents/
```

Abra outra sessão nova do Claude Code e rode:

```
/agents
```

Os agentes deste repo aparecem na lista, cada um com o papel na descrição: Coordinator, Author, Reviewer e os demais.

## O que você fez

A máquina tem as peças open source instaladas, o pstack configurado e os agentes ligados. O próximo passo é rodar o teste com uma atividade real: veja [como rodar o teste](../como-fazer/rodar-o-teste.md).

Para entender por que o repo é tão pequeno, leia [por que uma camada fina](../explicacao/por-que-uma-camada-fina.md).

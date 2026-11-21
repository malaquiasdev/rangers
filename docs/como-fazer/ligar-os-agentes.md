# Como ligar os agentes no Claude Code

O Claude Code carrega os subagentes do usuário a partir de `~/.claude/agents/`. Os scripts não fazem essa ligação. Para usar os agentes deste repo, crie um link simbólico para cada arquivo de `agents/`. Assim uma mudança no repo vale na próxima sessão, sem copiar nada.

1. Na raiz do repo, crie a pasta, se ela não existir, e os links:

   ```sh
   mkdir -p ~/.claude/agents
   ln -s "$PWD"/agents/*.md ~/.claude/agents/
   ```

   Se o `ln` disser `File exists` para algum arquivo, já existe um agente com esse nome na pasta. Confira o que é com `ls -l ~/.claude/agents` antes de decidir. Não apague um arquivo comum dessa pasta sem saber de onde ele veio.

2. Abra uma sessão nova do Claude Code. Os agentes carregam quando a sessão começa.

3. Rode `/agents` na sessão e confira se os agentes aparecem na lista.

## Conferir quais agentes estão ligados

```sh
ls -l ~/.claude/agents
```

Os agentes deste repo aparecem como links (`->`) apontando para a pasta `agents/` do clone.

## Desligar os agentes

Na raiz do repo, apague só os links que apontam para cá:

```sh
find ~/.claude/agents -type l -lname "$PWD/agents/*" -print -delete
```

O comando mostra cada link que apagou. Arquivos comuns e links para outros lugares ficam onde estão.

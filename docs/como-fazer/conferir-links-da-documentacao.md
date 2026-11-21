# Como conferir os links da documentação

Use este guia depois de mexer no `README.md` ou em `docs/`. Ele confere os links entre arquivos e as âncoras de cada título, sem acessar a internet.

O conferidor é o lychee, numa imagem de container. A versão da imagem está em [`deps.toml`](../../deps.toml).

1. Garanta que o Podman está de pé:

   ```sh
   podman machine start
   ```

   Se a máquina já estiver rodando, o comando avisa e não muda nada.

2. Na raiz do repo, rode:

   ```sh
   podman run --rm -v "$PWD:/w:ro" -w /w docker.io/lycheeverse/lychee:0.24.2 --offline --include-fragments README.md 'docs/**/*.md'
   ```

3. Leia o resumo no fim. Se aparecer algum erro, o lychee mostra o arquivo e o link quebrado. Corrija o caminho ou o título e rode de novo.

Se a versão da imagem mudar em `deps.toml`, troque a tag no comando.

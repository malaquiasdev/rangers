# How to check the documentation links

Use this guide after changing `README.md` or `docs/`. It checks the links between files and the anchor of each heading, without going to the internet. `make check` runs the same check among the others; this guide runs only this one and shows the errors.

The checker is lychee, in a container image. The version of the image is in [`deps.toml`](../../deps.toml).

1. Make sure Podman is up:

   ```sh
   podman machine start
   ```

   If the machine is already running, the command says so and changes nothing.

2. From the root of the repository, run:

   ```sh
   podman run --rm -v "$PWD:/w:ro" -w /w docker.io/lycheeverse/lychee:0.24.2 --offline --include-fragments README.md 'docs/**/*.md'
   ```

3. Read the summary at the end. For each error, lychee shows the file and the broken link. Fix the path or the heading and run it again.

If the image version changes in `deps.toml`, change the tag in the command and in [`internal/check/golang.go`](../../internal/check/golang.go).

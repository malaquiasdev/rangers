# How to run the checks before a commit

This guide turns on the Git hooks of a clone and runs the same checks by hand. What each check looks at is in the [checks and release reference](../reference/checks-and-release.md). The tools the checks need are listed in [what the machine needs](../reference/dependencies.md#what-the-machine-needs).

1. Build the program and put it in your `PATH`, as in [the install steps](../tutorials/prepare-the-machine.md#1-see-what-is-already-there). The hooks call it by name.

2. Turn on the hooks, once per clone:

   ```sh
   make hooks
   ```

   From now on, every commit runs the fast checks and every push runs the full ones. Build the program again after you pull changes to it, so the hooks run the new checks.

3. If you have a list of words that must never reach the repository, such as company names, put it in `~/.config/rangers/privacy.txt`, one pattern per line. Without the file, the privacy check is skipped with a warning.

4. Start Podman, which the link check needs:

   ```sh
   podman machine start
   ```

5. Run the full checks:

   ```sh
   make check
   ```

   The last line says `tudo certo (full)` when everything passed. Otherwise, each failure has a `FALHOU:` line above it. Fix them and run it again.

For a quick run without Go tests, lint and links, use `make quick`.

A commit message that the hook refuses needs the Conventional Commits form, such as `docs: fix the install steps`.

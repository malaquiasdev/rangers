# How to uninstall

This guide removes Rangers from the machine, whole or one piece at a time.

## Remove one piece

Run the remove command of the piece, from the table in [how to install and remove each piece](../reference/dependencies.md#how-to-install-and-remove-each-piece). If the piece is a Claude Code plugin, restart Claude Code afterwards.

Then run `scripts/verify.sh`. The removed piece shows as `FALTANDO` and the script ends with code `1`. That is expected.

## Remove everything

1. Disconnect the agents and the skill, as in [disconnect the agents](connect-agents.md#disconnect-the-agents).
2. Stop the learning job with `~/.local/bin/rangers schedule uninstall`, then delete `~/.local/bin/rangers` and `~/.local/state/rangers/learn`.
3. Run the remove command of each piece in the dependencies table.
4. Restart Claude Code.
5. Run `scripts/verify.sh`. Every piece shows as `FALTANDO`.

The skills listed in `deps.toml` live in `~/.claude/skills/`. The scripts did not install them, so the commands above do not remove them. Delete the folder of each skill if you want them gone too.

To turn off the Git hooks of a clone, run `git config --unset core.hooksPath` in it.

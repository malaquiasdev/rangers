# How to uninstall

This guide removes Rangers from the machine, whole or one piece at a time. Run the commands from the root of the clone, because `rangers install --check` reads `deps.toml` from there.

## Remove one piece

Run the remove command of the piece, from the table in [how to install and remove each piece](../reference/dependencies.md#how-to-install-and-remove-each-piece). If the piece is a Claude Code plugin, restart Claude Code afterwards.

Then run `rangers install --check`. The removed piece shows as `MISSING` and the command ends with code `1`. That is expected. An optional piece, such as openspec, shows as `not installed (optional)` instead, and does not change the exit code.

## Remove everything

1. Disconnect the agents, the rules and the skill, as in [disconnect the agents](connect-agents.md#disconnect-the-agents).
   Then delete the links that `rangers install` made from the old helper names to the program: `find ~/.claude/bin -maxdepth 1 -type l -lname "*/rangers" -print -delete`.
2. Stop the learning job with `~/.local/bin/rangers schedule uninstall`, then delete `~/.local/state/rangers/learn`.
3. Run the remove command of each piece in the dependencies table.
4. Delete the folder of each skill listed in `deps.toml` from `~/.claude/skills/`, if you want the skills gone too. Each one has a `.rangers-commit` file. The commands above do not remove them.
5. Restart Claude Code.
6. Run `~/.local/bin/rangers install --check` to see what is left:
   - the plugins, container-use and lychee show `MISSING`;
   - openspec shows `not installed (optional)`;
   - a skill whose folder you kept still shows its commit, or `present, unpinned` when its folder has no `.rangers-commit` file;
   - `docker` shows `responding` while the container runtime runs. Rangers does not install it, so this guide does not remove it.
7. Delete `~/.local/bin/rangers`.

To turn off the Git hooks of a clone, run `git config --unset core.hooksPath` in it.

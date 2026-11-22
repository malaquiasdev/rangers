# How to install the skills the agents use

The agents call skills that live outside this repository, such as `grilling` for the coordinator. `deps.toml` pins each one with `kind = "skill"`, but no script installs them. This guide copies one skill at the pinned commit into `~/.claude/skills/`.

1. Read the entry of the skill in [`deps.toml`](../../deps.toml): `name`, `repository`, `path` and `commit`.

2. Clone the repository at that commit, in a temporary folder:

   ```sh
   git clone https://github.com/<repository> /tmp/skill-src
   git -C /tmp/skill-src checkout <commit>
   ```

3. Copy the skill folder under its name:

   ```sh
   mkdir -p ~/.claude/skills
   cp -R /tmp/skill-src/<path> ~/.claude/skills/<name>
   rm -rf /tmp/skill-src
   ```

   When `path` is `.`, the whole repository is the skill.

4. Check that `~/.claude/skills/<name>/SKILL.md` exists, and open a new Claude Code session.

Repeat for each skill. To remove one, delete its folder in `~/.claude/skills/`.

# The `rangers install` command

`rangers install` installs each piece listed in [`deps.toml`](../../deps.toml) that is missing and links the agents, the rules, the entry skill and the `bin/` helpers into Claude Code. With `--check` it shows the state of each piece and changes nothing. The source is in [`cmd/rangers/`](../../cmd/rangers/) and [`internal/install/`](../../internal/install/).

## Usage

```sh
rangers install [--optional] [--repo DIR] [--claude-dir DIR]
rangers install --check [--optional] [--repo DIR] [--claude-dir DIR]
```

| Flag | Default | What it sets |
| --- | --- | --- |
| `--check` | off | Shows the state of each piece and changes nothing |
| `--optional` | off | Installs the pieces marked `optional = true`, and counts them as required in `--check` |
| `--repo` | the first folder, from the current one upward, that has a `deps.toml` | Clone of this repository |
| `--claude-dir` | `~/.claude` | Claude Code folder, where the skills are installed and the links are made |

## What the install does

Each piece that is already there is left alone. The command stops at the first step that fails.

| Kind | What it runs when the piece is missing |
| --- | --- |
| `claude-plugin` | `claude plugin marketplace add <marketplace>` and `claude plugin install <plugin>` |
| `npm` | `npm install -g <package>@<version>` |
| `brew` | `brew install <formula>` |
| `container-image` | `podman pull <image>` |
| `skill` | Clones `https://github.com/<repository>.git` into a temporary folder, checks out `commit`, copies `path` into `<claude-dir>/skills/<name>` without `.git` and writes the commit in `.rangers-commit` |

A skill folder that already exists is never replaced, with or without the `.rangers-commit` file.

Then it links, like `ln -sfn`, `agents/*.md` into `<claude-dir>/agents/`, `rules/*.md` into `<claude-dir>/rules/`, `skills/rangers` into `<claude-dir>/skills/rangers` and `bin/*` into `<claude-dir>/bin/`. It replaces a file or a link in the way and refuses to replace a real folder. The `bin/*` links are more careful: a regular file in the way is replaced only when its bytes are identical to the repository file. A different one is left alone, the line `bin/<name> differs from the installed file, left alone` is printed and the other links are still made. It ends with `Restart Claude Code and run /pstack:setup-pstack once.` It does not start the container runtime.

## Output of `--check`

One line per piece, with its name and what was found. Each `bin/*` link comes next, then the container runtime as the last piece line.

| Kind | How the version is found |
| --- | --- |
| `claude-plugin` | The `Version:` line under the plugin in `claude plugin list` |
| `npm` | First line of `<name> --version` |
| `brew` | First line of `<name> version` |
| `container-image` | The tag, when `podman image exists` finds the image |
| `skill` | The first 12 characters of the commit in `<claude-dir>/skills/<name>/.rangers-commit`, or `present, unpinned` when the folder has no marker |
| `bin/<name>` link | `linked`, `NOT LINKED` or `differs from the installed file, left alone` |
| Container runtime | `docker` shows `responding` when `docker info` works, or `NOT RESPONDING (container-use needs it)` |

A piece that is not there shows `MISSING`. An optional piece that is not there shows `not installed (optional)` and is not counted, unless `--optional` is set.

| Situation | Last line | Exit code |
| --- | --- | --- |
| Nothing missing | `All in place.` | `0` |
| One or more pieces or `bin/*` links missing, or `docker` not answering | `N item(s) missing. Run rangers install` | `1` |

A `bin/*` link that differs is shown and not counted, because `rangers install` leaves it alone. The command does not compare versions with `deps.toml`. It only says whether the piece exists and which version is installed.

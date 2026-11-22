# The `rangers install` command

`rangers install --check` shows the state of each piece listed in [`deps.toml`](../../deps.toml) and changes nothing. The source is in [`cmd/rangers/`](../../cmd/rangers/) and [`internal/install/`](../../internal/install/).

## Usage

```sh
rangers install --check [--optional] [--repo DIR] [--claude-dir DIR]
```

| Flag | Default | What it sets |
| --- | --- | --- |
| `--check` | off | Shows the state of each piece and changes nothing |
| `--optional` | off | Counts the pieces marked `optional = true` as required |
| `--repo` | the first folder, from the current one upward, that has a `deps.toml` | Clone of this repository |
| `--claude-dir` | `~/.claude` | Claude Code folder, where the skills are looked up |

## Output

One line per piece, with its name and what was found. The last piece line is the container runtime.

| Kind | How the version is found |
| --- | --- |
| `claude-plugin` | The `Version:` line under the plugin in `claude plugin list` |
| `npm` | First line of `<name> --version` |
| `brew` | First line of `<name> version` |
| `container-image` | The tag, when `podman image exists` finds the image |
| `skill` | The first 12 characters of the commit in `<claude-dir>/skills/<name>/.rangers-commit`, or `present, unpinned` when the folder has no marker |
| Container runtime | `docker` shows `responding` when `docker info` works, or `NOT RESPONDING (container-use needs it)` |

A piece that is not there shows `MISSING`. An optional piece that is not there shows `not installed (optional)` and is not counted, unless `--optional` is set.

| Situation | Last line | Exit code |
| --- | --- | --- |
| Nothing missing | `All in place.` | `0` |
| One or more pieces missing, or `docker` not answering | `N item(s) missing. Run rangers install` | `1` |

The command does not compare versions with `deps.toml`. It only says whether the piece exists and which version is installed.

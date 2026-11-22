#!/usr/bin/env bash
set -eu

root=$(git rev-parse --show-toplevel)
cd "$root"
chmod +x .githooks/* scripts/*.sh
git config core.hooksPath .githooks
echo "ganchos ligados: pre-commit, commit-msg e pre-push"

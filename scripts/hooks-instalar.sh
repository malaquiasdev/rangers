#!/usr/bin/env bash
set -eu

raiz=$(git rev-parse --show-toplevel)
cd "$raiz"
chmod +x .githooks/* scripts/*.sh
git config core.hooksPath .githooks
echo "ganchos ligados: pre-commit, commit-msg e pre-push"

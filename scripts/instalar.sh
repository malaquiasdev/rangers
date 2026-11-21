#!/usr/bin/env bash
set -eu

tem_plugin() {
  claude plugin list 2>/dev/null | grep -q "$1"
}

if ! tem_plugin 'pstack@cursor-plugins'; then
  claude plugin marketplace add cursor/plugins
  claude plugin install pstack@cursor-plugins
fi

if ! tem_plugin 'i-have-adhd@i-have-adhd'; then
  claude plugin marketplace add ayghri/i-have-adhd
  claude plugin install i-have-adhd@i-have-adhd
fi

if ! command -v container-use >/dev/null; then
  brew install dagger/tap/container-use
fi

if [ "${INSTALAR_OPCIONAIS:-0}" = 1 ] && ! command -v openspec >/dev/null; then
  npm install -g @fission-ai/openspec
fi

printf 'Feito. Reinicie o Claude Code e rode /pstack:setup-pstack uma vez.\n'
"$(dirname "$0")/conferir.sh"

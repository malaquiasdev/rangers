#!/usr/bin/env bash
set -eu

state=${1:?pasta de estado}
janela=${MADRUGA_JANELA_HORAS:-24}
projetos=${CLAUDE_PROJECTS_DIR:-$HOME/.claude/projects}

if [ -f "$state/checkpoint" ]; then
  mudados=$(find "$projetos" -maxdepth 2 -name '*.jsonl' -newer "$state/checkpoint")
else
  mudados=$(find "$projetos" -maxdepth 2 -name '*.jsonl' -mmin -$((janela * 60)))
fi

for arquivo in $mudados; do
  sessao=$(basename "$arquivo" .jsonl)
  jq -r --arg sessao "$sessao" '
    select(.type == "user" and .message.role == "user")
    | .timestamp as $quando
    | .message.content
    | if type == "string" then . else (map(select(.type == "text").text) | join(" ")) end
    | select(length > 0)
    | select(startswith("<") | not)
    | select(startswith("Base directory") | not)
    | select(startswith("Another Claude") | not)
    | select(startswith("[Request interrupted") | not)
    | select(startswith("Caveat:") | not)
    | select(startswith("[Cross-session") | not)
    | select(startswith("[Subagent") | not)
    | gsub("\n"; " ")
    | "\($sessao)\t\($quando)\t\(.[0:600])"
  ' "$arquivo" 2>/dev/null || true
done

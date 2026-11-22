#!/usr/bin/env bash
set -eu

aqui=$(cd "$(dirname "$0")" && pwd)
repo=${JAIMINHO_REPO:-$HOME/Developer/me/jaiminho}
state=${MADRUGA_STATE:-$HOME/.local/state/jaiminho/madruga}
mkdir -p "$state"

carimbo=$(date +%Y-%m-%d-%H%M)
entrada="$state/entrada-$carimbo.txt"

"$aqui/madruga-extrair.sh" "$state" > "$entrada"

if [ ! -s "$entrada" ]; then
  rm -f "$entrada"
  echo "nada novo desde o último checkpoint"
  exit 0
fi

ramo="madruga/$carimbo"
arvore="$repo-worktrees/madruga-$carimbo"
git -C "$repo" worktree add -q -b "$ramo" "$arvore" main

cd "$arvore"
claude -p --agent madruga \
  --permission-mode acceptEdits \
  --allowedTools "Read,Edit,Write,Bash(git add:*),Bash(git commit:*),Bash(git status:*),Bash(git diff:*)" \
  "O arquivo $entrada tem, uma por linha, as falas que o humano digitou desde o último checkpoint (sessão, hora, texto). Proponha as regras nos arquivos desta pasta e faça o commit local no ramo $ramo."

touch "$state/checkpoint"

osascript <<EOF
tell application "Reminders"
  tell list "Tarefas Claude"
    make new reminder with properties {name:"Revisar as regras que o Madruga propôs", body:"Ramo $ramo em $arvore"}
  end tell
end tell
EOF

echo "ramo $ramo em $arvore"

#!/usr/bin/env bash
set -eu

aqui=$(cd "$(dirname "$0")" && pwd)
repo=${CENTRO_DE_COMANDO_REPO:-$HOME/Developer/me/centro-de-comando}
state=${ZORDON_STATE:-$HOME/.local/state/centro-de-comando/zordon}
mkdir -p "$state"

carimbo=$(date +%Y-%m-%d-%H%M)
entrada="$state/entrada-$carimbo.txt"

"$aqui/zordon-extrair.sh" "$state" > "$entrada"

if [ ! -s "$entrada" ]; then
  rm -f "$entrada"
  echo "nada novo desde o último checkpoint"
  exit 0
fi

ramo="zordon/$carimbo"
arvore="$repo-worktrees/zordon-$carimbo"
git -C "$repo" worktree add -q -b "$ramo" "$arvore" main

cd "$arvore"
claude -p "O arquivo $entrada tem, uma por linha, as falas que o humano digitou desde o último checkpoint (sessão, hora, texto). Proponha as regras nos arquivos desta pasta e faça o commit local no ramo $ramo." \
  --agent zordon \
  --permission-mode acceptEdits \
  --allowedTools "Read,Edit,Write,Bash(git add:*),Bash(git commit:*),Bash(git status:*),Bash(git diff:*)"

touch "$state/checkpoint"

osascript <<EOF
tell application "Reminders"
  tell list "Tarefas Claude"
    make new reminder with properties {name:"Revisar as regras que o Zordon propôs", body:"Ramo $ramo em $arvore"}
  end tell
end tell
EOF

echo "ramo $ramo em $arvore"

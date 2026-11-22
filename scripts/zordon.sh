#!/usr/bin/env bash
set -eu

aqui=$(cd "$(dirname "$0")" && pwd)
repo=${RANGERS_REPO:-$HOME/Developer/me/rangers}
state=${ZORDON_STATE:-$HOME/.local/state/rangers/zordon}
mkdir -p "$state"

carimbo=$(date +%Y-%m-%d-%H%M)
entrada="$state/entrada-$carimbo.txt"

"$aqui/zordon-extrair.sh" "$state" > "$entrada"

falas=$(wc -l < "$entrada" | tr -d ' ')
minimo=${ZORDON_MIN_FALAS:-10}

if [ "$falas" -lt "$minimo" ]; then
  rm -f "$entrada"
  echo "só $falas falas novas, abaixo do mínimo de $minimo: não rodo o modelo"
  exit 0
fi

ramo="zordon/$carimbo"
arvore="$repo-worktrees/zordon-$carimbo"
git -C "$repo" worktree add -q -b "$ramo" "$arvore" main

cd "$arvore"
saida="$state/saida-$carimbo.json"
claude -p "O arquivo $entrada tem, uma por linha, as falas que o humano digitou desde o último checkpoint (sessão, hora, texto). Proponha as regras nos arquivos desta pasta e faça o commit local no ramo $ramo." \
  --agent zordon \
  --permission-mode acceptEdits \
  --allowedTools "Read,Edit,Write,Bash(git add:*),Bash(git commit:*),Bash(git status:*),Bash(git diff:*)" \
  --output-format json > "$saida"

jq -r '.result' "$saida"

jq -r --arg quando "$carimbo" --arg falas "$falas" '
  [$quando, $falas, (.duration_ms / 1000 | floor), .num_turns,
   (.total_cost_usd * 1000 | floor / 1000),
   (.usage.input_tokens + .usage.cache_creation_input_tokens + .usage.cache_read_input_tokens),
   .usage.output_tokens]
  | @tsv' "$saida" >> "$state/metricas.tsv"

touch "$state/checkpoint"

osascript <<EOF
tell application "Reminders"
  tell list "Tarefas Claude"
    make new reminder with properties {name:"Revisar as regras que o Zordon propôs", body:"Ramo $ramo em $arvore"}
  end tell
end tell
EOF

echo "ramo $ramo em $arvore"

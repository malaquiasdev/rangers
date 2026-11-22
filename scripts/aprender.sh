#!/usr/bin/env bash
set -eu

aqui=$(cd "$(dirname "$0")" && pwd)
repo=${RANGERS_REPO:-$HOME/Developer/me/rangers}
state=${APRENDER_STATE:-$HOME/.local/state/rangers/aprender}
mkdir -p "$state"

intervalo=${APRENDER_INTERVALO_HORAS:-2}
if [ -f "$state/checkpoint" ]; then
  idade=$(( $(date +%s) - $(stat -f %m "$state/checkpoint") ))
  if [ "$idade" -lt $((intervalo * 3600)) ]; then
    echo "a última rodada foi há $((idade / 60)) min, menos que o intervalo de ${intervalo}h: não rodo"
    exit 0
  fi
fi

carimbo=$(date +%Y-%m-%d-%H%M)
entrada="$state/entrada-$carimbo.txt"

"$aqui/aprender-extrair.sh" "$state" > "$entrada"

falas=$(wc -l < "$entrada" | tr -d ' ')
minimo=${APRENDER_MIN_FALAS:-10}

if [ "$falas" -lt "$minimo" ]; then
  rm -f "$entrada"
  echo "só $falas falas novas, abaixo do mínimo de $minimo: não rodo o modelo"
  exit 0
fi

ramo="aprender/$carimbo"
arvore="$repo-worktrees/aprender-$carimbo"
git -C "$repo" worktree add -q -b "$ramo" "$arvore" main

cd "$arvore"
saida="$state/saida-$carimbo.json"
claude -p "O arquivo $entrada tem, uma por linha, as falas que o humano digitou desde o último checkpoint (sessão, hora, texto). Proponha as regras nos arquivos desta pasta e faça o commit local no ramo $ramo." \
  --agent zordon \
  --model "${APRENDER_MODELO_REGRAS:-opus}" \
  --max-budget-usd "${APRENDER_TETO_USD:-3}" \
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
    make new reminder with properties {name:"Revisar as regras que as regras propostas", body:"Ramo $ramo em $arvore"}
  end tell
end tell
EOF

echo "ramo $ramo em $arvore"

#!/usr/bin/env bash
set -eu

raiz=$(git rev-parse --show-toplevel)
cd "$raiz"

modo=${1:-completo}
falhas=0

falha() {
  printf 'FALHOU: %s\n' "$1"
  falhas=$((falhas + 1))
}

for arquivo in scripts/*.sh .githooks/*; do
  [ -f "$arquivo" ] || continue
  bash -n "$arquivo" || falha "sintaxe de $arquivo"
  if sed 1d "$arquivo" | grep -qE '^[[:space:]]*#'; then
    falha "comentário em $arquivo"
  fi
done

for arquivo in agents/*.md; do
  nome=$(basename "$arquivo" .md)
  cabecalho=$(awk 'NR == 1 && $0 == "---" { dentro = 1; next } dentro && $0 == "---" { exit } dentro' "$arquivo")
  [ "$(printf '%s\n' "$cabecalho" | sed -n 's/^name: //p')" = "$nome" ] || falha "$arquivo: name diferente do nome do arquivo"
  printf '%s\n' "$cabecalho" | grep -qE '^model: (opus|sonnet|haiku)$' || falha "$arquivo: model inválido"
  printf '%s\n' "$cabecalho" | grep -qE '^effort: (low|medium|high|xhigh|max)$' || falha "$arquivo: effort inválido"
  printf '%s\n' "$cabecalho" | grep -q '^description: .' || falha "$arquivo: sem description"
  grep -q '^Character: ' "$arquivo" || falha "$arquivo: sem a linha Character"
done

for arquivo in skills/*/SKILL.md; do
  nome=$(basename "$(dirname "$arquivo")")
  grep -q "^name: $nome\$" "$arquivo" || falha "$arquivo: name diferente da pasta"
done

python3 -I - <<'PY' || falha "deps.toml inválido"
import sys
import tomllib

with open("deps.toml", "rb") as arquivo:
    dados = tomllib.load(arquivo)

for item in dados.get("dependency", []):
    for campo in ("name", "kind", "license", "function"):
        if campo not in item:
            print("dependência sem", campo, item.get("name"))
            sys.exit(1)
PY

personagens=$(cat scripts/personagens.txt)
achados=$(grep -rnwE "$personagens" --exclude-dir=.git --exclude-dir=.serena --exclude-dir=agents --exclude-dir=skills --exclude=personagens.txt . | grep -v -- '--agent ' || true)
if [ -n "$achados" ]; then
  printf '%s\n' "$achados"
  falha "nome de personagem fora de agents/ e skills/"
fi

padroes=${PRIVACIDADE_ARQUIVO:-$HOME/.config/rangers/privacidade.txt}
if [ -f "$padroes" ]; then
  if grep -rniEf "$padroes" --exclude-dir=.git --exclude-dir=.serena . ; then
    falha "nome de empresa no repositório"
  fi
else
  echo "aviso: sem $padroes, a checagem de privacidade foi pulada"
fi

if [ "$modo" = "completo" ]; then
  podman run --rm -v "$PWD:/w:ro" -w /w docker.io/lycheeverse/lychee:0.24.2 \
    --offline --include-fragments README.md 'docs/**/*.md' > /dev/null 2>&1 || falha "links da documentação"
fi

if [ "$falhas" -gt 0 ]; then
  printf '\n%s checagem(ns) falharam\n' "$falhas"
  exit 1
fi

echo "tudo certo ($modo)"

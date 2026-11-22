#!/usr/bin/env bash
set -eu

root=$(git rev-parse --show-toplevel)
cd "$root"

mode=${1:-full}
failures=0

fail() {
  printf 'FALHOU: %s\n' "$1"
  failures=$((failures + 1))
}

for path in scripts/*.sh .githooks/*; do
  [ -f "$path" ] || continue
  bash -n "$path" || fail "sintaxe de $path"
  if sed 1d "$path" | grep -qE '^[[:space:]]*#'; then
    fail "comentário em $path"
  fi
done

for path in agents/*.md; do
  agent=$(basename "$path" .md)
  header=$(awk 'NR == 1 && $0 == "---" { inside = 1; next } inside && $0 == "---" { exit } inside' "$path")
  [ "$(printf '%s\n' "$header" | sed -n 's/^name: //p')" = "$agent" ] || fail "$path: name diferente do nome do arquivo"
  printf '%s\n' "$header" | grep -qE '^model: (opus|sonnet|haiku)$' || fail "$path: model inválido"
  printf '%s\n' "$header" | grep -qE '^effort: (low|medium|high|xhigh|max)$' || fail "$path: effort inválido"
  printf '%s\n' "$header" | grep -q '^description: .' || fail "$path: sem description"
  grep -q '^Character: ' "$path" || fail "$path: sem a linha Character"
done

for path in skills/*/SKILL.md; do
  skill=$(basename "$(dirname "$path")")
  grep -q "^name: $skill\$" "$path" || fail "$path: name diferente da pasta"
done

python3 -I - <<'PY' || fail "deps.toml inválido"
import sys
import tomllib

with open("deps.toml", "rb") as handle:
    data = tomllib.load(handle)

for item in data.get("dependency", []):
    for field in ("name", "kind", "license", "function"):
        if field not in item:
            print("dependência sem", field, item.get("name"))
            sys.exit(1)
PY

characters=$(cat scripts/characters.txt)
findings=$(grep -rnwE "$characters" --exclude-dir=.git --exclude-dir=.serena --exclude-dir=agents --exclude-dir=skills --exclude=characters.txt --exclude=README.md . | grep -v -- '--agent ' || true)
if [ -n "$findings" ]; then
  printf '%s\n' "$findings"
  fail "nome de personagem fora de agents/ e skills/"
fi

patterns=${PRIVACY_FILE:-$HOME/.config/rangers/privacy.txt}
if [ -f "$patterns" ]; then
  if grep -rniEf "$patterns" --exclude-dir=.git --exclude-dir=.serena . ; then
    fail "nome de empresa no repositório"
  fi
else
  echo "aviso: sem $patterns, a checagem de nomes de empresa foi pulada"
fi

if [ -f go.mod ]; then
  if grep -rnE '(^|[[:space:]])//|^[[:space:]]*/\*' --include='*.go' --exclude-dir=.git --exclude-dir=.serena . ; then
    fail "comentário em arquivo Go"
  fi
  [ -z "$(gofmt -l .)" ] || fail "gofmt: $(gofmt -l . | tr '\n' ' ')"
  if [ "$mode" = "full" ]; then
    go vet ./... || fail "go vet"
    go test ./... || fail "go test"
    golangci-lint run || fail "golangci-lint"
  fi
fi

if [ "$mode" = "full" ]; then
  podman run --rm -v "$PWD:/w:ro" -w /w docker.io/lycheeverse/lychee:0.24.2 \
    --offline --include-fragments README.md 'docs/**/*.md' > /dev/null 2>&1 || fail "links da documentação"
fi

if [ "$failures" -gt 0 ]; then
  printf '\n%s checagem(ns) falharam\n' "$failures"
  exit 1
fi

echo "tudo certo ($mode)"

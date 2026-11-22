#!/usr/bin/env bash
set -u

missing=0

show() {
  local name=$1 version=$2
  if [ -n "$version" ]; then
    printf '%-14s %s\n' "$name" "$version"
  else
    printf '%-14s FALTANDO\n' "$name"
    missing=$((missing + 1))
  fi
}

plugin_version() {
  claude plugin list 2>/dev/null | awk -v target="$1" '
    $0 ~ target { found = 1 }
    found && /Version:/ { print $2; exit }
  '
}

show pstack "$(plugin_version 'pstack@cursor-plugins')"
show i-have-adhd "$(plugin_version 'i-have-adhd@i-have-adhd')"
show container-use "$(command -v container-use >/dev/null && container-use version 2>&1 | head -1)"
optional_version=$(command -v openspec >/dev/null && openspec --version 2>/dev/null | head -1)
printf '%-14s %s\n' openspec "${optional_version:-não instalado (opcional)}"

if docker info >/dev/null 2>&1; then
  printf '%-14s %s\n' docker "respondendo"
else
  printf '%-14s %s\n' docker "SEM RESPOSTA (o container-use precisa dele)"
  missing=$((missing + 1))
fi

if [ "$missing" -gt 0 ]; then
  printf '\n%s item(ns) faltando. Rode scripts/install.sh\n' "$missing"
  exit 1
fi

printf '\nTudo no lugar.\n'

#!/usr/bin/env bash
set -u

faltando=0

mostrar() {
  local nome=$1 versao=$2
  if [ -n "$versao" ]; then
    printf '%-14s %s\n' "$nome" "$versao"
  else
    printf '%-14s FALTANDO\n' "$nome"
    faltando=$((faltando + 1))
  fi
}

plugin_versao() {
  claude plugin list 2>/dev/null | awk -v alvo="$1" '
    $0 ~ alvo { achou = 1 }
    achou && /Version:/ { print $2; exit }
  '
}

mostrar pstack "$(plugin_versao 'pstack@cursor-plugins')"
mostrar i-have-adhd "$(plugin_versao 'i-have-adhd@i-have-adhd')"
mostrar ccmanager "$(command -v ccmanager >/dev/null && ccmanager --version 2>/dev/null | head -1)"
mostrar basic-memory "$(command -v basic-memory >/dev/null && basic-memory --version 2>/dev/null | head -1)"
mostrar container-use "$(command -v container-use >/dev/null && container-use version 2>&1 | head -1)"
mostrar openspec "$(command -v openspec >/dev/null && openspec --version 2>/dev/null | head -1)"

if docker info >/dev/null 2>&1; then
  printf '%-14s %s\n' docker "respondendo"
else
  printf '%-14s %s\n' docker "SEM RESPOSTA (o container-use precisa dele)"
  faltando=$((faltando + 1))
fi

if [ "$faltando" -gt 0 ]; then
  printf '\n%s item(ns) faltando. Rode scripts/instalar.sh\n' "$faltando"
  exit 1
fi

printf '\nTudo no lugar.\n'

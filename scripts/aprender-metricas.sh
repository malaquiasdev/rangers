#!/usr/bin/env bash
set -eu

state=${APRENDER_STATE:-$HOME/.local/state/rangers/aprender}
arquivo="$state/metricas.tsv"

if [ ! -s "$arquivo" ]; then
  echo "nenhuma rodada registrada"
  exit 0
fi

printf 'rodada\tfalas\tsegundos\tturnos\tcusto_usd\ttokens_entrada\ttokens_saida\n'
cat "$arquivo"

awk -F'\t' '
  { rodadas++; segundos += $3; custo += $5; entrada += $6; saida += $7 }
  END {
    printf "\n%d rodadas, %d s no total (média %d s), US$ %.2f no total (média US$ %.2f)\n", rodadas, segundos, segundos / rodadas, custo, custo / rodadas
    printf "tokens: %d de entrada, %d de saída\n", entrada, saida
  }
' "$arquivo"

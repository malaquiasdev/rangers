package learn

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

const metricsHeader = "rodada\tfalas\tsegundos\tturnos\tcusto_usd\ttokens_entrada\ttokens_saida\n"

func Metrics(cfg *Config, out io.Writer) error {
	raw, err := os.ReadFile(filepath.Join(cfg.State, metricsName))
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return err
	}
	rows := strings.FieldsFunc(string(raw), func(r rune) bool { return r == '\n' })
	if len(rows) == 0 {
		_, err = fmt.Fprintln(out, "nenhuma rodada registrada")
		return err
	}
	var seconds, tokensIn, tokensOut int
	var cost float64
	for _, row := range rows {
		fields := strings.Split(row, "\t")
		seconds += integerAt(fields, 2)
		cost += decimalAt(fields, 4)
		tokensIn += integerAt(fields, 5)
		tokensOut += integerAt(fields, 6)
	}
	rounds := len(rows)
	_, err = fmt.Fprintf(out, "%s%s\n%d rodadas, %d s no total (média %d s), US$ %.2f no total (média US$ %.2f)\ntokens: %d de entrada, %d de saída\n",
		metricsHeader, strings.Join(rows, "\n")+"\n",
		rounds, seconds, seconds/rounds, cost, cost/float64(rounds), tokensIn, tokensOut)
	return err
}

func decimalAt(fields []string, index int) float64 {
	if index >= len(fields) {
		return 0
	}
	number, err := strconv.ParseFloat(fields[index], 64)
	if err != nil {
		return 0
	}
	return number
}

func integerAt(fields []string, index int) int {
	return int(decimalAt(fields, index))
}

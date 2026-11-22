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
	"time"
)

const metricsHeader = "rodada\tfalas\tsegundos\tturnos\tcusto_usd\ttokens_entrada\ttokens_saida\tpassada\n"

const (
	rulesPass = "rules"
	notesPass = "notes"
)

type totals struct {
	rounds  int
	seconds int
	cost    float64
}

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
	var tokensIn, tokensOut int
	var whole totals
	perPass := map[string]*totals{rulesPass: {}, notesPass: {}}
	var shown strings.Builder
	for _, row := range rows {
		fields := strings.Split(row, "\t")
		pass := rulesPass
		if len(fields) > 7 && fields[7] != "" {
			pass = fields[7]
		} else {
			row += "\t" + pass
		}
		shown.WriteString(row + "\n")
		tokensIn += integerAt(fields, 5)
		tokensOut += integerAt(fields, 6)
		for _, sum := range []*totals{&whole, perPass[pass]} {
			if sum != nil {
				sum.add(integerAt(fields, 2), decimalAt(fields, 4))
			}
		}
	}
	_, err = fmt.Fprintf(out, "%s%s\n%d rodadas, %d s no total (média %d s), US$ %.2f no total (média US$ %.2f)\ntokens: %d de entrada, %d de saída\n",
		metricsHeader, shown.String(),
		whole.rounds, whole.seconds, whole.seconds/whole.rounds, whole.cost, whole.cost/float64(whole.rounds), tokensIn, tokensOut)
	if err != nil {
		return err
	}
	for _, pass := range []string{rulesPass, notesPass} {
		sum := perPass[pass]
		if _, err = fmt.Fprintf(out, "%s: %d rodadas, %d s, US$ %.2f\n", pass, sum.rounds, sum.seconds, sum.cost); err != nil {
			return err
		}
	}
	return nil
}

func (t *totals) add(seconds int, cost float64) {
	t.rounds++
	t.seconds += seconds
	t.cost += cost
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

func spentToday(state string, now time.Time) (float64, error) {
	raw, err := os.ReadFile(filepath.Join(state, metricsName))
	if errors.Is(err, fs.ErrNotExist) {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}
	prefix := now.Format("2006-01-02") + "-"
	var spent float64
	for _, row := range strings.Split(string(raw), "\n") {
		fields := strings.Split(row, "\t")
		if strings.HasPrefix(fields[0], prefix) {
			spent += decimalAt(fields, 4)
		}
	}
	return spent, nil
}

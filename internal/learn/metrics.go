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

const metricsHeader = "rodada\tfalas\tsegundos\tturnos\tcusto_usd\t" +
	"tokens_entrada\ttokens_saida\tpassada\n"

type totals struct {
	rounds    int
	seconds   int
	tokensIn  int
	tokensOut int
	cost      float64
}

type summary struct {
	whole   totals
	perPass map[string]*totals
	shown   strings.Builder
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
	_, err = io.WriteString(out, summarize(rows).report())
	return err
}

func summarize(rows []string) *summary {
	sum := &summary{perPass: map[string]*totals{rulesPass: {}, notesPass: {}}}
	for _, row := range rows {
		fields := strings.Split(row, "\t")
		pass := rulesPass
		if len(fields) > 7 && fields[7] != "" {
			pass = fields[7]
		} else {
			row += "\t" + pass
		}
		sum.shown.WriteString(row + "\n")
		sum.whole.add(fields)
		if byPass, ok := sum.perPass[pass]; ok {
			byPass.add(fields)
		}
	}
	return sum
}

func (s *summary) report() string {
	var text strings.Builder
	whole := &s.whole
	text.WriteString(metricsHeader + s.shown.String())
	fmt.Fprintf(&text, "\n%d rodadas, %d s no total (média %d s), ", whole.rounds, whole.seconds,
		whole.seconds/whole.rounds)
	fmt.Fprintf(&text, "US$ %.2f no total (média US$ %.2f)\n", whole.cost,
		whole.cost/float64(whole.rounds))
	fmt.Fprintf(&text, "tokens: %d de entrada, %d de saída\n", whole.tokensIn, whole.tokensOut)
	for _, pass := range []string{rulesPass, notesPass} {
		sum := s.perPass[pass]
		fmt.Fprintf(&text, "%s: %d rodadas, %d s, US$ %.2f\n", pass, sum.rounds, sum.seconds, sum.cost)
	}
	return text.String()
}

func (t *totals) add(fields []string) {
	t.rounds++
	t.seconds += integerAt(fields, 2)
	t.cost += decimalAt(fields, 4)
	t.tokensIn += integerAt(fields, 5)
	t.tokensOut += integerAt(fields, 6)
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

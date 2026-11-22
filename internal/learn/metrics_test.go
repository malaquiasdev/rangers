package learn

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func metricsIn(t *testing.T, content *string) (*Config, *bytes.Buffer) {
	t.Helper()
	state := t.TempDir()
	if content != nil {
		if err := os.WriteFile(filepath.Join(state, "metrics.tsv"), []byte(*content), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	return &Config{State: state}, &bytes.Buffer{}
}

func TestMetrics(t *testing.T) {
	t.Run("given no metrics file, when showing, then it says there is no round", func(t *testing.T) {
		cfg, out := metricsIn(t, nil)
		if err := Metrics(cfg, out); err != nil {
			t.Fatal(err)
		}
		if out.String() != "nenhuma rodada registrada\n" {
			t.Fatalf("got %q", out.String())
		}
	})

	t.Run("given an empty metrics file, when showing, then it says there is no round", func(t *testing.T) {
		empty := ""
		cfg, out := metricsIn(t, &empty)
		if err := Metrics(cfg, out); err != nil {
			t.Fatal(err)
		}
		if out.String() != "nenhuma rodada registrada\n" {
			t.Fatalf("got %q", out.String())
		}
	})

	t.Run("given old rows without the pass column, when showing, then it counts them as rules and shows the column", func(t *testing.T) {
		content := "2026-11-21-0900\t10\t100\t5\t1.5\t1000\t200\n2026-11-22-1030\t20\t51\t3\t0.5\t500\t100\n"
		cfg, out := metricsIn(t, &content)
		if err := Metrics(cfg, out); err != nil {
			t.Fatal(err)
		}
		expected := "rodada\tfalas\tsegundos\tturnos\tcusto_usd\ttokens_entrada\ttokens_saida\tpassada\n" +
			"2026-11-21-0900\t10\t100\t5\t1.5\t1000\t200\trules\n" +
			"2026-11-22-1030\t20\t51\t3\t0.5\t500\t100\trules\n" +
			"\n2 rodadas, 151 s no total (média 75 s), US$ 2.00 no total (média US$ 1.00)\n" +
			"tokens: 1500 de entrada, 300 de saída\n" +
			"rules: 2 rodadas, 151 s, US$ 2.00\n" +
			"notes: 0 rodadas, 0 s, US$ 0.00\n"
		if out.String() != expected {
			t.Fatalf("expected\n%s\ngot\n%s", expected, out.String())
		}
	})

	t.Run("given rows from both passes, when showing, then it totals each pass and the whole", func(t *testing.T) {
		content := "2026-11-22-0900\t10\t100\t5\t1.5\t1000\t200\n" +
			"2026-11-22-1030\t20\t60\t3\t0.5\t500\t100\trules\n" +
			"2026-11-22-1030\t20\t40\t2\t0.25\t300\t50\tnotes\n"
		cfg, out := metricsIn(t, &content)
		if err := Metrics(cfg, out); err != nil {
			t.Fatal(err)
		}
		for _, expected := range []string{
			"3 rodadas, 200 s no total (média 66 s), US$ 2.25 no total (média US$ 0.75)",
			"tokens: 1800 de entrada, 350 de saída",
			"rules: 2 rodadas, 160 s, US$ 2.00",
			"notes: 1 rodadas, 40 s, US$ 0.25",
			"2026-11-22-1030\t20\t40\t2\t0.25\t300\t50\tnotes\n",
		} {
			if !strings.Contains(out.String(), expected) {
				t.Errorf("missing %q in\n%s", expected, out.String())
			}
		}
	})

	t.Run("given a row with a non-numeric field, when showing, then it counts the field as zero", func(t *testing.T) {
		content := "2026-11-22-1030\t20\tx\t3\t0.5\t500\t100\n"
		cfg, out := metricsIn(t, &content)
		if err := Metrics(cfg, out); err != nil {
			t.Fatal(err)
		}
		if got := out.String(); !bytes.Contains([]byte(got), []byte("1 rodadas, 0 s no total (média 0 s), US$ 0.50")) {
			t.Fatalf("got %q", got)
		}
	})
}

func TestSpentToday(t *testing.T) {
	t.Run("given rows from both passes and old rows today, when summing, then all of them count", func(t *testing.T) {
		content := "2026-11-22-0800\t1\t1\t1\t1.5\t1\t1\n" +
			"2026-11-22-0900\t1\t1\t1\t0.5\t1\t1\trules\n" +
			"2026-11-22-0900\t1\t1\t1\t0.25\t1\t1\tnotes\n" +
			"2026-11-21-2300\t1\t1\t1\t9\t1\t1\tnotes\n"
		cfg, _ := metricsIn(t, &content)
		spent, err := spentToday(cfg.State, time.Date(2026, 11, 22, 12, 0, 0, 0, time.Local))
		if err != nil || spent != 2.25 {
			t.Fatalf("got %v, %v", spent, err)
		}
	})
}

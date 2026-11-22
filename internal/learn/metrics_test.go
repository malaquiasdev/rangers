package learn

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
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

	t.Run("given two rounds, when showing, then it lists each one with the total and the average", func(t *testing.T) {
		content := "2026-11-21-0900\t10\t100\t5\t1.5\t1000\t200\n2026-11-22-1030\t20\t51\t3\t0.5\t500\t100\n"
		cfg, out := metricsIn(t, &content)
		if err := Metrics(cfg, out); err != nil {
			t.Fatal(err)
		}
		expected := "rodada\tfalas\tsegundos\tturnos\tcusto_usd\ttokens_entrada\ttokens_saida\n" +
			content +
			"\n2 rodadas, 151 s no total (média 75 s), US$ 2.00 no total (média US$ 1.00)\n" +
			"tokens: 1500 de entrada, 300 de saída\n"
		if out.String() != expected {
			t.Fatalf("expected\n%s\ngot\n%s", expected, out.String())
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

package main

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/malaquiasdev/rangers/internal/gate"
)

func TestMain(m *testing.M) {
	for _, pair := range os.Environ() {
		name, _, _ := strings.Cut(pair, "=")
		if strings.HasPrefix(name, "GIT_") {
			_ = os.Unsetenv(name)
		}
	}
	os.Exit(m.Run())
}

func TestReportGate(t *testing.T) {
	t.Run("given an allowed gate, when reporting, then the exit code is 0", func(t *testing.T) {
		if got := reportGate(nil); got != 0 {
			t.Fatalf("got %d", got)
		}
	})

	t.Run("given a refused gate, when reporting, then the exit code is 1", func(t *testing.T) {
		if got := reportGate(fmt.Errorf("verdict: %w", gate.ErrRefused)); got != 1 {
			t.Fatalf("got %d", got)
		}
	})

	t.Run("given an error, when reporting, then the exit code is 2", func(t *testing.T) {
		if got := reportGate(errors.New("gh: not found")); got != 2 {
			t.Fatalf("got %d", got)
		}
	})
}

func TestRunGate(t *testing.T) {
	t.Run("given gate with bad flags, when running, then the exit code is 2", func(t *testing.T) {
		if got := run(t.Context(), []string{"gate", "--pr", "0"}); got != 2 {
			t.Fatalf("got %d", got)
		}
	})
}

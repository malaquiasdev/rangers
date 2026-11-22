package learn

import (
	"testing"
	"time"
)

func TestGates(t *testing.T) {
	now := time.Date(2026, 11, 22, 10, 0, 0, 0, time.UTC)

	t.Run("given a checkpoint 30 minutes ago and a 2 hour interval, when checking the interval gate, then it blocks with the reason", func(t *testing.T) {
		reason := blocksInterval(2, now.Add(-30*time.Minute), now)
		expected := "a última rodada foi há 30 min, menos que o intervalo de 2h: não rodo"
		if reason != expected {
			t.Fatalf("expected %q, got %q", expected, reason)
		}
	})

	t.Run("given a checkpoint 2 hours ago and a 2 hour interval, when checking the interval gate, then it lets the round through", func(t *testing.T) {
		if reason := blocksInterval(2, now.Add(-2*time.Hour), now); reason != "" {
			t.Fatalf("got %q", reason)
		}
	})

	t.Run("given 9 lines and a minimum of 10, when checking the lines gate, then it blocks with the reason", func(t *testing.T) {
		reason := blocksLines(10, 9)
		expected := "só 9 falas novas, abaixo do mínimo de 10: não rodo o modelo"
		if reason != expected {
			t.Fatalf("expected %q, got %q", expected, reason)
		}
	})

	t.Run("given 10 lines and a minimum of 10, when checking the lines gate, then it lets the round through", func(t *testing.T) {
		if reason := blocksLines(10, 10); reason != "" {
			t.Fatalf("got %q", reason)
		}
	})
}

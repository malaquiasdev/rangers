package learn

import (
	"fmt"
	"math"
	"strconv"
	"time"
)

func blocksInterval(intervalHours float64, checkpoint, now time.Time) string {
	age := now.Sub(checkpoint)
	limit := time.Duration(intervalHours * float64(time.Hour))
	if age >= limit {
		return ""
	}
	hours := strconv.FormatFloat(intervalHours, 'f', -1, 64)
	return fmt.Sprintf("a última rodada foi há %d min, menos que o intervalo de %sh: não rodo", int(age.Minutes()), hours)
}

func blocksLines(minimum, lines int) string {
	if lines >= minimum {
		return ""
	}
	return fmt.Sprintf("só %d falas novas, abaixo do mínimo de %d: não rodo o modelo", lines, minimum)
}

func blocksDaily(dailyBudget, spent float64) string {
	if spent < dailyBudget {
		return ""
	}
	return fmt.Sprintf("teto diário de US$ %.2f atingido (US$ %.2f gastos hoje): não rodo o modelo", dailyBudget, spent)
}

func roundBudget(budget, dailyBudget, spent float64) float64 {
	balance := math.Round((dailyBudget-spent)*1000) / 1000
	return math.Min(budget, balance)
}

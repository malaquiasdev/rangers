package gate

import (
	"slices"
	"strings"
	"unicode/utf8"
)

const ellipsis = "..."

func fit(lines []string, limit int) string {
	text := strings.Join(lines, "\n")
	excess := utf8.RuneCountInString(text) - limit
	if excess <= 0 {
		return text
	}
	longest := slices.MaxFunc(lines, func(a, b string) int { return len(a) - len(b) })
	runes := []rune(longest)
	keep := max(len(runes)-excess-len(ellipsis), 0)
	cut := string(runes[:keep]) + ellipsis
	return strings.Replace(text, longest, cut, 1)
}

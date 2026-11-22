package schedule

import "strings"

var statusKeys = []struct{ key, label string }{
	{"state", "estado"},
	{"runs", "rodadas"},
	{"last exit code", "último código de saída"},
	{"run interval", "intervalo"},
}

func describe(output string) string {
	found := map[string]string{}
	for _, line := range strings.Split(output, "\n") {
		key, value, ok := strings.Cut(line, "=")
		key = strings.TrimSpace(key)
		if _, seen := found[key]; ok && !seen {
			found[key] = strings.TrimSpace(value)
		}
	}
	var rows []string
	for _, item := range statusKeys {
		if value, ok := found[item.key]; ok {
			rows = append(rows, item.label+": "+value)
		}
	}
	if len(rows) == 0 {
		return "agendado, mas não consegui ler o estado"
	}
	return strings.Join(rows, "\n")
}

package learn

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

const agentPrefix = "Learning."

func findAgent(folder string) (string, error) {
	files, err := filepath.Glob(filepath.Join(folder, "*.md"))
	if err != nil {
		return "", err
	}
	sort.Strings(files)
	for _, file := range files {
		raw, err := os.ReadFile(file)
		if err != nil {
			return "", err
		}
		name, description := header(string(raw))
		if name != "" && strings.HasPrefix(description, agentPrefix) {
			return name, nil
		}
	}
	return "", fmt.Errorf("nenhum agente com description começando em %q em %s", agentPrefix, folder)
}

func header(content string) (name, description string) {
	rows := strings.Split(content, "\n")
	if len(rows) == 0 || rows[0] != "---" {
		return "", ""
	}
	for _, row := range rows[1:] {
		if row == "---" {
			break
		}
		if value, ok := strings.CutPrefix(row, "name: "); ok {
			name = strings.TrimSpace(value)
		}
		if value, ok := strings.CutPrefix(row, "description: "); ok {
			description = strings.TrimSpace(value)
		}
	}
	return name, description
}

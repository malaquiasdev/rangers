package learn

import (
	"path/filepath"
	"strings"
)

const (
	notesBase  = "Read,Write,Edit,Glob,Grep"
	auditTilde = "~/.claude/bin/audit"
)

func notesAllowed(home string) string {
	forms := []string{auditTilde, filepath.Join(home, ".claude", "bin", "audit")}
	rules := []string{notesBase}
	for _, form := range forms {
		rules = append(rules, "Bash("+form+":*)")
	}
	return strings.Join(rules, ",")
}

package learn

import (
	"path/filepath"
	"slices"
	"strings"
)

const (
	notesBase  = "Read,Write,Edit,Glob,Grep"
	auditTilde = "~/.claude/bin/audit"
)

func notesAllowed(home string) string {
	path := filepath.Join(home, ".claude", "bin", "audit")
	forms := []string{auditTilde, path}
	resolved, err := filepath.EvalSymlinks(path)
	if err == nil && !slices.Contains(forms, resolved) {
		forms = append(forms, resolved)
	}
	rules := []string{notesBase}
	for _, form := range forms {
		rules = append(rules, "Bash("+form+":*)")
	}
	return strings.Join(rules, ",")
}

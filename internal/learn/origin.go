package learn

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const originHeading = "Origin:"

var originNote = filepath.Join("claude", "estudos", "origem-das-regras-rangers.md")

func originSection(result string) string {
	lines := strings.Split(result, "\n")
	for i, line := range lines {
		if strings.TrimSpace(line) == originHeading {
			return strings.TrimSpace(strings.Join(lines[i+1:], "\n"))
		}
	}
	return ""
}

func (r *round) keepOrigin(ctx context.Context, result string) error {
	section := originSection(result)
	if section == "" || !r.brainWritable(ctx) {
		return nil
	}
	path := filepath.Join(r.cfg.Brain, originNote)
	if err := os.MkdirAll(filepath.Dir(path), 0o750); err != nil {
		return r.warnOrigin(err)
	}
	prefix := ""
	if info, err := os.Stat(path); err == nil && info.Size() > 0 {
		prefix = "\n"
	}
	entry := fmt.Sprintf("%s## %s\n\n%s\n", prefix, r.stamp, section)
	return r.warnOrigin(appendText(path, entry))
}

func (r *round) brainWritable(ctx context.Context) bool {
	return r.dep.Brain.Probe(ctx, r.cfg.Brain) == nil
}

func (r *round) warnOrigin(cause error) error {
	if cause == nil {
		return nil
	}
	notice := "aviso: não consegui anotar a origem das regras no brain (%v)\n"
	_, err := fmt.Fprintf(r.dep.Err, notice, cause)
	return err
}

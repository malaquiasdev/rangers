package gate

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"slices"
	"strings"
	"time"
	"unicode"

	"github.com/malaquiasdev/rangers/internal/transcript"
)

const (
	approvalWindow = 48 * time.Hour
	hashLength     = 8
	shortLength    = 7
)

var answers = []string{"yes", "sim", "ok", "pode"}

func approvalHash(repo string, pr int, head string) string {
	sum := sha256.Sum256([]byte(fmt.Sprintf("%s#%d@%s", repo, pr, head)))
	return hex.EncodeToString(sum[:])[:hashLength]
}

func affirmative(text string) bool {
	trimmed := strings.TrimRightFunc(strings.TrimSpace(text), func(r rune) bool {
		return unicode.IsPunct(r) || unicode.IsSpace(r)
	})
	return slices.Contains(answers, strings.ToLower(trimmed))
}

func approved(lines []transcript.Line, token string) bool {
	asked := map[string]bool{}
	for _, line := range lines {
		if line.Assistant {
			asked[line.Session] = asked[line.Session] || strings.Contains(line.Text, token)
			continue
		}
		if asked[line.Session] && affirmative(line.Text) {
			return true
		}
		asked[line.Session] = false
	}
	return false
}

func (g *gate) token() string {
	return "gate:" + approvalHash(g.opts.repo, g.opts.pr, g.pr.HeadRefOid)
}

func (g *gate) approval(_ context.Context) (outcome, error) {
	token := g.token()
	lines, err := g.dep.Lines(g.opts.transcripts, g.dep.Now().Add(-approvalWindow))
	if err != nil {
		return outcome{}, err
	}
	if !approved(lines, token) {
		return refuse("approval", "no yes from the human for "+token+"; ask with rangers gate ask"), nil
	}
	return pass("approval", "the human said yes to "+token), nil
}

func (g *gate) ask(ctx context.Context) error {
	if err := g.loadPR(ctx); err != nil {
		return err
	}
	token := g.token()
	_, err := fmt.Fprintf(g.dep.Out, "%s\nMerge %s#%d at %s? Answer sim. %s\n",
		strings.TrimPrefix(token, "gate:"), g.opts.repo, g.opts.pr, g.pr.short(), token)
	return err
}

package gate

import (
	"context"
	"fmt"
	"path/filepath"
	"strconv"
)

func (g *gate) merge(ctx context.Context) (outcome, error) {
	number := strconv.Itoa(g.opts.pr)
	_, err := g.dep.Runner.Output(ctx, "gh", "pr", "merge", number, "--repo", g.opts.repo,
		"--squash", "--match-head-commit", g.pr.HeadRefOid)
	if err != nil {
		return outcome{}, err
	}
	return pass("merge", "squash at "+g.pr.short()+g.audit(ctx)), nil
}

func (g *gate) audit(ctx context.Context) string {
	script := filepath.Join(g.home, ".claude", "bin", "audit")
	if !g.dep.Runner.Has(script) {
		return ""
	}
	link := fmt.Sprintf("https://github.com/%s/pull/%d", g.opts.repo, g.opts.pr)
	ci := "checks green"
	if g.ciNotApplicable {
		ci = "ci not applicable"
	}
	reason := fmt.Sprintf("rangers gate: %s, %s, squash at %s", g.token(), ci, g.pr.short())
	if _, err := g.dep.Runner.Output(ctx, script, g.opts.repo, "merge", link, reason); err != nil {
		return "; audit not recorded"
	}
	return "; audit recorded"
}

package gate

import (
	"context"
	"encoding/json"
	"fmt"
	"slices"
	"strconv"
	"strings"
)

const (
	prFields          = "state,isDraft,baseRefName,headRefOid,title,headRefName,labels"
	productionVerdict = "production needs the human's explicit yes; the gate stops here"
	releaseTitle      = "chore(main): release"
	releaseBranch     = "release-please--"
	productionLabel   = "production"
)

type label struct {
	Name string `json:"name"`
}

type pullRequest struct {
	State       string  `json:"state"`
	IsDraft     bool    `json:"isDraft"`
	BaseRefName string  `json:"baseRefName"`
	HeadRefOid  string  `json:"headRefOid"`
	Title       string  `json:"title"`
	HeadRefName string  `json:"headRefName"`
	Labels      []label `json:"labels"`
}

func (p *pullRequest) isProduction() bool {
	if strings.HasPrefix(p.Title, releaseTitle) || strings.HasPrefix(p.HeadRefName, releaseBranch) {
		return true
	}
	return slices.ContainsFunc(p.Labels, func(l label) bool {
		return strings.EqualFold(l.Name, productionLabel)
	})
}

func (g *gate) gh(ctx context.Context, into any, args ...string) error {
	out, err := g.dep.Runner.Output(ctx, "gh", args...)
	if err != nil {
		return err
	}
	return json.Unmarshal([]byte(out), into)
}

func (g *gate) loadPR(ctx context.Context) error {
	number := strconv.Itoa(g.opts.pr)
	return g.gh(ctx, &g.pr, "pr", "view", number, "--repo", g.opts.repo, "--json", prFields)
}

func (g *gate) defaultBranch(ctx context.Context) (string, error) {
	var repo struct {
		DefaultBranchRef struct {
			Name string `json:"name"`
		} `json:"defaultBranchRef"`
	}
	err := g.gh(ctx, &repo, "repo", "view", g.opts.repo, "--json", "defaultBranchRef")
	return repo.DefaultBranchRef.Name, err
}

func (g *gate) open(ctx context.Context) (outcome, error) {
	if err := g.loadPR(ctx); err != nil {
		return outcome{}, err
	}
	base, err := g.defaultBranch(ctx)
	if err != nil {
		return outcome{}, err
	}
	switch {
	case g.pr.State != "OPEN":
		return refuse("pr", "the pull request is "+strings.ToLower(g.pr.State)), nil
	case g.pr.IsDraft:
		return refuse("pr", "the pull request is a draft"), nil
	case g.pr.BaseRefName != base:
		reason := fmt.Sprintf("base is %s, not the default branch %s", g.pr.BaseRefName, base)
		return refuse("pr", reason), nil
	}
	return pass("pr", "open, ready, base "+base), nil
}

func (g *gate) production(_ context.Context) (outcome, error) {
	if g.pr.isProduction() {
		return refuse("production", productionVerdict), nil
	}
	return pass("production", "not a release"), nil
}

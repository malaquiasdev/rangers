package gate

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
)

const (
	bucketPass     = "pass"
	bucketSkipping = "skipping"

	noChecksReported = "no checks reported on the"
)

type check struct {
	Name   string `json:"name"`
	Bucket string `json:"bucket"`
}

func (g *gate) checks(ctx context.Context) ([]check, error) {
	number := strconv.Itoa(g.opts.pr)
	out, err := g.dep.Runner.Output(ctx, "gh", "pr", "checks", number,
		"--repo", g.opts.repo, "--json", "name,state,bucket")
	if strings.TrimSpace(out) == "" && err != nil && strings.Contains(err.Error(), noChecksReported) {
		return nil, nil
	}
	var checks []check
	if parseErr := json.Unmarshal([]byte(out), &checks); parseErr != nil {
		return nil, errors.Join(err, parseErr)
	}
	return checks, nil
}

func (g *gate) withoutChecks(ctx context.Context) (outcome, error) {
	runs, err := g.runsOnPullRequests(ctx)
	if err != nil {
		return outcome{}, err
	}
	if runs {
		return refuse("ci", "ci has no checks for this head"), nil
	}
	g.ciNotApplicable = true
	return pass("ci", "not applicable, no workflow runs on pull requests"), nil
}

func (g *gate) ci(ctx context.Context) (outcome, error) {
	checks, err := g.checks(ctx)
	if err != nil {
		return outcome{}, err
	}
	if len(checks) == 0 {
		return g.withoutChecks(ctx)
	}
	var failing []string
	for _, c := range checks {
		if c.Bucket != bucketPass && c.Bucket != bucketSkipping {
			failing = append(failing, c.Name+" ("+c.Bucket+")")
		}
	}
	if len(failing) > 0 {
		return refuse("ci", "ci not green: "+strings.Join(failing, ", ")), nil
	}
	return pass("ci", fmt.Sprintf("%d checks passed or skipped", len(checks))), nil
}

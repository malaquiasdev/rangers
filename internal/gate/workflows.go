package gate

import (
	"context"
	"encoding/json"
	"net/url"
	"path"
	"slices"
	"strings"

	"go.yaml.in/yaml/v3"
)

const (
	rawAccept      = "Accept: application/vnd.github.raw+json"
	notFoundMarker = "Not Found (HTTP 404)"
	globCharacters = "*?+[]!"
)

type workflowEntry struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

func (g *gate) runsOnPullRequests(ctx context.Context) (bool, error) {
	for _, ref := range []string{g.pr.BaseRefName, g.pr.HeadRefOid} {
		found, err := g.refRunsOnPullRequests(ctx, ref)
		if err != nil || found {
			return found, err
		}
	}
	return false, nil
}

func (g *gate) refRunsOnPullRequests(ctx context.Context, ref string) (bool, error) {
	names, err := g.workflowNames(ctx, ref)
	if err != nil {
		return false, err
	}
	for _, name := range names {
		body, err := g.workflowBody(ctx, ref, name)
		if err != nil {
			return false, err
		}
		if triggersOnPullRequest(body, g.pr.HeadRefName) {
			return true, nil
		}
	}
	return false, nil
}

func (g *gate) workflowsURL() string {
	return "repos/" + g.opts.repo + "/contents/.github/workflows"
}

func (g *gate) workflowNames(ctx context.Context, ref string) ([]string, error) {
	out, err := g.dep.Runner.Output(ctx, "gh", "api", g.workflowsURL()+refQuery(ref))
	if err != nil {
		if strings.Contains(err.Error(), notFoundMarker) {
			return nil, nil
		}
		return nil, err
	}
	var entries []workflowEntry
	if err := json.Unmarshal([]byte(out), &entries); err != nil {
		return nil, err
	}
	var names []string
	for _, entry := range entries {
		if entry.Type == "file" && isWorkflowFile(entry.Name) {
			names = append(names, entry.Name)
		}
	}
	return names, nil
}

func (g *gate) workflowBody(ctx context.Context, ref, name string) (string, error) {
	file := g.workflowsURL() + "/" + url.PathEscape(name) + refQuery(ref)
	return g.dep.Runner.Output(ctx, "gh", "api", "-H", rawAccept, file)
}

func refQuery(ref string) string {
	return "?ref=" + url.QueryEscape(ref)
}

func isWorkflowFile(name string) bool {
	ext := path.Ext(name)
	return ext == ".yml" || ext == ".yaml"
}

func triggersOnPullRequest(body, head string) bool {
	var doc struct {
		On any `yaml:"on"`
	}
	if err := yaml.Unmarshal([]byte(body), &doc); err != nil {
		return true
	}
	events, ok := eventsOf(doc.On)
	if !ok {
		return true
	}
	if _, found := events["pull_request"]; found {
		return true
	}
	if _, found := events["pull_request_target"]; found {
		return true
	}
	push, found := events["push"]
	return found && pushReachesBranch(push, head)
}

func eventsOf(on any) (map[string]any, bool) {
	switch value := on.(type) {
	case string:
		return map[string]any{value: nil}, true
	case []any:
		events := map[string]any{}
		for _, item := range value {
			name, ok := item.(string)
			if !ok {
				return nil, false
			}
			events[name] = nil
		}
		return events, true
	case map[string]any:
		return value, true
	}
	return nil, false
}

func pushReachesBranch(push any, head string) bool {
	if push == nil {
		return true
	}
	filters, ok := push.(map[string]any)
	if !ok {
		return true
	}
	if _, ignoring := filters["branches-ignore"]; ignoring {
		return true
	}
	branches, filtered := filters["branches"]
	if !filtered {
		return true
	}
	return slices.ContainsFunc(patternsOf(branches), func(pattern string) bool {
		return pattern == head || strings.ContainsAny(pattern, globCharacters)
	})
}

func patternsOf(branches any) []string {
	switch value := branches.(type) {
	case string:
		return []string{value}
	case []any:
		patterns := make([]string, 0, len(value))
		for _, item := range value {
			name, ok := item.(string)
			if !ok {
				return []string{"*"}
			}
			patterns = append(patterns, name)
		}
		return patterns
	}
	return []string{"*"}
}

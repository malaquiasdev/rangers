package learn

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var pendingBranch = regexp.MustCompile(`^Ramo (\S+) em `)

type branchLookup struct {
	git   Git
	repo  string
	known map[string]bool
}

func prunePending(ctx context.Context, cfg *Config, dep *Dependencies) error {
	path := filepath.Join(cfg.State, pendingName)
	cause := dropMissingBranches(ctx, path, cfg.Repo, dep.Git)
	if cause == nil {
		return nil
	}
	_, err := fmt.Fprintf(dep.Err,
		"aviso: não consegui limpar os lembretes pendentes (%v); %s ficou como estava\n", cause, path)
	return err
}

func dropMissingBranches(ctx context.Context, path, repo string, git Git) error {
	raw, err := os.ReadFile(path)
	if errors.Is(err, fs.ErrNotExist) {
		return nil
	}
	if err != nil {
		return err
	}
	kept, dropped, err := keepExisting(ctx, string(raw), repo, git)
	if err != nil || !dropped {
		return err
	}
	return replaceFile(path, kept)
}

func (l *branchLookup) exists(ctx context.Context, branch string) (bool, error) {
	if found, seen := l.known[branch]; seen {
		return found, nil
	}
	found, err := l.git.BranchExists(ctx, l.repo, branch)
	if err != nil {
		return false, err
	}
	l.known[branch] = found
	return found, nil
}

func keepExisting(ctx context.Context, content, repo string, git Git) (string, bool, error) {
	lookup := &branchLookup{git: git, repo: repo, known: map[string]bool{}}
	var kept strings.Builder
	dropped := false
	for _, line := range strings.SplitAfter(content, "\n") {
		found, err := lineSurvives(ctx, lookup, line)
		if err != nil {
			return "", false, err
		}
		if found {
			kept.WriteString(line)
		} else {
			dropped = true
		}
	}
	return kept.String(), dropped, nil
}

func lineSurvives(ctx context.Context, lookup *branchLookup, line string) (bool, error) {
	branch := pendingBranchOf(line)
	if branch == "" {
		return true, nil
	}
	return lookup.exists(ctx, branch)
}

func pendingBranchOf(line string) string {
	_, body, found := strings.Cut(line, "\t")
	if !found {
		return ""
	}
	match := pendingBranch.FindStringSubmatch(body)
	if match == nil {
		return ""
	}
	return match[1]
}

func replaceFile(path, content string) error {
	file, err := os.CreateTemp(filepath.Dir(path), ".pending-reminders-*")
	if err != nil {
		return err
	}
	if _, err = file.WriteString(content); err != nil {
		return errors.Join(err, file.Close(), os.Remove(file.Name()))
	}
	if err = file.Close(); err != nil {
		return errors.Join(err, os.Remove(file.Name()))
	}
	if err = os.Rename(file.Name(), path); err != nil {
		return errors.Join(err, os.Remove(file.Name()))
	}
	return nil
}

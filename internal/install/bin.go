package install

import (
	"bytes"
	"path/filepath"
	"slices"
)

const (
	binLinked binState = iota
	binLinkable
	binDiffers
)

var (
	binStatus = map[binState]string{
		binLinked:   "linked",
		binLinkable: "NOT LINKED",
		binDiffers:  "differs from the installed file, left alone",
	}

	binAliases = []string{"agent-watch", "audit", "complete-reminders", "stop-push-check"}
)

type binState int

func (in *installer) binLinks() ([]link, error) {
	matches, err := in.files.Glob(filepath.Join(in.opts.repo, "bin", "*"))
	if err != nil {
		return nil, err
	}
	links := make([]link, 0, len(matches)+len(binAliases))
	for _, match := range matches {
		if slices.Contains(binAliases, filepath.Base(match)) && in.binary != "" {
			continue
		}
		path := filepath.Join(in.opts.claudeDir, "bin", filepath.Base(match))
		links = append(links, link{target: match, path: path, guarded: true})
	}
	return in.withAliases(links), nil
}

func (in *installer) binState(l link) (binState, error) {
	if current, err := in.files.Readlink(l.path); err == nil && current == l.target {
		return binLinked, nil
	}
	if !in.files.IsRegular(l.path) {
		return binLinkable, nil
	}
	same, err := in.sameBytes(l.path, l.target)
	if err != nil {
		return binDiffers, err
	}
	if same {
		return binLinkable, nil
	}
	return binDiffers, nil
}

func (in *installer) sameBytes(a, b string) (bool, error) {
	left, err := in.files.ReadFile(a)
	if err != nil {
		return false, err
	}
	right, err := in.files.ReadFile(b)
	if err != nil {
		return false, err
	}
	return bytes.Equal(left, right), nil
}

func (in *installer) checkBins() (int, error) {
	links, err := in.binLinks()
	if err != nil {
		return 0, err
	}
	missing := 0
	for _, l := range links {
		state, stateErr := in.binState(l)
		if stateErr != nil {
			return 0, stateErr
		}
		in.out.row("bin/"+filepath.Base(l.path), binStatus[state])
		if state == binLinkable {
			missing++
		}
	}
	return missing, nil
}

func (in *installer) withAliases(links []link) []link {
	if in.binary == "" {
		return links
	}
	aliases := make([]link, 0, len(binAliases))
	for _, name := range binAliases {
		path := filepath.Join(in.opts.claudeDir, "bin", name)
		aliases = append(aliases, link{target: in.binary, path: path, guarded: true})
	}
	return append(links, aliases...)
}

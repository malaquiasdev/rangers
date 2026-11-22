package install

import "path/filepath"

const (
	skillLinked skillState = iota
	skillLinkable
	skillBlocked
)

var skillStatus = map[skillState]string{
	skillLinked:   "linked",
	skillLinkable: "NOT LINKED",
	skillBlocked:  "exists and is not a link, left alone",
}

type skillState int

func (in *installer) skillLinks() ([]link, error) {
	matches, err := in.files.Glob(filepath.Join(in.opts.repo, "skills", "*", "SKILL.md"))
	if err != nil {
		return nil, err
	}
	links := make([]link, 0, len(matches))
	for _, match := range matches {
		folder := filepath.Dir(match)
		path := filepath.Join(in.opts.claudeDir, "skills", filepath.Base(folder))
		links = append(links, link{target: folder, path: path, skill: true})
	}
	return links, nil
}

func (in *installer) skillState(l link) skillState {
	if current, err := in.files.Readlink(l.path); err == nil {
		if current == l.target {
			return skillLinked
		}
		return skillLinkable
	}
	if in.files.Exists(l.path) {
		return skillBlocked
	}
	return skillLinkable
}

package install

import (
	"fmt"
	"path/filepath"
)

type link struct {
	target, path string
	guarded      bool
	skill        bool
}

func (in *installer) link() error {
	links, err := in.links()
	if err != nil {
		return err
	}
	linked := 0
	for _, l := range links {
		placed, placeErr := in.place(l)
		if placeErr != nil {
			return placeErr
		}
		if placed {
			linked++
		}
	}
	in.out.row("links", fmt.Sprintf("%d linked into %s", linked, in.opts.claudeDir))
	return nil
}

func (in *installer) place(l link) (bool, error) {
	if l.skill {
		return in.placeSkill(l)
	}
	if !l.guarded {
		return true, in.files.Symlink(l.target, l.path)
	}
	state, err := in.binState(l)
	if err != nil {
		return false, err
	}
	switch state {
	case binDiffers:
		in.out.linef("bin/%s differs from the installed file, left alone", filepath.Base(l.path))
		return false, nil
	case binLinked:
		return true, nil
	}
	return true, in.files.Symlink(l.target, l.path)
}

func (in *installer) placeSkill(l link) (bool, error) {
	switch in.skillState(l) {
	case skillBlocked:
		in.out.linef("skills/%s %s", filepath.Base(l.path), skillStatus[skillBlocked])
		return false, nil
	case skillLinked:
		return true, nil
	}
	return true, in.files.Symlink(l.target, l.path)
}

func (in *installer) links() ([]link, error) {
	links, err := in.docLinks()
	if err != nil {
		return nil, err
	}
	skills, err := in.skillLinks()
	if err != nil {
		return nil, err
	}
	links = append(links, skills...)
	bins, err := in.binLinks()
	if err != nil {
		return nil, err
	}
	return append(links, bins...), nil
}

func (in *installer) docLinks() ([]link, error) {
	var links []link
	for _, folder := range []string{"agents", "rules"} {
		matches, err := in.files.Glob(filepath.Join(in.opts.repo, folder, "*.md"))
		if err != nil {
			return nil, err
		}
		for _, match := range matches {
			path := filepath.Join(in.opts.claudeDir, folder, filepath.Base(match))
			links = append(links, link{target: match, path: path})
		}
	}
	return links, nil
}

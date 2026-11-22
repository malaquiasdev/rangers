package install

import (
	"fmt"
	"path/filepath"
)

type link struct{ target, path string }

func (in *installer) link() error {
	links, err := in.links()
	if err != nil {
		return err
	}
	for _, l := range links {
		if err = in.files.Symlink(l.target, l.path); err != nil {
			return err
		}
	}
	in.out.row("links", fmt.Sprintf("%d linked into %s", len(links), in.opts.claudeDir))
	return nil
}

func (in *installer) links() ([]link, error) {
	var links []link
	for _, folder := range []string{"agents", "rules"} {
		matches, err := in.files.Glob(filepath.Join(in.opts.repo, folder, "*.md"))
		if err != nil {
			return nil, err
		}
		for _, match := range matches {
			path := filepath.Join(in.opts.claudeDir, folder, filepath.Base(match))
			links = append(links, link{match, path})
		}
	}
	skill := filepath.Join(in.opts.repo, "skills", "rangers")
	if in.files.Exists(skill) {
		links = append(links, link{skill, filepath.Join(in.opts.claudeDir, "skills", "rangers")})
	}
	return links, nil
}

package install

import (
	"context"
	"fmt"
	"path/filepath"
)

const doneMessage = "\nDone. Restart Claude Code and run /pstack:setup-pstack once."

func (in *installer) install(ctx context.Context, deps []Dependency) error {
	for _, d := range deps {
		if err := in.installOne(ctx, d); err != nil {
			return fmt.Errorf("%s: %w", d.Name, err)
		}
	}
	if err := in.link(); err != nil {
		return err
	}
	in.out.linef(doneMessage)
	return in.out.err
}

func (in *installer) installOne(ctx context.Context, d Dependency) error {
	if d.Optional && !in.opts.optional {
		in.out.row(d.Name, "skipped (optional)")
		return nil
	}
	if version := in.version(ctx, d); version != "" {
		in.out.row(d.Name, version)
		return nil
	}
	in.out.linef("installing %s", d.Name)
	if err := in.add(ctx, d); err != nil {
		return err
	}
	in.out.row(d.Name, installed)
	return nil
}

func (in *installer) add(ctx context.Context, d Dependency) error {
	switch d.Kind {
	case kindPlugin:
		return in.addPlugin(ctx, d)
	case kindNPM:
		return in.runner.Stream(ctx, "npm", "install", "-g", d.Package+"@"+d.Version)
	case kindBrew:
		return in.runner.Stream(ctx, "brew", "install", d.Formula)
	case kindImage:
		return in.runner.Stream(ctx, "podman", "pull", d.Image)
	}
	return in.addSkill(ctx, d)
}

func (in *installer) addPlugin(ctx context.Context, d Dependency) error {
	market := []string{"plugin", "marketplace", "add", d.Marketplace}
	if err := in.runner.Stream(ctx, "claude", market...); err != nil {
		return err
	}
	return in.runner.Stream(ctx, "claude", "plugin", "install", d.Plugin)
}

func (in *installer) addSkill(ctx context.Context, d Dependency) error {
	tmp, err := in.files.MkdirTemp()
	if err != nil {
		return err
	}
	defer func() { _ = in.files.RemoveAll(tmp) }()
	url := "https://github.com/" + d.Repository + ".git"
	clone := []string{"clone", "--quiet", "--no-checkout", url, tmp}
	if _, err = in.runner.Output(ctx, "git", clone...); err != nil {
		return err
	}
	if _, err = in.runner.Output(ctx, "git", "-C", tmp, "checkout", "--quiet", d.Commit); err != nil {
		return err
	}
	return in.copySkill(d, filepath.Join(tmp, d.Path))
}

func (in *installer) copySkill(d Dependency, source string) error {
	dst := filepath.Join(in.opts.claudeDir, "skills", d.Name)
	if err := in.files.CopyDir(source, dst); err != nil {
		_ = in.files.RemoveAll(dst)
		return err
	}
	return in.files.WriteFile(filepath.Join(dst, marker), []byte(d.Commit+"\n"))
}

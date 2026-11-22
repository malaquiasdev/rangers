package install

import (
	"context"
	"path/filepath"
	"strings"
)

const (
	marker      = ".rangers-commit"
	shortCommit = 12
	installed   = "installed"
)

func (in *installer) version(ctx context.Context, d Dependency) string {
	switch d.Kind {
	case kindPlugin:
		return in.pluginVersion(ctx, d.Plugin)
	case kindNPM:
		return in.commandVersion(ctx, d.Name, "--version")
	case kindBrew:
		return in.commandVersion(ctx, d.Name, "version")
	case kindImage:
		return in.imageVersion(ctx, d.Image)
	}
	return in.skillVersion(d)
}

func (in *installer) pluginVersion(ctx context.Context, plugin string) string {
	out, err := in.runner.Output(ctx, "claude", "plugin", "list")
	if err != nil {
		return ""
	}
	found := false
	for line := range strings.SplitSeq(out, "\n") {
		found = found || strings.Contains(line, plugin)
		if fields := strings.Fields(line); found && len(fields) == 2 && fields[0] == "Version:" {
			return fields[1]
		}
	}
	if found {
		return installed
	}
	return ""
}

func (in *installer) commandVersion(ctx context.Context, name, flag string) string {
	if !in.runner.Has(name) {
		return ""
	}
	out, err := in.runner.Output(ctx, name, flag)
	if line, _, _ := strings.Cut(strings.TrimSpace(out), "\n"); err == nil && line != "" {
		return line
	}
	return installed
}

func (in *installer) imageVersion(ctx context.Context, image string) string {
	if _, err := in.runner.Output(ctx, "podman", "image", "exists", image); err != nil {
		return ""
	}
	if _, tag, ok := strings.Cut(image, ":"); ok {
		return tag
	}
	return installed
}

func (in *installer) skillVersion(d Dependency) string {
	dir := filepath.Join(in.opts.claudeDir, "skills", d.Name)
	if !in.files.Exists(dir) {
		return ""
	}
	raw, err := in.files.ReadFile(filepath.Join(dir, marker))
	commit := strings.TrimSpace(string(raw))
	if err != nil || commit == "" {
		return "present, unpinned"
	}
	return commit[:min(shortCommit, len(commit))]
}

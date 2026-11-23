package agentwatch

import (
	"context"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

var ignoredNames = map[string]bool{"node_modules": true, ".git": true, "vendor": true}

type Git interface {
	LastCommit(ctx context.Context, dir string) (int64, bool)
}

type GitCLI struct{}

func (GitCLI) LastCommit(ctx context.Context, dir string) (int64, bool) {
	out, err := exec.CommandContext(ctx, "git", "-C", dir, "log", "-1", "--format=%ct").Output()
	if err != nil {
		return 0, false
	}
	stamp, err := strconv.ParseInt(strings.TrimSpace(string(out)), 10, 64)
	return stamp, err == nil
}

func (w *watcher) lastActivity(ctx context.Context, t target) int64 {
	var newest int64
	if isFile(t.file) {
		newest = modified(t.file)
	}
	for _, dir := range t.dirs {
		newest = max(newest, w.newestInDir(ctx, dir))
	}
	return newest
}

func (w *watcher) newestInDir(ctx context.Context, dir string) int64 {
	if info, err := os.Stat(dir); err != nil || !info.IsDir() {
		return 0
	}
	newest := newestFile(dir)
	if commit, ok := w.dep.Git.LastCommit(ctx, dir); ok {
		newest = max(newest, commit)
	}
	return newest
}

func newestFile(dir string) int64 {
	var newest int64
	_ = filepath.WalkDir(dir, func(_ string, entry fs.DirEntry, err error) error {
		switch {
		case err != nil:
			return nil
		case ignoredNames[entry.Name()] && entry.IsDir():
			return filepath.SkipDir
		case ignoredNames[entry.Name()] || !entry.Type().IsRegular():
			return nil
		}
		if info, infoErr := entry.Info(); infoErr == nil {
			newest = max(newest, info.ModTime().Unix())
		}
		return nil
	})
	return newest
}

func isFile(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.Mode().IsRegular()
}

func modified(path string) int64 {
	info, err := os.Stat(path)
	if err != nil {
		return 0
	}
	return info.ModTime().Unix()
}

package check

import (
	"bytes"
	"fmt"
	"io/fs"
	"strings"
)

const (
	binaryProbe = 8000
	gitDir      = ".git"
	serenaDir   = ".serena"
	readmeFile  = "README.md"
)

type scan struct {
	roots     []string
	skipDirs  map[string]bool
	skipFiles map[string]bool
	suffix    string
	matches   func(line string) bool
}

func (c *checker) search(s scan) []string {
	var found []string
	for _, root := range s.roots {
		_ = fs.WalkDir(c.dep.Root, root, func(path string, entry fs.DirEntry, err error) error {
			switch {
			case err != nil:
				return fs.SkipDir
			case entry.IsDir() && s.skipDirs[entry.Name()]:
				return fs.SkipDir
			case s.wants(entry):
				found = append(found, c.grep(s, root, path)...)
			}
			return nil
		})
	}
	return found
}

func (s scan) wants(entry fs.DirEntry) bool {
	name := entry.Name()
	return entry.Type().IsRegular() && !s.skipFiles[name] && strings.HasSuffix(name, s.suffix)
}

func (c *checker) grep(s scan, root, path string) []string {
	data, err := fs.ReadFile(c.dep.Root, path)
	if err != nil || bytes.IndexByte(data[:min(len(data), binaryProbe)], 0) >= 0 {
		return nil
	}
	prefix := ""
	if root == "." {
		prefix = "./"
	}
	lines := strings.Split(strings.TrimSuffix(string(data), "\n"), "\n")
	var found []string
	for number, line := range lines {
		if s.matches(line) {
			found = append(found, fmt.Sprintf("%s%s:%d:%s", prefix, path, number+1, line))
		}
	}
	return found
}

func (c *checker) report(found []string, message string) {
	for _, line := range found {
		c.say("%s\n", line)
	}
	if len(found) > 0 {
		c.fail("%s", message)
	}
}

func set(names ...string) map[string]bool {
	members := make(map[string]bool, len(names))
	for _, name := range names {
		members[name] = true
	}
	return members
}

func vcsDirs() map[string]bool {
	return set(gitDir, serenaDir)
}

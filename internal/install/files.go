package install

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

const gitDir = ".git"

type OSFiles struct{}

func (OSFiles) Exists(path string) bool {
	_, err := os.Lstat(path)
	return err == nil
}

func (OSFiles) ReadFile(path string) ([]byte, error) {
	return os.ReadFile(path)
}

func (OSFiles) WriteFile(path string, data []byte) error {
	return os.WriteFile(path, data, 0o600)
}

func (OSFiles) Glob(pattern string) ([]string, error) {
	return filepath.Glob(pattern)
}

func (OSFiles) MkdirTemp() (string, error) {
	return os.MkdirTemp("", "rangers-skill-")
}

func (OSFiles) RemoveAll(path string) error {
	return os.RemoveAll(path)
}

func (OSFiles) Symlink(target, link string) error {
	if current, err := os.Readlink(link); err == nil && current == target {
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(link), 0o750); err != nil {
		return err
	}
	if info, err := os.Lstat(link); err == nil {
		if info.IsDir() {
			return fmt.Errorf("%s is a directory, not replacing it", link)
		}
		if err = os.Remove(link); err != nil {
			return err
		}
	}
	return os.Symlink(target, link)
}

func (OSFiles) CopyDir(src, dst string) error {
	return filepath.WalkDir(src, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.Name() == gitDir {
			return skipGit(entry)
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		return copyEntry(path, filepath.Join(dst, rel), entry)
	})
}

func skipGit(entry fs.DirEntry) error {
	if entry.IsDir() {
		return filepath.SkipDir
	}
	return nil
}

func copyEntry(from, to string, entry fs.DirEntry) error {
	info, err := entry.Info()
	if err != nil {
		return err
	}
	switch {
	case entry.IsDir():
		return os.MkdirAll(to, info.Mode().Perm()|0o700)
	case info.Mode()&fs.ModeSymlink != 0:
		target, linkErr := os.Readlink(from)
		if linkErr != nil {
			return linkErr
		}
		return os.Symlink(target, to)
	}
	data, err := os.ReadFile(from)
	if err != nil {
		return err
	}
	return os.WriteFile(to, data, info.Mode().Perm())
}

package install

import "context"

type Runner interface {
	Has(name string) bool
	Output(ctx context.Context, name string, args ...string) (string, error)
	Stream(ctx context.Context, name string, args ...string) error
}

type Files interface {
	Exists(path string) bool
	ReadFile(path string) ([]byte, error)
	WriteFile(path string, data []byte) error
	Glob(pattern string) ([]string, error)
	Symlink(target, link string) error
	Readlink(path string) (string, error)
	IsRegular(path string) bool
	CopyDir(src, dst string) error
	MkdirTemp() (string, error)
	RemoveAll(path string) error
}

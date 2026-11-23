package install

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

const (
	repoDir   = "/repo"
	homeDir   = "/home/x"
	claudeDir = "/home/x/.claude"
	pinned    = "abeeaf95660ba9e8b7c753eb38f441ca707a8823"

	pluginList = `Installed plugins:

  ❯ i-have-adhd@i-have-adhd
    Version: 0.4.1
    Scope: user
    Status: ✔ enabled

  ❯ pstack@cursor-plugins
    Version: 0.15.46
    Scope: user
    Status: ✔ enabled
`
)

type reply struct {
	out string
	err error
}

type fakeRunner struct {
	present map[string]bool
	replies map[string]reply
	streams map[string]error
	calls   []string
}

type memFiles struct {
	files     map[string]string
	links     map[string]string
	removed   []string
	written   []string
	copyErr   error
	linkErr   error
	symlinked []string
	tempDirs  int
}

type setup struct {
	t      *testing.T
	runner *fakeRunner
	files  *memFiles
	out    *bytes.Buffer
	dir    string
	binary string
}

func (f *fakeRunner) Has(name string) bool {
	return f.present[name]
}

func (f *fakeRunner) Output(_ context.Context, name string, args ...string) (string, error) {
	cmd := words(name, args)
	f.calls = append(f.calls, cmd)
	if r, ok := f.replies[cmd]; ok {
		return r.out, r.err
	}
	if name == "git" {
		return "", nil
	}
	return "", errors.New("no reply for " + cmd)
}

func (f *fakeRunner) Stream(_ context.Context, name string, args ...string) error {
	cmd := words(name, args)
	f.calls = append(f.calls, "stream "+cmd)
	return f.streams[cmd]
}

func (m *memFiles) Exists(path string) bool {
	if _, ok := m.files[path]; ok {
		return true
	}
	if _, ok := m.links[path]; ok {
		return true
	}
	for name := range m.files {
		if strings.HasPrefix(name, path+"/") {
			return true
		}
	}
	return false
}

func (m *memFiles) ReadFile(path string) ([]byte, error) {
	content, ok := m.files[path]
	if !ok {
		return nil, errors.New("not found: " + path)
	}
	return []byte(content), nil
}

func (m *memFiles) WriteFile(path string, data []byte) error {
	m.files[path] = string(data)
	m.written = append(m.written, path)
	return nil
}

func (m *memFiles) Glob(pattern string) ([]string, error) {
	var found []string
	for name := range m.files {
		if ok, err := filepath.Match(pattern, name); err == nil && ok {
			found = append(found, name)
		}
	}
	sort.Strings(found)
	return found, nil
}

func (m *memFiles) Symlink(target, link string) error {
	if m.linkErr != nil {
		return m.linkErr
	}
	delete(m.files, link)
	m.links[link] = target
	m.symlinked = append(m.symlinked, link)
	return nil
}

func (m *memFiles) Readlink(path string) (string, error) {
	target, ok := m.links[path]
	if !ok {
		return "", errors.New("not a link: " + path)
	}
	return target, nil
}

func (m *memFiles) IsRegular(path string) bool {
	_, ok := m.files[path]
	return ok
}

func (m *memFiles) CopyDir(src, dst string) error {
	if m.copyErr != nil {
		return m.copyErr
	}
	for name, content := range m.files {
		if rest, ok := strings.CutPrefix(name, src+"/"); ok {
			m.files[dst+"/"+rest] = content
		}
	}
	return nil
}

func (m *memFiles) MkdirTemp() (string, error) {
	m.tempDirs++
	return fmt.Sprintf("/tmp/skill-%d", m.tempDirs), nil
}

func (m *memFiles) RemoveAll(path string) error {
	m.removed = append(m.removed, path)
	for name := range m.files {
		if name == path || strings.HasPrefix(name, path+"/") {
			delete(m.files, name)
		}
	}
	return nil
}

func newSetup(t *testing.T) *setup {
	t.Helper()
	return &setup{
		t:      t,
		runner: &fakeRunner{present: map[string]bool{}, replies: map[string]reply{}, streams: map[string]error{}},
		files: &memFiles{
			files: map[string]string{repoDir + "/deps.toml": sampleManifest},
			links: map[string]string{},
		},
		out: &bytes.Buffer{},
		dir: repoDir,
	}
}

func (s *setup) run(args ...string) error {
	env := func(name string) string {
		if name == "HOME" {
			return homeDir
		}
		return ""
	}
	dep := &Dependencies{Runner: s.runner, Files: s.files, Out: s.out, Dir: s.dir, Binary: s.binary}
	return Command(s.t.Context(), args, env, dep)
}

func (s *setup) allInstalled() {
	s.runner.present["openspec"] = true
	s.runner.present["container-use"] = true
	s.runner.replies["claude plugin list"] = reply{out: pluginList}
	s.runner.replies["openspec --version"] = reply{out: "1.7.0\n"}
	s.runner.replies["container-use version"] = reply{out: "container-use version 0.4.2\ncommit: 89ebe77\n"}
	s.runner.replies["podman image exists docker.io/lycheeverse/lychee:0.24.2"] = reply{}
	s.runner.replies["docker info"] = reply{out: "Server: ok"}
	s.files.files[claudeDir+"/skills/grilling/.rangers-commit"] = pinned + "\n"
}

func (s *setup) typed() string {
	return s.out.String()
}

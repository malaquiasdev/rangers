package check

import (
	"bytes"
	"context"
	"io/fs"
	"strings"
	"testing"
	"testing/fstest"
)

const (
	validAgentFormat = "---\nname: %s\ndescription: A role\nmodel: sonnet\neffort: high\n---\n\nCharacter: Someone\n"
	validSkill       = "---\nname: pilot\ndescription: A skill\n---\n"
	validDeps        = "format_version = 1\n\n[[dependency]]\nname = \"a\"\nkind = \"npm\"\nlicense = \"MIT\"\nfunction = \"does a thing\"\n"
	names            = "Foo|Bar Baz\n"
	repoDir          = "/repo"
	homeDir          = "/home/x"
	privacyDefault   = "/home/x/.config/rangers/privacy.txt"
	privacyWarning   = "aviso: sem /home/x/.config/rangers/privacy.txt, a checagem de nomes de empresa foi pulada\n"
)

type fakeRunner struct {
	calls   []string
	envs    map[string][]string
	quiet   map[string]bool
	failing map[string]bool
	outputs map[string]string
}

type harness struct {
	files  fstest.MapFS
	runner *fakeRunner
	env    map[string]string
	host   map[string]string
	out    bytes.Buffer
}

func (f *fakeRunner) Run(_ context.Context, call Invocation) error {
	key := strings.Join(append([]string{call.Name}, call.Args...), " ")
	f.calls = append(f.calls, key)
	f.envs[key] = call.Env
	f.quiet[key] = call.Quiet
	if f.failing[key] {
		return fs.ErrInvalid
	}
	return nil
}

func (f *fakeRunner) Output(_ context.Context, name string, args ...string) (string, error) {
	key := strings.Join(append([]string{name}, args...), " ")
	f.calls = append(f.calls, key)
	return f.outputs[key], nil
}

func clean() fstest.MapFS {
	return fstest.MapFS{
		"agents/alpha.md":        {Data: []byte(agent("alpha"))},
		"skills/pilot/SKILL.md":  {Data: []byte(validSkill)},
		"deps.toml":              {Data: []byte(validDeps)},
		"scripts/characters.txt": {Data: []byte(names)},
	}
}

func agent(name string) string {
	return strings.Replace(validAgentFormat, "%s", name, 1)
}

func file(content string) *fstest.MapFile {
	return &fstest.MapFile{Data: []byte(content)}
}

func newHarness(files fstest.MapFS) *harness {
	return &harness{
		files: files,
		runner: &fakeRunner{
			envs:    map[string][]string{},
			quiet:   map[string]bool{},
			failing: map[string]bool{},
			outputs: map[string]string{},
		},
		env:  map[string]string{"HOME": homeDir},
		host: map[string]string{},
	}
}

func (h *harness) run(t *testing.T, args ...string) error {
	t.Helper()
	dep := &Dependencies{
		Dir:    repoDir,
		Root:   h.files,
		Runner: h.runner,
		Out:    &h.out,
		ReadHost: func(path string) ([]byte, error) {
			if text, ok := h.host[path]; ok {
				return []byte(text), nil
			}
			return nil, fs.ErrNotExist
		},
	}
	return Command(t.Context(), args, func(name string) string { return h.env[name] }, dep)
}

func (h *harness) called(prefix string) bool {
	for _, call := range h.runner.calls {
		if strings.HasPrefix(call, prefix) {
			return true
		}
	}
	return false
}

func (h *harness) ranked() string {
	return strings.Join(h.runner.calls, "\n")
}

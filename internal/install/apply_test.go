package install

import (
	"errors"
	"slices"
	"strings"
	"testing"
)

const (
	cloneURL    = "https://github.com/mattpocock/skills.git"
	skillSource = "/tmp/skill-1/skills/productivity/grilling"
	doneText    = "\nDone. Restart Claude Code and run /pstack:setup-pstack once.\n"
)

func TestInstallPieces(t *testing.T) {
	t.Run("given a missing plugin, when installing, then it adds the marketplace and installs the plugin", func(t *testing.T) {
		s := newSetup(t)
		s.allInstalled()
		s.runner.replies["claude plugin list"] = reply{}
		if err := s.run(); err != nil {
			t.Fatal(err)
		}
		want := []string{
			"stream claude plugin marketplace add cursor/plugins",
			"stream claude plugin install pstack@cursor-plugins",
		}
		if got := streamed(s); !slices.Equal(got, want) {
			t.Fatalf("got %v", got)
		}
		if !strings.Contains(s.typed(), row("pstack", "installed")) {
			t.Fatalf("got\n%s", s.typed())
		}
	})

	t.Run("given an optional piece, when installing without the flag, then it is skipped", func(t *testing.T) {
		s := newSetup(t)
		if err := s.run(); err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(s.typed(), row("openspec", "skipped (optional)")) {
			t.Fatalf("got\n%s", s.typed())
		}
		if slices.Contains(streamed(s), "stream npm install -g @fission-ai/openspec@1.7.0") {
			t.Fatal("the optional piece was installed")
		}
	})

	t.Run("given the optional flag, when the npm package is missing, then it installs the pinned version", func(t *testing.T) {
		s := newSetup(t)
		if err := s.run("--optional"); err != nil {
			t.Fatal(err)
		}
		if !slices.Contains(streamed(s), "stream npm install -g @fission-ai/openspec@1.7.0") {
			t.Fatalf("got %v", streamed(s))
		}
	})

	t.Run("given a missing formula, when installing, then it runs brew install", func(t *testing.T) {
		s := newSetup(t)
		if err := s.run(); err != nil {
			t.Fatal(err)
		}
		if !slices.Contains(streamed(s), "stream brew install dagger/tap/container-use") {
			t.Fatalf("got %v", streamed(s))
		}
	})

	t.Run("given a missing image, when installing, then it pulls it with podman", func(t *testing.T) {
		s := newSetup(t)
		if err := s.run(); err != nil {
			t.Fatal(err)
		}
		pull := "stream podman pull docker.io/lycheeverse/lychee:0.24.2"
		if !slices.Contains(streamed(s), pull) {
			t.Fatalf("got %v", streamed(s))
		}
	})

	t.Run("given every piece in place, when installing, then it installs nothing and shows the versions", func(t *testing.T) {
		s := newSetup(t)
		s.allInstalled()
		if err := s.run(); err != nil {
			t.Fatal(err)
		}
		if len(streamed(s)) != 0 || s.files.tempDirs != 0 {
			t.Fatalf("installed something: %v", streamed(s))
		}
		if !strings.Contains(s.typed(), row("container-use", "container-use version 0.4.2")) {
			t.Fatalf("got\n%s", s.typed())
		}
	})

	t.Run("given a step that fails, when installing, then it stops and names the piece", func(t *testing.T) {
		s := newSetup(t)
		s.runner.streams["claude plugin install pstack@cursor-plugins"] = errors.New("boom")
		err := s.run()
		if err == nil || !strings.Contains(err.Error(), "pstack") || !strings.Contains(err.Error(), "boom") {
			t.Fatalf("got %v", err)
		}
		if slices.Contains(streamed(s), "stream brew install dagger/tap/container-use") {
			t.Fatal("kept going after the failure")
		}
	})

	t.Run("given a run that works, when it ends, then it tells to restart Claude Code and run setup-pstack", func(t *testing.T) {
		s := newSetup(t)
		s.allInstalled()
		if err := s.run(); err != nil {
			t.Fatal(err)
		}
		if !strings.HasSuffix(s.typed(), doneText) {
			t.Fatalf("got\n%s", s.typed())
		}
	})
}

func streamed(s *setup) []string {
	var got []string
	for _, call := range s.runner.calls {
		if strings.HasPrefix(call, "stream ") {
			got = append(got, call)
		}
	}
	return got
}

func TestInstallSkill(t *testing.T) {
	t.Run("given a missing skill, when installing, then it clones, checks out the commit and copies the path", func(t *testing.T) {
		s := newSetup(t)
		s.files.files[skillSource+"/SKILL.md"] = "skill"
		if err := s.run(); err != nil {
			t.Fatal(err)
		}
		clone := "git clone --quiet --no-checkout " + cloneURL + " /tmp/skill-1"
		checkout := "git -C /tmp/skill-1 checkout --quiet " + pinned
		if !slices.Contains(s.runner.calls, clone) || !slices.Contains(s.runner.calls, checkout) {
			t.Fatalf("got %v", s.runner.calls)
		}
		if got := s.files.files[claudeDir+"/skills/grilling/SKILL.md"]; got != "skill" {
			t.Fatalf("got %q", got)
		}
	})

	t.Run("given a missing skill, when installing, then it writes the commit marker and removes the temp folder", func(t *testing.T) {
		s := newSetup(t)
		s.files.files[skillSource+"/SKILL.md"] = "skill"
		if err := s.run(); err != nil {
			t.Fatal(err)
		}
		if got := s.files.files[claudeDir+"/skills/grilling/.rangers-commit"]; got != pinned+"\n" {
			t.Fatalf("got %q", got)
		}
		if !slices.Contains(s.files.removed, "/tmp/skill-1") {
			t.Fatalf("got %v", s.files.removed)
		}
	})

	t.Run("given a skill folder that already exists, when installing, then it is left alone", func(t *testing.T) {
		s := newSetup(t)
		s.files.files[claudeDir+"/skills/grilling/SKILL.md"] = "mine"
		if err := s.run(); err != nil {
			t.Fatal(err)
		}
		if got := s.files.files[claudeDir+"/skills/grilling/SKILL.md"]; got != "mine" {
			t.Fatalf("got %q", got)
		}
		if !strings.Contains(s.typed(), row("grilling", "present, unpinned")) {
			t.Fatalf("got\n%s", s.typed())
		}
		if s.files.tempDirs != 0 {
			t.Fatal("cloned a skill that was already there")
		}
	})

	t.Run("given a clone that fails, when installing, then the error names the skill and the temp folder is removed", func(t *testing.T) {
		s := newSetup(t)
		s.runner.replies["git clone --quiet --no-checkout "+cloneURL+" /tmp/skill-1"] = reply{err: errors.New("offline")}
		err := s.run()
		if err == nil || !strings.Contains(err.Error(), "grilling") || !strings.Contains(err.Error(), "offline") {
			t.Fatalf("got %v", err)
		}
		if !slices.Contains(s.files.removed, "/tmp/skill-1") {
			t.Fatalf("got %v", s.files.removed)
		}
	})

	t.Run("given a copy that fails, when installing, then the half copied folder is removed", func(t *testing.T) {
		s := newSetup(t)
		s.files.copyErr = errors.New("disk full")
		err := s.run()
		if err == nil || !strings.Contains(err.Error(), "disk full") {
			t.Fatalf("got %v", err)
		}
		if !slices.Contains(s.files.removed, claudeDir+"/skills/grilling") {
			t.Fatalf("got %v", s.files.removed)
		}
	})
}

func TestInstallLinks(t *testing.T) {
	t.Run("given agents, rules and the skill in the repository, when installing, then it links each one into the claude dir", func(t *testing.T) {
		s := newSetup(t)
		s.allInstalled()
		s.files.files[repoDir+"/agents/a.md"] = "a"
		s.files.files[repoDir+"/agents/b.md"] = "b"
		s.files.files[repoDir+"/agents/notes.txt"] = "x"
		s.files.files[repoDir+"/rules/r.md"] = "r"
		s.files.files[repoDir+"/skills/rangers/SKILL.md"] = "s"
		if err := s.run(); err != nil {
			t.Fatal(err)
		}
		want := map[string]string{
			claudeDir + "/agents/a.md":    repoDir + "/agents/a.md",
			claudeDir + "/agents/b.md":    repoDir + "/agents/b.md",
			claudeDir + "/rules/r.md":     repoDir + "/rules/r.md",
			claudeDir + "/skills/rangers": repoDir + "/skills/rangers",
		}
		if len(s.files.links) != len(want) {
			t.Fatalf("got %v", s.files.links)
		}
		for link, target := range want {
			if s.files.links[link] != target {
				t.Errorf("%s: got %q, want %q", link, s.files.links[link], target)
			}
		}
		if !strings.Contains(s.typed(), row("links", "4 linked into "+claudeDir)) {
			t.Fatalf("got\n%s", s.typed())
		}
	})

	t.Run("given a repository without the skill folder, when installing, then it links only what exists", func(t *testing.T) {
		s := newSetup(t)
		s.allInstalled()
		s.files.files[repoDir+"/agents/a.md"] = "a"
		if err := s.run(); err != nil {
			t.Fatal(err)
		}
		if len(s.files.links) != 1 {
			t.Fatalf("got %v", s.files.links)
		}
	})

	t.Run("given the claude dir flag, when installing, then the links go there", func(t *testing.T) {
		s := newSetup(t)
		s.allInstalled()
		s.files.files["/other/skills/grilling/.rangers-commit"] = pinned
		s.files.files[repoDir+"/agents/a.md"] = "a"
		if err := s.run("--claude-dir", "/other"); err != nil {
			t.Fatal(err)
		}
		if s.files.links["/other/agents/a.md"] != repoDir+"/agents/a.md" {
			t.Fatalf("got %v", s.files.links)
		}
	})

	t.Run("given a link that cannot be made, when installing, then it returns the error", func(t *testing.T) {
		s := newSetup(t)
		s.allInstalled()
		s.files.files[repoDir+"/agents/a.md"] = "a"
		s.files.linkErr = errors.New("is a directory")
		err := s.run()
		if err == nil || !strings.Contains(err.Error(), "is a directory") {
			t.Fatalf("got %v", err)
		}
	})
}

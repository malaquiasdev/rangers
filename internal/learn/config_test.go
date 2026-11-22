package learn

import (
	"strings"
	"testing"
)

func environment(values map[string]string) func(string) string {
	return func(name string) string { return values[name] }
}

func TestReadConfig(t *testing.T) {
	t.Run("given no flags and no variables, when reading the config, then it uses the defaults", func(t *testing.T) {
		cfg, err := ReadConfig(nil, environment(map[string]string{"HOME": "/home"}))
		if err != nil {
			t.Fatal(err)
		}
		expected := Config{
			IntervalHours:      2,
			MinLines:           10,
			BudgetUSD:          3,
			DailyBudgetUSD:     6,
			Model:              "opus",
			WindowHours:        24,
			State:              "/home/.local/state/rangers/learn",
			Repo:               "/home/Developer/me/rangers",
			Projects:           "/home/.claude/projects",
			Brain:              "/home/Documents/brain",
			NotesModel:         "sonnet",
			NotesMaxBytes:      60000,
			CallTimeoutMinutes: 20,
		}
		if cfg != expected {
			t.Fatalf("expected %+v, got %+v", expected, cfg)
		}
	})

	t.Run("given LEARN variables, when reading the config, then they replace the defaults", func(t *testing.T) {
		cfg, err := ReadConfig(nil, environment(map[string]string{
			"HOME":                       "/home",
			"LEARN_INTERVAL_HOURS":       "5",
			"LEARN_MIN_LINES":            "3",
			"LEARN_BUDGET_USD":           "1.5",
			"LEARN_DAILY_BUDGET_USD":     "8.5",
			"LEARN_MODEL":                "sonnet",
			"LEARN_WINDOW_HOURS":         "48",
			"LEARN_STATE":                "/x/state",
			"LEARN_REPO":                 "/x/repo",
			"LEARN_PROJECTS":             "/x/projects",
			"LEARN_BRAIN":                "/x/brain",
			"LEARN_NOTES_MODEL":          "haiku",
			"LEARN_NOTES_MAX_BYTES":      "1000",
			"LEARN_NOTES":                "off",
			"LEARN_CALL_TIMEOUT_MINUTES": "7.5",
		}))
		if err != nil {
			t.Fatal(err)
		}
		expected := Config{IntervalHours: 5, MinLines: 3, BudgetUSD: 1.5, DailyBudgetUSD: 8.5, Model: "sonnet", WindowHours: 48, State: "/x/state", Repo: "/x/repo", Projects: "/x/projects", Brain: "/x/brain", NotesModel: "haiku", NotesMaxBytes: 1000, SkipNotes: true, CallTimeoutMinutes: 7.5}
		if cfg != expected {
			t.Fatalf("expected %+v, got %+v", expected, cfg)
		}
	})

	t.Run("given a variable and a flag for the same value, when reading the config, then the flag wins", func(t *testing.T) {
		cfg, err := ReadConfig(
			[]string{"--budget-usd", "0.5", "--daily-budget-usd", "2", "--model", "haiku", "--min-lines", "1", "--interval-hours", "0", "--window-hours", "6", "--state", "/f/s", "--repo", "/f/r", "--projects", "/f/p", "--brain", "/f/b", "--notes-model", "opus", "--notes-max-bytes", "500", "--skip-notes", "--call-timeout-minutes", "3"},
			environment(map[string]string{"HOME": "/home", "LEARN_BUDGET_USD": "9", "LEARN_DAILY_BUDGET_USD": "20", "LEARN_MODEL": "opus", "LEARN_CALL_TIMEOUT_MINUTES": "9"}),
		)
		if err != nil {
			t.Fatal(err)
		}
		expected := Config{IntervalHours: 0, MinLines: 1, BudgetUSD: 0.5, DailyBudgetUSD: 2, Model: "haiku", WindowHours: 6, State: "/f/s", Repo: "/f/r", Projects: "/f/p", Brain: "/f/b", NotesModel: "opus", NotesMaxBytes: 500, SkipNotes: true, CallTimeoutMinutes: 3}
		if cfg != expected {
			t.Fatalf("expected %+v, got %+v", expected, cfg)
		}
	})

	t.Run("given an invalid numeric variable, when reading the config, then it returns an error naming the variable", func(t *testing.T) {
		_, err := ReadConfig(nil, environment(map[string]string{"HOME": "/home", "LEARN_BUDGET_USD": "lots"}))
		if err == nil || !strings.Contains(err.Error(), "LEARN_BUDGET_USD") {
			t.Fatalf("got %v", err)
		}
	})

	t.Run("given a zero daily budget, when reading the config, then it refuses", func(t *testing.T) {
		_, err := ReadConfig([]string{"--daily-budget-usd", "0"}, environment(map[string]string{"HOME": "/home"}))
		if err == nil {
			t.Fatal("expected an error")
		}
	})

	t.Run("given LEARN_NOTES set to something other than off, when reading the config, then the notes pass stays on", func(t *testing.T) {
		cfg, err := ReadConfig(nil, environment(map[string]string{"HOME": "/home", "LEARN_NOTES": "on"}))
		if err != nil {
			t.Fatal(err)
		}
		if cfg.SkipNotes {
			t.Fatal("the notes pass should stay on")
		}
	})

	t.Run("given a zero notes byte limit, when reading the config, then it refuses", func(t *testing.T) {
		_, err := ReadConfig([]string{"--notes-max-bytes", "0"}, environment(map[string]string{"HOME": "/home"}))
		if err == nil {
			t.Fatal("expected an error")
		}
	})

	t.Run("given a zero budget, when reading the config, then it refuses", func(t *testing.T) {
		_, err := ReadConfig([]string{"--budget-usd", "0"}, environment(map[string]string{"HOME": "/home"}))
		if err == nil {
			t.Fatal("expected an error")
		}
	})

	t.Run("given a zero call timeout, when reading the config, then it refuses", func(t *testing.T) {
		_, err := ReadConfig([]string{"--call-timeout-minutes", "0"}, environment(map[string]string{"HOME": "/home"}))
		if err == nil {
			t.Fatal("expected an error")
		}
	})

	t.Run("given a call timeout variable that is not a number, when reading the config, then it returns an error naming the variable", func(t *testing.T) {
		_, err := ReadConfig(nil, environment(map[string]string{"HOME": "/home", "LEARN_CALL_TIMEOUT_MINUTES": "soon"}))
		if err == nil || !strings.Contains(err.Error(), "LEARN_CALL_TIMEOUT_MINUTES") {
			t.Fatalf("got %v", err)
		}
	})
}

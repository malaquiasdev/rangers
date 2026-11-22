package learn

import (
	"errors"
	"flag"
	"fmt"
	"path/filepath"
	"strconv"
)

type Config struct {
	IntervalHours  float64
	MinLines       int
	BudgetUSD      float64
	DailyBudgetUSD float64
	Model          string
	WindowHours    float64
	State          string
	Repo           string
	Projects       string
	Brain          string
	NotesModel     string
	NotesMaxBytes  int
	SkipNotes      bool
}

func ReadConfig(args []string, env func(string) string) (Config, error) {
	var failure error
	register := func(name string, err error) {
		if err != nil && failure == nil {
			failure = fmt.Errorf("%s inválido: %w", name, err)
		}
	}
	text := func(name, fallback string) string {
		if value := env(name); value != "" {
			return value
		}
		return fallback
	}
	decimal := func(name string, fallback float64) float64 {
		value := env(name)
		if value == "" {
			return fallback
		}
		number, err := strconv.ParseFloat(value, 64)
		register(name, err)
		return number
	}
	integer := func(name string, fallback int) int {
		value := env(name)
		if value == "" {
			return fallback
		}
		number, err := strconv.Atoi(value)
		register(name, err)
		return number
	}

	home := env("HOME")
	var cfg Config
	fs := flag.NewFlagSet("learn", flag.ContinueOnError)
	fs.Float64Var(&cfg.IntervalHours, "interval-hours", decimal("LEARN_INTERVAL_HOURS", 2), "intervalo mínimo desde o checkpoint, em horas")
	fs.IntVar(&cfg.MinLines, "min-lines", integer("LEARN_MIN_LINES", 10), "mínimo de falas novas")
	fs.Float64Var(&cfg.BudgetUSD, "budget-usd", decimal("LEARN_BUDGET_USD", 3), "teto em dólar por rodada")
	fs.Float64Var(&cfg.DailyBudgetUSD, "daily-budget-usd", decimal("LEARN_DAILY_BUDGET_USD", 6), "teto em dólar por dia")
	fs.StringVar(&cfg.Model, "model", text("LEARN_MODEL", "opus"), "modelo")
	fs.Float64Var(&cfg.WindowHours, "window-hours", decimal("LEARN_WINDOW_HOURS", 24), "janela em horas quando não há checkpoint")
	fs.StringVar(&cfg.State, "state", text("LEARN_STATE", filepath.Join(home, ".local", "state", "rangers", "learn")), "pasta de estado")
	fs.StringVar(&cfg.Repo, "repo", text("LEARN_REPO", filepath.Join(home, "Developer", "me", "rangers")), "repositório das regras")
	fs.StringVar(&cfg.Projects, "projects", text("LEARN_PROJECTS", filepath.Join(home, ".claude", "projects")), "pasta das transcrições")
	fs.StringVar(&cfg.Brain, "brain", text("LEARN_BRAIN", filepath.Join(home, "Documents", "brain")), "pasta do brain")
	fs.StringVar(&cfg.NotesModel, "notes-model", text("LEARN_NOTES_MODEL", "sonnet"), "modelo da passada de notas")
	fs.IntVar(&cfg.NotesMaxBytes, "notes-max-bytes", integer("LEARN_NOTES_MAX_BYTES", 60000), "limite de bytes da entrada das notas")
	fs.BoolVar(&cfg.SkipNotes, "skip-notes", env("LEARN_NOTES") == "off", "pula a passada de notas")
	if err := fs.Parse(args); err != nil {
		return Config{}, err
	}
	if failure != nil {
		return Config{}, failure
	}
	if cfg.BudgetUSD <= 0 {
		return Config{}, errors.New("o teto em dólar precisa ser maior que zero")
	}
	if cfg.NotesMaxBytes <= 0 {
		return Config{}, errors.New("o limite de bytes das notas precisa ser maior que zero")
	}
	if cfg.DailyBudgetUSD <= 0 {
		return Config{}, errors.New("o teto diário em dólar precisa ser maior que zero")
	}
	return cfg, nil
}

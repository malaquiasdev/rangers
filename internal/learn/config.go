package learn

import (
	"errors"
	"flag"
	"fmt"
	"path/filepath"
	"strconv"
)

type Config struct {
	IntervalHours      float64
	MinLines           int
	BudgetUSD          float64
	DailyBudgetUSD     float64
	Model              string
	WindowHours        float64
	State              string
	Repo               string
	Projects           string
	Brain              string
	NotesModel         string
	NotesMaxBytes      int
	SkipNotes          bool
	CallTimeoutMinutes float64
}

type settings struct {
	get     func(string) string
	failure error
}

func ReadConfig(args []string, env func(string) string) (Config, error) {
	var cfg Config
	set := &settings{get: env}
	flags := flag.NewFlagSet("learn", flag.ContinueOnError)
	set.define(flags, &cfg)
	if err := flags.Parse(args); err != nil {
		return Config{}, err
	}
	if set.failure != nil {
		return Config{}, set.failure
	}
	if err := validate(&cfg); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

func (s *settings) define(flags *flag.FlagSet, cfg *Config) {
	home := s.get("HOME")
	flags.Float64Var(&cfg.IntervalHours, "interval-hours", s.decimal("LEARN_INTERVAL_HOURS", 2),
		"intervalo mínimo desde o checkpoint, em horas")
	flags.IntVar(&cfg.MinLines, "min-lines", s.integer("LEARN_MIN_LINES", 10), "mínimo de falas novas")
	flags.Float64Var(&cfg.BudgetUSD, "budget-usd", s.decimal("LEARN_BUDGET_USD", 3),
		"teto em dólar por rodada")
	flags.Float64Var(&cfg.DailyBudgetUSD, "daily-budget-usd", s.decimal("LEARN_DAILY_BUDGET_USD", 6),
		"teto em dólar por dia")
	flags.StringVar(&cfg.Model, "model", s.text("LEARN_MODEL", "opus"), "modelo")
	flags.Float64Var(&cfg.WindowHours, "window-hours", s.decimal("LEARN_WINDOW_HOURS", 24),
		"janela em horas quando não há checkpoint")
	flags.StringVar(&cfg.State, "state", s.text("LEARN_STATE",
		filepath.Join(home, ".local", "state", "rangers", "learn")), "pasta de estado")
	flags.StringVar(&cfg.Repo, "repo", s.text("LEARN_REPO",
		filepath.Join(home, "Developer", "me", "rangers")), "repositório das regras")
	flags.StringVar(&cfg.Projects, "projects", s.text("LEARN_PROJECTS",
		filepath.Join(home, ".claude", "projects")), "pasta das transcrições")
	s.defineNotes(flags, cfg, home)
}

func (s *settings) defineNotes(flags *flag.FlagSet, cfg *Config, home string) {
	flags.StringVar(&cfg.Brain, "brain", s.text("LEARN_BRAIN",
		filepath.Join(home, "Documents", "brain")), "pasta do brain")
	flags.StringVar(&cfg.NotesModel, "notes-model", s.text("LEARN_NOTES_MODEL", "sonnet"),
		"modelo da passada de notas")
	flags.IntVar(&cfg.NotesMaxBytes, "notes-max-bytes", s.integer("LEARN_NOTES_MAX_BYTES", 60000),
		"limite de bytes da entrada das notas")
	flags.BoolVar(&cfg.SkipNotes, "skip-notes", s.get("LEARN_NOTES") == "off",
		"pula a passada de notas")
	flags.Float64Var(&cfg.CallTimeoutMinutes, "call-timeout-minutes",
		s.decimal("LEARN_CALL_TIMEOUT_MINUTES", 20), "limite em minutos de cada chamada ao claude")
}

func (s *settings) text(name, fallback string) string {
	if value := s.get(name); value != "" {
		return value
	}
	return fallback
}

func (s *settings) decimal(name string, fallback float64) float64 {
	value := s.get(name)
	if value == "" {
		return fallback
	}
	number, err := strconv.ParseFloat(value, 64)
	s.register(name, err)
	return number
}

func (s *settings) integer(name string, fallback int) int {
	value := s.get(name)
	if value == "" {
		return fallback
	}
	number, err := strconv.Atoi(value)
	s.register(name, err)
	return number
}

func (s *settings) register(name string, err error) {
	if err != nil && s.failure == nil {
		s.failure = fmt.Errorf("%s inválido: %w", name, err)
	}
}

func validate(cfg *Config) error {
	switch {
	case cfg.BudgetUSD <= 0:
		return errors.New("o teto em dólar precisa ser maior que zero")
	case cfg.NotesMaxBytes <= 0:
		return errors.New("o limite de bytes das notas precisa ser maior que zero")
	case cfg.DailyBudgetUSD <= 0:
		return errors.New("o teto diário em dólar precisa ser maior que zero")
	case cfg.CallTimeoutMinutes <= 0:
		return errors.New("o limite de minutos de cada chamada precisa ser maior que zero")
	}
	return nil
}

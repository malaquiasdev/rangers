package learn

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/malaquiasdev/rangers/internal/transcript"
)

const (
	checkpointName = "checkpoint"
	metricsName    = "metrics.tsv"
	allowedTools   = "Read,Edit,Write,Bash(git add:*),Bash(git commit:*),Bash(git status:*),Bash(git diff:*)"
	reminderTitle  = "Revisar as regras propostas"
)

type Dependencies struct {
	Claude   Claude
	Git      Git
	Reminder Reminder
	Now      func() time.Time
	Out      io.Writer
}

type response struct {
	DurationMs int64   `json:"duration_ms"`
	Turns      int     `json:"num_turns"`
	CostUSD    float64 `json:"total_cost_usd"`
	Result     string  `json:"result"`
	Usage      struct {
		Input         int `json:"input_tokens"`
		CacheCreation int `json:"cache_creation_input_tokens"`
		CacheRead     int `json:"cache_read_input_tokens"`
		Output        int `json:"output_tokens"`
	} `json:"usage"`
}

func Run(ctx context.Context, cfg *Config, dep Dependencies) error {
	now := dep.Now()
	lines, reason, err := collect(cfg, now)
	if err != nil {
		return err
	}
	if reason != "" {
		_, err = fmt.Fprintln(dep.Out, reason)
		return err
	}
	spent, err := spentToday(cfg.State, now)
	if err != nil {
		return err
	}
	if reason = blocksDaily(cfg.DailyBudgetUSD, spent); reason != "" {
		_, err = fmt.Fprintln(dep.Out, reason)
		return err
	}
	return runModel(ctx, cfg, dep, now, lines, roundBudget(cfg.BudgetUSD, cfg.DailyBudgetUSD, spent))
}

func collect(cfg *Config, now time.Time) ([]transcript.Line, string, error) {
	if err := os.MkdirAll(cfg.State, 0o750); err != nil {
		return nil, "", err
	}
	cutoff := now.Add(-time.Duration(cfg.WindowHours * float64(time.Hour)))
	info, err := os.Stat(filepath.Join(cfg.State, checkpointName))
	switch {
	case err == nil:
		if reason := blocksInterval(cfg.IntervalHours, info.ModTime(), now); reason != "" {
			return nil, reason, nil
		}
		cutoff = info.ModTime()
	case !errors.Is(err, fs.ErrNotExist):
		return nil, "", err
	}
	lines, err := transcript.Extract(cfg.Projects, cutoff)
	if err != nil {
		return nil, "", err
	}
	return lines, blocksLines(cfg.MinLines, len(lines)), nil
}

func runModel(ctx context.Context, cfg *Config, dep Dependencies, now time.Time, lines []transcript.Line, budget float64) error {
	agent, err := findAgent(filepath.Join(cfg.Repo, "agents"))
	if err != nil {
		return err
	}
	stamp := now.Format("2006-01-02-1504")
	input := filepath.Join(cfg.State, "input-"+stamp+".txt")
	err = os.WriteFile(input, []byte(join(lines)), 0o600)
	if err != nil {
		return err
	}
	branch := "learn/" + stamp
	worktree := cfg.Repo + "-worktrees/learn-" + stamp
	err = dep.Git.NewWorktree(cfg.Repo, branch, worktree)
	if err != nil {
		return err
	}
	prompt := fmt.Sprintf("The file %s has, one per line, the lines the human typed since the last checkpoint (session, time, text). Propose the rules in the files of this folder and make the local commit on branch %s.", input, branch)
	r, err := invoke(ctx, cfg, dep.Claude, stamp, worktree, []string{
		"-p", prompt,
		"--agent", agent,
		"--permission-mode", "acceptEdits",
		"--allowedTools", allowedTools,
		"--model", cfg.Model,
		"--max-budget-usd", strconv.FormatFloat(budget, 'f', -1, 64),
		"--output-format", "json",
	})
	if err != nil {
		return err
	}
	return finish(cfg, dep, now, stamp, branch, worktree, len(lines), &r)
}

func invoke(ctx context.Context, cfg *Config, claude Claude, stamp, worktree string, args []string) (response, error) {
	var r response
	raw, err := claude.Run(ctx, worktree, args)
	if err != nil {
		return r, err
	}
	err = os.WriteFile(filepath.Join(cfg.State, "output-"+stamp+".json"), raw, 0o600)
	if err != nil {
		return r, err
	}
	err = json.Unmarshal(raw, &r)
	if err != nil {
		return r, fmt.Errorf("resposta do claude não é JSON: %w", err)
	}
	return r, nil
}

func finish(cfg *Config, dep Dependencies, now time.Time, stamp, branch, worktree string, lines int, r *response) error {
	_, err := fmt.Fprintln(dep.Out, r.Result)
	if err != nil {
		return err
	}
	err = record(cfg.State, stamp, lines, r)
	if err != nil {
		return err
	}
	err = touchCheckpoint(filepath.Join(cfg.State, checkpointName), now)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(dep.Out, "ramo %s em %s\n", branch, worktree)
	if err != nil {
		return err
	}
	return dep.Reminder.Create(reminderTitle, fmt.Sprintf("Ramo %s em %s", branch, worktree))
}

func join(lines []transcript.Line) string {
	var joined strings.Builder
	for _, line := range lines {
		joined.WriteString(line.String())
		joined.WriteByte('\n')
	}
	return joined.String()
}

func record(state, stamp string, lines int, r *response) error {
	cost := math.Floor(r.CostUSD*1000) / 1000
	fields := []string{
		stamp,
		strconv.Itoa(lines),
		strconv.FormatInt(r.DurationMs/1000, 10),
		strconv.Itoa(r.Turns),
		strconv.FormatFloat(cost, 'f', -1, 64),
		strconv.Itoa(r.Usage.Input + r.Usage.CacheCreation + r.Usage.CacheRead),
		strconv.Itoa(r.Usage.Output),
	}
	file, err := os.OpenFile(filepath.Join(state, metricsName), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o600)
	if err != nil {
		return err
	}
	if _, err := file.WriteString(strings.Join(fields, "\t") + "\n"); err != nil {
		return errors.Join(err, file.Close())
	}
	return file.Close()
}

func touchCheckpoint(path string, now time.Time) error {
	if err := os.WriteFile(path, nil, 0o600); err != nil {
		return err
	}
	return os.Chtimes(path, now, now)
}

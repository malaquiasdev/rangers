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
	pendingName    = "pending-reminders.txt"
	allowedTools   = "Read,Edit,Write,Bash(git add:*),Bash(git commit:*),Bash(git status:*),Bash(git diff:*)"
	notesTools     = "Read,Write,Edit,Glob,Grep,Bash(~/.claude/bin/audit:*)"
	reminderTitle  = "Revisar as regras propostas"
)

type Dependencies struct {
	Claude   Claude
	Git      Git
	Reminder Reminder
	Now      func() time.Time
	Out      io.Writer
	Err      io.Writer
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

func Run(ctx context.Context, cfg *Config, dep *Dependencies) error {
	now := dep.Now()
	err := runPass(ctx, cfg, dep, now, rulesPass)
	if cfg.SkipNotes {
		return err
	}
	return errors.Join(err, runPass(ctx, cfg, dep, now, notesPass))
}

func runPass(ctx context.Context, cfg *Config, dep *Dependencies, now time.Time, pass string) error {
	lines, reason, err := collect(cfg, now, pass)
	if err != nil {
		return err
	}
	if reason != "" {
		return say(dep.Out, pass, reason)
	}
	spent, err := spentToday(cfg.State, now)
	if err != nil {
		return err
	}
	if reason = blocksDaily(cfg.DailyBudgetUSD, spent); reason != "" {
		return say(dep.Out, pass, reason)
	}
	budget := roundBudget(cfg.BudgetUSD, cfg.DailyBudgetUSD, spent)
	if pass == notesPass {
		return runNotes(ctx, cfg, dep, now, lines, budget)
	}
	return runModel(ctx, cfg, dep, now, lines, budget)
}

func say(out io.Writer, pass, text string) error {
	_, err := fmt.Fprintln(out, label(pass)+text)
	return err
}

func label(pass string) string {
	if pass == notesPass {
		return "[notas] "
	}
	return ""
}

func checkpointOf(pass string) string {
	if pass == notesPass {
		return checkpointName + "-notes"
	}
	return checkpointName
}

func humans(lines []transcript.Line) int {
	count := 0
	for _, line := range lines {
		if !line.Assistant {
			count++
		}
	}
	return count
}

func collect(cfg *Config, now time.Time, pass string) ([]transcript.Line, string, error) {
	if err := os.MkdirAll(cfg.State, 0o750); err != nil {
		return nil, "", err
	}
	cutoff := now.Add(-time.Duration(cfg.WindowHours * float64(time.Hour)))
	info, err := os.Stat(filepath.Join(cfg.State, checkpointOf(pass)))
	switch {
	case err == nil:
		if reason := blocksInterval(cfg.IntervalHours, info.ModTime(), now); reason != "" {
			return nil, reason, nil
		}
		cutoff = info.ModTime()
	case !errors.Is(err, fs.ErrNotExist):
		return nil, "", err
	}
	extract := transcript.Extract
	if pass == notesPass {
		extract = transcript.ExtractWithAssistant
	}
	lines, err := extract(cfg.Projects, cutoff)
	if err != nil {
		return nil, "", err
	}
	return lines, blocksLines(cfg.MinLines, humans(lines)), nil
}

func runNotes(ctx context.Context, cfg *Config, dep *Dependencies, now time.Time, lines []transcript.Line, budget float64) error {
	agent, err := findAgent(filepath.Join(cfg.Repo, "agents"))
	if err != nil {
		return err
	}
	stamp := now.Format("2006-01-02-1504")
	input := filepath.Join(cfg.State, "notes-input-"+stamp+".txt")
	err = os.WriteFile(input, []byte(transcript.Format(lines, cfg.NotesMaxBytes)), 0o600)
	if err != nil {
		return err
	}
	prompt := fmt.Sprintf("This is the notes pass. The file %s has, one per line, what the human and the assistant said since the last notes checkpoint (project, session, time, text). Write the notes in the brain folder you were started in, as your agent file says.", input)
	r, err := invoke(ctx, cfg, dep.Claude, "notes-output-"+stamp, cfg.Brain, []string{
		"-p", prompt,
		"--agent", agent,
		"--model", cfg.NotesModel,
		"--permission-mode", "acceptEdits",
		"--allowedTools", notesTools,
		"--max-budget-usd", strconv.FormatFloat(budget, 'f', -1, 64),
		"--output-format", "json",
	})
	if err != nil {
		return err
	}
	return settle(cfg, dep, now, stamp, notesPass, humans(lines), &r)
}

func runModel(ctx context.Context, cfg *Config, dep *Dependencies, now time.Time, lines []transcript.Line, budget float64) error {
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
	r, err := invoke(ctx, cfg, dep.Claude, "output-"+stamp, worktree, []string{
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

func invoke(ctx context.Context, cfg *Config, claude Claude, outputName, dir string, args []string) (response, error) {
	var r response
	raw, err := claude.Run(ctx, dir, args)
	if err != nil {
		return r, err
	}
	err = os.WriteFile(filepath.Join(cfg.State, outputName+".json"), raw, 0o600)
	if err != nil {
		return r, err
	}
	err = json.Unmarshal(raw, &r)
	if err != nil {
		return r, fmt.Errorf("resposta do claude não é JSON: %w", err)
	}
	return r, nil
}

func settle(cfg *Config, dep *Dependencies, now time.Time, stamp, pass string, lines int, r *response) error {
	err := say(dep.Out, pass, r.Result)
	if err != nil {
		return err
	}
	err = record(cfg.State, stamp, pass, lines, r)
	if err != nil {
		return err
	}
	return touchCheckpoint(filepath.Join(cfg.State, checkpointOf(pass)), now)
}

func finish(cfg *Config, dep *Dependencies, now time.Time, stamp, branch, worktree string, lines int, r *response) error {
	err := settle(cfg, dep, now, stamp, rulesPass, lines, r)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(dep.Out, "ramo %s em %s\n", branch, worktree)
	if err != nil {
		return err
	}
	body := fmt.Sprintf("Ramo %s em %s", branch, worktree)
	err = dep.Reminder.Create(reminderTitle, body)
	if err == nil {
		return nil
	}
	return keepPending(cfg.State, dep.Err, err, body)
}

func keepPending(state string, warnings io.Writer, cause error, body string) error {
	path := filepath.Join(state, pendingName)
	_, err := fmt.Fprintf(warnings, "aviso: não consegui criar o lembrete (%v); anotei em %s\n", cause, path)
	if err != nil {
		return err
	}
	file, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o600)
	if err != nil {
		return err
	}
	if _, err := file.WriteString(reminderTitle + "\t" + body + "\n"); err != nil {
		return errors.Join(err, file.Close())
	}
	return file.Close()
}

func join(lines []transcript.Line) string {
	var joined strings.Builder
	for _, line := range lines {
		joined.WriteString(line.String())
		joined.WriteByte('\n')
	}
	return joined.String()
}

func record(state, stamp, pass string, lines int, r *response) error {
	cost := math.Floor(r.CostUSD*1000) / 1000
	fields := []string{
		stamp,
		strconv.Itoa(lines),
		strconv.FormatInt(r.DurationMs/1000, 10),
		strconv.Itoa(r.Turns),
		strconv.FormatFloat(cost, 'f', -1, 64),
		strconv.Itoa(r.Usage.Input + r.Usage.CacheCreation + r.Usage.CacheRead),
		strconv.Itoa(r.Usage.Output),
		pass,
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

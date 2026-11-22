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
	flagModel      = "--model"
	flagBudget     = "--max-budget-usd"
	rulesPass      = "rules"
	notesPass      = "notes"
	allowedTools   = "Read,Edit,Write,Bash(git add:*),Bash(git commit:*)," +
		"Bash(git status:*),Bash(git diff:*)"
	reminderTitle = "Revisar as regras propostas"
	accessTitle   = "Dar acesso ao brain ao rangers"
	accessSetting = "em Ajustes > Privacidade e Segurança"
	rulesPrompt   = "The file %s has, one per line, the lines the human typed since the last " +
		"checkpoint (session, time, text). Propose the rules in the files of this folder " +
		"and make the local commit on branch %s. Do not write quotes, session ids or Origin " +
		"in the rules files or in the commit messages. In your final message, after the list " +
		"of rules, add a section headed `Origin:` with one bullet per rule: the start of the " +
		"rule, the session id and the human's quote."
	notesPrompt = "This is the notes pass. The file %s has, one per line, what the human and " +
		"the assistant said since the last notes checkpoint (project, session, time, text). " +
		"Write the notes in the brain folder you were started in, as your agent file says."
)

type Dependencies struct {
	Claude   Claude
	Git      Git
	Reminder Reminder
	Brain    Brain
	Binary   string
	Home     string
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

type round struct {
	cfg    *Config
	dep    *Dependencies
	now    time.Time
	pass   string
	stamp  string
	lines  []transcript.Line
	budget float64
}

func newRound(cfg *Config, dep *Dependencies, now time.Time, pass string) *round {
	return &round{cfg: cfg, dep: dep, now: now, pass: pass, stamp: now.Format("2006-01-02-1504")}
}

func Run(ctx context.Context, cfg *Config, dep *Dependencies) error {
	now := dep.Now()
	err := newRound(cfg, dep, now, rulesPass).run(ctx)
	if cfg.SkipNotes {
		return err
	}
	return errors.Join(err, newRound(cfg, dep, now, notesPass).run(ctx))
}

func (r *round) run(ctx context.Context) error {
	reason, err := r.blocked()
	if err != nil {
		return err
	}
	if reason != "" {
		return r.say(reason)
	}
	if r.pass == notesPass {
		return r.notes(ctx)
	}
	return r.rules(ctx)
}

func (r *round) blocked() (string, error) {
	reason, err := r.collect()
	if err != nil || reason != "" {
		return reason, err
	}
	spent, err := spentToday(r.cfg.State, r.now)
	if err != nil {
		return "", err
	}
	r.budget = roundBudget(r.cfg.BudgetUSD, r.cfg.DailyBudgetUSD, spent)
	return blocksDaily(r.cfg.DailyBudgetUSD, spent), nil
}

func (r *round) collect() (string, error) {
	if err := os.MkdirAll(r.cfg.State, 0o750); err != nil {
		return "", err
	}
	cutoff, reason, err := r.cutoff()
	if err != nil || reason != "" {
		return reason, err
	}
	extract := transcript.Extract
	if r.pass == notesPass {
		extract = transcript.ExtractWithAssistant
	}
	r.lines, err = extract(r.cfg.Projects, cutoff)
	if err != nil {
		return "", err
	}
	return blocksLines(r.cfg.MinLines, humans(r.lines)), nil
}

func (r *round) cutoff() (time.Time, string, error) {
	window := time.Duration(r.cfg.WindowHours * float64(time.Hour))
	info, err := os.Stat(filepath.Join(r.cfg.State, r.checkpoint()))
	if errors.Is(err, fs.ErrNotExist) {
		return r.now.Add(-window), "", nil
	}
	if err != nil {
		return time.Time{}, "", err
	}
	return info.ModTime(), blocksInterval(r.cfg.IntervalHours, info.ModTime(), r.now), nil
}

func (r *round) rules(ctx context.Context) error {
	agent, err := findAgent(filepath.Join(r.cfg.Repo, "agents"))
	if err != nil {
		return err
	}
	input, err := r.writeInput("input-", join(r.lines))
	if err != nil {
		return err
	}
	branch, worktree, err := r.newWorktree(ctx)
	if err != nil {
		return err
	}
	prompt := fmt.Sprintf(rulesPrompt, input, branch)
	args := r.arguments(prompt, agent, r.cfg.Model, allowedTools)
	resp, err := r.invoke(ctx, "output-", worktree, args)
	if err != nil {
		return err
	}
	return r.finishRules(ctx, &resp, branch, worktree)
}

func (r *round) notes(ctx context.Context) error {
	agent, err := findAgent(filepath.Join(r.cfg.Repo, "agents"))
	if err != nil {
		return err
	}
	if err = r.dep.Brain.Probe(ctx, r.cfg.Brain); err != nil {
		return r.skipNotes(err)
	}
	input, err := r.writeInput("notes-input-", transcript.Format(r.lines, r.cfg.NotesMaxBytes))
	if err != nil {
		return err
	}
	prompt := fmt.Sprintf(notesPrompt, input)
	args := r.arguments(prompt, agent, r.cfg.NotesModel, notesAllowed(r.dep.Home))
	resp, err := r.invoke(ctx, "notes-output-", r.cfg.Brain, args)
	if err != nil {
		return err
	}
	return r.settle(&resp)
}

func (r *round) skipNotes(cause error) error {
	notice := "sem acesso ao brain (%v): dê Acesso Total ao Disco a %s %s; pulo as notas"
	err := r.say(fmt.Sprintf(notice, cause, r.dep.Binary, accessSetting))
	if err != nil {
		return err
	}
	body := fmt.Sprintf("Acesso Total ao Disco para %s %s", r.dep.Binary, accessSetting)
	return appendLine(filepath.Join(r.cfg.State, pendingName), accessTitle+"\t"+body)
}

func (r *round) newWorktree(ctx context.Context) (string, string, error) {
	branch := "learn/" + r.stamp
	worktree := r.cfg.Repo + "-worktrees/learn-" + r.stamp
	return branch, worktree, r.dep.Git.NewWorktree(ctx, r.cfg.Repo, branch, worktree)
}

func (r *round) writeInput(prefix, content string) (string, error) {
	path := filepath.Join(r.cfg.State, prefix+r.stamp+".txt")
	return path, os.WriteFile(path, []byte(content), 0o600)
}

func (r *round) arguments(prompt, agent, model, tools string) []string {
	return []string{
		"-p", prompt,
		"--agent", agent,
		flagModel, model,
		"--permission-mode", "acceptEdits",
		"--allowedTools", tools,
		flagBudget, strconv.FormatFloat(r.budget, 'f', -1, 64),
		"--output-format", "json",
	}
}

func (r *round) invoke(ctx context.Context, prefix, dir string, args []string) (response, error) {
	var resp response
	raw, err := r.call(ctx, dir, args)
	if err != nil {
		return resp, err
	}
	err = os.WriteFile(filepath.Join(r.cfg.State, prefix+r.stamp+".json"), raw, 0o600)
	if err != nil {
		return resp, err
	}
	err = json.Unmarshal(raw, &resp)
	if err != nil {
		return resp, fmt.Errorf("resposta do claude não é JSON: %w", err)
	}
	return resp, nil
}

func (r *round) call(ctx context.Context, dir string, args []string) ([]byte, error) {
	limit := time.Duration(r.cfg.CallTimeoutMinutes * float64(time.Minute))
	limited, cancel := context.WithTimeout(ctx, limit)
	defer cancel()
	raw, err := r.dep.Claude.Run(limited, dir, args)
	if err != nil && errors.Is(limited.Err(), context.DeadlineExceeded) && ctx.Err() == nil {
		return nil, fmt.Errorf("o claude passou de %s sem terminar; processo encerrado: %w", limit, err)
	}
	return raw, err
}

func (r *round) settle(resp *response) error {
	err := r.say(resp.Result)
	if err != nil {
		return err
	}
	err = r.record(resp)
	if err != nil {
		return err
	}
	return touchCheckpoint(filepath.Join(r.cfg.State, r.checkpoint()), r.now)
}

func (r *round) finishRules(ctx context.Context, resp *response, branch, worktree string) error {
	err := r.settle(resp)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(r.dep.Out, "ramo %s em %s\n", branch, worktree)
	if err != nil {
		return err
	}
	body := fmt.Sprintf("Ramo %s em %s", branch, worktree)
	err = r.dep.Reminder.Create(ctx, reminderTitle, body)
	if err != nil {
		err = r.keepPending(err, body)
	}
	if err != nil {
		return err
	}
	return r.keepOrigin(ctx, resp.Result)
}

func (r *round) keepPending(cause error, body string) error {
	path := filepath.Join(r.cfg.State, pendingName)
	_, err := fmt.Fprintf(r.dep.Err,
		"aviso: não consegui criar o lembrete (%v); anotei em %s\n", cause, path)
	if err != nil {
		return err
	}
	return appendLine(path, reminderTitle+"\t"+body)
}

func (r *round) record(resp *response) error {
	cost := math.Floor(resp.CostUSD*1000) / 1000
	fields := []string{
		r.stamp,
		strconv.Itoa(humans(r.lines)),
		strconv.FormatInt(resp.DurationMs/1000, 10),
		strconv.Itoa(resp.Turns),
		strconv.FormatFloat(cost, 'f', -1, 64),
		strconv.Itoa(resp.Usage.Input + resp.Usage.CacheCreation + resp.Usage.CacheRead),
		strconv.Itoa(resp.Usage.Output),
		r.pass,
	}
	return appendLine(filepath.Join(r.cfg.State, metricsName), strings.Join(fields, "\t"))
}

func (r *round) say(text string) error {
	prefix := ""
	if r.pass == notesPass {
		prefix = "[notas] "
	}
	_, err := fmt.Fprintln(r.dep.Out, prefix+text)
	return err
}

func (r *round) checkpoint() string {
	if r.pass == notesPass {
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

func join(lines []transcript.Line) string {
	var joined strings.Builder
	for _, line := range lines {
		joined.WriteString(line.String())
		joined.WriteByte('\n')
	}
	return joined.String()
}

func appendLine(path, line string) error {
	return appendText(path, line+"\n")
}

func appendText(path, text string) error {
	file, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o600)
	if err != nil {
		return err
	}
	if _, err = file.WriteString(text); err != nil {
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

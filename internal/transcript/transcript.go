package transcript

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

const (
	textLimit      = 600
	assistantLimit = 400
)

var ignoredPrefixes = []string{"<", "Base directory", "Another Claude", "[Request interrupted", "Caveat:", "[Cross-session", "[Subagent"}

type Line struct {
	Project   string
	Session   string
	Time      string
	Text      string
	Assistant bool
}

func (l Line) String() string {
	return l.Session + "\t" + l.Time + "\t" + l.Text
}

func (l Line) Notes() string {
	text := l.Text
	if l.Assistant {
		text = "[assistant] " + text
	}
	return l.Project + "\t" + l.Session + "\t" + l.Time + "\t" + text
}

func Format(lines []Line, maxBytes int) string {
	rows := make([]string, len(lines))
	total := 0
	for i, line := range lines {
		rows[i] = line.Notes() + "\n"
		total += len(rows[i])
	}
	start := 0
	for start < len(rows) && total > maxBytes {
		total -= len(rows[start])
		start++
	}
	return strings.Join(rows[start:], "")
}

type entry struct {
	Kind      string `json:"type"`
	Timestamp string `json:"timestamp"`
	Message   struct {
		Role    string          `json:"role"`
		Content json.RawMessage `json:"content"`
	} `json:"message"`
}

func Extract(projects string, cutoff time.Time) ([]Line, error) {
	return collect(projects, cutoff, false)
}

func ExtractWithAssistant(projects string, cutoff time.Time) ([]Line, error) {
	lines, err := collect(projects, cutoff, true)
	sort.SliceStable(lines, func(i, j int) bool { return lines[i].Time < lines[j].Time })
	return lines, err
}

func collect(projects string, cutoff time.Time, withAssistant bool) ([]Line, error) {
	files, err := changedFiles(projects, cutoff)
	if err != nil {
		return nil, err
	}
	var lines []Line
	for _, file := range files {
		session := strings.TrimSuffix(filepath.Base(file), ".jsonl")
		project := filepath.Base(filepath.Dir(file))
		found, err := readLines(file, project, session, cutoff, withAssistant)
		if err != nil {
			return nil, err
		}
		lines = append(lines, found...)
	}
	return lines, nil
}

func changedFiles(projects string, cutoff time.Time) ([]string, error) {
	shallow, err := filepath.Glob(filepath.Join(projects, "*.jsonl"))
	if err != nil {
		return nil, err
	}
	deep, err := filepath.Glob(filepath.Join(projects, "*", "*.jsonl"))
	if err != nil {
		return nil, err
	}
	var changed []string
	for _, path := range append(shallow, deep...) {
		info, err := os.Stat(path)
		if errors.Is(err, fs.ErrNotExist) {
			continue
		}
		if err != nil {
			return nil, err
		}
		if info.ModTime().After(cutoff) {
			changed = append(changed, path)
		}
	}
	sort.Strings(changed)
	return changed, nil
}

func readLines(file, project, session string, cutoff time.Time, withAssistant bool) ([]Line, error) {
	raw, err := os.ReadFile(file)
	if err != nil {
		return nil, err
	}
	var lines []Line
	decoder := json.NewDecoder(bytes.NewReader(raw))
	for {
		var r entry
		if decoder.Decode(&r) != nil {
			return lines, nil
		}
		human := r.Kind == "user" && r.Message.Role == "user"
		assistant := withAssistant && r.Kind == "assistant" && r.Message.Role == "assistant"
		if !human && !assistant {
			continue
		}
		at, err := time.Parse(time.RFC3339Nano, r.Timestamp)
		if err != nil || !at.After(cutoff) {
			continue
		}
		text := content(r.Message.Content)
		if text == "" || (human && ignored(text)) {
			continue
		}
		limit := textLimit
		if assistant {
			limit = assistantLimit
		}
		text = strings.ReplaceAll(text, "\n", " ")
		if body := []rune(text); len(body) > limit {
			text = string(body[:limit])
		}
		lines = append(lines, Line{Project: project, Session: session, Time: r.Timestamp, Text: text, Assistant: assistant})
	}
}

func content(raw json.RawMessage) string {
	var text string
	if json.Unmarshal(raw, &text) == nil {
		return text
	}
	var parts []struct {
		Kind string `json:"type"`
		Text string `json:"text"`
	}
	if json.Unmarshal(raw, &parts) != nil {
		return ""
	}
	var texts []string
	for _, part := range parts {
		if part.Kind == "text" {
			texts = append(texts, part.Text)
		}
	}
	return strings.Join(texts, " ")
}

func ignored(text string) bool {
	for _, prefix := range ignoredPrefixes {
		if strings.HasPrefix(text, prefix) {
			return true
		}
	}
	return false
}

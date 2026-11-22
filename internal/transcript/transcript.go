package transcript

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

const (
	textLimit      = 600
	assistantLimit = 400
	humanSpeaker   = "user"
	modelSpeaker   = "assistant"
	textPart       = "text"
)

var ignoredPrefixes = []string{
	"<", "Base directory", "Another Claude", "[Request interrupted",
	"Caveat:", "[Cross-session", "[Subagent",
}

type Line struct {
	Project   string
	Session   string
	Time      string
	Text      string
	Assistant bool
}

type entry struct {
	Kind      string `json:"type"`
	Timestamp string `json:"timestamp"`
	Message   struct {
		Role    string          `json:"role"`
		Content json.RawMessage `json:"content"`
	} `json:"message"`
}

type source struct {
	project       string
	session       string
	cutoff        time.Time
	withAssistant bool
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

func (s source) read(raw []byte) []Line {
	var lines []Line
	decoder := json.NewDecoder(bytes.NewReader(raw))
	for {
		var e entry
		if decoder.Decode(&e) != nil {
			return lines
		}
		if line, ok := s.line(&e); ok {
			lines = append(lines, line)
		}
	}
}

func (s source) line(e *entry) (Line, bool) {
	who := speaker(e, s.withAssistant)
	if who == "" || !s.after(e.Timestamp) {
		return Line{}, false
	}
	text := content(e.Message.Content)
	if text == "" || (who == humanSpeaker && ignored(text)) {
		return Line{}, false
	}
	limit := textLimit
	if who == modelSpeaker {
		limit = assistantLimit
	}
	return Line{
		Project:   s.project,
		Session:   s.session,
		Time:      e.Timestamp,
		Text:      clip(text, limit),
		Assistant: who == modelSpeaker,
	}, true
}

func (s source) after(timestamp string) bool {
	at, err := time.Parse(time.RFC3339Nano, timestamp)
	return err == nil && at.After(s.cutoff)
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

func Extract(projects string, cutoff time.Time) ([]Line, error) {
	return collect(projects, cutoff, false)
}

func ExtractWithAssistant(projects string, cutoff time.Time) ([]Line, error) {
	lines, err := collect(projects, cutoff, true)
	sort.SliceStable(lines, func(i, j int) bool { return lines[i].Time < lines[j].Time })
	return lines, err
}

func collect(projects string, cutoff time.Time, withAssistant bool) ([]Line, error) {
	root, err := os.OpenRoot(projects)
	if errors.Is(err, fs.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	lines, err := readAll(root, cutoff, withAssistant)
	return lines, errors.Join(err, root.Close())
}

func readAll(root *os.Root, cutoff time.Time, withAssistant bool) ([]Line, error) {
	fsys := root.FS()
	files, err := changedFiles(fsys, cutoff)
	if err != nil {
		return nil, err
	}
	var lines []Line
	for _, file := range files {
		raw, err := fs.ReadFile(fsys, file)
		if err != nil {
			return nil, err
		}
		src := source{
			project:       projectOf(root, file),
			session:       strings.TrimSuffix(path.Base(file), ".jsonl"),
			cutoff:        cutoff,
			withAssistant: withAssistant,
		}
		lines = append(lines, src.read(raw)...)
	}
	return lines, nil
}

func projectOf(root *os.Root, file string) string {
	dir := path.Dir(file)
	if dir == "." {
		return filepath.Base(root.Name())
	}
	return path.Base(dir)
}

func changedFiles(fsys fs.FS, cutoff time.Time) ([]string, error) {
	var changed []string
	for _, pattern := range []string{"*.jsonl", "*/*.jsonl"} {
		matches, err := fs.Glob(fsys, pattern)
		if err != nil {
			return nil, err
		}
		for _, match := range matches {
			newer, err := modifiedAfter(fsys, match, cutoff)
			if err != nil {
				return nil, err
			}
			if newer {
				changed = append(changed, match)
			}
		}
	}
	sort.Strings(changed)
	return changed, nil
}

func modifiedAfter(fsys fs.FS, file string, cutoff time.Time) (bool, error) {
	info, err := fs.Stat(fsys, file)
	if errors.Is(err, fs.ErrNotExist) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return info.ModTime().After(cutoff), nil
}

func speaker(e *entry, withAssistant bool) string {
	switch {
	case e.Kind == humanSpeaker && e.Message.Role == humanSpeaker:
		return humanSpeaker
	case withAssistant && e.Kind == modelSpeaker && e.Message.Role == modelSpeaker:
		return modelSpeaker
	}
	return ""
}

func clip(text string, limit int) string {
	flat := strings.ReplaceAll(text, "\n", " ")
	if body := []rune(flat); len(body) > limit {
		return string(body[:limit])
	}
	return flat
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
		if part.Kind == textPart {
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

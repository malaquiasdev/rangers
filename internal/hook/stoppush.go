package hook

import (
	"bufio"
	"encoding/json"
	"io"
	"os"
	"strings"
	"unicode"
)

const (
	blockLead  = "This turn ends waiting for the human but the notification is incomplete. "
	blockTrail = " Then end the turn."
	pushStep   = "Call PushNotification now (action first, up to 200 characters)."

	verdictOK       = "ok"
	verdictBoth     = "both"
	verdictPush     = "push"
	verdictReminder = "reminder"
)

var reminderStep = `Create a reminder in the "` + ListName + `" list of Reminders with osascript, ` +
	`named "` + Prefix + `" followed by the question: ` +
	`osascript -e 'tell application "Reminders" to make new reminder at end of reminders of list "` +
	ListName + `" with properties {name:"` + Prefix + `<question>"}'. ` +
	`It is closed automatically on the next human message.`

type payload struct {
	Active     any    `json:"stop_hook_active"`
	Transcript string `json:"transcript_path"`
}

type logEntry struct {
	Type    string `json:"type"`
	IsMeta  any    `json:"isMeta"`
	Message struct {
		Content json.RawMessage `json:"content"`
	} `json:"message"`
}

type contentBlock struct {
	Type  string `json:"type"`
	Text  string `json:"text"`
	Name  string `json:"name"`
	Input struct {
		Command string `json:"command"`
	} `json:"input"`
}

type turn struct {
	pushed, reminded bool
	last             string
}

func stopPushCheck(in io.Reader, out io.Writer) {
	path, ok := transcriptPath(in)
	if !ok {
		return
	}
	file, err := os.Open(path)
	if err != nil {
		return
	}
	defer func() { _ = file.Close() }()
	reason := reasonFor(verdict(file))
	if reason == "" {
		return
	}
	encoder := json.NewEncoder(out)
	encoder.SetIndent("", "  ")
	encoder.SetEscapeHTML(false)
	_ = encoder.Encode(struct {
		Decision string `json:"decision"`
		Reason   string `json:"reason"`
	}{"block", reason})
}

func transcriptPath(in io.Reader) (string, bool) {
	raw, err := io.ReadAll(in)
	if err != nil {
		return "", false
	}
	var p payload
	if json.Unmarshal(raw, &p) != nil || p.Active == true || p.Active == "true" {
		return "", false
	}
	info, err := os.Stat(p.Transcript)
	return p.Transcript, p.Transcript != "" && err == nil && info.Mode().IsRegular()
}

func verdict(in io.Reader) string {
	var current turn
	reader := bufio.NewReader(in)
	for {
		line, err := reader.ReadBytes('\n')
		if e, ok := parse(line); ok {
			current = current.next(&e)
		}
		if err != nil {
			break
		}
	}
	return current.verdict()
}

func parse(line []byte) (logEntry, bool) {
	var e logEntry
	if json.Unmarshal(line, &e) != nil {
		return logEntry{}, false
	}
	return e, true
}

func (t turn) next(e *logEntry) turn {
	switch {
	case e.startsTurn():
		return turn{}
	case e.Type == "assistant":
		return t.record(e.contentBlocks())
	}
	return t
}

func (t turn) record(blocks []contentBlock) turn {
	var texts []string
	for _, b := range blocks {
		t.pushed = t.pushed || b.isPush()
		t.reminded = t.reminded || b.isReminder()
		if b.Type == "text" {
			texts = append(texts, b.Text)
		}
	}
	if joined := strings.Join(texts, "\n"); joined != "" {
		t.last = joined
	}
	return t
}

func (t turn) verdict() string {
	switch {
	case !t.asks():
		return verdictOK
	case !t.pushed && !t.reminded:
		return verdictBoth
	case !t.pushed:
		return verdictPush
	case !t.reminded:
		return verdictReminder
	}
	return verdictOK
}

func (t turn) asks() bool {
	trimmed := strings.TrimRightFunc(t.last, unicode.IsSpace)
	return strings.HasSuffix(trimmed, "?") ||
		strings.Contains(t.last, "Responda") ||
		strings.Contains(t.last, "Answer sim")
}

func (e *logEntry) startsTurn() bool {
	if e.Type != "user" || e.IsMeta == true {
		return false
	}
	content := strings.TrimSpace(string(e.Message.Content))
	if strings.HasPrefix(content, `"`) {
		return true
	}
	for _, b := range e.contentBlocks() {
		if b.Type != "tool_result" {
			return true
		}
	}
	return false
}

func (e *logEntry) contentBlocks() []contentBlock {
	var raws []json.RawMessage
	if json.Unmarshal(e.Message.Content, &raws) != nil {
		return nil
	}
	blocks := make([]contentBlock, len(raws))
	for i, raw := range raws {
		if json.Unmarshal(raw, &blocks[i]) != nil {
			blocks[i] = contentBlock{}
		}
	}
	return blocks
}

func (b contentBlock) isPush() bool {
	return b.Type == "tool_use" && strings.HasSuffix(b.Name, "PushNotification")
}

func (b contentBlock) isReminder() bool {
	if b.Type != "tool_use" || b.Name != "Bash" {
		return false
	}
	for _, part := range []string{"osascript", "Reminders", ListName, Prefix} {
		if !strings.Contains(b.Input.Command, part) {
			return false
		}
	}
	return true
}

func reasonFor(verdict string) string {
	var steps string
	switch verdict {
	case verdictBoth:
		steps = pushStep + " " + reminderStep
	case verdictPush:
		steps = pushStep
	case verdictReminder:
		steps = reminderStep
	default:
		return ""
	}
	return blockLead + steps + blockTrail
}

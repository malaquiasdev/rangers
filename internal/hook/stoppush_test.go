package hook

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const (
	reminderCommand = `osascript -e 'tell application "Reminders" to make new reminder ` +
		`at end of reminders of list "Tarefas Claude" ` +
		`with properties {name:"Claude espera: which one?"}'`

	bothBlock = "{\n  \"decision\": \"block\",\n  \"reason\": \"This turn ends waiting for the human but the notification is incomplete. " +
		"Call PushNotification now (action first, up to 200 characters). " + reminderStepText +
		" Then end the turn.\"\n}\n"
	pushBlock = "{\n  \"decision\": \"block\",\n  \"reason\": \"This turn ends waiting for the human but the notification is incomplete. " +
		"Call PushNotification now (action first, up to 200 characters). Then end the turn.\"\n}\n"
	reminderBlock = "{\n  \"decision\": \"block\",\n  \"reason\": \"This turn ends waiting for the human but the notification is incomplete. " +
		reminderStepText + " Then end the turn.\"\n}\n"

	reminderStepText = `Create a reminder in the \"Tarefas Claude\" list of Reminders with osascript, ` +
		`named \"Claude espera: \" followed by the question: ` +
		`osascript -e 'tell application \"Reminders\" to make new reminder at end of reminders of list ` +
		`\"Tarefas Claude\" with properties {name:\"Claude espera: <question>\"}'. ` +
		`It is closed automatically on the next human message.`
)

type entry map[string]any

func user(text string) entry {
	return entry{"type": "user", "message": entry{"role": "user", "content": text}}
}

func userBlocks(blocks ...entry) entry {
	return entry{"type": "user", "message": entry{"role": "user", "content": blocks}}
}

func metaUser(text string) entry {
	return entry{"type": "user", "isMeta": true, "message": entry{"role": "user", "content": text}}
}

func toolResult() entry {
	return userBlocks(entry{"type": "tool_result", "tool_use_id": "t1", "content": "ok"})
}

func assistant(blocks ...entry) entry {
	return entry{"type": "assistant", "message": entry{"role": "assistant", "content": blocks}}
}

func said(text string) entry {
	return assistant(entry{"type": "text", "text": text})
}

func called(name string) entry {
	return assistant(entry{"type": "tool_use", "id": "t1", "name": name, "input": entry{}})
}

func ran(command string) entry {
	return assistant(entry{"type": "tool_use", "id": "t2", "name": "Bash", "input": entry{"command": command}})
}

func reminded() entry {
	return ran(reminderCommand)
}

func runStop(t *testing.T, payload string) (string, error) {
	t.Helper()
	var out bytes.Buffer
	dep := &Dependencies{In: strings.NewReader(payload), Out: &out}
	err := Command(t.Context(), []string{"stop-push-check"}, func(string) string { return "" }, dep)
	return out.String(), err
}

func writeTranscript(t *testing.T, lines []string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "session.jsonl")
	if err := os.WriteFile(path, []byte(strings.Join(lines, "\n")+"\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

func encode(t *testing.T, entries []entry) []string {
	t.Helper()
	lines := make([]string, 0, len(entries))
	for _, e := range entries {
		raw, err := json.Marshal(e)
		if err != nil {
			t.Fatal(err)
		}
		lines = append(lines, string(raw))
	}
	return lines
}

func payloadFor(t *testing.T, path string, active bool) string {
	t.Helper()
	raw, err := json.Marshal(map[string]any{"transcript_path": path, "stop_hook_active": active})
	if err != nil {
		t.Fatal(err)
	}
	return string(raw)
}

func classify(t *testing.T, out string) string {
	t.Helper()
	if out == "" {
		return "silent"
	}
	var decision struct {
		Decision string `json:"decision"`
		Reason   string `json:"reason"`
	}
	if err := json.Unmarshal([]byte(out), &decision); err != nil {
		t.Fatalf("output is not JSON: %q", out)
	}
	if decision.Decision != "block" {
		t.Fatalf("decision %q", decision.Decision)
	}
	push := strings.Contains(decision.Reason, "PushNotification")
	reminder := strings.Contains(decision.Reason, ListName) && strings.Contains(decision.Reason, Prefix)
	switch {
	case push && reminder:
		return "both"
	case push:
		return "push"
	case reminder:
		return "reminder"
	}
	t.Fatalf("reason asks for nothing: %q", decision.Reason)
	return ""
}

func TestStopPushCheckVerdict(t *testing.T) {
	other := strings.ReplaceAll(reminderCommand, "Tarefas Claude", "Groceries")
	plain := strings.ReplaceAll(reminderCommand, "Claude espera: ", "")
	mention := `echo "Tarefas Claude Reminders Claude espera: "`
	tests := []struct {
		name    string
		entries []entry
		active  bool
		want    string
	}{
		{"given stop_hook_active, when the turn ends with a question, then it stays silent", []entry{user("go"), said("Which one?")}, true, "silent"},
		{"given a question mark and no push or reminder, when the turn ends, then it asks for both", []entry{user("go"), said("Which one?")}, false, "both"},
		{"given a question mark followed by whitespace, when the turn ends, then it asks for both", []entry{user("go"), said("Which one?  \n")}, false, "both"},
		{"given a question mark followed by a non-breaking space, when the turn ends, then it asks for both", []entry{user("go"), said("Which one? ")}, false, "both"},
		{"given the word Responda, when the turn ends, then it asks for both", []entry{user("go"), said("Responda sim para seguir.")}, false, "both"},
		{"given the words Answer sim, when the turn ends, then it asks for both", []entry{user("go"), said("Answer sim to continue.")}, false, "both"},
		{"given a push and a reminder in the turn, when it ends with a question, then it stays silent", []entry{user("go"), called("PushNotification"), toolResult(), reminded(), toolResult(), said("Which one?")}, false, "silent"},
		{"given a prefixed push tool name, when the turn ends with a question and a reminder exists, then it stays silent", []entry{user("go"), called("mcp__phone__PushNotification"), toolResult(), reminded(), toolResult(), said("Which one?")}, false, "silent"},
		{"given a push only in the previous turn, when the turn ends with a question, then it asks for both", []entry{user("first"), called("PushNotification"), toolResult(), reminded(), toolResult(), said("Done."), user("second"), said("Which one?")}, false, "both"},
		{"given a push and no reminder, when the turn ends with a question, then it asks for the reminder", []entry{user("go"), called("PushNotification"), toolResult(), said("Which one?")}, false, "reminder"},
		{"given a reminder and no push, when the turn ends with a question, then it asks only for the push", []entry{user("go"), reminded(), toolResult(), said("Which one?")}, false, "push"},
		{"given a reminder only in the previous turn, when the turn ends with a question and a push, then it asks for the reminder", []entry{user("first"), reminded(), toolResult(), said("Done."), user("second"), called("PushNotification"), toolResult(), said("Which one?")}, false, "reminder"},
		{"given a reminder in another list, when the turn ends with a question, then it asks for the reminder", []entry{user("go"), called("PushNotification"), toolResult(), ran(other), toolResult(), said("Which one?")}, false, "reminder"},
		{"given a reminder without the waiting prefix, when the turn ends with a question, then it asks for the reminder", []entry{user("go"), called("PushNotification"), toolResult(), ran(plain), toolResult(), said("Which one?")}, false, "reminder"},
		{"given a command that only mentions the list, when the turn ends with a question, then it asks for the reminder", []entry{user("go"), called("PushNotification"), toolResult(), ran(mention), toolResult(), said("Which one?")}, false, "reminder"},
		{"given a reminder created by a tool other than Bash, when the turn ends with a question, then it asks for the reminder", []entry{user("go"), called("PushNotification"), toolResult(), assistant(entry{"type": "tool_use", "name": "Task", "input": entry{"command": reminderCommand}}), said("Which one?")}, false, "reminder"},
		{"given a final message that asks nothing, when the turn ends, then it stays silent", []entry{user("go"), said("Done. PR is open.")}, false, "silent"},
		{"given a question in the middle of the message, when the final line is a statement, then it stays silent", []entry{user("go"), said("Why? Because it was broken. Fixed.")}, false, "silent"},
		{"given a question mark before a later line, when the turn ends, then it stays silent", []entry{user("go"), said("Which one?\nPick later")}, false, "silent"},
		{"given a question followed by a tool call and a statement, when the turn ends, then it uses the last text message", []entry{user("go"), said("Which one?"), called("Bash"), toolResult(), said("Fixed it.")}, false, "silent"},
		{"given a last message without text, when the turn ends, then it uses the last message that has text", []entry{user("go"), said("Which one?"), called("Bash")}, false, "both"},
		{"given text blocks split across one message, when the turn ends, then it joins them", []entry{user("go"), assistant(entry{"type": "text", "text": "Which"}, entry{"type": "text", "text": "one?"})}, false, "both"},
		{"given tool results inside the turn, when the push came before them, then the turn still counts it", []entry{user("go"), called("PushNotification"), toolResult(), reminded(), toolResult(), called("Bash"), toolResult(), said("Which one?")}, false, "silent"},
		{"given a meta user entry inside the turn, when the turn ends with a question and a push, then the turn is not reset", []entry{user("go"), called("PushNotification"), toolResult(), reminded(), metaUser("context"), said("Which one?")}, false, "silent"},
		{"given a user entry with text blocks, when the turn ends with a question, then it starts a new turn", []entry{user("go"), called("PushNotification"), toolResult(), reminded(), userBlocks(entry{"type": "text", "text": "again"}), said("Which one?")}, false, "both"},
		{"given an empty turn, when the hook runs, then it stays silent", []entry{user("go")}, false, "silent"},
		{"given no human entry, when the transcript only has a question, then the whole transcript is the turn", []entry{said("Which one?")}, false, "both"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			path := writeTranscript(t, encode(t, tc.entries))
			out, err := runStop(t, payloadFor(t, path, tc.active))
			if err != nil {
				t.Fatal(err)
			}
			if got := classify(t, out); got != tc.want {
				t.Fatalf("got %s, want %s: %q", got, tc.want, out)
			}
		})
	}
}

func TestStopPushCheckOutput(t *testing.T) {
	tests := []struct {
		name    string
		entries []entry
		want    string
	}{
		{"given no push and no reminder, when it blocks, then the output is the pretty JSON with both steps", []entry{user("go"), said("Which one?")}, bothBlock},
		{"given a reminder and no push, when it blocks, then the output asks only for the push", []entry{user("go"), reminded(), toolResult(), said("Which one?")}, pushBlock},
		{"given a push and no reminder, when it blocks, then the output asks only for the reminder", []entry{user("go"), called("PushNotification"), toolResult(), said("Which one?")}, reminderBlock},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			path := writeTranscript(t, encode(t, tc.entries))
			out, err := runStop(t, payloadFor(t, path, false))
			if err != nil {
				t.Fatal(err)
			}
			if out != tc.want {
				t.Fatalf("got %q\nwant %q", out, tc.want)
			}
		})
	}
}

func TestStopPushCheckInput(t *testing.T) {
	question := encode(t, []entry{user("go"), said("Which one?")})
	broken := []string{question[0], "{broken", "", "[1,2]", `"text"`, question[1]}
	tests := []struct {
		name    string
		payload func(t *testing.T) string
		want    string
	}{
		{"given a transcript path that does not exist, when the hook runs, then it stays silent", func(*testing.T) string { return `{"transcript_path":"/nonexistent/session.jsonl"}` }, "silent"},
		{"given input that is not JSON, when the hook runs, then it stays silent", func(*testing.T) string { return "not json" }, "silent"},
		{"given empty input, when the hook runs, then it stays silent", func(*testing.T) string { return "" }, "silent"},
		{"given input without a transcript path, when the hook runs, then it stays silent", func(*testing.T) string { return `{"stop_hook_active":false}` }, "silent"},
		{"given a transcript path that is a directory, when the hook runs, then it stays silent", func(t *testing.T) string { return `{"transcript_path":"` + t.TempDir() + `"}` }, "silent"},
		{"given a transcript without the active flag, when the turn ends with a question, then it blocks", func(t *testing.T) string {
			return `{"transcript_path":"` + writeTranscript(t, question) + `"}`
		}, "both"},
		{"given a transcript with a broken line, when the turn ends with a question, then it still blocks", func(t *testing.T) string {
			return `{"transcript_path":"` + writeTranscript(t, broken) + `"}`
		}, "both"},
		{"given a transcript with a very long line, when the turn ends with a question, then it still blocks", func(t *testing.T) string {
			long := encode(t, []entry{user(strings.Repeat("x", 2<<20)), said("Which one?")})
			return `{"transcript_path":"` + writeTranscript(t, long) + `"}`
		}, "both"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			out, err := runStop(t, tc.payload(t))
			if err != nil {
				t.Fatal(err)
			}
			if got := classify(t, out); got != tc.want {
				t.Fatalf("got %s, want %s: %q", got, tc.want, out)
			}
		})
	}
}

func TestCommandUsage(t *testing.T) {
	tests := []struct {
		name string
		args []string
	}{
		{"given no subcommand, when running, then it returns the usage error", nil},
		{"given an unknown subcommand, when running, then it returns the usage error", []string{"nope"}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			dep := &Dependencies{In: strings.NewReader(""), Out: &bytes.Buffer{}}
			err := Command(t.Context(), tc.args, func(string) string { return "" }, dep)
			if !errors.Is(err, ErrUsage) {
				t.Fatalf("got %v", err)
			}
		})
	}
}

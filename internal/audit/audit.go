package audit

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

const (
	usageText      = "uso: audit <repo|lugar> <ação> <link> <por quê>"
	zoneName       = "America/Sao_Paulo"
	dayLayout      = "2006-01-02"
	hourLayout     = "15:04"
	fieldCount     = 4
	sessionWidth   = 8
	unknownSession = "?"
	notePermission = 0o600

	headerFormat = "---\n" +
		"type: Log\n" +
		"title: Auditoria %[1]s\n" +
		"description: \"O que os agentes fizeram sem aprovação em %[1]s: " +
		"merge, push, deploy em dev, escrita no brain\"\n" +
		"timestamp: %[1]s\n" +
		"---\n" +
		"# Auditoria %[1]s\n\n" +
		"| hora | sessão | onde | ação | link | por quê |\n" +
		"|---|---|---|---|---|---|\n"
	indexFormat = "- [[claude/auditoria/%[1]s|Auditoria %[1]s]] — o que os agentes fizeram sozinhos\n"
	rowFormat   = "| %s | %s | %s | %s | %s | %s |\n"
)

var ErrUsage = errors.New(usageText)

type Store interface {
	Exists(path string) bool
	Create(path string, data []byte) error
	Append(path string, data []byte) error
}

type Dependencies struct {
	Store Store
	Now   func() time.Time
}

type OSStore struct{}

func (OSStore) Exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func (OSStore) Create(path string, data []byte) error {
	return os.WriteFile(path, data, notePermission)
}

func (OSStore) Append(path string, data []byte) error {
	file, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, notePermission)
	if err != nil {
		return err
	}
	if _, err = file.Write(data); err != nil {
		_ = file.Close()
		return err
	}
	return file.Close()
}

func Command(_ context.Context, args []string, env func(string) string, dep *Dependencies) error {
	if len(args) < fieldCount {
		return ErrUsage
	}
	zone, err := time.LoadLocation(zoneName)
	if err != nil {
		return err
	}
	now := dep.Now().In(zone)
	dir := filepath.Join(brainDir(env), "claude", "auditoria")
	note := filepath.Join(dir, now.Format(dayLayout)+".md")
	if !dep.Store.Exists(note) {
		if err = openNote(dep.Store, dir, now.Format(dayLayout)); err != nil {
			return err
		}
	}
	hour := now.Format(hourLayout)
	row := fmt.Sprintf(rowFormat, hour, session(env), args[0], args[1], args[2], args[3])
	return dep.Store.Append(note, []byte(row))
}

func openNote(store Store, dir, day string) error {
	note := filepath.Join(dir, day+".md")
	if err := store.Create(note, []byte(fmt.Sprintf(headerFormat, day))); err != nil {
		return err
	}
	return store.Append(filepath.Join(dir, "index.md"), []byte(fmt.Sprintf(indexFormat, day)))
}

func brainDir(env func(string) string) string {
	if dir := env("BRAIN_DIR"); dir != "" {
		return dir
	}
	return filepath.Join(env("HOME"), "Documents", "brain")
}

func session(env func(string) string) string {
	id := env("CLAUDE_SESSION_ID")
	if id == "" {
		id = env("AUDIT_SESSION")
	}
	if id == "" {
		return unknownSession
	}
	if runes := []rune(id); len(runes) > sessionWidth {
		return string(runes[:sessionWidth])
	}
	return id
}

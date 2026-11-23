package agentwatch

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

const (
	defaultLimit    = 900
	defaultInterval = 120 * time.Second
	listVariable    = "AGENT_WATCH_LIST"
	intervalVar     = "AGENT_WATCH_INTERVAL"
)

var ErrUsage = errors.New("usage: rangers watch [limit-seconds]")

type Dependencies struct {
	Git   Git
	Out   io.Writer
	Now   func() time.Time
	Sleep func(ctx context.Context, d time.Duration) error
}

type watcher struct {
	dep      *Dependencies
	list     string
	limit    int64
	interval time.Duration
}

type target struct {
	name string
	file string
	dirs []string
}

func Command(ctx context.Context, args []string, env func(string) string, dep *Dependencies) error {
	limit, err := parseLimit(args)
	if err != nil {
		return err
	}
	w := &watcher{dep: dep, list: listPath(env), limit: limit, interval: parseInterval(env(intervalVar))}
	return w.loop(ctx)
}

func Sleep(ctx context.Context, d time.Duration) error {
	timer := time.NewTimer(d)
	defer timer.Stop()
	select {
	case <-timer.C:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func parseLimit(args []string) (int64, error) {
	if len(args) == 0 {
		return defaultLimit, nil
	}
	limit, err := strconv.ParseInt(args[0], 10, 64)
	if len(args) > 1 || err != nil || limit < 0 {
		return 0, ErrUsage
	}
	return limit, nil
}

func parseInterval(value string) time.Duration {
	seconds, err := strconv.ParseFloat(value, 64)
	if err != nil || seconds <= 0 {
		return defaultInterval
	}
	return time.Duration(seconds * float64(time.Second))
}

func listPath(env func(string) string) string {
	if path := env(listVariable); path != "" {
		return path
	}
	return filepath.Join(env("HOME"), ".claude", "agent-watch.txt")
}

func (w *watcher) loop(ctx context.Context) error {
	for {
		targets, ok := readList(w.list)
		if !ok {
			w.say("lista vazia")
			return nil
		}
		if w.pass(ctx, targets) {
			return nil
		}
		if err := w.dep.Sleep(ctx, w.interval); err != nil {
			return err
		}
	}
}

func readList(path string) ([]target, bool) {
	info, err := os.Stat(path)
	if err != nil || info.Size() == 0 {
		return nil, false
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, false
	}
	var targets []target
	for line := range strings.Lines(string(raw)) {
		fields := strings.Fields(line)
		if len(fields) == 0 || strings.HasPrefix(fields[0], "#") {
			continue
		}
		t := target{name: fields[0]}
		if len(fields) > 1 {
			t.file, t.dirs = fields[1], fields[2:]
		}
		targets = append(targets, t)
	}
	return targets, true
}

func (w *watcher) pass(ctx context.Context, targets []target) bool {
	now := w.dep.Now().Unix()
	for _, t := range targets {
		if message := w.check(ctx, now, t); message != "" {
			w.say(message)
			return true
		}
	}
	return false
}

func (w *watcher) check(ctx context.Context, now int64, t target) string {
	gone := "SUMIU " + t.name
	if !isFile(t.file) && len(t.dirs) == 0 {
		return gone
	}
	last := w.lastActivity(ctx, t)
	if last == 0 {
		return gone
	}
	age := now - last
	if age <= w.limit {
		return ""
	}
	return fmt.Sprintf("PARADO %s há %d min (última atividade: %s)",
		t.name, age/60, time.Unix(last, 0).Format("15:04"))
}

func (w *watcher) say(message string) {
	_, _ = fmt.Fprintln(w.dep.Out, message)
}

package schedule

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
)

const (
	label           = "com.rangers.learn"
	intervalSeconds = 7200
	usage           = "uso: rangers schedule install|uninstall|status"
)

type Launchctl interface {
	Bootout(ctx context.Context, target string) error
	Bootstrap(ctx context.Context, domain, plist string) error
	Print(ctx context.Context, target string) (string, error)
}

type Dependencies struct {
	Launchctl Launchctl
	UID       int
	Out       io.Writer
}

type job struct {
	home   string
	binary string
	plist  string
	state  string
	domain string
}

func Command(ctx context.Context, args []string, env func(string) string, dep *Dependencies) error {
	action, flags := "status", []string(nil)
	if len(args) > 0 {
		action, flags = args[0], args[1:]
	}
	if action != "install" && action != "uninstall" && action != "status" {
		return errors.New(usage)
	}
	j, err := newJob(flags, env, dep.UID)
	if err != nil {
		return err
	}
	switch action {
	case "install":
		return j.install(ctx, dep)
	case "uninstall":
		return j.uninstall(ctx, dep)
	}
	return j.show(ctx, dep)
}

func newJob(flags []string, env func(string) string, uid int) (*job, error) {
	home := env("HOME")
	binary := filepath.Join(home, ".local", "bin", "rangers")
	set := flag.NewFlagSet("schedule", flag.ContinueOnError)
	set.StringVar(&binary, "binary", binary, "caminho do binário rangers")
	if err := set.Parse(flags); err != nil {
		return nil, err
	}
	return &job{
		home:   home,
		binary: binary,
		plist:  filepath.Join(home, "Library", "LaunchAgents", label+".plist"),
		state:  filepath.Join(home, ".local", "state", "rangers", "learn"),
		domain: "gui/" + strconv.Itoa(uid),
	}, nil
}

func (j *job) install(ctx context.Context, dep *Dependencies) error {
	if err := j.writePlist(); err != nil {
		return err
	}
	_ = dep.Launchctl.Bootout(ctx, j.target())
	if err := dep.Launchctl.Bootstrap(ctx, j.domain, j.plist); err != nil {
		return err
	}
	hours := intervalSeconds / 3600
	_, err := fmt.Fprintf(dep.Out, "agendado: a cada %d horas, log em %s\n", hours, j.log())
	return err
}

func (j *job) uninstall(ctx context.Context, dep *Dependencies) error {
	_ = dep.Launchctl.Bootout(ctx, j.target())
	if err := os.Remove(j.plist); err != nil && !errors.Is(err, fs.ErrNotExist) {
		return err
	}
	_, err := fmt.Fprintln(dep.Out, "agendamento removido")
	return err
}

func (j *job) show(ctx context.Context, dep *Dependencies) error {
	out, err := dep.Launchctl.Print(ctx, j.target())
	text := describe(out)
	if err != nil {
		text = "não agendado"
	}
	_, err = fmt.Fprintln(dep.Out, text)
	return err
}

func (j *job) target() string {
	return j.domain + "/" + label
}

func (j *job) log() string {
	return filepath.Join(j.state, "launchd.log")
}

package install

import (
	"fmt"
	"io"
)

type printer struct {
	out   io.Writer
	width int
	err   error
}

func newPrinter(out io.Writer, deps []Dependency) printer {
	width := len(runtimeName)
	for _, d := range deps {
		width = max(width, len(d.Name))
	}
	return printer{out: out, width: width}
}

func (p *printer) linef(format string, args ...any) {
	if p.err == nil {
		_, p.err = fmt.Fprintf(p.out, format+"\n", args...)
	}
}

func (p *printer) row(name, status string) {
	p.linef("%-*s  %s", p.width, name, status)
}

func (p *printer) summary(missing int) error {
	if missing == 0 {
		p.linef("\nAll in place.")
		return p.err
	}
	p.linef("\n%d item(s) missing. Run rangers install", missing)
	if p.err != nil {
		return p.err
	}
	return ErrIncomplete
}

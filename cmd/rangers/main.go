package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"time"

	"github.com/malaquiasdev/rangers/internal/learn"
)

func main() {
	if len(os.Args) < 2 || os.Args[1] != "learn" {
		fmt.Fprintln(os.Stderr, "uso: rangers learn [metrics] [flags]")
		os.Exit(2)
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	dep := learn.Dependencies{
		Claude:   learn.ClaudeCLI{},
		Git:      learn.GitCLI{},
		Reminder: learn.OsascriptReminder{},
		Now:      time.Now,
		Out:      os.Stdout,
	}
	err := learn.Command(ctx, os.Args[2:], os.Getenv, dep)
	stop()
	if err != nil {
		fmt.Fprintln(os.Stderr, "rangers learn:", err)
		os.Exit(1)
	}
}

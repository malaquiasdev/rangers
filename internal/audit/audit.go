package audit

import (
	"context"
	"errors"
	"time"
)

const usageText = "uso: audit <repo|lugar> <ação> <link> <por quê>"

var (
	ErrUsage = errors.New(usageText)

	errNotImplemented = errors.New("not implemented")
)

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

func (OSStore) Exists(string) bool {
	return false
}

func (OSStore) Create(string, []byte) error {
	return errNotImplemented
}

func (OSStore) Append(string, []byte) error {
	return errNotImplemented
}

func Command(_ context.Context, _ []string, _ func(string) string, _ *Dependencies) error {
	return errNotImplemented
}

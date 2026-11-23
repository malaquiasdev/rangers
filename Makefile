.PHONY: check quick hooks

check:
	go run ./cmd/rangers check full

quick:
	go run ./cmd/rangers check quick

hooks:
	go run ./cmd/rangers hooks install

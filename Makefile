.PHONY: check quick hooks

check:
	scripts/check.sh full

quick:
	scripts/check.sh quick

hooks:
	scripts/install-hooks.sh

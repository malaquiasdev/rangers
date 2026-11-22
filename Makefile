.PHONY: check rapido hooks

check:
	scripts/checar.sh completo

rapido:
	scripts/checar.sh rapido

hooks:
	scripts/hooks-instalar.sh

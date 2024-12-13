.env:
	@# -o xtrace means set -x
	@# That causes sh to print out the result of the command.
	@# The top level trace is prefixed with `+ `, while the next level trace is prefixed with `++ `.
	@# We are only interested in the top level trace, so we pipe the trace to grep to keep those with `+ ` only.
	@# Finally we use cut to remove the `+ ` prefix.
	sh -o xtrace .env.example 2>&1 | grep '^+ ' | cut -c '3-' > .env

setup: .env tls
	@# The setup command perform database migration and necessary setup,
	@# before the server can run.
	docker compose --profile setup run --rm setup

upgrade: .env
	@# Perform migration BEFORE you run a newer version.
	@# See https://keygen.sh/docs/self-hosting/#:~:text=6%20months.-,Migrations,-Some%20upgrades%20to
	docker compose --profile upgrade run --rm upgrade

tls:
	mkcert -cert-file tls.crt -key-file tls.key "localhost"

.PHONY:
start:
	docker compose up

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

.PHONY: start
start:
	docker compose up

cat: KEYGEN_ACCOUNT_ID ::= ""
cat: KEYGEN_PRODUCT_ID ::= ""
cat:
	go build -o cat -ldflags "-X github.com/louischan-oursky/keygen-poc/pkg/buildtimeconstant.KeygenAccountID=$(KEYGEN_ACCOUNT_ID) -X github.com/louischan-oursky/keygen-poc/pkg/buildtimeconstant.KeygenProductID=$(KEYGEN_PRODUCT_ID)" .

.PHONY: clean
clean:
	rm -f cat

.PHONY: cat-image
cat-image: KEYGEN_ACCOUNT_ID ::= ""
cat-image: KEYGEN_PRODUCT_ID ::= ""
cat-image:
	docker build --tag cat:latest --build-arg KEYGEN_ACCOUNT_ID=$(KEYGEN_ACCOUNT_ID) --build-arg KEYGEN_PRODUCT_ID=$(KEYGEN_PRODUCT_ID) .

.PHONY: build
build:
		go build -v ./cmd/apiservice

RANDOM_TEXT = $(shell openssl rand -hex 32)
gen_app_key:
	sed -i '/app_key/s/\(^app_key: \).*/\1\"${RANDOM_TEXT}\"/' config.yml
migrate:
	migrate -database "$$DB_CONNECTION://$$DB_USERNAME:$$DB_PASSWORD@$$DB_HOST/$$DB_DATABASE?sslmode=disable" -path migrations/ up
.DEFAULT_GOAL := build
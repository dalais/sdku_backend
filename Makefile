include .env
export $(shell sed 's/=.*//' .env)

.PHONY: build
build:
		go build -v ./cmd/apiservice
		
RANDOM_TEXT = $(shell openssl rand -hex 16)
gen_app_key:
	@sed -i "s/${APP_KEY}/${RANDOM_TEXT}/" .env
migrate:
	@migrate -database "$$DB_CONNECTION://$$DB_USERNAME:$$DB_PASSWORD@$$DB_HOST/$$DB_DATABASE?sslmode=disable" -path migrations/ up
.DEFAULT_GOAL := build
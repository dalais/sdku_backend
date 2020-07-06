include .env
export $(shell sed 's/=.*//' .env)

.PHONY: build
build:
		go build -v ./cmd/apiservice
migrate:
	@migrate -database "$$DB_CONNECTION://$$DB_USERNAME:$$DB_PASSWORD@$$DB_HOST/$$DB_DATABASE?sslmode=disable" -path migrations/ up
.DEFAULT_GOAL := build
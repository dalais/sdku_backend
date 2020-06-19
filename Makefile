.PHONY: build
build:
		go build -v ./cmd/apiservice
		
.DEFAULT_GOAL := build
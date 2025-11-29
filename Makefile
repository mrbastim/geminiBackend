APP_NAME=gemini-backend
GO_FILES=$(shell find . -name '*.go')

.PHONY: run build tidy

run:
	go run ./cmd/app

build:
	go build -o bin/$(APP_NAME) ./cmd/app

tidy:
	go mod tidy
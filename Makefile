APP_NAME=gemini-backend
GO_FILES=$(shell find . -name '*.go')

.PHONY: run build tidy swagger swagger-clean

run: swagger
	go run ./cmd/app

build:
	go build -o bin/$(APP_NAME) ./cmd/app

tidy:
	go mod tidy

swagger:
	@echo "Generating Swagger docs"
	go install github.com/swaggo/swag/cmd/swag@latest
	$(shell go env GOPATH)/bin/swag init -g cmd/app/main.go -o docs

swagger-clean:
	rm -rf docs
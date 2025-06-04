APP_NAME = url-shortener

.PHONY: test lint run build tidy

run:
	go run ./cmd/main.go

build:
	go build -o $(APP_NAME) ./cmd/main.go

test:
	go test -v ./internal/...

lint:
	golangci-lint run ./...

tidy:
	go mod tidy

docker-up:
	docker-compose up --build

docker-down:
	docker-compose down

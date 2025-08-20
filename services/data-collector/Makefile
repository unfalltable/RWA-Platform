# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
BINARY_NAME=data-collector
BINARY_UNIX=$(BINARY_NAME)_unix

# Docker parameters
DOCKER_IMAGE=rwa-platform/data-collector
DOCKER_TAG=latest

.PHONY: all build clean test coverage deps docker-build docker-run docker-push help

all: test build

## Build the binary
build:
	$(GOBUILD) -o $(BINARY_NAME) -v ./cmd/main.go

## Clean build files
clean:
	$(GOCLEAN)
	rm -f $(BINARY_NAME)
	rm -f $(BINARY_UNIX)

## Run tests
test:
	$(GOTEST) -v ./...

## Run tests with coverage
coverage:
	$(GOTEST) -v -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html

## Download dependencies
deps:
	$(GOMOD) download
	$(GOMOD) tidy

## Build for Linux
build-linux:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) -o $(BINARY_UNIX) -v ./cmd/main.go

## Run the application
run:
	$(GOBUILD) -o $(BINARY_NAME) -v ./cmd/main.go
	./$(BINARY_NAME)

## Run with hot reload (requires air)
dev:
	air

## Format code
fmt:
	$(GOCMD) fmt ./...

## Lint code (requires golangci-lint)
lint:
	golangci-lint run

## Security scan (requires gosec)
security:
	gosec ./...

## Build Docker image
docker-build:
	docker build -t $(DOCKER_IMAGE):$(DOCKER_TAG) .

## Run Docker container
docker-run:
	docker run -p 8080:8080 --env-file .env $(DOCKER_IMAGE):$(DOCKER_TAG)

## Push Docker image
docker-push:
	docker push $(DOCKER_IMAGE):$(DOCKER_TAG)

## Generate mocks (requires mockgen)
mocks:
	mockgen -source=internal/services/price_service.go -destination=mocks/price_service_mock.go
	mockgen -source=internal/services/blockchain_service.go -destination=mocks/blockchain_service_mock.go
	mockgen -source=internal/services/news_service.go -destination=mocks/news_service_mock.go

## Database migration up
migrate-up:
	migrate -path migrations -database "$(DATABASE_URL)" up

## Database migration down
migrate-down:
	migrate -path migrations -database "$(DATABASE_URL)" down

## Create new migration
migrate-create:
	migrate create -ext sql -dir migrations -seq $(name)

## Install development tools
install-tools:
	$(GOGET) github.com/cosmtrek/air@latest
	$(GOGET) github.com/golang/mock/mockgen@latest
	$(GOGET) github.com/securecodewarrior/gosec/v2/cmd/gosec@latest
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin v1.54.2

## Show help
help:
	@echo ''
	@echo 'Usage:'
	@echo '  make [target]'
	@echo ''
	@echo 'Targets:'
	@awk 'BEGIN {FS = ":.*?## "} { \
		if (/^[a-zA-Z_-]+:.*?##.*$$/) printf "  %-20s%s\n", $$1, $$2 \
	}' $(MAKEFILE_LIST)

.PHONY: all build run test clean migrate seed

# Load environment variables
ifneq (,$(wildcard ./.env))
    include .env
    export
endif

# Default Binary Name
BINARY_NAME=data-automation

all: build

build:
	@echo "Building binary..."
	go build -o $(BINARY_NAME) main.go

run: build
	@echo "Running binary..."
	./$(BINARY_NAME)

run-report-sales:
	@echo "Generating sales trend report..."
	go run main.go report --type sales-trend

run-report-product:
	@echo "Generating product performance report..."
	go run main.go report --type product-performance

run-api-push:
	@echo "Pushing data to API..."
	go run main.go api-push

test:
	@echo "Running tests..."
	go test ./... -v

test-coverage:
	@echo "Running tests with coverage..."
	go test -coverprofile=coverage.out ./...
	grep -v /mocks/ coverage.out > coverage_filtered.out
	go tool cover -func=coverage_filtered.out

clean:
	@echo "Cleaning up..."
	go clean
	rm -f $(BINARY_NAME)
	rm -rf reports/

migrate:
	@echo "Running database migrations..."
	chmod +x scripts/migrate.sh
	./scripts/migrate.sh

seed:
	@echo "Running database migrations with seed data..."
	chmod +x scripts/migrate.sh
	./scripts/migrate.sh --seed

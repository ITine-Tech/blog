run:
    go run ./cmd/api

build:
	go build ./cmd/api -o bin/blog

run-b: build
	./bin/blog

# Run tests
test:
	go test -v ./...

# Run tests with coverage
test-coverage:
	go test -v -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Run tests without verbose output (for CI)
test-quiet:
	go test ./...

# Run linter
lint:
	golangci-lint run

health:
	curl localhost:3000/healthcheck

health-v:
	curl localhost:3000/healthcheck -v

MIGRATIONS_PATH := "./cmd/migrations"

migrate-create name:
	migrate create -seq -ext sql -dir {{MIGRATIONS_PATH}} {{name}}

migrate-up:
	migrate -path={{MIGRATIONS_PATH}} -database="postgres://postgres:mypassword@localhost/myblog?sslmode=disable" up

migrate-down args:
	migrate -path={{MIGRATIONS_PATH}} -database="postgres://postgres:mypassword@localhost/myblog?sslmode=disable" down {{args}}

gen-docs:
	go tool swag init -g "./cmd/api/main.go"

run-docs:
	go tool swag init -g "./cmd/api/main.go"
	go run ./cmd/api/.

# Clean build artifacts
clean:
	rm -rf bin/
	rm -f coverage.out coverage.html

# Show available commands
help:
	@echo "Available commands:"
	@echo "  run           - Run the API server"
	@echo "  build         - Build the application"
	@echo "  run-b         - Build and run the application"
	@echo "  test          - Run tests with verbose output"
	@echo "  test-coverage - Run tests with coverage report"
	@echo "  test-quiet    - Run tests without verbose output"
	@echo "  lint          - Run golangci-lint"
	@echo "  health        - Check health endpoint"
	@echo "  health-v      - Check health endpoint with verbose output"
	@echo "  migrate-create- Create new migration"
	@echo "  migrate-up    - Run database migrations"
	@echo "  migrate-down  - Rollback database migrations"
	@echo "  gen-docs      - Generate Swagger documentation"
	@echo "  run-docs      - Generate docs and run server"
	@echo "  clean         - Clean build artifacts"



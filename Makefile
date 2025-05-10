
build :
	@go build -o bin/berta2
run: build
	./bin/berta2

test: 
	@go test -v ./...

health:
	@curl localhost:3000/healthcheck

health-v:
	@curl localhost:3000/healthcheck -v

MIGRATIONS_PATH = ./cmd/migrations

.PHONY: migrate-create
migration:
	@migrate create -seq -ext sql -dir $(MIGRATIONS_PATH) $(filter-out $@, $(MAKECMDGOALS))

.PHONY: migrate-up
migrate-up:
	@migrate -path=$(MIGRATIONS_PATH) -database="postgres://postgres:mypassword@localhost/berta?sslmode=disable" up

.PHONY: migrate-down
migrate-down:
	@migrate -path=$(MIGRATIONS_PATH) -database="postgres://postgres:mypassword@localhost/berta?sslmode=disable" down $(filter-out $@, $(MAKECMDGOALS))

.PHONY: gen-docs
gen-docs:
	@swag init -g ./cmd/main/main.go

.PHONY: run-docs
run-docs:
	@swag init -g ./cmd/main/main.go
	@go run ./cmd/main .

.PHONY: test
test:
	@go test -v ./...
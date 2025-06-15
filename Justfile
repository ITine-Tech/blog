run:
    go run ./cmd/api

build:
	go build ./cmd/api -o bin/blog

run-b: build
	./bin/blog

test:
	go test -v ./...

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



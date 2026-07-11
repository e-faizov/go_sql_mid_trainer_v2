.PHONY: up down run test test-todo fmt

up:
	docker compose up -d

down:
	docker compose down -v

run:
	go run ./cmd/api

test:
	go test ./...

test-todo:
	TEST_DATABASE_URL=$${TEST_DATABASE_URL:-postgres://app:app@localhost:55433/app?sslmode=disable} go test -tags todo ./...

fmt:
	gofmt -w ./cmd ./internal

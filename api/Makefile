DATABASE_URL ?= postgres://postgres:postgres@localhost:5432/resumeranker?sslmode=disable

.PHONY: install-tools migrate-up migrate-down migrate-create

install-tools:
	go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

migrate-up:
	migrate -path migrations -database "$(DATABASE_URL)" up

migrate-down:
	migrate -path migrations -database "$(DATABASE_URL)" down 1

migrate-create:
	@if [ -z "$(NAME)" ]; then echo "NAME is required. Usage: make migrate-create NAME=my_migration"; exit 1; fi
	migrate create -ext sql -dir migrations -seq $(NAME)

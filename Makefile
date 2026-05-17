.PHONY: migrate-fresh migrate-up migrate-down migrate migrate-force migrate-create migrate-alter inject

DB_USER ?= postgres
DB_PASS ?=
DB_HOST ?= 127.0.0.1
DB_PORT ?= 5432
DB_NAME ?= go_zenith_on_stage
DB_SSL  ?= disable

DB_URL=postgres://$(DB_USER):$(DB_PASS)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=$(DB_SSL)

migrate-fresh:
	migrate -path migrations -database "$(DB_URL)" down
	migrate -path migrations -database "$(DB_URL)" up

migrate-up:
	migrate -path migrations -database "$(DB_URL)" up 1

migrate:
	migrate -path migrations -database "$(DB_URL)" up

migrate-down:
	migrate -path migrations -database "$(DB_URL)" down 1

migrate-force:
	migrate -path migrations -database "$(DB_URL)" force $(version)

migrate-create:
	migrate create -ext sql -dir migrations create_$(name)_table

migrate-alter:
	migrate create -ext sql -dir migrations alter_$(name)_table

inject:
	wire gen ./cmd/injection/injector.go

start:
	go run cmd/web/main.go
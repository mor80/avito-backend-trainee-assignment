ifneq ($(wildcard .env),)
include .env
export
else
$(warning WARNING: .env file not found! Using .env.example)
include .env.example
export
endif

MIGRATIONS_DIR=./migrations
DB_DRIVER=postgres
DB_URL ?= $(DB_DRIVER)://$(REVIEWER_POSTGRES__USER):$(REVIEWER_POSTGRES__PASSWORD)@$(REVIEWER_POSTGRES__HOST):$(REVIEWER_POSTGRES__PORT)/$(REVIEWER_POSTGRES__DB_NAME)?sslmode=$(REVIEWER_POSTGRES__SSLMODE)

DOCKER_APP_SERVICE=service-reviewer

.PHONY: help migrate-create migrate-up migrate-down migrate-status \
        docker-up docker-down docker-restart docker-logs docker-migrate-up docker-migrate-down \
        app-run lint

##@ Goose commands

help: ## Показать список доступных команд
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_-]+:.*?##/ { printf "  \033[36m%-18s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

## Migration commands

migrate-create: ## Создать новую миграцию
	@read -p "Enter migration name: " name; \
	goose -dir $(MIGRATIONS_DIR) -s create $$name sql

migrate-up: ## Применить все новые миграции
	goose -dir $(MIGRATIONS_DIR) $(DB_DRIVER) "$(DB_URL)" up

migrate-down: ## Откатить последнюю миграцию
	goose -dir $(MIGRATIONS_DIR) $(DB_DRIVER) "$(DB_URL)" down

migrate-status: ## Показать статус миграций
	goose -dir $(MIGRATIONS_DIR) $(DB_DRIVER) "$(DB_URL)" status

goose-help: ## Показать справку Goose
	goose help

## Docker commands

docker-up: ## Запустить контейнеры в фоне
	docker compose up -d --build

docker-down: ## Остановить контейнеры
	docker compose down

docker-restart: ## Перезапустить контейнеры
	docker compose down && docker compose up -d

docker-logs: ## Смотреть логи
	docker compose logs -f $(DOCKER_APP_SERVICE)

docker-migrate-up: ## Применить миграции внутри контейнера
	docker compose exec $(DOCKER_APP_SERVICE) goose -dir /app/migrations $(DB_DRIVER) "$(DB_URL)" up

docker-migrate-down: ## Откатить миграции внутри контейнера
	docker compose exec $(DOCKER_APP_SERVICE) goose -dir /app/migrations $(DB_DRIVER) "$(DB_URL)" down

## App commands

app-run: ## Запустить приложение
	go run ./cmd/service-reviewer/main.go

lint: ## Запустить golangci-lint
	golangci-lint run

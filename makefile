SHELL := /bin/bash

# Path to Docker‐Compose file
Docker := docker-compose -f docker-compose.yml
MIGRATIONS_PATH = ./db/migrate/
DOCKER_MIGRATE = docker run --rm -v migrate-volume:/data --network="host" migrate/migrate:latest

DB_ADDR = postgres://admin:adminpassword@localhost:5432/dashboard?sslmode=disable
# Services
SERVICES := ingestor broker worker aggregator dashboard

.PHONY: help up down restart build build-images logs clean migrate-create migrate-up migrate-down sync-from-volume\
        run-aggregator run-worker run-ingestor run-broker run-dashboard


up:
	@echo "⏫ Bringing up all services..."
	$(Docker) up -d

down:
	@echo "⏬ Tearing down all services..."
	$(Docker) down

restart: down up

build:
	@echo "🛠️  Building Docker images..."
	$(Docker) build -d

logs:
ifndef SERVICE
	$(error SERVICE is not set. Usage: make logs SERVICE=<service name>)
endif
	@echo "📜 Tailing logs for service '$(SERVICE)'..."
	$(Docker) logs -f $(SERVICE)

## Local runs (for quick dev iterations)

run-aggregator:
	@echo "▶ Running Aggregator locally..."
	cd aggregator && go run main.go

run-worker:
	@echo "▶ Running Worker locally..."
	cd worker && python worker.py

run-ingestor:
	@echo "▶ Running Ingestor locally..."
	cd ingestor && go run main.go

run-dashboard:
	@echo "▶ Running Dashboard locally..."
	cd dashboard && go run main.go

clean:
	@echo "🧹 Removing all containers, networks, volumes, and images..."
	docker system prune -af --volumes

.PHONY: migrate-create
migrate-create:
	$(DOCKER_MIGRATE) create -seq -ext sql -dir /data $(name)



.PHONY: migrate-up
migrate-up:
	$(DOCKER_MIGRATE) -path=/data -database "$(DB_ADDR)" up

.PHONY: migrate-down
migrate-down:
	$(DOCKER_MIGRATE) -path=/data -database "$(DB_ADDR)" down -all

.PHONY: sync-from-volume
sync-from-volume:
	docker run --rm -v migrate-volume:/from -v "$(CURDIR)/db/migrate:/to" alpine sh -c "cp -r /from/. /to"

sync-to-volume:
	docker run --rm -v "$(CURDIR)/db/migrate:/from" -v migrate-volume:/to alpine sh -c "cp -r /from/. /to"

SHELL := /bin/bash

# Path to Docker‐Compose file
Docker := docker-compose -f docker-compose.yml

# Services
SERVICES := ingestor broker worker aggregator dashboard

.PHONY: help up down restart build build-images logs clean \
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
	$(Docker) build

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

run-broker:
	@echo "▶ Running Broker locally..."
	cd broker && go run main.go

run-dashboard:
	@echo "▶ Running Dashboard locally..."
	cd dashboard && go run main.go

## Cleanup everything (use with caution!)

clean:
	@echo "🧹 Removing all containers, networks, volumes, and images..."
	docker system prune -af --volumes

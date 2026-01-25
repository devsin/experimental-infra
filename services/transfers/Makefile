SHELL := /bin/bash
APP := transfers
REGISTRY ?= local
TAG ?= latest
IMAGE := $(REGISTRY)/$(APP):$(TAG)
DOCKER_COMPOSE := docker compose -f docker-compose.yml

.PHONY: build run lint docker-build docker-push compose-up compose-down compose-reset compose-logs smoke

build:
	go build -o bin/$(APP) ./cmd/transfers

run:
	go run ./cmd/transfers

lint:
	command -v golangci-lint >/dev/null 2>&1 || (echo "golangci-lint not installed. install from https://golangci-lint.run/" && exit 1)
	golangci-lint run ./...

docker-build:
	docker build -t $(IMAGE) .

docker-push: docker-build
	docker push $(IMAGE)

up:
	$(DOCKER_COMPOSE) up --build -d

down:
	$(DOCKER_COMPOSE) down

reset:
	$(DOCKER_COMPOSE) down -v

logs:
	$(DOCKER_COMPOSE) logs -f $(APP)

smoke:
	bash scripts/smoke.sh

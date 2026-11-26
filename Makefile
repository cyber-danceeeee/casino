COMPOSE ?= docker compose

.PHONY: init build up down logs ps install check dev-api dev-h5 dev-admin
init:
	python3 scripts/init-env.py
build:
	$(COMPOSE) build api web
up:
	$(COMPOSE) up -d --build --wait --wait-timeout 600
down:
	$(COMPOSE) down
logs:
	$(COMPOSE) logs -f --tail=100
ps:
	$(COMPOSE) ps -a
install:
	npm --prefix apps/h5 ci
	npm --prefix apps/admin ci
check:
	npm --prefix apps/h5 run build
	npm --prefix apps/admin run build
	npm --prefix apps/admin test
	cd server && GOTOOLCHAIN=go1.24.13 go test ./...
dev-api:
	cd server && GOTOOLCHAIN=go1.24.13 go run ./cmd/casino -c configs/casino.yml
dev-h5:
	npm --prefix apps/h5 run dev
dev-admin:
	npm --prefix apps/admin run dev -- --port 5174

GO_BIN_DIR := $(shell go env GOBIN)
ifeq ($(GO_BIN_DIR),)
GO_BIN_DIR := $(shell go env GOPATH)/bin
endif
AIR_BIN := $(GO_BIN_DIR)/air

.PHONY: processing-install processing-test core-test ui-build test run-core run-core-dev run-ui process-sample air-install network up down seed-nexus-rules seed-companion-assist

processing-install:
	cd processing/python && \
	if [ ! -x .venv/bin/python ]; then python3 -m venv .venv; fi && \
	. .venv/bin/activate && pip install -e ".[dev]"

process-sample: processing-install
	cd processing/python && . .venv/bin/activate && argos-processing process-capture-group --input ../../sample --output ../../var/outputs/sample

processing-test: processing-install
	cd processing/python && . .venv/bin/activate && pytest

core-test:
	cd core && go test ./...

ui-build:
	cd ui && npm run build

test: core-test processing-test ui-build

run-core:
	cd core && go run ./cmd/api

air-install:
	go install github.com/air-verse/air@latest

run-core-dev:
	@if [ ! -x "$(AIR_BIN)" ]; then $(MAKE) air-install; fi
	cd core && "$(AIR_BIN)"

run-ui:
	cd ui && npm run dev

network:
	@docker network inspect axis-local >/dev/null 2>&1 || docker network create axis-local >/dev/null

up: network
	docker compose up --build -d

down:
	docker compose down

seed-nexus-rules:
	bash scripts/seed-nexus-rules.sh

seed-companion-assist:
	bash scripts/seed-companion-assist.sh

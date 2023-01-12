.DEFAULT_GOAL := help
PQ_HOST = $(shell docker inspect -f '{{range.NetworkSettings.Networks}}{{.IPAddress}}{{end}}' postgres)
PQ_CONTAINER = postgres

install-swag: 
	@go install github.com/swaggo/swag/cmd/swag@latest;

.PHONY: docs
docs: install-swag ## Create the swagger docs
	@go mod vendor; 
	@$(shell go env GOPATH)/bin/swag init -g internal/explorer/server.go --parseVendor;
	@rm -rf vendor;

build: ## Bulil the project
	@cd cmds/proxy_server && CGO_ENABLED=0 GOOS=linux go build -ldflags "-w -s -X main.GitCommit=$(shell git describe --tags --abbrev=0) -extldflags '-static'"  -o server

db-start: ## Start postgres server on a docker container
	@docker run --rm --name $(PQ_CONTAINER) \
		-e POSTGRES_USER=postgres \
		-e POSTGRES_PASSWORD=postgres \
		-e POSTGRES_DB=tfgrid-graphql \
		-d \
		postgres

db-fill: ## Fill the database with a randomly generated data
	@echo "Loading...   It takes some time."
	@cd ./tools/db &&\
	go run . \
		--postgres-host $(PQ_HOST) \
		--postgres-db tfgrid-graphql \
		--postgres-password postgres \
		--postgres-user postgres \
		--reset \
		--seed 13

db-dump: ## Load a dump of the database 		(Args: `p=<path/to/file.sql`)
	@docker cp $(p) postgres:/dump.sql;
	@docker exec $(PQ_CONTAINER) bash -c "psql -U postgres  -d tfgrid-graphql < ./dump.sql"

db-stop: ## Stop the database container if running
	@if [ ! "$(shell docker ps | grep '$(PQ_CONTAINER)' )" = "" ]; then \
		docker stop postgres; \
	fi

start: ## Start the proxy server
	@go run cmds/proxy_server/main.go \
		-no-cert \
		--address :8080 \
		--log-level debug \
		--postgres-host $(PQ_HOST) \
		--postgres-db tfgrid-graphql \
		--postgres-password postgres \
		--postgres-user postgres

restart: db-stop db-start sleep db-fill start ## Full start of the database and the server

sleep:
	@sleep 5

test-queries: ## Run all queries tests
	@cd tests/queries/ &&\
	go test -v \
		--seed 13 \
		--postgres-host $(PQ_HOST) \
		--postgres-db tfgrid-graphql \
		--postgres-password postgres \
		--postgres-user postgres \
		--endpoint http://localhost:8080

test-query: ## Run specific test query 			(Args: `t=TestName`).
	@cd tests/queries/ &&\
	go test -v \
		--seed 13 \
		--postgres-host $(PQ_HOST) \
		--postgres-db tfgrid-graphql \
		--postgres-password postgres \
		--postgres-user postgres \
		--endpoint http://localhost:8080 \
		-run $(t)

test-unit: ## Run only unit tests
	@go test -v ./pkg/client

test-all: test-unit test-queries ## Run all unit/queries tests

.PHONY: help
help:
	@printf "%s\n" "Avilable targets:"
	@grep -E '^[a-zA-Z0-9_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m  make %-15s\033[0m %s\n", $$1, $$2}'
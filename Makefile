.PHONY: docs

docs:
	go mod vendor; swag init -g internal/explorer/server.go --parseVendor; rm -rf vendor
build:
	cd cmds/proxy_server && CGO_ENABLED=0 GOOS=linux go build -ldflags "-w -s -X main.GitCommit=$(shell git describe --tags --abbrev=0) -extldflags '-static'"  -o server

run_db:
	docker run --rm --name postgres \
		-e POSTGRES_USER=postgres \
		-e POSTGRES_PASSWORD=postgres \
		-e POSTGRES_DB=tfgrid-graphql \
		-d \
		postgres

fill_db:
	cd ./tools/db &&\
	go run . \
		--postgres-host $(shell docker inspect -f '{{range.NetworkSettings.Networks}}{{.IPAddress}}{{end}}' postgres) \
		--postgres-db tfgrid-graphql \
		--postgres-password postgres \
		--postgres-user postgres \
		--reset \
		--seed 13

start_server:
	go run cmds/proxy_server/main.go \
		-no-cert \
		--address :8080 \
		--log-level debug \
		--postgres-host $(shell docker inspect -f '{{range.NetworkSettings.Networks}}{{.IPAddress}}{{end}}' postgres) \
		--postgres-db tfgrid-graphql \
		--postgres-password postgres \
		--postgres-user postgres

run_tests:
	cd tests/queries/ &&\
	go test -v \
		--seed 13 \
		--postgres-host $(shell docker inspect -f '{{range.NetworkSettings.Networks}}{{.IPAddress}}{{end}}' postgres) \
		--postgres-db tfgrid-graphql \
		--postgres-password postgres \
		--postgres-user postgres \
		--endpoint http://localhost:8080

# run specific test `make run_test t=<test_name>`
run_test:
	cd tests/queries/ &&\
	go test -v \
		--seed 13 \
		--postgres-host $(shell docker inspect -f '{{range.NetworkSettings.Networks}}{{.IPAddress}}{{end}}' postgres) \
		--postgres-db tfgrid-graphql \
		--postgres-password postgres \
		--postgres-user postgres \
		--endpoint http://localhost:8080 \
		-run $(t)

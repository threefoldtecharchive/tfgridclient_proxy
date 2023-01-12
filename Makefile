.PHONY: docs

PQ_HOST = $(shell docker inspect -f '{{range.NetworkSettings.Networks}}{{.IPAddress}}{{end}}' postgres)

docs:
	go mod vendor; swag init -g internal/explorer/server.go --parseVendor; rm -rf vendor
build:
	cd cmds/proxy_server && CGO_ENABLED=0 GOOS=linux go build -ldflags "-w -s -X main.GitCommit=$(shell git describe --tags --abbrev=0) -extldflags '-static'"  -o server

db_start:
	docker run --rm --name postgres \
		-e POSTGRES_USER=postgres \
		-e POSTGRES_PASSWORD=postgres \
		-e POSTGRES_DB=tfgrid-graphql \
		-d \
		postgres

db_fill:
	cd ./tools/db &&\
	go run . \
		--postgres-host $(PQ_HOST) \
		--postgres-db tfgrid-graphql \
		--postgres-password postgres \
		--postgres-user postgres \
		--reset \
		--seed 13

db_dump:
	docker cp $(p) postgres:/dump.sql &&\
	docker exec postgres bash -c "psql -U postgres  -d tfgrid-graphql < ./dump.sql"

db_stop:
	if [ ! "$(shell docker ps | grep postgres )" = "" ]; then \
		docker stop postgres; \
	fi

start:
	go run cmds/proxy_server/main.go \
		-no-cert \
		--address :8080 \
		--log-level debug \
		--postgres-host $(PQ_HOST) \
		--postgres-db tfgrid-graphql \
		--postgres-password postgres \
		--postgres-user postgres

restart: db_stop db_start sleep db_fill start

sleep:
	sleep 2

tests:
	cd tests/queries/ &&\
	go test -v \
		--seed 13 \
		--postgres-host $(PQ_HOST) \
		--postgres-db tfgrid-graphql \
		--postgres-password postgres \
		--postgres-user postgres \
		--endpoint http://localhost:8080

test:
	cd tests/queries/ &&\
	go test -v \
		--seed 13 \
		--postgres-host $(PQ_HOST) \
		--postgres-db tfgrid-graphql \
		--postgres-password postgres \
		--postgres-user postgres \
		--endpoint http://localhost:8080 \
		-run $(t)

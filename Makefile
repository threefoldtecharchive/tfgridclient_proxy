.PHONY: docs

docs:
	go mod vendor; swag init -g internal/explorer/server.go --parseVendor; rm -rf vendor
build:
	cd cmds/proxy_server && CGO_ENABLED=0 GOOS=linux go build -ldflags "-w -s -X main.GitCommit=$(shell git describe --tags --abbrev=0) -extldflags '-static'"  -o server


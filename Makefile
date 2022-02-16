.PHONY: docs

docs:
	go mod vendor; swag init -g internal/explorer/server.go --parseVendor; rm -rf vendor

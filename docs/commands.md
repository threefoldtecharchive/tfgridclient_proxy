# Commands

The Makefile make it easier to do mostly all the frequently commands needed to work on the project.

## Work on Docs

we are using [swaggo/swag](https://github.com/swaggo/swag) to generate swagger docs based on the annotation inside the code.

- install swag executable binary

  ```bash
  go install github.com/swaggo/swag/cmd/swag@latest
  ```

- now if you check the binary directory inside go directory you will find the executable file.

  ```bash
  ls $(go env GOPATH)/bin
  ```

- to run swag you can either use the full path `$(go env GOPATH)/bin/swag` or export go binary to `$PATH`

  ```bash
  export PATH=$PATH:$(go env GOPATH)/bin
  ```

- use swag to format code comments.

  ```bash
  swag fmt
  ```

- update the docs

  ```bash
  swag init
  ```

- to parse external types from vendor

  ```bash
  swag init --parseVendor
  ```

- for a full generate docs command

  ```bash
  make docs
  ```

## To start the GridProxy server

After preparing the postgres database you can `go run` the main file in `cmds/proxy_server/main.go` which responsible for starting all the needed server/clients.

The server options

| Option             | Description                                                                                                             |
| ------------------ | ----------------------------------------------------------------------------------------------------------------------- |
| -address           | Server ip address (default `":443"`)                                                                                    |
| -ca                | certificate authority used to generate certificate (default `"https://acme-staging-v02.api.letsencrypt.org/directory"`) |
| -cert-cache-dir    | path to store generated certs in (default `"/tmp/certs"`)                                                               |
| -domain            | domain on which the server will be served                                                                               |
| -email             | email address to generate certificate with                                                                              |
| -log-level         | log level `[debug\|info\|warn\|error\|fatal\|panic]` (default `"info"`)                                                 |
| -no-cert           | start the server without certificate                                                                                    |
| -postgres-db       | postgres database                                                                                                       |
| -postgres-host     | postgres host                                                                                                           |
| -postgres-password | postgres password                                                                                                       |
| -postgres-port     | postgres port (default 5432)                                                                                            |
| -postgres-user     | postgres username                                                                                                       |
| -tfchain-url       | tF chain url (default `"wss://tfchain.dev.grid.tf/ws"`)                                                                 |
| -relay-url         | RMB relay url (default`"wss://relay.dev.grid.tf"`)                                                                      |
| -mnemonics         | Dummy user mnemonics for relay calls                                                                                    |
| -v                 | shows the package version                                                                                               |

For a full server setup:

```bash
make restart
```

## Run tests

There is two types of tests in the project

- Unit Tests
  - Found in `pkg/client/*_test.go`
  - Run with `go test -v ./pkg/client`
- Integration Tests
  - Found in `tests/queries/`
  - Run with:

    ```bash
    go test -v \
    --seed 13 \
    --postgres-host <postgres-ip> \
    --postgres-db tfgrid-graphql \
    --postgres-password postgres \
    --postgres-user postgres \
    --endpoint <server-ip> \
    --mnemonics <insert user mnemonics>
    ```

  - Or to run a specific test you can append the previous command with

    ```bash
    -run <TestName>
    ```

    You can found the TestName in the `tests/queries/*_test.go` files.

To run all the tests use

```bash
make test-all
```

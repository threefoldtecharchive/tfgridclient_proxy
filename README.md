# Grid proxy server

![golang workflow](https://github.com/threefoldtech/grid_proxy_server/actions/workflows/go.yml/badge.svg)

Interact with TFgridDB using rest APIs

## Live Instances

- Dev network: <https://gridproxy.dev.grid.tf>
  - Swagger: https://gridproxy.dev.grid.tf/swagger/index.html
- Test network: <https://gridproxy.test.grid.tf>
  - Swagger: https://gridproxy.test.grid.tf/swagger/index.html
- Main network: <https://gridproxy.grid.tf>
  - Swagger: https://gridproxy.grid.tf/swagger/index.html

## Run for Development & Testing

```bash
$ make help
```

To list all the available tasks for running:

- Database
- Server
- Tests


## Prerequisites

1. Golang compiler > 1.13 to run the grid proxy server.
2. Postgres database

## Generate swagger doc files

1. Install swag and export to exec path

    ```bash
    go install github.com/swaggo/swag/cmd/swag@latest
    export PATH=$(go env GOPATH)/bin:$PATH
    ```

2. Run `make docs`.

## Build

  ```bash
  GIT_COMMIT=$(git describe --tags --abbrev=0) && \
  cd cmds/proxy_server && CGO_ENABLED=0 GOOS=linux go build -ldflags "-w -s -X main.GitCommit=$GIT_COMMIT -extldflags '-static'"  -o server
  ```

## Development Run

- To run in development envornimnet see [here](tools/db/README.md) how to generate test db or load a db dump then use:
    ```sh
    go run cmds/proxy_server/main.go --address :8080 --log-level debug -no-cert --postgres-host 127.0.0.1 --postgres-db tfgrid-graphql --postgres-password postgres --postgres-user postgres
    ```
- all server Options:

| Option | Description |
| --- | --- |
| -address | Server ip address (default `":443"`)  |
| -log-level | log level `[debug\|info\|warn\|error\|fatal\|panic]` (default `"info"`)  |
| -no-cert | start the server without certificate  |
| -postgres-db | postgres database  |
| -postgres-host | postgres host  |
| -postgres-password | postgres password  |
| -postgres-port | postgres port (default 5432)  |
| -postgres-user | postgres username  |
| -v | shows the package version |


- Then visit `http://localhost:8080/<endpoint>`

## Production Run

- Download the latest binary [here](https://github.com/threefoldtech/tfgridclient_proxy/releases)
- add the execution permission to the binary and move it to the bin directory

  ```bash
  chmod +x ./gridproxy-server
  mv ./gridproxy-server /usr/local/bin/gridproxy-server
  ```

- Add a new systemd service

```bash
# create msgbus service
cat << EOF > /etc/systemd/system/gridproxy-server.service
[Unit]
Description=grid proxy server
After=network.target
After=msgbus.service

[Service]
ExecStart=gridproxy-server --postgres-host 127.0.0.1 --postgres-db db --postgres-password password --postgres-user postgres
Type=simple
Restart=always
User=root
Group=root

[Install]
WantedBy=multi-user.target
Alias=gridproxy.service
EOF
```

- enable the service

  ```
   systemctl enable gridproxy.service
  ```

- start the service

  ```
  systemctl start gridproxy.service
  ```

- check the status

  ```
  systemctl status gridproxy.service
  ```

- The command options:
  - postgre-\*: postgres connection info.

## To upgrade the machine

- just replace the binary with the new one and apply

```
systemctl restart gridproxy-server.service
```

- it you have changes in the `/etc/systemd/system/gridproxy-server.service` you have to run this command first

```
systemctl daemon-reload
```

## Dockerfile

To build & run dockerfile

```bash
docker build -t threefoldtech/gridproxy .
docker run --name gridproxy -e POSTGRES_HOST="127.0.0.1" -e POSTGRES_PORT="5432" -e POSTGRES_DB="db" -e POSTGRES_USER="postgres" -e POSTGRES_PASSWORD="password" --cap-add=NET_ADMIN threefoldtech/gridproxy
```

## Update helm package

- Do `helm lint charts/gridproxy`
- Regenerate the packages `helm package -u charts/gridproxy`
- Regenerate index.yaml `helm repo index --url https://threefoldtech.github.io/tfgridclient_proxy/ .`
- Push your changes

## Install the chart using helm package

- Adding the repo to your helm

  ```bash
  helm repo add gridproxy https://threefoldtech.github.io/tfgridclient_proxy/
  ```

- install a chart

  ```bash
  helm install gridproxy/gridproxy
  ```

## Release
- Update the `appVersion` in `charts/Chart.yaml`. (push, open PR, merge)
- Draft new release with [Github UI Releaser](https://github.com/threefoldtech/tfgridclient_proxy/releases/new) 
  - In the tags dropdown menu write the new tag `appVersion` and create it.
  - Generate release notes
  - Mark as release or pre-release and publish 
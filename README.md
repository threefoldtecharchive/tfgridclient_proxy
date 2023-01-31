<!-- Header -->
<div class="header" align="center">
    <h1>TFGrid Proxy</h1>
    <p><strong>A RESTful API to TFGridDB + RMB Proxy.</strong></p>
    <img src="https://github.com/threefoldtech/tfgridclient_proxy/actions/workflows/unit.yml/badge.svg" > <img src="https://github.com/threefoldtech/tfgridclient_proxy/actions/workflows/integration.yml/badge.svg" > <img src="https://github.com/threefoldtech/tfgridclient_proxy/actions/workflows/release.yml/badge.svg" > <img src="https://github.com/threefoldtech/tfgridclient_proxy/actions/workflows/go.yml/badge.svg" >
</div>

<!-- ToC -->

## Table of Content

- [About](##About)
- [Used Technologies & Prerequisites](##Used-Technologies-&-Prerequisites)
- [Start for Development](##Start-Development)
- [How to use the project](##How-to-use-the-project)
- [Setup for Production](##Setup-for-Production)

<!-- About -->

## About

The TFGrid Proxy contains two projects that can do a lot of grid usage work on behave of you like sending/receiving requests from and to twins on the grid and retrieving various information about the grid stats.

- Grid Explorer:

  The explorer can retrieve a lot of distracted grid/chain data and organize it in standard objects besides providing filtering, limitation, and pagination. [More About The Explorer](./docs/explorer.md)

- RMB Proxy:

  Every twin on the chain should run a local RMB instance along with a Redis server to be able to send/receive requests from other twins. The Proxy makes this easier by running the required services and with your provided `mnemonics` it can act on the chain on behave of you. [More About The Proxy](./docs/proxy.md)

  The Grid Proxy is very helpful when it used with other clients like:

  - The [Dashboard](https://github.com/threefoldtech/tfgrid_dashboard)
  - The [Playground](https://github.com/threefoldtech/grid_weblets)
  - The [GridClient](https://github.com/threefoldtech/grid3_client_ts)


<!-- Prerequisites -->
## Used Technologies & Prerequisites

1. **GoLang**: Mainly the two parts of the project written in `Go 1.17`, otherwise you can just download the compiled binaries from github [releases](https://github.com/threefoldtech/tfgridclient_proxy/releases)
   > Explorer:
2. **Postgresql**: Used to load the TFGrid DB
3. **Docker**: Containerize the running services such as Postgres and Redis.
   > RMB:
4. **MsgBus**: Aims to abstract inter-process communication between multiple processes running over multiple nodes.
5. **Redis**: Used as a message queue.
6. **Yggdrasil Network**: Peer-to-peer decentralized routing protocol among all chain twins. [see official docs](https://yggdrasil-network.github.io/)
7. **Twin ID**: Can be obtained from the dashboard with your Yggdrasil IP.
1. **GoLang**: Mainly the two parts of the project written in `Go 1.17`

- Explorer:

  2. **Postgresql**: Used to load the TFGrid DB
  3. **Docker**: Containerize the running services such as Postgres and Redis.

- RMB:

  4. **MsgBus**: Aims to abstract inter-process communication between multiple processes running over multiple nodes.
  5. **Redis**: Used as a message queue.

- Chain:

  6. **Mnemonics**: Secret seeds for your identity on the tf chain.
  7. **Yggdrasil Network**: Peer-to-peer decentralized routing protocol among all chain twins. [see official docs](https://yggdrasil-network.github.io/)
  8. **Twin ID**: Can be obtained from the dashboard with your Yggdrasil IP.

For more about the prerequisites and how to set up and configure them. follow the [Setup guide](./docs/setup.md)

<!-- Development -->

## Start for Development

To start the services for development or testing make sure first you have all the [Prerequisites](#Used-Technologies-&-Prerequisites).

- Clone this repo
  ```bash
   git clone https://github.com/threefoldtech/tfgridclient_proxy.git
   cd tfgridclient_proxy/
  ```
- The `Makefile` has all that you need to deal with Db, Explorer, RMB, Tests, and Docs.
  ```bash
   make help     # list all the available subcommands.
  ```
- For a quick test explorer server.
  ```bash
   make restart
  ```
  Now you can access the server at [:8080](http://loaclhost:8080)
- Run the tests
  ```bash
   make test-all
  ```
- Generate docs.

  ```bash
   make docs
  ```

  For more illustrations about the commands needed to work on the project. see [commands.md](./docs/commands.md). And for more about the project structure and contributions guidelines check [contributions.md](./docs/contributions.md)

<!-- Usage -->

## How to use the project

If you don't want to care about setting up your instance you can use one of the live instances. each works against a different TFChain network.

- Access instance for Chain: [DevNet](https://gridproxy.dev.grid.tf/), [QaNet](https://gridproxy.qa.grid.tf/), [TestNet](https://gridproxy.test.grid.tf/) and [MainNet](https://gridproxy.grid.tf/).

- Or follow the [development guide](#Start-Development) to run yours.
  By default, the instance runs against devnet. to configure that you will need to config this while running the server.

Either way, using a live instance or running yours. you will be able to access a swagger endpoint `<URL>/swagger/index.html` where you will find a list of all endpoints with descriptions about their usage and supported queries for filtering, limitation, or pagination.

> Note: You may face some differences between each instance and the others. that is normal because each network is in a different stage of development and works correctly with others parts of the Grid on the same network.

<!-- Production-->

## Setup for Production

### Get and install the binary

- You can either build the project:
  ```bash
   make build
   chmod +x cmd/proxy_server/server \
    && mv cmd/proxy_server/server /usr/local/bin/gridproxy-server
  ```
- Or download a release:
  Check the [releases](https://github.com/threefoldtech/tfgridclient_proxy/releases) page and edit the next command with the chosen version.
  ```bash
   wget https://github.com/threefoldtech/tfgridclient_proxy/releases/download/v1.6.7-rc2/tfgridclient_proxy_1.6.7-rc2_linux_amd64.tar.gz \
    && tar -xzf tfgridclient_proxy_1.6.7-rc2_linux_amd64.tar.gz \
    && chmod +x server \
    && mv server /usr/local/bin/gridproxy-server
  ```

### Add as a Systemd service

- Create the service file

  ```bash
  cat << EOF > /etc/systemd/system/gridproxy-server.service
  [Unit]
  Description=grid proxy server
  After=network.target
  After=msgbus.service

  [Service]
  ExecStart=gridproxy-server --domain gridproxy.dev.grid.tf --email omar.elawady.alternative@gmail.com -ca https://acme-v02.api.letsencrypt.org/directory --substrate wss://tfchain.dev.grid.tf/ws --postgres-host 127.0.0.1 --postgres-db db --postgres-password password --postgres-user postgres
  Type=simple
  Restart=always
  User=root
  Group=root

  [Install]
  WantedBy=multi-user.target
  Alias=gridproxy.service
  EOF
  ```

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

[Service]
ExecStart=gridproxy-server --domain gridproxy.dev.grid.tf --email omar.elawady.alternative@gmail.com -ca https://acme-v02.api.letsencrypt.org/directory --postgres-host 127.0.0.1 --postgres-db db --postgres-password password --postgres-user postgres
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
  - domain: the host domain which will generate ssl certificate to.
  - email: the mail used to run generate the ssl certificate.
  - ca: certificate authority server url, e.g.
    - let's encrypt staging: `https://acme-staging-v02.api.letsencrypt.org/directory`
    - let's encrypt production: `https://acme-v02.api.letsencrypt.org/directory`
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

## Development Run

- To run in development envornimnet see [here](tools/db/README.md) how to generate test db or load a db dump then use:
    ```sh
    go run cmds/proxy_server/main.go --address :8080 --log-level debug -no-cert --postgres-host 127.0.0.1 --postgres-db tfgrid-graphql --postgres-password postgres --postgres-user postgres
    ```
- all server Options:

| Option | Description |
| --- | --- |
| -address | Server ip address (default `":443"`)  |
| -ca | certificate authority used to generate certificate (default `"https://acme-v02.api.letsencrypt.org/directory"`)  |
| -cert-cache-dir | path to store generated certs in (default `"/tmp/certs"`)  |
| -domain | domain on which the server will be served  |
| -email | email address to generate certificate with  |
| -log-level | log level `[debug\|info\|warn\|error\|fatal\|panic]` (default `"info"`)  |
| -no-cert | start the server without certificate  |
| -postgres-db | postgres database  |
| -postgres-host | postgres host  |
| -postgres-password | postgres password  |
| -postgres-port | postgres port (default 5432)  |
| -postgres-user | postgres username  |
| -v | shows the package version |


- Then visit `http://localhost:8080/<endpoint>`

## Dockerfile

To build & run dockerfile

```bash
docker build -t threefoldtech/gridproxy .
docker run --name gridproxy -e POSTGRES_HOST="127.0.0.1" -e POSTGRES_PORT="5432" -e POSTGRES_DB="db" -e POSTGRES_USER="postgres" -e POSTGRES_PASSWORD="password" threefoldtech/gridproxy
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
- Command options:
  - domain: the host domain which will generate ssl certificate to.
  - email: the mail used to run generate the ssl certificate.
  - ca: certificate authority server url
  - substrate: substrate websocket link.
  - postgres-\*: postgres connection info.

For more about the production environment and how the deployed instances are upgraded. see [production.md](./docs/production.md)

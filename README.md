# Grid proxy server

![golang workflow](https://github.com/threefoldtech/grid_proxy_server/actions/workflows/go.yml/badge.svg)

Interact with TFgridDB using rest APIs

## Prerequisites

1. A msgbusd instance must be running on the node. This client uses RMB (message bus) to send messages to nodes, and get the responses.
2. A valid TwinID.
3. yggdrasil service running with a valid ip assigned to the twin on polkadot.
4. Golang compiler > 1.13 to run the grid proxy server.

## Build and run

- Start the msgbus with your twin ID
- Then to run `go run cmds/proxy_server/main.go`
- To run without certificate use `go run cmds/proxy_server/main.go -no-cert`
- To build `CGO_ENABLED=0 && GIT_COMMIT=$(git describe --tags --abbrev=0) && go build -ldflags "-X main.GitCommit=$GIT_COMMIT -extldflags '-static'"  cmds/proxy_server/main.go`
- Then visit `http://localhost:8080/<endpoint>`

## Production Run

- Start the msgbus systemd service with a machine twinId linked to its yggdrasil IP or public ip if there, [download and more info](https://github.com/threefoldtech/go-rmb)
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
ExecStart=gridproxy-server --domain gridproxy.dev.grid.tf --email omar.elawady.alternative@gmail.com -ca https://acme-v02.api.letsencrypt.org/directory --substrate wss://tfchain.dev.grid.tf/ws --explorer https://graphql.dev.grid.tf/graphql
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
  - ca: certificate authority server url
  - substrate: substrate websocket link.
  - explorer: explorer url which will get queries from.

## To upgrade the machine
- just replace the binary with the new one and apply
```
systemctl restart gridproxy-server.service
```
- it you have changes in the `/etc/systemd/system/gridproxy-server.service` you have to run this command first
```
systemctl daemon-reload
```

## Endpoints

### `/farms`

- Bring all nodes information and public ips

    Example

    ```json
    // 20210815123713
    // http://localhost:8080/farms

    {
      "data": {
        "farms": [
          {
            "name": "devnet",
            "farmId": 1,
            "twinId": 3,
            "version": 1,
            "cityId": 0,
            "countryId": 0,
            "pricingPolicyId": 1
          }
        ],
        "publicIps": [
          {
            "id": "kGoHpiNM1_R",
            "ip": "185.206.122.40/24",
            "farmId": "OubK0WQyJT",
            "contractId": 0,
            "gateway": "185.206.122.1"
          }
        ]
      }
    }
    ```

### `/nodes`

- Bring all nodes information and public configurations

    Example

    ```json
    // 20210815123555
    // http://localhost:8080/nodes

    {
      "data": {
        "nodes": [
          {
            "version": 1,
            "id": "LWeENXIU2a",
            "nodeId": 1,
            "farmId": 1,
            "twinId": 5,
            "countryId": 0,
            "gridVersion": 1,
            "cityId": 0,
            "uptime": 0,
            "created": 1628862798,
            "farmingPolicyId": 2,
            "cru": "24",
            "mru": "202875785216",
            "sru": "512110190592",
            "hru": "9001778946048"
          },
        ],
        "publicConfigs": [
          {
            "gw4": "185.206.122.1",
            "ipv4": "185.206.122.31/24",
            "ipv6": "2a10:b600:1:0:fc38:90ff:feb4:b15d/ 64",
            "gw6": "fe80::2e0:ecff:fe7b:7a67"
          }
        ]
      }
    }
    ```

- Query params

  - farm_id:

    select nodes from specific farm using farm id, example: `?farm_id=1`
  
  - page:

    default view is for 50 nodes and paginated to make it faster and easier to parse, example: `?page=1`
  
  - max_result:

    default max result for page, example: ``?max_result=50`

- Example full query

```json
// 20210824113426
// http://localhost:8080/nodes?farm_id=1&page=1

  {
    "data": {
      "nodes": [
        {
          "version": 1,
          "id": "-ldFCBmX8Y_",
          "nodeId": 2,
          "farmId": 1,
          "twinId": 2,
          "country": "BE",
          "gridVersion": 1,
          "city": "Unknown",
          "uptime": 0,
          "created": 1629466038,
          "farmingPolicyId": 1,
          "cru": "24",
          "mru": "202875785216",
          "sru": "512110190592",
          "hru": "9001778946048",
          "publicConfig": {
            "gw4": "185.206.122.1",
            "ipv4": "185.206.122.31/24",
            "ipv6": "2a10:b600:1:0:fc38:90ff:feb4:b15d/64",
            "gw6": "fe80::2e0:ecff:fe7b:7a67"
          }
        .
        .
        .
  ```

### `/gateways`

- Bring all gateways information and public configurations and domains

    Example

    ```json
    // 20211012115620
    // http://localhost:8080/gateways?max_result=1

      {
        "data": {
          "nodes": [
            {
              "version": 19,
              "id": "a5UM_zebCN-",
              "nodeId": 1,
              "farmId": 1,
              "twinId": 4,
              "country": "BE",
              "gridVersion": 1,
              "city": "Unknown",
              "uptime": 594154,
              "created": 1633094058,
              "farmingPolicyId": 1,
              "updatedAt": "2021-10-07T14:28:36.715Z",
              "cru": "24",
              "mru": "202802958336",
              "sru": "512110190592",
              "hru": "9001778946048",
              "publicConfig": {
                "domain": "gent01.devnet.grid.tf",
                "gw4": "185.206.122.1",
                "gw6": "<nil>",
                "ipv4": "185.206.122.31/24",
                "ipv6": "<nil>"
              }
            }
          ]
        }
      }
    ```

- Query params

  - farm_id:

    select nodes from specific farm using farm id, example: `?farm_id=1`
  
  - page:

    default view is for 50 nodes and paginated to make it faster and easier to parse, example: `?page=1`
  
  - max_result:

    default max result for page, example: ``?max_result=50`

### `/nodes/<node-id>` or `/gateways/<node-id>`

- Bring the node active used and total resources

    Example

    ```json
    // 20210815123807
    // http://localhost:8080/nodes/1

    {
      "total": {
        "cru": 24,
        "sru": 512110190592,
        "hru": 9001778946048,
        "mru": 201863462912,
        "ipv4u": 0
      },
      "used": {
        "cru": 2,
        "sru": 126701535232,
        "hru": 0,
        "mru": 25548913455,
        "ipv4u": 0
      }
    }
    ```

### `/nodes/<node-id>/status` or `/gateways/<node-id>/status`

- Bring the node status up or down

    Example

    ```json
    // 20211130123101
    // http://localhost:8080/nodes/7/status

    {
      "status": "up"
    }
    ```

### Dockerfile

- get public and private key for a yggdrasil configuration

To build & run dockerfile

```bash
docker build -t threefoldtech/gridproxy .
docker run --name gridproxy -e TWIN=296 -e EXPLORER="https://graphql.dev.grid.tf/graphql" -e SUBSTRATE="wss://tfchain.dev.grid.tf/ws" -e PUBLIC_KEY="5011157c2451b238c99247b9f0793f66e5b77998272c00676d23767fe3d576d8" -e PRIVATE_KEY="ff5b3012dbec23e86e2fde7dcd3c951781e87fe505be225488b50a6bb27662f75011157c2451b238c99247b9f0793f66e5b77998272c00676d23767fe3d576d8" --cap-add=NET_ADMIN threefoldtech/gridproxy
```

### Update helm package

- Do `helm lint charts/gridproxy`
- Regenerate the packages `helm package -u charts/gridproxy`
- Regenerate index.yaml `helm repo index --url https://mattiaperi.github.io/helm-chart/ .`
- Push your changes

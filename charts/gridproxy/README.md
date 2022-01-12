# How to use grid proxy helm chart

- Install helm

- `Helm repo update`

- Install the chart

  **Note**: EXPLORER_URL, SERVER_IP and REDIS_URL has default values you may not pass them if you want to use the defaults

  ```bash
  helm install -f values.yaml gridproxy . --set ingress.host="gridproxy.3botmain.grid.tf" --set env.TWIN=60 --set env.SERVER_PORT=":8080" --set env.EXPLORER="https://graphql.dev.grid.tf/graphql" --set env.SUBSTRATE="wss://tfchain.dev.grid.tf/ws" --set env.REDIS="localhost:6379"
  ```

- Update [polkadot](https://polkadot.js.org/apps/?rpc=wss%3A%2F%2Ftfchain.dev.grid.tf%2Fws#/extrinsics) with yggdrasail ip from `kubectl logs <pod name>`
  
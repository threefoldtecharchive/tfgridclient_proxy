# How to use grid proxy helm chart

- Install helm

- `Helm repo update`

- Install the chart

  **Note**: EXPLORER_URL, SERVER_IP and REDIS_URL has default values you may not pass them if you want to use the defaults

  ```bash
  helm install -f values.yaml gridproxy . --set ingress.host="ellolproxy.webg1dev.grid.tf" --set env.TWIN=7 --set env.SERVER_IP="0.0.0.0:8080" --set env.EXPLORER_URL="https://explorer.devnet.grid.tf/graphql/" --set env.REDIS_URL="localhost:6379"
  ```

- Update [polkadot](https://polkadot.js.org/apps/?rpc=wss%3A%2F%2Fexplorer.devnet.grid.tf%2Fws#/extrinsics) with yggdrasail ip from `kubectl logs <pod name>`
  
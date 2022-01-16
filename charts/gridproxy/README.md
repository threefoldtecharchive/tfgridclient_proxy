# How to use grid proxy helm chart

- Install helm

- `Helm repo update`

- get public and private key for a yggdrasil configuration

- Install the chart

  **Note**: EXPLORER_URL, SERVER_IP and REDIS_URL has default values you may not pass them if you want to use the defaults

  ```bash
  helm install -f values.yaml gridproxy . --set ingress.host="gridproxy.3botmain.grid.tf" --set env.TWIN=296 --set env.EXPLORER="https://graphql.dev.grid.tf/graphql" --set env.SUBSTRATE="wss://tfchain.dev.grid.tf/ws" --set env.PUBLIC_KEY="5011157c2451b238c99247b9f0793f66e5b77998272c00676d23767fe3d576d8" --set env.PRIVATE_KEY="ff5b3012dbec23e86e2fde7dcd3c951781e87fe505be225488b50a6bb27662f75011157c2451b238c99247b9f0793f66e5b77998272c00676d23767fe3d576d8"
  ```
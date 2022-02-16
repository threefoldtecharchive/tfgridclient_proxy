# How to use grid proxy helm chart

- Install helm

- `Helm repo update`

- get public and private key for a yggdrasil configuration, **note**: each setup should has it's own public key and private key and not to be shared with anyone

  `yggdrasil -genconf -json > yggdrasil.conf`

  see [here](https://yggdrasil-network.github.io/configuration.html) and example file `ygg_sample.conf`

- Remove traefik controller & service and Install nginx controller and cert manager (if not there)

  for nginx:

    ```bash
    helm upgrade --install ingress-nginx ingress-nginx \
      --repo https://kubernetes.github.io/ingress-nginx \
      --namespace ingress-nginx --create-namespace
    ```

  for cert manager:

    ```bash
    helm repo add jetstack https://charts.jetstack.io
    helm repo update
    helm install cert-manager jetstack/cert-manager --namespace cert-manager --create-namespace --set installCRDs=true
    ```

- Apply certificate `kubectl create -f prod_issuer.yaml`

- If you want to add more peers add using `--set env.PEERS="  tls:\\\/\\\/62.210.85.80:39575\\\n   tls:\\\/\\\/54.37.137.221:11129\\\n"` add `\\\` as escape characters before each `/`

- Install the chart

  **Note**: EXPLORER_URL, SERVER_IP and REDIS_URL has default values you may not pass them if you want to use the defaults

  ```bash
  helm install -f values.yaml gridproxy . --set ingress.host="gridproxy.3botmain.grid.tf" --set env.MNEMONICS="" --set env.SUBSTRATE="wss://tfchain.dev.grid.tf/ws" --set env.PUBLIC_KEY="5011157c2451b238c99247b9f0793f66e5b77998272c00676d23767fe3d576d8" --set env.PRIVATE_KEY="ff5b3012dbec23e86e2fde7dcd3c951781e87fe505be225488b50a6bb27662f75011157c2451b238c99247b9f0793f66e5b77998272c00676d23767fe3d576d8" --set env.PEERS="  tls:\\\/\\\/62.210.85.80:39575\\\n   tls:\\\/\\\/54.37.137.221:11129\\\n" --set env.POSTGRES_HOST="127.0.0.1" --set env.POSTGRES_PORT="5432" --set env.POSTGRES_DB="db" --set env.POSTGRES_USER="postgres" --set env.POSTGRES_PASSWORD="password"
  ```

- PUBLIC_KEY: yggdrasil public key
- PRIVATE_KEY: yggdrasil private key
- PEERS: yggdrasil peers, get from [here](https://publicpeers.neilalexander.dev/)
- SUBSTRATE: substrate url

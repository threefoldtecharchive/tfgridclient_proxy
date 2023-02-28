# How to use grid proxy helm chart

- Install helm

- `Helm repo update`

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

- Install the chart

  **Note**: these envs has default values you may not pass them if you want to use the defaults
    - SERVER_PORT: Addr optionally specifies the TCP address for the server to listen on, in the form "host:port". default `:80`
    - POSTGRES_PORT: default `5432`
    - POSTGRES_USER: default `postgres`

  ```bash
  helm install -f values.yaml gridproxy . --set ingress.host="gridproxy.dev.grid.tf" --set env.POSTGRES_HOST="127.0.0.1" --set env.POSTGRES_PORT="5432" --set env.POSTGRES_DB="db" --set env.POSTGRES_USER="postgres" --set env.POSTGRES_PASSWORD="password"
  ```

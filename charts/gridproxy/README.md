# How to use grid proxy helm chart

- Install helm

- `Helm repo update`

` Install the chart

  ```bash
  helm install -f values.yaml gridproxy . --set ingress.host="example.yourdomain.com"
  ```
  
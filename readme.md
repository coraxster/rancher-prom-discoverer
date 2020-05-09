rancher-prom-discoverer
=======

Auto discovery prometheus metrics for Rancher.

## Use
Rancher deployment `docker.compose.yml`
```yaml
version: '2'
services:
  your-super-service:
    image: $IMAGE
    labels:
      prometheus.endpoint: /metrics # can be any
      prometheus.label.foo: bar # extra label: foo -> bar
    ports:
      - 3000 # /metrics port
```
You'll got
```
service_metric{hostname="some-host-01", stack="your-super-stack", service="your-super-service", foo="bar" ...}
```

## Deploy
Used as prometheus sidekick [here](https://git.exness.io/xdata/prometheus).
 
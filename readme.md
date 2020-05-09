rancher-prom-discoverer
=======

Auto discovery prometheus metrics for Rancher 1.6.

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
rancher `docker-compose.yml`
```yaml
version: '2'
services:
  prometheus:
    image: prom/prometheus
    volumes:
      - /etc/prometheus/auto:/etc/prometheus/auto
    labels:
      io.rancher.sidekicks: prom-discoverer
    command: ["--config.file=/etc/prometheus/auto/prometheus.yml"]

  prom-discoverer:
    image: registry.exness.io/xdata/rancher-prom-discoverer:0-0-4
    environment:
      RANCHER_TOKEN: {YOUR-API-TOKEN}
      RANCHER_PROJECT: {YOUR-PROJECT-NAME}
      RANCHER_HOST: {YOUR-RANCHER_HOST}
      FILE: /etc/prometheus/auto/targets.json
      SENTRY_DSN: {YOUR-SENTRY_DSN} # if you need it
    volumes:
      - /etc/prometheus/auto:/etc/prometheus/auto
```
prometheus.yml
```yaml
...
scrape_configs:
  - job_name: 'auto'  # This is a default value, it is mandatory.
    file_sd_configs:
      - files:
        - /etc/prometheus/auto/targets.json

```
 
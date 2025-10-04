# Prometheus Alert Manager Route Generator


## Building

```shell
go mod tidy
go build .
```

## Running

```shell
./amgenerator --service-data ./service-data.yaml --routes-directory ./am_routes --alert-routing-file am_routes.yaml -alert-receivers-file am_receivers.yaml --orphanAlertEmail orphan-alerts@example.com
```

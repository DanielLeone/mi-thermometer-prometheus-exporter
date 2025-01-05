# mi-thermometer-prometheus-exporter
Prometheus exporter for the Xiaomi Temperature &amp; Humidity monitors running ATC firmware

# build and run on linux (dbus)
```
docker compose up --build
```

# test go binary builds
```
goreleaser build --snapshot --clean
```

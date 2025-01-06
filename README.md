# mi-thermometer-prometheus-exporter
Prometheus exporter for the Xiaomi Temperature and Humidity Monitor running ATC firmware

# build and run on linux (dbus)
```
docker compose up --build
```

# test go binary builds
```
goreleaser build --snapshot --clean
```

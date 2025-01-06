docker run --rm --privileged \
  -v "$(pwd):/go/src/github.com/danielleone/mi-thermometer-prometheus-exporter" \
  -v '/var/run/docker.sock:/var/run/docker.sock' \
  -w '/go/src/github.com/danielleone/mi-thermometer-prometheus-exporter' \
  --env-file '.env' \
  'goreleaser/goreleaser' build --snapshot --clean

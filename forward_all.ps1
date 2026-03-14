$services = @(
    "kafka:9092:9092",
    "postgres:5432:5432",
    "clickhouse:8123:8123",
    "clickhouse:9000:9000",
    "gateway:8080:8080",
    "prometheus-service:9090:9090",
    "grafana-service:3000:3000"
)

foreach ($svc in $services) {
    $parts = $svc.Split(":")
    $name = $parts[0]
    $local = $parts[1]
    $remote = $parts[2]

    Start-Process powershell -ArgumentList "kubectl port-forward svc/$name ${local}:${remote}" -WindowStyle Normal
}


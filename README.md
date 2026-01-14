# Pinger

A concurrent service health checker written in Go. Uses goroutines to perform parallel health checks against HTTP endpoints, TCP ports, and DNS records.

Built as a learning project to explore Go concurrency patterns (worker pools, channels, WaitGroups, context cancellation).

## Installation

```bash
go build -o pinger .
```

## Usage

### Quick Check

```bash
# Check a single URL
./pinger check https://google.com

# Check multiple URLs concurrently
./pinger check https://google.com https://github.com https://example.com

# Adjust concurrency and timeout
./pinger check -n 20 -t 10 https://example.com

# Verbose output (shows worker count, target count, and full errors)
./pinger check -v https://google.com https://github.com
```

### Using a Config File

```bash
./pinger check -c configs/example.yaml
```

### JSON Output (for CI/CD pipelines)

```bash
./pinger check -c configs/example.yaml --json
```

Exit code is `1` if any targets are unhealthy — useful for shell scripts and pipelines.

## Configuration

```yaml
# configs/example.yaml
concurrency: 10  # Number of worker goroutines
timeout: 5s      # Default timeout for all checks

targets:
  # HTTP/HTTPS endpoints
  - name: "GitHub"
    url: "https://github.com"
    type: http
    expected_status: 200
    headers:
      Authorization: "Bearer token"

  # TCP port checks (databases, caches, etc.)
  - name: "Redis"
    url: "localhost:6379"
    type: tcp
    timeout: 2s

  # DNS resolution checks
  - name: "Google DNS"
    url: "google.com"
    type: dns
```

## Key Concepts (for learning)

### Worker Pool (`internal/worker/pool.go`)

This is the core concurrency piece. Key patterns used:

- **Buffered channels** for job queue and results collection
- **`sync.WaitGroup`** to wait for all workers to finish
- **`context.Context`** for timeouts and cancellation
- **Fan-out/fan-in** — distribute jobs to workers, collect results

```go
// Simplified flow:
jobs := make(chan Job)       // Fan-out: send jobs to workers
results := make(chan Result) // Fan-in: collect results

for i := 0; i < numWorkers; i++ {
    go worker(jobs, results) // Each worker is a goroutine
}
```

### Interface Pattern (`internal/checker/`)

The `Checker` interface allows different check types (HTTP, TCP, DNS) to be used interchangeably:

```go
type Checker interface {
    Check(ctx context.Context, target Target) Result
}
```

Adding a new check type (gRPC, Redis, etc.) only requires implementing this interface.

---

## TODO

### CLI Enhancements
- [x] Add `--verbose` flag for detailed output
- [ ] Add `--silent` flag (only output on failure)
- [x] Add `pinger version` command
- [ ] Add retry logic with configurable attempts (`--retries 3`)
- [ ] Add `--fail-fast` to exit on first failure

### Watch Mode
- [ ] Add `pinger watch -c config.yaml` for continuous monitoring
- [ ] Use `time.Ticker` for interval-based re-checks
- [ ] Add `--interval` flag (default 30s)

### New Check Types
- [ ] gRPC health check (implement `grpc.health.v1.Health`)
- [ ] Redis `PING` command
- [ ] PostgreSQL/MySQL connection check
- [ ] Custom script/command execution

### Output & Observability
- [ ] Add Prometheus metrics endpoint (`pinger serve --metrics`)
- [ ] OpenTelemetry tracing support
- [ ] Webhook notifications on failure (Slack, PagerDuty)
- [ ] Response body matching (`expected_body: "OK"`)

### TUI Mode
- [ ] Add interactive TUI using [Bubbletea](https://github.com/charmbracelet/bubbletea)
- [ ] Live-updating status table
- [ ] Sparkline graphs for latency history
- [ ] Keyboard shortcuts to retry individual targets

### Kubernetes Integration
- [ ] **Helm chart** for deploying pinger as a CronJob or Deployment
- [ ] **ServiceMonitor** CRD for Prometheus Operator scraping
- [ ] Run as **init container** to wait for dependencies before app starts
- [ ] **ConfigMap** support for loading targets from K8s config
- [ ] Check **K8s Services** by DNS name (e.g., `my-service.namespace.svc.cluster.local`)
- [ ] Autodiscover endpoints from **K8s Ingress** or **Service** resources
- [ ] **Pod readiness gate** integration
- [ ] Output format compatible with `kubectl wait --for=condition`

### Code Quality
- [ ] Add unit tests for checkers (mock HTTP responses)
- [ ] Add integration tests
- [ ] Add benchmarks for worker pool
- [ ] Run with `-race` flag to catch concurrency bugs
- [ ] Add golangci-lint configuration

---

## Development

```bash
# Build
go build -o pinger .

# Build with race detection (catches concurrency bugs)
go build -race -o pinger .

# Run tests
go test ./...

# Tidy dependencies
go mod tidy
```

## License

MIT

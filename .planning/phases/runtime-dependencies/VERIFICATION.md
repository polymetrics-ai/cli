# Verification: runtime-dependencies

## Commands Run

```bash
go test ./...
go vet ./...
go build ./cmd/poly
make verify
./poly perf compare --iterations 50 --json
./poly perf compare --iterations 50 --runtime --json
scripts/runtime.sh up
scripts/runtime.sh down
```

## Results

- `go test ./...`: passed
- `go vet ./...`: passed
- `go build ./cmd/poly`: passed
- `make verify`: passed
- dependency-free performance: passed
- runtime-backed performance: degraded because services were unavailable

## Dependency-Free Performance

```text
iterations: 50
records: 150
duration: 38.28725ms
average: 765.745us
records/sec: 3917.75
```

## Runtime-Backed Status

Runtime startup was attempted twice with `scripts/runtime.sh up`.

Both attempts failed during image pull with:

```text
lookup production.cloudfront.docker.com: no such host
```

After the failed pull, `scripts/runtime.sh down` was run to remove partial compose state.

`./poly perf compare --iterations 50 --runtime --json` correctly reported degraded runtime health.


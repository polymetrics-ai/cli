# Measurement record — CLI package test-ceiling foundation

## Test inventory

The following commands were run before and after the fixture change:

```sh
go test -list '.*' ./internal/cli
awk '/^Test/ {print}' <output> | sort
```

Both captures contained 263 runnable test names. Their SHA-256 hashes are identical:

```text
198d206817b81ff1f9480421f2aad978958f275bce83c1e105eb28023a4e61b6
```

`TestMain` is intentionally not listed by `go test -list`; it is the package's 264th declared `Test…` function. There are no name additions, removals, or renames.

## Full package timing

| Run | Command | Package duration | Wall duration | Result |
| --- | --- | ---: | ---: | --- |
| Before | `/usr/bin/time -p go test -v -timeout 30m ./internal/cli` | 623.128s | 627.73s | pass |
| After | `/usr/bin/time -p go test -v -timeout 30m ./internal/cli` | 532.694s | 537.29s | pass |
| Aggregate verification | `make verify` → `go test -timeout 20m ./...` | 685.504s for `internal/cli` | n/a | pass |

The local like-for-like package reduction is 90.434s (14.5%). The documented integration-base CI measurement is 1180.982s; applying the measured relative reduction yields approximately 1010s, or 15.9% under the unchanged 1200s per-binary ceiling. This is an inference; the hosted Verify result is the authoritative CI measurement.

## Root-cause measurement

- `rg -n 'buildTransportPM\\(' internal/cli --glob '*_test.go'` found 18 callers.
- The pre-change helper built identical `./cmd/pm` source into each caller's `t.TempDir`.
- `rg -n 't\\.Parallel\\(' internal/cli --glob '*_test.go'` found no calls.
- Process-wide `t.Setenv` and `t.Chdir` calls prevent safely applying package-wide test parallelism without a broad isolation rewrite.

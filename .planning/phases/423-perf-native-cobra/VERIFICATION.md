# Phase 423 Verification

## Required gate checklist

- [x] `gofmt -w cmd internal`
- [x] `go test ./internal/cli/... -run 'Perf|CobraRouterShell|Golden' -count=1`
- [x] `go test ./internal/cli/ -run Certify -count=1`
- [x] `go vet ./...`
- [x] `go test ./...`
- [x] `go build ./cmd/pm`
- [x] `make verify`
- [x] `git diff --check origin/feat/cli-architecture-v2...HEAD`
- [x] `git diff -- go.mod go.sum`

## CLI parity checklist

- [x] Golden transcript diff empty; no fixture changes.
- [x] `./pm help perf` checked: exit 0, docs-map canonical help.
- [x] Bare `./pm perf` checked: exit 0, byte-identical to `pm help perf`.
- [x] `./pm perf --help` checked: exit 0, byte-identical to docs-map help.
- [x] JSON manual checked: `./pm perf --json` exit 0 with `CommandManual` envelope.
- [x] Invalid action checked: `./pm perf bogus --json` exit 2, JSON category `usage`.
- [x] Native flag semantics checked: `compare --iterations`, `compare --runtime`, `sync-modes --records`; space/equals forms, repeated scalar last-wins, bare bool/value sentinels, unknown flags, extra args, and late `--root`/`--json`.
- [x] Runtime compare config use checked with loopback endpoints only in focused tests; no services started.
- [x] Completion metadata preserved; Phase 15 completion implementation explicitly not included.
- [x] `docs/cli/perf.md` parity checked by docs-generate-diff/golden docs test; no update needed because help unchanged.
- [x] Website docs/source/generated data checked under `website/**`; no update needed because generated docs unchanged.
- [x] Generated help/manual artifacts checked via existing generator/docs validation.

## Optional / safety-limited

- [x] Runtime-backed integration tests not run; no services started.
- [x] No credentialed connector checks.
- [x] No external services started.
- [x] No reverse ETL execution beyond repository local temp-dir smoke inside `make verify`.
- [x] No new dependencies.

## Results

```bash
go test ./internal/cli/ -run 'Perf|CobraRouterShell' -count=1
```

Result: red as expected before implementation. Failed because `perf` was still a legacy wrapper and native perf subcommands/flags were missing.

```bash
gofmt -w internal/cli/cobra_router.go internal/cli/cli.go internal/cli/cobra_router_test.go internal/cli/perf_cli_test.go
go test ./internal/cli/ -run 'Perf|CobraRouterShell' -count=1
```

Result: pass (`ok  	polymetrics.ai/internal/cli	13.101s`).

```bash
gofmt -w cmd internal
go test ./internal/cli/... -run 'Perf|CobraRouterShell|Golden' -count=1
go vet ./...
go build ./cmd/pm
```

Result: focused/golden test pass (`ok  	polymetrics.ai/internal/cli	18.543s`); `go vet` and `go build` exited 0 with no output.

```bash
go test ./internal/cli/ -run Certify -count=1
```

Result: pass (`ok  	polymetrics.ai/internal/cli	91.433s`).

```bash
go test ./...
```

Result: pass. Full package output emitted in terminal run; slow packages included `ok  	polymetrics.ai/internal/cli	179.296s` and `ok  	polymetrics.ai/internal/connectors/certify	367.560s`.

```bash
make verify
```

Result: pass. Completed gofmt, tidy-check, vet, `go test -timeout 20m ./...`, `go build ./cmd/pm`, docs validate, local smoke flow, golangci-lint, and `connectorgen validate`; terminal tail included `connectorgen validate: 547 connector(s) checked, 0 findings`.

```bash
go build ./cmd/pm
```

Result: pass; exited 0 with no output.

```bash
./pm help perf
./pm perf
./pm perf --help
./pm perf --json
./pm perf bogus --json
```

Result: pass. Help/bare/`--help` byte-identical (874 bytes); JSON manual emitted `CommandManual` (1004 bytes); invalid action exited 2 with usage JSON and stderr `error: unknown command "bogus" for "pm perf"`.

```bash
./pm perf compare --iterations 1 --runtime=false --json
./pm perf sync-modes --records 5 --json
```

Result: pass. Compare emitted `PerformanceComparison` with dependency-free iterations=1, records=3, no runtime-backed result; sync-modes emitted `SyncModeBenchmark` with records=5 for every result.

```bash
./pm docs generate --dir "$TMP_DOCS/cli" --connectors-dir "$TMP_DOCS/connectors"
diff -ru docs/cli "$TMP_DOCS/cli"
./pm docs validate --connectors-dir docs/connectors
npm --prefix website run gen:docs
```

Result: pass. Docs diff had no output; docs validate passed; website generator wrote 11 docs pages with no tracked diff.

```bash
git diff --check origin/feat/cli-architecture-v2...HEAD
git diff -- go.mod go.sum
```

Result: pass / no output.

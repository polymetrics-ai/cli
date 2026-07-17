# Phase 423 Verification

## Required gate checklist

- [x] `gofmt -w cmd internal`
- [x] `go test ./internal/cli/... -run 'Perf|CobraRouterShell|Golden' -count=1`
- [x] `go test ./internal/cli/ -run Certify -count=1`
- [x] `go vet ./...`
- [ ] `go test ./...`
- [x] `go build ./cmd/pm`
- [ ] `make verify`
- [ ] `git diff --check origin/feat/cli-architecture-v2...HEAD`
- [ ] `git diff -- go.mod go.sum`

## CLI parity checklist

- [ ] Golden transcript diff empty or reviewed intentionally.
- [ ] `./pm help perf` checked: exit 0, docs-map canonical help.
- [ ] Bare `./pm perf` checked: exit 0, byte-identical to `pm help perf`.
- [ ] `./pm perf --help` checked: exit 0, byte-identical to docs-map help.
- [ ] JSON manual checked: `./pm perf --json` exit 0 with `CommandManual` envelope.
- [ ] Invalid action checked: `./pm perf bogus --json` exit 2, JSON category `usage`.
- [ ] Native flag semantics checked: `compare --iterations`, `compare --runtime`, `sync-modes --records`; space/equals forms, repeated scalar last-wins, bare bool/value sentinels, unknown flags, extra args, and late `--root`/`--json`.
- [ ] Runtime compare config use checked with loopback endpoints only; no services started.
- [ ] Completion metadata preserved; Phase 15 completion implementation explicitly not included.
- [ ] `docs/cli/perf.md` parity checked by docs-generate-diff/golden docs test; update only if help changes.
- [ ] Website docs/source/generated data checked under `website/**`; update only if generated docs change.
- [ ] Generated help/manual artifacts checked via existing generator/docs validation.

## Optional / safety-limited

- [ ] Runtime-backed integration tests not run unless explicitly requested; no services started.
- [ ] No credentialed connector checks.
- [ ] No external services started.
- [ ] No reverse ETL execution beyond repository local temp-dir smoke inside `make verify`.
- [ ] No new dependencies.

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

Full gate, runtime help/docs/website parity, and diff guards pending.

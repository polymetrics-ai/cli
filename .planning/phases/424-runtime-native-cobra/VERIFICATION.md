# Phase 424 Verification

## Required gate checklist

- [x] `gofmt -w cmd internal`
- [x] `go test ./internal/cli/... -run 'Runtime|CobraRouterShell|Golden' -count=1`
- [x] `go test ./internal/cli/...`
- [x] `go vet ./...`
- [x] `go test ./...`
- [x] `go build ./cmd/pm`
- [x] `make verify`
- [x] `git diff --check origin/feat/cli-architecture-v2...HEAD`
- [x] `git diff -- go.mod go.sum`

## CLI parity checklist

- [x] Golden transcript diff empty; no fixture changes.
- [x] `./pm help runtime` checked: exit 0, docs-map canonical help.
- [x] Bare `./pm runtime` checked: exit 0, same canonical help as `pm help runtime`.
- [x] `./pm runtime --help` checked: exit 0, docs-map canonical help.
- [x] JSON manual checked: `./pm runtime --json` exit 0 with `CommandManual` envelope.
- [x] Invalid action checked: `./pm runtime bogus --json` exit 2, JSON category `usage`.
- [x] Native doctor semantics checked: `doctor --json`, unknown flags ignored, extra args ignored, late `--json`, late `--root`, and config-file endpoints.
- [x] Runtime service optionality checked: tests use loopback/config-only endpoints; no Podman/PostgreSQL/DragonflyDB/Temporal startup.
- [x] Completion metadata/no-file fallback seam preserved; Phase 15 completion implementation explicitly not included.
- [x] `docs/cli/runtime.md` parity checked by docs-generate-diff; no update needed because help unchanged.
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
gofmt -w internal/cli/cobra_router.go internal/cli/cli.go internal/cli/cobra_router_test.go internal/cli/runtime_cli_test.go
go test ./internal/cli/ -run 'Runtime|CobraRouterShell' -count=1
```

Result: pass (`ok  \tpolymetrics.ai/internal/cli\t11.749s`).

```bash
go test ./internal/cli/... -run 'Runtime|CobraRouterShell|Golden' -count=1
go test ./internal/cli/...
go vet ./...
go build ./cmd/pm
```

Result: pass. Focused/golden passed (`ok  \tpolymetrics.ai/internal/cli\t17.329s`); full internal CLI package passed (`ok  \tpolymetrics.ai/internal/cli\t195.015s`); `go vet` and `go build` emitted no output.

```bash
gofmt -w cmd internal
go vet ./...
go test ./...
go build ./cmd/pm
make verify
```

Result: pass. `go test ./...` passed; slow packages included `internal/connectors/certify 342.381s`. `make verify` passed and ended with `connectorgen validate: 547 connector(s) checked, 0 findings`; `gofmt`, `go vet`, and `go build` emitted no output.

```bash
./pm help runtime
./pm runtime
./pm runtime --help
./pm runtime --json
./pm runtime bogus --json
./pm runtime doctor --unknown ignored extra --root "$root" --json
```

Result: pass. Help/bare/`--help` byte-identical (`help bytes=470`); JSON manual emitted `CommandManual` for `runtime` (`600` bytes); invalid action exited 2 with usage JSON and stderr `error: unknown command "bogus" for "pm runtime"`; loopback doctor emitted `RuntimeDoctor`, redacted PostgreSQL endpoint, Dragonfly/Temporal topology endpoints, all statuses `error`, stderr_bytes=0, and secret_leak=0.

```bash
./pm docs generate --dir "$TMP_DOCS/cli" --connectors-dir "$TMP_DOCS/connectors"
diff -ru docs/cli "$TMP_DOCS/cli"
./pm docs validate --connectors-dir docs/connectors
npm --prefix website run gen:docs
git diff --check origin/feat/cli-architecture-v2...HEAD
git diff -- go.mod go.sum
```

Result: pass. Docs generated to temp dir; docs diff emitted no output; docs validate passed; website docs generator wrote 11 docs pages; diff-check and go.mod/go.sum diff emitted no output.

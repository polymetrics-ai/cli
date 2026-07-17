# Phase 422 Verification

## Required gate checklist

- [x] `gofmt -w cmd internal`
- [x] `go test ./internal/cli/... -run 'Query|CobraRouterShell|Golden' -count=1`
- [x] `go test ./internal/cli/ -run Certify -count=1`
- [x] `go vet ./...`
- [x] `go test ./...`
- [x] `go build ./cmd/pm`
- [x] `make verify`
- [x] `git diff --check origin/feat/cli-architecture-v2...HEAD`
- [x] `git diff -- go.mod go.sum`

## CLI parity checklist

- [x] Golden transcript diff empty; no fixture changes.
- [x] `./pm help query` checked: exit 0, docs-map canonical help.
- [x] Bare `./pm query` checked: exit 0, byte-identical to `pm help query`.
- [x] `./pm query --help` checked: exit 0, byte-identical to docs-map help.
- [x] JSON manual checked: `./pm query --json` exit 0 with `CommandManual` envelope.
- [x] Invalid action checked: `./pm query bogus --json` exit 2, JSON category `usage`, no project open.
- [x] Query JSON output checked with local temp fixture only.
- [x] Read-only SQL rejection checked; no generic SQL write exposed or validation weakened.
- [x] Native flag semantics checked: `--table`, `--sql`, `--limit`, `--fields`, `--agent-mode`, `--sample`; space/equals forms, repeated scalar last-wins, repeated/comma `--fields`, bare bool sentinels, unknown flags, extra args, and late `--root`/`--json`.
- [x] Completion metadata preserved; Phase 15 completion implementation explicitly not included.
- [x] `docs/cli/query.md` parity checked by docs-generate-diff/golden docs test; no update needed because help unchanged.
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
go test ./internal/cli/... -run 'Query|CobraRouterShell|Golden' -count=1
```

Result: pass (`ok  	polymetrics.ai/internal/cli	10.691s`).

```bash
go test ./internal/cli/ -run Certify -count=1
```

Result: pass (`ok  	polymetrics.ai/internal/cli	91.638s`).

```bash
gofmt -w cmd internal
go vet ./...
go build ./cmd/pm
```

Result: pass; `go vet` and `go build` exited 0 with no output.

```bash
go test ./...
```

Result: pass. Full package output emitted in terminal run; slow packages included `ok  	polymetrics.ai/internal/cli	205.284s` and `ok  	polymetrics.ai/internal/connectors/certify	380.015s`.

```bash
make verify
```

Result: pass. Completed gofmt, tidy-check, vet, `go test -timeout 20m ./...`, `go build ./cmd/pm`, docs validate, local smoke flow, golangci-lint, and `connectorgen validate`; terminal tail included `connectorgen validate: 547 connector(s) checked, 0 findings`.

```bash
./pm help query
./pm query
./pm query --help
./pm query --json
./pm query bogus --json
```

Result: pass. Help/bare/`--help` byte-identical (1189 bytes); JSON manual emitted `CommandManual` (1368 bytes); invalid action exited 2 with usage JSON and stderr `error: unknown command "bogus" for "pm query"`.

```bash
./pm query run --table customers --fields id,email --limit 1 --root "$QUERY_ROOT" --json
./pm query run --sql "DELETE FROM customers" --root "$QUERY_ROOT" --json
```

Result: pass using local temp fixture. Query run emitted `QueryResult`; read-only SQL rejection exited 1 with `Error` envelope.

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

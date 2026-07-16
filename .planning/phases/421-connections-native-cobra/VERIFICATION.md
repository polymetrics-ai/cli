# Phase 421 Verification

## Required gate checklist

- [x] `gofmt -w cmd internal`
- [x] `go test ./internal/cli/... -run 'Connections|CobraRouterShell|Golden' -count=1`
- [x] `go test ./internal/cli/ -run Certify -count=1`
- [x] `go vet ./...`
- [x] `go test ./...`
- [x] `go build ./cmd/pm`
- [x] `make verify`
- [x] `git diff --check origin/feat/cli-architecture-v2...HEAD`
- [x] `git diff -- go.mod go.sum`

## CLI parity checklist

- [x] Golden transcript diff empty; no fixture changes.
- [x] `./pm help connections` checked: exit 0, docs-map canonical help.
- [x] Bare `./pm connections` checked: exit 0, byte-identical to `pm help connections`.
- [x] `./pm connections --help` checked: exit 0, byte-identical to docs-map help.
- [x] JSON manual checked: `./pm connections --json` exit 0 with `CommandManual` envelope.
- [x] Invalid action checked: `./pm connections bogus --json` exit 2, JSON category `usage`.
- [x] Native flag semantics checked: `--source`, `--destination`, `--stream`, `--sync-mode`, `--cursor`, `--table`, `--source-config`, `--destination-config`, `--primary-key`; space/equals forms, repeated singleton last-wins, repeated primary key accumulation, bare bool values, unknown flags, extra args, and late `--root`/`--json`.
- [x] Completion compatibility seam preserved; Phase 15 implementation explicitly not included.
- [x] `docs/cli/connections.md` parity checked by docs-generate-diff/golden docs test; no update needed because help unchanged.
- [x] Website docs/source/generated data checked under `website/**`; no update needed because generated docs unchanged.
- [x] Generated help/manual artifacts checked via existing generator/docs validation.

## Optional / safety-limited

- [x] Runtime-backed integration tests not run; no services started.
- [x] No credentialed connector checks.
- [x] No reverse ETL external execution beyond repository local temp-dir smoke in `make verify`.
- [x] No new dependencies.

## Results

```bash
go test ./internal/cli/... -run 'Connections|CobraRouterShell|Golden' -count=1
```

Result: pass (`ok  	polymetrics.ai/internal/cli	12.034s`).

```bash
go test ./internal/cli/ -run Certify -count=1
```

Result: pass (`ok  	polymetrics.ai/internal/cli	95.754s`).

```bash
gofmt -w cmd internal
go vet ./...
go test ./...
go build ./cmd/pm
make verify
```

Result: pass. Full `go test ./...` included `ok  	polymetrics.ai/internal/cli	199.324s` and `ok  	polymetrics.ai/internal/connectors/certify	398.504s`. `make verify` completed fmt, tidy-check, vet, `go test -timeout 20m ./...`, build, docs validate, local smoke, golangci-lint, and `connectorgen validate` with `0 findings`.

```bash
git diff --check origin/feat/cli-architecture-v2...HEAD
git diff -- go.mod go.sum
```

Result: pass / no output.

```bash
./pm help connections
./pm connections
./pm connections --help
./pm connections --json
./pm connections bogus --json
```

Result: pass. Help/bare/`--help` byte-identical (1148 bytes); JSON manual emitted `CommandManual` (1299 bytes); invalid action exited 2 with usage JSON and stderr `error: unknown command "bogus" for "pm connections"`.

```bash
./pm docs generate --dir "$TMP/cli" --connectors-dir "$TMP/connectors"
diff -ru docs/cli "$TMP/cli"
./pm docs validate --connectors-dir docs/connectors
npm run gen:docs --prefix website
```

Result: pass. Docs diff had no output; docs validate passed; website generator wrote 11 docs pages with no tracked diff.

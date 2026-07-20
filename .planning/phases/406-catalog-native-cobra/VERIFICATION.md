# Phase 406 Verification

## Required gate checklist

- [x] `gofmt -w cmd internal`
- [x] `go test ./internal/cli/ -run 'Catalog|CobraRouterShell|Golden' -count=1`
- [x] `go test ./internal/cli/ -run Certify -count=1`
- [x] `go vet ./...`
- [x] `go test ./...`
- [x] `go build ./cmd/pm`
- [x] `make verify`
- [x] `git diff --check origin/feat/cli-architecture-v2...HEAD` (rerun after final commit; no output)
- [x] `git diff -- go.mod go.sum` (no output)

## CLI parity checklist

- [x] Golden transcript diff empty: `go test ./internal/cli/ -run 'Catalog|CobraRouterShell|Golden' -count=1` passed and `git diff -- internal/cli/testdata/golden_transcripts.json docs/cli website | wc -l` returned `0`.
- [x] `./pm help catalog` checked: exit 0, 391 bytes.
- [x] Bare `./pm catalog` checked: exit 0, 391 bytes; byte-identical to `pm help catalog`.
- [x] `./pm catalog --help` checked: exit 0, 391 bytes; byte-identical to `pm help catalog`.
- [x] JSON manual checked: `./pm catalog --json` exit 0, 536 bytes.
- [x] Invalid action checked: `./pm catalog bogus --json` exit 2, JSON category `usage`, stderr `error: unknown command "bogus" for "pm catalog"`.
- [x] `--connection` space, equals, repeated last-wins, bare `--connection`, unknown action flags, and late `--root`/`--json` placement covered by `TestCatalogConnectionFlagFormsPreserveLegacySemantics` and focused CLI gate.
- [x] `docs/cli/catalog.md` parity checked by golden docs-generate-diff test; no update needed.
- [x] Website docs/source/generated data grep checked under `website/content` and `website/lib`; no output change, no update needed.
- [x] No `pm connectors catalog` scope drift: connector catalog tests/goldens included in focused gate.

## Optional / safety-limited

- [x] Runtime-backed integration tests not run; no services started.
- [x] No credentialed connector checks.
- [x] No reverse ETL external execution. `make verify` ran the repository's local smoke flow, including local reverse run into a temp outbox, as required by the project gate.
- [x] No new dependencies.

## Results

```bash
go test ./internal/cli/ -run 'Catalog|CobraRouterShell|Golden' -count=1
```

Result: pass (`ok  	polymetrics.ai/internal/cli	16.830s`).

```bash
go test ./internal/cli/ -run Certify -count=1
```

Result: pass (`ok  	polymetrics.ai/internal/cli	90.214s`).

```bash
gofmt -w cmd internal
go vet ./...
go test ./...
go build ./cmd/pm
make verify
```

Result: pass. `make verify` completed fmt, tidy-check, vet, `go test -timeout 20m ./...`, build, docs validate, smoke, golangci-lint, and `connectorgen validate` with `0 findings`.

```bash
git diff --check origin/feat/cli-architecture-v2...HEAD
git diff -- go.mod go.sum
```

Result: pass / no output after final implementation commit.

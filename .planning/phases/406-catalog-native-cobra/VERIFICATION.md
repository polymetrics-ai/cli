# Phase 406 Verification

## Required gate checklist

- [ ] `gofmt -w cmd internal`
- [ ] `go test ./internal/cli/ -run 'Catalog|CobraRouterShell|Golden' -count=1`
- [ ] `go test ./internal/cli/ -run Certify -count=1`
- [ ] `go vet ./...`
- [ ] `go test ./...`
- [ ] `go build ./cmd/pm`
- [ ] `make verify`
- [ ] `git diff --check origin/feat/cli-architecture-v2...HEAD`
- [ ] `git diff -- go.mod go.sum` (must be empty)

## CLI parity checklist

- [ ] Golden transcript diff empty or explicitly justified.
- [ ] `pm help catalog` checked.
- [ ] Bare `pm catalog` checked: contextual help, exit 0.
- [ ] `pm catalog --help` checked: canonical docs-map help, exit 0.
- [ ] JSON manual checked (`pm catalog --json` or equivalent).
- [ ] Invalid action checked: usage error, not contextual help.
- [ ] `--connection` space, equals, repeated last-wins, and bare `--connection` legacy behavior checked.
- [ ] Late global `--root`/`--json` placement checked for catalog path.
- [ ] `docs/cli/catalog.md` parity checked; no update expected.
- [ ] Website docs/source/generated data grep/generator checked or marked not applicable; no output change expected.
- [ ] No `pm connectors catalog` scope drift; focused tests/goldens cover connector catalog unchanged.

## Optional / safety-limited

- [ ] Runtime-backed integration tests not run unless services already available and explicitly needed.
- [ ] No credentialed connector checks.
- [ ] No reverse ETL execution outside existing local project smoke gates.
- [ ] No new dependencies.

## Results

Pending.

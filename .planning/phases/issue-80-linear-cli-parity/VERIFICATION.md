# Verification Checklist — Issue #80 Linear CLI parity parent

Date: 2026-07-09

## Required adapter checks

- [x] `scripts/gsd doctor` — pass.
- [x] `scripts/gsd verify-pi` — pass.
- [x] `scripts/gsd list --json` — ran; registry available.
- [x] `scripts/gsd prompt plan-phase issue-80-linear-cli-parity --skip-research` — generated prompt.
- [!] `scripts/gsd prompt programming-loop init --phase issue-80-linear-cli-parity --dry-run` — unavailable; manual-GSD fallback recorded.

## Parent minimum gates before handoff

```bash
gofmt -w cmd internal
go vet ./...
go test ./...
go build ./cmd/pm
make verify
go run ./cmd/connectorgen validate internal/connectors/defs
```

## Issue #97 focused gates

```bash
go test ./internal/connectors/engine -run 'TestLinearCLISurface' -count=1
go run ./cmd/connectorgen validate internal/connectors/defs/linear --json
go test ./internal/connectors/engine ./internal/connectors/commandrunner -count=1
```

## CLI help/docs/website parity checks

Applies because connector command surface metadata is CLI-visible. Complete or explicitly exempt:

- [ ] `pm help linear` or nearest supported help topic.
- [ ] `pm linear` bare connector command help when implemented, or record not applicable for metadata-only slice.
- [ ] `pm linear --help` when implemented, or record not applicable for metadata-only slice.
- [ ] `docs/cli/**` grep/update.
- [ ] `website/**` grep/update.
- [ ] generated help/manual artifacts, if affected.

## Current status

#97 focused slice verification completed:

```bash
go test ./internal/connectors/engine -run TestLinearCLISurfaceMapsImplementedStreams -count=1
# pass

go test ./internal/connectors/engine ./internal/connectors/commandrunner -count=1
# pass

go test ./internal/connectors/conformance -run 'TestConformance/linear' -count=1
# pass

go run ./cmd/connectorgen validate internal/connectors/defs --json
# pass: 0 findings, 547 connectors checked

npm --prefix website run gen:website-data
# pass

npm --prefix website run test:unit
# blocked: local website deps missing (`vitest: command not found`)

go vet ./...
# pass

go test ./...
# pass

go build ./cmd/pm
# pass

./pm docs validate --connectors-dir docs/connectors
# pass

make verify
# pass

git diff --check
# pass
```

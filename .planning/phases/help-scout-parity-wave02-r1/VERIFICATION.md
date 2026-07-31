# Help Scout parity wave02 r1 — verification

## Focused gates

```bash
go test ./cmd/connectorgen -run 'TestHelpScoutFullSurface' -count=1
```

Result: pass (`ok polymetrics.ai/cmd/connectorgen`).

```bash
tmp=$(mktemp -d); mkdir -p "$tmp/help-scout"; cp -R internal/connectors/defs/help-scout/. "$tmp/help-scout/"; go run ./cmd/connectorgen validate "$tmp"; rm -rf "$tmp"
```

Result: pass (`1 connector(s) checked, 0 findings`).

```bash
go test ./internal/connectors/conformance -run 'TestConformance/help-scout' -count=1
```

Result: pass (`ok polymetrics.ai/internal/connectors/conformance`).

```bash
go test ./internal/connectors/engine -count=1
go test ./internal/connectors/bundleregistry -count=1
go test ./cmd/connectorgen -count=1
```

Result: pass.

## Broader local gates

```bash
go vet ./internal/connectors/... ./internal/cli/...
go vet ./...
go build ./cmd/pm
make connector-boundary
make connectorgen-validate
make smoke-no-build
make docs-check
make lint
git diff --check
```

Result: pass. `make connector-boundary` outcome clean; `make connectorgen-validate` checked 549 connectors with 0 findings; `make lint` reported 0 issues.

```bash
go test ./...
go test ./internal/connectors/...
```

Result: incomplete/timeout at 10 minutes each. Before timeout, `go test ./...` had passed through `internal/app`; `go test ./internal/connectors/...` had passed `internal/connectors`, `internal/connectors/boundary`, and `internal/connectors/bundleregistry`. No failing package output was captured before timeout. Focused Help Scout, engine, conformance, bundleregistry, connectorgen, vet, build, smoke, docs, lint, and boundary gates passed.

## CLI/help/docs parity

- Added connector-local `cli_surface.json` with 144 fixed Help Scout command metadata rows.
- Added/updated `docs.md`, `api_surface.json`, `operations.json`, `certification.json`, stream/write schemas, and sanitized fixtures under `internal/connectors/defs/help-scout/**`.
- No shared runtime CLI files, website generated data, or broad generated artifacts were edited in this worker slice.
- Direct/report provider-query and binary-download operations remain blocked/planned behind shared foundations; no raw generic API/query/file escape hatch is exposed.

## GitHub issue addendum

Verified `help-scout-captain-policy-addendum-wave02-r1` marker is present on #212-#219 after `gh-axi issue edit --body-file` updates.

# Verification: GitHub CLI Surface Metadata

Date: 2026-07-06

## Commands

```bash
jq empty internal/connectors/defs/github/cli_surface.json .planning/phases/github-cli-surface-metadata/RUN-STATE.json
go test ./internal/connectors/engine -run CLISurface
go test ./cmd/connectorgen -run CLISurface
go test ./cmd/connectorgen ./internal/connectors/engine
go test ./internal/connectors/conformance -run 'TestConformance/github'
go vet ./...
go build ./cmd/pm
go run ./cmd/connectorgen validate internal/connectors/defs
./pm docs validate --connectors-dir docs/connectors
cd website && pnpm run gen:website-data
cd website && pnpm run typecheck
cd website && pnpm run build
go test ./...
make verify
```

## Result

- JSON parse checks passed.
- Focused engine and connectorgen CLI-surface tests passed.
- Full engine and connectorgen package tests passed.
- GitHub conformance passed.
- Vet and `go build ./cmd/pm` passed.
- Full connector-bundle validation passed: 547 connector(s) checked, 0 findings.
- Checked-in connector docs validation passed.
- Website data generation, typecheck, and production build passed; the build generated 1113 static
  pages.
- Full `go test ./...` passed.
- `make verify` passed, including gofmt, tidy-check, vet, full tests, build, connector docs
  validation, smoke, lint, and connector definition validation.

## Not Yet Run

- CodeRabbit review is pending until a PR is opened or updated.

# Verification: GitHub CLI Surface Metadata

Date: 2026-07-07

## Commands

```bash
jq empty internal/connectors/defs/github/cli_surface.json .planning/phases/github-cli-surface-metadata/RUN-STATE.json
go test ./internal/connectors/engine -run CLISurface
go test ./cmd/connectorgen -run CLISurface
go test ./cmd/connectorgen -run TestValidate_CLISurfaceAPIRefFailsWhenSurfaceHasZeroEndpoints -count=1
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
- Review-driven zero-endpoint API-surface regression passed.
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

## Review Loop Notes

- CodeRabbit PR #48 produced one actionable finding: CLI-surface endpoint-reference validation was
  skipped when `api_surface.json` existed but had zero endpoints.
- Fixed with a red/green regression and the validator guard change.
- Stacked PR review routing was hardened in `.agents/agentic-delivery/` so skipped CodeRabbit
  statuses on non-`main` bases are not treated as review completion.

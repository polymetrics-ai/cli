# TDD Ledger: GitHub CLI Surface Metadata

## Red Tests

- `go test ./internal/connectors/engine -run CLISurface`
  - Initial failure: `Bundle.CLISurface undefined`.
- `go test ./cmd/connectorgen -run CLISurface`
  - Initial failure: `ruleCLISurfaceUnknownTarget` and `ruleCLISurfaceMissingMapping` undefined.
- Review-driven safety red cases were added before tightening the validator:
  - Excluded `api_surface.json` endpoint references should fail.
  - API endpoint references whose `covered_by` target does not match the command stream/write should
    fail.
  - Implemented or partial reverse ETL commands without `risk` and `approval` text should fail.
  - `availability=implemented` with `raw_api` or `direct_write` intent should fail.
- Embedded runtime regression:
  - `engine.Load(defs.FS, "github")` must load non-nil `CLISurface`.
- CodeRabbit review regression:
  - `go test ./cmd/connectorgen -run TestValidate_CLISurfaceAPIRefFailsWhenSurfaceHasZeroEndpoints -count=1`
  - Initial failure: a present `api_surface.json` with `endpoints: []` reported only
    `surface_incomplete`; the `cli_surface.json` endpoint references were skipped because
    validation was guarded by `len(endpoints) > 0`.

## Green Tests

- `jq empty internal/connectors/defs/github/cli_surface.json .planning/phases/github-cli-surface-metadata/RUN-STATE.json`
  - Passed.
- `go test ./internal/connectors/engine -run CLISurface`
  - Passed.
- `go test ./cmd/connectorgen -run CLISurface`
  - Passed.
- `go test ./cmd/connectorgen -run TestValidate_CLISurfaceAPIRefFailsWhenSurfaceHasZeroEndpoints -count=1`
  - Passed after changing CLI endpoint-reference validation to run whenever `b.Surface != nil`,
    including an intentionally empty endpoint set.
- `go test ./cmd/connectorgen ./internal/connectors/engine`
  - Passed.
- `go test ./internal/connectors/conformance -run 'TestConformance/github'`
  - Passed.
- `go vet ./...`
  - Passed.
- `go build ./cmd/pm`
  - Passed.
- `go run ./cmd/connectorgen validate internal/connectors/defs`
  - Passed: 547 connector(s) checked, 0 findings.
- `./pm docs validate --connectors-dir docs/connectors`
  - Passed.
- `cd website && pnpm run gen:website-data`
  - Passed.
- `cd website && pnpm run typecheck`
  - Passed.
- `cd website && pnpm run build`
  - Passed: 1113 static pages generated.
- `go test ./...`
  - Passed.
- `make verify`
  - Passed, including gofmt, tidy-check, vet, full tests, build, docs validation, smoke, lint, and
    connectorgen validation.

## Refactor Notes

- The repo's JSON schema compiler does not support `definitions`/`$ref`; the `cli_surface.json`
  schema keeps repeated flag shapes inline.
- Secret-looking example detection was tightened to catch GitHub token prefixes such as `ghp_` and
  `github_pat_`.
- `connectorgen validate` expects a root containing connector directories. Passing
  `internal/connectors/defs/github` treats `fixtures` and `schemas` as bundle directories and is not
  the right validation gate.
- `defs.FS` must embed `*/cli_surface.json`; otherwise disk validation passes but shipped bundles
  lose command-surface metadata at runtime.
- The website command map should only list dispatched `pm` commands. GitHub CLI parity metadata now
  lives in a separate connector metadata section until a dispatcher exists.

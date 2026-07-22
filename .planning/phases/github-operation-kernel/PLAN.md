# GitHub Operation Kernel Plan

Issue: #56
Parent issue: #44
Branch: `feat/56-operation-kernel`
Base: `feat/44-github-cli-parity`

## Objective

Add a JSON-first connector operation kernel foundation so `cli_surface.json`
commands can reference typed operation definitions without enabling execution
before the dedicated executors land.

## GSD Mode

Manual fallback. The repository does not contain the expected
`scripts/programming-loop.mjs` or `scripts/tdd-gate.mjs` helpers in this
checkout. The phase still follows the GSD programming loop manually:

1. Plan the slice.
2. Add red tests before production code.
3. Implement the minimum green behavior.
4. Refactor only after focused tests pass.
5. Run targeted and broader verification.
6. Commit and push green checkpoints.
7. Open a stacked sub-PR against the parent branch.
8. Use the automated review routing and Claude loop before integration.

## Scope

- Add optional `operations.json` loading to connector bundles.
- Add typed operation metadata with closed operation kinds.
- Add `operation` references to command surface commands.
- Validate operation references and unsafe/generic operation shapes.
- Expose operation IDs through the public command surface.
- Keep operation execution blocked by default in `commandrunner`.
- Add a small GitHub operation metadata file as reviewed examples.

## Non-Goals

- No GraphQL executor implementation.
- No reverse-ETL command execution.
- No binary transfer executor.
- No local git executor.
- No new dependencies.
- No unrestricted raw HTTP, GraphQL, SQL, or shell escape hatch.

## PR Slice

This PR should be a foundation-only stacked PR:

- PR title: `feat(connectors): add typed operation metadata foundation`
- PR base: `feat/44-github-cli-parity`
- PR body: `Refs #56` and `Refs #44`

## Expected Files

- `internal/connectors/engine/bundle.go`
- `internal/connectors/engine/metaschemas.go`
- `internal/connectors/engine/schema/operations.schema.json`
- `internal/connectors/engine/schema/cli_surface.schema.json`
- `internal/connectors/defs/defs.go`
- `internal/connectors/defs/github/operations.json`
- `internal/connectors/defs/github/cli_surface.json`
- `internal/connectors/command_surface.go`
- `internal/connectors/engine/connector.go`
- `internal/connectors/commandrunner/runner.go`
- `internal/connectors/engine/bundle_test.go`
- `cmd/connectorgen/main_test.go`
- `cmd/connectorgen/validate.go`
- `internal/connectors/commandrunner/runner_test.go`
- `docs/architecture/connector-operation-kernel.md`

## TDD Plan

1. Red: loader rejects/does not understand `operations.json`.
2. Red: `cli_surface.json` command `operation` field is rejected.
3. Red: validator misses unknown operation references and unsafe generic
   operation kinds.
4. Red: runner does not expose feature-gated operation blocking.
5. Green: add schema/types/load/validation/public conversion/blocking.
6. Refactor: keep stream/write/direct-read behavior unchanged.

## Verification Plan

Focused:

```bash
jq . internal/connectors/defs/github/operations.json
go test ./internal/connectors/engine -run 'TestBundleLoad.*Operation|TestBundleLoadEmbeddedGitHub'
go test ./cmd/connectorgen -run 'TestValidate_.*Operation|TestValidate_CLISurface'
go test ./internal/connectors/commandrunner
go build ./cmd/pm
```

Broader:

```bash
go test ./internal/connectors/...
go test ./cmd/...
go vet ./...
go test ./...
go build ./cmd/pm
make verify
```

## Human Gates

- New Go dependencies.
- Auth scope changes.
- Secret handling changes.
- Destructive external actions.
- Production deployment.
- Lowering verification gates.
- Generic shell, unrestricted HTTP write, generic SQL write, or arbitrary
  GraphQL execution.

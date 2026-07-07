# GitHub GraphQL Engine Phase Plan

Issue: #39
Parent issue: #44
Branch: `feat/39-github-graphql-engine`

## Objective

Add a safe declarative GraphQL transport primitive to the connector engine so GitHub CLI parity work
can promote GraphQL-backed commands from `partial`/`unsupported_api` to implemented command slices.

## Scope

- Support fixed GraphQL read bodies for declarative streams.
- Support fixed GraphQL mutation bodies for declarative write actions.
- Fail closed on GraphQL top-level `errors`.
- Validate the schema/type shape for static GraphQL documents and declared variable templates.
- Preserve reverse ETL plan, preview, approval, execute behavior for mutations.

## Non-Scope

- No arbitrary `gh api graphql` style raw query execution.
- No user-supplied GraphQL document text in records, flags, prompts, or config.
- No connector bundle expansion beyond minimal engine proof in this phase.
- No GitHub live API calls or credential use.
- No new Go dependencies.

## Implementation Tasks

1. Red tests: read path sends declared GraphQL payload and fails on `errors[]`.
2. Red tests: write path sends declared GraphQL mutation payload and ignores record-provided `query`.
3. Add shared GraphQL request payload helpers using existing interpolation rules.
4. Wire GraphQL read payloads into `Requester.Do` without changing REST default behavior.
5. Wire GraphQL write payloads behind `body_type: graphql`.
6. Add loader/schema validation for GraphQL write/read definitions.
7. Run focused tests, format, and local verification.

## Safety

- GraphQL documents live in bundle metadata and are reviewed code artifacts.
- Runtime input can only fill declared `variables` templates.
- GraphQL mutation execution remains under existing reverse ETL writer entry points.
- Dry run stays network-free and previews only method/path, not credentials.

## Verification Plan

- `go test ./internal/connectors/engine -run 'TestReadGraphQL|TestWriteGraphQL|TestBundleLoad.*GraphQL'`
- `go test ./internal/connectors/engine`
- `go run ./cmd/connectorgen validate internal/connectors/defs/github`
- `git diff --check`

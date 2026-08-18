Refs #4015

## Intent

Deliver P2's connector-neutral GraphQL certification mechanism without creating a second GraphQL executor or adding connector identifiers to shared Go.

## What Changed

- Added a definition-driven GraphQL certification profile, schema-lock compiler, report capability, and `unexecutable` handling for declared profiles that cannot compile.
- Added the GitHub-owned source-lock/profile data and two bounded read-only produced-value assertions.
- Classified all 305 generated GraphQL commands without promoting unexecuted work to pass: 29 `schema_conformant`, 2 `eligible_pending_live`, and 274 `fixture_required`.
- Regenerated `certification-sweep.json` from the connector definition.

## Truthfulness and Test Cases

- Happy: a serial product-path `pm connectors certify github --direct-read-only` run used the disposable identity and passed both GraphQL assertions (`rate-limit` value type and `viewer.__typename`).
- Bad: after schema compilation, changing only the definition-owned viewer assertion from `User` to `Organization` made `TestGraphQLCertificationInventoryRejectsIncorrectProducedValueAfterSchemaConformance` fail at the asserted JSON pointer; restoring it passed.
- Edge: a declared GraphQL profile whose source lock cannot compile is emitted as `unexecutable`, not skipped; all other unexecuted command rows retain a concrete non-pass status/reason.

Schema conformance verifies source-pinned fixed root selection and source argument binding. It does not prove returned provider values, authorization, or mutation effects; those remain live-required or fixture-bound.

## GSD / TDD

- Lifecycle: `discuss-phase` → `plan-phase --tdd` → `execute-phase` → `verify-work` → `code-review`.
- Commands were resolved through `scripts/gsd sources` and prompts were completed inline because the direct-PR brief and canonical contract prohibit spawned roles.
- Red, green, live-probe, verification, and inline review records: `.planning/phases/github-graphql-certification-mechanism-r1/`.

## Skills

`golang-how-to`, `golang-graphql`, `golang-design-patterns`, `golang-structs-interfaces`, `golang-error-handling`, `golang-security`, `golang-safety`, and `golang-testing`.

## Verification

- `go test -timeout 20m ./internal/connectors/certify`
- `go test -timeout 20m ./internal/connectors/engine`
- `go test -timeout 20m ./cmd/connectorgen`
- `go test -timeout 20m ./internal/cli`
- `go vet ./...`; `go build ./cmd/pm`
- `make fmt tidy-check lint agent-contract-check connectorgen-validate connectorgen-surface-sync github-parity-artifacts-check connectorgen-certification-matrix connectorgen-certification-sweep connector-boundary connector-canon-check release-workflow-check docs-check smoke-no-build`
- `pnpm --dir website run gen:docs` twice, then `git diff --exit-code -- website`
- `scripts/verify-gsd-workflow`

All passed locally. Full `go test ./...` / `make verify` were not run as one process because repository guidance requires package-scoped commands under the per-command timeout; the changed packages, their `internal/cli` consumer, and all non-test `make verify` gates above were run separately.

## CLI parity

No command, flag, help text, manual source, or website documentation surface changed. `pm help connectors` was exercised before the live certification command; `make docs-check` and the twice-stable website generator passed.

## Safety

No mutation was sent. The disposable token was read only into an exported variable by command substitution, passed through `--from-env`, unset after use, and never printed, stored, logged, or committed.

## Review

Inline review: no findings (recorded in `REVIEW.md`). Automated route: `claude_auto`, pending PR-open trigger; fallback: none. No Claude/Copilot comments exist yet, so no dispositions are required.

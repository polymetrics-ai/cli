Refs #4359

## Intent

Provide the narrowly closed engine proof that lets the retained CircleCI
provider identity `{project-slug}` correspond to the existing source-backed
runtime transport `vcs_type/org/repo`, without provider-route rewriting or a
generic path-template substitution feature.

## What changed

- Added the explicitly embedded and inventory-classified
  `circleci/composite_provider_path_identity.json` runtime declaration.
- Require exact CircleCI URL, retained SHA-256, placeholder, ordered config
  keys, and all eleven named source/binding rows before equivalence is possible.
- Resolve only the six ETL stream and five reverse-ETL write bindings through
  `composite_provider_path_identity`; reject all other lanes and malformed,
  conflicting, partial, reordered, repeated, extra, absolute, query, route,
  base, literal, and hook transport variants.
- Added red/green/refactor evidence and a fresh independent-Codex review record.

## TDD and lifecycle

- Red: the eleven-binding test initially failed because the proof/configuration
  type did not exist.
- Green: focused engine matrix passes for all eleven positive bindings and all
  adversarial closure cases.
- GSD lifecycle was executed inline because compatible isolated GSD workers
  were unavailable; resolved commands and fallback are recorded in the phase
  artifacts. Required Go skills: `golang-how-to`, `golang-cli`,
  `golang-testing`, `golang-error-handling`, `golang-security`,
  `golang-safety`, `golang-design-patterns`, and `golang-structs-interfaces`.

## Verification

- `go test -timeout 20m ./internal/connectors/defs ./internal/connectors/engine`
- `go test -timeout 20m ./internal/connectors/commandrunner`
- `go test -timeout 20m ./internal/app`
- `go build ./cmd/pm`
- `go vet ./internal/connectors/engine ./internal/connectors/defs ./internal/connectors/commandrunner ./internal/app ./internal/cli`
- `go run ./cmd/connectorgen validate internal/connectors/defs`
- `go run ./cmd/connectorgen surface-sync internal/connectors/defs --check`
- `go run ./cmd/connectorgen declaration-admission internal/connectors/defs`
- `go run ./cmd/connectorgen operation-evidence . --check`
- `make docs-check`
- `make tidy-check lint smoke-no-build agent-contract-check connectorgen-validate connectorgen-surface-sync connector-boundary release-workflow-check`
- `go test -timeout 20m ./internal/cli`

All listed commands passed locally; their exact results are recorded in the
verification artifact and handoff.
`source-import circleci --check` is deliberately unavailable on this foundation
base: Batch 1's retained CircleCI source-lock directory is not imported here.

## Credential and command-surface boundary

This foundation intentionally does not add Batch 1's generated CircleCI CLI
surface. The built binary therefore reports `unknown command "circleci"` for
the eleven future paths in this branch. The required fresh-project,
credential-free `missing --credential` proof belongs to the post-foundation
Batch 1 integration and is explicitly deferred there; no credentials, provider
I/O, request/body escape, or raw transport capability is introduced here.

## Review disposition

Claude Code was unavailable. A fresh-context independent Codex re-review of
`b9b2478…880c3c452` passed with no findings. It verified source identity,
closed endpoint surface, runtime embedding policy, six-lane isolation, and the
honest Batch 1 credential-boundary limitation. No Copilot fallback was needed.

Refs #4359

## Intent

Provide the narrowly closed engine proof that lets the retained CircleCI
provider identity `{project-slug}` correspond to the existing source-backed
runtime transport `vcs_type/org/repo`, without provider-route rewriting or a
generic path-template substitution feature.

## What changed

- Added the explicitly embedded and inventory-classified
  `circleci/composite_provider_path_identity.json` runtime declaration.
- Integrated current `main` at `2165619e` by a normal merge. Its only conflict
  was the generator-owned certification subject, regenerated canonically; no
  CircleCI route, source lock, CLI surface, or command mapping changed.
- Keep the CircleCI URL, retained SHA-256, placeholder, ordered config keys,
  and all eleven named source/binding rows in its one declaration. The
  provider-neutral engine validates a closed source-cited record shape and
  resolves only a row's exact declared inverse; Batch 1's retained source lock
  is the independent witness for the declaration's exact CircleCI values.
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
- `go run ./cmd/connectorgen certification-subject --check`
- `make connectorgen-certification-matrix connectorgen-certification-candidates connectorgen-certification-sweep`
- `make docs-check`
- `make tidy-check lint smoke-no-build agent-contract-check connectorgen-validate connectorgen-surface-sync connector-boundary release-workflow-check`
- `go test -timeout 20m ./internal/cli`

All listed commands passed locally; their exact results are recorded in the
verification artifact and handoff. The CI-boundary repair additionally passed
the full `make tidy-check lint smoke-no-build agent-contract-check
connectorgen-validate connectorgen-surface-sync connector-boundary
release-workflow-check` gate.
`source-import circleci --check` is deliberately unavailable on this foundation
base: Batch 1's retained CircleCI source-lock directory is not imported here.
Current-head Verify initially found the derived certification subject stale after
this declaration change; its owning generator refreshed only
`internal/connectors/certifications/current-subject.json`, and the strict
subject plus dependent certification checks now pass.
After the current-main merge, full engine/generator/definitions/commandrunner/
App/CLI tests, definition/source/generator checks, docs, lint, smoke, release,
and connector-canon gates passed locally. The current-head CI remains the
terminal record for the complete gate set.

## Credential and command-surface boundary

This foundation intentionally does not add Batch 1's generated CircleCI CLI
surface. The built binary therefore reports `unknown command "circleci"` for
the eleven future paths in this branch. The required fresh-project,
credential-free `missing --credential` proof belongs to the post-foundation
Batch 1 integration and is explicitly deferred there; no credentials, provider
I/O, request/body escape, or raw transport capability is introduced here.

## Review disposition

Claude Code was unavailable. Fresh-context independent Codex re-reviews of the
production repair (`b9b2478…56808f8d2`) and final certification recovery
(`b9b2478…fbc320592`) found no blocker. The first review's one low-severity
evidence wording finding is resolved here: exact CircleCI values are
declaration/source-lock evidence, while the shared engine enforces the closed
declaration shape and exact declared inverse. The final review confirmed that
the recovery changes only the generator-owned declaration fingerprint and
preserves the closed endpoint surface, runtime embedding policy, six-lane
isolation, and honest Batch 1 credential-boundary limitation. No Copilot
fallback was needed.

After current main #4356 was normally merged, a further fresh-context Codex
audit of exact base `2165619e` and head `50cb0e41d` passed with zero findings.
It verified the generator-only merge resolution, exact eleven-row source/CLI
witness, closed six lanes, no engine literals, and no CircleCI CLI expansion.

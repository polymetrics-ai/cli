# Issue #4299 — shipped definition embed inventory and binary budget guardrails

## Task Delivery Header

- Issue: Refs #4299 — perf(defs): stop embedding connector source locks in the shipped binary
- Base branch: main
- Merges into: main
- Delivery: Pull request open against `main` from `fm/cli-defs-embed-slim-r1`, with the declared local checks green and GitHub base verified through the API.
- Working branch: fm/cli-defs-embed-slim-r1
- Task: Make the runtime definition embed inventory explicit and bounded. Exclude all provider source locks except the one GitHub GraphQL lock that full installed certification needs, then add deterministic inventory and package-size checks. Preserve source files on disk, exact GitHub lock bytes and provenance, and offline certification behavior.
- Verification: Record fail-first inventory/budget tests; run focused defs/certification/release checks, a fresh release-like before/after binary measurement, connector generators and boundary checks, full repository verification, and an installed GitHub certification proof without credentials or network access.

## Evidence Table

| Acceptance criterion | Evidence | Observable assertion or fake reason |
| --- | --- | --- |
| Non-exempt `sources/**`, `api_surface.json`, and `fixtures/**` do not ship | live | Walk the real `defs.FS` and fail with the exact embedded path if any forbidden artifact remains. |
| The GitHub certification lock remains the sole source-lock exception | live | The real embedded byte slice equals the committed file byte-for-byte and its SHA-256 equals the committed provenance digest. |
| Certification retains its source-pinned GraphQL contract | live | `graphQLCertificationInventoryFor("github")` compiles all fixed GraphQL operations from the exception, with the expected classifications. |
| Package/binary guardrails are deterministic and attributable | live | The guardrail emits a sorted report with embedded files attributed by class and rejects an oversized binary/archive or forbidden inventory before release packaging. |
| An installed archive stays self-contained | live | A release-like archive is extracted outside the checkout; its `pm` executes the GitHub full-certification preflight/schema path without reading repository files. |

## GSD lifecycle and manual fallback

- `scripts/gsd doctor` and `go run ./cmd/agentcontractgen check` passed on 2026-08-20.
- Resolved commands: `discuss-phase`, `plan-phase`, `execute-phase`, `verify-work`, and `code-review` via `scripts/gsd sources <command>`.
- Generated and followed inline prompts: `scripts/gsd prompt discuss-phase 4299`, `scripts/gsd prompt plan-phase 4299 --tdd`, `scripts/gsd prompt execute-phase 4299`, `scripts/gsd prompt verify-work 4299`, and `scripts/gsd prompt code-review 4299`.
- Manual-GSD fallback: Pi role execution is unavailable in this worker and the canonical delivery contract forbids spawning planning/execution/review roles. The captain supplied the resolved Option A scope; this phase directory records equivalent discussion, TDD, execution, verification, and review evidence.
- Required skills loaded: `golang-how-to`, `golang-design-patterns`, `golang-structs-interfaces`, `golang-error-handling`, `golang-security`, `golang-safety`, `golang-testing`, `golang-graphql`, and `golang-documentation`.

## Locked decisions

1. Stage 1/2 only: inventory and package-size guardrails plus an embed-pattern correction. No JSON minification, compression, fleet lazy loading, checkout-root resolution, or connector-definition changes.
2. `github/sources/github-operation-source-lock.json` is the **only** raw runtime source-lock exception. Its repository bytes, `Raw*` identities, SHA-256 provenance, source URL, and offline behavior are contracts.
3. The exception is required in installed builds: GitHub full certification operates in an ephemeral project and `graphQLCertificationInventoryFor` reads the source-pinned schema at runtime. It cannot depend on an adjacent source checkout.
4. `api_surface.json`, `fixtures/**`, and all non-exempt `*/sources/*` remain repository/build-time artifacts and must be absent from `defs.FS`.
5. A source lock stays committed/readable on disk. This change governs shipped content only.

## TDD execution plan

| Order | Slice | Red evidence first | Green implementation | Guard |
| --- | --- | --- | --- | --- |
| 1 | Embed inventory | A real `defs.FS` walk reports non-GitHub `sources/**` paths from the broad glob and therefore fails the allowlist | Replace the glob with the one GitHub lock and expose a deterministic, sorted inventory | Reject `api_surface.json`, `fixtures/**`, any non-exempt source lock, duplicate paths, and non-deterministic report order |
| 2 | GitHub exception integrity | No test binds the embedded exception to the committed source bytes/digest | Assert literal byte equality and SHA-256 equality, then compile the actual GraphQL inventory | Keep `Raw*`/provenance semantics unchanged; no re-encoding/minification |
| 3 | Release-like budget | A broad-source synthetic/inventory fixture is admitted and an oversized package report has no enforcement | Add deterministic binary/archive report and enforce the committed budget from the release-like layout | Inspect extracted archive outside its build tree and invoke the installed binary's non-network certification boundary |
| 4 | Regression gates | Current source lock glob permits future provider locks automatically | Run generator/certification/boundary/release checks plus the full verification campaign | No connector definition changes; no runtime checkout fallback |

## Commit checkpoints

1. Planning/TDD/checklist checkpoint before production changes.
2. Red-test checkpoint recording the broad-glob inventory failure and absent archive budget contract.
3. Green inventory/exception/package guardrails checkpoint.
4. Review-fix checkpoint only for an in-scope defect.

## Explicit exclusions

- JSON minification or compression.
- Changes to `internal/connectors/defs/<connector>/` definitions or provider source locks.
- Disk/check-out-root fallback for certification.
- Fleet lazy loading, changed connector execution semantics, or new dependencies.
- Credentialed/live provider certification and reverse-ETL execution.

# Issue #3860 — Polling truthfulness

## Task Delivery Header

- Issue: Refs #3860 — docs(sync): surface polling-watermark limits and eligibility.
- Base branch: `integration/4015-mvp-flat-r1` (`ff6a8710199c10f209d9d47cce87e5c8f7c429e6` confirmed before edits).
- Merges into: `integration/4015-mvp-flat-r1` → `main`.
- Delivery: A pull request open against `integration/4015-mvp-flat-r1`, with the base API-read back, all applicable local checks recorded, and automated-review routing completed.
- Working branch: `fm/cli-3860-polling-truthfulness-r1`.
- Task: Establish polling-watermark behavior from the built binary, reconcile its eligibility and limits across runtime help, declaration-derived surfaces, CLI manual, and website generation inputs; regenerate derived documentation; and cover happy, bad, and edge behavior with observable assertions.
- Verification: Build `./cmd/pm`; run real-binary help/inspect/catalog probes for eligible and ineligible declarations; run targeted Go tests and the real-preflight sweep; run the documentation generator twice and compare output; run the repository verification gates named in the shared context; read the opened PR base through the GitHub API.

## Evidence Table

| Acceptance criterion | Evidence | Observable assertion or fake reason |
| --- | --- | --- |
| Runtime eligibility is established before declaration changes | live | A built `pm` is run against representative eligible and ineligible definitions; its rendered values or typed refusal are captured without credentials or provider I/O. |
| Public polling claims match that behavior | live | Runtime help, inspect/catalog output, CLI manual, and website-generator output express the same mechanism, eligibility, and limits. |
| Unsupported polling cannot be presented as executable | live | A test asserts the concrete typed preflight refusal and proves its transport is not invoked. |
| Valid polling remains executable | fake | Deterministic unit fixture is needed because this task does not authorize provider credentials; it asserts the produced eligibility value rather than only `err == nil`. |
| Generated documentation is stable | live | Two generator runs leave `git diff --exit-code` clean after the first generated update. |

## GSD / TDD ledger

This direct-PR task uses the required GSD sequence via `scripts/gsd prompt`: `discuss-phase`, `plan-phase 3860 --tdd`, `execute-phase`, `verify-work`, and `code-review`. The Pi runtime cannot provide the compatible isolated roles and repository policy forbids role spawning, so the generated workflows are executed inline; this trace is their durable artifact.

Loaded skills: `golang-how-to`, `golang-cli`, `golang-testing`, `golang-error-handling`, `golang-security`, `golang-safety`, `golang-documentation`, `vercel-react-best-practices`, and `vercel-composition-patterns`. The routing-referenced `frontend-design` and `web-design-guidelines` skills are not installed.

### Discuss — recorded scope

- Treat current built-binary behavior as authoritative over stale issue text.
- Preserve the fail-closed `integration_type == database` allowlist unless real-binary evidence proves a different executable capability.
- Do not alter help dispatch ordering, certification, generic transport registration, the four-path matrix, broker/MCP/UI, or unapproved I/O.

### Plan — red / green / refactor

1. **Red:** Add or extend a targeted test that demonstrates every source of a rendered polling claim is derived from the real eligibility/preflight result. Assert the exact result value or typed error and that unavailable mode rejects before transport I/O.
2. **Green:** Implement the smallest shared derivation or declaration correction necessary for runtime help, inspect/catalog, and documentation-generator inputs to agree.
3. **Refactor:** Remove duplicate/copyable capability wording, retain declaration ownership, and regenerate outputs through repository tooling.
4. **Verify:** Build the binary and run the planned real-binary probes, targeted tests, real preflight sweep, double generation, and named repository gates.

### CLI parity checklist

- [x] `pm <namespace>` behavior evaluated; no dispatch-order changes.
- [x] `pm help <topic>` evaluated against the built binary.
- [x] `pm <command> --help` evaluated against the built binary.
- [x] Invalid/unsafe polling behavior evaluated without credential or transport I/O.
- [x] `pm connectors inspect <name> --json` and catalog output evaluated.
- [x] `docs/cli/**`, generated manual/help artifacts, and `website/**` generator inputs reconciled.
- [x] PR body records help/manual/website parity, happy/bad/edge evidence, the chosen build-vs-correct-claim resolution, and verification outcomes.

### Verification checklist

- [x] `go build -o ./bin/pm ./cmd/pm`
- [x] Real-binary polling eligibility probes (happy, bad, edge)
- [x] Targeted package tests with `-timeout 20m`
- [x] Real `commandrunner.Preflight` sweep
- [x] Generated documentation twice, byte-stability confirmed
- [x] Individual applicable `make verify` gates
- [x] GSD verify-work and code-review recorded

## Execution evidence

### Runtime observation before source reconciliation

Built `./cmd/pm` and ran it in an isolated fixture project. A PostgreSQL
`incremental_dedupe` polling connection completed with `records_read: 3`,
`records_loaded: 3`, and `batch_count: 2`; its second run completed with zero
read and loaded records, proving the persisted polling checkpoint was used.
The same binary rejected an unknown cursor with exit `1` and
`postgres polling cursor field is absent from the selected relation`; no
`postgres-fixture-missing-cursor:public.users` checkpoint was written.

The implementation choice is **correct the claim, not remove the capability**.
The executable dynamic PostgreSQL polling provider was already present and the
binary proved it. The divergent text was the broad assertion that every
`planned` declaration was non-executable, even though a planned *static*
descriptor can be a placeholder for an implemented per-catalog declaration.

### TDD ledger

- **Red:** `go test -timeout 20m ./internal/cli -run '^(TestInspectPostgresKeepsStaticPollingWatermarkPlannedWhileRuntimeBindsItPerStream|TestPollingHelpDistinguishesStaticDeclarationsFromDynamicRuntimeEligibility)$' -count=1` failed: the manual omitted `Static declaration status` and global help denied dynamically constructed declarations.
- **Green:** The same focused command passed after the guide labels static status, the runtime help scopes non-implementation to the static declaration alone, and website source docs carry the same rule.
- **Refactor/generation:** `go run ./cmd/pm docs generate --dir docs/cli` regenerated CLI and connector documents; `pnpm --dir website run gen:docs` regenerated website input output. Each was run twice with identical post-generation diff hashes.
- **Happy:** `TestPMBinaryExecutesPostgresFixturePollingResume` asserts the compiled binary emits the 3-record/3-loaded, then 0-record/0-loaded results.
- **Bad:** `TestPostgresPollingTransportRefusesMissingPerStreamCursorBeforeIO` asserts the typed `ErrPollingCursorFieldRequired` and zero I/O; `TestPMBinaryRefusesPostgresFixturePollingUnknownStreamCursorBeforePageRead` asserts the surfaced missing-column refusal and no page/checkpoint state.
- **Edge:** `TestPollingPreflightRefusesEachUnsafeDeclarationBeforeSourceIO` checks unsafe declaration, stale-evidence, hard-delete, and incompatible-apply rejection before source or apply activity.

## Verify-work — inline execution

### Built-binary surface evidence

- `./pm help connectors`, bare `./pm connectors`, and `./pm connectors --help`
  all rendered the scoped rule: a planned/unsupported/absent **static**
  declaration alone does not implement polling, while an eligible connector may
  construct an implemented declaration per selected catalog object.
- `./pm connectors inspect postgres` rendered `Static declaration status:
  planned` and the dynamic runtime-eligibility explanation. `./pm connectors
  inspect postgres --json` retained the truthful static reason that the live
  transport constructs and preflights the declaration.
- `./pm etl --help` does not claim a polling capability. `./pm connections
  --help` renders the bounded, non-CDC polling limits, hard-delete limitation,
  and explicit-rebootstrap rule.

### Commands run locally

| Command | Result |
| --- | --- |
| `go build -o ./bin/pm ./cmd/pm` | pass |
| `go test -timeout 20m ./internal/cli -count=1` | pass |
| `go test -timeout 20m ./internal/connectors -count=1` | pass |
| `go test -timeout 20m ./internal/connectors/engine -count=1` | pass |
| `go test -timeout 20m ./internal/connectors/native/postgres -count=1` | pass |
| `go test -timeout 20m ./cmd/connectorgen -count=1` | pass; required consumer-package gate for changed guide and generated documentation surfaces |
| `go test -timeout 20m ./internal/cli -run '^(TestGoldenTranscripts\|TestInspectPostgresKeepsStaticPollingWatermarkPlannedWhileRuntimeBindsItPerStream\|TestPollingHelpDistinguishesStaticDeclarationsFromDynamicRuntimeEligibility\|TestPMBinaryExecutesPostgresFixturePollingResume\|TestPMBinaryRefusesPostgresFixturePollingUnknownStreamCursorBeforePageRead)$' -count=1` | pass |
| `go test -timeout 20m ./internal/connectors/engine -run '^(TestPollingPreflightAdmitsDeclaredPollingBeforeGuardedSourceRead\|TestPollingPreflightRefusesEachUnsafeDeclarationBeforeSourceIO\|TestPollingModeEligibilitySweepsEveryImplementedPollingModeThroughRuntimePreflight)$' -count=1` | pass |
| `go test -timeout 20m ./internal/connectors/native/postgres -run '^(TestPostgresPollingTransportResumesFixtureCursor\|TestPostgresPollingTransportRefusesMissingPerStreamCursorBeforeIO)$' -count=1` | pass |
| `go vet ./...` and `go build ./cmd/pm` | pass |
| `make docs-check tidy-check agent-contract-check connectorgen-validate connectorgen-surface-sync` | pass (run as individual make targets) |
| `make lint connector-boundary release-workflow-check` | pass (run as individual make targets) |
| `pnpm --dir website run typecheck` and `pnpm --dir website run test:scripts` | pass |
| `go run ./cmd/pm docs generate --dir docs/cli` twice | pass; identical SHA-256 after each: `41852315887f45cba7d3759911d1b0e2604173a6a62e3353f3c4a2f944d57bf7` |
| `pnpm --dir website run gen:docs` twice | pass; identical SHA-256 after each: `48d097ea7446e03e1afdd3d9d9ded4ee99fa6d23220108232f9949d42de4e0d8` |
| `git diff --check` | pass |

The monolithic `go test -timeout 20m ./...` is left to CI because this
repository's own AGENTS.md says not to run it as one command under a
per-command timeout; all changed and directly related packages above were run
individually. `make smoke-no-build` was not run: it executes reverse-ETL
behavior, which issue #3860 names as a non-goal for this lane.

## Code review — inline execution

Reviewed the final diff against `integration/4015-mvp-flat-r1` for
truthfulness, generated-file provenance, failure-before-I/O coverage, and
scope. No actionable finding: the static `planned` descriptor still says only
what it can know, the dynamic capability remains guarded by runtime preflight,
and all edited documentation is generated from or sourced by the documented
generators.

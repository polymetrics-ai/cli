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

- [ ] `pm <namespace>` behavior evaluated; no dispatch-order changes.
- [ ] `pm help <topic>` evaluated against the built binary.
- [ ] `pm <command> --help` evaluated against the built binary.
- [ ] Invalid/unsafe polling behavior evaluated without credential or transport I/O.
- [ ] `pm connectors inspect <name> --json` and catalog output evaluated.
- [ ] `docs/cli/**`, generated manual/help artifacts, and `website/**` generator inputs reconciled or explicitly not applicable.
- [ ] PR body records help/manual/website parity, happy/bad/edge evidence, the chosen build-vs-correct-claim resolution, and verification outcomes.

### Verification checklist

- [ ] `go build -o ./bin/pm ./cmd/pm`
- [ ] Real-binary polling eligibility probes (happy, bad, edge)
- [ ] Targeted package tests with `-timeout 20m`
- [ ] Real `commandrunner.Preflight` sweep
- [ ] Generated documentation twice, byte-stability confirmed
- [ ] Individual applicable `make verify` gates
- [ ] GSD verify-work and code-review recorded

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

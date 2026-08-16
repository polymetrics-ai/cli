# GitHub dedupe modes r1

## Task Delivery Header

- Issue: Refs #4166 — test(certification): prove the three unexercised-coverage gaps are closed (F3); related parent #4015.
- Base branch: `integration/4015-mvp-flat-r1`
- Merges into: `integration/4015-mvp-flat-r1` → `main`
- Delivery: direct PR open against `integration/4015-mvp-flat-r1`, with the API-reported base read back after creation.
- Working branch: `fm/cli-truth-github-dedupe-modes-build-r1`
- Task: Correct generator-derived certification-matrix mode cells, then make GitHub `incremental_dedupe` and `incremental_dedupe_history` run with real, observable warehouse dedupe semantics through the shipped binary. Regenerate affected matrix shards and prove a repeat run is byte-stable.
- Verification: red/green focused Go tests; `connectorgen certification-matrix` generation plus repeat-byte check; real registry preflight and fresh-built-binary checks; live private-GitHub happy/bad/edge proof with independent provider read-back; required scoped test, lint, generated-file, website-generator, connector-boundary, build, and diff gates.

## Evidence Table

| Acceptance criterion | Evidence | Observable assertion or fake reason |
| --- | --- | --- |
| Generator uses the real mode/transport intersection | live | A focused generator test shows a source that omits a mode cannot produce `declared` or `implemented` for that mode; before the correction it does. |
| Regenerated matrix shards stay stable | live | A scoped generation changes exactly derived matrix shards, then a second scoped run leaves their SHA-256 values unchanged. |
| Both GitHub modes pass production registry preflight | live | The real production registry returns the selected source, warehouse destination, and apply strategy for each mode; unsupported modes retain their typed pre-I/O refusal. |
| Dedupe does not duplicate a repeated source key | live | A built `pm` ETL run reaches the same GitHub item twice and an independent warehouse query sees one active key for dedupe mode. |
| History mode is replay-idempotent | live | A repeated built-binary run against the same GitHub key produces no second open/current history record; an independent read shows the expected retained history state. |
| Real-provider result is independently read back | live | GitHub REST is queried separately from the run report against the retained private repository and asserts the test item/data used by the source. |

## Lifecycle and skills

- Completed before planning: `scripts/gsd doctor`; `scripts/gsd sources` for `discuss-phase`, `plan-phase`, `execute-phase`, `verify-work`, and `code-review`; all five generated prompts; `go run ./cmd/agentcontractgen check`.
- Manual-GSD fallback: prompts are executed inline because no compatible isolated GSD worker is available and the task forbids role spawning.
- Skills loaded: `golang-how-to`, `golang-cli`, `golang-testing`, `golang-error-handling`, `golang-security`, `golang-safety`, `golang-design-patterns`, `golang-structs-interfaces`, `golang-context`, and `golang-concurrency`.

## Foundation check

| Need | Required proof | Planned evidence |
| --- | --- | --- |
| Truthful mode-cell derivation | Generator receives the mode-specific transport admission, not only a coarse capability bit. | Focused red/green generator test and generated-shard repeat stability. |
| Executable source and destination | Registry preflight resolves both modes into real registered source/destination/apply adapters. | Real `synctransport.Registry.Preflight` tests and a built-binary preflight route. |
| Warehouse materialization | Bounded GitHub reads write the owned warehouse table using the selected apply strategy. | Fresh `pm` executable with an isolated project and deterministic fixture tests. |
| Live provider proof | Private retained GitHub repository supports a bounded item with a stable key and a read-back independent of the writer's report. | Live `pm` proof and a separate GitHub REST read only after the ETL command completes. |

## TDD slices and checkpoints

1. **Generator intersection (red → green).** Add a focused test covering an otherwise-readable connector mode absent from `sync_transport.json`; verify it initially receives fabricated true cells, then derive its mode facts from the admitted source transport. Regenerate affected certification matrices and commit the generator plus output checkpoint.
2. **GitHub mode admission (red → green).** Add real-registry preflight tests for `incremental_dedupe` and `incremental_dedupe_history`, including a typed no-I/O refusal for a still-unadmitted mode. Add only definition-owned source-mode, destination, and apply declarations needed to make the two modes executable; do not name GitHub in shared production code.
3. **Observable dedupe behavior (red → green).** Test the production ETL path for duplicate key suppression and history replay idempotency. Keep test doubles for deterministic CI, but drive the shipped command construction path and assert record counts/state instead of just exit status.
4. **Live proof and delivery.** Build `pm`; create a private retained repository named `pm-truth-github-dedupe-modes-build-r1`; run bounded happy, bad, and edge scenarios with the authenticated CLI identity injected only by environment command substitution; independently read GitHub state; record only non-secret commands/results. Run full scoped local gates, then inline verify-work and code review, commit/push, open the direct PR, and API-verify its base.

## CLI help/manual/website parity

No command name, flag, output schema, help topic, or connector CLI surface is intended to change: this makes two already-declared sync modes executable. Runtime `pm help`, bare namespace, and GitHub command-help smoke checks remain required; generated matrix and website connector documentation regeneration/checks remain required. If inspection reveals changed user-visible mode documentation, update runtime help, `docs/cli/**`, website docs, and generated artifacts in the same slice.

## Scope fences

- Do not implement `incremental_append`.
- Do not alter PostgreSQL history mode, certification `pass` semantics, or `--full-parity`.
- Do not expand `issue_label_destination` beyond `add_issue_labels` and `set_issue_labels`.
- Do not add dependencies, generic write tools, provider-specific shared runtime branches, or allowlist exceptions.

# Plan: Gong calls list bounded date range and output-limit correction

Issue: #596
Branch: `fix/gong-calls-list-date-range-limit`
Mode: focused bounded calls-list correction

## GSD command path

- `scripts/gsd doctor`: passed (`commands 69`).
- Required GSD command attempted: `scripts/gsd prompt programming-loop init --phase issue-596 --dry-run`; unavailable in this adapter (`unknown GSD command: programming-loop`).
- Manual-GSD fallback active per AGENTS.md and `.agents/agentic-delivery/references/gsd-pi-adapter.md`.
- Supplemental adapter prompt inspected: `scripts/gsd prompt gsd-quick "issue 596 Gong calls list correction" --dry-run`.
- Resume checkpoint after compaction: reread ship instructions and decision record, reran `scripts/gsd doctor`, and inspected `scripts/gsd prompt gsd-execute-phase issue-596-gong-calls-list-date-range-limit --dry-run`; continuing manual-GSD execution because the phase artifact is issue-scoped rather than a numeric GSD phase and `programming-loop` remains unavailable.
- Reference checkpoint: current authoritative Gong OpenAPI docs verified before implementation finalization; private comparative notes are kept outside the repository and must not appear in tracked/public surfaces.
- PR 597 architecture-revision checkpoint: read captain revision instructions, definition-driven decision, architecture report, private-reference addendum, and review-transparency correction; reconciled old no-mistakes run `01KYKYYAPSP34HVH9VH5ZKK3WN` through `no-mistakes axi status/help` and cancelled it before new edits; reran `scripts/gsd doctor` and inspected `scripts/gsd prompt gsd-execute-phase issue-596-gong-calls-list-date-range-limit --dry-run`.

## Required skills loaded

- `gsd-core`
- `golang-how-to`
- `golang-cli`
- `golang-testing`
- `golang-error-handling`
- `golang-security`
- `golang-safety`
- `golang-design-patterns`
- `golang-structs-interfaces`
- `golang-documentation`
- `no-mistakes`

## Objective

Implement the captain-authorized focused correction for `pm gong calls list`: add bounded `--from`/`--to` filters, make output-limit semantics deterministic and explicit, and reconcile/document calls-list request page-size behavior without unrelated Gong endpoint changes.

## Scope

- Add command-specific `--from` and `--to` flags to `calls list` in `internal/connectors/defs/gong/cli_surface.json`.
- Validate ISO-8601 timestamps and reject `--from >= --to` before HTTP for this command.
- Map `--from` to Gong `fromDateTime`; map `--to` to `toDateTime`.
- Preserve `--config start_date=...` compatibility and use the established explicit command flag precedence over config-derived lower bounds.
- Keep `--limit` as PM emitted-record cap. Verify N=1, below page boundary, at boundary, and across cursor pages.
- Reconcile `limit={{ config.page_size }}` with current Gong OpenAPI by avoiding undocumented provider-side claim in help/docs while preserving backwards-safe request behavior unless tests prove removal is required.
- Update user-facing connector docs and generated website data only; do not update changelog, tag, or release.

## Non-goals

- No participant/parties ETL stream or caller/recipient output change.
- No live Gong API calls or real credentials.
- No reverse ETL execution.
- No unrelated Gong endpoint changes.
- No dependency changes, release tag, package publication, or merge.

## Existing convention / precedence decision

Existing commandrunner `queryOverrides` treats command-specific flags as explicit request query overrides. Existing engine `buildInitialQuery` currently applies incremental `start_config_key` after request query values, which can mask an explicit `--from` mapped to the same query key. For issue #596, use the smallest general correction: preserve command-specific query overrides over config-derived incremental lower bounds when both set the same query parameter. This matches CLI explicit-input precedence and preserves `--config start_date` when `--from` is absent.

## Architecture revision scope for PR 597

- Remove every Gong/provider-specific validation branch from `internal/connectors/commandrunner/runner.go`.
- Add only the smallest generic CLI-surface validation declarations/interpreter needed for string date-time format, non-empty values, one order constraint over mapped targets, config fallback when a side is absent, and definition-owned validation messages.
- Keep provider-specific names (`fromDateTime`, `toDateTime`, `start_date`, Gong wording) in Gong connector definitions and tests/docs, not shared runner code.
- Preserve the existing generic engine explicit-query precedence behavior with non-Gong tests.
- Keep current dynamic connector seam: definition-driven/manual-parsed command-specific flags; no Gong-specific Cobra/Viper binding or broad CLI framework migration.
- Revise PR 597 in place with additive transparent commits; no reset, force-push, competing PR, merge, tag, or release.

## Implementation slices

1. **Red tests — request shape and validation**
   - Commandrunner or CLI mock tests for `calls list --from`, `--to`, both together, invalid timestamp, and invalid range.
   - Test that explicit `--from` wins over configured `start_date` while `start_date` still works alone.
2. **Red tests — limit/cursor behavior**
   - Local mock server returns two cursor pages; assert counts and request counts for N=1, N=2, page boundary, and above page boundary.
3. **Green implementation**
   - Add `calls list` flags to Gong CLI surface.
   - Add generic CLI-surface validation for mapped date-time query/body values before connector dispatch.
   - Adjust engine query construction so explicit request query keys are not overwritten by incremental lower bound.
   - Preserve request `limit={{ config.page_size }}` but document it as compatibility/internal page-size config, not official Gong schema support.
4. **Docs/generated surfaces**
   - Update checked-in connector docs/manual/website generated surfaces as applicable.
   - Leave changelog/release tagging out of this PR; issue #596 remains scoped to the bounded Gong calls-list correction.
5. **Verification and PR**
   - Run targeted tests, broader selected gates, no-mistakes AXI, push branch, open PR with `Closes #596`.

## CLI help/docs/website parity checklist

- [x] `pm help gong` reflects `calls list` fields/command notes.
- [x] `pm gong calls list --help` documents `--from`, `--to`, and `--limit` output-cap language.
- [x] `docs/connectors/gong/MANUAL.md` and `docs/connectors/gong/SKILL.md` updated from generated output; `internal/connectors/defs/gong/docs.md` updated.
- [x] `docs/cli/**` generated docs unchanged/not applicable for connector-specific command surface.
- [x] `website/**` generated connector data updated from the Gong bundle/docs source.
- [x] Tests cover help rendering and command metadata.

## Safety

Synthetic fixtures/mock servers only. Do not use real Gong credentials, live account data, customer data, or write operations. Reverse ETL remains plan -> preview -> approval -> execute and is not executed here.

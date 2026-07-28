# Plan: Gong calls list bounded date range and output-limit correction

Issue: #596
Branch: `fix/gong-calls-list-date-range-limit`
Mode: focused PM v0.1.1 correction

## GSD command path

- `scripts/gsd doctor`: passed (`commands 69`).
- Required GSD command attempted: `scripts/gsd prompt programming-loop init --phase issue-596 --dry-run`; unavailable in this adapter (`unknown GSD command: programming-loop`).
- Manual-GSD fallback active per AGENTS.md and `.agents/agentic-delivery/references/gsd-pi-adapter.md`.
- Supplemental adapter prompt inspected: `scripts/gsd prompt gsd-quick "issue 596 Gong calls list correction" --dry-run`.

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

Implement the captain-authorized focused v0.1.1 correction for `pm gong calls list`: add bounded `--from`/`--to` filters, make output-limit semantics deterministic and explicit, and reconcile/document calls-list request page-size behavior without unrelated Gong endpoint changes.

## Scope

- Add command-specific `--from` and `--to` flags to `calls list` in `internal/connectors/defs/gong/cli_surface.json`.
- Validate ISO-8601 timestamps and reject `--from >= --to` before HTTP for this command.
- Map `--from` to Gong `fromDateTime`; map `--to` to `toDateTime`.
- Preserve `--config start_date=...` compatibility and use the established explicit command flag precedence over config-derived lower bounds.
- Keep `--limit` as PM emitted-record cap. Verify N=1, below page boundary, at boundary, and across cursor pages.
- Reconcile `limit={{ config.page_size }}` with current Gong OpenAPI by avoiding undocumented provider-side claim in help/docs while preserving backwards-safe request behavior unless tests prove removal is required.
- Update user-facing docs/changelog/release surface for v0.1.1 preparation only; do not tag/release.

## Non-goals

- No participant/parties ETL stream or caller/recipient output change.
- No live Gong API calls or real credentials.
- No reverse ETL execution.
- No unrelated Gong endpoint changes.
- No dependency changes, release tag, package publication, or merge.

## Existing convention / precedence decision

Existing commandrunner `queryOverrides` treats command-specific flags as explicit request query overrides. Existing engine `buildInitialQuery` currently applies incremental `start_config_key` after request query values, which can mask an explicit `--from` mapped to the same query key. For issue #596, use the smallest general correction: preserve command-specific query overrides over config-derived incremental lower bounds when both set the same query parameter. This matches CLI explicit-input precedence and preserves `--config start_date` when `--from` is absent.

## Implementation slices

1. **Red tests — request shape and validation**
   - Commandrunner or CLI mock tests for `calls list --from`, `--to`, both together, invalid timestamp, and invalid range.
   - Test that explicit `--from` wins over configured `start_date` while `start_date` still works alone.
2. **Red tests — limit/cursor behavior**
   - Local mock server returns two cursor pages; assert counts and request counts for N=1, N=2, page boundary, and above page boundary.
3. **Green implementation**
   - Add `calls list` flags to Gong CLI surface.
   - Add narrow validation for Gong calls list query values before connector read dispatch.
   - Adjust engine query construction so explicit request query keys are not overwritten by incremental lower bound.
   - Preserve request `limit={{ config.page_size }}` but document it as compatibility/internal page-size config, not official Gong schema support.
4. **Docs/release surface**
   - Update checked-in connector docs/manual/website generated surfaces as applicable.
   - Add PM v0.1.1 changelog entry.
5. **Verification and PR**
   - Run targeted tests, broader selected gates, no-mistakes AXI, push branch, open PR with `Closes #596`.

## CLI help/docs/website parity checklist

- [ ] `pm help gong` reflects `calls list` fields/command notes.
- [ ] `pm gong calls list --help` documents `--from`, `--to`, and `--limit` output-cap language.
- [ ] `docs/connectors/gong.md` updated or regenerated.
- [ ] `docs/cli/**` generated docs updated if generator changes them.
- [ ] `website/**` generated connector data/docs updated if source surfaces changed.
- [ ] Tests cover help rendering and command metadata.

## Safety

Synthetic fixtures/mock servers only. Do not use real Gong credentials, live account data, customer data, or write operations. Reverse ETL remains plan -> preview -> approval -> execute and is not executed here.

# PLAN — #3860 polling-watermark truth surfaces

## Task Delivery Header

- Issue: `Refs #3860 — docs(sync): surface polling-watermark limits and eligibility`
- Base branch: `integration/4015-mvp-flat-r1`
- Merges into: `integration/4015-mvp-flat-r1 → main`
- Delivery: Pull request open against `integration/4015-mvp-flat-r1` with green CI.
- Working branch: `fm/cli-3860-polling-docs-sweep-r1`
- Task: Project the real native polling-watermark preflight into connector help, inspection, catalog, manuals, generated artifacts, and the website without promoting polling to CDC or fabricating a REST surface.
- Verification: Targeted CLI and connector tests; fresh `pm` build before sanctioned generation; parity command output; documented generator checks; the required PostgreSQL database-integration run; individual repository gates; CI.

| Acceptance criterion | Evidence | Observable assertion or fake reason |
| --- | --- | --- |
| Runtime help, bare namespace behavior, inspect JSON, and catalog show real per-binding polling eligibility | fake | These outputs are deterministic projections of declaration/preflight fixtures. Tests execute the real renderer and assert implemented/blocked status and exact absence of a false implemented claim; provider calls and credentials are neither needed nor authorized. |
| Polling text discloses ordering/checkpoint, at-least-once replay, snapshot/deletion limitations, and rebootstrap rules without calling it CDC | fake | Tests inspect the rendered surface and generated artifacts. A fixture has no credentialed provider dependency; it is necessary to exercise unavailable, unsafe, soft-delete-only, identity-mismatch, and valid declaration states deterministically. |
| Native database discovery remains protocol-truthful with no invented REST endpoint | fake | The PostgreSQL bundle's real `api_surface.json` and inspected/generated projections are asserted to contain zero endpoints. No live provider can prove a metadata fabrication absence. |
| The native PostgreSQL connector still passes its database integration lane | live (waived) | The supplied Docker/Colima command was attempted but timed out while Docker blocked during the harness's image-store capacity probe; a read-only `docker info` probe also stalled. The supervisor waived this lane as machine saturation, not a code finding. No retry is authorized for this documentation/surface issue. |

## Discussed decisions

- `PollingModeEligibilityOf` remains the one authority for mode rows. No docs generator, help renderer, catalog formatter, or test may reconstruct its capability decision from copied strings.
- PostgreSQL now carries a declaration-owned `polling_watermark.json` with `status=planned` and an explicit no-binding reason. That status is visible in inspection/catalog output but serializes no empty source/target contract, and never display it as implemented merely because the shared executor exists.
- A legacy `changefeed.json` declaration and its logical-replication mechanism remain distinct from a polling scan. This issue does not alter #3748's changefeed ownership or describe polling as CDC.
- A database connector has no REST endpoint surface. PostgreSQL's zero-entry `api_surface.json` stays zero; tests assert the absence rather than substituting a synthetic endpoint.
- Rebootstrap outcomes are explicit operator action. State incompatibility, source identity mismatch, snapshot expiration, and retention failures must never be described as automatic full scans.
- This is an inline/manual GSD execution: the runtime adapter generated the required prompts, but this environment has no compatible Pi runtime and the canonical single-worker contract forbids role spawning. The artifacts below preserve the full discuss → plan (`--tdd`) → execute → verify → review record.

## Required skills and parity checklist

Loaded: `golang-how-to`, `golang-cli`, `golang-testing`, `golang-error-handling`, `golang-security`, `golang-documentation`, and `vercel-react-best-practices`. The routing-required `frontend-design` and `web-design-guidelines` skills are not installed in this runner; website documentation changes will follow the existing site patterns and no component API work is planned.

- [x] `pm connectors` exits successfully with contextual help.
- [x] `pm help connectors` and `pm connectors --help` accurately describe polling truth.
- [x] `pm connectors inspect <name> --json` and catalog output expose no unavailable mode as implemented.
- [x] `docs/cli/**` and `website/**` cover the same bounded polling contract.
- [x] Generated help/manual and website artifacts are regenerated from a freshly built binary.
- [x] No generic shell, HTTP write, SQL write, credential, or reverse-ETL execution surface is added.

## TDD slices

1. **Red — surface contract.** Add a focused renderer/CLI test for an unavailable declaration, unsafe cursor, soft-delete-only declaration, source-identity rebootstrap, and a registered valid declaration. Assert every status and the absence of false `implemented`/CDC wording.
2. **Red — native database REST absence.** Add an observable bundle/surface assertion that PostgreSQL remains a protocol-native connector with zero REST `api_surface` endpoints.
3. **Green — preflight-derived projection.** Add the smallest shared surface model/rendering path necessary to use `PollingModeEligibilityOf` and show mechanism, ordering/checkpoint scope, delivery/replay, snapshot, deletion, and rebootstrap facts.
4. **Green — parity documentation and sanctioned generation.** Update source manuals and website content, build a fresh `pm`, invoke the sanctioned generator, then review the generated diff line by line for only intended entries.
5. **Refactor/review.** Keep vocabulary closed, remove duplicated status logic, run help/inspect/catalog command parity, and audit docs for forbidden CDC/delete-complete/automatic-scan claims.

## Planned files

- `internal/cli/**` and existing focused CLI tests — preflight-derived help, inspect, and catalog surfaces.
- `internal/connectors/**` focused tests — declaration/preflight and native API-surface absence coverage only.
- `docs/cli/**`, `website/**`, and sanctioned generated artifacts — synchronized documentation.
- `.planning/phases/issue-3860-polling-docs-sweep-r1/*` — delivery evidence.

## Commit checkpoints

1. Plan/TDD evidence.
2. Focused red-test checkpoint.
3. Green surface/documentation/generation checkpoint.
4. Verification/review-fix checkpoint, rebase, push, and PR delivery.

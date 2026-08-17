# Summary — reverse-ETL API-surface derivation r1

## Outcome

`surface-sync` now derives endpoint metadata from an implemented command's declared `METHOD /path` summary before the intent-specific operation rules run. The parsed address must already occur in the bundle's `api_surface.json`; this keeps the derivation generic while preventing a summary-shaped sentence from creating a fictional binding.

Regeneration recovered all 214 GitHub reverse-ETL commands with one-to-one canonical endpoints. The remaining 14 intentional convenience aliases remain unbound: `cache delete`, `issue close`, `issue reopen`, `pr close`, `pr comment`, `pr lock`, `pr reopen`, `pr unlock`, `repo archive`, `repo create`, `repo delete`, `repo unarchive`, `secret delete`, and `secret set`.

## TDD and invariant evidence

- Red: the new all-bundle predicate exposed the existing broken state (214 genuine GitHub records plus 34 punctuated Workday summary strings). The focused one-command reverse-ETL fixture also failed with `api surface fills = 0`.
- Green: the source-bound predicate and focused derivation test pass after the generic synchronization change. The explicit punctuation and friendly-summary subtests prove that neither punctuation nor aliases are normalized into endpoints.
- The GitHub certification buckets are unchanged: `1466 fixture_required + 25 eligible_pending_live + 50 not_applicable + 29 schema_conformant + 1 provider_refused = 1571` before and after. This metadata repair makes no status reclassification.

## GSD delivery record

The required `discuss-phase`, `plan-phase --tdd`, `execute-phase`, `verify-work`, and `code-review` stages were executed inline. The canonical delivery contract forbids role spawning in this context, so `inline_manual_fallback` is recorded in `RUN-STATE.json`. Required Go skills were loaded: `golang-how-to`, `golang-cli`, `golang-testing`, `golang-error-handling`, `golang-design-patterns`, `golang-structs-interfaces`, `golang-security`, and `golang-safety`.

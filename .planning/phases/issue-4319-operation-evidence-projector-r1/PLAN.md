# Issue #4319 — operation evidence projector plan

## GSD and skills

- Commands resolved and performed inline: `discuss-phase`, `plan-phase --tdd`, `execute-phase`, `verify-work`, and `code-review`.
- Required skills loaded: `golang-how-to`, `golang-cli`, `golang-testing`, `golang-design-patterns`, `golang-structs-interfaces`, `golang-error-handling`, `golang-security`, `golang-safety`, and `golang-documentation`.
- CLI parity: this adds a `connectorgen` developer command, not a `pm` user command. Update its runtime help and test it; `pm` help/manual/website pages are not applicable. The evidence artifact's website field is generated data from connector projection, not a website UI change.

## Plan

1. Add red behavioral tests for a source-locked connector definition, inspecting emitted artifact rows rather than internal declaration counts. Include complete evidence, each individual missing surface, provider-evidenced absence, duplicate rollups, deterministic generation, and the 100-row aggregate validator.
2. Introduce a small projector package/command seam that reads existing source-lock/projection inputs without modifying their parser or schema. Map one source operation into a stable evidence row with complete/explicit-gap semantics and a fixed-sort output artifact plus deduplicated foundation rollup.
3. Integrate the command in `connectorgen` help/dispatch and generate a checked-in fixed-100 reference artifact from the external provider source inventory. The validator compares projector output to that independently selected cohort and names each failure.
4. Run targeted Go tests, generation/check commands, and the full repository `make verify`. Record byte-stability and magnitude checks; investigate an unexpected result rather than relabel it as complete.
5. Execute the inline `verify-work` and `code-review` passes, document any findings/dispositions, commit, push, open a `main`-targeted PR, and read its base through the GitHub API.

## Magnitude check

- The current `main` tree contains the GitHub source-lock v2 inventory only,
  so this projector emits **1,525** rows and five deduplicated gap rollups.
  This is not accepted as a replacement count for Batch 8–10's 11,128: the
  remote batch ledger's remaining 9,599 source operations plus four
  provider-evidenced unavailable/dynamic surfaces arrive with the
  concurrent v3 source-lock foundation. The projector reads v2 and v3 without
  changing that owner’s parser or schema, so it expands when those locks land.

## Design constraints

- Treat source locks as read-only inputs; source lock schema/provenance is owned by the concurrent source-lock foundation.
- A source operation has one row even when it has multiple classifications. Stable source identity and sorted serialization make duplicate handling deterministic.
- Do not treat an absent generated surface as not applicable. Only provider-owned source evidence may establish an absence, and the artifact retains that evidence pointer.
- The gate must compare actual source-derived rows to an independent fixed-100 cohort, never a count created by the projector itself.

# PLAN — issue #3852 output-policy declaration

Issue: #3852. Branch: `fm/cli-found-output-policy-declaration-r1`.

## GSD path

- `scripts/gsd doctor`: passed.
- `scripts/gsd list`: passed; the installed adapter contains the five required lifecycle commands.
- `scripts/gsd sources discuss-phase|plan-phase|execute-phase|verify-work|code-review`: passed.
- `go run ./cmd/agentcontractgen check`: passed.
- Discuss prompt: `scripts/gsd prompt discuss-phase issue-3852-output-policy-declaration-r1 --auto`.
- Plan prompt: `scripts/gsd prompt plan-phase issue-3852-output-policy-declaration-r1 --tdd --skip-research`.
- Execute prompt: `scripts/gsd prompt execute-phase issue-3852-output-policy-declaration-r1`; the
  three TDD slices were executed inline.
- Verify prompt: `scripts/gsd prompt verify-work issue-3852-output-policy-declaration-r1`; automated
  UAT is recorded in `UAT.md`.
- Review prompt: `scripts/gsd prompt code-review issue-3852-output-policy-declaration-r1
  --depth=standard --files=...`; its inline standard review is recorded in `REVIEW.md`.
- Inline/manual fallback is required because the issue contract forbids role spawning.

## Required skills loaded

- `golang-how-to` — selected the Go implementation/test/documentation skills.
- `golang-design-patterns`, `golang-structs-interfaces` — smallest explicit policy-set model.
- `golang-error-handling`, `golang-security`, `golang-safety` — preserve closed validation,
  bounded behavior, and no-secret/no-raw-output safety boundaries.
- `golang-testing` — red/green observable schema and runtime tests.
- `golang-cli`, `golang-documentation` — connector-surface and authoring-documentation parity.
- `github-issue-first-delivery` — issue-first branch, TDD, and evidence record.

## Deliverable slices

### Slice 1 — prove the declaration/runtime mismatch (TDD RED)

1. Add an engine bundle-load test with one `rest_write` operation and one implemented
   `direct_write` command declaring `output_policy: "json"`.
2. In the same test, prove the current direct-write policy validator accepts `json` and its response
   executor returns the decoded body unchanged.
3. Run the focused engine test. It must fail only because `cli_surface.schema.json` rejects `json`.
   Record the real failing output in `TDD-LEDGER.md`; retain the test.

### Slice 2 — reconcile the closed enum and prevent drift (TDD GREEN)

1. Represent the existing direct-read and direct-write supported policy sets in enumerable,
   behavior-preserving runtime registries outside #3771-owned functions.
2. Extend `cli_surface.schema.json` with the exact union of those two sets, retaining the existing
   binary-download compatibility value `binary_file_bounded`.
3. Add a command-runner test that parses the schema enum and fails on either direction of drift.
4. Run focused engine and command-runner tests, then `connectorgen surface-sync --check` and
   `connectorgen validate` to prove existing bundles remain valid.

### Slice 3 — authoring guidance and verification

1. Replace the redaction-default wording in the direct-read/output-policy convention with concise
   policy selection guidance: `json` for complete write JSON, `none` for intentionally no body,
   special policies only for their existing response families, and binary download unchanged.
2. Confirm the authoring guide does not prescribe a redacting policy as the default.
3. Run targeted tests, formatting, vet, schema/connectorgen gates, docs check, boundary check, and
   the issue-appropriate non-monolithic verification components. Record outputs in
   `VERIFICATION.md`.

## Scope fences

- Do not alter #3771-owned `BuildWriteCommand`, `Run`, `runDirectRead`,
  `runOperationDirectRead`, `runBinaryDownload`, or any redaction helper.
- Do not modify direct-write or direct-read behavior, operations schema, existing connector JSON
  bundles, generated docs, runtime help, public CLI commands, credentials, or provider state.
- Do not add dependencies, raw response bypasses, or new output-policy names.

## Commit checkpoints

1. Planning/context/TDD ledger checkpoint.
2. Retained RED test checkpoint with failing output recorded.
3. GREEN schema/regression-test/documentation checkpoint after focused verification.
4. Verification/review fixes checkpoint if needed.

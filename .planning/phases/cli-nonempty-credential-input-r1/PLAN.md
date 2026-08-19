# Plan — provider-neutral non-empty credential input

**Status:** in progress

## Task Delivery Header

- Issue: No standalone GitHub issue was supplied by the Firstmate foundation brief; Twenty consumer PR is #4298 (`feat(twenty): recover CRM all-ops connector`).
- Base branch: `main` (`origin/main` at `51dd6d468e4a40ece70c36efb81df4fdede8a8b6`).
- Merges into: `main`.
- Delivery: Branch `fm/cli-nonempty-credential-input-r1` committed and ready for the normal Firstmate no-mistakes validation/PR gate; no push or merge in this task stage.
- Working branch: `fm/cli-nonempty-credential-input-r1`.
- Task: Prevent empty connector secret values from being persisted or emitted, while preserving stdin byte fidelity except for one documented terminal line-ending transport delimiter. Do not edit the Twenty lane.
- Verification: Focused red/green tests for CLI stdin, App/vault round trips, and engine/connsdk auth emission; `gofmt`, changed-package tests, full repository checks, connector boundary, generated-artifact checks, `go vet`, and no-mistakes at the subsequent gate.

## Evidence Table

| Acceptance criterion | Evidence | Observable assertion or fake reason |
| --- | --- | --- |
| Non-empty stdin values persist and round-trip through App/vault | live | CLI-created credential resolves from a fresh App; byte length and SHA-256 fingerprint match the synthetic input and vault ciphertext lacks its fingerprinted plaintext assertion path. |
| Long stdin values avoid argv and round-trip exactly | live | A generated long canary reaches a fresh App with matching length and SHA-256 fingerprint; test argv carries only the field name. |
| Empty and newline-only stdin fail before persistence | live | CLI exits validation with the typed non-secret category and no credential metadata or encrypted entry is created. |
| Only one terminal LF/CRLF is normalized | live | Table cases assert expected byte lengths/SHA-256 fingerprints and preserve extra terminal bytes. |
| Direct App callers cannot persist an empty secret | live | `AddCredential` returns the typed empty-secret error and leaves no credential/vault entry. |
| Existing non-empty credentials and repeated writes remain usable | live | Existing stored credential resolves across a fresh App; rotation/rewrite remains non-empty and retains a matching fingerprint. |
| Selected required authentication cannot emit an empty credential form | live | Bearer, Basic, API-key-header, API-key-query, and credential-grant paths fail before request mutation; no header/query credential is present. |
| Optional no-auth selection remains supported | live | A declaration whose optional credential route does not match selects `none` and sends no Authorization header. |
| Twenty lane is isolated | live | Changed-path check shows no `internal/connectors/defs/twenty/**` path; consumer commit SHA is recorded for PR #4298. |

## Lifecycle and skills

- Resolved adapter commands: `scripts/gsd doctor`, `sources` for
  `discuss-phase`, `plan-phase`, `execute-phase`, `verify-work`, and
  `code-review`; generated prompts were inspected and performed inline.
- Required skills: `golang-how-to`, `golang-cli`, `golang-testing`,
  `golang-error-handling`, `golang-security`, `golang-safety`,
  `golang-design-patterns`, and `golang-structs-interfaces`.
- CLI help/manual/website parity: behavior is behind the existing documented
  `--value-stdin` contract; help synopsis, flag name, docs, and website command
  surfaces do not change. We will verify `pm help credentials`,
  `pm credentials`, and `pm credentials --help`, and record the deliberate
  no-doc-change result.

## Red-green-refactor slices

1. **Red — real CLI/App stdin path.** Add synthetic stdin cases for empty,
   newline-only, LF/CRLF, byte-preserving terminal content, and long values.
   Confirm the current broad trim or empty value can reach persistence.
2. **Green — shared input and persistence contract.** Add a small provider-
   neutral contract package/type; normalize one terminal delimiter in CLI and
   reject emptiness again in `App.AddCredential` before the vault write.
3. **Red — selected auth emission.** Add table tests showing an empty selected
   bearer/basic/API-key route currently reaches header/query construction.
4. **Green — auth admission guard.** Make each selected credential-bearing
   declarative route refuse blank resolved material before it can mutate a
   request, while retaining non-selected optional/no-auth behavior.
5. **Refactor and verify.** Consolidate duplicated test helpers, preserve
   redaction, run generated and repository gates, complete `VERIFICATION.md`,
   and perform the inline code review.

## Commit checkpoints

1. Planning/TDD evidence checkpoint before production edits.
2. Red test checkpoint, retained in the ledger even if not separately committed.
3. Green implementation and focused-test checkpoint.
4. Review/verification fix checkpoint if required.

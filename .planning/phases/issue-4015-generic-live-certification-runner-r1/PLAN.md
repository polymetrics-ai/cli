# Generic live certification runner — plan

## Task Delivery Header

- Issue: Refs #4015 — Production MVP
- Base branch: integration/4015-mvp-flat-r1 (`cc2bfe6a2252e94ad27630b7fa4857a4d5b07d6e` fetched and rebased before the full-sweep extension)
- Merges into: integration/4015-mvp-flat-r1 → main
- Delivery: A direct PR is open against the exact base after the runner has persisted only individually validated live evidence, with its local verification results recorded.
- Working branch: fm/cli-github-certify-now-r1
- Task: Add one connector-parameterized Node script. It derives commands, assertions, credential defaults/requirements, eligibility, and command paths from the selected connector's `metadata.json`, `cli_surface.json`, `certification.json`, and optional `certification-sweep.json`; it runs one operation at a time, writes an accepted evidence record only after an observed response satisfies the declared assertion, and validates the repository immediately after each record. It does not modify Go production code or hard-code connector names, provider paths, authentication conventions, or response shapes.
- Verification: Run Node syntax and definition-shape checks, execute the runner against the authorized connector, execute the unchanged runner against a differently shaped connector, and run `go run ./cmd/connectorgen certification-matrix --check` after every accepted record and at the end.

## Evidence Table

| Acceptance criterion | Evidence | Observable assertion or fake reason |
| --- | --- | --- |
| Runner accepts a connector argument and only loads its own definition directory | live | A second connector invocation reaches its independently declared candidate set without a script edit or connector-specific branch. |
| A passing command produces a valid, secret-free evidence record immediately | live | The operation's declared JSON assertions hold; the record passes `connectorgen certification-matrix --check` before the next operation begins. |
| A command with no produced value is not certified | live | The runner records a non-pass receipt and leaves the accepted-evidence directory unchanged for that operation. |
| Credential scope remains bounded | live | Every accepted record uses `observed_operations`, its exact v2 note, and `protocol_exchanges`; no record asserts full parity. |
| Sensitive values remain out of the repository | live | The runner fingerprints every scalar in command, request, and response proof material and scans written records for the provided credential before persistence succeeds. |

## Scope and delivery guard

Captain direction on 2026-08-17 explicitly widened the runner from one provider to every connector. This is a script-only certification foundation: connector-specific input remains in each definition bundle, and the runner has no connector-specific conditionals. The direct-PR task forbids spawned roles, so the required GSD lifecycle is executed inline and recorded here.

## Slice plan

1. **Plan and red proof.** Record the missing generic runner and verify no connector-specific command runner exists.
2. **Green runner.** Add one Node script that constructs candidates strictly from the selected definition bundle, captures only in-memory raw output, fingerprints it before persistence, and validates after every accepted record.
3. **Real provider proof.** Run the selected connector with the disposable credential. Certify each passing operation before attempting another; classify every non-pass with a sanitized provider receipt.
4. **Portability proof and review.** Run the unchanged script with a second connector argument, complete matrix/check and targeted validations, then perform inline code review.
5. **Direct never-stall full sweep.** Firstmate decided that `scripts/certify-connector-live.mjs` remains untouched because its declared-direct-read candidate model is deliberately narrower than the captain's full-surface direction. Execute each selected command directly instead: inspect the bundle for command/API shape, invoke the built binary with a point-in-time Keychain credential, assert the observed response, write one schema-v2 accepted-evidence JSON record, and immediately run `certification-matrix --check`. A derived assertion is labelled `agent_derived` in strict-schema evidence provenance; unsupported commands are recorded as `not_implemented`, never executed.
6. **Bounded live batches.** Resume serially in batches of at most 100 from individual local receipts. Missing fixture inputs, credential mismatch, entitlement, unassessed mutation containment, and provider defects become an immediately persisted non-pass. A provider mutation used solely as fixture setup is directly deleted through the provider API and independently read back before the next command; the CLI delete exit code is never cleanup proof.

## Executed evidence

- The authorized live sweep executed all 122 definition-owned `eligible_pending_live` candidates: 38 accepted records were written under `internal/connectors/certifications/evidence/`, 80 provider refusals and 4 missing-fixture non-passes were written immediately to `GITHUB-LIVE-RUN-ACCOUNTED.json`, and no corrected-sweep product defect occurred.
- Each of the 38 record writes was immediately followed by a passing `go run ./cmd/connectorgen certification-matrix --check`; the same check passed again at the end.
- The unchanged runner recorded Freshchat's definition-owned no-candidate outcome (`executed=0`, `missing_fixture=1`), and its definition-only path passed for all 36 connector command surfaces. Connectors without `certification.json` now produce a definition-owned missing-fixture receipt instead of an invocation error.
- After PR #4219 opened, a data-selected GitHub App retry excluded every enterprise path. It executed the 16 rows whose captain-supplied direct probe reported `new_status=200`, certified 15, and recorded one HTTP 400 product defect without retrying it.

## Required skills and lifecycle

- Loaded: `golang-how-to`, `golang-cli`, `golang-design-patterns`, `golang-structs-interfaces`, `golang-error-handling`, `golang-security`, `golang-safety`, and `golang-testing`.
- Resolved and reviewed inline: `scripts/gsd prompt discuss-phase issue-3993-github-live-certification`, `plan-phase ... --tdd`, `execute-phase ...`, `verify-work ...`, and `code-review ...`; compatible isolated GSD roles are unavailable/forbidden by this direct-PR brief.
- CLI help/manual/website parity: not applicable; this adds no `pm` command, flag, help topic, generated manual, or website content.

## Safety constraints

- The runner never prints, writes, or includes a credential in argv. The credential is supplied only through an inherited environment variable set by command substitution outside the script.
- It never writes outside the selected repository's local project or `internal/connectors/certifications/evidence/`; it performs no provider mutation unless a definition-declared candidate explicitly requires it and the caller opts in.
- Provider output is never persisted raw: every scalar is replaced by a repository-salted HMAC fingerprint before it reaches an evidence record.
- Captain standing directive, 2026-08-18: the full 1,571-command sweep is the priority. One bounded retry is permitted per obstacle, then the runner records the honest outcome and advances. Direct provider fixture setup/cleanup inside the disposable boundary, agent-derived assertions, and a credential retry are authorised. Only real money, real people, public visibility under the disposable organisation, or a third-party repository/organisation is an escalation; unclassified mutations become `unassessed` and do not stall the batch.

## Slice 1 continuation — manual GSD/TDD fallback

This direct-PR evidence-only continuation changes `internal/connectors/certifications/evidence/` but does not change production behavior. The project lifecycle is recorded inline because the task requires real serial provider execution and the direct-PR brief forbids spawned GSD roles.

- Scope: after PR #4219 merged, rebase onto `origin/integration/4015-mvp-flat-r1`; add only fresh GitHub schema-v2 records backed by external protocol proofs; execute the remaining non-mutating Slice 1 direct-read, binary-download, and ETL paths one at a time; do not run paused direct-write or reverse-ETL commands.
- Red: the accepted importer rejected the completed direct-read cohort even though 36 operations passed, because 84 legitimate non-passes made the aggregate report fail. It wrote zero records (`importer_all_or_nothing_cohort`, `36 passed / 84 legitimate non-pass / 0 records written`).
- Green: rerun each passing stage with its own external proof, write the 36 bounded `observed_operations` schema-v2 records, and immediately run `go run ./cmd/connectorgen certification-matrix --check` after each. Record every later non-pass with GitHub's response; leave a successful command uncertified when it does not emit a captured protocol exchange rather than inventing one.
- Verification and review: the exact final local gates and results are in `VERIFICATION.md`; `bash scripts/verify-gsd-workflow origin/integration/4015-mvp-flat-r1` must pass before pushing the direct PR.

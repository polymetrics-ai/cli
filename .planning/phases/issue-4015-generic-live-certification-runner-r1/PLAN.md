# Generic live certification runner — plan

## Task Delivery Header

- Issue: Refs #4015 — Production MVP
- Base branch: integration/4015-mvp-flat-r1 (`9d01ab98a86d0d78cac04c579d548778f8197674` fetched before work)
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

## Executed evidence

- The authorized live sweep executed all 122 definition-owned `eligible_pending_live` candidates: 38 accepted records were written under `internal/connectors/certifications/evidence/`, 80 provider refusals and 4 missing-fixture non-passes were written immediately to `GITHUB-LIVE-RUN-ACCOUNTED.json`, and no corrected-sweep product defect occurred.
- Each of the 38 record writes was immediately followed by a passing `go run ./cmd/connectorgen certification-matrix --check`; the same check passed again at the end.
- The unchanged runner recorded Freshchat's definition-owned no-candidate outcome (`executed=0`, `missing_fixture=1`), and its definition-only path passed for all 36 connector command surfaces. Connectors without `certification.json` now produce a definition-owned missing-fixture receipt instead of an invocation error.

## Required skills and lifecycle

- Loaded: `golang-how-to`, `golang-cli`, `golang-design-patterns`, `golang-structs-interfaces`, `golang-error-handling`, `golang-security`, `golang-safety`, and `golang-testing`.
- Resolved and reviewed inline: `scripts/gsd prompt discuss-phase issue-3993-github-live-certification`, `plan-phase ... --tdd`, `execute-phase ...`, `verify-work ...`, and `code-review ...`; compatible isolated GSD roles are unavailable/forbidden by this direct-PR brief.
- CLI help/manual/website parity: not applicable; this adds no `pm` command, flag, help topic, generated manual, or website content.

## Safety constraints

- The runner never prints, writes, or includes a credential in argv. The credential is supplied only through an inherited environment variable set by command substitution outside the script.
- It never writes outside the selected repository's local project or `internal/connectors/certifications/evidence/`; it performs no provider mutation unless a definition-declared candidate explicitly requires it and the caller opts in.
- Provider output is never persisted raw: every scalar is replaced by a repository-salted HMAC fingerprint before it reaches an evidence record.

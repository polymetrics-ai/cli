## Task Delivery Header

- Issue: Refs #4015 — GitHub live certification coverage.
- Base branch: fm/cli-github-certify-now-r1 (PR #4219).
- Merges into: fm/cli-github-certify-now-r1 → integration/4015-mvp-flat-r1 → main.
- Delivery: A stacked PR containing schema-validated live evidence and honest receipts for every command in the assigned actions-security slice.
- Working branch: fm/cli-cert-slice2-actions-security.
- Task: Certify the 301-command GitHub actions-security slice, retaining a validated record for every live pass and a bounded receipt for every honest non-pass; remove all run-owned credentials and fixtures.
- Verification: `go run ./cmd/connectorgen certification-matrix --check`, scoped live receipts, `git diff --check`, and the repository gates applicable to evidence-only changes.

## Evidence Table

| Acceptance criterion | Evidence | Observable assertion or fake reason |
| --- | --- | --- |
| A live direct read produces schema-valid evidence | live | The GitHub response satisfies its declared JSON-pointer assertion, and `certification-matrix --check` accepts the new record; an empty or wrong-shaped response is rejected. |
| Non-passes are distinguished by observed provider/runtime result | live | Each receipt records the command, exact bounded diagnostic, and one of the task taxonomy outcomes rather than treating a successful process exit as a pass. |
| Run-owned credentials are removed | live | Every completed receipt reports `local_cleanup.credential_removed: true`; a missing cleanup leaves it false. |

## Inline GSD / TDD Record

The direct-PR lane cannot use compatible isolated GSD agents, so the lifecycle is executed inline.

- `discuss-phase 4015`: scope is the supplied actions-security slice; do not regenerate shared artifacts.
- `plan-phase 4015 --tdd`: direct-read candidates first, then fixture-dependent reads and contained mutations with plan → preview → approval → execute.
- Red: the first direct-read pass attempted without non-secret `owner`/`repo` credential config failed with `missing path variable "owner"`; generated evidence was not accepted for those calls.
- Green: retrying with the declared disposable fixture owner/repo and resolving the real configuration id `17` produced 11 accepted records, each immediately accepted by `certification-matrix --check`.
- `execute-phase 4015`: live command execution is represented by the receipts in this directory.
- `verify-work 4015`: final receipt aggregation and matrix validation remain required before delivery.
- `code-review 4015`: review the committed evidence and receipts for secret exposure, scope drift, and false success claims before opening the PR.

Loaded skills: `golang-how-to`, `golang-cli`, `golang-testing`, `golang-error-handling`, `golang-security`, and `golang-safety`.

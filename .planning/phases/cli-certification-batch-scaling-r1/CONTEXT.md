# Certification batch-scaling measurement — context

## Task Delivery Header

- Issue: Refs #4015 — Production MVP certification; operationally blocks publication under #4211.
- Base branch: `integration/4015-mvp-flat-r1` at `eba2658c5fd671a1eebfb71463cbe6a3045d3c65` (fetched before planning).
- Merges into: `integration/4015-mvp-flat-r1` → `main`.
- Delivery: A direct PR is open against the stated base, with measured 10/100/repeated-100 live-read curves, staged non-published evidence inputs, verification, and an API-read-back of the PR base.
- Working branch: `fm/cli-certification-batch-scaling-r1`.
- Task: Measure end-to-end serial GitHub direct-read certification throughput, including a fresh project setup, credential validation, checkpoint I/O, report persistence, and teardown. Classify every execution as produced-value pass, provider refusal, missing fixture, or product defect; record rate-limit/header evidence; project the remaining read surface only from those measurements; do not publish evidence while #4211 is unresolved.
- Verification: Current-head `pm` help; current #4214 candidate-projection check; live 10/100/repeated-100 runs; structural checks that every successful row asserted `/response` as object or array; resume checkpoint inspection; generated checks; changed-package/consumer tests including `./cmd/connectorgen`; repository verification gates; diff review; and GitHub API base read-back.

## Evidence Table

| Acceptance criterion | Evidence | Observable assertion or fake reason |
| --- | --- | --- |
| Ten-operation end-to-end cost is measured | live | A fresh disposable project executes exactly ten declaration-derived direct reads; the report has ten terminal stage rows, and each pass contains a produced `/response` object or array assertion. |
| Hundred-operation end-to-end cost is measured | live | A fresh disposable project executes exactly 100 direct reads from the pinned #4214 candidate projection; result buckets sum to 100. |
| Repeated 100-operation behavior is measured | live | At least three fresh, serial 100-operation batches execute with the same deterministic candidate manifest; the comparison includes wall clock, mean, rate events, and checkpoint counts. |
| Throttling conclusion has header evidence | live | Saved sanitized run facts include real rate-limit event type/reason/reset fields. If no throttle occurs, the report explicitly bounds that conclusion to the measured operations and records the absence of Retry-After/HTTP-429 evidence. |
| Evidence is ready but publication is honest | live | A staged generic live-report input contains only sanitized execution facts and no `credential_scope`; no accepted evidence file or certification matrix is written before #4211 is verified. |
| Reusable run rules are captured | live | The final report contains explicit rules for serial batching, checkpoint/resume, Retry-After handling, and produced-value classification. |

## Decisions

- Source candidates come from PR #4214 head `7306b9ec3e079b51ac9c70a674605a3a27f6e09b`, because it remains open. Its source is used only in a disposable, uncommitted measurement copy and is not merged, copied into production code, or rebased into this branch.
- The 10-operation manifest contains the first ten generated direct reads in source order. The 100-operation manifest contains all 97 generated direct reads plus the first three existing direct-read overrides. This is deterministic and uses direct reads only.
- Repeated hundred-operation batches reuse the exact 100-operation manifest. Reads are idempotent; using one stable workload separates load effects from changing fixture mix.
- Each run uses a fresh disposable `pm` project and runs serially. The timing starts before project initialization and ends only after the project is removed. The committed report retains counts, timing, stage names, and safe rate-limit metadata, never credentials or provider response bodies.
- #4211 / PR #4215 is a hard publishing gate. This lane may stage non-accepted evidence inputs but must never create an accepted record or claim `credential_scope: full_parity`.

## GSD and skills record

Resolved lifecycle: `scripts/gsd doctor`; `scripts/gsd sources discuss-phase`; `plan-phase`; `execute-phase`; `verify-work`; `code-review`; corresponding generated prompts; and `go run ./cmd/agentcontractgen check` before production work.

Inline/manual fallback: this direct-PR task has no compatible isolated Pi-role runtime, and the delivery contract prohibits role spawning. Discussion, planning, measurement execution, verification, and review are therefore recorded inline.

Skills loaded: `golang-how-to`, `golang-cli`, `golang-testing`, `golang-error-handling`, `golang-security`, and `golang-safety`.

CLI/docs parity: `pm connectors certify` behavior is not changed. Its existing help was read before the live operation; no shipped help, manual, website, or generated surface changes are expected.

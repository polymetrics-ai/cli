---
phase: "3993"
depth: standard
mode: inline_manual
status: historical_harness_pass_and_fresh_lineage_1_review_complete
---

# Code review — Issue #3993 GitHub live certification harness

`scripts/gsd prompt code-review 3993 --depth=standard` was generated on
2026-08-11. The GSD adapter cannot run a roadmap phase for issue 3993 and the
delivery contract/dispatch forbid spawning the reviewer role, so this standard
review was completed inline after rebase and focused validation.

## Scope reviewed

- Boundary resolution and current-ledger generation in
  `scripts/github-live-cases.mjs`.
- Barrier release, terminal process lifetime, report redaction, write/read-back
  validation, and App credential scope validation in
  `scripts/github-live-proof-sweep.mjs`.
- Source-derived manifest/bootstrap guards and their generated JSON artifacts.
- Focused Node regression coverage, UAT, verification, and live-measurement
  evidence.

## Findings and disposition

| Severity | Finding | Disposition |
| --- | --- | --- |
| Critical | None. | No secret, destructive, boundary, or report-integrity Critical finding remains. |
| Warning | [#4020](https://github.com/polymetrics-ai/cli/issues/4020): a forged write case could set `--config owner=…` / `repo=…`, bypassing the former literal `--owner`/`--repo` check and reach credential inspection. | Fixed in this child with a RED/GREEN regression. The guard now covers direct owner/repository aliases and their config forms before a PM child starts; focused and full harness tests are green. |
| Info | The one-barrier 665-child run still reaches the 45,000 ms local terminal bound before meaningful provider quota use. | Retained exactly as measured: isolated App returned-record controls succeed; REST changed 15,000 → 14,997 and GraphQL stayed 5,000. No #3990 coordinator was added or implied. |
| Info | GitHub certification still needs shared provider-admission policy and outbound foundations. | Explicit dependency only: #3990 for coordinated admission, #3994 for approved action execution, and #3992 for schedule firing. This review does not implement any of them. |

## Review conclusions

- Boundary input is default-deny, normalizes slugs, binds organization and
  repository immutable IDs, and refuses the protected owner before a provider
  process is constructed.
- Every eligible operation is queued before the single shared barrier is
  released. The deliberately uncapped release remains measurement semantics,
  not a covert sequential retry policy.
- A read is proved only by a returned provider envelope. A write is unavailable
  unless it has an independent non-write read-back, and the final scope guard
  rejects direct/config owner and repository overrides before process launch.
- Report projection discards raw stdout, stderr, response bodies, approvals,
  grants, and credentials. Generated evidence passed the credential-shaped scan
  and the live temporary directory is absent.
- Manifest and bootstrap self-checks are generated from the current 1,521-row
  implemented surface rather than the stale 957-row pre-skip ledger.
- The legacy rate-limit artifact is historical-only, retains its archived
  case-file binding, and explicitly states that current certification is not
  proven and `rate-limit get` is untestable.
- Bootstrap inventory input, completed output, and supplied inventory are
  guarded for credential-shaped persisted material before hashing, validation,
  serialization, or output.

## Final-gate correction history

`1/#4020 -> 2/#4022 -> 3/#4027 -> 4/#4039 -> 5/#4050` completed after the
focused 50-test GREEN suite. Round 5 ([#4050](https://github.com/polymetrics-ai/cli/issues/4050)) retired the legacy
rate proof's current-certification semantics and guarded the remaining
bootstrap artifact path. The historical 0/665/856 result remains separate from
the current 182/1,339 classifier.

## Verdict

**PASS for the #3993 harness slice.** All five permitted in-scope correction
loops are recorded complete. GitHub as a provider remains **uncertified**
because the truthful 0/665 measurement and the external dependencies are
unresolved.

## Fresh current-SHA lineage 1/5 — standard review

This is a separate captain-authorized delivery lineage, not a sixth correction
round. `scripts/gsd prompt code-review 3993 --depth=standard` was regenerated
on 2026-08-11 and reviewed inline because the issue-backed GSD phase has no
compatible isolated reviewer runtime and the delivery contract forbids role
spawning.

### Scope

- `scripts/github-live-proof-sweep.mjs`
- `scripts/tests/github-live-proof-sweep.test.mjs`
- Fresh lineage plan, RED/GREEN ledger, run state, UAT, summary, and
  verification evidence.

### Findings

| Severity | Finding | Disposition |
| --- | --- | --- |
| Critical | None. | The external proof runner cannot inspect a credential, start a `pm` child, or reach a provider after it identifies itself as `external_pm_per_operation`. |
| Warning | None. | Both report statuses require a closed, safe execution-model enum; neither can be relabelled across the credentialed-live/blocker boundary. |
| Info | Existing schema-version-2 reports without `execution_model` now fail validation. | Intentional fail-closed migration: a historical or external-process report cannot be promoted as current-SHA one-process certification evidence. |

### Review evidence and reasoning

- JavaScript syntax checks and `git diff --check` pass for the production and
  regression-test change.
- `execution_model` is accepted only from the two fixed safe strings. A
  `credentialed_live` report must be `built_pm_in_process`; an
  `external_blocker` report must be `external_pm_per_operation`.
- In `executeLive`, the external-runner assertion is immediately after
  canonical boundary/case validation and before `readFile(binary)`,
  `validateCredentialScope`, `executeBarrier`, and every `runProcess` call.
  The review therefore found no path from this runner to a credentialed
  provider request under a false in-process provenance claim.
- The focused regression proves both forbidden relabellings are rejected, and
  the existing structural-redaction assertion avoids scanner-prone serialized
  URL matching while retaining exact redaction coverage.
- Removing the unused `buildCases` import addresses the outstanding static
  scan finding without changing the canonical case builder used by the test.

### Fresh-line verdict

**PASS for the bounded provenance recovery.** It does not certify GitHub live
behavior: the approved App credential and immutable run-owned boundary are
still absent, so the live slice remains `needs-decision` with no provider
operation performed.

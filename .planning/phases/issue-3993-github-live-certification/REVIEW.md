---
phase: "3993"
depth: standard
mode: inline_manual
status: clean_after_one_in_scope_correction
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

## Verdict

**PASS for the #3993 harness slice.** One local boundary defect was corrected
in correction loop 1 of 5. GitHub as a provider remains **uncertified** because
the truthful 0/665 measurement and the external dependencies are unresolved.

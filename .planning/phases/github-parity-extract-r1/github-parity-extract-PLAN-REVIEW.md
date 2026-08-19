# Plan Review — GitHub live-operation proof

**Date:** 2026-08-08
**Mode:** Inline/manual plan-checker fallback

The GSD adapter is healthy, but this session has no compatible isolated planner/checker and is not
authorised to delegate. I therefore reviewed the planned work against the generated GSD plan-phase
contract inline.

## Checks

| Check | Result | Evidence |
| --- | --- | --- |
| Context decisions covered | PASS | `gsd-sdk query check.decision-coverage-plan ...` reports 13/13. |
| TDD before behavior | PASS | Plans 01 and 02 define real RED commands before their declaration/harness GREEN work. |
| Dependency order | PASS | Rate declaration precedes sweep; live proof consumes both. |
| Scope | PASS | GitHub-only changes; non-GitHub phantom-flag work is a reported count only. |
| Security | PASS | Scope inputs are non-secret; runner persists redacted records only; writes require the dedicated private repository. |
| Gate completeness | PASS | Generated artifact confinement, CLI/help parity, provider rate proof, destructive gate truthfulness, and final zero-failure accounting are explicit acceptance criteria. |

## Revisions made during review

- Added front-matter `must_haves` rather than relying on prose, because the decision-coverage tool
  only reads plan front matter.
- Kept the phantom-flag regression test GitHub-scoped. The other-connector debt is counted and
  recorded, never made a repository-wide blocker.

## Approval

The plan is ready for inline execution. Any implementation change that expands beyond the
GitHub-only boundary, creates a second limiter, or weakens write confirmation requires a new
captain decision rather than a plan adjustment.

---

## Plan 04 review — response/body parity foundation

**Date:** 2026-08-08
**Mode:** Inline/manual plan-checker fallback

| Check | Result | Evidence |
| --- | --- | --- |
| Context decisions covered | PASS | `gsd-sdk query check.decision-coverage-plan .planning/phases/github-parity-extract-r1 .planning/phases/github-parity-extract-r1/github-parity-extract-CONTEXT.md` reports 15/15. |
| TDD order | PASS | Plan 04 requires focused no-content/text/raw-body RED tests before each matching engine or transport change. |
| Transport scope | PASS | The plan admits only operation-declared `POST text/plain` root-string input; it does not add generic raw HTTP. |
| Provider-auth safety | PASS | OAuth application endpoints are deliberately not promoted by this response slice; D-15 prohibits falling back to the ordinary bearer credential. |
| Generated-artifact ownership | PASS | The generator source changes first, then emits GitHub declarations and `surface-sync` verifies them. |
| Delivery scope | PASS | The work remains within the existing GitHub parent lane and explicitly does not claim final parity while further classified gaps remain. |

**Approval:** Plan 04 is ready for inline TDD execution. The next implementation checkpoint must
record actual test failures and green results in `TDD-LEDGER.md` before any generated GitHub
surface is promoted.

---

## Plan 05 review — oneOf request-body parity

**Date:** 2026-08-08
**Mode:** Inline/manual plan-checker fallback

| Check | Result | Evidence |
|---|---|---|
| Context decisions covered | PASS | Plan 05 carries D-13, D-14, D-15 exclusion, and D-16's exact 19-arm scope. |
| TDD order | PASS | The first task adds a loaded-bundle/preflight test that must fail before generator changes. |
| Runtime boundary | PASS | `covered_by.writes` already supports multiple actions per endpoint. The plan adds only declared `record.*` structured JSON parsing because the existing CLI schema already exposes the type; no generic union, raw body, or transport feature is proposed. |
| Safety classification | PASS | Bulk attestation deletion is destructive despite its POST verb; all documented creation arms remain approval-only. |
| Generated-artifact ownership | PASS | Only `scripts/gen-github-parity.py` is edited as the declaration source; GitHub JSON and the shared ledger are regenerated. |
| Scope restraint | PASS | OAuth Basic-auth, anyOf, root polymorphism, genuine duplicates, credential minting, and live writes are explicitly deferred. Structured JSON stays within the captain-authorized request-body work and is schema-gated before planning. |

**Approval:** Plan 05 is ready for inline RED/GREEN execution. The ledger must record the actual
red command result before any GitHub declaration is regenerated.

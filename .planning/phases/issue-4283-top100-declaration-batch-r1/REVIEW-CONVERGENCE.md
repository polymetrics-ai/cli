# Firstmate exhaustive review convergence — Batch-1 source-rigidity R2

## Immutable discovery frame

- Candidate SHA: `6ec34964fda7a78a1736a5cd8933a46418346261`.
- Candidate ref: `origin/fm/cli-top100-declaration-batch-r1`.
- Required ancestor: `6ec34964fda7a78a1736a5cd8933a46418346261` (verified with `git merge-base --is-ancestor`).
- Existing PR/base: #4294, `fm/cli-top100-declaration-batch-r1 → main`; GitHub API read-back recorded in `CONTEXT.md`.
- Captain review substitution: requested Codex-only fresh-context audit is recorded as the intended audit route. No audit verdict is claimed before there is a valid scoped candidate.

## Discovery ledger

The committed machine-checkable discovery input is
`data/connector-canon/batch1-source-rigidity-r2-cohort-ledger.json`. It pins
each cohort's source-lock path, whole-file SHA-256, schema form, identity
location, and source denominator. It intentionally has zero projected cells:
it is the immutable denominator for the later source-operation-to-lane report,
not a coverage or usability claim.

| Cohort | Source identities | Requested classification/reachability state |
| --- | ---: | --- |
| Asana | 249 | Pending cohort matrix and fresh review. |
| Bitbucket | 297 | Pending cohort matrix; retain default-only evidence gaps visibly. |
| CircleCI | 111 | Pending cohort matrix and fresh review. |
| Docker Hub | 54 | Pending cohort matrix and fresh review. |
| GitLab | 1,752 | Pending cohort matrix; source bridge/provider fragments are reviewed as a named foundation. |
| Jira | 617 | Pending cohort matrix; response/media exact-selection evidence is required. |
| Notion | 49 | Pending cohort matrix and fresh review. |
| Sentry | 223 | Pending cohort matrix and fresh review. |
| Stripe | 589 | Pending cohort matrix; serialization/media shapes are reviewed as named foundations. |
| Vercel | 400 | Pending cohort matrix; binary/media shapes are reviewed as named foundations. |
| **Total** | **4,341** | No identity has been admitted, erased, or reclassified in this task worktree. |

## Initial findings and disposition

| Finding group | Audit disposition | Review action |
| --- | --- | --- |
| F1, F2 | Reusable parser/normalization foundation | Named bounded-cohort foundation; red-test and Atlas-backed. |
| P1–P6 | Shared source retention/projection/generator foundation plus connector materialization | Named bounded-cohort foundation and affected mappings may ship together; preserve ledger and citations. |
| R1–R4 | Shared exact runtime foundation | Named bounded-cohort foundation; prove closed, source-cited contracts before reuse. |
| E1–E3 | Provider-evidence gap | Per-connector generated surface must retain a cited typed blocked/unsupported outcome; no invented execution. |

## Review status

`IN PROGRESS — policy correction, P1 retention, and P2 mapping evidence`. Captain instruction `006.msg` changes the delivery policy itself: a bounded named cohort may carry its named shared foundations after the Atlas check. This is neither a code-quality verdict nor a fresh-context zero-blocker review. A fresh review must be rerun against the exact final code SHA after the policy correction, matrix, implementation, and validation gates are green.

## Checkpoint Codex-only audit (not final review)

- Audited code candidate: `6ae39d7df956ea06d94ac5f6a9b590c25e4cd385`.
- Remote parent at review: `624f76617cfb57eb5934e43527ba41af8e75ceff`.
- Review route: Captain-approved Codex-only audit; no Claude dependency was requested or used.
- Scope: P2's connector-filtered, read-only `operation-evidence` bridge; its tests, generated Batch-1 report, and same-change Atlas/GSD documentation. No runtime transport, provider request, credential, or approval semantics changed.

### Evidence reviewed

- `git merge-base --is-ancestor origin/fm/cli-top100-declaration-batch-r1 6ae39d7df956ea06d94ac5f6a9b590c25e4cd385` passed; the remote parent was unchanged during review.
- `git diff --check origin/fm/cli-top100-declaration-batch-r1..6ae39d7df956ea06d94ac5f6a9b590c25e4cd385`, the two focused P2 tests, `go vet ./cmd/connectorgen`, `make docs-check`, and JSON parsing for the Atlas/report passed.
- The scoped generator's `--check` confirmed exactly 4,341 selected rows: Asana 249, Bitbucket 297, CircleCI 111, Docker Hub 54, GitLab 1,752, Jira 617, Notion 49, Sentry 223, Stripe 589, and Vercel 400. Its selector requires an explicit output and refuses both duplicate selections and a mixed valid/missing selection.
- The full `./cmd/connectorgen` suite was not recorded as green: this task harness interrupts a command at its 30-second boundary while that package's broad suite is still running. CI must run the full suite.

### Findings and disposition

| Finding | Evidence | Disposition |
| --- | --- | --- |
| Filtered report could hide a selected connector without a lock | The first implementation accepted `asana` plus `missing`; `TestOperationEvidenceConnectorFilterIsBounded` now rejects the set. | Fixed in this checkpoint. |
| GitLab source identities could remain invisible in a cohort-specific report | `TestOperationEvidenceGitLabSourceLockBridge` verifies all 1,752 unique locked identities and complete citations. | Fixed for read-only evidence projection only. |
| The required generated command/help/docs disposition invariant is not yet met | Report totals: 1,846 `canonical.missing`, 3,439 no-CLI rows, 3,721 runtime-disabled rows; existing gaps have not yet become named-foundation or provider-evidenced command outcomes. | Blocking remaining P2/P3/P4–P6 work; do not call the cohort reconciled. |
| GitLab cannot complete the v2/v3 materialization path | `source-materialize gitlab --check` refuses v1–v3 locks; `source-import gitlab --check` lacks `gitlab-retained-artifacts.json`. | Blocking P2 materialization; retain as explicit red evidence. |
| Full package check did not complete within the local command boundary | The local harness cut it off while the suite was active. | CI required; no local green claim. |

This checkpoint has **no zero-blocker or merge-ready verdict**. It records a reviewed incremental commit only; the final exact-SHA review must include all remaining foundations, source dispositions, command/help/docs reachability, PR checks, and green CI.

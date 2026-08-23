---
phase: cli-current-foundations-main-integration-r1
reviewed: 2026-08-21T03:27:40Z
source_sha: 8a8a866ff6d5282c28bda12acceed8a624218f01
inputs:
  - .planning/phases/cli-current-foundations-main-integration-r1/reviews/POSTFIX-MAPPING-REVIEW.md
  - .planning/phases/cli-current-foundations-main-integration-r1/reviews/POSTFIX-RUNTIME-REVIEW.md
  - .planning/phases/cli-current-foundations-main-integration-r1/reviews/POSTFIX-ORCHESTRATION-REVIEW.md
  - .planning/phases/cli-current-foundations-main-integration-r1/REVIEW.md
  - .planning/phases/cli-current-foundations-main-integration-r1/REVIEW-FIX.md
raw_ids: 47
atomic_claims: 73
compound_raw_ids: 24
intra_id_regrouping_reductions: 26
cross_lens_duplicate_merges: 1
canonical_findings: 46
canonical_blockers: 38
canonical_warnings: 8
status: issues_found
verdict: blockers
---

# Foundation Post-Fix Review Convergence

## Verdict and arithmetic

**Merge remains blocked.** The three ledgers contribute 47 raw IDs: mapping 14 (12 blocker, 2 warning), runtime 16 (13 blocker, 3 warning), and orchestration 17 (14 blocker, 3 warning). Compound expansion produces 73 atomic claims. Twenty-four compound raw IDs account for 26 intra-ID regrouping reductions, returning the 47 logical source findings. One true cross-lens duplicate pair then yields **46 canonical findings: 38 BLOCKER and 8 WARNING**.

```text
47 raw IDs
  -> 73 atomic claims
  -  26 intra-ID regrouping reductions
  =  47 logical source findings
  -   1 true cross-lens duplicate merge
  =  46 canonical findings (38 blocker + 8 warning)
```

At the atomic layer, the duplicate is the foreign-file overwrite/delete claim common to `PFR-BL-10` and `ORCH-PF-B12`. The canonical finding also retains orchestration's additional visible-placeholder/crash and directory-fsync claims; nothing is discarded.

## Dedupe rule and result

A merge was allowed only when all three conditions held: materially identical failure mode, the same affected implementation path, and the same minimal production fix/red behavioral contract. Only the binary pair passed:

| Merged raw IDs | Canonical | Exact rationale |
|---|---|---|
| `PFR-BL-10`, `ORCH-PF-B12` | `PF-CF-B22` | Both cite `internal/connectors/engine/binary_read.go:461-537`: replaceable visible reservation, ordinary rename over the final path, and unowned cleanup. Both require atomic no-replace publication and owned cleanup. Orchestration's crash/directory-fsync edges are preserved in the same state-machine fix/tests. |

No other overlap was deduplicated:

- `PFM-B06` classifies GraphQL source input/output locations; `PFR-BL-01` fixes the later syntax-aware public receipt boundary.
- `PFM-B08`, `PFR-BL-09`, `ORCH-PF-B05`, and `ORCH-PF-B06` are four different numeric boundaries: GraphQL Int32 domain, CLI lexical coercion/minimum, source cloning, and provider read-back equality/hash.
- `PFR-BL-08` is live destination authorization immediately before each effect; `ORCH-PF-B11` is cross-process durable auth epoch admission before every send/query.
- `PFM-B04/B05`, `PFR-BL-11`, and `ORCH-PF-B02` concern GraphQL direction, cursor authority/bounds, and ETL budget-as-EOF respectively.
- `PFR-BL-05`, `PFR-BL-12`, `ORCH-PF-B09`, and `ORCH-PF-B10` lose results at runner/native, SQS-native, ordinary ETL CLI, and reverse finalization persistence boundaries respectively.

## Raw-ID and atomic crosswalk (47/47, 73 atoms)

`Residual/regression` means the prior fix ledger claimed the relevant boundary closed. `New` means the post-fix review exposed a distinct gap. `Mixed` preserves both statuses within a compound source finding.

| # | Raw ID | Sev | Atoms | Atomic claim labels | Canonical | Prior-closure classification |
|---:|---|:---:|---:|---|---|---|
| 1 | `PFM-B01` | B | 1 | `.a` noncommutative/stale parity artifact set | `PF-CF-B01` | Residual/regression (`CF-B02`, evidence) |
| 2 | `PFM-B02` | B | 1 | `.a` gap identities skipped by projection+coverage while implemented | `PF-CF-B02` | Residual (`CF-B02/B03/B22`) |
| 3 | `PFM-B03` | B | 1 | `.a` Google Ads legal nested bodies closed empty | `PF-CF-B03` | New |
| 4 | `PFM-B04` | B | 1 | `.a` GraphQL connection accepts zero direction | `PF-CF-B04` | Residual (`CF-B04`) |
| 5 | `PFM-B05` | B | 1 | `.a` backward GraphQL reads forward pageInfo | `PF-CF-B05` | Residual (`CF-B04`) |
| 6 | `PFM-B06` | B | 2 | `.a` secret query input allowed inline; `.b` provider-issued result secrets unclassified | `PF-CF-B06` | Residual (`CF-B08/B23`) |
| 7 | `PFM-B07` | B | 1 | `.a` mutation documents omit nested resource identity | `PF-CF-B07` | New GraphQL gap |
| 8 | `PFM-B08` | B | 1 | `.a` GraphQL Int lacks signed-32-bit domain | `PF-CF-B08` | New |
| 9 | `PFM-B09` | B | 1 | `.a` embedded GraphQL projection not digest-bound | `PF-CF-B09` | Residual (`CF-B01`) |
| 10 | `PFM-B10` | B | 1 | `.a` Foundation evidence ignores current code/base/components | `PF-CF-B10` | Residual (`CF-W08`) |
| 11 | `PFM-B11` | B | 1 | `.a` historical certification evidence stays live after subject drift | `PF-CF-B11` | Residual (`CF-W08`) |
| 12 | `PFM-B12` | B | 2 | `.a` deleted route skipped; `.b` unresolved parameter ref skipped | `PF-CF-B12` | Residual (`CF-B02`) |
| 13 | `PFM-W01` | W | 1 | `.a` surface-sync preserves changed/deleted provider-owned fields | `PF-CF-W01` | New robustness residual |
| 14 | `PFM-W02` | W | 2 | `.a` website loses flag semantics; `.b` skills lose flag semantics | `PF-CF-W02` | Mixed (`CF-W01/W07`) |
| 15 | `PFR-BL-01` | B | 2 | `.a` raw/declared receipt secrets leak; `.b` GraphQL error metadata leaks | `PF-CF-B13` | Residual (`CF-B08/B11/B23`) |
| 16 | `PFR-BL-02` | B | 2 | `.a` concrete substring rewrites keys/headers/IDs; `.b` keyword heuristic rewrites messages | `PF-CF-B14` | Regression (`CF-B23`) |
| 17 | `PFR-BL-03` | B | 2 | `.a` hook wire sequence/record not approval-bound; `.b` compound responses absent | `PF-CF-B15` | Mixed (`CF-B17` + new receipt gap) |
| 18 | `PFR-BL-04` | B | 2 | `.a` JSON/config state rematerialized; `.b` multipart file/digest TOCTOU | `PF-CF-B16` | Residual (`CF-B17`) |
| 19 | `PFR-BL-05` | B | 2 | `.a` commandrunner erases error result; `.b` Ashby wrapper erases result | `PF-CF-B17` | Residual (`CF-B11`) |
| 20 | `PFR-BL-06` | B | 2 | `.a` stream loses last response; `.b` buffered retry loses last response | `PF-CF-B18` | Residual (`CF-B13`) |
| 21 | `PFR-BL-07` | B | 2 | `.a` status success has no complete receipt; `.b` status errors erase response | `PF-CF-B19` | New receipt-kind gap |
| 22 | `PFR-BL-08` | B | 1 | `.a` authorization stale before actual destination effect | `PF-CF-B20` | New boundary residual |
| 23 | `PFR-BL-09` | B | 2 | `.a` CLI numeric lexeme rounds; `.b` minimum metadata/comparison rounds | `PF-CF-B21` | Residual (`CF-B19`) |
| 24 | `PFR-BL-10` | B | 1 | `.a` no-overwrite path overwrites/deletes competing final file | `PF-CF-B22` | New; duplicate of `ORCH-PF-B12.b` |
| 25 | `PFR-BL-11` | B | 2 | `.a` cursor URL adds undeclared query; `.b` cursor lacks control/encoded-size admission | `PF-CF-B23` | Residual (`CF-B21/B22`) |
| 26 | `PFR-BL-12` | B | 2 | `.a` SQS success omits receipt; `.b` SQS post-I/O errors discard response | `PF-CF-B24` | New native residual (`CF-B11`) |
| 27 | `PFR-BL-13` | B | 1 | `.a` SQS redirect can forward session token cross-origin | `PF-CF-B25` | Residual (`CF-B14`) |
| 28 | `PFR-WR-01` | W | 1 | `.a` declared idempotency header previewed then stripped | `PF-CF-W03` | New latent gap |
| 29 | `PFR-WR-02` | W | 1 | `.a` minLength witness falsely rejects valid schema | `PF-CF-W04` | New latent gap |
| 30 | `PFR-WR-03` | W | 1 | `.a` multipart symlink security test races/false-passes | `PF-CF-W05` | New |
| 31 | `ORCH-PF-B01` | B | 2 | `.a` transport effects precede ownership CAS; `.b` CDC files precede CAS | `PF-CF-B26` | Explicit regression (`CF-B24`) |
| 32 | `ORCH-PF-B02` | B | 2 | `.a` capped full overwrite publishes prefix; `.b` incremental cap cannot resume | `PF-CF-B27` | New |
| 33 | `ORCH-PF-B03` | B | 2 | `.a` conformance authority self-supplied; `.b` claimed-keyed POST lacks provider proof | `PF-CF-B28` | Residual (`CF-B27`) |
| 34 | `ORCH-PF-B04` | B | 2 | `.a` readback ignores receipt; `.b` bounded prefix is unrelated proof | `PF-CF-B29` | Residual (`CF-B07`) |
| 35 | `ORCH-PF-B05` | B | 1 | `.a` declarative clone rounds >2^53 | `PF-CF-B30` | New |
| 36 | `ORCH-PF-B06` | B | 2 | `.a` numeric equality is Go-type-sensitive; `.b` identity hash is type-sensitive | `PF-CF-B31` | New |
| 37 | `ORCH-PF-B07` | B | 1 | `.a` apply/publish and readback share deadline | `PF-CF-B32` | New |
| 38 | `ORCH-PF-B08` | B | 2 | `.a` runtime/catalog clone shallow; `.b` Arrow segment request aliases state | `PF-CF-B33` | New |
| 39 | `ORCH-PF-B09` | B | 3 | `.a` ordinary CLI drops App error run; `.b` sidecar error hides App run; `.c` ambiguous save returns zero | `PF-CF-B34` | Residual (`CF-B09`) |
| 40 | `ORCH-PF-B10` | B | 2 | `.a` definite no-commit run fabricated; `.b` indeterminate commit not reloaded | `PF-CF-B35` | New finalization gap |
| 41 | `ORCH-PF-B11` | B | 2 | `.a` cross-process admitted work keeps sending; `.b` ambiguous fence CAS skips local cancel | `PF-CF-B36` | New auth-cohort gap |
| 42 | `ORCH-PF-B12` | B | 3 | `.a` visible placeholder/crash; `.b` competing file overwrite/delete; `.c` no directory fsync | `PF-CF-B22` | New; `.b` duplicates `PFR-BL-10.a` |
| 43 | `ORCH-PF-B13` | B | 1 | `.a` CDC restore accepts unrelated regular files | `PF-CF-B37` | New |
| 44 | `ORCH-PF-B14` | B | 2 | `.a` receipt-retire error after checkpoint; `.b` marker error after checkpoint | `PF-CF-B38` | Residual (`CF-B06`) |
| 45 | `ORCH-PF-W01` | W | 1 | `.a` parking store indeterminate commit diverges live state | `PF-CF-W06` | New durability gap |
| 46 | `ORCH-PF-W02` | W | 1 | `.a` expired/revoked parked authorization rearms forever | `PF-CF-W07` | New |
| 47 | `ORCH-PF-W03` | W | 1 | `.a` declared preflight failure relabeled absent | `PF-CF-W08` | Residual (`CF-W06`) |

## Methodology assessment

- **Coverage:** all three post-fix ledgers are complete at their declared depth and frozen SHA. Mapping covers generator/source/certification breadth; runtime traces request, receipt, masking, filesystem, and native boundaries; orchestration traces durable state, concurrency, recovery, and CLI terminal publication.
- **Evidence quality:** findings cite concrete symbols/lines, end-to-end call paths, behavioral examples, six-surface impact, minimal fixes, and red-first tests. Runtime additionally reproduced a race; mapping reproduced required parity-check failures; orchestration identified tests that explicitly encode losing side effects/truncated success.
- **Tension resolution:** narrower passing statements were not treated as contradictions. For example, ordinary engine receipts can exist while runner/native adapters erase them; declarative redirect repair can coexist with native SQS redirect leakage; ordinary occurrence IDs can be preserved while generated GraphQL documents omit resource fields.
- **Severity:** highest source severity is preserved. The one merged pair is BLOCKER in both ledgers. No warning was promoted or blocker downgraded during convergence.
- **Losslessness:** every raw ID appears exactly once in the table; atom counts sum to 73; every row maps to exactly one canonical ID; both members of the only merged pair are retained as sources of `PF-CF-B22`.

## Artifact and merge disposition

Canonical fix/test contracts and the dependency-ordered one-wave plan are in `.planning/phases/cli-current-foundations-main-integration-r1/POSTFIX-REVIEW.md`.

**Verdict: blockers.** Thirty-eight blockers remain, so the Foundation lane is not mergeable at `8a8a866ff6d5282c28bda12acceed8a624218f01`.

---

_Converged: 2026-08-21T03:27:40Z_

_Frozen diff: e62ae21d428f0d27225f9bff564dc2cd797f6b65..8a8a866ff6d5282c28bda12acceed8a624218f01_

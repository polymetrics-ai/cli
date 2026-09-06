# CP11 acceptance projection — Firstmate079/081

This is a compact navigation view of [CP11-ACCEPTANCE.json](CP11-ACCEPTANCE.json), not a second acceptance ledger. The JSON owns every stable row, original authority, object/phase, expected state, assertion and receipt gap. Counts are contract accounting, not new bugs, new requirements or one-test-per-row targets.

Behavior/test/dependency candidate: `7481d1770a21cc95869fd10bf281f632af48c089`; tree `a2e583336ffa8ad86a0de95110259342bfa6dab0`. Audited evidence commit: `053853368a1514eaadf0b2411ab8740959559797`. Original audit SHA-256: `bc109e85fdde9d1958b2cde7874a3f7b30b8e5d06b1b0c2764088fb34fa3e0a0`. CP11 remains unaccepted.

Design: **complete, design only** — [full report](/Users/karthiksivadas/pm-cli-agent-workspace/data/cli-batch1-pi-takeover/cp11-ownership-design-079.md). Report SHA-256: `35f8926266fdae784e88d429eeea0b27adcd114a1550ca47cd7dece3f592a417`. The owner read the full return. It selects explicit partial-open ownership, common-writer cleanup of proven-owned incomplete records, retention of complete bound authority, and separate collision/completion classification. No unresolved architecture or product question was reported. Firstmate’s coherent implementation brief remains required.

| Finding | Kind | Rows | Bounded contract and caller projection |
| --- | --- | ---: | --- |
| 7481-01 | Production | 35 | **Creation and partial-record ownership.** Prepared, marker and phase writers; actual phase callers and valid bootstrap/successor classes. |
| 7481-02 | Production | 14 | **Complete collision outcome.** CreateTemp plus CURRENT/JOURNAL and recovery consumers; pure versus meaningful compound, absence/conflict compatibility. |
| 7481-03 | Proof | 11 | **Failed-allocation A/B and recovery.** Public Publish, direct CURRENT/JOURNAL, generation Prune, stage Recover; empty/nonempty DIRECTORY B. |
| 7481-04 | Proof | 30 | **Independent expected state.** Nine unheld generation schedules × empty/nonempty lease-file B; held reader; owned stage; separate Check; oracle counterexamples. |
| 7481-05 | Claim / retained proof | 6 | **Correct unsupported claims.** Keep mandatory nested Recover-stage/Prune-generation proof. Nested Publish currently reaches initial recovery; extra final-prune scenario optional. |
| 7481-06 | Proof | 23 | **Stipulated compound controls.** Write/short Write/Sync + Close for each record kind; opened-control siblings, link, capture, stage, one-Close ownership. |

## Actual caller/cut navigation

- `01.*`: first-side-effect/partial-success states; prepared/marker/phase frontiers; CURRENT/JOURNAL bootstrap/successor classes; exact phase-call inventory. The linked design supplies representation/ownership; these draft rows remain unclosed.
- `02.*`: allocation outcome cases and the complete classifier chain. `evidence_binding_gap` means the current explicit binding is missing; it does not assert a new standalone test is required.
- `03.*`: actual failed allocations must preserve actor-created DIRECTORY B and opened A, exact residue and fresh recovery. Existing direct CURRENT/JOURNAL fixtures allocate successfully and fail later cleanup; their wrong-phase status is explicit.
- `04.cut.*`: explicit Prune; no-JOURNAL Recover; Open; Publish initial recovery; prepared/new-selected Recover; committed/new-selected Recover; successful Publish final prune; rejected-new fresh Recover; immediate rollback; held-reader Prune; owned stale-stage Recover. Check is separate and non-destructive. Only the nine unheld generation rows have both empty/nonempty regular lease-file B requirements. Stage uses its marker/root; held reader keeps empty B.
- `04.oracle.*`: unrelated readable root, same bytes/wrong identity, missing/replaced stable history, wrong phase, F03A old-history skips and actor-B timing. Legitimate new controls/phases/history must remain permitted.
- `05.*`: preserve genuine mandatory nested resource coverage; correct initial-recovery/final-prune and post-Stat child-identity overclaims without inventing new guarantees.
- `06.*`: retain useful existing compounds and complete the explicitly missing 058 combinations. Existing successful-link+Close or pre-Sync+Close does not silently replace failed-link+Close or real writable completion.

## Current readiness

All 119 rows remain open or partial. Current disposition counts, derived from JSON:

- `claim_correction_pending`: 2.
- `evidence_binding_gap`: 26.
- `incomplete_assertion`: 42.
- `missing_negative_control`: 7.
- `missing_test`: 13.
- `oracle_counterexample`: 1.
- `original_desired_red`: 6.
- `partial_existing_coverage`: 14.
- `retained_bounded_coverage`: 4.
- `wrong_phase_existing_test`: 4.

101 rows have no bound exact assertion location; 33 have no receipt reference. These are draft evidence gaps, not 101 proven absent tests or 33 required new executions. Structural ID/reference/symbol/hash/caller-variant checks passed; semantic completeness, repaired GREEN, final original-range lint and independent acceptance remain pending.

## Evidence custody and prompt use

External test bytes are separate from production SHA 7481. The JSON records each content hash, overlay hash and added virtual test path. The supplied three-probe result is a **transcription** (exit 1: two production REDs and an oracle counterexample). The new marker/phase four-cell result is an **original tool command/result capture**, not a rollout export (exit 1, 1.079s/5.26s). Neither modifies existing project sources/tests or supplies GREEN.

The normal/race package receipts (319.928s/783.834s) bind intended final precommit source, not a postcommit run or semantic closure. Cached preflight, recovered scanner success, separate package failure and CP29 debt remain distinct. Prior resolved B01/B02/R3/F01/F02/F04R/F08R/F07 protections remain in JSON by reference.

Use this projection plus exact relevant obligation IDs in later Firstmate-authored prompts; fetch detail on demand. Do not paste the full JSON or duplicate the original reports into each prompt. The approved execution policy retains CP11 original-base full-range review; accepted-baseline scopes apply later.

Existing delivery: https://github.com/polymetrics-ai/cli/pull/4294.

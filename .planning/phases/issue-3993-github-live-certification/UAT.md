---
phase: "3993"
status: passed_for_harness_slice_certification_not_ready
mode: inline_manual
---

# UAT — GitHub live certification harness and inbound warehouse proof

`scripts/gsd prompt verify-work 3993 --auto` was run on 2026-08-11. The
project-local GSD adapter cannot instantiate a roadmap phase for issue 3993,
and this dispatch prohibits role spawning, so the required UAT is completed
inline against committed code and recorded evidence.

The acceptance target for this child is a truthful, runnable harness and an
inbound warehouse proof—not a successful provider certification. The measured
0/665 barrier result remains a passing observation for evidence integrity and
an open certification failure for the parent program.

| ID | Deliverable | Automated / recorded evidence | Result |
| --- | --- | --- | --- |
| D1 | A supplied immutable, run-owned `Polymetrics-Cert` boundary controls generation and dispatch. | `github-live-cases.test.mjs`; boundary and slug/immutable-ID rejection cases; proof-sweep regression rejects direct and `--config` write-scope overrides before any PM child starts; source-derived manifest check. | pass |
| D2 | All applicable commands launch through one barrier and every result is terminally accounted for. | `github-live-proof-sweep.test.mjs`; self-test; recorded R2 run with `single_barrier_release`. | pass |
| D3 | Live credential use is App-only, secret-free in artifacts, and a result is read back rather than inferred from exit status. | Sanitized R2 evidence: one App-authenticated returned-record control and a separate 16-way returned-record control; no revoked fine-grained-token probe. | pass |
| D4 | The current case ledger is reclassified from the supplied boundary rather than a pinned historical repository. | Fresh ledger: 665 attemptable / 856 blocked; reason-family movement in `SUMMARY.md`; manifest and bootstrap checks pass. | pass |
| D5 | GitHub source flows through the warehouse and is read back through DuckDB. | Built `pm`; `pm flow plan`, preview, and run read 1/wrote 1; `PAR1` materialization; `pm query` returned 1 row. | pass |
| D6 | Temporary credential and warehouse material is gone after evidence capture. | Sanitized evidence scan plus final temporary-run residue check; no provider mutation was dispatched. | pass |

## Explicit non-acceptance

The full credentialed provider measurement is **0 proven / 665 failed / 856
untestable**: each admitted child reached the required 45,000 ms terminal
bound. REST changed 15,000 to 14,997 and GraphQL stayed 5,000, while isolated
App reads succeeded. This is neither invalid-credential evidence nor a
provider-quota result. It is a retained #3993 local process-admission finding.

GitHub certification remains blocked by a provider-admission solution that
preserves the barrier semantics (#3990), 767 cleanup-safe mutation fixtures,
and outbound dependencies #3994 and #3992. No action approval, schedule, or
shared rate-policy implementation was added in this child.

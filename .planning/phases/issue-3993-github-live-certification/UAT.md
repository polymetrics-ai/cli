---
phase: "3993"
status: partial_current_sha_needs_decision
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
| D4 | The current case ledger is reclassified from the supplied boundary rather than a pinned historical repository. | Canonical classifier: 1,521 total, 182 attemptable / 1,339 blocked, with 169 / 470 direct reads and static -483 / +483 movement from the separately named historical terminal measurement; manifest readiness and case digest are bound to the exact case order. | pass |
| D5 | GitHub source flows through the warehouse and is read back through DuckDB. | Built `pm`; `pm flow plan`, preview, and run read 1/wrote 1; `PAR1` materialization; `pm query` returned 1 row. | pass |
| D6 | Temporary credential and warehouse material is gone after evidence capture. | Sanitized evidence scan plus final temporary-run residue check; no provider mutation was dispatched. | pass |

## Explicit non-acceptance

The historical credentialed provider measurement is **0 proven / 665 failed /
856 untestable**: each admitted child reached the required 45,000 ms terminal
bound. REST changed 15,000 to 14,997 and GraphQL stayed 5,000, while isolated
App reads succeeded. This is neither invalid-credential evidence nor a
provider-quota result. It is a retained #3993 local process-admission finding.

GitHub certification remains blocked by a provider-admission solution that
preserves the barrier semantics (#3990), typed cleanup-safe fixtures for current
writes, and outbound dependencies #3994 and #3992. No action approval,
schedule, or shared rate-policy implementation was added in this child.

## Fresh current-SHA auto-verification — lineage 1/5

The captain-authorized `verify-work 3993 --auto` path is executed inline: this
issue is not an interactive roadmap phase, dispatch prohibits GSD role spawning,
and the only unavailable observation is an external credential/boundary gate.
The entries below distinguish automatically observed behavior from a provider
result rather than asking a human to attest to a secret or external mutation.

| ID | Expected observable result | Evidence | Result |
| --- | --- | --- | --- |
| F1 | An external per-operation `pm` runner cannot be accepted as one-process credentialed-live evidence and cannot reach credential/provider dispatch. | Fresh RED then GREEN in `github-live-proof-sweep.test.mjs`; focused suite 15/15. | pass (automated) |
| F2 | The current built `pm` wires the in-process certify harness and its focused GitHub/Go gates remain green. | SHA-256-bound `./pm`; combined Node suite 51/51; `internal/connectors/certify` and `internal/cli` package tests pass. | pass (automated) |
| F3 | A current credentialed GitHub read/write/warehouse proof is performed only inside the approved App-owned immutable boundary. | Approved GitHub App credential source and `Polymetrics-Cert` boundary were absent; no provider request began. | blocked (`third-party` safety prerequisite) |

**Fresh result:** partial / `needs-decision`. There is no user-facing action to
test until an authorized operator supplies the existing approved credential and
boundary through the sanctioned secret channel. This is not a code gap and does
not create a GSD gap plan; it is an external safety gate.

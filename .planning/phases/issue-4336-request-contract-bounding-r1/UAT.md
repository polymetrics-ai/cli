# UAT — issue-4336 request-contract execution envelopes

Inline verify-work fallback: the canonical delivery contract forbids spawning a
verifier role in this lane. All six SUMMARY deliverables are deterministic and
carry passing automated evidence; none requires human visual judgment or a live
provider credential.

| ID | Deliverable | Evidence | Verdict |
| --- | --- | --- | --- |
| D1 | Exact Gong-shaped valid source is represented without synthetic schema bounds | Real importer fixture preserves `{"type":"string"}` and emits no gap | passed |
| D2 | Finite envelope is mandatory and canonical | Missing/altered parameter and body envelope regressions plus full generator suite | passed |
| D3 | Encoded cap is enforced before I/O | Local HTTP boundary receives cap exactly once and receives no cap+1 request | passed |
| D4 | Provider numerics remain exact and compatible | Huge integer/decimal bounds, identical wire lexemes, and legacy numeric grammar cases | passed |
| D5 | Operators can inspect PM provenance | Built `pm` help, bare namespace, command help, and JSON inspection checks | passed |
| D6 | Unsupported semantics remain honest gaps | String/numeric header quarantine, dynamic/composition/untyped controls, and clean 552-connector boundary scan | passed |

Verdict: **passed**. This verifies the shared policy foundation, not a final
Batch-1 command count. That recount depends on the independent #4339 quarantine
and dialect work.

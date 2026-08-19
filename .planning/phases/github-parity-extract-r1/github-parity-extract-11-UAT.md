# UAT — plan 11 generated fixed GraphQL contracts

**Mode:** automated inline/manual fallback (`verify-work --auto`). The parent lane forbids spawning
GSD roles and no human/provider interaction is appropriate for a no-provider local checkpoint.

| Acceptance behavior | Automated exercise | Result |
| --- | --- | --- |
| Every source-root has a deterministic typed artifact | generator and combined-ledger tests | PASS — 31 query + 274 mutation roots, exact one-to-one command/operation binding |
| The transport cannot become raw GraphQL | loopback engine tests and exact named endpoint-ledger checks | PASS — no caller document, selection, header, endpoint, or cursor channel |
| A direct-read query cannot select an appended mutation | mixed-document preflight regression | PASS — exactly one named operation of the declared kind is required |
| Failed GraphQL responses cannot echo raw provider bodies | loopback non-2xx query/mutation regressions | PASS — both error paths redact the fixture value |
| Mutations retain their lifecycle safety | command help and shared approval tests | PASS — plan/preview/approval; destructive acknowledgement where declared |
| `deleteIssue` remains unavailable | generator and `pm ... delete-issue --help` | PASS — `unsafe_or_disallowed` |
| Source inventory cannot silently stale | hermetic source-drift tests and scheduled read-only workflow | PASS — independent REST/GraphQL totals and hashes are enforced |
| Lab safety remains closed before live work | PM-only lab test suite and boundary check | PASS — deny by default; exactly one declared target |

No human judgment is pending for this checkpoint. Its handoff condition is deliberately local:
future live cohorts may proceed only under the existing PM-only lab boundary and must not reinterpret
these UAT results as a provider acceptance result.

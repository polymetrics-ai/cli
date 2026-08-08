# Verification Checklist — Zoom Clips parity, R1

## Planned checks

- [x] Live artifact URL, retrieval timestamp, byte count, digest, operation count, and ledger
  comparison recorded before RED.
- [x] Prior-worker five-file handoff inspected without copying or deleting the old worktree.
- [x] RED capture committed before production declaration or foundation changes.
- [x] Closed root-array direct-write foundation is green with object-body regressions.
- [x] Declared bearer redirect foundation is green without permitting arbitrary credential forwarding.
- [x] Closed operation-level base64 image upload foundation is green with pre-network bounds,
  media/name, snapshot, redaction, and provider-bounded bearer-redirect assertions.
- [ ] All 21 endpoint rows and 23 concrete commands run through real preflight and fixtures.
- [ ] Every documented `204` action is status-only and destructive confirmation-gated.
- [ ] Endpoint ledger reconciliation is confined to Zoom Clips; zero Zoom rows are
  `unsafe_or_disallowed`.
- [ ] Generated docs/site output retains Zoom-only changes after whole-file generation.
- [ ] Fresh `pm` binary reaches base, namespace, provider group, and all command help routes.
- [ ] Scoped local gates, inline verify-work, and manual code review are complete.

## Captured results

The 2026-08-08 RED checkpoint recorded the failing 102/1,740/55/42 inventory totals, all
twenty-three unknown Clips command paths, root `json_array` rejection, bearer stripping on the
declared binary redirect fixture, and missing direct-write base64 transformation. The verbatim
output is retained in `TDD-LEDGER.md`; its fixtures contain only synthetic values.

# Descriptor-Free Retention Admission — Plan

## Goal

Allow a source-backed, non-executable seven-lane contract to retain complete
mapping accounting when its immutable source lock reconciles exactly, while
preserving canonical-descriptor requirements for every executable claim.

## Red → green → refactor sequence

1. Add failing tests for a `retention_only` contract constructed from the real
   Jira, Sentry, and Vercel frozen lock/matrix evidence. The tests must prove
   exact-ID coverage, support ordinary spaces in Sentry provider IDs, and no
   descriptor import.
2. Add failing negative tests for implemented lanes, unmapped cells, legacy
   method partitions, missing/duplicate IDs, unsafe IDs, runtime artifacts,
   and a descriptor-less executable claim.
3. Extend the closed contract schema and contract validator with a dedicated
   retention-only invariant. Keep legacy ordinary contracts compatible.
4. Teach only `checkSourceProjection` to accept the missing descriptor after
   the existing enabled-contract bridge proves exact retained-source evidence
   and the dedicated invariant passes.
5. Update `source.projection-admission.v1` and terminology with the
   mapping-only boundary and proof names.
6. Run format, focused normal/race tests, affected package tests/vet,
   JSON/agent-contract/diff gates. Commit locally only if green.

## Admission invariant

`retention_only` is valid only when all of the following are true:

1. There is one primary source lock and no supplemental source locks.
2. Every claimed source cell uses exact operation IDs, never method selectors.
3. All source operations reconcile exactly once in a primary exact-ID
   partition; overlays reference only those retained IDs.
4. No lane or source coverage count claims `implemented` or
   `unmapped_mapping`; remaining source counts are mapped-unproven, deferred
   foundation, or provider-evidenced/not-applicable.
5. Each lane references only the primary source-lock artifact; it has no
   transport or warehouse binding.
6. Retention-only source IDs remain opaque provider data, including ordinary
   spaces and `/`; reject only empty IDs and control characters. IDs are never
   normalized or used as filesystem paths.
7. A missing descriptor remains a failure for every ordinary or executable
   contract.

## Foundation Atlas update

Update the existing mapping-admission foundation
`source.projection-admission.v1`, rather than introducing a runtime
foundation. The entry must state that descriptor-free retention is source
accounting only, cannot admit execution, and requires the exact-ID,
source-only invariant above.

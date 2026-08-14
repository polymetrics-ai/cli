# Inline code review — Issue #3974 typed database foundation recovery

## Verdict

**PASS — no actionable source finding.** The review was performed inline because the canonical
contract permits one worker only and the numbered GSD role runtime is unavailable for this
issue-local phase. It is a manual `code-review` fallback, not a skipped review.

## Reviewed boundaries

| Area | Result |
| --- | --- |
| `database.Load` and schema | Closed members, duplicate/case-alias rejection, bounded numeric policy, no raw-value errors, context checks, and defensive projections are intact. |
| Logical types/catalog/read plans | Opaque mappings cannot compile, compatibility remains exact/lossless only, identities are structured, resource limits are finite, and a read plan requires a non-null unique-key suffix. |
| Native admission | Driver identity is exact; a manifest declaration alone cannot admit a command; sealed inbound/outbound warehouse legs require distinct matching native evidence. |
| Warehouse mediation | `ArtifactIdentity` is shared with structural layout ownership; inbound owns a source artifact, outbound contains no source, and no direct database pair can be expressed. |
| Engine/defs | `database.json` is optional fleet-wide, strict when present, embedded in the binary, and cannot alter metadata capabilities. |
| PostgreSQL | The reference driver is unregistered/non-executing; `write`, `query`, and `cdc` remain false; CDC remains fail-closed before source contact. |
| Current canon | The three replay overlaps preserve #4003 documentation, current bundle conventions, TLS, and the pglogrepl fail-closed staging boundary. |
| Generated artifacts | #4026 owns the sole post-replay matrix synchronization. Its diff is 15 `bundle.go` source-location references only, with no capability change. |

## Scope and safety checks

- All 12 source-range commits are present in order after the planning checkpoint; no #3864
  commit was imported or recreated.
- `git diff --check` and `gofmt -d` produced no findings.
- The focused/regression/static results are recorded in `VERIFICATION.md`.
- The #3995 compatibility result is `RETRY` only for a future certification transition; it is not
  an automatic approval and does not expand this F1 scope.

## Disposition

No source correction follows from review. The only correction round is #4026's derived-artifact
synchronization, recorded as round 1/5 in `traces/capability-matrix-red.txt`.

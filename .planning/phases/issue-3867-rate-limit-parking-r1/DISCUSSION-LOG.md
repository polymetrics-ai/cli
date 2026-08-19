# Issue #3867: rate-limit parking and automatic resumption - Discussion Log

> **Audit trail only.** Decisions are captured in `CONTEXT.md`.

**Date:** 2026-08-15
**Issue:** #3867 rate-limit parking and automatic resumption
**Mode:** `scripts/gsd prompt discuss-phase issue-3867-rate-limit-parking-r1 --auto`

The issue supplies unambiguous acceptance criteria and the automated discussion
mode selected all material implementation areas.

## Parking authority and payload

| Option | Description | Selected |
| --- | --- | --- |
| Typed rate-limit error with reset evidence | Park only when the provider supplied a parsed reset instant. | ✓ |
| Generic request failure | Would fabricate a resume time or hide a non-rate failure. | |

**Selection:** Preserve the actual typed rate-limit reason/source and reset
instant; do not park generic errors.

## Resumption and checkpoint boundary

| Option | Description | Selected |
| --- | --- | --- |
| Resume from committed checkpoint | Re-enter source execution without replaying an acknowledged apply. | ✓ |
| Restart from the beginning | Replays source and risks duplicate destination work. | |

**Selection:** The durable parking record carries a defensive committed
checkpoint clone and resumes only from it.

## Scope isolation and scheduler restart

| Option | Description | Selected |
| --- | --- | --- |
| Opaque scope-key coordinator with persisted re-arm | Same scope waits; unrelated scope proceeds; reconstructed scheduler restores pending work. | ✓ |
| Process-local generic delay | Loses state on restart and cannot distinguish unrelated scopes. | |

**Selection:** Persist opaque scope/run records and re-arm only at/after reset.

## Operator events

| Option | Description | Selected |
| --- | --- | --- |
| Typed truthful events | Report the real parsed reason/source and reset timestamp. | ✓ |
| Generic failure event | Does not establish why or when resumption will occur. | |

**Selection:** Closed, secret-free park/resume event values.

## Deferred Ideas

- #4125 `window_seconds` duration overflow.
- #4136 certification validation ordering.
- #4090.

# POSTMORTEM TEMPLATE — Flow Engine

Use this template for any incident or regression discovered in the flow-engine phase.

## Incident summary

- Date:
- Severity: (P0 / P1 / P2)
- Duration:
- Affected command(s):
- Reported by:

## Timeline

| Time | Event |
|------|-------|
| | Issue first observed |
| | Investigation started |
| | Root cause identified |
| | Fix deployed |
| | Verified resolved |

## Root cause

Describe the technical root cause in one paragraph.

## Impact

- Commands affected:
- Data integrity impact (Y/N):
- Ledger/checkpoint state impact (Y/N):
- User-visible data loss (Y/N):

## Detection

How was the issue found? (test failure / user report / CI / manual)

## Resolution

What code change fixed the issue? Reference the commit.

## Contributing factors

List any process or design factors that made this possible.

## Action items

| Action | Owner | Due date |
|--------|-------|----------|
| | | |

## Lessons learned

What should we do differently in future phases?

## Invariant violations (if any)

Did this incident violate any of the Phase 0 non-negotiables?
- [ ] New dependency introduced
- [ ] Secret value logged
- [ ] Side effect without plan→preview→approval
- [ ] Test weakened or skipped
- [ ] `make verify` was red at commit time

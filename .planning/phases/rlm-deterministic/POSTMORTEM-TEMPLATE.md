# POSTMORTEM TEMPLATE — Phase 2: RLM Deterministic Backend

Fill this out if a defect is found in production or during phase verification.

---

## Incident summary

**Date:**
**Severity:** [P0 / P1 / P2 / P3]
**Detected by:** [test / manual / user report]
**Phase gate passed?** [yes / no — if no, describe which gate missed it]

---

## Timeline

| Time | Event |
|---|---|
| T+0 | Issue first observed |
| T+N | Root cause identified |
| T+N | Fix deployed |

---

## What happened

[1-3 sentences describing the observable failure.]

---

## Root cause

[Technical root cause. Reference specific file/function.]

---

## Impact

- Records affected:
- Tables affected:
- Determinism violated? [yes/no]
- Data loss? [yes/no]
- Security impact? [yes/no]

---

## Detection gap

[Why did existing tests not catch this? Which test would have caught it?]

---

## Corrective actions

| Action | Owner | Deadline |
|---|---|---|
| Add failing test that reproduces the bug | | |
| Fix implementation | | |
| Verify fix with `go test -count=1 ./internal/rlm/...` | | |
| Update threat model if relevant | | |

---

## Lessons learned

[What invariant was violated? What process change prevents recurrence?]

---

## Common failure modes to watch for in this phase

1. **Non-determinism:** caused by map iteration order in feature scoring. All map reads must be over a sorted key slice.
2. **Path traversal in table names:** ensure `validateTableName` is called before any `filepath.Join`.
3. **Score normalization divide-by-zero:** all-zero weights must be handled before division.
4. **Partial write corruption:** `os.Rename` is atomic; never write directly to the final OutTable path.
5. **Model stub accidentally invoked:** backend selection must never default to model; `ErrUnknownMode` for any unrecognized mode.

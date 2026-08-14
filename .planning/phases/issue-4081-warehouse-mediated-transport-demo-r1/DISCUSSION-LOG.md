# #4081 — Transport demo discussion log

**Date:** 2026-08-13
**Command:** `scripts/gsd prompt discuss-phase issue-4081-warehouse-mediated-transport-demo-r1 --auto`
**Mode:** inline/manual fallback; no question was reopened because the issue,
task brief, repository contract, and live topology fix every material choice.

| Area | Options considered | Selected decision | Reason |
| --- | --- | --- | --- |
| Walking slice | Four-pair matrix / GitHub-only first leg | GitHub-only first leg | The task explicitly narrows this child to the smallest demonstrable path. |
| Warehouse boundary | Pass staged source records / durable immutable handle then reopen | Durable immutable handle then reopen | Prevents source-record aliasing and proves the actual WAL/DuckDB/Parquet mediator. |
| Destination path | Generic HTTP writer / existing typed approval-gated action | Existing typed path | Retains closed provider actions and plan-preview-approval-execute safety. |
| Evidence verifier | Descriptor self-certification / read-only accepted evidence | Read-only accepted evidence | Defaults remain fail closed and executors cannot certify themselves. |
| Local proof | Exit status / returned counts and independently observed receipts | Returned counts, hashes, receipt order, read-back, cleanup | A successful exit can hide a truncated or bypassed route. |
| Provider evidence | Unapproved token or personal fixture / approved App or safe local server | Safe local server; optional approved App only | No secret discovery or unsafe mutation boundary is permitted. |
| CLI surface | Invent a new broad command / use bounded harness unless an accepted namespace exists | Bounded harness first | Avoids a speculative public surface; parity remains mandatory if a command is added. |
| Dependency state | Implement on stale combined branch / wait for accepted #4019/#4077 content | Wait, then resume on `e7d2b296...` | The requested base guard was explicit; the squash merge is admitted by direct content identity rather than source-commit ancestry. |

## Questions intentionally not reopened

- The final combined #4016 remains draft and human-gated into `main`.
- #4079 is a corrected dependency, not work to duplicate.
- Only one #4015 sub-issue may own the production-construction slice; #4081 is
  now linked directly under #4015.
- Real-provider use is optional and restricted by the task's credential/resource
  boundary; no authorization question is needed for local planning.

## Carrier decision update — 2026-08-13

The later accepted carrier report replaces only the provisional CLI-surface row
above. The rejected one-shot `pm demo` harness is not made Green. The closed
operator route is `pm etl transport github-issue-label` for plan, preview, and
typed cleanup plus the existing `pm etl run` with an exact plan selector,
one-bounded-line stdin token carrier, and `--confirm destructive`. This is not
a broad generic namespace: repository, base URL, source/target issue, label,
action, record, and credentials remain persisted connection/App state. The
faithful binary harness invokes those commands in separate subprocesses.

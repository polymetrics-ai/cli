# #4081 scoped checkpoint plan

| Checkpoint | Allowed paths | Must prove | Push/PR state |
| --- | --- | --- | --- |
| Base admission | none | Combined `e7d2b296...` has accepted #4019/#4077 content by exact blob identity (squash-aware) | final child branch created only after guard |
| Plan | `.planning/phases/issue-4081-warehouse-mediated-transport-demo-r1/**`, `.planning/traces/issue-4015-*` | GSD context, scope, RED/GREEN, verification, safety, skill evidence, immutable topology trace | commit/push to final child branch |
| RED | focused test files + phase artifacts | Dormant registry/verifier/nil stage, raw-page rejection, tamper failure, failure ordering, #4079 control | commit/push after expected failure |
| GREEN | app construction, transport adapter/stage, bounded harness, tests, phase artifacts | durable reopen and receipt/read-back before CAS | commit/push after focused green |
| Refactor/review | only in-scope review/no-mistakes findings + evidence | formatting, gates, dispositions | commit/push, then draft PR |

The temporary local branch `fm/cli-transport-demo-mvp-construction-r1` was never
pushed and contained planning work only. The final issue branch was created only
from the admitted remote combined head; no speculative base is permitted.

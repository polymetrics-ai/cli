# Slice 0 TDD ledger

| Step | State | Required observable | Evidence |
| --- | --- | --- | --- |
| Mutation inventory projects all eligible declared commands | pending RED | A test fails until it sees exactly 279 `direct_write` and 577 `reverse_etl` fixture-required candidates derived from declarations. | Not started: depends on #4214 landing. |
| Generic projection emits execution identity without I/O | pending GREEN | The preceding test passes; each candidate has a declaration binding and derived invocation fields, with no provider request. | Not started: depends on #4214 landing. |
| Classification is exhaustive and exclusive | pending RED | A test fails on zero, duplicate, or unknown classification and on a bucket total other than 856. | Not started: depends on #4214 landing. |
| Unassessed fails closed | pending RED | A test fails if an unmatched candidate is classified `contained`; it must emit `unassessed` with evidence. | Not started: depends on #4214 landing. |
| Escape classifier refuses unsafe declarations | pending RED | Paid-seat, outside-invitation, public-publication, and third-party test declarations must not be contained. | Not started: depends on #4214 landing. |
| Contained declaration is accepted | pending GREEN | A disposable-scope declaration produces `contained` with the connector-owned evidence path. | Not started: depends on #4214 landing. |
| Generated artifacts are canonical | pending GREEN | Generator rerun produces no diff and the sweep remains 1,571. | Not started: depends on #4214 landing. |

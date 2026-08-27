# TDD ledger — issue 4364 deferred visibility/admission bridge

| Behavior | Red evidence | Green evidence | Refactor / safety note |
| --- | --- | --- | --- |
| 4,341-row manifest admits exactly 1,910 concrete deferred rows | Pending: add a named reconciliation test before manifest-loader production code. | Pending. | The test must derive totals from the source records and assert the published invariant; it must not bless the obsolete 1,908 matrix. |
| Deferred row preserves exact citation, lane, stable CLI path, target, and one foundation | Pending: malformed fixture table for duplicate/missing/generic/policy/multiple-gap cases. | Pending. | Reuse existing closed citation, endpoint, and foundation-component validators. |
| Generated deferred declaration has no fabricated typed executor | Pending: fixture asserts no new stream/write/operation while source/declaration visibility is present. | Pending. | A later typed-lane PR, not this bridge, owns promotion. |
| Public command returns exact missing foundation before I/O | Pending: request/credential-counting commandrunner/app/CLI regression against a generated command. | Pending. | Keep the typed terminal separate from `missing --credential`; it is visibility proof, not runnable proof. |
| All source classes remain visible and evidence rolls up exactly | Pending: per-provider/lane/component/semantic count regression. | Pending. | Generated operation evidence must state visibility and runnable delta separately. |


# Review Disposition — Issue #475 Cycle 9

Frozen reviewed candidate: `0cdcda7e049b7ecfa2fdc52027c66c5de161f2c8`  
Immutable base: `e659d6f1b666f58748e2d8c86599ceb4bbc62ff8`  
Sources: `/tmp/475-REVIEW-CYCLE8-1.md`, `/tmp/475-REVIEW-CYCLE8-2.md`

Both independent reviews are accepted as one blocking correction batch. Their overlapping findings
are deduplicated below; none is declined or deferred outside Cycle 9 except the explicitly
parent-owned physical-workspace identity guarantee.

| Cycle 9 group | Review findings | Disposition / proof boundary |
|---|---|---|
| Creation-result ownership | CR-01, BL-01 | normalize once; one session read; same owned session for validation, execution, cancellation, cleanup, and late creation |
| Private tool oracle | BL-02 | frozen private expected names; distinct immutable Pi arrays; mutation/reorder/replacement cannot validate forbidden tools |
| Deep schema/result snapshots | HI-02, BL-03 | bounded data-only deep clone/freeze before awaits; one-read immutable capability/workspace DTO results |
| Retryable setup settlement | HI-04, BL-04 | explicit fulfilled/rejected/pending state; settled reload/create rejection remains primary and reusable, not quarantining |
| Bounded teardown | HI-01, BL-05 | unsubscribe/dispose independently exactly once and independently bounded; dispose remains reachable; late rejections consumed |
| Exact signal lease | HI-03, BL-06 | capture add/remove operations; native fallback detach after pre-detach throw; request and parent coverage |
| Public typed errors | HI-05, BL-07 | every public async rejection is `AgentSessionRuntimeError` with own `cause`; aggregate primary and cleanup deterministically |
| Terminal event DTO | MD-02, BL-08 | known closed event kinds parsed during delivery; no raw references; bounded keys/arrays/scalars; proxy/accessor/sparse/wide rejection |
| Shared redaction grammar | CR-02 redaction portion, BL-09 | equals multiword, opaque Authorization, URL credentials, implicit flow pairs, malformed/mid escapes, and 63/64/65 worst-case keys across all consumers |
| Credential-bearing paths | CR-02 path portion, BL-10 | bounded case-insensitive path classification denies registry/package/netrc/Git/Kubernetes/cloud/container auth before callback |
| Capability name authority | BL-11 | tokenized sensitive noun + acquisition verb denial in either order, plurals, aliases, and read-only roles |
| Terminal-safe handoff | MD-01, WR-01 | reject HT/LF/CR/CRLF, all C0/C1, Unicode line/paragraph separators, and bidi formatting in every public text field |
| Direct Pi custom-tool contract | WR-02 | exported Pi 0.80.6 `ToolDefinition`/`AgentToolResult`, TypeBox `TSchema`, required `details`, no hiding `unknown` cast, offline no-model exercise |

Cycle 8's bounded disjoint-mutator lease contract remains mandatory. The runtime continues to use
the coordinator-supplied stable `workspace.id` as its physical collision key; proving symlink/case
identity belongs to #479 and requires no #475 scheduler or workspace edit.

Status before RED: accepted, planned, and unimplemented. Fresh stable-head review remains
parent-owned after the complete PLAN -> RED -> GREEN -> evidence chain.

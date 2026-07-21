# Review Disposition — Issue #475 Cycle 10

Frozen reviewed candidate: `f63957aed6fd1406eb3bd9a82adbd10b23b34c33`
Immutable base: `e659d6f1b666f58748e2d8c86599ceb4bbc62ff8`
Sources: `/tmp/475-REVIEW-CYCLE9-1.md`, `/tmp/475-REVIEW-CYCLE9-2.md`

Both complete reports, including WR-01, are accepted as one blocking correction batch. No report
row is declined, weakened, or deferred.

| Cycle 10 group | Review findings | Disposition / proof boundary |
|---|---|---|
| Native signal authority | R1-1, BL-03 | canonical native add/remove always runs; captured hooks remain observable but cannot defeat cancellation or detach |
| Staged returned-session ownership | R1-2, BL-01 | cleanup capsule captures each available operation independently before operational validation; successful forced cleanup retries, actual cleanup failure quarantines |
| Detached timeout ownership | R1-3 | late abort/idle/unsubscribe/dispose deadlines are unreferenced; foreground awaited bounds remain referenced |
| Multi-phase close join | BL-02 | pending creation stays bounded; once cleanup starts, close/shutdown/coalesced close await its internally bounded terminal |
| Exact SDK result capture | R1-4 | creation result, extension result, canonical empty arrays, and fallback are closed one-read data snapshots; malformed containers still clean the exact session |
| Pi cumulative/terminal events | R1-5 | known closed envelopes charge only novel cumulative delta and fully account terminal handoff evidence under joint maxima |
| Prototype-safe DTOs | R1-6 | schema/result keys including `__proto__`, `prototype`, and `constructor` remain own data properties and serialize identically |
| Incremental breadth / closed events | BL-07 | enumerable breadth rejects before full hostile materialization; known terminal kinds reject unknown fields |
| Sanitized public failure graph | R1-7 | SDK/workspace/capability/listener/cleanup failures cross boundaries only as bounded typed redacted snapshots; raw external errors are never retained |
| Remaining redaction grammar | BL-04 | equals assignments, Proxy-Authorization, quoted YAML/flow keys, and OAuth fragments redact through every consumer; harmless colon prose is preserved |
| Sensitive workspace paths | BL-05 | cloud configs/token stores, `.envrc`, and DSA/ECDSA key names reject before workspace callbacks for nested/case variants |
| Capability authority vocabulary | BL-06 | sensitive nouns/compounds and acquisition/display aliases are structurally denied for every role regardless of order/plural |
| Original-text terminal safety | WR-01 | handoff fields reject forbidden controls before redaction, including strings that also contain credentials |

Status after GREEN: all accepted groups are implemented at
`a88cbe5242f070059ea49446ffac6914716a8c5d`; all 102 focused tests and both strict TypeScript scopes
pass. The single RED checkpoint `6df77689` preserved all 86 prior tests, executed 16 independent
new behavior failures, and left production frozen. The complete serialized suite remains honestly
environment-blocked only at the unchanged 31 controller/state-store `/bin/ps` `spawn EPERM` rows;
isolation excluding those files passes 165/165. Parent orchestration owns the permitted-environment
rerun and fresh exact-head review; this lane performed no push or GitHub mutation.

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

Status after GREEN: all accepted groups are implemented at `94918f4e` and the consolidated focused
suite passes 86/86. Strict focused and all-production TypeScript pass against pinned Pi 0.80.6;
the no-model tool row exercises Pi's real argument validator and required result `details`.
Fresh stable-head review remains parent-owned after the evidence commit.

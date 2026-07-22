# Review Disposition — Issue #475 Cycle 15

Frozen reviewed candidate: `f41cde91e01e439a5ebbbaa4867729e0fa80b371`
Immutable base: `e659d6f1b666f58748e2d8c86599ceb4bbc62ff8`
Sources: `/tmp/475-REVIEW-CYCLE14-1.md`, `/tmp/475-REVIEW-CYCLE14-2.md`

Both reports are accepted completely. Review 1's two blockers and Review 2's one blocker are three
unique behavior families; none is declined, deferred, or answered with a case-by-case key/prototype
patch.

| Cycle 15 group | Review findings | Implemented disposition / proof boundary |
|---|---|---|
| Quoted unknown whole-value closure | C14-R1-01 | CLOSED at `38e95460`: every unknown-sensitive quote style goes directly to complete value redaction, never Authorization component parsing; all 13 shared consumers and exact public controls pass |
| Conservative assignment/flow grammar | C14-R2-B1 | CLOSED at `38e95460`: one bounded total candidate parser separates assignment syntax from canonical eligibility; uncertain/cutoff/malformed fields redact and resume later siblings under 25/50/100-KiB work bounds |
| Own-descriptor post-create capture | C14-R1-02 | CLOSED at `38e95460`/`ee4943f4`: non-approved direct prototypes reject without traps; no runtime `for...in` remains; own fields/indexes and callback/barrier seams are split while cleanup stays owned |

The Cycle 14 exact host-capability registry remains closed and retained. Test-only RED `5d83d519`
executed 137 rows with all 134 retained passes and exactly three intended failures while production
remained frozen. Cohesive GREEN `38e95460` and REFACTOR `ee4943f4` pass 137/137 focused and every
declared terminal gate except the unchanged environment-blocked complete-suite process family.
Both source reports were re-read in full after GREEN; no blocker was omitted or narrowed.

# Review Disposition — Issue #475 Cycle 14

Frozen reviewed candidate: `67050a4a3cf62d0d40660de76938ab72ac68ee96`
Immutable base: `e659d6f1b666f58748e2d8c86599ceb4bbc62ff8`
Sources: `/tmp/475-REVIEW-CYCLE13-1.md`, `/tmp/475-REVIEW-CYCLE13-2.md`

Both reports are accepted completely. Review 1's three blockers and Review 2's blocker/warning are
consolidated into three architecture families; no finding is declined, deferred, or answered with
another semantic synonym list.

| Cycle 14 group | Review findings | Disposition / proof boundary |
|---|---|---|
| Post-create lifecycle barriers | R1-F1 | ordered cleanup-root/validation/subscription acquisition with an active-scope assertion after every callback and before every later side effect; abort/close/shutdown matrix proves exact cleanup and zero residue |
| Closed host capability registry | R1-F2, R2-B1 | replace all pattern/token semantic denial with an exported exact safe registry/literal union; reject every unknown string and forged mutability/identity through policy, prompt, and runtime |
| Closed structured-field grammar | R1-F3, R2-W1 | canonical segment parser; exact sensitive paths/terminal compounds, explicit public metadata terminals/paths, every unknown assignment fail-closed through all consumers |

The Cycle 13 C13-01, C13-03, C13-06, and C13-07 closures remain accepted. C13-02 is extended from
pre-create SDK callbacks to every post-create session seam. C13-04 is replaced architecturally
rather than extended enumeratively. C13-05's declared secret cases remain, while ancestor
subsequence matching is removed to close the public-leaf fidelity warning. All 131 retained rows
remain mandatory.

The #479 handoff is closed and narrow: AgentSession host tools are exactly `host_inspect` and
`host_verify`; scoped workspace tools remain separate; scheduler, Git/worktree, GitHub, decision,
review, and integration authority stays in #479's controller-owned adapters. There is no arbitrary
host-name extension mechanism. Status: exact three-row RED captured at `229217f4` with 131 retained
passes and three intended failures; strict TypeScript passes and production is frozen at its
pre-GREEN boundary.

Status after GREEN/refactor: cohesive GREEN `9af22e72` and REFACTOR `27c07eec` close every accepted
finding without adding a dynamic capability escape or a semantic alias list. All three C14 rows
pass with all 131 retained rows (134/134). Both strict TypeScript scopes, the explicit pinned Pi
0.80.6 offline RPC, retained actual no-tool/one-tool sessions, and safe isolation 197/197 pass.
The serialized complete suite executes 271 tests with 240 passes and only the unchanged 31
controller/state-store `spawn EPERM` environment failures. Both complete Cycle 13 reports were
re-read in full after implementation; no blocker or warning is omitted, deferred, or silently
narrowed. Parent orchestration retains the process-capable rerun, exact-head review, integration,
and delivery boundary.

# Review Disposition — Issue #475 Cycle 13

Frozen reviewed candidate: `5dafc5725167bb74ce88a723073b8c4ceb8314e0`
Immutable base: `e659d6f1b666f58748e2d8c86599ceb4bbc62ff8`
Sources: `/tmp/475-REVIEW-CYCLE12-1.md`, `/tmp/475-REVIEW-CYCLE12-2.md`

Both complete reports are accepted as one blocking correction batch. Their seven unique findings
are mapped below; no row is declined, weakened, or deferred.

| Cycle 13 group | Review finding | Disposition / proof boundary |
|---|---|---|
| Bounded dense-array ingress | R1-F1 | remove whole-key materialization; capture only bounded length/index own-data descriptors into a fresh frozen DTO; every non-index peer is untouched and inert |
| SDK-seam terminality | R1-F2 | recheck active scope after every re-entrant SDK callback and immediately before scheduling creation |
| Public policy snapshots | R2-F1 | intrinsically dense-capture authority arrays before all validation, normalization, or formatting |
| Semantic capability grammar | R2-F2 | deny equivalent generic execution/process and protected-data/export authority regardless of order or synonym |
| Qualified sensitive keys | R2-F3 | classify split dotted/qualified sensitive compounds through the shared scanner and every consumer |
| Public prompt snapshots | R2-F4 | intrinsically dense-capture context/authority arrays before validation and serialization; return immutable prompts |
| Cross-event tool identity | R2-F5 | correlate authorized assistant call, execution, result message, turn results, next turn, and final handoff exactly |

Prior-family disposition is explicit: Cycle 7 and 9–10 remain retained; R1-F1 and R2-F1/F4 reopen
the direct-boundary extension of Cycle 8/12 immutable snapshot coverage; R2-F2/F3 reopen the Cycle
11 capability and qualified-key grammar at semantically equivalent forms; R2-F5 extends Cycle
11/12 ordered lifecycle evidence to one complete tool identity. All 124 Cycle 12 rows remain
mandatory. RED was captured at test-only checkpoint `974d2e79`: all 124 retained rows passed,
exactly seven compiled behavior rows fail at their intended boundaries, strict focused TypeScript
passes, and production remains frozen.

R1-F1 cannot literally prove absence of arbitrary hidden strings/symbols under a bounded-work
contract: ECMAScript offers only whole-key materializers for those fields. Cycle 13 therefore
accepts the stronger authority property that is implementable without breaking the public API:
only bounded indexed own-data is copied into a fresh immutable array, while every non-index source
peer is neither read nor preserved and cannot affect policy, prompt, lifecycle, or handoff state.

Status after GREEN/refactor: all seven accepted findings are implemented by cohesive GREEN
`48f546a5` and hardened at `e50b5f97`. The unchanged seven top-level C13 rows now pass together
with all 124 retained rows (131/131), both strict TypeScript scopes and explicit Pi 0.80.6 offline
RPC pass, and safe isolation passes 194/194. Complete Shepherd is environment-blocked at 237/268
only by the same 31 controller/state-store `spawn EPERM` rows. Both source reports were re-read in
full after implementation; no finding is omitted, deferred, or silently narrowed.

# Review Disposition — Issue #475 Cycle 11

Frozen reviewed candidate: `1571dc4d4f45ad4285107d04f2d7c489a7f357ab`
Immutable base: `e659d6f1b666f58748e2d8c86599ceb4bbc62ff8`
Sources: `/tmp/475-REVIEW-CYCLE10-1.md`, `/tmp/475-REVIEW-CYCLE10-2.md`

Both complete reports are accepted as one blocking correction batch. Their unique union is mapped
below; no report row is declined, weakened, or deferred.

| Cycle 11 group | Review findings | Disposition / proof boundary |
|---|---|---|
| Real Pi result compatibility | R1-1 | accept required extension `runtime` by own descriptor as inert compatibility evidence; actual no-model factory/result path cleans the real returned session |
| Native signal operations | R1-2 | invoke only canonical `EventTarget` add/remove operations; shadow hooks cannot attach alternate tuples or defeat rollback |
| Run-associated creation terminal | R2 BL-01 | `abort(runId)` joins terminal creation/cleanup or returns typed pending ownership while preserving observable quarantine |
| Admission/close linearization | R2 BL-02 | admission begins before caller/SDK callbacks; close waits admissions and reservation rechecks closing before create/prompt |
| Stateful bounded Pi streams | R1-3, R2 BL-03 | outer/inner snapshots agree; content families prove actual suffix growth; complete envelope and replacement state consume the aggregate bound |
| Ordered complete terminal DTO | R1-4, R2 BL-04 | exactly one ordered terminal pair; required api/usage and all routing/response/diagnostic/error/content/tool identity fields compare |
| Pre-materialization-safe adapters | R1-5, R2 BL-06 | fixed envelopes use allowlisted descriptors; bounded arbitrary JSON copies enumerable data without whole-source key collection; hidden/symbol peers are inert |
| Total failure sanitizer | R1-7, R2 BL-05 | proxy/prototype traps collapse safely; AggregateError members are manually capped and iterator close is guarded |
| Closed capability grammar | R1-6, R2 BL-07 | concatenated, separated, plural, and mixed forbidden authority compounds deny for every role |
| AWS cache paths | R1-8 | AWS SSO and CLI cache families reject root/nested/case variants before callbacks |
| Cookie header redaction | R2 BL-08 | Cookie and Set-Cookie auth/session values redact through every shared consumer while reviewed harmless prose remains exact |
| Qualified-key redaction | R1-9 | final/compound dotted sensitive key segments redact in equals/colon forms through every shared consumer |

Status after GREEN/refactor: all accepted groups are implemented at `d9b4eaee` with first GREEN
`1e605675`; all 114 focused tests and both strict TypeScript scopes pass. The single RED checkpoint
`c5886520` preserved all 102 prior tests, executed exactly 12 intended behavior failures, and left
production frozen. Complete Shepherd remains honestly environment-blocked only at the unchanged 31
controller/state-store `/bin/ps` `spawn EPERM` rows; isolation excluding those files passes
177/177. Parent orchestration owns the permitted-environment rerun and fresh exact-head review; this
lane performed no push, network, GitHub, Go, connector, service, credential, model, or `make` action.

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

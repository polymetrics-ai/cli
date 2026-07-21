# PRD Coverage — Issue #475

## Cycle 8 Stable-Head Diagnostic

The frozen candidate is `f219b730c63adc9188c93093a40511433a3d0110`; the immutable base remains
`e659d6f1b666f58748e2d8c86599ceb4bbc62ff8`. Cycle 8 uses the accepted issue contract,
`/tmp/475-REVIEW-SECURITY-CYCLE7.md`, and the parent-provided lifecycle disposition as its
phase-equivalent coverage gate.

| Required outcome | RED boundary | Cycle 8 result |
|---|---|---|
| Listener cleanup owns the acquired target despite add/remove throws, parent removal, or signal mutation | independent fake signal leases with reservation/timer/listener accounting | focused GREEN |
| Literal `undefined` rejection/throw is not erased | cleanup and validation reasonless failures | focused GREEN |
| Thenable dispose/unsubscribe is awaited and rejection quarantines | delayed and rejecting thenable hooks | focused GREEN |
| Request, authority, workspace, binding, and signal normalize/freeze once | hostile getters, mutation, reload, mutator-fence, cwd/head/prompt assertions | focused GREEN |
| Disjoint isolated mutators can use bounded concurrency without weakening collision fences | two active canonical authority leases, same-scope denial, capacity denial, per-lease release | focused GREEN |
| Every configurable limit has a hard reviewed maximum | one-above-ceiling table including Node timer maximum | focused GREEN |
| Comma-bearing line/Auth parameters redact completely | direct plus all shared-redactor consumers | focused GREEN |
| Multiline-flow key-only/continued scalars redact | mapping and sequence cross-products | focused GREEN |
| Escaped quoted secret keys classify after bounded decode | JSON escapes, YAML doubled quote, malformed fail-closed controls | focused GREEN |
| Event accounting is bounded and cycle-safe before materialization | oversized, deep, cyclic event probes | focused GREEN |
| Tools, prompt, and handoff share canonical prefixes | trailing/redundant separator compatibility | focused GREEN |
| Handoff strings cannot carry terminal controls | summary/finding/verification C0/C1 matrix | focused GREEN |
| Prior behavior and declared lane gates remain intact | 53 retained focused tests; full Shepherd, strict Pi 0.80.6 TS, offline RPC, diff/base/head/scope | 70/70 focused, both strict scopes, RPC, diff/base/scope pass; full suite and push environment-blocked |

No dependency, Go, connector, CLI/help/docs/website, runtime-service, live-GitHub, review-bot,
merge, or parent-artifact work is required or authorized.

## Cycle 7 Stable-Head Diagnostic

Issue #475 remains a narrow Pi AgentSession runtime slice under parent issue #471. The repository
program PRD is connector-focused, so this phase's accepted issue contract, exact-head review
findings, and existing PLAN define the phase-equivalent coverage gate. The frozen candidate is
`a3cd85a5d0871dd1c4c99dd8b30bcd609a228c45`, the immutable base is
`e659d6f1b666f58748e2d8c86599ceb4bbc62ff8`, and the 11-finding campaign source is
<https://github.com/polymetrics-ai/cli/pull/486#issuecomment-5037079867>.

| Required outcome | Artifact / test boundary | Status before Cycle 7 execution |
|---|---|---|
| Throwing external-signal attach/remove cannot strand an admitted run | two independent lifecycle rows with close, timer, reservation, and hook accounting | focused GREEN |
| Successful close never precedes owned late creation work | abandoned create resolves or rejects after close begins | focused GREEN |
| An uncancellable create cannot hang or falsely satisfy close | pending create at bounded close; quarantine and later-dispatch assertions | focused GREEN |
| Malformed late fulfillment is consumed and fails closed | abandoned create resolves malformed; close and `unhandledRejection` assertions | focused GREEN |
| Multiline outer flow and indented/key-only/continued YAML cannot hide sensitive values | shared direct and consumer payloads | focused GREEN |
| Numeric secrets and all Authorization schemes redact; unmatched quotes recover | direct and serialized consumer payloads | focused GREEN |
| Shepherd repository aliases and generic PKCS#8 are recognized | environment/path vocabulary plus `BEGIN PRIVATE KEY` payloads | focused GREEN |
| Safe multiline quoted assignment prose remains byte-identical | direct preservation control | pass |
| Total scanner work is near-linear with leading padding and dense flow assignments | deterministic 25/50/100 KiB diagnostics, including key-start work | focused GREEN |
| Every redaction form reaches every relevant trust boundary | direct, prompt, `workspace_read`, typed capability, handoff summary/finding | focused GREEN |
| Prior lifecycle and redaction invariants remain intact | existing 40 focused regressions | mandatory retained passes |
| Declared phase verification | focused/full Shepherd tests, pinned Pi 0.80.6 strict TypeScript, offline RPC, diff/base/head/scope | pass: 53/53 focused, 190/190 full, both strict scopes, RPC, and scope gates |

No dependency, CLI/help/docs/website, Go, connector, runtime-backed service, live credential, or
external mutation work is required. Parent orchestration owns the stable-head review campaign and
integration after this worker returns a clean pushed head; this lane does not merge parent commit
`2a89142e` or edit shared parent artifacts.

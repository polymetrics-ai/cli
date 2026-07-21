# PRD Coverage — Issue #475

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

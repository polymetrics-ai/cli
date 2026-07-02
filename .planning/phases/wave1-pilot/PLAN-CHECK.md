# PLAN-CHECK — wave1-pilot

Verdict: **APPROVED-WITH-FLAGS**

Checked against: `.planning/ROADMAP.md` (wave1-pilot acceptance), SPEC.md, PLAN.md, TEST-PLAN.md,
API-CONTRACT.md, EVAL-PLAN.md, DECISIONS.md, OBSERVABILITY.md, RUNBOOK.md, PRD-COVERAGE.md,
wave0-engine-harness/SUMMARY.md, and the live codebase (`internal/connectors/engine/hooks.go`,
`internal/connectors/defs/defs.go`, `cmd/connectorgen/gen.go`, `internal/connectors/hooks/hookset/
hookset_gen.go`, `internal/connectors/conformance/conformance_test.go`, legacy packages for
zendesk-support/monday/gmail/github).

No blockers. Two flags (should-fix, non-blocking) and one info item below.

---

## 1. Requirement coverage — PASS

Every ROADMAP wave1-pilot acceptance bullet maps to a task with a runnable verify command:

| # | ROADMAP bullet | Task(s) | Verify command |
|---|---|---|---|
| 1 | All 10 `migrated`/typed-blocked, self-check green | P-1..P-10 | conventions.md §7 block (`connectorgen validate`, `go build`+`go vet`, `TestConformance/<name>`, `paritytest/<name>`) |
| 2 | Per-connector parity green | P-1..P-10, P-14 | `go test ./internal/connectors/paritytest/... ./internal/connectors/engine -run 'TestParity\|Test.*Parity'` |
| 3 | Wave gate green (path guard, gen, build/test/lint) | P-14 | RUNBOOK.md exact sequence (git status --porcelain → connectorgen gen → registryset diff empty → build/vet → test → conformance → race → make verify) |
| 4 | Fable 100% line-by-line review, zero unresolved blockers | P-11 | `review.schema.json` verdicts → `traces/review-<name>.json`, repair-then-quarantine loop |
| 5 | conventions.md + executor template patched | P-12 | diff review of the two files |
| 6 | `pilot-costs.json` written | P-13 | file inspection vs `traces/cost-log.jsonl` dispatch count |
| 7 | Pass B decision made with user | P-14 step 4 | DECISIONS.md entry (human gate, not automated) |

Also cross-checked EVAL-PLAN.md's 10 exit metrics — all trace to the same task set, no orphaned
metric. No ROADMAP requirement is dropped or partially covered.

## 2. Wave0 carried flags — PASS, all accounted for

| Flag | Disposition | Where |
|---|---|---|
| N1 (`formatCursorForAssertion` github_date_range) | Fixed, red-first TDD | P-0 |
| N4 (stale doc comment) | Batched into P-0 | P-0 |
| N2 (all-digits non-unix guard) | Bundled into P-0 per DECISIONS #3 | P-0 |
| N3 (relative next-URL fail-closed) | Noted-not-blocking; bitly/calendly confirmed to use absolute URLs (spot-checked legacy comments) | SPEC §4, P-3/P-4 task rows |
| N5 (`..%5C` residuals) | Noted-not-blocking, correctly scoped (no pilot interpolates untrusted values into paths beyond config/record ids) | SPEC §4 |
| gmail 3-legged OAuth open question | Resolved with evidence: verified `gmail/auth.go` (127 loc) implements ONLY the refresh-token grant, no 3-legged dance in legacy — confirmed by direct read, matches SPEC §5.7 claim exactly | SPEC §5.7, P-10 |
| 11MB blob in history | Deferred to pre-push per DECISIONS #4 | RUNBOOK.md |
| conformance pkg coverage 83.1% | No gate this phase (informational) | — |
| searxng subreddit narrowing | wave6 human-gate item, correctly out of scope here | — |

Nothing dropped.

## 3. `paritytest/<name>` package decision and the hookset_gen.go collision question — VERIFIED SAFE, no collision

Traced the actual collision risk end-to-end against the live codebase, not just the plan's prose:

- `cmd/connectorgen/gen.go`'s `genHookset` regenerates `hooks/hookset/hookset_gen.go` by scanning
  **directory names** under `internal/connectors/hooks/` (`wiringPackageNames`) — it does not run
  during any agent's task, only at P-14 (`go run ./cmd/connectorgen gen`, orchestrator-only,
  explicitly listed FORBIDDEN for agents in conventions.md §7 and reiterated in RUNBOOK.md's DW-1
  step: "do NOT run `go run ./cmd/connectorgen gen` ... from any agent").
- The three hook-authoring agents (monday/github/gmail, P-8/P-9/P-10) each write ONLY to their own
  `internal/connectors/hooks/<name>/` directory — never to `hooks/hookset/hookset_gen.go` itself.
  Confirmed: today only `hooks/hookset/` exists (empty scaffold); `monday/`, `github/`, `gmail/`
  subdirectories do not exist pre-pilot, so there is no pre-existing file three agents contend to
  edit — each creates a brand-new, disjoint directory.
- Parity tests do **not** depend on `hookset_gen.go` at all: SPEC §6 states "Tier-2 pilots
  blank-import their own `hooks/<name>` package from the parity test to trigger
  `engine.RegisterHooks` init." Verified this mechanism is real and sufficient: `hooks.go`'s
  `RegisterHooks(name, factory)` is called from a hooks package's own `init()`
  (`hookset_gen.go`'s own doc comment: "package so their init() functions run engine.RegisterHooks
  side effects" — the identical blank-import mechanism Go's `_ "path"` import performs, whether
  triggered by `hookset_gen.go` or by a test file directly importing `hooks/<name>` the same way).
  A parity test file doing `import _ "polymetrics.ai/internal/connectors/hooks/monday"` runs that
  package's `init()` -> `RegisterHooks("monday", ...)` with zero dependency on the generated
  wiring file.
- `defs.FS`'s `//go:embed all:*` (`internal/connectors/defs/defs.go`) confirmed to auto-discover
  any new `defs/<name>/` subdirectory with **no edit to `defs.go` itself** — genuinely disjoint
  file sets for the declarative side too.
- Net: the 10 DW-1 tasks' writable sets truly do not intersect. The only shared artifact
  (`hookset_gen.go`) is touched exactly once, by the orchestrator, at P-14 — after all 10 agents
  are done — exactly as PLAN.md's dispatch table and RUNBOOK.md's gate sequence specify. This is
  the correct fix and it is already in the plan; no revision needed here.

## 4. Per-task self-verify commands — spot-checked, runnable as written

- `go test ./internal/connectors/conformance -run 'TestConformance/<name>'` — verified
  `TestConformance` in `conformance_test.go` does `t.Run(b.Name, ...)` over `engine.LoadAll(defs.FS)`
  bundles, so the subtest path is exactly `TestConformance/<connector-name>` as every task assumes.
- `go test ./internal/connectors/paritytest/<name> -v` (a hook connector: monday) — the directory
  doesn't exist yet (expected, pre-execution); the pattern matches wave0's `native/postgres/
  parity_test.go` precedent (external test package driving both legacy and engine-backed
  connectors against one httptest server) and `engine.Load(defs.FS, name)` is a real, used API
  (confirmed in `parity_stripe_test.go`). Runnable once P-8 authors the file.
- `make verify` — confirmed composition (`fmt tidy-check vet test build docs-check smoke lint
  connectorgen-validate`) in the Makefile; matches every task's reference.
- zendesk-support dual-auth claim (P-7) — verified directly against
  `internal/connectors/zendesk-support/zendesk_support.go` and its test file: Basic
  `<email>/token:<api_token>` AND OAuth Bearer are both real, legacy resolves whichever secret is
  present preferring OAuth — SPEC §5's table is accurate.
- monday GraphQL in-body pagination claim (P-8) — verified `next_items_page` cursor field and the
  `query { next_items_page (limit: %d, cursor: %q) ... }` template in `monday.go` — accurate.
- gmail refresh-token-only claim (P-10) — verified `gmail/auth.go`'s `oauthRefreshAuth` type and
  doc comment match SPEC §5.7 verbatim.

All spot-checked commands/claims are accurate and runnable.

## 5. github task (P-9) sizing — FLAG (should-fix, non-blocking)

Legacy github package: 2712 loc of production Go (`github.go` 1980 + `auth.go` 295 + `streams.go`
352 + `manifest.go` 85) + ~1153 loc of tests, 33 stream/schema definitions in streams.go
(SPEC/inventory say 19 streams — the 33 count includes non-stream helper struct literals; not a
contradiction, just note the raw grep isn't the stream count), 16 write actions, two Tier-2 hook
interfaces at the ≤300-line cap. This is genuinely the largest, highest-risk single task in the
wave (SPEC itself calls it "XL, highest risk" and PLAN.md flags it as "the long pole").

The plan's mitigations are real and reasonable: a stated "parity floor" (§5.6) allows `partial`
status with typed blockers for any write action that doesn't fit the hook budget, DECISIONS.md #1
dispatches it FIRST for maximum wall-clock runway, and DECISIONS.md #2 assigns it to the
stronger-scrutiny review batch. Given the `partial`-with-typed-blockers escape valve is explicit
and acceptable per EVAL-PLAN metric 1 ("10/10 ... OR partial/blocked ONLY with typed blockers"),
a single-agent attempt is a reasonable bet for a pilot (the pilot's purpose is partly to learn
whether XL connectors need pre-splitting for wave3). Not a blocker, but flag for the coordinator:
if P-9 reports `partial` with more than ~3 unported write actions, that's a strong signal wave3's
XL bucket (1/agent per ROADMAP) needs a sub-task split convention, not just a hook-budget escape
valve — capture this explicitly as a candidate P-12 learning if it happens.

**Minor data inconsistency (info only)**: SPEC.md/inventory.json list github at "3664 loc"; this
figure does not match either the production-only line count (2712) or the total-with-tests count
(3865) from a direct `wc -l`. This is almost certainly inventorygen's own counting convention
(e.g., excluding blank lines/comments or a different test-file inclusion rule) and is internally
consistent across SPEC/PLAN/inventory.json, so it does not affect task sizing decisions made from
it. No action needed beyond awareness.

## 6. Cost-capture (P-13) mechanics — PASS

OBSERVABILITY.md specifies the coordinator (not agents) records role/model/connector/wall-clock/
token counts "from the agent-run usage stats" at each dispatch boundary into
`traces/cost-log.jsonl`, then P-13 aggregates into `pilot-costs.json`. This mirrors wave0's already-
proven method (SUMMARY.md's "Numbers for the pilot cost model" section shows this exact capture
already happened once). The mechanism is implementable by a Claude-Code-style coordinator loop that
receives usage stats in each agent dispatch's return envelope; no undefined data source.

## 7. TDD pairing + red-first protocol — PASS

Every `[behavior]` task has a paired red-first test task per the Test-pairing summary table:
P-0 (new conformance self-test bundle, RED before fix), P-1..P-10 (parity test written first,
RED on missing bundle, plus a required MEANINGFUL red→green transition beyond the trivial load
failure — TEST-PLAN.md §5 explicitly guards against only capturing the trivial RED). Hook unit
tests (`hooks_test.go`) are additionally required for the three Tier-2 pilots. P-11/12/13/14 are
correctly exempted as process/docs tasks with no production behavior, recorded as such in
TDD-GATE.json per the plan's own instruction — this addresses wave0's B3 lesson (empty-array gate
files) directly and by name.

## 8. Dependency graph / wave structure — PASS

DW-0 (P-0, serial) -> DW-1 (P-1..P-10, 10-way parallel, disjoint dirs, verified in §3 above) ->
DW-2 (P-11 + repairs) -> DW-3 (P-12, P-13, parallel, different files: conventions.md/prompts doc
vs pilot-costs.json) -> DW-4 (P-14, serial single-writer). No cycles, no forward references, wave
numbers consistent with `depends_on` semantics. DECISIONS.md #1's stagger (github+monday first)
refines but does not contradict PLAN.md's softer "if staggering is needed, start P-9 and P-8"
language — same two tasks, DECISIONS.md is the authoritative, more specific instruction (minor:
PLAN.md's wording wasn't tightened to match, cosmetic only).

## 9. Scope sanity — PASS with expected variance

P-1..P-10 are each single-connector, single-task dispatches (not 4-5 tasks per plan as the
generic threshold table would flag) — appropriate given SPEC's explicit shared task shape and the
1-agent-per-connector design. P-9/P-10/P-8 carry extra hook directories but stay within their own
task. P-14 is a single serial gate task by design (single-writer wave close) — appropriately not
parallelized.

## 10. Context compliance (DECISIONS.md) — PASS

All 5 coordinator decisions are correctly reflected: #1 (stagger) in RUNBOOK.md dispatch step 3
and PLAN.md's sequencing note; #2 (2 reviewer batches, hook-heavy grouped together) matches P-11's
description of 5+5 split reasoning; #3 (N2 bundled into P-0) confirmed in P-0's task description;
#4 (11MB blob deferred to pre-push) in RUNBOOK.md; #5 (monday Tier-2 StreamHook approach, no
engine change) matches SPEC §5.5 and PLAN.md's P-8 row exactly. No contradictions found; no
deferred-idea scope creep (Pass B / wave5 surface expansion, wave6 registry flip, and go.mod
changes are all explicitly excluded in SPEC §1 "Out of scope").

## Top findings

1. **[VERIFIED SAFE]** The suspected `hookset_gen.go` three-way collision does not occur: the file
   is orchestrator-only, regenerated once at P-14 from directory names, and parity tests bypass it
   entirely via direct blank-import of `hooks/<name>` (confirmed against live `RegisterHooks`/
   `gen.go`/`defs.go` code, not just plan prose).
2. **[FLAG, non-blocking]** P-9 (github) is a large single-agent bet (2712 loc legacy, 16 writes,
   2 hook interfaces at cap); mitigated by the plan's own `partial`+typed-blocker escape valve and
   first-dispatch stagger. Watch for ≥3 unported write actions as a wave3 XL-splitting signal.
3. **[INFO]** SPEC/inventory's github "3664 loc" figure doesn't match a direct `wc -l` of either
   production-only (2712) or total-with-tests (3865) — almost certainly a tooling convention
   difference, internally consistent, no action needed.
4. **[INFO]** PLAN.md's optional stagger wording for P-9/P-8 is softer than DECISIONS.md #1's firm
   mandate; DECISIONS.md correctly supersedes it, cosmetic only.

No revision required. Plan is ready for execution.

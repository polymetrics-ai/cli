# Agent Trace: reviewer

## Rendered Prompt Or Prompt Reference

TBD

## Files Inspected

- TBD

## Actions Taken

- TBD

## Commands Run

- TBD

## Findings

- TBD

## Handoff Summary

TBD

## Verification Evidence

TBD

## Unresolved Risks

- TBD

---

# Batch B review (xkcd, vitally, bitly, calendly, zendesk-support) — 2026-07-02, HEAD 48cbff5

## Commands run
- `go test -count=1 ./internal/connectors/paritytest/{xkcd,vitally,bitly,calendly,zendesk-support}` — all ok (uncached)
- `go test -count=1 ./internal/connectors/conformance -run 'TestConformance/(xkcd|vitally|bitly|calendly|zendesk-support)'` — PASS x5
- `go run ./cmd/connectorgen validate internal/connectors/defs` — 13 connectors, 0 findings
- `go build ./internal/connectors/... && go vet ./internal/connectors/...` — clean

## Sources read in full
defs/{xkcd,vitally,bitly,calendly,zendesk-support}/** (every file), paritytest/{same}/,
traces/p{1,2,3,4,7}-*-ledger.md, legacy internal/connectors/{same}/ packages, engine
read.go/paginate.go/interpolate.go/auth.go(spot)/connsdk/{paginate,http}.go (targeted),
docs/migration/{conventions.md,review.schema.json}, phase SPEC/TEST-PLAN/TDD-GATE/VERIFICATION.

## Findings
Full severity-classified list in `.planning/phases/wave1-pilot/REVIEW-B.md`. Verdicts:
xkcd FAIL (schema drops legacy passthrough fields alt/img/link/news/transcript; undocumented;
fixtures not real-shape), vitally PASS (1 minor), bitly PASS (3 minors incl. false docs claim
about size-param-on-first-page-only — engine provably re-sends it on next_url pages),
calendly FAIL (dropped derived `id` field adjudicated NOT acceptable under §5 meta-rule; plus
undocumented page_size de-facto-required regression, dead spec keys, base-pagination on the
single-object stream), zendesk-support FAIL (invented `updated_at[gte]` incremental param legacy
never sent and Zendesk does not document — behavior addition misfiled as acceptable; has_more
stop-signal divergence; dead spec keys).

## Key adjudications
1. calendly `id`-drop: NOT acceptable; P-12 `last_path_segment` filter (RecordHook only interim).
2. Optional/config-driven query params: ≥5 occurrences — mandatory P-12 ENGINE_GAP increment;
   opt-in per-param optional-query field (+ optional `default`), NOT blanket absent-key-falsy.
3. zendesk incremental: behavior addition, not scope narrowing — remove; Pass B via real
   incremental-export endpoints. TEST-PLAN row was a planning error; plan-vs-legacy conflicts
   must escalate, not be satisfied by invented API params.

## Gate verdict
NO-GO as-is. xkcd/calendly/zendesk-support to gap loop; vitally/bitly pass. No gate weakening
involved in any fix; no human gate triggered.

---

# Batch A review (github, gmail, monday, sentry, chargebee — hook-heavy) — 2026-07-02, HEAD 48cbff5

## Rendered prompt or prompt reference
P-11 adversarial-reviewer dispatch, batch A override: 100% line-by-line, hook-heavy set; three
named adjudications (computed_fields stringification; parity-test field stripping principle;
github `repository` marker / ENGINE_GAP G0); line-cap adjudication for github 363 / monday 340.

## Files inspected (full reads unless noted)
- defs/github/** (streams.json, writes.json all 25 actions, spec, metadata, api_surface, docs.md,
  schemas: commits/issues/releases/milestones + spot others, fixtures: check, commits, issues 2p,
  writes: close_issue/create_pull_request/update_pull_request/close_pull_request + spot others)
- defs/gmail/**, defs/monday/**, defs/sentry/**, defs/chargebee/** (all core files + schemas +
  fixtures spot-checked for wire realism and numeric-cursor shape)
- hooks/{github,gmail,monday,sentry}/hooks.go in full; hooks_test.go function inventories +
  targeted reads
- paritytest/{github,gmail,monday,sentry,chargebee}/parity_test.go (github/gmail in full; monday/
  sentry/chargebee structure + every normalization/strip helper + incremental tests in full)
- LEGACY (read-only): github/{github,streams,auth}.go in full for writes/auth/read paths;
  gmail/{auth,gmail,streams}.go; monday/{monday,streams}.go; sentry/sentry.go targeted;
  chargebee/{chargebee,streams}.go targeted (harvest/sort_by, baseURL/site, secret)
- ENGINE: read.go (Read/readDeclarative/newRuntime/computed_fields/filters/cursor), write.go in
  full, interpolate/bundle targeted; conformance/dynamic.go+replay.go (write shape subset-match,
  capture server, marker handling)
- Phase docs: PLAN.md, SPEC.md §5.3-5.7 §6, TDD-GATE.json, TDD-LEDGER.md, VERIFICATION.md,
  SUMMARY.md, traces/p5,p6,p8,p9,p10 ledgers + r3-ledger + coordinator trace (stub),
  docs/migration/{conventions.md,review.schema.json}, REVIEW-B.md (for A2 consistency)

## Commands run (all green unless noted)
- `go run ./cmd/connectorgen validate internal/connectors/defs` → 13 connectors, 0 findings
- `go build ./...` → clean
- `go test ./internal/connectors/conformance -run TestConformance` → ok
- `go test -count=1 ./internal/connectors/paritytest/{github,gmail,monday,sentry,chargebee}
  ./internal/connectors/hooks/...` → all ok, uncached
- independent secret-shape grep (sk_live/ghp_/AIza/Bearer-long/PEM/password=) over batch-A defs,
  hooks, paritytest → clean (only a documented empty-password Basic-auth doc line)

## Findings
Full severity-classified report: `.planning/phases/wave1-pilot/REVIEW-A.md`. Verdicts: all 5 PASS
(0 blockers; majors per connector: github 3, gmail 2, sentry 1, chargebee 2, monday 0). Notable
majors found by legacy diffing that no test/ledger caught: github auth_type+secret-alias surface
drop (silent unauthenticated fallback), github color '#'-strip drop, github state-cursor
incremental = new behavior misdocumented as "matches legacy exactly", gmail stale post-R3 docs
bullet, sentry hostname / chargebee site dead-config with FALSE docs claims of legacy base-URL
derivation, chargebee missing incremental sort_by[asc]=updated_at.

## Adjudications (full reasoning in REVIEW-A.md)
- A1 stringification: NOT fan-out-acceptable; typed-extraction engine increment required (>=3
  recurrence met: chargebee+gmail+github). Pilots stand as documented deviations.
- A2 strip principle: 4-condition rule (ledgered / genuinely-unproducible / enumerated-visible /
  companion-pinned); batch A compliant; consistent with REVIEW-B's calendly FAIL.
- A3 github repository marker: right pilot call, blocker-class for fan-out; wire Config (never
  Secrets) into applyComputedFields Vars, restore field.
- C1 line caps 363/340: accept with conventions wording change (soft 300 / hard 400 / >400 or 3rd
  interface = Tier 3).
- C2 skip_dynamic markers: verified honest (nothing skipped that replay could genuinely run).

## Handoff summary
GO for batch A connectors with gap-loop fixes (docs/ledger/spec corrections + 2 test additions —
no gate weakening anywhere). NO-GO for phase close until P-14 fills the hollow gate artifacts
(TDD-GATE.json passed:true with zero task rows, TBD SUMMARY, evidence-free VERIFICATION.md — C4).
NO-GO for Pass A fan-out until the combined engine mini-wave (A1+A3+conditional-query+C3 base-URL
decision) and the P-12 conventions pass land.

## Verification evidence
See "Commands run" — reviewer re-ran the full gate uncached rather than trusting ledger
transcripts; ledger RED evidence cross-checked against code for plausibility (embed-pattern,
undefined-symbol, unresolved-key signatures all consistent).

## Unresolved risks
- C3: no spec-default materialization anywhere → every migrated connector requires explicit
  base_url that legacy defaulted; hard-blocks wave6 flip if undecided.
- Harness: conformance fixtureResponse lacks header support + spec-default-aware config synthesis;
  sentry's markers should be revisited (not grandfathered) if that lands.
- github incremental state filtering changes destination-visible record sets for
  incremental_append (non-deduped) syncs at flip time; needs an explicit product decision note.

---

# Re-review + phase-close trace (gap-loop cycle 1, P-14) — gsd-loop-reviewer (Fable)

Date: 2026-07-02. HEAD f7632b9165fd87623105d93015c608d51b36d6e3. Scope: `git diff d96253a...HEAD`.

Method: read every repaired file directly (defs/{xkcd,calendly,zendesk-support,github,gmail,
sentry,chargebee,vitally,bitly}, hooks/github, paritytest/{xkcd,calendly,zendesk-support,github,
gmail,chargebee,bitly}, engine/{read,paginate,interpolate,bundle,schema}.go + tests,
cmd/connectorgen/validate.go, conventions.md) against the prior findings; never trusted ledgers
for dispositions. Audited every removed test assertion in the diff (`git diff -- '*_test.go'` +
grep for removed Fatalf/func Test) — zero weakenings; all removals are strengthenings or the
mandated zendesk inversion. Cross-checked 4 red-evidence transcripts against shipped format
strings (s2-xkcd, s2-zendesk, gaploop-s1 item 5, p4-calendly embed error) — all verbatim.
Independently verified: engine passthrough projection support (bundle.go:182, read.go:596-614);
buildInitialQuery Cursor="" ordering (read.go:409-441) confirming the chargebee sort_by STOP;
EvalWhen ==/in vs ResolveCheck bare-ref parsing (interpolate.go:363-395, 513-540) confirming the
public_access adjudication; app-layer end-of-sync-only cursor persistence (internal/app/app.go)
bounding the sort_by order-delta risk; hooks/github line count exactly 400 (at the new hard
ceiling).

Gate re-run at HEAD (transcripts in VERIFICATION.md): go build clean; go test -count=1
./internal/connectors/... ./cmd/... exit 0 (582 ok / 0 FAIL); connectorgen validate 13/0;
make lint 0 issues.

Verdict: GO for phase close as completed_with_warnings; 16 findings resolved; 2 ENGINE_GAP items
carried as pre-fan-out blockers + 7 carried minors (full detail in REVIEW-A.md "Re-review (gap
loop cycle 1)" and SUMMARY.md). Artifacts written this pass: TDD-LEDGER.md (index + verification
notes), TDD-GATE.json (14 task + 14 behavior rows), SUMMARY.md, VERIFICATION.md, RUN-STATE.json,
REVIEW-A.md re-review section.

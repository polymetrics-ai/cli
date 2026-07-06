# Phase Summary

Phase: wave1-pilot
Status: COMPLETE (closed with warnings — 2 queued engine items, see "Carried queue").
HEAD at close: f7632b9165fd87623105d93015c608d51b36d6e3 (branch connector-architecture-v2).
Written at P-14 by gsd-loop-reviewer, 2026-07-02.

## What the pilot delivered

**10/10 pilot connectors migrated to declarative bundles at legacy parity** (xkcd, vitally, bitly,
calendly, sentry, chargebee, zendesk-support, monday, github, gmail), each with:
- a `defs/<name>/` bundle (spec/streams/schemas/fixtures/docs/api_surface/metadata),
- a `paritytest/<name>/` package driving BOTH the legacy connector and the engine bundle against
  shared httptest servers with raw-equality record comparison,
- hook-aware conformance coverage (R-3) with honesty-verified `skip_dynamic` markers where replay
  is genuinely inexpressible (gmail, monday, sentry) and full dynamic coverage elsewhere (github
  incl. AuthHook + WriteHook in replay),
- per-task TDD trace ledgers with verified red-first evidence (TDD-LEDGER.md index).

Tier-2 escape hatches stayed within cap: github (AuthHook+WriteHook, 400 lines — at the hard
ceiling), monday (StreamHook+CheckHook, 340), sentry (StreamHook, 281), gmail (AuthHook, 267).
chargebee/calendly/zendesk-support/xkcd/vitally/bitly are pure Tier-1.

## Review outcomes (P-11) and gap-loop cycle 1

- REVIEW-A (batch A, hook-heavy): 5/5 pass with 8 majors + adjudications A1 (stringified
  computed_fields NOT fan-out-acceptable), A2 (parity-strip rules), A3 (repository marker G0),
  flags C1 (line caps), C3 (base-URL defaults), C4 (hollow gate artifacts), C5 (fan-out readiness).
- REVIEW-B (batch B, adversarial): xkcd/calendly/zendesk-support FAIL (schema-projection field
  drop; dropped derived `id` PK + page_size hard-error; invented `updated_at[gte]` incremental +
  has_more stop divergence); vitally/bitly pass with minors.
- Gap-loop cycle 1, Step 1 (engine mini-wave, dc7ad63): typed computed_fields extraction for bare
  `{{ record.path }}` templates; `config.*` wired into computed_fields Vars (Secrets excluded by
  design); optional-query dialect (`{template, omit_when_absent, default}`); `last_path_segment`
  filter; token_path cursor `stop_path` + loop guard; spec.json `default` materialization into
  RuntimeConfig (C3) + `default_type_mismatch` validate rule; conventions.md consolidated updates.
- Gap-loop cycle 1, Step 2 (pilot repairs, f7632b9): all 3 FAIL verdicts repaired and all 8
  REVIEW-A majors closed (re-review confirms; see REVIEW-A.md "Re-review (gap loop cycle 1)").
  Native types restored fleet-wide (chargebee ~30 fields, gmail 4, github 4 — schemas retightened,
  masking test helpers deleted); calendly `id` PK restored via last_path_segment; zendesk invented
  incremental removed + stop_path parity; github fail-loud auth (`public_access` opt-in gating
  `mode:none`), repository marker on all 19 streams, label color-strip, since-param honesty +
  parity tests both paths.

## Gap-loop learnings (carried into conventions.md / fan-out brief)

1. Legacy is ground truth over TEST-PLAN — a plan-vs-legacy conflict is an escalation, not
   something to satisfy by inventing API surface (zendesk `updated_at[gte]`).
2. Schema source must be the legacy RECORD-SHAPING function (or passthrough when legacy passes
   through raw) — Catalog() field lists are an attractive wrong source (xkcd).
3. JSON type IS record data: typed extraction is required, string-widening schemas is not an
   acceptable fan-out pattern (A1; now closed by the engine increment).
4. Parity-test strips/coercions must be ledgered, genuinely unproducible, enumerated, and pinned
   (A2); producible-field strips are FAIL (calendly stripDerivedID, now deleted).
5. Green harness output is necessary but not sufficient — fixture shape can mask real-API
   divergence; reviews must read the wire shape against the real API docs.
6. Dual-auth candidate order is load-bearing (zendesk golden pattern, now conventions §3).
7. ENGINE_GAP recurrence >=3 triggers a mini engine wave — happened once (6 items landed in S-1);
   two further engine items were discovered DURING repairs and are queued (below), not silently
   worked around.

## Costs (docs/migration/pilot-costs.json, coordinator-recorded)

- 10 migrated / 0 blocked / 0 quarantined; total executor tokens 2,721,525; avg 272,153/connector.
- By shape: Tier-1 S ~160k, Tier-1 M ~246k, Tier-2 hooked ~355k, XL github 503k (long pole,
  43 min). Wave wall-clock ~45 min for the 10-way parallel fan-out.
- Overheads: planner 136k, plan-checker 95k, P-0 follow-ups 184k.
- Pass A projection revised: 544 remaining connectors, naive ~147M subagent tokens; bundled
  (S=12/M=7/L=4/XL=1, shared-context amortization) ~75-90M across ~77 bundle agents. The wave0
  estimate of 30M was optimistic. Pass B cost is unmeasured — measure on 3-5 connectors early in
  wave5 before committing the full roster.

## Engine increments landed this phase

wave0-carryover P-0 (github_date_range assertion alignment, start-date warning rule) + R-3
(hook-aware conformance) + gap-loop S-1 (7 items above). `connectorgen validate` grew
`default_type_mismatch`. Meta-schemas needed no change (verified, not assumed — s1 ledger item 7).

## Carried queue (MUST land before Pass A fan-out dispatch)

1. **ENGINE_GAP — incremental-lower-bound query vars** (chargebee `sort_by[asc]=updated_at`):
   `buildInitialQuery` resolves `stream.Query` templates with `Cursor=""` BEFORE computing the
   incremental lower bound (engine/read.go:409-441), so a param legacy sends only-when-incremental
   is inexpressible; the S-2 agent correctly STOPped rather than shipping a config-gated
   approximation that would silently omit the param on every state-cursor sync. Documented OPEN in
   defs/chargebee/docs.md Known limits + traces/s2-chargebee-sentry-ledger.md. Risk today is
   bounded (app persists cursor only at successful sync completion — internal/app/app.go — so
   order cannot cause resume data loss), but the "extra param only-when-incremental" legacy
   pattern will recur at fan-out scale, and any future mid-sync checkpointing would convert the
   order delta into a data-loss risk.
2. **ENGINE_GAP — ResolveCheck `when`-grammar (`==`/`in`) static parsing**: runtime `EvalWhen`
   supports equality/membership but `ResolveCheck`/`ResolveCheckAuthSpec` (connectorgen validate +
   conformance interpolations_resolve) parse only bare `namespace.key` truthiness refs — an
   `auth_type == 'public'`-shaped `when` fails validate. github's `public_access` boolean opt-in is
   the sanctioned interim rendering (fail-loud, documented, parity-tested); fan-out needs ONE
   sanctioned pattern for legacy auth-mode enums, not ~500 improvised booleans. Documented in
   defs/github/spec.json + docs.md (ledger G14) + traces/s2-github-gmail-ledger.md.

Carried minors (non-blocking, fold into P-12/fan-out template pass): chargebee parity's blanket
`normalizeRecordsStringify` main-compare should flip to raw DeepEqual now that types match (its
justifying deviation is RESOLVED; comment also cites the old test name); github compound-write
test still compares method/path only (bodies compared in all non-compound tests); github
close_issue/close_pull_request fixture expect.body omit `state`; monday `max_pages` consumed but
undeclared in spec.json + permissive-parse note; sentry inert static `per_page` query entries +
check-per_page=1 note; gmail hooks interpolateOptional comment claims CRLF/unknown-filter errors
propagate but code returns "" on any error; github hooks.go sits exactly AT the 400-line hard
ceiling (zero headroom for any future fix).

## Verification at close

See VERIFICATION.md: build, full uncached test suite, connectorgen validate (13 connectors,
0 findings), and make lint all green at HEAD f7632b9.

## Next

P-12 conventions/fan-out-template consolidation is partially delivered (S-1 item 7); the two
queued engine items + carried minors go to a pre-fan-out mini dispatch. Then AskUserQuestion on
Pass A budget per pilot-costs.json (75-90M revised projection).

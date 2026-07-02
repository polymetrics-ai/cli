# Agent Trace: planner

Phase: wave1-pilot · Role: gsd-loop-planner (fable) · Date: 2026-07-02 · Branch:
connector-architecture-v2 (HEAD ~9457b48)

## Rendered Prompt Or Prompt Reference

Coordinator dispatch: plan wave1-pilot (10-connector pilot migration per ROADMAP/orchestration
plan), sequencing wave0 carried follow-ups N1+N4 first; produce all phase planning artifacts;
resolve the gmail-OAuth and parity-test-location questions with code evidence.

## Files Inspected

- .planning/ROADMAP.md; .planning/phases/wave0-engine-harness/{SUMMARY.md, REVIEW.md (full incl.
  re-review), DECISIONS.md, PRD-COVERAGE.md}
- docs/migration/{conventions.md (full), orchestration-plan.md, inventory.json (pilot rows),
  result.schema.json, review.schema.json}; docs/prompts/universal-programming-loop-prompts.md
- Legacy pilots (targeted): gmail/{auth.go FULL, gmail.go, streams.go}, github/{auth.go,
  github.go:236,254-289,1759-1900, streams.go:12-30, manifest.go}, monday/monday.go:131-160,396,
  sentry/sentry.go:7-9,89,144-152,243, chargebee/chargebee.go:262-264 + streams.go:40-47,
  bitly/bitly.go:36,200-240, calendly/calendly.go:97,172-176,303-317,
  zendesk-support/zendesk_support.go:8-13,269-284, vitally/vitally.go:104, xkcd/xkcd.go:55-108
- Engine/harness: engine/{bundle.go (AuthSpec:91-110, StreamSpec:155-180, LoadAll/Load:271-316),
  hooks.go (5 interfaces + registry), read.go:142, paginate.go:12-76}, defs/defs.go (embed
  all:*), engine/parity_stripe_test.go:1-50, hooks/hookset/, cmd/connectorgen/gen.go:14-24
- Wave1-pilot scaffold: PRD-COVERAGE.md, RUN-STATE.json, AGENTS.md, SUMMARY.md, traces/

## Actions Taken

- Wrote .planning/phases/wave1-pilot/{SPEC.md, PLAN.md, TEST-PLAN.md, THREAT-MODEL.md,
  RUNBOOK.md, DATA-MODEL.md, API-CONTRACT.md, OBSERVABILITY.md, EVAL-PLAN.md}; updated
  PRD-COVERAGE.md (3 not-applicable lines + artifact references); this trace.

## Commands Run

- Read-only inspection: ls/grep/sed over legacy pilot packages, engine, conformance, cmd/,
  docs/; python3 extraction of the 10 inventory.json rows; `ls internal/connectors/*/write.go`
  (write survey). No build/test runs (planning only, no code changed).

## Findings

1. **Gmail decision — KEEP, Tier-2 AuthHook.** Legacy gmail is a real OAuth2 refresh-token-grant
   connector (auth.go oauthRefreshAuth: refresh→access exchange, expiry-60s cache, Bearer apply);
   the 3-legged consent/acquisition is NOT in legacy — refresh token arrives via secrets
   (credentials layer), which answers REVIEW.md carried item 7. Engine lacks a refresh-grant auth
   mode (bundle.go:92) → hooks/gmail AuthHook (~130 loc port) suffices; no roster swap needed.
2. **Parity-test location — new per-connector dirs `internal/connectors/paritytest/<name>/`.**
   Wave0 parity lives in engine/ as shared `package engine_test`; 10 parallel agents there would
   break the path-guard disjointness rule and collide on test-helper identifiers. defs.go's
   `//go:embed all:*` means no shared-file edit for new bundles; stripe/searxng parity stays in
   engine/ (wave0 artifacts).
3. **Only github has writes among pilots** (write survey verified; github.go:236 + 16 actions at
   :1759+; all other pilots return ErrUnsupportedOperation).
4. **monday cannot be Tier-1**: reads are GraphQL POSTs with in-body pagination; engine read path
   sends nil body (read.go:142 — StreamSpec.Body is unwired at runtime; latent gap noted for
   P-12) → StreamHook.
5. **sentry link_header risk** pre-specified with a resolution ladder (results= attribute vs
   connsdk unconditional rel=next).
6. **N1 red-first path exists**: conformance self-test-bundle pattern
   (testdata/good/acme-numeric-cursor) reusable for a github_date_range cursor bundle.

## Handoff Summary

15 tasks in 5 dispatch waves: DW-0 P-0 (N1+N4, serial, TDD) → DW-1 P-1..P-10 (10-way parallel,
disjoint dirs defs/<name> + paritytest/<name> + hooks/<name> for monday/github/gmail) → DW-2 P-11
(Fable 100% line-by-line review + repairs) → DW-3 P-12 (conventions/prompt patch) + P-13
(pilot-costs.json) → DW-4 P-14 (wave gate + phase close + Pass B human gate). Executors sonnet,
review fable, prompts from docs/prompts/universal-programming-loop-prompts.md with PLAN.md rows
inlined.

## Verification Evidence

Planning-only change set; verification is deferred to P-0/P-14 command blocks defined in
TEST-PLAN.md §6 and RUNBOOK.md. All code claims above carry file:line citations checked during
this session.

## Unresolved Risks

- github (XL, 3664 loc) may land `partial` — SPEC §5.6 sets a parity floor so partial is honest
  and measurable.
- monday hook may exceed the 300-line cap → Tier-3 escalation path pre-authorized via
  coordinator flag (SPEC §5.5).
- sentry termination behavior unproven until fixtures exist (ladder in SPEC §5.3).
- Coordinator open questions: DW-1 staggering; P-11 reviewer fan-out shape; bundle N2 guard into
  P-0?; 11 MB history blob before any push (wave0 N7).

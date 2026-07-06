# TDD Ledger

Phase: wave1-pilot
Status: POPULATED (P-14 close pass, gap-loop cycle 1 complete). Written by gsd-loop-reviewer at
HEAD f7632b9165fd87623105d93015c608d51b36d6e3, 2026-07-02.

Record failing test evidence before production code for every behavior-adding task. This ledger is
an INDEX: the authoritative RED->GREEN transcripts live in the per-task trace ledgers under
`.planning/phases/wave1-pilot/traces/`. Every row below was written only after the referenced
ledger's red evidence was spot-checked against the shipped test/code (failure signatures must match
the actual Fatalf/compiler format strings — see "Reviewer verification" at the bottom).

| Task | Scope | Trace ledger (red evidence) | Red evidence shape | Status |
|---|---|---|---|---|
| P-0 | conformance/connectorgen pre-pilot follow-ups (N1, N2, N4) | traces/p0-ledger.md | failing conformance assertions + seeded-invalid-bundle validate failures before rule changes | GREEN |
| P-1 | xkcd bundle + paritytest | traces/p1-xkcd-ledger.md | bundle-missing embed failure, then per-assertion RED | GREEN (re-repaired in S2, see below) |
| P-2 | vitally bundle + paritytest | traces/p2-vitally-ledger.md | bundle-missing RED -> parity GREEN | GREEN |
| P-3 | bitly bundle + paritytest | traces/p3-bitly-ledger.md | bundle-missing RED + 2-page next_url pagination RED | GREEN |
| P-4 | calendly bundle + paritytest | traces/p4-calendly-ledger.md | `pattern all:*: cannot embed directory calendly: contains no embeddable files` (verified verbatim) | GREEN (re-repaired in S2) |
| P-5 | sentry bundle + StreamHook + paritytest | traces/p5-sentry-ledger.md | hook-not-registered + Link-header results= stop REDs | GREEN |
| P-6 | chargebee bundle + paritytest | traces/p6-chargebee-ledger.md | bundle-missing RED; envelope-unwrap parity REDs | GREEN (re-tightened in S2) |
| P-7 | zendesk-support bundle + paritytest | traces/p7-zendesk-support-ledger.md | dual-auth candidate-order RED->GREEN (genuine, per REVIEW-B) | GREEN (incremental invention REPAIRED in S2) |
| P-8 | monday bundle + StreamHook/CheckHook + paritytest | traces/p8-monday-ledger.md | GraphQL in-body pagination REDs | GREEN |
| P-9 | github bundle + AuthHook/WriteHook + paritytest | traces/p9-github-ledger.md | 25-write-action parity REDs; JWT exchange RED | GREEN (majors repaired in S2) |
| P-10 | gmail bundle + OAuth2 AuthHook + paritytest | traces/p10-gmail-ledger.md | unresolved-key transitions, https-guard REDs, token-endpoint failure RED | GREEN (re-tightened in S2) |
| R-3 | conformance hook-aware dynamic checks | traces/r3-ledger.md | undefined-symbol compile REDs + marker-skip honesty assertions | GREEN |
| S-1 | gap-loop cycle-1 engine mini-wave (typed computed_fields; config in computed_fields Vars; optional-query dialect; last_path_segment; token_path stop_path + loop guard; spec-default materialization; validate default_type_mismatch rule) | traces/gaploop-s1-ledger.md | per-item REDs: `count_typed = "42" (string), want a native number`; `unresolved key "owner" in config`; QueryParam undefined-symbol compile RED; `unknown filter "last_path_segment"`; `does not implement Err() error` (all verified against shipped format strings) | GREEN |
| S-2 | gap-loop cycle-1 pilot repairs (xkcd, calendly, zendesk-support, github+gmail, chargebee+sentry; vitally/bitly doc fixes) | traces/s2-xkcd-ledger.md, traces/s2-calendly-ledger.md, traces/s2-zendesk-ledger.md, traces/s2-github-gmail-ledger.md, traces/s2-chargebee-sentry-ledger.md | tests edited FIRST against unrepaired bundles: xkcd 11-field drop RED (verified verbatim); zendesk inverted-filter RED `updated_at[gte] = "2026-01-01T00:00:00Z", want empty` (verified verbatim); zendesk loop-guard RED (matches paginate.go:296 message); chargebee/gmail stringify-test REDs (predicted by S-1 handoff, re-verified independently) | GREEN |

Known non-TDD items (docs-only, no RED/GREEN applies, verified by gate re-run instead): S-1 item 7
(conventions.md/meta-schema pass); S-2 sentry/chargebee dead-config spec+docs edits (dropping an
already-dead key); all docs.md honesty corrections.

One deliberate no-test item: S-2 chargebee item 2 (`sort_by[asc]=updated_at`) is a STOPPED
engine-scope gap (see s2-chargebee-sentry-ledger.md "STOP") — no RED was added because no GREEN is
reachable without an out-of-scope engine change; carried on the phase queue (SUMMARY.md).

## Reviewer verification (gsd-loop-reviewer, P-14)

Spot-checked red-evidence claims against shipped code (failure text must match actual format
strings / compiler output):
1. traces/s2-xkcd-ledger.md RED — matches `parity_test.go`'s exact Fatalf string ("engine record
   dropped real API field %q (schema-projection silently discarding a field legacy passes
   through)"); field iteration order (month, num, link -> first missing = "link") consistent.
2. traces/s2-zendesk-ledger.md RED — `updated_at[gte] = %q, want empty (...not real Zendesk API
   surface)` matches shipped assertion; loop-guard message matches engine/paginate.go:296 verbatim.
3. traces/gaploop-s1-ledger.md item-5 RED — "unexpected cursor (must not be requested)" and
   "does not implement Err() error" match paginate_test.go:378/465 exactly.
4. traces/p4-calendly-ledger.md RED — Go embed error text is the real toolchain output for an
   empty `all:*`-embedded directory.

TDD-GATE.json task rows populated from this index in the same close pass.

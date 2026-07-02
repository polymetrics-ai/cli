# wave1-pilot gap loop — cycle 1 (coordinator plan, from REVIEW-A.md + REVIEW-B.md)

Status: reviews complete (A: 5/5 pass + 8 majors; B: vitally/bitly pass, xkcd/calendly/zendesk-support FAIL).
Phase close + Pass A fan-out BLOCKED until this cycle lands.

## Step 1 — Engine mini-wave (ONE sonnet agent, serial; engine pkg + connectorgen + meta-schemas + conventions)
Mandated by adjudications A1/A3/B2 + flags:
1. Typed computed_fields extraction: bare `{{ record.path }}` (no filters) copies the raw typed JSON
   value (number/bool/null preserved); filtered/mixed templates keep string semantics. TDD.
2. Wire Config (NOT Secrets) into applyComputedFields Vars (github G0 repository marker).
3. Optional-query dialect: stream.Query entry may be object {template, omit_when_absent: true,
   default?} — absent-key-falsy ONLY inside these; string entries keep hard-error. (6 occurrences.)
4. `last_path_segment` interpolation filter (calendly id).
5. token_path cursor paginator: stop_path support + loop guard (zendesk has_more).
6. C3 decision: engine materializes spec.json `default` values for absent config keys at
   runtimeConfig build (single mechanism for all legacy base-URL defaults). Validate rule: default
   must type-check.
7. connectorgen/meta-schema/conventions.md updates for all of the above + line-cap wording
   (soft 300 / self-report / hard 400 / 3rd interface = Tier 3) + bitly §4 next_url fixture
   exception + dual-auth ordering golden pattern (zendesk ledger → conventions §3).

## Step 2 — Pilot repair wave (parallel, disjoint dirs, after Step 1)
- xkcd (FAIL): full 11-field schemas or projection passthrough + recorded-real-shape fixtures.
- calendly (FAIL): restore id via last_path_segment, x-primary-key [id], delete stripDerivedID;
  page_size → spec default 100 (C3) or static; dead keys (max_pages, mode) dropped; users stream
  pagination none; docs honesty.
- zendesk-support (FAIL): remove invented updated_at[gte] incremental (keep x-cursor-field);
  adopt stop_path=meta.has_more; drop dead keys; fix parity comment + TEST-PLAN note.
- github (majors): auth_type + secret-alias surface restored or documented w/ no silent
  mode:none fall-through (fail loud); docs incremental honesty + since-param parity test;
  label color-strip normalization (RecordHook or computed_fields).
- gmail (majors): docs.md Known-limits stale claim fixed; dead keys max_pages/mode dropped.
- sentry (major): hostname dead config — either wire via C3 default derivation decision or drop +
  docs fix.
- chargebee (majors): sort_by[asc]=updated_at on incremental requests (optional-query dialect);
  site dead config resolved; adopt typed computed_fields (drop string-widened schemas + lock-in
  tests updated to assert native types).
- vitally/bitly (minors): docs corrections; vitally Check note; adopt typed extraction where it
  simplifies.

## Step 3 — P-12/P-14 close
- conventions.md consolidated patch (typed extraction, optional-query, defaults, null-vs-absent
  sparse-record parity case in fan-out template, "legacy is ground truth over TEST-PLAN — escalate
  conflicts" rule, comment-vs-bundle contradiction grep in reviews).
- Fix hollow gate artifacts: TDD-LEDGER.md index rows for P-0..P-10 + repairs (evidence in
  traces/p*-ledgers), SUMMARY.md, VERIFICATION.md with HEAD + rerun evidence, RUN-STATE.
- Focused Fable re-review of the gap-loop diff → GO/NO-GO → phase close.
- Then AskUserQuestion: Pass B budget (docs/migration/pilot-costs.json has real numbers; revised
  Pass A projection 75–90M subagent tokens / ~77 bundle agents).

## Wave gate after each step
go build ./... && go test ./internal/connectors/... ./cmd/... && make lint &&
go run ./cmd/connectorgen validate internal/connectors/defs && git commit (coordinator).

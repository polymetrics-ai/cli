# REVIEW-B — wave1-pilot batch B adversarial review (xkcd, vitally, bitly, calendly, zendesk-support)

Reviewer: gsd-loop-reviewer (Fable). HEAD 48cbff5, branch connector-architecture-v2.
Scope: full line-by-line read of `defs/<name>/**`, `paritytest/<name>/`, per-connector trace
ledgers, spot-checked against the LEGACY packages (`internal/connectors/<name>/`) and the engine
source (`read.go`, `paginate.go`, `interpolate.go`, `connsdk/paginate.go`, `connsdk/http.go`).
All checks per `docs/migration/review.schema.json` + `docs/migration/conventions.md` §5 meta-rule.

Verification independently re-run at review time (not trusted from ledgers):

```
go test -count=1 ./internal/connectors/paritytest/{xkcd,vitally,bitly,calendly,zendesk-support}   # all ok
go test -count=1 ./internal/connectors/conformance -run 'TestConformance/(xkcd|vitally|bitly|calendly|zendesk-support)'  # PASS x5
go run ./cmd/connectorgen validate internal/connectors/defs   # 13 connectors, 0 findings
go build ./internal/connectors/... && go vet ./internal/connectors/...   # clean
```

Green harness output is necessary but NOT sufficient — three of the findings below are cases
where the parity suite's server/fixture shape masks a real-API divergence the meta-rule forbids.

---

## Verdict summary

| connector | verdict | blockers | majors | minors |
|---|---|---|---|---|
| xkcd | **fail** | 1 | 0 | 1 |
| vitally | **pass** | 0 | 0 | 1 |
| bitly | **pass** | 0 | 0 | 3 |
| calendly | **fail** | 2 | 2 | 0 |
| zendesk-support | **fail** | 1 | 2 | 2 |

---

## xkcd — FAIL

### Findings

1. **BLOCKER — `internal/connectors/defs/xkcd/schemas/{latest,comic}.json`, `streams.json`:
   schema projection silently drops fields legacy emits on every real API response.**
   Legacy's live read path is a raw passthrough: `json.Unmarshal(resp.Body, &rec); emit(rec)`
   (`xkcd.go:93-97`) — EVERY field of the response body is emitted. The real XKCD JSON API (the
   bundle's own `docs_url`, https://xkcd.com/json.html) returns `month, num, link, year, news,
   safe_title, transcript, alt, img, title, day`. The bundle's schemas declare only 6 of those
   (`num,title,safe_title,year,month,day` — copied from legacy's `Catalog()` field list, which is
   NOT the record-shaping function conventions.md §2 mandates as the source). In `"schema"`
   projection mode (`engine/read.go:535` `projectRecord`), `link/news/transcript/alt/img` are
   silently dropped — changed emitted record DATA for every input legacy accepts. Undocumented:
   neither `docs.md` "Known limits" nor `p1-xkcd-ledger.md` mentions it; `docs.md:37-39` actively
   misattributes the schema to "legacy's ... direct pass-through record shape". The parity suite
   masks it: `xkcdFixtureBody` (parity_test.go:95) and both stream fixtures contain ONLY the 6
   declared fields, so DeepEqual never sees the divergence — and the fixtures therefore also fail
   §4's recorded-real-shape rule (a real `info.0.json` response always carries `alt`/`img`/etc.).
   **Fix (either):** (a) `"projection": "passthrough"` on both streams — the exact analog of
   legacy's behavior — keeping the 6-field schema as documentation of the core fields; or
   (b) declare all 11 real fields. Then re-record fixtures/parity bodies in the real 11-field
   shape so parity actually proves passthrough.

2. **minor — `internal/connectors/defs/xkcd/docs.md:60-63`:** claims `spec.json`'s `base_url`
   "still declares `\"default\": \"https://xkcd.com\"`" — it does not (no `default` key in
   `spec.json`). Doc/code drift; add the annotation or fix the sentence.

### Checked
`{"schema_fidelity_spot_checks": 2, "write_actions_verified": true (none declared, legacy read-only verified), "fixtures_realistic": false, "escape_hatches_justified": true (none), "secret_scan_clean": true, "conventions_adherent": false}`

Everything else is strong: hostile-path fail-closed parity test is exemplary; base_url
no-default deviation properly documented; no dead spec keys.

---

## vitally — PASS

### Findings

1. **minor — `internal/connectors/defs/vitally/streams.json` (`base.check`) vs legacy
   `vitally.go:33-47`:** legacy `Check` is offline config/secret validation only (never dials);
   the bundle's check issues a real `GET /resources/accounts`. Fail-loud improvement, zero record
   data impact, but it is an undocumented check-behavior deviation (a bad credential now fails at
   Check instead of first Read; a network outage now fails Check). One sentence in docs.md
   "Known limits" closes it.

### Adjudication notes
- `api_key_header` + empty `prefix` for the pre-built `Authorization` header value is the
  byte-exact reproduction of legacy `connsdk.APIKeyHeader("Authorization", auth, "")` — verified
  against `engine/auth.go:85-90` and locked by `TestParityVitally_AuthorizationHeaderByteExact`.
  Correctly avoided `basic` mode (would re-encode a pair the connector never receives).
- `status` optional query filter: NOT declared, NOT wired, documented in docs.md + ledger, base
  case parity-tested (`NoStatusParamSentWhenUnset`). Consistent with the searxng §5 item-7
  precedent. ACCEPTABLE as documented scope narrowing — and counted as occurrence #2 of the
  recurring optional-query-param gap (see P-12 adjudication below).
- Schema `{id,name,traits}` matches legacy's mapper field-for-field; `traits` correctly typed as
  an opaque object bag. Fixtures realistic for this endpoint's envelope.

### Checked
`{"schema_fidelity_spot_checks": 1, "write_actions_verified": true, "fixtures_realistic": true, "escape_hatches_justified": true, "secret_scan_clean": true, "conventions_adherent": true}`

---

## bitly — PASS

### Findings

1. **minor — `internal/connectors/defs/bitly/docs.md:41-46`:** claims `size=50` "is sent as a
   static per-stream query value on the FIRST request only — subsequent pages are driven entirely
   by the absolute `pagination.next` URL ... matching legacy's `harvest` loop". False for the
   engine: `readDeclarative` merges `baseQuery` into EVERY page request (`read.go:143`
   `mergeQuery(baseQuery, page.Query)`), and `connsdk.Requester.resolveURL` (http.go:145-165)
   re-applies it onto the absolute next URL (Del+Add, replacing any same-named param). Legacy
   explicitly clears the query on next pages (`bitly.go:180-183`). Verified benign in data terms
   (Bitly's next URL carries the same `size`, and the replace is idempotent), but the doc claim
   is wrong and must be corrected to "re-sent on every page; value-identical to what the next URL
   already carries".
2. **minor — `internal/connectors/paritytest/bitly/parity_test.go:230-232`:** the
   `bitlyTwoPageServer` comment says the test "asserts the request-shape (... size query param on
   page 1 only) matches legacy" — no such assertion exists (only path + search_after are matched),
   and per finding 1 the engine in fact sends `size=50` on page 2. Fix the comment (or add the
   honest assertion of the actual engine behavior).
3. **minor — `internal/connectors/defs/bitly/fixtures/streams/bitlinks/` single-page fixture for
   a paginated stream:** deviates from §4's 2-page rule, but the ledger's justification is
   correct and independently verified — a `next_url` fixture cannot carry the replay server's URL
   (unknown until runtime), `pagination_terminates` exercises `Streams[0]` (organizations,
   non-paginated), and real 2-page `next_url` correctness IS proven live by
   `TestParityBitly_BitlinksStreamPaginates`. Accepted; MUST be codified in conventions.md at
   P-12 (see flags) so fan-out agents don't cargo-cult either way.

### Adjudication notes
- Bearer auth byte-parity proven. `next_url` + same-host SSRF guard correct for Bitly's absolute
  `pagination.next`. Group-scoped path templating matches legacy's `url.PathEscape`.
- `page_size`/`max_pages` config-driven overrides not modeled: documented (docs.md Known limits
  + ledger), static `size=50` matches the stripe `limit=100` precedent, keys correctly NOT
  declared in spec.json (F6). ACCEPTABLE; occurrence #3 of the recurring gap (P-12).
- `group_guid`-missing error-text delta: same failure classification, postgres item-9 precedent.
  ACCEPTABLE.
- Schemas match legacy mappers field-for-field on all 4 streams (spot-checked all 4 against
  legacy `streams.go` mappers and Bitly v4 docs field names). Null-vs-absent: legacy emits
  explicit nils for keys absent from the raw item; the engine omits them — only observable if
  Bitly omits a documented field; noted in the wave-level flags, not a bitly-specific defect.

### Checked
`{"schema_fidelity_spot_checks": 4, "write_actions_verified": true, "fixtures_realistic": true, "escape_hatches_justified": true, "secret_scan_clean": true, "conventions_adherent": true}`

---

## calendly — FAIL

### Findings

1. **BLOCKER (adjudication) — `internal/connectors/defs/calendly/schemas/*.json`,
   `paritytest/calendly/parity_test.go` (`stripDerivedID`): the legacy-emitted derived `id` field
   is dropped from every record of every stream, and the parity suite strips it on both sides
   before comparison.** Legacy stamps `id = idFromURI(uri)` on all 5 streams and publishes it as
   the primary key (`streams.go` mappers; `calendly.go:402-412`). The bundle never emits `id` and
   switches `x-primary-key` to `uri`. The §5 meta-rule is categorical: a deviation is acceptable
   only iff it "never changes the emitted record DATA for any input legacy itself would accept" —
   removing a field from every record is a data change on the default path, full stop. The
   ledger's ACCEPTABLE self-verdict ("fully recoverable", "zero-data-loss alternative") is a
   misclassification: recoverability is not the meta-rule's test, and the in-repo precedent cuts
   the other way — searxng's derived `stream` marker (§5 item 6) and comma-joined `engines`
   (item 4) were both legacy-derived fields whose omission/reshaping was treated as MUST-FIX via
   engine increments, with the parity-test stripping workarounds explicitly removed on
   resolution. The ledger's claim that the §6 decision tree offers no hook trigger here is also
   wrong: `RecordHook` is defined verbatim for "per-record post-processing beyond schema
   projection". To the agent's credit the deviation is exhaustively documented and the strip is
   individually justified — this is an honest, well-lit misclassification, not a stealth one.
   **Adjudication: dropping a legacy-emitted field is NOT acceptable under the meta-rule. Fan-out
   fix: file ENGINE_GAP (URI→trailing-segment transform); P-12 adds a `last_path_segment` filter
   (or generic `split:<sep>` + `last`) to `interpolate.go`'s filter set — a few lines, closes the
   same gap for every HAL/URI-keyed API in the fan-out — then `"id": "{{ record.uri |
   last_path_segment }}"` in `computed_fields`, restore `x-primary-key: ["id"]`, and delete
   `stripDerivedID`. Interim (if calendly must ship before P-12): a Tier-2 `RecordHook`. Either
   way the strip must not survive into the fan-out template.**
2. **BLOCKER — `internal/connectors/defs/calendly/spec.json` + `streams.json` `query`
   (`"count": "{{ config.page_size }}"`): undocumented accepted-input regression + false doc
   claim.** `stream.Query` templating has no absent-key tolerance (conventions §3, verified at
   `read.go:355-364`), so every organization-scoped read hard-errors when `page_size` is unset —
   an input legacy accepts (defaults to 100, `calendly.go:363-376`). `page_size` is not in
   `required[]`, so the spec advertises it as optional while the bundle cannot run without it;
   conventions §3 explicitly forbids declaring a query-referenced property "unless the query
   template can actually tolerate its absence (currently: never, for query)". Every parity test
   masks it by setting `page_size: "100"` (parity_test.go:73); conformance masks it via
   synthetic-value config. `docs.md:48-49` compounds it with a false parity claim ("matching
   legacy's `calendlyPageSize` default of 100"). Not in docs.md Known limits, not in the ledger —
   an UNDOCUMENTED parity deviation (blocker class per review.schema.json). **Fix: static
   `"count": "100"` (bitly/zendesk/stripe pattern) and drop `page_size` from spec.json — or keep
   it and add it to `required[]` with an honest docs.md note. Static literal is the
   fan-out-consistent choice.**
3. **major — `internal/connectors/defs/calendly/spec.json`: dead config keys `max_pages` and
   `mode` (F6 violation).** Neither is consumed by any template or engine mechanism; `mode`'s
   description ("fixture for credential-free conformance") describes a legacy affordance the
   bundle explicitly does not have (its own docs.md says so). `max_pages`'s description implies a
   wired cap that does not exist. Delete both.
4. **major — `internal/connectors/defs/calendly/streams.json` `base.pagination` applies the
   `next_url` paginator to the `users` single-object stream.** Harmless today (`/users/me` has no
   `pagination.next_page`, so it stops after one page) but it makes the bundle's declared shape
   wrong for the stream and will misbehave if Calendly ever adds a `pagination` envelope to a
   single-object response. Give `users` an explicit `"pagination": {"type": "none"}` override
   (stream-level replaces base wholesale).

### Adjudication notes (accepted items)
- `organization_uri` as required config replacing legacy's per-read `/users/me` auto-discovery:
  ACCEPTABLE. Request bytes and emitted data are identical given the same URI; the parity suite
  proves legacy's own discovery resolves the identical value from the same server; the URI is
  per-account-invariant; documented in spec/docs/ledger. This is input-surface relocation, not a
  data change — distinct in kind from finding 1. (True auto-discovery, if ever needed, is a
  StreamHook.)
- 3 streams publishing `x-cursor-field: updated_at` with NO incremental block and NO
  client_filtered: correct and well-argued — matches legacy's real (non-)filtering exactly;
  adding filtering would be new behavior (this exact reasoning is what zendesk-support violated,
  see below).
- `min_start_time` incremental for scheduled_events only, from state-cursor or `start_date`,
  rfc3339 verbatim: matches legacy `incrementalLowerBound` + `calendly.go:155-161`; parity-tested
  both sources plus the negative case on the other 3 streams. Excellent.
- computed_fields flatten of `user.uri/name/email`: matches legacy's membership mapper.
- Single-page `next_url` fixtures: same accepted harness limitation as bitly (see bitly finding
  3); the live 2-page proof is `TestParityCalendly_ScheduledEventsTwoPagePagination`.

### Checked
`{"schema_fidelity_spot_checks": 5, "write_actions_verified": true, "fixtures_realistic": true, "escape_hatches_justified": false (RecordHook wrongly ruled unavailable), "secret_scan_clean": true, "conventions_adherent": false}`

---

## zendesk-support — FAIL

### Findings

1. **BLOCKER — `internal/connectors/defs/zendesk-support/streams.json`: every stream declares
   `incremental.request_param: "updated_at[gte]"` — a server-side filter legacy NEVER sent, on a
   query parameter the Zendesk Support API does not document for these collection endpoints.**
   Legacy `harvest()` (zendesk_support.go:152-195) sends no incremental filter of any kind;
   `start_date` is doc-comment-only. The bundle's own parity file says the quiet part out loud —
   `parity_test.go:297-306`: "Zendesk's real API has no documented updated_at>= query filter for
   these collection endpoints ... this bundle therefore intentionally does NOT declare an
   incremental.request_param either" — a comment that directly contradicts the shipped
   `streams.json` (which declares it on all 5 streams) and the very next test
   (`StartDateConfigRaisesLowerBound`, an engine-only test asserting the param IS sent). The
   ledger (item 4) records the flip: the agent first correctly concluded "declaring request_param
   would be a BEHAVIOR CHANGE ... not parity", then kept it anyway to satisfy TEST-PLAN's
   "start_date-raised" row. Adjudication:
   - Under the §5 meta-rule this fails BOTH ways. If the param worked, a configured
     `start_date`/persisted cursor (inputs legacy accepts and ignores) would change the emitted
     record set vs legacy — the ledger's own text admits records "would be extra-filtered-out by
     the engine relative to legacy". "Narrows sync SCOPE" IS changed emitted data; the
     "capability addition" framing is the exact "new behavior under the guise of a migration"
     conventions.md's rate_limit rule forbids. Since the param is NOT real Zendesk API surface,
     what actually happens live is worse: Zendesk ignores unknown params, so the bundle
     advertises incremental filtering that silently no-ops in production and is "proven" only
     against fixture servers that honor an invented parameter — a schema/API-fidelity blocker
     (review.schema.json's first blocker example) plus unrealistic fixtures for the incremental
     path.
   - TEST-PLAN's "✓ start_date-raised" row for zendesk-support was a planning error (legacy has
     no such wire behavior to be parity-tested); per the migration discipline a plan-vs-legacy
     conflict is an escalation, not something to satisfy by inventing an API parameter.
   **Fix: remove the `incremental` blocks from all 5 streams (keep `x-cursor-field: updated_at`
   in the schemas for manifest-surface parity — the exact calendly-3-streams pattern this same
   wave already established); delete `TestParityZendesk_StartDateConfigRaisesLowerBound` or
   invert it to assert NO filter param is ever sent (matching legacy); reconcile the stale test
   comment; amend TEST-PLAN row to "N/A — legacy sends no incremental filter"; real Zendesk
   incremental belongs to Pass B via the documented `/api/v2/incremental/*` export endpoints
   (different endpoint + `start_time` unix param), not this param.**
2. **major — `streams.json` `base.pagination` (cursor/token_path) ignores `meta.has_more`,
   diverging from legacy's stop rule against Zendesk's documented behavior.** Legacy stops when
   `has_more != "true" || after_cursor == ""` (zendesk_support.go:189); the engine's
   `connsdk.CursorPaginator` (paginate.go:106-117) stops ONLY on an absent/empty `after_cursor`
   — the token_path cursor variant has no `stop_path` support and, unlike `nextURL`, no
   loop-detection guard. Zendesk's own cursor-pagination docs instruct clients to use `has_more`
   and warn the cursor properties may be populated even when `has_more` is false; docs.md:59-65's
   contrary claim ("live-API behavior always emit after_cursor: null on the final page") cites
   only legacy's own test fixture as evidence, and both the fixtures and every parity server are
   shaped `has_more:false + after_cursor:null`, so the divergent case (has_more:false with a
   non-null cursor) is never exercised. Best case live: one extra empty-page request per stream
   per sync (request-count delta, meta-rule-acceptable IF documented); worst case: a pagination
   loop with no guard. **Fix (ordered): (a) P-12 engine increment — support `stop_path` on the
   token_path cursor variant (trivial: reuse lastRecordCursor's stop logic) and declare
   `"stop_path": "meta.has_more"`, restoring legacy's exact stop rule; (b) until then, document
   the divergence honestly in Known limits (replacing the unverified "always null" claim) and add
   a parity/regression test for the has_more:false + non-null-cursor page.**
3. **minor — `spec.json`: dead config keys `page_size`, `max_pages`, `mode` (F6 violation).**
   `page[size]=100` is a static query literal (correct, stripe pattern) but then `page_size` must
   not be declared; `max_pages`/`mode` are wired to nothing. Same defect class as calendly
   finding 3. Delete all three.
4. **minor — `spec.json` `email` is secrets-only; legacy also accepted a non-secret
   `config.email` and dotted `credentials.*` secret keys (zendesk_support.go:271-287, 366-378).**
   The bare-key secret surface is a reasonable canonicalization (and the parity helper documents
   it), but the config-key fallback narrowing is not mentioned in docs.md's config-surface
   deviation note. One sentence fixes it.

### Adjudication notes (accepted items)
- **Dual-auth ordering: CORRECT.** Bearer (`when: {{ secrets.access_token }}`) is declared before
  Basic (`when: {{ secrets.api_token }}`), reproducing legacy's access_token-first precedence
  under `selectAuth`'s first-match-wins; the both-secrets-present corner is explicitly
  parity-tested (`OAuthBearerAuthParity`). Basic username template `{{ secrets.email }}/token` +
  password `{{ secrets.api_token }}` is byte-identical to legacy's `connsdk.Basic(email+"/token",
  apiToken)` — asserted against the exact base64 string. The ledger's auth candidate-order RED→
  GREEN is a genuine, well-evidenced TDD transition. This is the fan-out template for dual-auth.
- `base_url`-required / no `subdomain` derivation: ACCEPTABLE config-surface narrowing —
  operator-reachable equivalent for every legacy input, no request/data delta, documented in
  docs.md + ledger with the searxng-subreddit precedent correctly applied.
- 2-page tickets cursor fixture: present and correct (token cursors are expressible in static
  fixtures, unlike next_url); `pagination_terminates` has a real second page here.
- Schemas: all 5 match legacy mappers field-for-field (spot-checked all 5).

### Checked
`{"schema_fidelity_spot_checks": 5, "write_actions_verified": true, "fixtures_realistic": false (incremental + has_more cases fixture-shaped to pass), "escape_hatches_justified": true (none used), "secret_scan_clean": true, "conventions_adherent": false}`

---

## Cross-cutting adjudications (requested rulings)

### 1. calendly's dropped derived `id` (field-strip in parity)
**Ruling: NOT acceptable under the §5 meta-rule** — a legacy-emitted field disappearing from
every record is emitted-DATA change on the default path; recoverability and documentation
quality do not satisfy the rule's condition, and the searxng items 4/6 precedent (engine grew
`join:`/static-literal computed_fields rather than dropping legacy-derived fields, and the
parity strips were removed on resolution) is directly on point. The strip itself was properly
individually justified (the reviewable-workmanship bar was met); the self-verdict was not.
**Fan-out needs the `last_path_segment` filter (P-12 mini engine increment), not per-connector
RecordHooks**: the transform is pure, single-purpose, and will recur across URI/HAL-keyed APIs;
a RecordHook per such connector would bleed Tier-2 packages for a one-line derivation. RecordHook
is the sanctioned interim only if calendly must close before P-12.

### 2. Recurring "optional/config-driven query param not expressible" (vitally `status`, bitly
`page_size`/`max_pages`, calendly `page_size` workaround, zendesk dead keys, gmail
`start_date`/`include_spam_and_trash`, searxng wave0 F6)
**Ruling: ≥3-occurrence threshold is met several times over — this is a mandatory P-12
ENGINE_GAP mini-increment, no longer a per-connector documented narrowing.** Recommended shape:
**an explicit opt-in optional-query dialect field, NOT blanket absent-key-falsy query
templating.** Blanket absent-key-falsy would convert every future mis-declared/mistyped required
query key from a fail-loud error into a silently-unfiltered request (the F4 fail-open class the
engine deliberately rejects for headers/secrets) and would retroactively change every existing
bundle's query semantics. Instead, per-entry opt-in on `stream.Query`, e.g.:

```json
"query": {
  "status":   { "template": "{{ config.status }}", "omit_when_absent": true },
  "count":    { "template": "{{ config.page_size }}", "default": "100" },
  "page[size]": "100"
}
```

— string entries keep today's exact hard-error semantics (zero migration risk); object entries
get `when`-grammar absent-key-falsy resolution (omit the param when the resolved value is
empty), plus an optional `default` literal that also cleanly closes the calendly-page_size and
legacy-default-base-url-style gaps. Static validation stays strict: the referenced key must
still be DECLARED in spec.json (mirror of `when`'s validate-vs-runtime split). Batch-B rewiring
once landed: vitally `status`, bitly `size` (config override), calendly `count`, gmail's two
filters. Separately and lower-priority: config-driven `page_size`/`max_pages` runtime overrides
for paginators would need PaginationSpec-level config wiring — do NOT bundle it into this
increment; the static-literal defaults are behaviorally sufficient for parity.

### 3. zendesk-support's added server-side incremental filtering
**Ruling: NOT acceptable — this is scope/behavior ADDITION (not narrowing), on an undocumented
(non-existent) API parameter, adopted to satisfy a TEST-PLAN row that mis-modeled legacy.** The
same wave's calendly ledger states the governing principle exactly: "adding `client_filtered:
true` here would be NEW behavior legacy never had, and conventions.md's meta-rule forbids
introducing new deviations under the guise of parity." Remove the incremental blocks; keep
schema-level `x-cursor-field`; real incremental is Pass B via Zendesk's incremental-export
endpoints.

---

## Fan-out readiness (Tier-1 patterns from batch B)

**Ready to template now:**
- Bearer / api_key_header(empty-prefix) / dual-auth `when`-gated candidate lists with
  declared-order precedence (zendesk is the golden dual-auth example; its ledger item 3 should be
  lifted into conventions.md §3 as an authoring rule: "auth candidate order is load-bearing").
- Templated path segments (`{{ config.comic_number }}`, `{{ config.group_guid }}`) with
  urlencode-by-default + dot-dot guard; hostile-input fail-closed parity test shape (xkcd's).
- `next_url` pagination with same-host SSRF guard (bitly/calendly), including the documented
  "single-page conformance fixture + live 2-page parity proof" pattern for next_url streams —
  MUST be written into conventions.md §4 at P-12 as the sanctioned exception, with the
  bitly-ledger explanation attached.
- Token-path cursor pagination WITH the caveat that any API whose real stop signal is a boolean
  (`has_more`) needs the `stop_path` extension (P-12) before fan-out to Zendesk-shaped APIs.
- Config-relocation of per-account-invariant request scoping (calendly `organization_uri`) —
  acceptable pattern when the value is invariant and operator-resolvable; document per calendly's
  ledger item 1 wording.
- `x-cursor-field`-without-incremental-block for legacy's informational-only cursors (calendly's
  3 streams / zendesk post-fix).

**NOT ready — engine work needed before wide Tier-1 fan-out (P-12 backlog, in priority order):**
1. Optional-query dialect field (+ `default`) — adjudication 2. Highest recurrence.
2. `last_path_segment` (or `split`+`last`) interpolation filter — adjudication 1. Unblocks
   calendly and every URI-keyed API.
3. `stop_path` support on the token_path cursor paginator (+ a loop guard mirroring nextURL's
   `seen` map) — zendesk finding 2.
4. Conventions §4 amendment: next_url single-page-fixture exception; §2 amendment: schema source
   is the legacy RECORD-SHAPING function (or passthrough when legacy passes through raw) — the
   xkcd failure shows "Catalog field list" is an attractive wrong source worth naming explicitly.

**Process flags for the orchestrator:**
- Three connectors in the SAME wave handled the SAME optional-param gap three different ways
  (not declared / static literal / unconditionally wired) — fan-out at 200 connectors will
  amplify any un-templated judgment call; P-12 must make these patterns prescriptive.
- TEST-PLAN rows can contradict legacy ground truth (zendesk "start_date-raised", "has_more=false
  stop"); migration agents treated the plan as overriding legacy. Add to the fan-out brief: when
  TEST-PLAN and legacy disagree, STOP and escalate — legacy is ground truth for parity.
- Stale rationale drift: zendesk's parity comment describes a design that was reversed without
  reconciliation. Fan-out reviews should grep for comment-vs-bundle contradictions.
- Null-vs-absent (wave-level, informational): every mapRecord-style legacy connector emits
  explicit `nil` for keys absent from the raw item; `projectRecord` omits absent keys. Parity
  suites only prove equality for fully-populated fixtures. Against APIs that serialize nulls
  (Zendesk, Bitly largely do) this is invisible; for APIs that omit optional fields it will
  surface as absent-vs-null record deltas. Worth one conventions.md sentence and a
  deliberately-sparse-record parity case in the fan-out template.

## Gate verdict

**NO-GO for phase close as-is.** xkcd, calendly, zendesk-support return to the gap loop
(backend/migration agents) with the blocker fixes above; vitally and bitly pass (bitly with doc
corrections that can ride along in the same gap loop). None of the blockers require weakening
any test or gate — every fix strengthens parity. No human gate is triggered by the fixes
themselves; the P-12 engine-increment recommendations (items 1-3) are orchestrator decisions.

# REVIEW-A — wave1-pilot batch A (hook-heavy): github, gmail, monday, sentry, chargebee

Reviewer: gsd-loop-reviewer (fable), READ-ONLY line-by-line review per PLAN.md P-11 /
docs/migration/review.schema.json / orchestration-plan verification-pyramid layer 3.
HEAD: 48cbff5, branch connector-architecture-v2. Date: 2026-07-02.

Verification re-run by reviewer (not trusted from ledgers): `go run ./cmd/connectorgen validate
internal/connectors/defs` → 13 connectors, 0 findings · `go build ./...` clean ·
`go test -count=1 ./internal/connectors/paritytest/{github,gmail,monday,sentry,chargebee}
./internal/connectors/hooks/...` all PASS uncached · `go test ./internal/connectors/conformance
-run TestConformance` PASS (incl. github full-dynamic, gmail/monday/sentry marker-skips) ·
independent secret-shape grep over all batch-A defs/hooks/paritytest files → clean.

Legacy sources read for spot-checks: internal/connectors/github/{github,streams,auth}.go,
gmail/{gmail,streams,auth}.go, monday/{monday,streams}.go, sentry/sentry.go (targeted),
chargebee/{chargebee,streams}.go. Engine read against: engine/{read,write,interpolate,bundle}.go,
conformance/{dynamic,replay}.go.

---

## Verdict summary

| connector | verdict | blockers | majors | minors |
|---|---|---|---|---|
| github | **pass** | 0 | 3 | 4 |
| gmail | **pass** | 0 | 2 | 2 |
| monday | **pass** | 0 | 0 | 3 |
| sentry | **pass** | 0 | 1 | 2 |
| chargebee | **pass** | 0 | 2 | 0 |

No blocker-severity findings in batch A (contrast batch B: 3 fails). Every major below routes to
the gap loop (backend fix or docs/ledger correction) before wave close; none requires weakening a
gate, none triggers a human gate.

---

## github — pass

checked: schema_fidelity_spot_checks=4 (commits, issues, releases, milestones vs
streams.go record fns + docs.github.com), write_actions_verified=true (all 25 method/path/required
fields diffed against githubWriteActionSpecs + executeWriteAction and GitHub REST docs — all 25
match), fixtures_realistic=true, escape_hatches_justified=true (AuthHook: token-exchange auth;
WriteHook: compound multi-request writes — both named conventions §1 triggers; 0 skip markers,
full dynamic conformance coverage incl. WriteHook), secret_scan_clean=true,
conventions_adherent=true (with the line-cap adjudication below).

Strengths worth naming: all 25 write actions' method/path verified correct (incl. delete_file's
DELETE-with-JSON-body via body_fields, dispatch_workflow's string workflow_id matching legacy's
githubRequiredString, PUT merge/contents); the missing_ok_status REMOVAL (G12) is exactly the
right meta-rule call — declaring it would have been new leniency legacy never had, and
TestParityGithub_WriteDeleteLabelNotFoundFailsOnBothSides pins it; the G-ledger (G0–G13) is the
best deviation record in the wave; conformance write fixture response block (create_pull_request
201/number:42) fixed a real harness gap rather than hiding behind a marker.

Findings:
- **major — internal/connectors/defs/github/spec.json + docs.md**: legacy's explicit auth-mode
  selection and credential aliases are dropped UNDOCUMENTED. Legacy honors
  `auth_type`/`auth`/`authentication` (auth.go:61-96 — e.g. `auth_type=github_app` forces app auth
  even when a token secret is also set; `auth_type=public` forces anonymous), token secret aliases
  (`personalAccessToken`/`accessToken`/`oauthToken`/`installationToken`/`githubToken`/
  `GITHUB_TOKEN`, auth.go:1634-1644), private-key aliases (`privateKey`/`githubAppPrivateKey`/
  `privateKeyBase64`/`githubAppPrivateKeyBase64`), and `app_id` aliases (`client_id`/
  `github_app_id`). The bundle reads only `token`/`private_key`/`private_key_base64`/`app_id`, and
  the failure mode for an alias-shaped legacy config is a SILENT fall-through to `mode: none`
  (unauthenticated reads) — not an error. G1 ledgers the write-action-name alias drop but nothing
  ledgers the auth config-surface drop; docs.md's Auth setup implies the new keys are the only
  ones that ever existed. Fix: ledger entry + docs.md "config surface changes vs legacy" list (the
  owner/repo split already has this treatment — extend it), and consider dropping the trailing
  `{"mode":"none"}` only-when-typo hazard note into docs.
- **major — internal/connectors/paritytest/github/parity_test.go + docs.md**: the `issues`
  incremental block (`request_param: since`) forwards the app-persisted STATE cursor; legacy github
  never reads `req.State` anywhere (github.go Read — config `since` only), so a sync with persisted
  state emits a different (smaller, correctly-incremental) record set than legacy. Direction is an
  improvement and consistent with engine-wide semantics, but docs.md says "matches legacy exactly"
  (false for the state path) and NO parity test covers `since` forwarding at all (neither
  config-since equality nor the state-cursor round trip PLAN.md's parity minimum names for
  incremental connectors). Fix: ledger the state-filtering behavior delta + add the since-param
  parity test (config path asserting legacy equality; state path asserting the engine's documented
  new behavior).
- **major — internal/connectors/defs/github/writes.json (create_label/update_label)**: legacy
  strips a leading `#` from `color` (`strings.TrimPrefix(color, "#")`, github.go:1120,1133); the
  declarative body passes the record verbatim, so `color: "#ff0000"` — an input legacy accepts and
  normalizes — is sent raw and GitHub 422s. Undocumented value-normalization drop (G4 covers type
  coercion, not this). Fix: ledger + docs.md entry (or a `pattern: "^[0-9a-fA-F]{6}$"` schema
  constraint to fail loudly instead of silently diverging).
- minor — paritytest/github: TestParityGithub_WriteCreatePullRequestCompound compares only
  method/path per request, not bodies (every non-compound write test compares full decoded
  bodies). Bodies would compare equal today — strengthen to close the gap.
- minor — defs/github/fixtures/writes/{close_issue,close_pull_request}.json: expect.body omits
  `"state": "closed"` — conformance's subset body match means the hook's core payload key is never
  asserted by the fixture (it IS asserted in hooks_test.go + parity). Add `state` to expect.body.
- minor — api_surface.json scope prose + paritytest header comment say "24" write actions; there
  are 25 (ledger/docs.md correctly say 25).
- minor — hooks/github/hooks.go createPullRequest copies every non-meta/reviewer record field into
  the POST body; legacy's issue-XOR-title branch (github.go:1021-1039) sends `issue` OR
  `title`(+`body`), never both. A record carrying both (legacy-accepted; legacy ignores title) gets
  a different wire body. Edge of G3/G4's documented permissiveness — fold into the G3 ledger text.

Line-cap adjudication (hooks.go = 363): **accept with wording change** — see cross-cutting §C1.

## gmail — pass

checked: schema_fidelity_spot_checks=4 (messages/threads/drafts/labels vs streams.go mapRecord fns
— field-for-field match incl. drafts' nested message reach-in), write_actions_verified=n/a
(read-only, matches legacy ErrUnsupportedOperation), fixtures_realistic=true,
escape_hatches_justified=true (AuthHook = OAuth2 refresh-token grant, a named trigger; hook is a
near-verbatim port of auth.go's oauthRefreshAuth incl. 60s margin, 3600s default ttl, injectable
clock; https-only token_url tightening is a documented, justified deviation),
secret_scan_clean=true, conventions_adherent=mostly (F6 finding below).

Bundle-level skip marker honesty: **genuine**. Sole auth candidate is mode:custom; conformance's
synthetic config can never carry an https token_url, and the only way to "fix" it would be
inventing a bearer fallback legacy doesn't have or weakening the https guard — both forbidden.
Marker reason correctly names paritytest/gmail as the substitute, which drives the REAL AuthHook
(TLS token server, header parity after refresh, token-endpoint failure path). Parity-test
integrity is high: the stringification deviation is coerced ONLY via an enumerated 4-field helper
AND pinned by a dedicated type-assertion test — the required pattern (see adjudication A2).

Findings:
- **major — internal/connectors/defs/gmail/docs.md (Known limits, 3rd bullet)**: stale and now
  FALSE. It still states conformance "always call[s] engine.Check/engine.Read with a nil Hooks
  argument (by design — conformance has no per-connector hook wiring mechanism)" and that github's
  bearer candidate "masks" the same gap. R3 made conformance hook-aware (dynamic.go passes
  engine.HooksFor, blank-imports hooks/hookset) and replaced this exact framing with the
  metadata.json skip marker this bundle already carries; github now has FULL dynamic hook coverage,
  the opposite of what this bullet claims. r3-ledger.md lists defs/gmail/docs.md as updated, but
  the old ENGINE_GAP text survived. Fan-out agents are told to model on pilot docs — a false
  engine-behavior claim here propagates. Fix: rewrite the bullet around the marker (the marker's
  own reason text is accurate).
- **major — internal/connectors/defs/gmail/spec.json**: dead config keys, F6 violation with
  misleading operator-facing text. `max_pages` ("use 0, all, or unlimited to exhaust the stream")
  is consumed by NOTHING — PaginationSpec is static; an operator setting max_pages=5 silently gets
  unbounded. `mode` ("fixture for credential-free conformance") is a legacy-only affordance
  (gmail.go fixtureMode); no engine path reads it. Both are undocumented in Known limits (unlike
  `start_date`/`include_spam_and_trash`, which at least carry an explicit forward-compat note —
  itself in tension with the searxng/github F6 precedent of NOT declaring unwired keys; P-12 must
  pick one rule and apply it wave-wide). Fix: drop `max_pages`/`mode` (and either drop or clearly
  Known-limits-mark the two filter keys per the P-12 decision).
- minor — hooks/gmail/hooks.go interpolateOptional (lines 112-128): doc comment claims CRLF/
  unknown-filter errors "still propagate"; the code returns "" on ANY error. Behavior is safe
  (worst case omits client_secret → loud invalid_client downstream) but the comment lies. Align
  code or comment.
- minor — legacy's `refresh_token` secret alias (gmail.go refreshTokenSecret) and dotted
  `credentials.*` secret forms are dropped, undocumented (same class as github's alias finding,
  smaller blast radius). Ledger it.

## monday — pass

checked: schema_fidelity_spot_checks=5 (all 5 streams vs streams.go record fns — hook mapping
functions are byte-for-byte ports incl. stringField id coercion and itemRecord group/board
hoisting; GraphQL selections diffed exactly), write_actions_verified=n/a (read-only),
fixtures_realistic=true (retained as record-shape documentation; marker-skipped for replay,
honestly labeled as such in docs.md), escape_hatches_justified=true (StreamHook: GraphQL POST
reads with in-body pagination — the canonical named trigger, SPEC-pre-identified; CheckHook:
GraphQL check; exactly 2 interfaces), secret_scan_clean=true, conventions_adherent=true.

Per-stream skip markers (all 5): **genuine**. StreamHook returns handled=true unconditionally for
every declared stream; a replay server cannot match per-page GraphQL POST bodies (fixture matching
is method/path/query — pagination state lives in the body). Nothing marked skipped could run. The
R3 simplification genuinely removed the fictional GET-shaped shadow pagination (streams.json now
carries only honest metadata), and docs.md documents the never-live declarative path plainly.
monday docs.md also explicitly documents the one auth divergence I independently found (no-token
config: legacy hard-errors pre-request, bundle falls to mode:none and surfaces monday's 401 via
error_map) — exemplary disclosure.

Findings:
- minor — hooks/monday/hooks.go mondayPageSize/mondayMaxPages "never error" (invalid values fall
  back to default/unbounded) vs legacy's hard errors (monday.go:446-474). More-permissive-only, and
  self-documented in a code comment, but absent from docs.md/ledger. A typoed page_size now
  silently syncs with 50 instead of erroring. Ledger it (or restore legacy's error behavior —
  sentry's hook kept it, so the two hook ports are inconsistent on this point).
- minor — spec.json omits `max_pages` even though the hook consumes config.max_pages (line 73,
  262) — the inverse of gmail's dead-key problem: a consumed-but-undeclared key an operator cannot
  discover. Declare it (it is genuinely wired here, unlike gmail's).
- minor — api_surface.json declares GraphQL READ endpoints as method GET (real wire verb POST) to
  satisfy validate's covered-mutation rule; documented inline in scope prose. Acceptable
  workaround, but P-12 should add a sanctioned representation for GraphQL surfaces so fan-out
  agents don't improvise.

Line-cap (hooks.go = 340): **accept with wording change** — §C1.

## sentry — pass

checked: schema_fidelity_spot_checks=4 (projects/issues/events/releases schemas vs legacy mapRecord
— raw-name projection, no renames needed; fixture shapes match Sentry's real wire incl. count as
string), write_actions_verified=n/a (read-only), fixtures_realistic=true,
escape_hatches_justified=true (StreamHook: Link-header `results=` twist — the resolution ladder was
followed with direct code evidence, rung 1 rejected for a REAL reason (extra trailing request is a
hard 404/HTTPError, not benign), nextCursor/splitAttr/cursorFromURL are byte-for-byte ports),
secret_scan_clean=true, conventions_adherent=mostly.

Per-stream skip markers (all 4): **genuine**. Two independent reasons: (a) the hook faithfully
errors on conformance's non-integer synthetic page_size exactly as legacy would (weakening that to
pass replay is forbidden), and (b) fixtureResponse carries status/body only — no response headers —
so a 2-page Link-header/results= replay is inexpressible regardless. Parity-test integrity is the
strongest in the batch: RAW reflect.DeepEqual on records with NO stripping/normalization at all,
2-page results=false stop, bearer header parity, non-2xx parity, hostile-base-URL fail-closed
parity.

Findings:
- **major — internal/connectors/defs/sentry/spec.json + docs.md**: `hostname` is dead config and
  docs.md's claim is false. Legacy derives the base URL from `hostname` (default sentry.io) as
  https://<hostname> when base_url is unset (sentry.go:293-308); the bundle's base url is
  `{{ config.base_url }}` and NO template references hostname, so a legacy-canonical config
  (auth_token + organization + project, no base_url) hard-errors "unresolved key base_url" instead
  of reading sentry.io. docs.md Auth setup asserts "matching legacy's sentryBaseURL resolution" —
  untrue. This is an accepted-input regression + F6 dead key + misdocumentation. Fix now: mark
  base_url required in spec.json, drop hostname, correct docs.md, ledger the config-surface change.
  Real fix: the batch-level default/derived-base-URL engine decision (§C3).
- minor — base.check omits legacy Check's `per_page=1` query (sentry.go:89) — check request shape
  differs; harmless, ledger-level note.
- minor — streams.json still carries the pre-R3 static `query: {"per_page": "100"}` on every
  stream; inert (hook builds its own per_page) but fictional — monday's R3 cleanup removed its
  equivalents; remove for consistency.

## chargebee — pass

checked: schema_fidelity_spot_checks=5 (all 5 schemas vs chargebeeStreams field lists —
field-for-field, stringification documented per-property in-schema),
write_actions_verified=n/a (read-only), fixtures_realistic=true (real envelope wire shape, numeric
unix-seconds cursors as JSON NUMBERS — the required B1/B2 shape; 2-page offset/next_offset
propagation recorded), escape_hatches_justified=true (none used; the RecordHook-rejection analysis
in the ledger is correct engine-reading, not convenience), secret_scan_clean=true,
conventions_adherent=true.

Parity-test integrity: the stringify-normalizing helper is at the outer boundary of acceptability
(it coerces EVERY field to string form) but is saved by (a) value equality still being asserted
per-field, and (b) TestParityChargebee_ComputedFieldsStringifyNumericAndBooleanFields pinning the
exact type deviation (legacy json.Number/bool vs engine string) so no regression can hide.
Incremental parity is the wave's model: app-persisted digit-string cursor AND RFC3339 start_date
fallback both asserted to the exact wire value on both sides.

Findings:
- **major — internal/connectors/defs/chargebee/streams.json**: incremental requests omit
  `sort_by[asc]=updated_at`. Legacy sets it alongside `updated_at[after]` on every incremental
  request (chargebee.go:152-154); the engine has no conditional-query mechanism (a static query key
  would send it on full-refresh reads too, which legacy doesn't). Emitted record SET is unchanged
  (app-side cursor is a max over emitted records) but the wire request and record ORDER for
  incremental syncs diverge from legacy, undocumented — the parity test captures only
  `updated_at[after]` so it can't see this. Fix: ledger + docs.md entry now; real fix rides batch
  B's already-mandated optional/conditional-query ENGINE_GAP increment (this is another occurrence
  for that counter — a param sent only-when-incremental is the same conditional-query class).
- **major — internal/connectors/defs/chargebee/spec.json + docs.md**: `site` is dead config and
  docs.md's claim ("derived from the required-ish site config value as
  https://{site}.chargebee.com/api/v2 — base_url wins when both are set, matching legacy's
  chargebeeBaseURL") is false — nothing consumes `site`; a site-only legacy config hard-errors on
  base_url. Identical class to sentry's hostname finding (§C3). Same fix shape.

---

## Adjudications

### A1 — computed_fields stringification (chargebee ~30 fields, gmail 4 fields, github 4 fields)

**Ruling: NOT acceptable as a fan-out pattern. Typed extraction is REQUIRED before Pass A
fan-out. The three pilot bundles may stand for the pilot only, as documented deviations, and must
be re-tightened once the engine feature lands.**

Reasoning: the §5 meta-rule protects "emitted record DATA". JSON type IS record data: legacy
publishes these fields as integer/boolean (chargebee `created_at` int64, `deleted` bool; gmail
`messages_total` json.Number; github `user_id` number) and every warehouse destination derives
column types from them. The engine emits strings. At the wave6 flip every existing destination
table for these streams takes a column-type change (INTEGER→VARCHAR, BOOLEAN→VARCHAR); numeric
aggregation/filter semantics silently degrade. "Textually identical information" holds only for a
human reader, not for a typed consumer — and the chargebee ledger itself concedes it is "a real
type-shape change". What makes the pilots acceptable AS PILOTS is process: single honest schema
types (["string","null"], never widened unions), per-property wire-type descriptions, explicit
lock-in tests, ledger entries. That is the correct way to carry a known-bad constraint, not a
license to replicate it ~500 times: envelope unwraps and camelCase renames are ubiquitous in the
remaining inventory, so the blast radius compounds with every fan-out batch.

Recurrence math: chargebee (envelope unwrap) + gmail (rename) + github (nested-id flatten, G0b)
= 3 independent same-wave occurrences → conventions §6's "ENGINE_GAP recurs ≥3 times → mini wave-0
engine increment" threshold is MET. Required increment (matches the chargebee ledger's own
candidate design): when a computed_fields template is a single bare `{{ record.<path> }}` with no
filters, copy the raw typed value into the projected record instead of routing through
Interpolate's stringify; filter chains and multi-part templates keep string semantics. Then:
chargebee/gmail/github schemas revert to real wire types, the stringify lock-in tests flip to
raw-type equality, and the §5 ledger entries move to RESOLVED — before fan-out dispatch.

### A2 — the PRINCIPLE for parity tests stripping/normalizing fields before comparison

(calendly's concrete instance was adjudicated FAIL in REVIEW-B; this is the batch-A ruling on the
general rule both batches feed to P-12.)

**Ruling: field exclusion or coercion in a parity comparison is acceptable ONLY when ALL four
hold:**
1. **Ledgered**: the strip corresponds 1:1 to a §5-ledgered, docs.md-documented deviation (dropped
   field or type change) — never an undocumented convenience.
2. **Genuinely unproducible**: the engine dialect actually cannot produce the field today (github's
   labels_count/assignees_count/assets_count — no length/count filter; is_pull_request — no typed
   boolean literal). If the engine COULD produce it (calendly's `id` = last path segment of a
   field the record carries), stripping it masks a fixable gap → FAIL, per REVIEW-B.
3. **Visible and enumerated**: implemented as a named, per-field helper citing the docs
   (github's isDocumentedDrop/isStringifiedNestedID; gmail's 4-key stringifyCountField loop) —
   never a blanket normalize-the-whole-record transform that would also hide unrelated drift.
   chargebee's stringify-everything helper is the outer boundary: tolerable only because values
   are still compared field-for-field and (4) holds; fan-out template should use the enumerated
   form.
4. **Pinned by a companion assertion**: a dedicated test asserts the deviation itself (field
   absent / exact type change, e.g. TestParityGmail_ComputedFieldsStringifyLabelCountFields,
   TestParityChargebee_ComputedFieldsStringifyNumericAndBooleanFields,
   github's string-form-only compare limited to 4 named ids) so the strip cannot silently absorb a
   future regression.

Additionally: strips must never be bidirectional cover for ENGINE-only extra fields (batch-A
github's per-field loop iterates legacy keys only — one-directional; acceptable because schema
projection bounds the engine side, but the P-12 template should compare key SETS minus the
enumerated drops). All batch-A strips satisfy 1–4.

### A3 — github `repository` marker field dropped (ENGINE_GAP G0: computed_fields cannot reference config.*)

**Ruling: correct pilot-scope decision; blocker-class engine gap for fan-out. Close G0 in the same
mini wave-0 increment as A1, then restore the field.**

Reasoning: within the pilot's options the executor chose right — a 3rd hook interface (RecordHook)
would breach the Tier-2 cap for a convenience field; inventing declarative syntax is forbidden; the
drop is typed (ENGINE_GAP), ledgered (G0), docs-documented, and identity-safe (PKs are
node_id/id/sha/name — never `repository`). But fan-out impact is real: legacy stamps the marker on
EVERY record of ALL 19 streams, so at the wave6 flip a column silently disappears from every github
destination table; any consumer unioning multi-repo syncs by `repository` breaks with no
deprecation path. And the gap class is guaranteed-recurrent across ~500 connectors: "stamp the
connector's config scope onto each record" (account_id, site, org/project, workspace) is a standard
legacy pattern — searxng's static `stream` marker already proved marker fields are load-bearing;
config-derived markers are the same pattern one notch up. The fix is small and contained: wire
`Config` (and `Secrets` should stay EXCLUDED — a computed field must never be able to copy a secret
into a record; note this explicitly in the increment's threat-model line) into
`applyComputedFields`' Vars in engine/read.go, static-validate the references via ResolveCheck,
then `"repository": "{{ config.owner }}/{{ config.repo }}"` restores parity and G0 moves to
RESOLVED. Count G0 toward the §6 recurrence threshold together with A1's increment — one combined
engine mini-wave covers both since they touch the same function.

---

## Cross-cutting (batch-level) findings & flags

### C1 — hooks.go line caps: github 363, monday 340 vs "~300"

**Ruling: accept both with a conventions wording change (no split, no Tier-3 escalation).**
conventions §1 is self-contradictory ("hard-capped at ~300 lines" — a tilde is not a hard cap).
Both files were audited line-by-line: github carries 2 mandated interfaces (full RS256
JWT/installation exchange + 4 compound writes, each a named parity-floor item); monday carries 2
pagination shapes, 5 record mappers, and the GraphQL-errors-in-HTTP-200 envelope — no fluff,
no dead code; both ledgers self-reported the overrun honestly. Splitting into a second .go file to
game a single-file count is worse than the overrun; Tier-3 for ~13-21% overage is
disproportionate. P-12 wording: "~300 soft target; >300 requires a self-reported justification in
the trace ledger; 400 is a hard ceiling; >400 or a 3rd interface → Tier 3." (Both files sit under
400.) sentry (281) and gmail (267) are under target.

### C2 — skip_dynamic marker honesty: VERIFIED for all three marked connectors

Nothing is marker-skipped that could genuinely run in replay: gmail (custom-auth-only + https-only
token_url vs synthetic config), monday (in-body GraphQL pagination unmatchable by
method/path/query replay), sentry (legacy-faithful page_size error on synthetic config AND
fixtureResponse has no header support for Link-header replay). Every marker reason names its real
substitute and those substitutes were re-run green uncached. github's zero-marker full-dynamic
coverage (AuthHook + WriteHook live in replay, incl. the response-block fixture fix) is the
correct worked example for fan-out. One watch-item: sentry's replay limitation is partly a HARNESS
limitation (no response headers, no spec-default-aware config synthesis) — if the harness ever
gains those, the markers should be revisited rather than grandfathered.

### C3 — FLAG (batch-level, pre-fan-out decision required): spec defaults are annotation-only;
legacy in-code base-URL defaults/derivations have no engine home

Every batch-A bundle sets `base.url: {{ config.base_url }}`; the engine never materializes
spec.json `default` values into RuntimeConfig, and no app-layer merge exists (verified: no
default-materialization code outside connector-local legacy helpers). Every legacy connector had
in-code defaults (github api.github.com, gmail base+token URLs, monday api.monday.com/v2,
chargebee site-derived, sentry hostname-derived) — so EVERY migrated connector currently
hard-errors on a config shape legacy accepted. The wave0 stripe golden set this pattern, so batch-A
connectors are not individually faulted beyond the sentry/chargebee misdocumentation majors — but
this MUST be decided before fan-out (and hard-blocks the wave6 registry flip): either (a) an
engine/app increment materializing spec defaults into config (recommended; also subsumes gmail's
token_url default), plus a template-level answer for DERIVED urls (sentry hostname / chargebee
site — e.g. default-bearing config refs or a documented "base_url required, derivation dropped"
convention with a config-migration step), or (b) a convention that base_url is always `required`
with no default — applied uniformly and ledgered per-connector. Today's state (optional,
default-annotated, unconsumed) is the worst of both.

### C4 — FLAG (process): phase gate artifacts are hollow at review time

TDD-GATE.json is `passed: true` with EMPTY tasks/behaviorTasks; TDD-LEDGER.md is a stub;
SUMMARY.md is "TBD"; VERIFICATION.md carries no recorded run output or HEAD. The real red-first
evidence EXISTS and is substantive (per-task trace ledgers carry genuine RED transcripts whose
failure signatures I cross-checked against the code: embed-pattern failures, undefined-symbol
compile errors, the gmail unresolved-key transitions), and I re-ran the full gate locally — but a
`passed: true` gate with zero task rows is exactly the stubbed-gate shape this review is required
to reject as evidence. NOT a code defect; P-14 must populate TDD-GATE.json task rows (P-0..P-10),
refresh VERIFICATION.md with the actual command outputs + HEAD, and write SUMMARY.md before the
phase may close. Do not close on the current artifacts.

### C5 — Fan-out-readiness assessment for the patterns batch A establishes

Ready to template now: the paritytest/<name> package shape (RED-first bundle-missing failure,
shared-httptest both-sides drive, RAW DeepEqual bar, app-persisted digit-cursor round trip —
chargebee is the reference); the deviation-ledger discipline (github's G-table is the model); the
skip-marker rule + its honesty bar (github as the prefer-full-coverage example, gmail/monday/sentry
as the three legitimate marker shapes); the write-fixture `response` block for WriteHooks; the
missing_ok_status "don't add leniency legacy lacked" rule; hook ports that keep legacy's exact
config-validation strictness (sentry yes, monday drifted — pick sentry's rule).

NOT ready — hold fan-out until closed (one combined engine mini-wave + P-12 docs pass):
1. typed computed_fields extraction (A1 — met the ≥3 recurrence bar);
2. config.* in computed_fields Vars (A3/G0, secrets excluded);
3. conditional/optional query params (REVIEW-B's mandated increment; chargebee's sort_by and the
   github G8 / gmail start_date filter drops are further occurrences of the same class);
4. the C3 base-URL default/derivation decision;
5. conventions wording: line-cap (C1), F6 dead-key rule applied ONE way (gmail vs github/searxng
   currently contradict), config-surface-change ledger requirement (auth aliases, auth_type,
   secret key renames — three connectors hit it this wave), GraphQL api_surface representation
   (monday's GET annotation).

Fan-out before items 1–4 land would replicate, at ~500-connector scale, exactly the deviations
this batch had to hand-document per connector.

---

## Gap-loop routing

- backend (defs/docs edits, no gate changes): github spec/docs auth-surface ledger + color-strip
  entry + "24→25" typos + close_* fixture bodies; gmail docs.md stale bullet + dead spec keys;
  sentry hostname/base_url spec+docs correction + stale per_page query cleanup; chargebee
  sort_by ledger + site/docs correction; monday max_pages spec declaration + permissive-parse
  ledger note.
- tester: github since-param/incremental parity tests; github compound-write body assertions.
- orchestrator/P-12: adjudications A1/A2/A3, flags C1/C3/C5 (engine mini-wave scoping), C4 (P-14
  bookkeeping).

## Machine-readable verdicts

```json
{
  "verdicts": {"github": "pass", "gmail": "pass", "monday": "pass", "sentry": "pass", "chargebee": "pass"},
  "adjudications": [
    {"id": "A1", "topic": "computed_fields stringification (chargebee/gmail/github)", "ruling": "not acceptable at fan-out scale; typed-extraction engine increment REQUIRED before Pass A (>=3 recurrence met); pilots stand as documented deviations, re-tighten after"},
    {"id": "A2", "topic": "parity-test field stripping principle", "ruling": "acceptable only if ledgered + genuinely unproducible + enumerated/visible + pinned by companion assertion; batch A compliant; calendly-style producible-field strips FAIL"},
    {"id": "A3", "topic": "github repository marker drop (G0)", "ruling": "right pilot call; blocker-class for fan-out; wire Config (not Secrets) into applyComputedFields Vars in same mini-wave, then restore field"}
  ],
  "blocks": [
    "fan-out (Pass A) blocked until: typed computed_fields extraction (A1), config-in-computed_fields (A3), conditional-query-param increment (REVIEW-B #2 + chargebee sort_by), base-url default/derivation decision (C3)",
    "phase close blocked until P-14 populates TDD-GATE.json task rows / VERIFICATION.md run evidence / SUMMARY.md (C4)"
  ],
  "flags": [
    "github: auth_type + secret-alias config surface dropped undocumented (silent unauthenticated fallback path) — major, gap loop",
    "github: state-cursor incremental filtering is new behavior vs legacy, docs claim 'matches legacy exactly', no since-param parity test — major, gap loop",
    "github: create_label/update_label color '#'-strip not reproduced, undocumented — major, gap loop",
    "gmail: docs.md Known-limits bullet is stale/false post-R3 (claims hooks-blind conformance) — major, gap loop",
    "gmail: dead spec keys max_pages/mode with misleading descriptions (F6) — major, gap loop",
    "sentry: hostname dead config + false docs claim of legacy base-URL derivation — major, gap loop",
    "chargebee: incremental sort_by[asc]=updated_at omitted vs legacy, undocumented — major, gap loop",
    "chargebee: site dead config + false docs derivation claim — major, gap loop",
    "line caps 363/340 accepted with conventions wording change (C1)",
    "skip_dynamic markers verified honest for gmail/monday/sentry; github full-dynamic is the fan-out example (C2)"
  ]
}
```

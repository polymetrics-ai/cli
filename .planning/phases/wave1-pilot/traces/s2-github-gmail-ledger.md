# Gap-loop cycle-1 — Step 2 (pilot repair wave): github + gmail — TDD ledger

Executor: gsd-loop-backend. HEAD at start: dc7ad63, branch connector-architecture-v2 (includes Step
1's engine mini-wave: typed computed_fields extraction, config.* in computed_fields, optional-query
dialect, last_path_segment filter, token_path stop_path+loop guard, spec default materialization —
`.planning/phases/wave1-pilot/traces/gaploop-s1-ledger.md`).

Scope: `internal/connectors/defs/github/**`, `internal/connectors/hooks/github/**`,
`internal/connectors/paritytest/github/**`, `internal/connectors/defs/gmail/**`,
`internal/connectors/paritytest/gmail/**`, this ledger. `hooks/gmail/**` untouched (read-only per
dispatch; no fix in this task required touching it — see gmail section). No engine/cmd/connsdk
edits (Step 1 territory, already closed) beyond what static validation already supports.

Mandated by GAP-LOOP-PLAN.md Step 2's github/gmail bullets, cross-referenced against
REVIEW-A.md's github findings (majors 1-3, minors) and gmail findings (majors 1-2).

Baseline (before any edit in this task): `go build ./...` clean; `go run ./cmd/connectorgen
validate internal/connectors/defs` → 13 connectors, 0 findings; `go test
./internal/connectors/paritytest/github/... ./internal/connectors/hooks/github/...` all PASS;
`go test ./internal/connectors/paritytest/gmail/...` → exactly 1 FAIL,
`TestParityGmail_ComputedFieldsStringifyLabelCountFields` (the Step-1-predicted, Step-2-scoped
breakage — engine now emits `json.Number` where the test still asserts the pre-increment
stringified form). No other pilot suite affected.

---

## GITHUB

### Item 1 — restore the `repository` marker field via config-capable computed_fields (ledger G0)

RED test added first: `TestParityGithub_RepositoryMarkerFieldRestored` in
`internal/connectors/paritytest/github/parity_test.go` (also removed `"repository"` from
`isDocumentedDrop`, which had been silently stripping it from the generic per-field comparison
loop in `TestParityGithub_StreamRecords`).

```
$ go test ./internal/connectors/paritytest/github/... -run 'TestParityGithub_RepositoryMarkerFieldRestored|TestParityGithub_StreamRecords' -v
--- FAIL: TestParityGithub_StreamRecords (stream "repository"/"issues"/"pull_requests"/"workflows" record 0 missing field "repository" in engine output (legacy=octocat/hello-world))
--- FAIL: TestParityGithub_RepositoryMarkerFieldRestored (engine stream ... repository = <nil> (<nil>), want string "octocat/hello-world")
```

GREEN: per Step-1's now-available `config.*` in `computed_fields` (A3/G0 RESOLVED), added
`"repository": "{{ config.owner }}/{{ config.repo }}"` to every one of the 19 streams'
`computed_fields` in `internal/connectors/defs/github/streams.json`, and a `repository` schema
property (documented, `type: string`) to all 19 `internal/connectors/defs/github/schemas/*.json`
files (matching the existing convention that every computed field is schema-declared, e.g.
`relation` on collaborators/contributors/stargazers/subscribers).

```
$ go run ./cmd/connectorgen validate internal/connectors/defs
connectorgen validate: 13 connector(s) checked, 0 findings

$ go test ./internal/connectors/paritytest/github/... -run 'TestParityGithub_RepositoryMarkerFieldRestored|TestParityGithub_StreamRecords' -v
--- PASS (all subtests)
```

`docs.md` updated: Known limits' `ENGINE_GAP` bullet flipped to "G0 — RESOLVED" with the fix
description; Streams-notes bullet updated to say the marker IS reproduced.

### Item 2a — auth_type + secret-alias config surface: no silent mode:none fall-through (ledger G14)

Legacy read in full (`internal/connectors/github/auth.go`): `auth_type`/`auth`/`authentication`
explicit mode selection (many synonyms per mode, auth.go:61-96); token secret aliases
(`personalAccessToken`/`accessToken`/`oauthToken`/`installationToken`/`githubToken`/`GITHUB_TOKEN`,
github.go:1634-1644); private-key aliases (`privateKey`/`githubAppPrivateKey`/`privateKeyBase64`/
`githubAppPrivateKeyBase64`); app-id aliases (`client_id`/`github_app_id`, auth.go:256-257). None of
these are reproduced by the bundle (only canonical `token`/`private_key`/`private_key_base64`/
`app_id` are read) — the dangerous part REVIEW-A.md's major flagged is that the bundle's
pre-existing `base.auth` chain (`bearer` when `token` truthy → `custom` github_app when `app_id`
truthy → unconditional `{"mode":"none"}`) let a caller who set ONLY an alias-shaped secret (e.g.
`personalAccessToken`, not `token`) fall through both real candidates and reach `mode:none` —
silent unauthenticated reads, zero error.

RED tests added first (`internal/connectors/paritytest/github/parity_test.go`):
`TestParityGithub_AuthNoCredentialsFailsLoudRatherThanSilentlyPublic` (no token, no app config, no
opt-in → must hard-error) and `TestParityGithub_AuthExplicitPublicOptIn` (lock-in: an explicit
opt-in must still work).

```
$ go test ./internal/connectors/paritytest/github/... -run 'TestParityGithub_AuthNoCredentialsFailsLoudRatherThanSilentlyPublic|TestParityGithub_AuthExplicitPublicOptIn' -v
--- FAIL: TestParityGithub_AuthNoCredentialsFailsLoudRatherThanSilentlyPublic
    engine Read with no credentials and no explicit auth_type = nil error, want a hard failure
--- PASS: TestParityGithub_AuthExplicitPublicOptIn   (already true — auth chain's unconditional
    mode:none happened to satisfy it before the fix too, kept as a lock-in for the fixed shape)
```

**Design constraint discovered mid-fix (recorded, not worked around):** the natural declarative fix
is gating the `mode:"none"` candidate on `{{ config.auth_type in ['public','none','anonymous',
'unauthenticated'] }}` (matching legacy's exact enum via the `when` grammar's documented `in`
operator, conventions.md §3). This FAILS `connectorgen validate` —
`engine.ResolveCheck` (reused verbatim for every `when` field, not just ordinary interpolation
templates) parses the ENTIRE `{{ }}` inner content as a single `namespace.key` reference split on
`.`; it has no `in`/`==` grammar awareness at all (confirmed: EVERY existing `when` clause in every
bundle in this repo — github's own token/app_id gates, monday, searxng, zendesk-support — is bare
truthiness only, zero prior `in`/`==` usage exists in any production bundle). Fixing
`ResolveCheck` itself is an `engine/cmd` change outside this task's file scope (not one of Step 1's
7 mandated items, and Step 1 is already closed/committed) — flagged here as a genuine follow-up
`ENGINE_GAP` for P-12/a future engine increment (`ResolveCheck` needs `when`-aware static parsing
for `==`/`in`, not just bare truthiness), NOT worked around by silently declaring an
unvalidatable/unchecked auth field.

**GREEN (within current static-validation constraints):** introduced a dedicated boolean-shaped
`spec.json` opt-in key, `public_access` (any non-empty value = explicit opt-in), and gated the
`none` candidate on its bare truthiness — statically valid under the current `ResolveCheck`, and
closes the actual security-relevant gap (no more silent fallthrough for absent/typo'd/alias-shaped
credentials) even though it does not reproduce legacy's exact `auth_type=public` string-value
selection 1:1. `internal/connectors/defs/github/streams.json`:
`{"mode":"none","when":"{{ config.public_access }}"}` (was unconditional `{"mode":"none"}`).
`internal/connectors/defs/github/spec.json`: added `public_access` (documented narrowing + the
security rationale), dropped the earlier abandoned `auth_type`-as-wired-key attempt (never
committed as a validate-passing state).

```
$ go run ./cmd/connectorgen validate internal/connectors/defs
connectorgen validate: 13 connector(s) checked, 0 findings

$ go test ./internal/connectors/paritytest/github/... -run 'TestParityGithub_Auth' -v
--- PASS: TestParityGithub_AuthGithubAppInstallationTokenBearerHeader
--- PASS: TestParityGithub_AuthNoCredentialsFailsLoudRatherThanSilentlyPublic
--- PASS: TestParityGithub_AuthExplicitPublicOptIn
```

`docs.md` updated: "Auth setup" section rewritten (candidate list + new "Auth config surface vs
legacy" paragraph naming every dropped alias, the silent-fallthrough hazard, the fix, and the
`in`/`==` static-validation limitation); "Known limits" gained a matching bullet.

### Item 2b — docs incremental honesty + since-param parity test (ledger note)

Legacy's `since` incremental filter (github.go) reads `req.Config.Config["since"]` ONLY — grep
confirms no `req.State` reference anywhere in the legacy github package. The bundle's
`streams.json` `incremental.request_param: "since"` goes through the ENGINE-WIDE
`incrementalLowerBoundValue` (engine/read.go), which prefers an app-persisted STATE cursor over
`start_config_key` when state is present — correct engine-wide semantics, but docs.md previously
claimed "matches legacy exactly" unconditionally, which is false for the state-cursor path, and no
test covered `since` forwarding at all (config-path equality or the state-cursor round trip).

Two tests added to `internal/connectors/paritytest/github/parity_test.go` (both pin EXISTING
behavior — no engine/bundle code change needed here, this is a docs+test gap, per the review's own
framing "direction is an improvement... NO parity test covers since forwarding at all"):
- `TestParityGithub_SinceConfigOnlyMatchesLegacy` — config `since` -> query param, asserted equal
  on both connectors (the shape legacy actually supports).
- `TestParityGithub_SinceStateCursorForwardingIsEngineOnlyBehavior` — app-persisted state cursor:
  legacy ignores it (`since` query param stays empty), engine forwards it. Pins the documented
  divergence so a future regression on either side is caught.

```
$ go test ./internal/connectors/paritytest/github/... -run 'TestParityGithub_Since' -v
--- PASS: TestParityGithub_SinceConfigOnlyMatchesLegacy
--- PASS: TestParityGithub_SinceStateCursorForwardingIsEngineOnlyBehavior
```

`docs.md`'s "Streams notes" `issues` bullet rewritten: removed the blanket "matches legacy exactly"
claim, added the config-path/state-path split with test cross-references.

### Item 3 — create_label/update_label '#'-color-strip normalization (ledger G16)

Legacy (`github.go:1120,1133`) does `strings.TrimPrefix(color, "#")` before sending
`create_label`/`update_label` bodies; the bundle's declarative write path passes record fields
verbatim, so a caller-supplied `"#ff0000"` (a value legacy explicitly accepts and normalizes, and
GitHub's own docs/UI commonly show with the leading hash) reached the wire unstripped and GitHub's
API 422s. Per dispatch ("RecordHook-in-WriteHook or writes.json — legacy is ground truth"):
implemented via the EXISTING `WriteHook` interface (`hooks/github/hooks.go`'s `ExecuteWrite`
already implements `WriteHook` for the 4 compound actions — adding 2 more `case`s is not a 3rd hook
interface, stays within the Tier-2 cap).

RED tests added first (`internal/connectors/paritytest/github/parity_test.go`):
`TestParityGithub_WriteCreateLabelStripsLeadingHashFromColor`,
`TestParityGithub_WriteCreateLabelColorWithoutHashUnaffected` (lock-in, no-op case),
`TestParityGithub_WriteUpdateLabelStripsLeadingHashFromColor`,
`TestParityGithub_WriteUpdateLabelNoColorFieldOmitsColor` (lock-in: never invents a color key).

```
$ go test ./internal/connectors/paritytest/github/... -run 'TestParityGithub_Write.*Label' -v
--- FAIL: TestParityGithub_WriteCreateLabelStripsLeadingHashFromColor
    engine request body color = "#ff0000", want "ff0000" (legacy)
--- PASS: TestParityGithub_WriteCreateLabelColorWithoutHashUnaffected  (no-op case, already correct)
--- FAIL: TestParityGithub_WriteUpdateLabelStripsLeadingHashFromColor
    engine request body color = "#00ff00", want "00ff00" (legacy)
--- PASS: TestParityGithub_WriteUpdateLabelNoColorFieldOmitsColor  (already correct)
```

GREEN: `internal/connectors/hooks/github/hooks.go` — added `case "create_label"`/`case
"update_label"` to `ExecuteWrite`'s switch, plus `createLabel`/`updateLabel` functions that build
the same JSON body the declarative path would (name/color/description for create; new_name/color/
description optional-field shape for update, matching ledger G3's "at least one mutable field"
permissiveness) but with `strings.TrimPrefix(color, "#")` applied, exactly like legacy's
`githubCreateLabelPayload`/`githubUpdateLabelPayload`.

```
$ go test ./internal/connectors/paritytest/github/... -run 'TestParityGithub_Write.*Label' -v
--- PASS (all 4)
```

Also added direct hook-level unit tests to `internal/connectors/hooks/github/hooks_test.go`:
`TestExecuteWrite_CreateLabelStripsLeadingHash`, `TestExecuteWrite_CreateLabelMissingColorErrors`,
`TestExecuteWrite_UpdateLabelStripsLeadingHashWhenColorPresent`,
`TestExecuteWrite_UpdateLabelMissingNameErrors` — all PASS.

**Line-cap discipline (conventions.md §C1 "400 is a hard ceiling"):** the 2 new cases pushed
`hooks.go` from 363 to 424 lines. Trimmed comments/whitespace (package doc, Authenticator doc,
ExecuteWrite doc, createLabel/updateLabel doc, `updateLabel`'s 3 near-identical optional-field ifs
collapsed into one loop) back down to exactly **400 lines** — at the hard ceiling, not over it — with
zero functional/behavioral change (re-ran the full hooks+paritytest suite green after each trim
pass). No 3rd hook interface introduced; still exactly AuthHook+WriteHook (Tier-2 cap).

### Item 4a — docs incremental honesty (folded into item 2b above)

Covered by item 2b's `since` docs fix (the dispatch's "docs incremental honesty" bullet under
item (2) and the since-param test are the same finding).

### Item 4b — stale G0b ledger prose fix

`p9-github-ledger.md`'s G0b row (and this bundle's earlier undocumented assumption) described
`user_id`/`author_id`/`committer_id`/`workflow_run_id` (bare single-reference `computed_fields`
templates, e.g. `"{{ record.user.id }}"`) as permanently stringified, with schemas widened to
`["integer","string"]` and the parity suite comparing them string-form-only via
`isStringifiedNestedID`. Step 1's typed `computed_fields` extraction (bare single `{{ record.path
}}` now preserves native JSON type) already made this claim STALE — the fields have been emitting
native `json.Number` since dc7ad63, but nothing in defs/paritytest had been retightened to match
(github's suite stayed green throughout only because `isStringifiedNestedID`'s
`fmt.Sprint`-based comparison tolerates either type transparently, per gaploop-s1-ledger.md's own
note).

Added `TestParityGithub_NestedIDComputedFieldsEmitNativeNumbers` (asserts `json.Number` directly,
bypassing the masking helper) — passed immediately (lock-in test, confirms the stale claim, not a
behavior bug):

```
$ go test ./internal/connectors/paritytest/github/... -run TestParityGithub_NestedIDComputedFieldsEmitNativeNumbers -v
--- PASS
```

Retightened `internal/connectors/defs/github/schemas/{issues,pull_requests,commits,
issue_comments,workflow_artifacts}.json`: `user_id`/`author_id`/`committer_id`/`workflow_run_id`
type `["integer","string"]` -> `["integer","null"]`, description updated to explain the now-native
type. Removed `isStringifiedNestedID` entirely from `paritytest/github/parity_test.go` (not just
bypassed) and its special-cased branch in `TestParityGithub_StreamRecords`'s comparison loop — these
4 fields now flow through the same plain `reflect.DeepEqual` RAW-equality path as every other field.

```
$ go run ./cmd/connectorgen validate internal/connectors/defs
connectorgen validate: 13 connector(s) checked, 0 findings

$ go test ./internal/connectors/paritytest/github/... -v
--- PASS (all subtests, including TestParityGithub_StreamRecords with the helper removed)
```

`docs.md` Known limits gained a "G0b — RESOLVED (stale ledger prose fixed)" bullet.
`p9-github-ledger.md`'s original G0b row struck through (`~~...~~`) with a **RESOLVED** annotation
appended (conventions.md §5's own strikethrough+RESOLVED pattern), preserving the original RED/GREEN
historical evidence rather than deleting it.

### GitHub self-verify (final)

```
$ go build ./...
(clean)

$ go run ./cmd/connectorgen validate internal/connectors/defs
connectorgen validate: 13 connector(s) checked, 0 findings

$ go test ./internal/connectors/paritytest/github/... ./internal/connectors/hooks/github/... -v
(all subtests PASS — 20 parity tests, 15 hook unit tests)

$ go test ./internal/connectors/conformance -run TestConformance/github -v
--- PASS: TestConformance/github

$ wc -l internal/connectors/hooks/github/hooks.go
400  (hard ceiling, not exceeded; 2 hook interfaces: AuthHook + WriteHook, Tier-2 cap unchanged)
```

---

## GMAIL

Legacy read for spot-checks: `internal/connectors/gmail/{gmail,streams,auth}.go` (read-only
reference, never edited). `internal/connectors/hooks/gmail/**` was NOT touched — none of the 3
mandated gmail fixes require a hook change (docs-only, spec-only, schema+test-only respectively).

### Item 1 — docs.md Known-limits stale hooks-blind-conformance claim fixed

`internal/connectors/defs/gmail/docs.md`'s Known-limits 3rd bullet still stated conformance
"always call[s] engine.Check/engine.Read with a nil Hooks argument (by design — conformance has no
per-connector hook wiring mechanism)" and that github's bearer candidate "masks" the identical gap.
This was TRUE at gmail's original migration time but R3 (wave1-pilot's earlier repair round, per
REVIEW-A.md's finding) made `conformance/dynamic.go` hook-aware (`engine.HooksFor` + blank-import of
`hooks/hookset`) and replaced this exact framing with the `metadata.json` skip-marker mechanism —
confirmed by re-reading `internal/connectors/conformance/dynamic.go` directly: it DOES pass
`engine.HooksFor(b.Name)` today, not a hard-coded `nil`. github now has FULL dynamic hook coverage
(zero skip markers — confirmed via `internal/connectors/defs/github/metadata.json`), the opposite
of what the stale bullet claimed.

No test change needed (this is a pure prose correction; conformance's actual hook-aware behavior is
already covered by `TestConformance/github`'s full-dynamic pass and `TestConformance/gmail`'s own
marker-skip pass, both pre-existing green). Rewrote the bullet to describe the marker mechanism
accurately: gmail's marker is genuine (sole auth candidate is `mode:custom`, conformance's synthetic
config can never carry an https `token_url`, and inventing a bearer fallback or weakening the https
guard are both forbidden) — the marker's own reason text was already accurate; only the surrounding
prose (which claimed a repo-wide hook-blindness that no longer exists) was stale.

```
$ go test ./internal/connectors/conformance -run TestConformance/gmail -v
--- PASS: TestConformance/gmail   (marker-skip path, unaffected by the docs fix — confirms the
    corrected prose matches actual runtime behavior)
```

### Item 2 — dead spec keys max_pages/mode dropped

REVIEW-A.md major: `spec.json`'s `max_pages` ("use 0, all, or unlimited to exhaust the stream") is
consumed by NOTHING (`PaginationSpec` is static bundle JSON, never templated/config-read — confirmed
via `grep -rn "max_pages" internal/connectors/engine/` finding zero references outside the const
default-materialization path, which itself never reads a `max_pages` key); `mode` ("fixture for
credential-free conformance") is a legacy-only affordance (`gmail.go`'s `fixtureMode`) with no
engine-side path that ever reads `cfg.Config["mode"]` for gmail's declarative bundle. Both are F6
violations (conventions.md: a declared-but-unwireable key is worse than an absent one) — unlike
`start_date`/`include_spam_and_trash`, which at least have an explicit Known-limits forward-compat
note, `max_pages`/`mode` had NO such disclosure and misleading operator-facing description text.

RED-equivalent check first (docs-only/spec-only change, so "RED" here is the validate/grep evidence
that these keys are genuinely dead, not a failing test — matches gaploop-s1-ledger.md item 7's
precedent for docs-only items):

```
$ grep -rn '"max_pages"\|"mode"' internal/connectors/engine/*.go internal/connectors/defs/gmail/streams.json
(zero matches in engine/*.go for either key; zero matches in gmail/streams.json — confirms both are
 declared in spec.json and consumed by NOTHING anywhere in the bundle or the engine)
```

GREEN: removed `max_pages` and `mode` properties entirely from `internal/connectors/defs/gmail/
spec.json`.

```
$ go run ./cmd/connectorgen validate internal/connectors/defs
connectorgen validate: 13 connector(s) checked, 0 findings

$ go test ./internal/connectors/paritytest/gmail/... -v
(unaffected — no test referenced either key; all PASS except the pre-existing, separately-scoped
 TestParityGmail_ComputedFieldsStringifyLabelCountFields, see item 3 below)
```

`docs.md`'s "Streams notes"/Known-limits no longer need a `max_pages`/`mode` mention (they were
never disclosed there to begin with — REVIEW-A.md's point was exactly that these dead keys had NO
Known-limits treatment while the two filter keys did; removing the keys resolves the F6 violation
directly rather than requiring a new disclosure for something no longer declared).

### Item 3 — flip TestParityGmail_ComputedFieldsStringifyLabelCountFields to native types

Baseline RED (pre-existing at task start, exactly as Step 1 predicted and scoped to Step 2):

```
$ go test ./internal/connectors/paritytest/gmail/... -run TestParityGmail_ComputedFieldsStringifyLabelCountFields -v
--- FAIL: TestParityGmail_ComputedFieldsStringifyLabelCountFields
    engine labels[0].messages_total = "10" (json.Number), want string (computed_fields always
    stringifies — engine/interpolate.go's resolveExpr/stringify)
```

`labels`' 4 computed_fields (`messages_total`/`messages_unread`/`threads_total`/`threads_unread`,
sourced via bare single `"{{ record.messagesTotal }}"`-shaped templates in
`internal/connectors/defs/gmail/streams.json`) all qualify for Step 1's typed extraction (bare
single reference, no filter, no literal text) — confirmed by inspecting `streams.json` directly.
The test itself still asserted the PRE-increment stringified form, so it was the test lagging
behind already-landed engine behavior, not an engine regression.

GREEN:
- `internal/connectors/paritytest/gmail/parity_test.go`:
  `TestParityGmail_ComputedFieldsStringifyLabelCountFields` renamed to
  `TestParityGmail_ComputedFieldsPreserveLabelCountFieldsNativeType` and rewritten to assert
  `json.Number` (native type) on the engine side instead of `string`, matching legacy's own
  `json.Number` (connsdk's `UseNumber`-decoded envelope) — RAW equality via `reflect.DeepEqual`
  between the two now that both sides emit the identical type, not just the identical stringified
  value.
- `internal/connectors/defs/gmail/schemas/labels.json`: retightened all 4 fields'
  `["string","null"]` -> `["integer","null"]` (their real Gmail API wire type), descriptions updated
  to point at the RESOLVED state instead of the stringify limitation.

```
$ go test ./internal/connectors/paritytest/gmail/... -run TestParityGmail_ComputedFieldsPreserveLabelCountFieldsNativeType -v
--- PASS

$ go run ./cmd/connectorgen validate internal/connectors/defs
connectorgen validate: 13 connector(s) checked, 0 findings

$ go test ./internal/connectors/conformance -run TestConformance/gmail -v
--- PASS   (records_match_schema now validates the tightened integer type against the
    now-native-typed emitted records cleanly)
```

`docs.md`'s "Streams notes" stringify-deviation paragraph rewritten from "documented parity
deviation" framing to "RESOLVED" framing (typed extraction landed; schemas tightened; test renamed
and flipped to native-type equality) — matches the chargebee/github treatment of the identical A1
finding.

### Gmail self-verify (final)

```
$ go build ./...
(clean)

$ go run ./cmd/connectorgen validate internal/connectors/defs
connectorgen validate: 13 connector(s) checked, 0 findings

$ go test ./internal/connectors/paritytest/gmail/... -v
(all subtests PASS, zero failures — the one pre-existing failure this task started with is fixed)

$ go test ./internal/connectors/conformance -run TestConformance/gmail -v
--- PASS: TestConformance/gmail

$ git diff --stat -- internal/connectors/hooks/gmail
(empty — hooks/gmail untouched, per dispatch's read-only default)
```

---

## Combined final self-verify (whole Step 2 github+gmail scope)

```
$ go build ./...
(clean)

$ go run ./cmd/connectorgen validate internal/connectors/defs
connectorgen validate: 13 connector(s) checked, 0 findings

$ go test ./internal/connectors/conformance -run 'TestConformance/(github|gmail)' -v
--- PASS: TestConformance/github
--- PASS: TestConformance/gmail

$ go test ./internal/connectors/paritytest/github/... ./internal/connectors/paritytest/gmail/... ./internal/connectors/hooks/github/... -v
(all PASS, zero failures)

$ go test ./internal/connectors/... 2>&1 | grep -v '^ok'
(only "no test files" notices for packages that never had tests — zero FAIL lines; the gmail
 failure this task started with is gone, no new failures introduced anywhere else in the fleet)

$ make lint
(golangci-lint run over engine/defs/hooks/native/conformance/certify/connectorgen/inventorygen)
0 issues.
```

No dependency additions, no schema migrations (JSON-Schema bundle "schemas" only — no DB
migrations), no auth/security WEAKENING (the auth_type/public_access fix is strictly a
STRICTER-than-legacy tightening — closes a silent fail-open path, never opens one), no destructive
data actions, no secret access beyond what the existing AuthHook/paritytest fixtures already
exercised, no quality-gate reductions. No commits made (per dispatch instructions).

# Overview

GitHub started as the wave1-pilot Tier-2 (AuthHook + WriteHook) migration of
`internal/connectors/github` (the largest, highest-risk legacy connector: 1980+352+295+85 = ~2712
loc across `github.go`/`streams.go`/`auth.go`/`manifest.go`, 19 read streams, 25 write actions — the
only pilot with real writes), then underwent a **Pass B full-surface expansion** (this pass): every
repository-scoped endpoint in GitHub's published OpenAPI description (500 operations under
`/repos/{owner}/{repo}/...`, `docs/github/rest-api-description`) was reviewed and mapped to either a
stream, a write action, or a documented exclusion in `api_surface.json` — no blanket
`out_of_scope` bucket. The bundle now reads GitHub repository, issue, pull request, code, release,
collaboration, Actions, webhook, deploy-key, environment, fork, invitation, issue-event, and security
(code scanning / Dependabot / secret scanning / security-advisory) data, plus repository rulesets and
autolinks, and writes 67 approved reverse-ETL actions through the GitHub REST API. This bundle is
engine-vs-legacy parity-tested against `internal/connectors/github` (read-only reference, frozen at
its own 19-stream/25-write surface); the legacy package stays registered and unchanged until wave6's
registry flip — parity tests now assert the bundle is a **superset** of legacy's surface, not an
exact match, since Pass B intentionally adds streams/writes legacy never had.

Declarative bundle: `metadata.json`, `spec.json`, `streams.json` (33 streams), `writes.json` (67
actions), `schemas/*.json`, `fixtures/**`, this file. Go escape hatch:
`internal/connectors/hooks/github/hooks.go` implements exactly 2 hook interfaces (the Tier-2 cap):
`AuthHook` (github_app JWT -> installation-token exchange) and `WriteHook` (compound multi-request
write actions a single declarative action cannot express). None of the 42 write actions Pass B added
need a hook — every one of them is a single-request declarative action — so `hooks.go` is unchanged
at exactly 400 lines (see the STANDING EXCEPTION note in Known limits, also unchanged).

Note: PLAN.md/SPEC.md's per-connector row cites "16 actions"; the legacy source
(`github.go:1759` `githubWriteActionSpecs`) actually enumerates 25 distinct write actions. The
wave1-pilot bundle implemented the full 25 for capability parity (the task mandate at the time), not
the 16 the planning doc cited — flagged for the P-12 conventions/planning-doc correction pass. Pass B
adds 42 more on top of that 25 (67 total; see "Write actions & risks" below).

## Pass B full-surface expansion (this pass)

`api_surface.json` was rewritten from a 63-entry "wave1-pilot scope, everything else `out_of_scope`"
manifest into a full enumeration of the connector's natural surface: every one of the 500
`/repos/{owner}/{repo}/...` operations in GitHub REST API v3 1.1.4's published OpenAPI description.
Org-level, user-level, enterprise-level, and gist/GraphQL surfaces remain out of scope — this
connector's `spec.json` configures a single `owner`+`repo` identity, not an org or user identity, so
org/user-scoped resources (teams, org secrets, gists, enterprise admin, etc.) have no config surface
to hang off of without a distinct, separate scope-widening change.

**New streams (14)**: `commit_comments`, `deploy_keys`, `webhooks`, `environments`, `forks`,
`invitations`, `issue_events`, `code_scanning_alerts`, `dependabot_alerts`,
`secret_scanning_alerts`, `security_advisories`, `repo_rulesets`, `autolinks`, `languages`.

**New write actions (42)**: `create_webhook`/`update_webhook`/`delete_webhook`,
`create_deploy_key`/`delete_deploy_key`, `create_or_update_environment`/`delete_environment`,
`create_commit_comment`/`update_commit_comment`/`delete_commit_comment`,
`update_issue_comment`/`delete_issue_comment`, `lock_issue`/`unlock_issue`,
`set_issue_labels`/`add_issue_labels`/`remove_issue_label`,
`add_issue_assignees`/`remove_issue_assignees`,
`create_review_comment`/`update_review_comment`/`delete_review_comment`,
`submit_pull_request_review`/`dismiss_pull_request_review`, `update_pull_request_branch`,
`update_release_asset`/`delete_release_asset`, `replace_repo_topics`,
`add_collaborator`/`remove_collaborator`, `create_ref`/`update_ref`/`delete_ref`, `merge_branch`,
`update_code_scanning_alert`/`update_dependabot_alert`/`update_secret_scanning_alert`,
`create_deployment`, `create_fork`,
`create_repo_ruleset`/`update_repo_ruleset`/`delete_repo_ruleset`.

**Exclusion category breakdown** (all 402 excluded endpoints carry a real category + reason, closed
vocabulary, no blanket bucket): `requires_elevated_scope` (168 — Actions/Dependabot/environment
secrets and variables, branch protection, org-scoped resources surfaced under the repo path,
security-feature toggles, CodeQL/code-quality/traffic analytics requiring elevated scopes),
`out_of_scope` (143 — narrow preview features like sub-issues/issue-field-values/Codespaces,
caller-chosen-path/ref/SHA lookups with no bulk enumeration, CI-provider-credential-conventional
endpoints like commit statuses and check runs), `duplicate_of` (67 — single-resource detail GETs
whose data is already covered by the corresponding list stream, or an action that's a narrower
variant of one already implemented), `binary_payload` (10 — zip/tarball/SBOM/SARIF/raw-asset-upload
endpoints), `non_data_endpoint` (8 — boolean checks, pings, generators with no persisted record),
`destructive_admin` (5 — repo/branch/deployment deletion, ownership transfer), `deprecated` (1 — the
legacy repository events timeline). See `api_surface.json` for the full per-endpoint mapping.

## Auth setup

Three credential shapes, reproduced via `streams.json` `base.auth`'s ordered candidates. Legacy's
`auto` resolution order (`auth.go:73-80`: token wins, then github_app, then **silently** public) is
matched for the token/github_app precedence, but the final fallback is now a documented, deliberate
**stricter-than-legacy** deviation (see the "auth config surface" paragraph below and ledger G14):

1. **Token** (`{"mode":"bearer","token":"{{ secrets.token }}","when":"{{ secrets.token }}"}`) — a
   classic PAT, fine-grained PAT, OAuth token, GitHub Actions `GITHUB_TOKEN`, or a pre-generated
   installation access token, supplied via the `token` secret. Sent as `Authorization: Bearer
   <token>`. This candidate's `when` truthiness check means it is skipped entirely (falls through
   to the next candidate) when `token` is unset — the engine's absent-key-falsy semantics for `when`
   (never for ordinary interpolation) make this safe.
2. **GitHub App** (`{"mode":"custom","hook":"github","when":"{{ config.app_id }}"}`) — server-to-
   server auth. `hooks/github/hooks.go`'s `Authenticator` signs an RS256 JWT (stdlib
   `crypto/rsa`, exactly like legacy `auth.go:154+` `githubAppJWT`) from `app_id`
   (config) + the private key (`private_key`/`private_key_base64` secrets, base64-decoded when
   using the latter), then POSTs `/app/installations/{installation_id}/access_tokens` on the
   connector's own base URL to exchange it for a one-hour installation access token, and returns a
   `connsdk.Authenticator` that sets `Authorization: Bearer <installation_token>`. Matches legacy
   exactly, including the **uncached** re-mint-on-every-call behavior (Known limits).
3. **Public** — unauthenticated reads only, reachable via EITHER of two candidates (see the config
   surface paragraph below): `{"mode":"none","when":"{{ config.public_access }}"}` (the primary,
   dedicated boolean opt-in), OR `{"mode":"none","when":"{{ config.auth_type in ['public', 'none',
   'anonymous', 'unauthenticated'] }}"}` (S3 engine mini-wave item 2: legacy's exact string-enum
   selection, restored as an additional opt-in once `engine.ResolveCheckWhen` made the `in` operator
   statically validatable — see ledger G14 R1 CONDITION). Writes fail per-request with a GitHub
   401/403 (no separate "requires write auth" precheck is reproduced — see Known limits).

`app_id`+`installation_id` must both be configured for the github_app path; `installation_id`
absence is caught by the hook itself (not the `when` gate, which only tests `app_id`'s truthiness —
see the parity-deviation ledger item G5) with the same error class legacy raises.

**Auth config surface vs legacy — secret ALIASES and `auth_type`'s non-public modes are NOT
reproduced; `public_access` (primary) and `auth_type`'s 4 public synonyms (additional, S3 restored)
close a silent-fallthrough hazard (ledger G14).** Legacy honors an explicit `auth_type`/`auth`/
`authentication` config value (`auth.go:61-96`, case/hyphen-insensitive, many synonyms per mode:
`auth_type=github_app` forces app auth even when a token secret is ALSO set,
`auth_type=public`/`none`/`anonymous`/`unauthenticated` forces anonymous reads) plus a wide set of
secret aliases for the token (`personalAccessToken`/`accessToken`/`oauthToken`/`installationToken`/
`githubToken`/`GITHUB_TOKEN`, `github.go:1634-1644`), the private key
(`privateKey`/`githubAppPrivateKey`/`privateKeyBase64`/`githubAppPrivateKeyBase64`), and the app ID
(`client_id`/`github_app_id`, `auth.go:256-257`). None of the secret ALIASES, and none of `auth_type`'s
non-public-synonym behavior (forcing github_app over a simultaneously-set token; the
token/oauth/actions/installation mode distinctions, which this bundle collapses into the single
`bearer` candidate since it never needs to distinguish them), are reproduced — this bundle reads
only the canonical `token`/`private_key`/`private_key_base64`/`app_id` keys for credential material.
The dangerous failure mode this created (REVIEW-A.md major finding) was that a caller who supplied
ONLY an alias-shaped secret (e.g. `personalAccessToken` but not `token`) got a **silent,
unauthenticated** read with zero error — the bundle's `base.auth` chain fell through the token
candidate (unset `token`) and the github_app candidate (no `app_id`) straight to an unconditional
`mode:none`. This is now fixed: the `none` outcome requires an EXPLICIT opt-in — either the
dedicated `public_access` config key (any non-empty value; the primary, documented surface) OR
`auth_type` set to one of the 4 legacy public synonyms (`public`/`none`/`anonymous`/`unauthenticated`
— S3 engine mini-wave item 2, restored once `engine.ResolveCheckWhen` made the `in` operator
statically validatable; see ledger G14 R1 CONDITION) — so a config that resolves to none of
token/github_app/either-public-opt-in now hard-errors ("select auth: no auth spec matched") instead
of silently reading unauthenticated (F4, THREAT-MODEL: never fail open). This is intentionally
**stricter than legacy** (legacy's own `auto` mode silently falls through to public in this exact
shape) — a deliberate, documented deviation, not a parity-preserving one, closing a real
security-relevant gap rather than reproducing it. The two opt-ins are two SEPARATE `mode:none`
candidates in the `auth` list (not one OR'd condition) since `EvalWhen`'s grammar has no `||`
operator; both gate on an explicit signal, so this is purely additive — no new way to reach
`mode:none` implicitly exists. `auth_type`'s restoration is intentionally NARROW: it reproduces only
the 4 public-synonym STRING VALUES, not legacy's full mode-selection semantics (e.g.
`auth_type=github_app` forcing app auth ahead of a configured token) — any other `auth_type` value
is inert (auth selection still resolves via the token/app_id-presence candidates ahead of it in the
list). Any caller who previously depended on an alias secret name, `auth_type=github_app` forcing
app auth over a simultaneously-set token, or another non-public `auth_type` synonym must migrate to
the canonical `token`/`app_id`+`installation_id`+`private_key` keys (see ledger G14; parity tests:
`TestParityGithub_AuthNoCredentialsFailsLoudRatherThanSilentlyPublic`,
`TestParityGithub_AuthExplicitPublicOptIn`, `TestParityGithub_AuthTypePublicEnumOptIn`,
`TestParityGithub_AuthTypeUnrelatedValueDoesNotGrantPublicAccess`).

## Streams notes

All 19 legacy streams (`repository`, `issues`, `pull_requests`, `branches`, `commits`, `tags`,
`releases`, `labels`, `milestones`, `issue_comments`, `pull_request_review_comments`,
`collaborators`, `contributors`, `stargazers`, `subscribers`, `workflows`, `workflow_runs`,
`workflow_artifacts`, `deployments`) are implemented. Pagination is uniformly `page_number`
(`page`/`per_page`, `page_size: 100`, short-page stop — matches legacy's `readPages`/
`readEnvelopePages` short-page-stop-when-`len(page) < perPage` behavior exactly). `owner`/`repo` are
two separate config keys (see Known limits) interpolated into every stream's `/repos/{{ config.owner
}}/{{ config.repo }}/...` path.

**`per_page`/`max_pages` are NOT runtime-configurable** (documented deviation from legacy's default
of `max_pages=1`/`per_page=100`, both config-overridable): `PaginationSpec` fields are static bundle
JSON, never templated (conventions.md's dialect reference: no field references `PaginationSpec` in
`connectorgen validate`, and no `{{ }}` resolution exists for pagination knobs at all), so this
bundle's `page_size: 100` and the ABSENCE of a declared `max_pages` (which defaults to 0/unbounded
per `engine/read.go`) are fixed values, not per-sync-configurable. This means the bundle's default
behavior is **unbounded** pagination (reads every page until a short/empty page), while legacy's own
default (`githubDefaultMaxPages = 1`) reads only ONE page unless a caller explicitly sets
`config.max_pages`. Parity is asserted with legacy configured for the SAME effective behavior
(`max_pages=all`), not against legacy's capped default — see
`TestParityGithub_IssuesPaginationFiltersOutPullRequests`.

- `repository` is `single_object: true` (a single JSON object response, not a list).
- `issues` filters out pull requests via `records.filter.field_absent: pull_request` (declarative
  equivalent of legacy's `if _, ok := item["pull_request"]; ok { return nil, false }`), and is the
  only stream with SERVER-SIDE incremental filtering (`since` query param). `issue_comments`/
  `pull_request_review_comments` are also server-side `since`-filtered.
  **This matches legacy exactly ONLY on the config path** (`config.since` -> `since` query param,
  `TestParityGithub_SinceConfigOnlyMatchesLegacy`): legacy's `since` filter reads
  `req.Config.Config["since"]` only and never consults `req.State` anywhere in the package. The
  engine-wide incremental mechanism (`engine/read.go`'s `incrementalLowerBoundValue`) prefers an
  app-persisted STATE cursor over `start_config_key` when one is present, so a sync with persisted
  state forwards the state cursor as `since` on the engine side while legacy would ignore it and
  re-request the full (or config-`since`-bounded) set — a smaller, correctly-incremental record set,
  and a deliberate, documented IMPROVEMENT consistent with every other incremental stream in the
  fleet, not a parity bug (`TestParityGithub_SinceStateCursorForwardingIsEngineOnlyBehavior` pins
  both halves: config-path equality AND the state-path divergence).
- `pull_requests`/`releases`/`milestones` declare an `incremental.cursor_field` for manifest/
  sync-mode surface parity (legacy's own `Stream{CursorFields: [...]}` declares these too) but
  neither legacy nor this bundle actually filters by it server- or client-side — both are
  always-full-stream reads, matching legacy's real (non-)behavior exactly (no `request_param`, no
  `client_filtered`).
- `workflows`/`workflow_runs`/`workflow_artifacts` are envelope responses
  (`{"workflows":[...],"total_count":N}` etc.) — `records.path` names the envelope array key
  (`workflows`/`workflow_runs`/`artifacts`), matching legacy's `readEnvelopePages`.
- Heavy `computed_fields` flattening reproduces legacy's `nestedString`/`nestedValue` field
  flattening (`user_login`, `author_login`, `commit_author_name`, `base_ref`, `head_sha`,
  `workflow_run_id`, etc.). Legacy's `repository` marker field (every emitted record carries the
  `owner/repo` string) IS reproduced on all 19 streams via
  `"repository": "{{ config.owner }}/{{ config.repo }}"` (gap-loop cycle-1 engine mini-wave closed
  `ENGINE_GAP` G0 — `computed_fields` can now reference `config.*`, never `secrets.*` — see Known
  limits / ledger G0, RESOLVED).
- `checkOrigin`/link-header pagination is NOT used: legacy's own `readPages` is `page`/`per_page`
  query-param pagination, not RFC 5988 Link-header following, so `page_number` (not `link_header`)
  is the byte-accurate parity choice despite GitHub's REST API also supporting Link headers.

### Pass B streams (14 new, this pass)

All 14 use the same `page_number` pagination as the legacy-parity streams (`page`/`per_page: 100`,
short-page stop) except `languages`, which declares `pagination: {type: none}` (GitHub's
`/languages` endpoint returns exactly one object, never paginated).

- `commit_comments`, `deploy_keys`, `webhooks`, `forks`, `invitations`, `issue_events`,
  `repo_rulesets`, `autolinks`: plain array responses, `records.path: "."`, same shape as the
  legacy-parity streams.
- `environments` is an envelope response (`{"total_count": N, "environments": [...]}`) —
  `records.path: "environments"` names the envelope array key, same convention as
  `workflows`/`workflow_runs`/`workflow_artifacts`.
- `code_scanning_alerts`, `dependabot_alerts`, `secret_scanning_alerts`, `security_advisories`
  declare `incremental.cursor_field: updated_at` (client-side surface parity only — no
  `request_param`, matching the `pull_requests`/`releases`/`milestones` precedent of declaring the
  cursor field for manifest/sync-mode purposes without a server-side filter, since none of these
  four alert/advisory list endpoints document a since-style incremental query parameter).
  `secret_scanning_alerts`' schema deliberately does NOT project the API's own `secret` field (the
  actual leaked-credential value GitHub echoes back) — schema-mode projection already drops any
  undeclared field by default, and this omission is intentional, not an oversight: a record is what
  flows to a destination warehouse, and a leaked secret has no business landing there.
- `languages` is a single-object-per-repository stream whose body is a raw `{"Go": 123, ...}` map
  keyed by an arbitrary, dynamic language name (not a fixed property set) — it declares
  `projection: "passthrough"` (conventions.md §3) so every language key survives verbatim; only the
  `repository` marker field is statically declared in `schemas/languages.json`.
- `repo_rulesets`'s stream covers ruleset LIST/detail data (name/target/enforcement/source); the
  full `rules[]`/`conditions`/`bypass_actors` nested structures returned by GitHub are passed through
  raw only via `create_repo_ruleset`/`update_repo_ruleset`'s write-side `record_schema` (loosely
  typed as `object`/`array` there, not decomposed field-by-field) — the read-side schema stays
  intentionally narrower (the summary fields most consumers need), matching the same "read schema
  narrower than write schema" shape `repository`'s own PATCH-equivalent settings surface has.

## Write actions & risks

All 25 legacy write actions (`github.go:1759+` `githubWriteActionSpecs`) are implemented — 21
purely declarative (`writes.json` actions with `path_fields`/`body_fields`/JSON-Schema `record_schema`
validation), 4 requiring `hooks/github/hooks.go`'s `WriteHook.ExecuteWrite` because they are
genuinely compound (multiple HTTP requests per logical write, matching legacy's own follow-up
request helpers): `close_issue` (state PATCH + optional comment POST via legacy's
`writeIssueComment`), `create_pull_request` (create POST + optional issue-metadata PATCH + optional
reviewers POST via legacy's `writePullRequestFollowups`/`writeReviewers`), `update_pull_request`
(optional core PATCH + optional issue-metadata PATCH + optional reviewers POST), `close_pull_request`
(optional comment POST + state PATCH). `request_reviewers` and `create_pull_request_review`, despite
sounding related, are each a SINGLE request in legacy and are fully declarative here.

Declarative actions: `create_issue`, `update_issue`, `comment_issue`, `request_reviewers`,
`merge_pull_request`, `create_label`, `update_label`, `delete_label`, `create_milestone`,
`update_milestone`, `delete_milestone`, `create_release`, `update_release`, `delete_release`,
`dispatch_workflow`, `rerun_workflow_run`, `cancel_workflow_run`, `delete_workflow_run`,
`create_pull_request_review`, `create_or_update_file`, `delete_file` (DELETE with a JSON body via
`body_fields`, matching legacy's contents-API delete semantics). None of the 4 `delete_*` actions
declare `delete.missing_ok_status` — legacy's own `doJSONWithAuth` treats ANY non-2xx status
(including a 404 for an already-deleted resource) as a hard failure with no idempotent-delete
special-casing at all, so this bundle intentionally does NOT add the engine's `missing_ok_status`
leniency (conventions.md §3's delete semantics are available but declaring them here would be new,
more-lenient behavior legacy never had, changing write-accounting for a 404 input from "failed" to
"written" — not a parity-preserving deviation, see the ledger).

Every write action carries the exact legacy `Risk` prose (github.go's `githubWriteActionSpecs`).
`merge_pull_request` is the highest-risk action (irreversibly changes repository history unless
branch protection blocks the merge); `dispatch_workflow`/`rerun_workflow_run` start or repeat CI/CD
automation.

### Pass B write actions (42 new, this pass)

All 42 are purely declarative — none needed a `WriteHook` (every one of them is genuinely a single
HTTP request): `create_webhook`/`update_webhook`/`delete_webhook`,
`create_deploy_key`/`delete_deploy_key`, `create_or_update_environment`/`delete_environment`,
`create_commit_comment`/`update_commit_comment`/`delete_commit_comment`,
`update_issue_comment`/`delete_issue_comment`, `lock_issue`/`unlock_issue`,
`set_issue_labels`/`add_issue_labels`/`remove_issue_label`,
`add_issue_assignees`/`remove_issue_assignees`,
`create_review_comment`/`update_review_comment`/`delete_review_comment`,
`submit_pull_request_review`/`dismiss_pull_request_review`, `update_pull_request_branch`,
`update_release_asset`/`delete_release_asset`, `replace_repo_topics`,
`add_collaborator`/`remove_collaborator`, `create_ref`/`update_ref`/`delete_ref`, `merge_branch`,
`update_code_scanning_alert`/`update_dependabot_alert`/`update_secret_scanning_alert`,
`create_deployment`, `create_fork`,
`create_repo_ruleset`/`update_repo_ruleset`/`delete_repo_ruleset`.

- `delete_webhook`/`delete_deploy_key`/`delete_environment`/`delete_commit_comment`/
  `delete_issue_comment`/`remove_issue_label`/`delete_review_comment`/`delete_release_asset`/
  `remove_collaborator`/`delete_repo_ruleset` all declare `delete.missing_ok_status: [404]` (an
  already-deleted resource counts as written, not failed) — unlike the 4 wave1-pilot `delete_*`
  actions above, these are NEW actions with no legacy behavior to stay byte-parity with, so the
  engine's idempotent-delete leniency (conventions.md §3) is the correct default here, not a
  deviation from anything. `delete_ref` additionally declares `missing_ok_status: [404, 422]` —
  GitHub's git/refs delete returns 422 (not 404) for a ref that doesn't exist under certain
  ref-name shapes.
- `update_ref`/`delete_ref`'s `ref` path field genuinely contains a literal `/` (GitHub's own
  convention: `heads/my-branch`, `tags/v1.0.0`) — `InterpolatePath`'s default per-segment urlencoding
  turns this into a percent-encoded `%2F` on the wire, which both Go's own `net/http` server (used by
  this bundle's fixture replay/parity harness) and GitHub's real API decode back into a literal `/`
  for path routing — this is standard percent-decoding behavior, not a workaround, so no deviation is
  recorded for it.
- `create_ref` intentionally does NOT reproduce a convenience default for `sha` (create-a-branch-off-
  HEAD); the caller supplies the full 40-character commit SHA explicitly, matching GitHub's own API
  contract (`sha` is a required field on the create-ref request body).
- `merge_branch` is GitHub's plain two-ref merge-commit endpoint (POST `/merges`) — distinct from
  `merge_pull_request` (which merges an existing pull request via its `pull_number`) and from
  `update_pull_request_branch` (which merges the base INTO a PR's head to resolve a stale/behind PR,
  POST `/pulls/{pull_number}/update-branch`); all three create a merge commit but target different
  ref pairs and are not interchangeable.
- `create_repo_ruleset`/`update_repo_ruleset`'s `rules`/`conditions`/`bypass_actors` fields are typed
  loosely (`"type": "array"`/`"type": "object"` with no nested property decomposition) in
  `record_schema` — GitHub's ruleset rule shapes are a large, evolving `oneOf` union (branch naming
  patterns, required status checks, required signatures, merge-queue settings, etc.); validating the
  full union would require `anyOf`/`oneOf` support the engine's draft-07 subset does not have (see
  the parity-deviation ledger's stripe item 1 precedent for the same subset limitation) — a caller
  supplies a well-formed rules array per GitHub's own docs, and GitHub's API is the final validator.
- `add_issue_assignees`/`remove_issue_assignees` both hit the identical `/assignees` endpoint (POST
  to add, DELETE to remove — DELETE-with-body, matching `delete_file`'s established DELETE-with-body
  precedent above) — this is GitHub's own REST convention (an unusual but real, documented shape),
  not a bundle-specific oddity.

## Known limits

- **`auth`/`authentication` aliases, every legacy secret ALIAS, and `auth_type`'s non-public modes
  are NOT reproduced** (`personalAccessToken`/`accessToken`/`oauthToken`/`installationToken`/
  `githubToken`/`GITHUB_TOKEN`; `privateKey`/`githubAppPrivateKey`/`privateKeyBase64`/
  `githubAppPrivateKeyBase64`; `client_id`/`github_app_id`; `auth_type=github_app`'s
  override-token-precedence behavior) — only the canonical `token`/`private_key`/
  `private_key_base64`/`app_id` config/secret keys are read for credential material. **RESOLVED (S3
  engine mini-wave item 2) — `auth_type`'s 4 public synonyms ARE now reproduced**: `auth_type`
  set to `public`/`none`/`anonymous`/`unauthenticated` is an additional, purely-additive opt-in for
  unauthenticated reads, restored once `engine.ResolveCheckWhen` made the when-grammar's `in`
  operator statically validatable (previously `connectorgen validate`'s `engine.ResolveCheck` only
  parsed bare `namespace.key` truthiness, hard-failing any `==`/`in`-shaped `when` clause even when
  the referenced key was spec-declared). See "Auth setup"'s config-surface paragraph above for the
  full alias list and the fix for the silent-fallthrough hazard this previously caused (ledger G14):
  a config resolving to none of token/github_app/either-public-opt-in now hard-errors instead of
  silently reading unauthenticated.
- **`ENGINE_GAP` G0 — RESOLVED.** Legacy's `repository` marker field (every stream stamps the
  `owner/repo` string onto every emitted record) was NOT reproducible at pilot time because
  `streams.json`'s `computed_fields` templates were resolved via `Vars{Record: raw}` only
  (`engine/read.go`'s `applyComputedFields`) — `config.*` was never wired into that interpolation
  environment, unlike every OTHER templating surface in the dialect (base URL, headers, query, path,
  auth all receive both `Config` and `Record`/`Secrets`). The gap-loop cycle-1 engine mini-wave wired
  `Config` (never `Secrets` — a computed field must never be able to copy a secret into record data)
  into `applyComputedFields`'s `Vars`, so every stream now declares
  `"repository": "{{ config.owner }}/{{ config.repo }}"` in its `computed_fields`, restoring the
  marker on all 19 streams (see `TestParityGithub_RepositoryMarkerFieldRestored`,
  `docs/migration/conventions.md` §3 "`config.*` in `computed_fields`"). `owner`/`repo` also remain
  available on `RuntimeConfig.Config` directly for any caller that needs them.
- **G0b — RESOLVED (stale ledger prose fixed).** `p9-github-ledger.md`'s G0b entry (and this file's
  prior text) described `user_id`/`author_id`/`committer_id`/`workflow_run_id` (all sourced via a
  bare, single `computed_fields` template like `"{{ record.user.id }}"`, no filter/literal text) as
  emitted STRINGS, with `issues.json`/`pull_requests.json`/`commits.json`/`issue_comments.json`/
  `workflow_artifacts.json` widening these 4 fields' schema type to `["integer","string"]` and the
  parity suite comparing them string-form-only (`isStringifiedNestedID`). The gap-loop cycle-1 typed
  `computed_fields` extraction increment (same engine change that closed G0 above — REVIEW-A.md
  adjudication A1) now preserves the native JSON type for exactly this bare-single-reference
  template shape, so these 4 fields emit real `json.Number` values, matching legacy's own
  raw-JSON-passthrough numeric type exactly. Schemas are retightened to `["integer","null"]` (no
  more widened union) and the parity suite compares them via plain RAW equality (see
  `TestParityGithub_NestedIDComputedFieldsEmitNativeNumbers`; the old string-form-only
  `isStringifiedNestedID` helper is removed, not just bypassed).
- **`owner`/`repo` are two config keys, not legacy's single `repository` ("owner/repo") field.**
  The engine's `InterpolatePath` urlencodes every `{{ }}`-resolved value as one opaque path segment
  (a literal `/` inside a resolved value becomes `%2F`, not a segment delimiter), so a single
  `repository` config value cannot be split into two path segments declaratively (the dialect has no
  string-split filter). `owner` and `repo` are declared as separate required `spec.json` properties
  instead — an honest config-surface change from legacy, not a silent behavior narrowing (SPEC.md
  §5.6 anticipates this exact shape).
- **`labels_count`/`assignees_count`/`assets_count` are NOT reproduced.** Legacy derives these via
  `len(item["labels"])`/`len(item["assignees"])`/`len(item["assets"])`. The dialect's only
  array-aware `computed_fields` filter is `join:<sep>` (string-join, not count); there is no
  length/count filter. These three fields are omitted from `issues`/`releases` schemas entirely
  rather than approximated with a wrong value.
- **`is_pull_request` is NOT reproduced on the `issues` stream.** It is always legacy's literal
  `false` (issues stream already filters out PRs), but `computed_fields`' `Interpolate` always
  produces a STRING (`"false"`), never JSON-Schema `boolean` `false` — stamping it would introduce a
  byte-level record-shape mismatch with legacy rather than removing one. Omitted entirely.
  `pull_requests` stream's `draft`/other native booleans are unaffected (they pass through raw JSON
  values via schema projection, not `computed_fields`, so they keep their real boolean type).
- **Optional per-request passthrough filters are not wired**: legacy's `sort`/`direction` (issues/
  PRs/milestones), and the full `sha`/`path`/`author`/`committer`/`until` (commits) and
  `actor`/`branch`/`event`/`status`/`created`/`head_sha`/`check_suite_id` (workflow_runs) config
  filters are conditionally-sent-if-non-empty in legacy. The dialect's `stream.Query` templating has
  no absent-key-falsy tolerance (only `auth`'s `when` does), so an unconditional `{{ config.x }}`
  reference hard-errors whenever the caller leaves that filter unset — the common case. These
  filters are not declared in `spec.json` at all (F6, conventions.md: a declared-but-unwireable key
  is worse than an absent one). `state` (issues/pull_requests/milestones) IS always sent, but as the
  static literal `"all"` (legacy's own default when unconfigured) rather than a runtime-overridable
  config value, for the identical reason.
- **github_app installation-token exchange is uncached** (matches legacy exactly, not a new
  limitation introduced here): `hooks/github/hooks.go`'s `Authenticator` mints a fresh JWT and POSTs
  a fresh installation-token exchange on every call, exactly like legacy's `authorizationHeader` ->
  `githubAppInstallationToken` (auth.go:117-152) does on every single HTTP request during a sync.
  This is real, legacy-inherited inefficiency (documented, not silently "fixed" by adding caching
  this migration never asked for — conventions.md's rate-limit-placement precedent).
  `installation_repositories`/`installation_repository_ids`/`installation_permissions` (restricted
  installation-token scoping) are read from `config.*` inside the hook, matching legacy's
  `githubInstallationTokenPayload`.
  A `public`-mode write attempt is NOT pre-validated the way legacy's `githubHasWriteAuth` check
  short-circuits it before ever building a request; this bundle relies on GitHub's own 401/403
  response for an unauthenticated write instead (still fails, just one HTTP round-trip later — never
  silently succeeds).
- **Legacy's write-action name ALIASES are not reproduced** (`issue_create`, `pr_merge`, etc. all
  normalizing to a canonical action name via `githubNormalizeWriteAction`). This bundle's
  `writes.json` declares only the 25 canonical action names; a caller must supply the canonical name
  (documented scope narrowing, parity-deviation ledger item G1).
- **OR-rule / "at least one mutable field" validations are approximated, not exact**, for
  `create_pull_request` (title+body XOR issue), `update_issue`/`update_pull_request`/
  `update_milestone`/`update_release`/`update_label` ("at least one field present"), matching
  stripe's existing documented deviation #1 precedent (draft-07 subset has no `anyOf`/`oneOf`) —
  strictly more permissive than legacy, never stricter, never diverges for a legacy-valid record.
- **`create_or_update_file`'s dual `content`/`content_base64` convenience fallback is not
  reproduced.** Legacy accepts either a pre-base64-encoded `content_base64` OR raw `content` (which
  it then base64-encodes itself before sending). The engine has no filter that base64-encodes a body
  FIELD value (only `{{ }}`-templated string values support the `base64` filter, and body
  construction passes record fields through verbatim). This bundle's `content` field is the
  pre-encoded (GitHub API's actual wire shape) form only — a caller must supply already-base64
  content, matching GitHub's real contents API contract directly.
- **SUPERSEDED (Pass B full-surface expansion, this pass)**: this bullet previously read "Full
  GitHub REST surface (orgs, teams, projects v2, notifications, code scanning, dependabot, secrets
  administration, webhooks, GraphQL) is out of scope" — that blanket framing is no longer accurate.
  Webhooks, code scanning, and dependabot are now real streams/writes (see "Pass B streams"/"Pass B
  write actions" above). Org-level, user-level, enterprise-level, gist, and GraphQL surfaces remain
  genuinely out of scope (this connector's `spec.json` has no org/user identity to hang those
  resources off), each with its own specific category+reason in `api_surface.json`
  (`requires_elevated_scope` for org-scoped resources surfaced under the repo path,
  `out_of_scope` for narrow preview features and caller-chosen-parameter lookups with no bulk
  enumeration) — see `api_surface.json`'s `scope` field and the "Pass B full-surface expansion"
  section above for the full breakdown, not a single blanket bucket.
- **Parity-test methodology changed from exact-equality to superset (Pass B)**:
  `paritytest/github/parity_test.go`'s `TestParityGithub_BundleLoadsAndValidates` and
  `TestParityGithub_ManifestSurface` asserted `reflect.DeepEqual` against legacy's exact 19
  streams/25 writes at wave1-pilot time. Both now assert the bundle's stream/write name set is a
  SUPERSET containing every legacy-parity name (via a shared `assertSupersetOf` helper), since Pass B
  intentionally adds streams/writes legacy never had — an exact-equality assertion would have to be
  deleted or perpetually rewritten on every future capability addition, which superset-checking
  avoids while still catching an accidental regression (a legacy-parity stream/write silently
  dropped).
- **STANDING EXCEPTION — `hooks/github/hooks.go` stays at exactly 400 lines (the Tier-2 hard
  ceiling), zero headroom, evaluated and NOT reduced this pass** (S3 engine mini-wave item 3, carried
  minor: "github hooks.go sits exactly AT the 400-line hard ceiling — reduce ONLY if achievable
  without gaming"). Re-evaluated at S3: the wave1-pilot gap-loop repair round (REVIEW-A.md's
  github label major fix) already trimmed this file from 424 to exactly 400 lines by removing
  redundant comment prose and collapsing 3 near-identical `updateLabel` field-ifs into one loop — the
  cheap, safe reductions were already taken. Surveyed again here for any REMAINING safe reduction:
  the file's ~21 comment lines are all load-bearing (explaining non-obvious legacy-parity behavior —
  e.g. the uncached JWT re-mint, the OR-rule approximation, the label color-strip rationale — not
  restatements of the code); its ~29 blank lines are single standard `gofmt`-conventional separators
  between logical blocks (`gofmt -l` reports the file clean; removing any would violate normal Go
  formatting purely to hit a number). The one candidate logic consolidation considered
  (`createLabel`/`updateLabel` sharing a payload-builder) was rejected: the two functions have
  different validation rules (create requires name+color both; update requires only name, every
  other field optional) and merging them would trade a couple of lines for reduced clarity/increased
  coupling risk in Tier-2 escape-hatch code, which is a net-negative trade, not a genuine
  simplification. Conclusion: no further reduction is achievable without removing genuine
  documentation or forcing an artificial merge — this is a standing exception, not deferred work.
  Reviewer citation: REVIEW-A.md's "Re-review (gap loop cycle 1)" disposition table (github label
  major row: "hooks.go trimmed 424→400 (AT the hard ceiling — watch item)") and SUMMARY.md's carried
  minors list (wave1-pilot phase summary) name this exact line as the standing watch item this
  bullet resolves the evaluation of.

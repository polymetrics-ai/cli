# Overview

GitHub is the wave1-pilot Tier-2 (AuthHook + WriteHook) migration of `internal/connectors/github`
(the largest, highest-risk legacy connector: 1980+352+295+85 = ~2712 loc across
`github.go`/`streams.go`/`auth.go`/`manifest.go`, 19 read streams, 25 write actions — the only pilot
with real writes). It reads GitHub repository, issue, pull request, code, release, collaboration,
and Actions data, and writes approved reverse-ETL actions through the GitHub REST API. This bundle
is engine-vs-legacy parity-tested against `internal/connectors/github` (read-only reference); the
legacy package stays registered and unchanged until wave6's registry flip.

Declarative bundle: `metadata.json`, `spec.json`, `streams.json` (19 streams), `writes.json` (25
actions), `schemas/*.json`, `fixtures/**`, this file. Go escape hatch:
`internal/connectors/hooks/github/hooks.go` implements exactly 2 hook interfaces (the Tier-2 cap):
`AuthHook` (github_app JWT -> installation-token exchange) and `WriteHook` (compound multi-request
write actions a single declarative action cannot express).

Note: PLAN.md/SPEC.md's per-connector row cites "16 actions"; the legacy source
(`github.go:1759` `githubWriteActionSpecs`) actually enumerates 25 distinct write actions. This
bundle implements the full 25 for capability parity (the task mandate), not the 16 the planning doc
cites — flagged for the P-12 conventions/planning-doc correction pass.

## Auth setup

Three credential shapes, matching legacy's `auto` resolution order (`auth.go:73-80`: token wins,
then github_app, then public) exactly, reproduced via `streams.json` `base.auth`'s ordered
candidates:

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
3. **Public** (`{"mode":"none"}`) — unauthenticated reads only; writes fail per-request with a
   GitHub 401/403 (no separate "requires write auth" precheck is reproduced — see Known limits).

`app_id`+`installation_id` must both be configured for the github_app path; `installation_id`
absence is caught by the hook itself (not the `when` gate, which only tests `app_id`'s truthiness —
see the parity-deviation ledger item G5) with the same error class legacy raises.

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
  only stream with SERVER-SIDE incremental filtering (`since` query param, matches legacy exactly).
  `issue_comments`/`pull_request_review_comments` are also server-side `since`-filtered.
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
  `owner/repo` string) is NOT reproduced — see Known limits (`ENGINE_GAP`: `computed_fields`
  templates can only reference `record.*`, never `config.*`).
- `checkOrigin`/link-header pagination is NOT used: legacy's own `readPages` is `page`/`per_page`
  query-param pagination, not RFC 5988 Link-header following, so `page_number` (not `link_header`)
  is the byte-accurate parity choice despite GitHub's REST API also supporting Link headers.

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

## Known limits

- **`ENGINE_GAP` — legacy's `repository` marker field (every stream stamps the `owner/repo` string
  onto every emitted record) is NOT reproduced.** `streams.json`'s `computed_fields` templates are
  resolved via `Vars{Record: raw}` only (`engine/read.go`'s `applyComputedFields`) — `config.*` is
  never wired into that interpolation environment, unlike every OTHER templating surface in the
  dialect (base URL, headers, query, path, auth all receive both `Config` and `Record`/`Secrets`).
  A `computed_fields` value referencing `{{ config.owner }}` therefore hard-errors with "unresolved
  key owner in config" on every single record. This is a genuine dialect gap, not a workaround
  opportunity: there is no declarative way to stamp a config-derived constant onto every record
  today. Filed as `ENGINE_GAP` (not worked around with a 3rd hook interface, which would exceed the
  Tier-2 cap already spent on AuthHook+WriteHook). `owner`/`repo` remain available on
  `RuntimeConfig.Config` for any caller that needs them; they are simply not copied onto each record.
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
- Full GitHub REST surface (orgs, teams, projects v2, notifications, code scanning, dependabot,
  secrets administration, webhooks, GraphQL) is out of scope; see `api_surface.json`'s
  `excluded: {category: out_of_scope, ...}` entries.

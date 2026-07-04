# Overview

dbt Cloud is a workflow orchestration product for dbt (data build tool). This bundle reads dbt
Cloud projects, runs, repositories, users, environments, jobs, invites, licenses, notifications,
and SSH tunnels, and writes job/notification/SSH-tunnel mutations plus job/run control actions
(trigger, retry, cancel), through the dbt Cloud Administrative API v2, account-scoped under
`/accounts/{account_id}/`. This bundle migrates `internal/connectors/dbt` (the hand-written
connector; the legacy package stays registered and unchanged until wave6's registry flip) and,
as of this Pass B full-surface expansion, covers every documented v2 operation — see
`api_surface.json` for the endpoint-by-endpoint accounting (cross-checked against the published
OpenAPI spec, `github.com/dbt-labs/dbt-cloud-openapi-spec/openapi-v2.yaml`).

## Auth setup

Provide a dbt Cloud service token via the `api_key_2` secret; it is sent as the `Authorization`
header in the form `Token <api_key_2>` (`auth: [{"mode": "api_key_header", "header":
"Authorization", "prefix": "Token "}]`), matching legacy's `connsdk.APIKeyHeader("Authorization",
secret, "Token ")`. A required `account_id` config value scopes every list/write endpoint under
`/accounts/{account_id}/`.

## Streams notes

All 10 streams (`projects`, `runs`, `repositories`, `users`, `environments`, `jobs`, `invites`,
`notifications`, `ssh_tunnels`, and singleton `licenses`) share the identical envelope shape: `GET`
against the account-scoped resource, records at `data`, primary key `["id"]` (`licenses` alone
keys off `account_id` — it is a single account-wide summary object, not a per-record list). dbt
Cloud's list envelope is `{"data": [...], "extra": {"pagination": {"count", "total_count"},
"filters": {"limit", "offset"}}, "status": {...}}`; pagination uses `pagination.type:
offset_limit` with `limit_param: limit`, `offset_param: offset`, `page_size: 100` (matches
legacy's `dbtDefaultPageSize`/`dbtMaxPageSize`, both 100) for every paginated stream — the
engine's offset/limit paginator stops on a short page. `licenses` overrides `pagination.type` to
`none`: the real `GET /accounts/{account_id}/licenses/` endpoint accepts no `limit`/`offset` query
params at all (it always returns the account's one summary record), so declaring the base
offset/limit pagination here would send dead query parameters the API silently ignores — an honest
per-stream override, not a workaround.

New in this Pass B expansion: `jobs` (job definitions — schedule, steps, dbt version, docs
generation flags), `invites` (pending/redeemed user invitations, including nested `account`/
`groups` context flattened by ordinary schema projection), `notifications` (per-user job-status
notification configurations — email or Slack channel destinations, `on_cancel`/`on_failure`/
`on_success`/`on_warning` arrays of job ids), and `ssh_tunnels` (`GET
/accounts/{account_id}/encryptions/` — SSH tunnel configurations for warehouse connections;
`private_key` is never returned by the real API's list response and is not declared in this
stream's schema, only in `create_ssh_tunnel`/`update_ssh_tunnel`'s write-side `record_schema`
where a caller legitimately supplies one).

No stream declares an `incremental` block: the dbt Cloud Administrative API v2 list endpoints this
bundle reads have no server-side updated-since filter parameter (`runs`/`jobs` accept rich
query-param filters — `created_at__range`, `state`, `project_id__in`, etc. — but none of them is a
cursor-shaped "changed since X" filter the engine's `incremental.request_param` dialect can drive),
so every sync here is full-refresh, exactly matching legacy (which never declared cursor fields for
any dbt stream either).

## Write actions & risks

13 write actions now cover every dialect-expressible dbt Cloud mutation:

- **Job lifecycle**: `create_job`/`update_job`/`delete_job` (job definition CRUD — schedule, steps,
  environment, docs generation). `delete_job` is idempotent-delete (`missing_ok_status: [404]`).
- **Run control**: `trigger_job_run` (kicks off a real run against the job's configured warehouse
  connection — genuine warehouse side effects), `retry_failed_job` (retries a job's most recent
  failed run from the point of failure), `cancel_run` (stops an in-progress run), `retry_run`
  (retries a specific failed run by id). All four run-control actions carry an explicit
  warehouse-side-effect risk string; none is a pure metadata mutation.
- **Notifications**: `create_notification`/`update_notification`/`delete_notification` — registers,
  repoints, or removes an outbound job-status notification (email address or Slack channel of the
  caller's choosing).
- **SSH tunnels**: `create_ssh_tunnel`/`update_ssh_tunnel`/`delete_ssh_tunnel` — SSH tunnel
  configuration for a warehouse connection; the create/update `record_schema` accepts an optional
  `private_key` field (the real API accepts one on write, though it is never echoed back on read),
  so this is the one write-side schema in this bundle that legitimately declares a
  credential-shaped field outside `spec.json`'s `x-secret` mechanism — see Known limits.

`capabilities.write` is now `true` (previously `false`); `metadata.json`'s `risk.write` and
`risk.approval` document per-action risk tiers (irreversible deletes and warehouse-affecting run
control both require approval; job/notification/SSH-tunnel create is low-risk).

## Known limits

- Legacy's `harvest` loop has a defensive extra stop condition this bundle does not reproduce:
  when the response's `extra.pagination.total_count` is present and the running emitted count has
  already reached it, legacy stops even if the just-fetched page was exactly `page_size` long (a
  page that happens to end precisely on a `page_size` boundary). The engine's `offset_limit`
  paginator only recognizes the short-page stop signal (`recordCount < page_size`). For any real
  dbt Cloud account this is a one-extra-page-request difference at most (the very next page would
  come back empty or short and stop there), never a difference in which records are emitted —
  documented parity deviation, not a fixable gap (`offset_limit`'s stop rule is fixed engine
  behavior, not spec-declarable). See `docs/migration/conventions.md` §5's meta-rule: this never
  changes emitted record data for any input legacy would accept.
- `page_size`/`max_pages` config keys legacy exposed are not declared in `spec.json`: the engine's
  `offset_limit` paginator reads its page size and `MaxPages` only from `streams.json`'s
  statically-declared `pagination` block, with no mechanism to source either from
  `RuntimeConfig.Config` at read time (same limitation documented for searxng's `page_size`/
  `max_pages`, `docs/migration/conventions.md`'s Tier-1 read-only variant section). Declaring a
  `spec.json` property no template ever consumes is dead config (F6, REVIEW.md), so this bundle
  fixes `page_size: 100` (legacy's own default and max) and leaves `max_pages` unbounded (`0`,
  legacy's own "unlimited" default) rather than declaring unwireable config.
- `account_id` is required in `spec.json` even though the dbt Cloud API technically supports a
  "default account" lookup for some UI flows; legacy always requires it explicitly
  (`dbtAccountID`), so this bundle matches that stricter, unambiguous behavior.
- **Run/job artifact downloads (`manifest.json`/`run_results.json`/`catalog.json`) are out of
  scope**: `GET /accounts/{account_id}/runs/{run_id}/artifacts/{remainder}` and the job-scoped
  equivalent return the raw artifact file body (often large, and shaped by the dbt project itself,
  not by the dbt Cloud API), not a structured JSON list/detail record — `binary_payload` per
  `api_surface.json`. `GET .../runs/{run_id}/artifacts/` (no filename) lists only the available
  artifact filenames as a precursor to a per-file download and carries no independently syncable
  record data of its own.
- **`create_ssh_tunnel`/`update_ssh_tunnel`'s `private_key` field is a write-only credential the
  engine's `x-secret` mechanism does not cover**: `x-secret` governs `spec.json` CONNECTION
  properties (redacted from `DryRunWrite` previews via `Schema.SecretKeys()`); a write action's
  `record_schema` has no equivalent secret-marking mechanism in this dialect (write records are
  caller-supplied per-call data, not persisted connector config). A caller passing a real private
  key through `create_ssh_tunnel`/`update_ssh_tunnel` should treat it with the same care as any
  other credential; this bundle does not invent a redaction mechanism outside the engine's existing
  `x-secret` scope (an `ENGINE_GAP` would be the correct escalation if write-record secret redaction
  becomes a requirement, not a per-connector workaround).
- The v2 Projects/Environments/Repositories CRUD endpoints (`POST`/`PUT`/`DELETE` on those three
  resources) are dbt-labs-marked `Deprecated. Consider using the v3 API instead` in the published
  OpenAPI spec; this bundle targets v2 only (matching its `base_url` default and legacy's own v2
  usage) and does not implement writes against a deprecated surface — see `api_surface.json`'s
  `deprecated`-category exclusions.
- Account-admin-scoped endpoints (account settings mutation, user permission grants, a user's own
  profile self-update, the account-listing endpoint) are excluded as `requires_elevated_scope`/
  `deprecated`: these are identity/billing-administration actions, not syncable connector data or a
  data-record mutation this connector's `account_id`-scoped read/write model addresses.

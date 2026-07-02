# Overview

n8n is a workflow automation platform. This bundle reads workflows, executions, tags, and users
from a self-hosted or cloud n8n instance through its public REST API. It is read-only, migrating
`internal/connectors/n8n` (the hand-written legacy connector, which stays registered and unchanged
until wave6's registry flip) at capability parity.

## Auth setup

Provide the n8n API key via the `api_key` secret; it is sent as the `X-N8N-API-KEY` header on
every request and is never logged, matching legacy's `connsdk.APIKeyHeader("X-N8N-API-KEY", ...)`
wiring.

## Streams notes

All 4 streams (`workflows`, `executions`, `tags`, `users`) list against
`{"data":[...],"nextCursor":"..."}`-enveloped n8n endpoints, mapped field-for-field from legacy's
own `mapRecord` functions (`internal/connectors/n8n/streams.go`). Pagination is `cursor`
(`cursor_param: "cursor"`, `token_path: "nextCursor"`, no `stop_path` — n8n's `nextCursor` is
`null`/absent on the final page and that alone stops pagination, matching legacy's own stop
condition exactly, which never consults a separate boolean flag). Every request sends
`limit={{ config.page_size | default 100 }}` (matches legacy's default `n8nDefaultPageSize`).

## Write actions & risks

None. n8n is exposed read-only in this bundle (`capabilities.write: false`), matching legacy's
`Write` stub that always returns `connectors.ErrUnsupportedOperation`.

## Known limits

- **`base_url` is required and must be fully-qualified, including the `/api/v1` version path** —
  legacy derives its base URL from either an explicit `base_url` OR a bare `host`/instance-URL
  config value, defaulting the scheme to `https` and appending `/api/v1` automatically when the
  caller's value doesn't already include it (`n8nBaseURL`, `internal/connectors/n8n/n8n.go:256-284`).
  The engine's `spec.json` `"default"` materialization mechanism only fills in a FIXED literal
  default, not one derived/concatenated from another config value at read time — there is no
  declarative way to express "append `/api/v1` unless already present" without inventing ad hoc Go
  (a Tier-2 escalation this bundle does not need for anything else). This bundle therefore requires
  the caller to configure the fully-qualified `base_url` (e.g.
  `https://your-instance.n8n.cloud/api/v1`) directly. This is a documented config-surface
  narrowing per `docs/migration/conventions.md` §3's "derived default" guidance, not a silent
  behavior change to any request this bundle actually sends once configured. The separate `host`
  config key legacy also accepted is not modeled.
- Full n8n API surface (workflow create/activate/deactivate/delete, execution retry, credentials,
  variables, source control, audit logs) is out of scope for this wave; see `api_surface.json`'s
  `excluded` entries. Only the 4 legacy-parity read streams are implemented.
- `max_pages` config accepts an integer, `all`, or `unlimited` (0/absent means unbounded), matching
  legacy's `n8nMaxPages` parsing exactly.

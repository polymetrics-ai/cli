```
NAME
  pm schedule - manage authorized flow schedules

SYNOPSIS
  pm schedule create --name nightly --cron "0 2 * * *" --flow nightly_leads --authorization auth_<opaque-id> [--json]
  pm schedule list [--json]
  pm schedule inspect nightly [--json]
  pm schedule status nightly [--json]
  pm schedule install nightly [--crontab] [--json]
  pm schedule remove nightly [--crontab] [--json]
  pm schedule fire nightly --authorization auth_<opaque-id> [--json]

DESCRIPTION
  Schedules bind a cron expression and a named flow to a durable, revocable
  authorization reference. The selected local scheduler backend invokes
  schedule fire with that reference. On every firing, pm re-derives the
  content-free action scope and obtains a run-scoped grant before it sends a
  connector request. Use inspect or status to view the safe reference and the
  last fire state. Use --crontab on install or remove to force the crontab
  backend.

SECURITY
  A manifest, rendered scheduler payload, and fire receipt retain only the
  opaque authorization reference and safe receipt identifiers; they never
  retain approval tokens, credentials, payloads, or secret-derived values.
  Expired, revoked, or scope-changed authorization stops before a provider
  request. Failed, rate-limited, ambiguous, or cleanup-failed fires park and
  never replay automatically.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```

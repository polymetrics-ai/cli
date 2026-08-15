```
NAME
  pm schedule - run existing approved-job flows on a local scheduler

SYNOPSIS
  pm schedule create --name nightly --cron "0 2 * * *" --flow nightly_leads [--json]
  pm schedule list [--json]
  pm schedule inspect nightly [--json]
  pm schedule status nightly [--json]
  pm schedule install nightly [--crontab] [--json]
  pm schedule remove nightly [--crontab] [--json]
  pm schedule fire nightly [--json]

DESCRIPTION
  Approve each ETL or reverse-ETL job once, compose those existing approved jobs
  into a stored flow with pm flow create, then schedule that existing flow.
  Create refuses a missing or invalid flow before writing a schedule. Install
  revalidates the flow before touching the scheduler backend.

  The selected backend invokes exactly:

    pm --root <root> flow run <name> --json

  No approval token or authorization reference is placed in crontab, argv, a
  schedule manifest, or schedule JSON. Use inspect or status to view terminal
  flow status, safe prepared-execution identities, and opaque receipt IDs. Use
  --crontab on install or remove to force the crontab backend.

SECURITY
  Each unattended firing reloads every referenced job and revalidates credential
  revision, manifest/schema, source scope, mappings, destination action,
  confirmation policy, expiry, and revocation before a provider request. Any
  drift refuses and parks the schedule. Failed, rate-limited, ambiguous, or
  cleanup-failed writes also park or halt and never replay automatically.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```

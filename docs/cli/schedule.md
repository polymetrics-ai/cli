```
NAME
  pm schedule - create, list, install, and remove flow schedules

SYNOPSIS
  pm schedule create --name nightly --cron "0 2 * * *" --flow nightly_leads [--json]
  pm schedule list [--json]
  pm schedule install nightly [--crontab] [--json]
  pm schedule remove nightly [--crontab] [--json]

DESCRIPTION
  Schedules bind a cron expression to a named flow and install it into the
  selected local scheduler backend. Use --crontab on install or remove to force
  the crontab backend. The payload is pm flow run.

SECURITY
  Schedules do not embed secret values. Flow execution still uses the normal
  project credential references and approval gates.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```

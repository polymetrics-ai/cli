```
NAME
  pm environments - list cached PM Broker Environment metadata

SYNOPSIS
  pm environments list [--workspace <workspace-id>] [--json]
  pm environments show <environment-id> [--json]

DESCRIPTION
  Shows cached Environment metadata from safe PM Broker contexts only.
  Production environments default to remote runtime mode, and production writes
  or scheduled production jobs cannot use local fallback.

SECURITY
  Output contains Environment IDs, parent IDs, display names, and environment
  type only. Live provider operations remain unsupported.

EXIT STATUS
  0 success
  2 usage error
  3 validation error

```

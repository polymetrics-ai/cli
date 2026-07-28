```
NAME
  pm organizations - list cached PM Broker Organization metadata

SYNOPSIS
  pm organizations list [--json]
  pm organizations show <organization-id> [--json]

DESCRIPTION
  Shows cached Organization metadata from safe PM Broker contexts only. This
  foundation does not call a live broker and does not infer membership from
  display names or project defaults.

SECURITY
  Output contains Organization IDs and display names only. No credentials or
  raw secret values are read or printed.

EXIT STATUS
  0 success
  2 usage error
  3 validation error

```

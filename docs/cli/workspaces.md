```
NAME
  pm workspaces - list cached PM Broker Workspace metadata

SYNOPSIS
  pm workspaces list [--organization <org-id>] [--json]
  pm workspaces show <workspace-id> [--json]

DESCRIPTION
  Shows cached Workspace metadata from safe PM Broker contexts only. Workspaces
  remain bound to immutable Organization IDs; display names never grant access.

SECURITY
  Output contains Workspace IDs, Organization IDs, and display names only.

EXIT STATUS
  0 success
  2 usage error
  3 validation error

```

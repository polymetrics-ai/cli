# Security Trace

- Operation rows are metadata only and do not participate in command dispatch.
- `blocked_by_default` must be true for every operation row.
- Sensitive/admin/destructive/disallowed models require `source_url` or `notes`.
- Duplicate rows require `duplicate_of`.
- No secrets, credentials, raw API execution, or generic write surfaces were added.

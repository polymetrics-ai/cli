# Inline code review — issue #3628 Email IMAP/SMTP connector

Date: 2026-08-06. Scope: native Email reader/writer, bundle declaration, registration, command
surface, generated docs, and dependency change.

## Resolved finding

| Severity | Finding | Resolution | Evidence |
| --- | --- | --- | --- |
| warning | `mailboxes` ignored `ReadRequest.Limit`, so a caller requesting one record received every listed mailbox. | Replaced `LIST.Collect()` with a streaming bounded loop that drains/closes the command safely. | Red: `TestMailboxListHonorsRequestedLimit` emitted 2 for limit 1. Green: same test passes. |

## Current disposition

Email exposes mailbox listing and the SMTP send path only. Message reads, full refresh, and sparse
UID continuation are blocked pending #3810: `internal/app/app.go:350-376` lacks catalog sync-mode
validation, `internal/app/app.go:543-551` forwards persisted cursor state,
`internal/app/types.go:40-47` lacks scan-continuation state, and
`internal/app/local_warehouse.go:246-256` persists only an emitted cursor. Both become available
when #3810 lands. Review also confirms no SMTP read surface, no fabricated REST endpoint, no live
provider call, and prepared SMTP payload binding to the destructive approval digest.

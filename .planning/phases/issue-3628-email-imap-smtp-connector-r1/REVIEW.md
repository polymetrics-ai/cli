# Inline code review — issue #3628 Email IMAP/SMTP connector

Date: 2026-08-06. Scope: native Email reader/writer, bundle declaration, registration, command
surface, generated docs, and dependency change.

## Resolved finding

| Severity | Finding | Resolution | Evidence |
| --- | --- | --- | --- |
| warning | `mailboxes` ignored `ReadRequest.Limit`, so a caller requesting one record received every listed mailbox. | Replaced `LIST.Collect()` with a streaming bounded loop that drains/closes the command safely. | Red: `TestMailboxListHonorsRequestedLimit` emitted 2 for limit 1. Green: same test passes. |

## Final disposition

No open critical, warning, or security findings remain. Review confirmed: no SMTP read surface,
no fabricated REST endpoint, no redacting output declaration, no live provider call, bounded MIME
and attachment handling, UIDVALIDITY+UID cursor use, hard-delete documentation, and prepared SMTP
payload binding to the destructive approval digest. `make lint` reports zero issues.

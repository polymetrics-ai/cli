# Overview

Email is a native protocol connector, not the Gmail or Outlook API connectors. IMAP4rev2 lists
mailboxes. SMTP submission is the send-only write side: it submits a typed message and does not
list, fetch, or search mail.

The implementation uses [RFC 9051](https://www.rfc-editor.org/rfc/rfc9051.html) for IMAP and
[RFC 6409](https://www.rfc-editor.org/rfc/rfc6409.html) for SMTP submission. RFC 9051 explicitly
separates IMAP message access from mail posting, which is handled by a submission protocol.

## Auth setup

Supply the same connection fields as a mail client:

- `imap_host`, `imap_port` (`143` or `993`), and `imap_security` (`tls`, `starttls`, or `none`)
- `smtp_host`, `smtp_port` (`25`, `465`, or `587`), and `smtp_security` (`tls`, `starttls`, or `none`)
- `username` and secret `password`
- optional `smtp_username`, `from_address`, and `connection_timeout_seconds`

The port and security fields are closed constraints validated before credential persistence. The
password is a secret and is never printed, logged, or put in a preview.
`tls` means implicit TLS; `starttls` upgrades the established protocol connection. Use `none` only
for a trusted local/test server: remote password authentication over an unencrypted connection is
rejected.

## Streams notes

`pm email mailboxes list` issues IMAP `LIST`. Message reads are not exposed while their full-refresh
and sparse-UID continuation semantics cannot be enforced by the shared ETL boundary.

Email message reads, full-refresh enforcement, and sparse UID scan continuation are blocked pending
#3810. Full-refresh enforcement needs catalog sync-mode validation at
`internal/app/app.go:350-376` and a mode-aware change to persisted-cursor forwarding at
`internal/app/app.go:555-563`; until then, a default full refresh could silently become incremental.
Sparse UID scan continuation needs scan-continuation state at `internal/app/types.go:40-47` and non-emitted checkpoint persistence at
`internal/app/local_warehouse.go:246-256`; until then, an empty sparse range could be scanned
repeatedly. Both become available when #3810 lands.

When #3810 enables message polling, its RFC 9051 UIDVALIDITY plus UID cursor will not observe hard
deletes: a removed message simply stops appearing and no tombstone is emitted. IMAP IDLE/push
subscriptions remain out of scope and belong to the webhook/subscription seam in #3614.

## Write actions & risks

`pm email message send` is the only SMTP capability. It accepts typed `to`, optional `cc` and
`bcc`, `subject`, `body`, optional `body_content_type`, and attachment paths relative to the Email
attachment staging root, `<project-root>/.polymetrics/email-attachments/`. Stage regular files there; absolute paths,
traversal, and escaping symlinks are rejected before preview.
Attachments become part of the preview-bound RFC 5322 MIME payload.

SMTP submission is externally visible and irreversible after the server accepts the message. It is
non-batchable and always follows plan → preview → approval with typed destructive confirmation →
execute. The preview includes the exact unmasked envelope and MIME data that will be sent; BCC
recipients appear in the envelope preview but not the RFC 5322 headers.

## Known limits

- Email message reads, full-refresh enforcement, and sparse UID scan continuation are blocked
  pending #3810. Full-refresh enforcement needs `internal/app/app.go:350-376` catalog sync-mode
  validation and a mode-aware change to `internal/app/app.go:555-563` persisted-cursor forwarding.
  Sparse UID scan continuation needs scan-continuation state at `internal/app/types.go:40-47` and persistence at
  `internal/app/local_warehouse.go:246-256`. Both become available when #3810 lands.
- When message polling becomes available, hard deletions will not be observable and no deletion
  tombstone will be produced. IMAP IDLE, webhooks, and subscriptions remain #3614's seam.
- SMTP is send-only. It does not and will not back mailbox, message, search, or stream reads.
- Attachments must be relative regular files beneath `<project-root>/.polymetrics/email-attachments/`.
  The aggregate attachment limit is 25 MiB and each file is limited to 10 MiB.

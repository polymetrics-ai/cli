# Overview

Email is a native protocol connector, not the Gmail or Outlook API connectors. IMAP4rev2 is the
polled read side: it lists mailboxes and reads mailbox messages. SMTP submission is the send-only
write side: it submits a typed message and does not list, fetch, or search mail.

The implementation uses [RFC 9051](https://www.rfc-editor.org/rfc/rfc9051.html) for IMAP and
[RFC 6409](https://www.rfc-editor.org/rfc/rfc6409.html) for SMTP submission. RFC 9051 explicitly
separates IMAP message access from mail posting, which is handled by a submission protocol.

## Auth setup

Supply the same connection fields as a mail client:

- `imap_host`, `imap_port` (`143` or `993`), and `imap_security` (`tls`, `starttls`, or `none`)
- `smtp_host`, `smtp_port` (`25`, `465`, or `587`), and `smtp_security` (`tls`, `starttls`, or `none`)
- `username` and secret `password`
- optional `smtp_username`, `from_address`, `connection_timeout_seconds`, and `mailbox`

The port and security fields are closed constraints validated before credential persistence. The
password is a secret and is never printed, logged, included in fixtures, or put in a preview.
`tls` means implicit TLS; `starttls` upgrades the established protocol connection. Use `none` only
for a trusted local/test server: remote password authentication over an unencrypted connection is
rejected.

## Streams notes

`pm email mailboxes list` issues IMAP `LIST`. `pm email messages list` selects one mailbox and
emits a bounded number of messages with envelope, flags, internal date, RFC822 size, and bounded
leaf body parts. The connector requests at most 1 MiB per body part and never advertises message
search as an SMTP feature.

The `messages` incremental cursor is a mailbox-scoped encoding of `UIDVALIDITY` plus UID, not a
received-date timestamp. RFC 9051 §§2.3.1 and 2.3.2 define UIDs as the mailbox's stable message
identity while `UIDVALIDITY` detects when that UID namespace changes; a date is neither monotonic
nor a mailbox identity. The cursor cannot be moved to a different mailbox. When UIDVALIDITY
changes, the connector starts from that mailbox's new UID namespace.

Polling cannot see a hard delete. A message removed from the mailbox simply stops appearing; this
connector emits no tombstone and makes no claim to detect it. IMAP IDLE/push subscriptions are out
of scope here and belong to the webhook/subscription seam in #3614.

## Write actions & risks

`pm email message send` is the only SMTP capability. It accepts typed `to`, optional `cc` and
`bcc`, `subject`, `body`, optional `body_content_type`, and project-relative attachment paths.
Attachments are regular files and become part of the preview-bound RFC 5322 MIME payload.

SMTP submission is externally visible and irreversible after the server accepts the message. It is
non-batchable and always follows plan → preview → approval with typed destructive confirmation →
execute. The preview includes the exact unmasked envelope and MIME data that will be sent; BCC
recipients appear in the envelope preview but not the RFC 5322 headers.

## Known limits

- This connector polls; it does not implement IMAP IDLE, webhooks, or subscriptions (#3614 owns that seam).
- Hard deletions are not observable in an incremental poll; no deletion tombstone is produced.
- Message body parts are intentionally partial (1 MiB each; at most 32 leaf parts per message)
  and the command `--limit` bounds emitted messages. Large or nested MIME content can therefore
  be truncated by design.
- SMTP is send-only. It does not and will not back mailbox, message, search, or stream reads.
- Attachment paths must resolve to regular files under the project root. The aggregate attachment
  limit is 25 MiB and each file is limited to 10 MiB.

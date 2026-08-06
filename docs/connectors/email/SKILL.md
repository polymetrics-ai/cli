---
name: pm-email
description: Email (IMAP + SMTP) connector knowledge and safe action guide.
---

# pm-email

## Purpose

Reads mailboxes and bounded message data through IMAP4rev2, and sends one typed RFC 5322 message through SMTP submission after plan, preview, approval, and destructive confirmation.

## Icon

- id: pm-sample
- asset: icons/pm-sample.svg
- source: polymetrics
- review_status: polymetrics
- review_url: https://github.com/polymetrics-ai/cli

## Capabilities

- check=true catalog=true read=true write=true query=false
- Integration type: email

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- imap_host (required)
- imap_port (required)
- imap_security (required)
- smtp_host (required)
- smtp_port (required)
- smtp_security (required)
- username (required)
- smtp_username
- from_address
- connection_timeout_seconds default=30
- mailbox default=INBOX
- password (secret) (required)

## ETL Streams

- mailboxes: Mailboxes returned by IMAP LIST.
  - primary key: name
  - fields: name(string), delimiter(string), attributes(array)
- messages: Messages from one IMAP mailbox, keyed and incremented by mailbox UIDVALIDITY plus UID. Hard deletions are not observable by polling. Full refresh is unavailable pending #3810.
  - primary key: mailbox, uid_validity, uid
  - cursor: imap_cursor
  - fields: mailbox(string), uid_validity(string), uid(string), imap_cursor(string), envelope(object), flags(array), internal_date(string), size(integer), body_parts(array)

## Sync Modes

- ETL sync modes: incremental_append
- Source modes: incremental

## Reverse ETL Actions

- send_message: Submit one RFC 5322/MIME message through SMTP; attachment paths are relative to the runtime .polymetrics staging root.
  - endpoint: SMTP MAIL/RCPT/DATA
  - required fields: to, subject, body
  - optional fields: cc, bcc, body_content_type, attachments
  - risk: submits one externally visible SMTP message; it cannot be undone after the server accepts DATA
  - bulk reverse ETL: refused (non-batchable; run it as its own pm command, one record at a time)

## Security

- read risk: polled IMAP reads; hard deletion is not observable by polling
- write risk: SMTP send-only submission
- approval: plan, unmasked preview, typed destructive confirmation, and approval are required before SMTP submission
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Command Surface

- Read IMAP mailboxes and send explicitly approved SMTP messages.
- Usage: pm email <command> [options]
- Source CLI: IMAP4rev2 and SMTP submission (https://www.rfc-editor.org/rfc/rfc6409.html)
- IMAP reads
- SMTP submission
- Other Commands
  - mailboxes list - List mailboxes through IMAP LIST. [intent=etl availability=implemented stream=mailboxes]
  - messages list - Read bounded message envelopes, flags, dates, sizes, and bounded body parts from one IMAP mailbox. [intent=etl availability=implemented stream=messages]; flags: --mailbox
  - message send - Plan, preview, approve, and submit one typed SMTP message; the preview includes the exact unmasked MIME payload. [intent=reverse_etl availability=implemented write=send_message]; approval: Requires plan, unmasked preview, approval evidence, and typed destructive confirmation.; risk: SMTP submission is externally visible and irreversible after the server accepts it.; flags: --to (required), --cc, --bcc, --subject (required), --body (required), --body-content-type, --attachment
- Help topics:
  - incremental - Messages use a mailbox-specific UIDVALIDITY+UID cursor; hard deletions are not observable by polling; full message refresh is unavailable pending #3810.
  - submission-safety - SMTP is send-only; message send is non-batchable, destructive, approval-gated, and previewed without masking.

## Commands

### Inspect as a manual

```bash
pm connectors inspect email
```

### Inspect as structured JSON

```bash
pm connectors inspect email --json
```

## Agent Rules

- Run pm connectors inspect email before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
- For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.

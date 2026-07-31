# Overview

Reads Gmail messages, threads, drafts, labels, history, filters, send-as aliases, delegates,
forwarding addresses, mailbox profile, and bounded redacted direct detail/settings resources, and
writes approved reverse-ETL mutations through the Google OAuth 2.0 refresh-token grant.

Executable coverage after the wave03 re-audit against Gmail Discovery revision `20260727`: 10 ETL
streams, 11 bounded direct-read commands, and 40 typed reverse-ETL write actions. The only
non-executable official operations are Gmail watch/stop changefeed controls blocked on shared CDC
foundations and 11 Google Workspace Client-Side Encryption admin operations that remain not
applicable to this general Gmail mailbox connector. Fixture-only validation is not live provider
certification.

Service API documentation: https://developers.google.com/gmail/api/reference/rest.

## Auth setup

Connection fields include OAuth refresh-token secrets (`client_id`, `client_secret`,
`client_refresh_token`), `token_url`, `scopes`, `base_url`, `user_id`, `userId`, `page_size`,
`start_date`, `include_spam_and_trash`, and `start_history_id`. Secret values must come from an
environment variable or stdin; do not paste them in chat, docs, or shell history.

`user_id` is the legacy-compatible config key used by ETL streams and reverse-ETL writes. The
provider-style direct-read commands also declare `userId` (default `me`) because the direct-read
surface resolves official `{userId}` path variables by exact name. Use `me` for direct reads unless
the shared direct-read path-variable validator is expanded to accept email-style delegated user IDs.

Authentication behavior uses the connector-specific `gmail` AuthHook. It exchanges the configured
refresh token for a short-lived bearer token at the HTTPS `token_url`; non-HTTPS token endpoints fail
closed. Connection checks call GET `/users/{{ config.user_id }}/labels`.

## Streams notes

Stream-backed reads are `messages`, `threads`, `drafts`, `labels`, `history`, `filters`, `send_as`,
`delegates`, `forwarding_addresses`, and `profile`. Cursor pagination is used for `messages`,
`threads`, `drafts`, and `history`; settings/profile streams are single-page. `messages`, `threads`,
and `drafts` can apply `start_date` as a Gmail search `after:<unix-seconds>` query. `history` uses
Gmail `startHistoryId` and is the mailbox change-history stream; watch/stop lifecycle execution is
blocked until the shared CDC foundations define safe state/renewal/teardown semantics.

Provider-style direct reads are definition-owned in `cli_surface.json`. They are fixed-target,
bounded, and `json_redacted`: message/thread/draft detail, label/filter getters, singleton settings
(`autoForwarding`, `vacation`, `language`, `imap`, `pop`), and attachment get with `data` redacted. These are not raw
HTTP escape hatches and do not change warehouse `pm query` semantics.

## Write actions & risks

Reverse ETL writes must be planned, previewed, explicitly approved, and then executed. Gmail write
coverage includes message send/insert/import/modify/trash/untrash/delete, message batch modify and
batch delete, thread modify/trash/untrash/delete, draft create/update/send/delete, label
create/update/patch/delete, filter create/delete, send-as create/update/patch/delete/verify,
S/MIME insert/set-default/delete, delegate create/delete, forwarding address create/delete, and
account-wide auto-forwarding/vacation/language/IMAP/POP updates.

Destructive actions such as `delete_message`, `batch_delete_messages`, `delete_thread`, and
`delete_smime_info` require destructive confirmation through the reverse-ETL approval flow. Message
raw bodies, draft raw bodies, S/MIME PKCS#12 material/passwords, send-as signatures/SMTP passwords,
forwarding/delegate/send-as emails, vacation response bodies, and attachment/certificate direct-read
payload fields are redacted from operator-visible previews/errors/output where declared.

## Known limits

- Fixture-only work remains uncertified. No live Gmail credentials, provider calls, provider writes,
  VPS, or Thaalam checks were run for this wave.
- `users.watch` and `users.stop` are blocked by shared CDC/changefeed dependencies #2986/#2988;
  they are tracked as operation rows, not advertised as executable writes.
- Google Workspace Client-Side Encryption (`settings/cse/**`) is tracked as not-applicable/blocked
  admin operation rows with official source evidence. It requires enterprise CSE/KACLS setup and is
  outside the general Gmail mailbox connector.
- Direct-read path-variable validation currently supports identifier-safe values. The direct-read
  `userId` default is `me`; email-address path variables such as `sendAsEmail`, `delegateEmail`,
  and `forwardingEmail` are planned/blocked until the shared direct-read safety encoder is expanded.
  ETL/write templates continue to support those email path values via `user_id` and record fields.
- API coverage counts after this change: total=79, executable=61 (10 streams + 11 direct reads +
  40 writes), blocked/planned=18 (5 email-path direct reads + 2 CDC controls + 11 CSE
  not-applicable/admin rows), excluded/not-applicable=11 by semantic disposition, certified=0.

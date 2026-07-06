# Overview

Reads Gmail messages, threads, drafts, labels, history, filters, send-as aliases, delegates,
forwarding addresses, and mailbox profile, and writes approved reverse-ETL mutations
(send/insert/import/modify/trash/delete messages and threads; draft and label lifecycle; filter,
send-as, delegate, and forwarding-address management; vacation/language/IMAP/POP/auto-forwarding
settings) via the Google OAuth 2.0 refresh-token grant.

Readable streams: `messages`, `threads`, `drafts`, `labels`, `history`, `filters`, `send_as`,
`delegates`, `forwarding_addresses`, `profile`.

Write actions: `send_message`, `insert_message`, `import_message`, `modify_message`,
`trash_message`, `untrash_message`, `delete_message`, `modify_thread`, `trash_thread`,
`untrash_thread`, `delete_thread`, `create_draft`, `update_draft`, `send_draft`, `delete_draft`,
`create_label`, `update_label`, `patch_label`, `delete_label`, `create_filter`, `delete_filter`,
`create_send_as`, `update_send_as`, `patch_send_as`, `delete_send_as`, `verify_send_as`,
`create_delegate`, `delete_delegate`, `create_forwarding_address`, `delete_forwarding_address`,
`update_auto_forwarding`, `update_vacation`, `update_language`, `update_imap`, `update_pop`.

Service API documentation: https://developers.google.com/gmail/api/reference/rest.

## Auth setup

Connection fields:

- `base_url` (optional, string); default `https://gmail.googleapis.com/gmail/v1`; format `uri`;
  Gmail API base URL override for tests or proxies.
- `client_id` (required, secret, string); Google OAuth 2.0 client ID for the refresh-token grant.
  Used only in the token-request form; never logged.
- `client_refresh_token` (required, secret, string); Long-lived Google OAuth 2.0 refresh token.
  Exchanged for a short-lived access token at token_url; never logged. The 3-legged
  consent/acquisition dance is out of scope for this connector (credentials layer already owns it).
- `client_secret` (optional, secret, string); Google OAuth 2.0 client secret (optional for some
  client types). Used only in the token-request form; never logged.
- `include_spam_and_trash` (optional, string); When 'true', includes SPAM and TRASH in list results
  (includeSpamTrash=true).
- `page_size` (optional, string); default `100`; Records per page (1-500, maxResults).
- `scopes` (optional, string); default `https://mail.google.com/`; OAuth scope requested on the
  token-refresh grant.
- `start_date` (optional, string); format `date-time`; RFC3339 lower bound; translated to a Gmail
  search query 'after:<unix-seconds>' filter.
- `start_history_id` (optional, string); Gmail historyId lower bound for the history stream's
  incremental sync (users.history.list's required startHistoryId). Unlike messages/threads/drafts,
  history genuinely publishes a server-recognized, monotonically-increasing cursor (historyId) --
  see Streams notes in docs.md. Obtain an initial value from any message/thread/label/profile
  record's historyId field.
- `token_url` (optional, string); default `https://oauth2.googleapis.com/token`; format `uri`;
  Google OAuth 2.0 token endpoint override. MUST be https in production; the hook fails closed on a
  non-https or unparseable value to prevent exfiltrating the refresh token to an attacker-chosen
  endpoint.
- `user_id` (optional, string); default `me`; Gmail user ID path segment (email address for
  delegated access, or 'me').

Secret fields are redacted in logs and write previews: `client_id`, `client_refresh_token`,
`client_secret`.

Default configuration values: `base_url=https://gmail.googleapis.com/gmail/v1`, `page_size=100`,
`scopes=https://mail.google.com/`, `token_url=https://oauth2.googleapis.com/token`, `user_id=me`.

Authentication behavior:

- Connector-specific authentication using `secrets.client_refresh_token`, `config.token_url`,
  `secrets.client_id`, `secrets.client_secret`, `config.scopes`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/users/{{ config.user_id }}/labels`.

## Streams notes

Default pagination: cursor pagination; cursor parameter `pageToken`; next token from
`nextPageToken`.

Pagination by stream: cursor: `messages`, `threads`, `drafts`, `history`; none: `labels`, `filters`,
`send_as`, `delegates`, `forwarding_addresses`, `profile`.

Incremental streams use their declared cursor fields and send lower-bound parameters only when a
lower bound is available.

- `messages`: GET `/users/{{ config.user_id }}/messages` - records path `messages`; query
  `includeSpamTrash` from template `{{ config.include_spam_and_trash }}`, omitted when absent;
  `maxResults`=`{{ config.page_size }}`; `q` from template `after:{{ config.start_date |
  unix_seconds }}`, omitted when absent; cursor pagination; cursor parameter `pageToken`; next token
  from `nextPageToken`; computed output fields `thread_id`.
- `threads`: GET `/users/{{ config.user_id }}/threads` - records path `threads`; query
  `includeSpamTrash` from template `{{ config.include_spam_and_trash }}`, omitted when absent;
  `maxResults`=`{{ config.page_size }}`; `q` from template `after:{{ config.start_date |
  unix_seconds }}`, omitted when absent; cursor pagination; cursor parameter `pageToken`; next token
  from `nextPageToken`; computed output fields `history_id`.
- `drafts`: GET `/users/{{ config.user_id }}/drafts` - records path `drafts`; query
  `includeSpamTrash` from template `{{ config.include_spam_and_trash }}`, omitted when absent;
  `maxResults`=`{{ config.page_size }}`; `q` from template `after:{{ config.start_date |
  unix_seconds }}`, omitted when absent; cursor pagination; cursor parameter `pageToken`; next token
  from `nextPageToken`; computed output fields `message_id`, `thread_id`.
- `labels`: GET `/users/{{ config.user_id }}/labels` - records path `labels`; query
  `includeSpamTrash` from template `{{ config.include_spam_and_trash }}`, omitted when absent; `q`
  from template `after:{{ config.start_date | unix_seconds }}`, omitted when absent; computed output
  fields `label_list_visibility`, `message_list_visibility`, `messages_total`, `messages_unread`,
  `threads_total`, `threads_unread`.
- `history`: GET `/users/{{ config.user_id }}/history` - records path `history`; query
  `maxResults`=`{{ config.page_size }}`; `startHistoryId` from template `{{ config.start_history_id
  }}`, omitted when absent; cursor pagination; cursor parameter `pageToken`; next token from
  `nextPageToken`; incremental cursor `id`; sent as `startHistoryId`; formatted as `rfc3339`;
  initial lower bound from `start_history_id`; computed output fields `labels_added`,
  `labels_removed`, `messages_added`, `messages_deleted`.
- `filters`: GET `/users/{{ config.user_id }}/settings/filters` - records path `filter`.
- `send_as`: GET `/users/{{ config.user_id }}/settings/sendAs` - records path `sendAs`; computed
  output fields `display_name`, `is_default`, `is_primary`, `reply_to_address`, `send_as_email`,
  `treat_as_alias`, `verification_status`.
- `delegates`: GET `/users/{{ config.user_id }}/settings/delegates` - records path `delegates`;
  computed output fields `delegate_email`, `verification_status`.
- `forwarding_addresses`: GET `/users/{{ config.user_id }}/settings/forwardingAddresses` - records
  path `forwardingAddresses`; computed output fields `forwarding_email`, `verification_status`.
- `profile`: GET `/users/{{ config.user_id }}/profile` - records at response root; computed output
  fields `email_address`, `history_id`, `messages_total`, `threads_total`.

## Write actions & risks

Overall write risk: external Gmail API mutation, including sending real outbound email, permanently
deleting messages/threads/drafts, granting mailbox delegation, and changing account-wide
forwarding/vacation/IMAP/POP settings.

Reverse ETL writes should be planned, previewed, approved, and then executed. Declared actions:

- `send_message`: POST `/users/{{ config.user_id }}/messages/send` - kind `create`; body type
  `json`; required record fields `raw`; accepted fields `raw`, `threadId`; risk: sends a real
  outbound email on behalf of the mailbox owner; irreversible once delivered.
- `insert_message`: POST `/users/{{ config.user_id }}/messages` - kind `create`; body type `json`;
  required record fields `raw`; accepted fields `labelIds`, `raw`, `threadId`; risk: inserts a
  message directly into the mailbox without sending it (no SMTP delivery, no notifications) -- still
  a real, visible mailbox mutation.
- `import_message`: POST `/users/{{ config.user_id }}/messages/import` - kind `create`; body type
  `json`; required record fields `raw`; accepted fields `labelIds`, `raw`, `threadId`.
- `modify_message`: POST `/users/{{ config.user_id }}/messages/{{ record.id }}/modify` - kind
  `update`; body type `json`; path fields `id`; required record fields `id`; accepted fields
  `addLabelIds`, `id`, `removeLabelIds`; risk: changes label state on an existing message (e.g.
  moving in/out of INBOX/TRASH/UNREAD), visible to the mailbox owner.
- `trash_message`: POST `/users/{{ config.user_id }}/messages/{{ record.id }}/trash` - kind
  `update`; body type `none`; path fields `id`; required record fields `id`; accepted fields `id`;
  risk: moves a message to Trash; auto-purged by Gmail after 30 days.
- `untrash_message`: POST `/users/{{ config.user_id }}/messages/{{ record.id }}/untrash` - kind
  `update`; body type `none`; path fields `id`; required record fields `id`; accepted fields `id`;
  risk: restores a trashed message back to its prior labels.
- `delete_message`: DELETE `/users/{{ config.user_id }}/messages/{{ record.id }}` - kind `delete`;
  body type `none`; path fields `id`; required record fields `id`; accepted fields `id`; missing
  records treated as success for status `404`; confirmation `destructive`; risk: permanently deletes
  a message immediately, bypassing Trash; irreversible.
- `modify_thread`: POST `/users/{{ config.user_id }}/threads/{{ record.id }}/modify` - kind
  `update`; body type `json`; path fields `id`; required record fields `id`; accepted fields
  `addLabelIds`, `id`, `removeLabelIds`; risk: changes label state on every message in an existing
  thread.
- `trash_thread`: POST `/users/{{ config.user_id }}/threads/{{ record.id }}/trash` - kind `update`;
  body type `none`; path fields `id`; required record fields `id`; accepted fields `id`; risk: moves
  an entire thread to Trash; auto-purged by Gmail after 30 days.
- `untrash_thread`: POST `/users/{{ config.user_id }}/threads/{{ record.id }}/untrash` - kind
  `update`; body type `none`; path fields `id`; required record fields `id`; accepted fields `id`;
  risk: restores a trashed thread back to its prior labels.
- `delete_thread`: DELETE `/users/{{ config.user_id }}/threads/{{ record.id }}` - kind `delete`;
  body type `none`; path fields `id`; required record fields `id`; accepted fields `id`; missing
  records treated as success for status `404`; confirmation `destructive`; risk: permanently deletes
  every message in a thread immediately, bypassing Trash; irreversible.
- `create_draft`: POST `/users/{{ config.user_id }}/drafts` - kind `create`; body type `json`;
  required record fields `message`; accepted fields `message`; risk: creates a new unsent draft,
  visible to the mailbox owner.
- `update_draft`: PUT `/users/{{ config.user_id }}/drafts/{{ record.id }}` - kind `update`; body
  type `json`; path fields `id`; required record fields `id`, `message`; accepted fields `id`,
  `message`; risk: replaces the entire content of an existing draft.
- `send_draft`: POST `/users/{{ config.user_id }}/drafts/send` - kind `custom`; body type `json`;
  required record fields `id`; accepted fields `id`; risk: sends a real outbound email from an
  existing draft on behalf of the mailbox owner; irreversible once delivered.
- `delete_draft`: DELETE `/users/{{ config.user_id }}/drafts/{{ record.id }}` - kind `delete`; body
  type `none`; path fields `id`; required record fields `id`; accepted fields `id`; missing records
  treated as success for status `404`; risk: permanently deletes a draft; irreversible.
- `create_label`: POST `/users/{{ config.user_id }}/labels` - kind `create`; body type `json`;
  required record fields `name`; accepted fields `color`, `labelListVisibility`,
  `messageListVisibility`, `name`; risk: creates a new custom label visible in the mailbox owner's
  label list.
- `update_label`: PUT `/users/{{ config.user_id }}/labels/{{ record.id }}` - kind `update`; body
  type `json`; path fields `id`; required record fields `id`, `name`; accepted fields `color`, `id`,
  `labelListVisibility`, `messageListVisibility`, `name`; risk: replaces the full definition of an
  existing label (name/visibility/color); a system label's name cannot actually be changed by Gmail
  even though the request is accepted.
- `patch_label`: PATCH `/users/{{ config.user_id }}/labels/{{ record.id }}` - kind `update`; body
  type `json`; path fields `id`; required record fields `id`; accepted fields `color`, `id`,
  `labelListVisibility`, `messageListVisibility`, `name`; risk: partially updates an existing
  label's fields, leaving unset fields unchanged.
- `delete_label`: DELETE `/users/{{ config.user_id }}/labels/{{ record.id }}` - kind `delete`; body
  type `none`; path fields `id`; required record fields `id`; accepted fields `id`; missing records
  treated as success for status `404`; risk: removes a user label from the account and from every
  message/thread that carried it; system labels reject deletion with an error.
- `create_filter`: POST `/users/{{ config.user_id }}/settings/filters` - kind `create`; body type
  `json`; required record fields `criteria`; accepted fields `action`, `criteria`; risk: creates a
  mail filter that automatically acts on future incoming messages matching its criteria (may
  auto-forward mail externally).
- `delete_filter`: DELETE `/users/{{ config.user_id }}/settings/filters/{{ record.id }}` - kind
  `delete`; body type `none`; path fields `id`; required record fields `id`; accepted fields `id`;
  missing records treated as success for status `404`; risk: removes an existing mail filter; future
  messages stop being auto-actioned by it.
- `create_send_as`: POST `/users/{{ config.user_id }}/settings/sendAs` - kind `create`; body type
  `json`; required record fields `sendAsEmail`; accepted fields `displayName`, `replyToAddress`,
  `sendAsEmail`, `signature`, `smtpMsa`, `treatAsAlias`; risk: adds a new custom From: alias; Google
  emails a verification link to the new address before it can send mail.
- `update_send_as`: PUT `/users/{{ config.user_id }}/settings/sendAs/{{ record.sendAsEmail }}` -
  kind `update`; body type `json`; path fields `sendAsEmail`; required record fields `sendAsEmail`;
  accepted fields `displayName`, `isDefault`, `replyToAddress`, `sendAsEmail`, `signature`,
  `smtpMsa`, `treatAsAlias`; risk: replaces the full send-as alias configuration, including which
  alias is the account default.
- `patch_send_as`: PATCH `/users/{{ config.user_id }}/settings/sendAs/{{ record.sendAsEmail }}` -
  kind `update`; body type `json`; path fields `sendAsEmail`; required record fields `sendAsEmail`;
  accepted fields `displayName`, `isDefault`, `replyToAddress`, `sendAsEmail`, `signature`,
  `smtpMsa`, `treatAsAlias`; risk: partially updates an existing send-as alias, leaving unset fields
  unchanged.
- `delete_send_as`: DELETE `/users/{{ config.user_id }}/settings/sendAs/{{ record.sendAsEmail }}` -
  kind `delete`; body type `none`; path fields `sendAsEmail`; required record fields `sendAsEmail`;
  accepted fields `sendAsEmail`; missing records treated as success for status `404`; risk: removes
  a custom From: alias (the account's primary address cannot be deleted; Gmail rejects that
  request).
- `verify_send_as`: POST `/users/{{ config.user_id }}/settings/sendAs/{{ record.sendAsEmail
  }}/verify` - kind `custom`; body type `none`; path fields `sendAsEmail`; required record fields
  `sendAsEmail`; accepted fields `sendAsEmail`; risk: re-sends the verification email for a pending
  custom From: alias.
- `create_delegate`: POST `/users/{{ config.user_id }}/settings/delegates` - kind `create`; body
  type `json`; required record fields `delegateEmail`; accepted fields `delegateEmail`; risk: grants
  another account read/send/delete access to this mailbox (Google Workspace accounts only); a
  significant access-control change.
- `delete_delegate`: DELETE `/users/{{ config.user_id }}/settings/delegates/{{ record.delegateEmail
  }}` - kind `delete`; body type `none`; path fields `delegateEmail`; required record fields
  `delegateEmail`; accepted fields `delegateEmail`; missing records treated as success for status
  `404`; risk: revokes another account's delegated access to this mailbox.
- `create_forwarding_address`: POST `/users/{{ config.user_id }}/settings/forwardingAddresses` -
  kind `create`; body type `json`; required record fields `forwardingEmail`; accepted fields
  `forwardingEmail`; risk: proposes a new external forwarding address; Google emails a verification
  link before it can be used by update_auto_forwarding.
- `delete_forwarding_address`: DELETE `/users/{{ config.user_id }}/settings/forwardingAddresses/{{
  record.forwardingEmail }}` - kind `delete`; body type `none`; path fields `forwardingEmail`;
  required record fields `forwardingEmail`; accepted fields `forwardingEmail`; missing records
  treated as success for status `404`; risk: removes a forwarding address; if it is the account's
  current auto-forwarding target, forwarding stops.
- `update_auto_forwarding`: PUT `/users/{{ config.user_id }}/settings/autoForwarding` - kind
  `update`; body type `json`; required record fields `enabled`; accepted fields `disposition`,
  `emailAddress`, `enabled`; risk: changes the account-wide auto-forwarding singleton; when enabled,
  silently copies all future incoming mail to an external address.
- `update_vacation`: PUT `/users/{{ config.user_id }}/settings/vacation` - kind `update`; body type
  `json`; accepted fields `enableAutoReply`, `endTime`, `responseBodyHtml`, `responseBodyPlainText`,
  `responseSubject`, `restrictToContacts`, `restrictToDomain`, `startTime`; risk: changes the
  account-wide vacation-responder singleton; when enabled, auto-replies to external senders with the
  configured message.
- `update_language`: PUT `/users/{{ config.user_id }}/settings/language` - kind `update`; body type
  `json`; required record fields `displayLanguage`; accepted fields `displayLanguage`; risk: changes
  the Gmail web interface display language for the account.
- `update_imap`: PUT `/users/{{ config.user_id }}/settings/imap` - kind `update`; body type `json`;
  accepted fields `autoExpunge`, `enabled`, `expungeBehavior`, `maxFolderSize`; risk: changes the
  account-wide IMAP-access singleton; disabling breaks any external IMAP client currently connected.
- `update_pop`: PUT `/users/{{ config.user_id }}/settings/pop` - kind `update`; body type `json`;
  accepted fields `accessWindow`, `disposition`; risk: changes the account-wide POP-access
  singleton, including what happens to mail after it is fetched via POP.

## Known limits

- Batch defaults: read_page_size=100.
- API coverage includes 10 stream-backed endpoint group(s), 35 write-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  binary_payload=2, destructive_admin=1, duplicate_of=15, non_data_endpoint=2,
  requires_elevated_scope=14.

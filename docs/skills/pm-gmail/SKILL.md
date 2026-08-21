---
name: pm-gmail
description: Gmail connector knowledge and safe action guide.
---

# pm-gmail

## Purpose

Reads Gmail messages, threads, drafts, labels, history, filters, send-as aliases, delegates, forwarding addresses, and mailbox profile, and writes approved reverse-ETL mutations (send/insert/import/modify/trash/delete messages and threads; draft and label lifecycle; filter, send-as, delegate, and forwarding-address management; vacation/language/IMAP/POP/auto-forwarding settings) via the Google OAuth 2.0 refresh-token grant.

## Icon

- id: simple-icons-gmail
- asset: icons/simple-icons/gmail.svg
- title: Gmail
- simple_icon_slug: gmail
- simple_icon_hex: EA4335
- source: simple-icons
- license: CC0-1.0
- review_status: cc0_with_trademark_caveat
- review_url: https://simpleicons.org/?q=Gmail
- match: exact-name-or-slug
- matched_by: gmail

## Capabilities

- check=true catalog=true read=true write=true query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- base_url
- include_spam_and_trash
- page_size
- scopes
- start_date
- start_history_id
- token_url
- user_id
- client_id (secret) (required)
- client_refresh_token (secret) (required)
- client_secret (secret)

## ETL Streams

- messages:
  - primary key: id
  - fields: id(string), thread_id(string)
- threads:
  - primary key: id
  - fields: history_id(string), id(string), snippet(string)
- drafts:
  - primary key: id
  - fields: id(string), message_id(string), thread_id(string)
- labels:
  - primary key: id
  - fields: id(string), label_list_visibility(string), message_list_visibility(string), messages_total(integer), messages_unread(integer), name(string), threads_total(integer), threads_unread(integer), type(string)
- history:
  - primary key: id
  - cursor: id
  - fields: id(string), labels_added(array), labels_removed(array), messages_added(array), messages_deleted(array)
- filters:
  - primary key: id
  - fields: action(object), criteria(object), id(string)
- send_as:
  - primary key: send_as_email
  - fields: display_name(string), is_default(boolean), is_primary(boolean), reply_to_address(string), send_as_email(string), signature(string), smtpMsa(object), treat_as_alias(boolean), verification_status(string)
- delegates:
  - primary key: delegate_email
  - fields: delegate_email(string), verification_status(string)
- forwarding_addresses:
  - primary key: forwarding_email
  - fields: forwarding_email(string), verification_status(string)
- profile:
  - primary key: email_address
  - fields: email_address(string), history_id(string), messages_total(integer), threads_total(integer)

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped, incremental_append, incremental_append_deduped

## Reverse ETL Actions

- send_message:
  - endpoint: POST /users/{{ config.user_id }}/messages/send
  - required fields: raw
  - risk: sends a real outbound email on behalf of the mailbox owner; irreversible once delivered
- insert_message:
  - endpoint: POST /users/{{ config.user_id }}/messages
  - required fields: raw
  - risk: inserts a message directly into the mailbox without sending it (no SMTP delivery, no notifications) -- still a real, visible mailbox mutation
- import_message:
  - endpoint: POST /users/{{ config.user_id }}/messages/import
  - required fields: raw
  - risk: imports a message into the mailbox from an external mail migration source, bypassing spam classification by default
- modify_message:
  - endpoint: POST /users/{{ config.user_id }}/messages/{{ record.id }}/modify
  - required fields: id
  - risk: changes label state on an existing message (e.g. moving in/out of INBOX/TRASH/UNREAD), visible to the mailbox owner
- trash_message:
  - endpoint: POST /users/{{ config.user_id }}/messages/{{ record.id }}/trash
  - required fields: id
  - risk: moves a message to Trash; auto-purged by Gmail after 30 days
- untrash_message:
  - endpoint: POST /users/{{ config.user_id }}/messages/{{ record.id }}/untrash
  - required fields: id
  - risk: restores a trashed message back to its prior labels
- delete_message:
  - endpoint: DELETE /users/{{ config.user_id }}/messages/{{ record.id }}
  - required fields: id
  - risk: permanently deletes a message immediately, bypassing Trash; irreversible
- modify_thread:
  - endpoint: POST /users/{{ config.user_id }}/threads/{{ record.id }}/modify
  - required fields: id
  - risk: changes label state on every message in an existing thread
- trash_thread:
  - endpoint: POST /users/{{ config.user_id }}/threads/{{ record.id }}/trash
  - required fields: id
  - risk: moves an entire thread to Trash; auto-purged by Gmail after 30 days
- untrash_thread:
  - endpoint: POST /users/{{ config.user_id }}/threads/{{ record.id }}/untrash
  - required fields: id
  - risk: restores a trashed thread back to its prior labels
- delete_thread:
  - endpoint: DELETE /users/{{ config.user_id }}/threads/{{ record.id }}
  - required fields: id
  - risk: permanently deletes every message in a thread immediately, bypassing Trash; irreversible
- create_draft:
  - endpoint: POST /users/{{ config.user_id }}/drafts
  - required fields: message
  - risk: creates a new unsent draft, visible to the mailbox owner
- update_draft:
  - endpoint: PUT /users/{{ config.user_id }}/drafts/{{ record.id }}
  - required fields: id, message
  - risk: replaces the entire content of an existing draft
- send_draft:
  - endpoint: POST /users/{{ config.user_id }}/drafts/send
  - required fields: id
  - risk: sends a real outbound email from an existing draft on behalf of the mailbox owner; irreversible once delivered
- delete_draft:
  - endpoint: DELETE /users/{{ config.user_id }}/drafts/{{ record.id }}
  - required fields: id
  - risk: permanently deletes a draft; irreversible
- create_label:
  - endpoint: POST /users/{{ config.user_id }}/labels
  - required fields: name
  - risk: creates a new custom label visible in the mailbox owner's label list
- update_label:
  - endpoint: PUT /users/{{ config.user_id }}/labels/{{ record.id }}
  - required fields: id, name
  - risk: replaces the full definition of an existing label (name/visibility/color); a system label's name cannot actually be changed by Gmail even though the request is accepted
- patch_label:
  - endpoint: PATCH /users/{{ config.user_id }}/labels/{{ record.id }}
  - required fields: id
  - risk: partially updates an existing label's fields, leaving unset fields unchanged
- delete_label:
  - endpoint: DELETE /users/{{ config.user_id }}/labels/{{ record.id }}
  - required fields: id
  - risk: removes a user label from the account and from every message/thread that carried it; system labels reject deletion with an error
- create_filter:
  - endpoint: POST /users/{{ config.user_id }}/settings/filters
  - required fields: criteria
  - risk: creates a mail filter that automatically acts on future incoming messages matching its criteria (may auto-forward mail externally)
- delete_filter:
  - endpoint: DELETE /users/{{ config.user_id }}/settings/filters/{{ record.id }}
  - required fields: id
  - risk: removes an existing mail filter; future messages stop being auto-actioned by it
- create_send_as:
  - endpoint: POST /users/{{ config.user_id }}/settings/sendAs
  - required fields: sendAsEmail
  - risk: adds a new custom From: alias; Google emails a verification link to the new address before it can send mail
- update_send_as:
  - endpoint: PUT /users/{{ config.user_id }}/settings/sendAs/{{ record.sendAsEmail }}
  - required fields: sendAsEmail
  - risk: replaces the full send-as alias configuration, including which alias is the account default
- patch_send_as:
  - endpoint: PATCH /users/{{ config.user_id }}/settings/sendAs/{{ record.sendAsEmail }}
  - required fields: sendAsEmail
  - risk: partially updates an existing send-as alias, leaving unset fields unchanged
- delete_send_as:
  - endpoint: DELETE /users/{{ config.user_id }}/settings/sendAs/{{ record.sendAsEmail }}
  - required fields: sendAsEmail
  - risk: removes a custom From: alias (the account's primary address cannot be deleted; Gmail rejects that request)
- verify_send_as:
  - endpoint: POST /users/{{ config.user_id }}/settings/sendAs/{{ record.sendAsEmail }}/verify
  - required fields: sendAsEmail
  - risk: re-sends the verification email for a pending custom From: alias
- create_delegate:
  - endpoint: POST /users/{{ config.user_id }}/settings/delegates
  - required fields: delegateEmail
  - risk: grants another account read/send/delete access to this mailbox (Google Workspace accounts only); a significant access-control change
- delete_delegate:
  - endpoint: DELETE /users/{{ config.user_id }}/settings/delegates/{{ record.delegateEmail }}
  - required fields: delegateEmail
  - risk: revokes another account's delegated access to this mailbox
- create_forwarding_address:
  - endpoint: POST /users/{{ config.user_id }}/settings/forwardingAddresses
  - required fields: forwardingEmail
  - risk: proposes a new external forwarding address; Google emails a verification link before it can be used by update_auto_forwarding
- delete_forwarding_address:
  - endpoint: DELETE /users/{{ config.user_id }}/settings/forwardingAddresses/{{ record.forwardingEmail }}
  - required fields: forwardingEmail
  - risk: removes a forwarding address; if it is the account's current auto-forwarding target, forwarding stops
- update_auto_forwarding:
  - endpoint: PUT /users/{{ config.user_id }}/settings/autoForwarding
  - required fields: enabled
  - risk: changes the account-wide auto-forwarding singleton; when enabled, silently copies all future incoming mail to an external address
- update_vacation:
  - endpoint: PUT /users/{{ config.user_id }}/settings/vacation
  - risk: changes the account-wide vacation-responder singleton; when enabled, auto-replies to external senders with the configured message
- update_language:
  - endpoint: PUT /users/{{ config.user_id }}/settings/language
  - required fields: displayLanguage
  - risk: changes the Gmail web interface display language for the account
- update_imap:
  - endpoint: PUT /users/{{ config.user_id }}/settings/imap
  - risk: changes the account-wide IMAP-access singleton; disabling breaks any external IMAP client currently connected
- update_pop:
  - endpoint: PUT /users/{{ config.user_id }}/settings/pop
  - risk: changes the account-wide POP-access singleton, including what happens to mail after it is fetched via POP
- insert_smime_info:
  - endpoint: POST /users/{{ config.user_id }}/settings/sendAs/{{ record.sendAsEmail }}/smimeInfo
  - required fields: sendAsEmail
  - risk: high: insert_smime_info
- delete_smime_info:
  - endpoint: DELETE /users/{{ config.user_id }}/settings/sendAs/{{ record.sendAsEmail }}/smimeInfo/{{ record.id }}
  - required fields: sendAsEmail, id
  - risk: high: delete_smime_info
- set_default_smime_info:
  - endpoint: POST /users/{{ config.user_id }}/settings/sendAs/{{ record.sendAsEmail }}/smimeInfo/{{ record.id }}/setDefault
  - required fields: sendAsEmail, id
  - risk: high: set_default_smime_info
- watch_mailbox:
  - endpoint: POST /users/{{ config.user_id }}/watch
  - risk: high: watch_mailbox
- stop_mailbox_watch:
  - endpoint: POST /users/{{ config.user_id }}/stop
  - risk: high: stop_mailbox_watch

## Security

- read risk: external Gmail API read of message/thread/draft/label/history/filter/send-as/delegate/forwarding-address/profile metadata
- write risk: external Gmail API mutation, including sending real outbound email, permanently deleting messages/threads/drafts, granting mailbox delegation, and changing account-wide forwarding/vacation/IMAP/POP settings
- approval: reverse ETL plan approval required before writes; several actions (send_message, send_draft, delete_message, delete_thread, delete_draft, create_delegate, update_auto_forwarding) warrant elevated operator scrutiny -- see docs.md Write actions & risks
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Command Surface

- Gmail command surface
- Usage: pm gmail <command> [flags]
- Source CLI: Gmail API v1 (Official Google Discovery document, revision 20260803)
- Global flags:
  - --credential (string): Credential name to use for the Gmail request.
  - --json (boolean): Emit machine-readable JSON output.
  - --max-bytes (integer): Maximum direct-read response bytes.
  - --approval-token-stdin (boolean): Read the approval token as one bounded line from standard input.
- Messages
  - messages list - List Gmail messages as ETL records. [intent=etl availability=implemented stream=messages]
  - messages get - Get a message by id. [intent=direct_read availability=implemented]; risk: bounded Gmail JSON read; response is size-limited and secret-shaped fields are redacted; flags: --id (max 4096 bytes) (string): id path value: maps_to=path.id, --page (integer): Page number to return. Only for connectors whose declared pagination addresses pages by number (page_number, offset_limit); defaults to the first page., --page-cursor (string): Continuation token to return the next page. Use the next_cursor value from a previous page, for connectors whose declared pagination addresses pages by token (cursor, next_url, link_header).
  - messages attachments download - Download a message attachment to a bounded destination. [intent=binary_download availability=partial operation=gmail.messages.attachment_download]; approval: filesystem writes require an explicit destination policy; risk: bounded binary download; size-capped and written only to an explicit destination; notes: Blocked: Gmail returns attachment data in a base64url JSON envelope, and the bounded binary executor accepts only declared direct response bytes.; flags: --messageId (max 4096 bytes) (string): messageId path value: maps_to=path.messageId, --id (max 4096 bytes) (string): id path value: maps_to=path.id, --dest-root (required) (string): directory the download is written beneath; traversal outside it is refused., --file-name (string): name for the downloaded file within --dest-root; must be a single path segment., --max-bytes (integer): lower the operation's declared size cap; it can never raise it.
  - messages send - sends a real outbound email on behalf of the mailbox owner; irreversible once delivered [intent=reverse_etl availability=implemented write=send_message]; approval: reverse ETL plan -> preview -> approval -> execute; risk: sends a real outbound email on behalf of the mailbox owner; irreversible once delivered; flags: --raw (string): Entire RFC 2822 message, base64url-encoded, including headers and MIME body.: maps_to=record.raw
  - messages insert - inserts a message directly into the mailbox without sending it (no SMTP delivery, no notifications) -- still a real, visible mailbox mutation [intent=reverse_etl availability=implemented write=insert_message]; approval: reverse ETL plan -> preview -> approval -> execute; risk: inserts a message directly into the mailbox without sending it (no SMTP delivery, no notifications) -- still a real, visible mailbox mutation; flags: --raw (string): Entire RFC 2822 message, base64url-encoded.: maps_to=record.raw
  - messages import - imports a message into the mailbox from an external mail migration source, bypassing spam classification by default [intent=reverse_etl availability=implemented write=import_message]; approval: reverse ETL plan -> preview -> approval -> execute; risk: imports a message into the mailbox from an external mail migration source, bypassing spam classification by default; flags: --raw (string): Entire RFC 2822 message, base64url-encoded.: maps_to=record.raw
  - messages modify - changes label state on an existing message (e.g. moving in/out of INBOX/TRASH/UNREAD), visible to the mailbox owner [intent=reverse_etl availability=implemented write=modify_message]; approval: reverse ETL plan -> preview -> approval -> execute; risk: changes label state on an existing message (e.g. moving in/out of INBOX/TRASH/UNREAD), visible to the mailbox owner; flags: --id (string): id value: maps_to=record.id
  - messages trash - moves a message to Trash; auto-purged by Gmail after 30 days [intent=reverse_etl availability=implemented write=trash_message]; approval: reverse ETL plan -> preview -> approval -> execute; risk: moves a message to Trash; auto-purged by Gmail after 30 days; flags: --id (string): id value: maps_to=record.id
  - messages untrash - restores a trashed message back to its prior labels [intent=reverse_etl availability=implemented write=untrash_message]; approval: reverse ETL plan -> preview -> approval -> execute; risk: restores a trashed message back to its prior labels; flags: --id (string): id value: maps_to=record.id
  - messages delete - permanently deletes a message immediately, bypassing Trash; irreversible [intent=reverse_etl availability=implemented write=delete_message]; approval: reverse ETL plan -> preview -> approval -> execute; destructive confirmation required; risk: permanently deletes a message immediately, bypassing Trash; irreversible; flags: --id (string): id value: maps_to=record.id
- Threads
  - threads list - List Gmail threads as ETL records. [intent=etl availability=implemented stream=threads]
  - threads get - Get a thread by id. [intent=direct_read availability=implemented]; risk: bounded Gmail JSON read; response is size-limited and secret-shaped fields are redacted; flags: --id (max 4096 bytes) (string): id path value: maps_to=path.id, --page (integer): Page number to return. Only for connectors whose declared pagination addresses pages by number (page_number, offset_limit); defaults to the first page., --page-cursor (string): Continuation token to return the next page. Use the next_cursor value from a previous page, for connectors whose declared pagination addresses pages by token (cursor, next_url, link_header).
  - threads modify - changes label state on every message in an existing thread [intent=reverse_etl availability=implemented write=modify_thread]; approval: reverse ETL plan -> preview -> approval -> execute; risk: changes label state on every message in an existing thread; flags: --id (string): id value: maps_to=record.id
  - threads trash - moves an entire thread to Trash; auto-purged by Gmail after 30 days [intent=reverse_etl availability=implemented write=trash_thread]; approval: reverse ETL plan -> preview -> approval -> execute; risk: moves an entire thread to Trash; auto-purged by Gmail after 30 days; flags: --id (string): id value: maps_to=record.id
  - threads untrash - restores a trashed thread back to its prior labels [intent=reverse_etl availability=implemented write=untrash_thread]; approval: reverse ETL plan -> preview -> approval -> execute; risk: restores a trashed thread back to its prior labels; flags: --id (string): id value: maps_to=record.id
  - threads delete - permanently deletes every message in a thread immediately, bypassing Trash; irreversible [intent=reverse_etl availability=implemented write=delete_thread]; approval: reverse ETL plan -> preview -> approval -> execute; destructive confirmation required; risk: permanently deletes every message in a thread immediately, bypassing Trash; irreversible; flags: --id (string): id value: maps_to=record.id
- Drafts
  - drafts list - List Gmail drafts as ETL records. [intent=etl availability=implemented stream=drafts]
  - drafts get - Get a draft by id. [intent=direct_read availability=implemented]; risk: bounded Gmail JSON read; response is size-limited and secret-shaped fields are redacted; flags: --id (max 4096 bytes) (string): id path value: maps_to=path.id, --page (integer): Page number to return. Only for connectors whose declared pagination addresses pages by number (page_number, offset_limit); defaults to the first page., --page-cursor (string): Continuation token to return the next page. Use the next_cursor value from a previous page, for connectors whose declared pagination addresses pages by token (cursor, next_url, link_header).
  - drafts create - creates a new unsent draft, visible to the mailbox owner [intent=reverse_etl availability=partial write=create_draft]; approval: reverse ETL plan -> preview -> approval -> execute; risk: creates a new unsent draft, visible to the mailbox owner; notes: Required field with no scalar leaf; use the typed reverse-ETL record path for the full object. Connector command execution is metadata-only for complex object/array records.
  - drafts update - replaces the entire content of an existing draft [intent=reverse_etl availability=partial write=update_draft]; approval: reverse ETL plan -> preview -> approval -> execute; risk: replaces the entire content of an existing draft; notes: Required field with no scalar leaf; use the typed reverse-ETL record path for the full object. Connector command execution is metadata-only for complex object/array records.; flags: --id (string): id value: maps_to=record.id
  - drafts send - sends a real outbound email from an existing draft on behalf of the mailbox owner; irreversible once delivered [intent=reverse_etl availability=implemented write=send_draft]; approval: reverse ETL plan -> preview -> approval -> execute; risk: sends a real outbound email from an existing draft on behalf of the mailbox owner; irreversible once delivered; flags: --id (string): The draft ID to send (Gmail's drafts.send accepts { "id": "<draftId>" } as its full request body).: maps_to=record.id
  - drafts delete - permanently deletes a draft; irreversible [intent=reverse_etl availability=implemented write=delete_draft]; approval: reverse ETL plan -> preview -> approval -> execute; risk: permanently deletes a draft; irreversible; flags: --id (string): id value: maps_to=record.id
- Labels
  - labels list - List Gmail labels as ETL records. [intent=etl availability=implemented stream=labels]
  - labels get - Get a label by id. [intent=direct_read availability=implemented]; risk: bounded Gmail JSON read; response is size-limited and secret-shaped fields are redacted; flags: --id (max 4096 bytes) (string): id path value: maps_to=path.id, --page (integer): Page number to return. Only for connectors whose declared pagination addresses pages by number (page_number, offset_limit); defaults to the first page., --page-cursor (string): Continuation token to return the next page. Use the next_cursor value from a previous page, for connectors whose declared pagination addresses pages by token (cursor, next_url, link_header).
  - labels create - creates a new custom label visible in the mailbox owner's label list [intent=reverse_etl availability=implemented write=create_label]; approval: reverse ETL plan -> preview -> approval -> execute; risk: creates a new custom label visible in the mailbox owner's label list; flags: --name (string): name value: maps_to=record.name
  - labels update - replaces the full definition of an existing label (name/visibility/color); a system label's name cannot actually be changed by Gmail even though the request is  [intent=reverse_etl availability=implemented write=update_label]; approval: reverse ETL plan -> preview -> approval -> execute; risk: replaces the full definition of an existing label (name/visibility/color); a system label's name cannot actually be changed by Gmail even though the request is accepted; flags: --id (string): id value: maps_to=record.id, --name (string): name value: maps_to=record.name
  - labels patch - partially updates an existing label's fields, leaving unset fields unchanged [intent=reverse_etl availability=implemented write=patch_label]; approval: reverse ETL plan -> preview -> approval -> execute; risk: partially updates an existing label's fields, leaving unset fields unchanged; flags: --id (string): id value: maps_to=record.id
  - labels delete - removes a user label from the account and from every message/thread that carried it; system labels reject deletion with an error [intent=reverse_etl availability=implemented write=delete_label]; approval: reverse ETL plan -> preview -> approval -> execute; risk: removes a user label from the account and from every message/thread that carried it; system labels reject deletion with an error; flags: --id (string): id value: maps_to=record.id
- History
  - history list - List Gmail history as ETL records. [intent=etl availability=implemented stream=history]
- Filters
  - filters list - List Gmail filters as ETL records. [intent=etl availability=implemented stream=filters]
- Send As
  - send-as list - List Gmail send as as ETL records. [intent=etl availability=implemented stream=send_as]
- Delegates
  - delegates list - List Gmail delegates as ETL records. [intent=etl availability=implemented stream=delegates]
- Forwarding Addresses
  - forwarding-addresses list - List Gmail forwarding addresses as ETL records. [intent=etl availability=implemented stream=forwarding_addresses]
- Profile
  - profile list - List Gmail profile as ETL records. [intent=etl availability=implemented stream=profile]
- Settings
  - settings filters get - Get a filter by id. [intent=direct_read availability=implemented]; risk: bounded Gmail JSON read; response is size-limited and secret-shaped fields are redacted; flags: --id (max 4096 bytes) (string): id path value: maps_to=path.id, --page (integer): Page number to return. Only for connectors whose declared pagination addresses pages by number (page_number, offset_limit); defaults to the first page., --page-cursor (string): Continuation token to return the next page. Use the next_cursor value from a previous page, for connectors whose declared pagination addresses pages by token (cursor, next_url, link_header).
  - settings send-as get - Get a send-as alias. [intent=direct_read availability=implemented]; risk: bounded Gmail JSON read; response is size-limited and secret-shaped fields are redacted; flags: --sendAsEmail (max 4096 bytes) (string): sendAsEmail path value: maps_to=path.sendAsEmail, --page (integer): Page number to return. Only for connectors whose declared pagination addresses pages by number (page_number, offset_limit); defaults to the first page., --page-cursor (string): Continuation token to return the next page. Use the next_cursor value from a previous page, for connectors whose declared pagination addresses pages by token (cursor, next_url, link_header).
  - settings delegates get - Get a delegate. [intent=direct_read availability=implemented]; risk: bounded Gmail JSON read; response is size-limited and secret-shaped fields are redacted; flags: --delegateEmail (max 4096 bytes) (string): delegateEmail path value: maps_to=path.delegateEmail, --page (integer): Page number to return. Only for connectors whose declared pagination addresses pages by number (page_number, offset_limit); defaults to the first page., --page-cursor (string): Continuation token to return the next page. Use the next_cursor value from a previous page, for connectors whose declared pagination addresses pages by token (cursor, next_url, link_header).
  - settings forwarding-addresses get - Get a forwarding address. [intent=direct_read availability=implemented]; risk: bounded Gmail JSON read; response is size-limited and secret-shaped fields are redacted; flags: --forwardingEmail (max 4096 bytes) (string): forwardingEmail path value: maps_to=path.forwardingEmail, --page (integer): Page number to return. Only for connectors whose declared pagination addresses pages by number (page_number, offset_limit); defaults to the first page., --page-cursor (string): Continuation token to return the next page. Use the next_cursor value from a previous page, for connectors whose declared pagination addresses pages by token (cursor, next_url, link_header).
  - settings auto-forwarding get - Get auto-forwarding settings. [intent=direct_read availability=implemented]; risk: bounded Gmail JSON read; response is size-limited and secret-shaped fields are redacted; flags: --page (integer): Page number to return. Only for connectors whose declared pagination addresses pages by number (page_number, offset_limit); defaults to the first page., --page-cursor (string): Continuation token to return the next page. Use the next_cursor value from a previous page, for connectors whose declared pagination addresses pages by token (cursor, next_url, link_header).
  - settings vacation get - Get vacation responder settings. [intent=direct_read availability=implemented]; risk: bounded Gmail JSON read; response is size-limited and secret-shaped fields are redacted; flags: --page (integer): Page number to return. Only for connectors whose declared pagination addresses pages by number (page_number, offset_limit); defaults to the first page., --page-cursor (string): Continuation token to return the next page. Use the next_cursor value from a previous page, for connectors whose declared pagination addresses pages by token (cursor, next_url, link_header).
  - settings language get - Get display-language settings. [intent=direct_read availability=implemented]; risk: bounded Gmail JSON read; response is size-limited and secret-shaped fields are redacted; flags: --page (integer): Page number to return. Only for connectors whose declared pagination addresses pages by number (page_number, offset_limit); defaults to the first page., --page-cursor (string): Continuation token to return the next page. Use the next_cursor value from a previous page, for connectors whose declared pagination addresses pages by token (cursor, next_url, link_header).
  - settings imap get - Get IMAP settings. [intent=direct_read availability=implemented]; risk: bounded Gmail JSON read; response is size-limited and secret-shaped fields are redacted; flags: --page (integer): Page number to return. Only for connectors whose declared pagination addresses pages by number (page_number, offset_limit); defaults to the first page., --page-cursor (string): Continuation token to return the next page. Use the next_cursor value from a previous page, for connectors whose declared pagination addresses pages by token (cursor, next_url, link_header).
  - settings pop get - Get POP settings. [intent=direct_read availability=implemented]; risk: bounded Gmail JSON read; response is size-limited and secret-shaped fields are redacted; flags: --page (integer): Page number to return. Only for connectors whose declared pagination addresses pages by number (page_number, offset_limit); defaults to the first page., --page-cursor (string): Continuation token to return the next page. Use the next_cursor value from a previous page, for connectors whose declared pagination addresses pages by token (cursor, next_url, link_header).
  - settings send-as smime list - List S/MIME configs for a send-as alias. [intent=direct_read availability=implemented]; risk: bounded Gmail JSON read; response is size-limited and secret-shaped fields are redacted; flags: --sendAsEmail (max 4096 bytes) (string): sendAsEmail path value: maps_to=path.sendAsEmail, --page (integer): Page number to return. Only for connectors whose declared pagination addresses pages by number (page_number, offset_limit); defaults to the first page., --page-cursor (string): Continuation token to return the next page. Use the next_cursor value from a previous page, for connectors whose declared pagination addresses pages by token (cursor, next_url, link_header).
  - settings send-as smime get - Get an S/MIME config. [intent=direct_read availability=implemented]; risk: bounded Gmail JSON read; response is size-limited and secret-shaped fields are redacted; flags: --sendAsEmail (max 4096 bytes) (string): sendAsEmail path value: maps_to=path.sendAsEmail, --id (max 4096 bytes) (string): id path value: maps_to=path.id, --page (integer): Page number to return. Only for connectors whose declared pagination addresses pages by number (page_number, offset_limit); defaults to the first page., --page-cursor (string): Continuation token to return the next page. Use the next_cursor value from a previous page, for connectors whose declared pagination addresses pages by token (cursor, next_url, link_header).
  - settings filters create - creates a mail filter that automatically acts on future incoming messages matching its criteria (may auto-forward mail externally) [intent=reverse_etl availability=partial write=create_filter]; approval: reverse ETL plan -> preview -> approval -> execute; risk: creates a mail filter that automatically acts on future incoming messages matching its criteria (may auto-forward mail externally); notes: Required field with no scalar leaf; use the typed reverse-ETL record path for the full object. Connector command execution is metadata-only for complex object/array records.
  - settings filters delete - removes an existing mail filter; future messages stop being auto-actioned by it [intent=reverse_etl availability=implemented write=delete_filter]; approval: reverse ETL plan -> preview -> approval -> execute; risk: removes an existing mail filter; future messages stop being auto-actioned by it; flags: --id (string): id value: maps_to=record.id
  - settings send-as create - adds a new custom From: alias; Google emails a verification link to the new address before it can send mail [intent=reverse_etl availability=implemented write=create_send_as]; approval: reverse ETL plan -> preview -> approval -> execute; risk: adds a new custom From: alias; Google emails a verification link to the new address before it can send mail; flags: --sendAsEmail (string): sendAsEmail value: maps_to=record.sendAsEmail
  - settings send-as update - replaces the full send-as alias configuration, including which alias is the account default [intent=reverse_etl availability=implemented write=update_send_as]; approval: reverse ETL plan -> preview -> approval -> execute; risk: replaces the full send-as alias configuration, including which alias is the account default; flags: --sendAsEmail (string): sendAsEmail value: maps_to=record.sendAsEmail
  - settings send-as patch - partially updates an existing send-as alias, leaving unset fields unchanged [intent=reverse_etl availability=implemented write=patch_send_as]; approval: reverse ETL plan -> preview -> approval -> execute; risk: partially updates an existing send-as alias, leaving unset fields unchanged; flags: --sendAsEmail (string): sendAsEmail value: maps_to=record.sendAsEmail
  - settings send-as delete - removes a custom From: alias (the account's primary address cannot be deleted; Gmail rejects that request) [intent=reverse_etl availability=implemented write=delete_send_as]; approval: reverse ETL plan -> preview -> approval -> execute; risk: removes a custom From: alias (the account's primary address cannot be deleted; Gmail rejects that request); flags: --sendAsEmail (string): sendAsEmail value: maps_to=record.sendAsEmail
  - settings send-as verify - re-sends the verification email for a pending custom From: alias [intent=reverse_etl availability=implemented write=verify_send_as]; approval: reverse ETL plan -> preview -> approval -> execute; risk: re-sends the verification email for a pending custom From: alias; flags: --sendAsEmail (string): sendAsEmail value: maps_to=record.sendAsEmail
  - settings delegates create - grants another account read/send/delete access to this mailbox (Google Workspace accounts only); a significant access-control change [intent=reverse_etl availability=implemented write=create_delegate]; approval: reverse ETL plan -> preview -> approval -> execute; risk: grants another account read/send/delete access to this mailbox (Google Workspace accounts only); a significant access-control change; flags: --delegateEmail (string): delegateEmail value: maps_to=record.delegateEmail
  - settings delegates delete - revokes another account's delegated access to this mailbox [intent=reverse_etl availability=implemented write=delete_delegate]; approval: reverse ETL plan -> preview -> approval -> execute; risk: revokes another account's delegated access to this mailbox; flags: --delegateEmail (string): delegateEmail value: maps_to=record.delegateEmail
  - settings forwarding-addresses create - proposes a new external forwarding address; Google emails a verification link before it can be used by update_auto_forwarding [intent=reverse_etl availability=implemented write=create_forwarding_address]; approval: reverse ETL plan -> preview -> approval -> execute; risk: proposes a new external forwarding address; Google emails a verification link before it can be used by update_auto_forwarding; flags: --forwardingEmail (string): forwardingEmail value: maps_to=record.forwardingEmail
  - settings forwarding-addresses delete - removes a forwarding address; if it is the account's current auto-forwarding target, forwarding stops [intent=reverse_etl availability=implemented write=delete_forwarding_address]; approval: reverse ETL plan -> preview -> approval -> execute; risk: removes a forwarding address; if it is the account's current auto-forwarding target, forwarding stops; flags: --forwardingEmail (string): forwardingEmail value: maps_to=record.forwardingEmail
  - settings auto-forwarding update - changes the account-wide auto-forwarding singleton; when enabled, silently copies all future incoming mail to an external address [intent=reverse_etl availability=implemented write=update_auto_forwarding]; approval: reverse ETL plan -> preview -> approval -> execute; risk: changes the account-wide auto-forwarding singleton; when enabled, silently copies all future incoming mail to an external address; flags: --enabled (boolean): enabled value: maps_to=record.enabled
  - settings vacation update - changes the account-wide vacation-responder singleton; when enabled, auto-replies to external senders with the configured message [intent=reverse_etl availability=implemented write=update_vacation]; approval: reverse ETL plan -> preview -> approval -> execute; risk: changes the account-wide vacation-responder singleton; when enabled, auto-replies to external senders with the configured message
  - settings language update - changes the Gmail web interface display language for the account [intent=reverse_etl availability=implemented write=update_language]; approval: reverse ETL plan -> preview -> approval -> execute; risk: changes the Gmail web interface display language for the account; flags: --displayLanguage (string): displayLanguage value: maps_to=record.displayLanguage
  - settings imap update - changes the account-wide IMAP-access singleton; disabling breaks any external IMAP client currently connected [intent=reverse_etl availability=implemented write=update_imap]; approval: reverse ETL plan -> preview -> approval -> execute; risk: changes the account-wide IMAP-access singleton; disabling breaks any external IMAP client currently connected
  - settings pop update - changes the account-wide POP-access singleton, including what happens to mail after it is fetched via POP [intent=reverse_etl availability=implemented write=update_pop]; approval: reverse ETL plan -> preview -> approval -> execute; risk: changes the account-wide POP-access singleton, including what happens to mail after it is fetched via POP
  - settings send-as smime insert - high: insert_smime_info [intent=reverse_etl availability=implemented write=insert_smime_info]; approval: reverse ETL plan -> preview -> approval -> execute; risk: high: insert_smime_info; flags: --sendAsEmail (string): sendAsEmail value: maps_to=record.sendAsEmail
  - settings send-as smime delete - high: delete_smime_info [intent=reverse_etl availability=implemented write=delete_smime_info]; approval: reverse ETL plan -> preview -> approval -> execute; destructive confirmation required; risk: high: delete_smime_info; flags: --id (string): id value: maps_to=record.id, --sendAsEmail (string): sendAsEmail value: maps_to=record.sendAsEmail
  - settings send-as smime set-default - high: set_default_smime_info [intent=reverse_etl availability=implemented write=set_default_smime_info]; approval: reverse ETL plan -> preview -> approval -> execute; risk: high: set_default_smime_info; flags: --id (string): id value: maps_to=record.id, --sendAsEmail (string): sendAsEmail value: maps_to=record.sendAsEmail
- Watch
  - watch start - high: watch_mailbox [intent=reverse_etl availability=implemented write=watch_mailbox]; approval: reverse ETL plan -> preview -> approval -> execute; risk: high: watch_mailbox
  - watch stop - high: stop_mailbox_watch [intent=reverse_etl availability=implemented write=stop_mailbox_watch]; approval: reverse ETL plan -> preview -> approval -> execute; risk: high: stop_mailbox_watch
- Help topics:
  - gmail-auth - Gmail uses OAuth2 authorization-code credentials; never pass secrets in command text.
  - gmail-writes - Gmail mutations are typed reverse-ETL writes: plan -> preview -> approval -> execute.

## Commands

### Inspect as a manual

```bash
pm connectors inspect gmail
```

### Inspect as structured JSON

```bash
pm connectors inspect gmail --json
```

## Agent Rules

- Run pm connectors inspect gmail before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
- For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.

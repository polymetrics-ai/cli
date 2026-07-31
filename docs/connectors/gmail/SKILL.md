---
name: pm-gmail
description: Gmail connector knowledge and safe action guide.
---

# pm-gmail

## Purpose

Reads Gmail messages, threads, drafts, labels, history, filters, send-as aliases, delegates, forwarding addresses, mailbox profile, and bounded direct detail/settings resources; writes approved reverse-ETL mutations for message/thread/draft/label/filter/send-as/delegate/forwarding/S/MIME settings via the Google OAuth 2.0 refresh-token grant.

## Icon

- asset: icons/pm-sample.svg
- source: polymetrics
- review_status: polymetrics

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
- userId
- user_id
- client_id (secret)
- client_refresh_token (secret)
- client_secret (secret)

## ETL Streams

- messages:
  - primary key: id
  - fields: id(), thread_id()
- threads:
  - primary key: id
  - fields: history_id(), id(), snippet()
- drafts:
  - primary key: id
  - fields: id(), message_id(), thread_id()
- labels:
  - primary key: id
  - fields: id(), label_list_visibility(), message_list_visibility(), messages_total(), messages_unread(), name(), threads_total(), threads_unread(), type()
- history:
  - primary key: id
  - cursor: id
  - fields: id(), labels_added(), labels_removed(), messages_added(), messages_deleted()
- filters:
  - primary key: id
  - fields: action(), criteria(), id()
- send_as:
  - primary key: send_as_email
  - fields: display_name(), is_default(), is_primary(), reply_to_address(), send_as_email(), signature(), smtpMsa(), treat_as_alias(), verification_status()
- delegates:
  - primary key: delegate_email
  - fields: delegate_email(), verification_status()
- forwarding_addresses:
  - primary key: forwarding_email
  - fields: forwarding_email(), verification_status()
- profile:
  - primary key: email_address
  - fields: email_address(), history_id(), messages_total(), threads_total()

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
- batch_delete_messages:
  - endpoint: POST /users/{{ config.user_id }}/messages/batchDelete
  - required fields: ids
  - risk: permanently deletes every message ID in ids[] immediately, bypassing Trash; Gmail documents no per-ID existence guarantees
- batch_modify_messages:
  - endpoint: POST /users/{{ config.user_id }}/messages/batchModify
  - required fields: ids
  - risk: changes labels and Classification Label values across a bounded ids[] set in one Gmail batchModify request; validate the staged IDs before approval
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
- insert_smime_info:
  - endpoint: POST /users/{{ config.user_id }}/settings/sendAs/{{ record.sendAsEmail }}/smimeInfo
  - required fields: sendAsEmail, pkcs12
  - risk: uploads PKCS#12 S/MIME certificate material for a send-as alias; certificate bytes and password are redacted and require approved reverse ETL execution
- set_default_smime_info:
  - endpoint: POST /users/{{ config.user_id }}/settings/sendAs/{{ record.sendAsEmail }}/smimeInfo/{{ record.id }}/setDefault
  - required fields: sendAsEmail, id
  - risk: changes which S/MIME certificate Gmail uses by default for the send-as alias
- delete_smime_info:
  - endpoint: DELETE /users/{{ config.user_id }}/settings/sendAs/{{ record.sendAsEmail }}/smimeInfo/{{ record.id }}
  - required fields: sendAsEmail, id
  - risk: deletes an S/MIME certificate configuration from the send-as alias; future encrypted sending may fail until another cert is configured
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

## Security

- read risk: external Gmail API read of message/thread/draft/label/history/filter/send-as/delegate/forwarding-address/profile metadata plus bounded, redacted direct detail/settings reads
- write risk: external Gmail API mutation, including sending real outbound email, permanently deleting messages/threads/drafts, granting mailbox delegation, changing account-wide forwarding/vacation/IMAP/POP settings, and managing S/MIME certificate settings
- approval: reverse ETL plan approval required before writes; sending mail, permanent delete/batch delete, delegation, forwarding, and S/MIME certificate actions warrant elevated operator scrutiny -- see docs.md Write actions & risks
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Command Surface

- Inspect Gmail mailbox metadata, run bounded redacted direct reads, and use generic reverse ETL for approved Gmail mutations.
- Usage: pm gmail <command> [flags]
- Source CLI: Gmail API (Gmail API v1 Discovery revision 20260727)
- Global flags:
  - --credential (string): Credential name to use for the Gmail request.
  - --connection (string): Alias for --credential.
  - --config (string_array): Connector config override as key=value. For direct reads, config userId defaults to me.
  - --json (boolean): Emit machine-readable JSON output.
  - --limit (integer): Maximum PM ETL records to emit; does not control Gmail page size.
  - --max-bytes (integer): Maximum direct-read response bytes; operation metadata clamps Gmail binary/detail reads.
  - --preview (boolean): Preview a reverse-ETL write command without making a network mutation.
  - --approve (string): Approval token required to execute a reverse-ETL plan.
  - --confirm (string): Typed confirmation challenge for destructive reverse-ETL writes.
- Messages
  - messages list - List Gmail messages as ETL records. [intent=etl availability=implemented stream=messages]
  - messages get - Get one Gmail message by ID with message content redacted. [intent=direct_read availability=implemented]; notes: Uses the shared bounded direct-read executor and json_redacted output policy.; flags: --user-id, --id, --format
  - attachments get - Get one Gmail message attachment with base64url attachment bytes redacted. [intent=direct_read availability=implemented]; notes: Capped at 16 MiB by operation metadata; attachment data is redacted from JSON output.; flags: --user-id, --message-id, --id
- Threads
  - threads list - List Gmail threads as ETL records. [intent=etl availability=implemented stream=threads]
  - threads get - Get one Gmail thread by ID with embedded message content redacted. [intent=direct_read availability=implemented]; notes: Uses the shared bounded direct-read executor and json_redacted output policy.; flags: --user-id, --id, --format
- Drafts
  - drafts list - List Gmail drafts as ETL records. [intent=etl availability=implemented stream=drafts]
  - drafts get - Get one Gmail draft by ID with draft message content redacted. [intent=direct_read availability=implemented]; notes: Uses the shared bounded direct-read executor and json_redacted output policy.; flags: --user-id, --id, --format
- Labels
  - labels list - List Gmail labels as ETL records. [intent=etl availability=implemented stream=labels]
  - labels get - Get one Gmail label by ID. [intent=direct_read availability=implemented]; notes: Uses the shared bounded direct-read executor and json_redacted output policy.; flags: --user-id, --id
- Mailbox settings
  - settings auto-forwarding - Get the Gmail auto-forwarding singleton with forwarding email redacted. [intent=direct_read availability=implemented]; notes: Uses the shared bounded direct-read executor and json_redacted output policy.; flags: --user-id
  - settings vacation - Get the Gmail vacation responder singleton with response body redacted. [intent=direct_read availability=implemented]; notes: Uses the shared bounded direct-read executor and json_redacted output policy.; flags: --user-id
  - settings language - Get the Gmail display-language singleton. [intent=direct_read availability=implemented]; notes: Uses the shared bounded direct-read executor and json_redacted output policy.; flags: --user-id
  - settings imap - Get the Gmail IMAP settings singleton. [intent=direct_read availability=implemented]; notes: Uses the shared bounded direct-read executor and json_redacted output policy.; flags: --user-id
  - settings pop - Get the Gmail POP settings singleton. [intent=direct_read availability=implemented]; notes: Uses the shared bounded direct-read executor and json_redacted output policy.; flags: --user-id
  - filters list - List Gmail filters as ETL records. [intent=etl availability=implemented stream=filters]
  - filters get - Get one Gmail filter by ID. [intent=direct_read availability=implemented]; notes: Uses the shared bounded direct-read executor and json_redacted output policy.; flags: --user-id, --id
  - send-as list - List Gmail send-as aliases as ETL records. [intent=etl availability=implemented stream=send_as]
  - send-as get - Get one Gmail send-as alias with sensitive alias fields redacted. [intent=direct_read availability=planned]; notes: Blocked on shared direct-read path-variable safety: the current direct-read executor validates path variables with identifier-safe characters only, while the official Gmail path requires an email address value. ETL streams and typed reverse-ETL writes keep supporting these email path values through definition templates.; flags: --user-id, --send-as-email
  - delegates list - List Gmail delegates as ETL records. [intent=etl availability=implemented stream=delegates]
  - delegates get - Get one Gmail delegate with delegate email redacted. [intent=direct_read availability=planned]; notes: Blocked on shared direct-read path-variable safety: the current direct-read executor validates path variables with identifier-safe characters only, while the official Gmail path requires an email address value. ETL streams and typed reverse-ETL writes keep supporting these email path values through definition templates.; flags: --user-id, --delegate-email
  - forwarding list - List Gmail forwarding addresses as ETL records. [intent=etl availability=implemented stream=forwarding_addresses]
  - forwarding get - Get one Gmail forwarding address with address redacted. [intent=direct_read availability=planned]; notes: Blocked on shared direct-read path-variable safety: the current direct-read executor validates path variables with identifier-safe characters only, while the official Gmail path requires an email address value. ETL streams and typed reverse-ETL writes keep supporting these email path values through definition templates.; flags: --user-id, --forwarding-email
  - smime list - List S/MIME certificate configs for one send-as alias with certificate material redacted. [intent=direct_read availability=planned]; notes: Blocked on shared direct-read path-variable safety: the current direct-read executor validates path variables with identifier-safe characters only, while the official Gmail path requires an email address value. ETL streams and typed reverse-ETL writes keep supporting these email path values through definition templates.; flags: --user-id, --send-as-email
  - smime get - Get one S/MIME certificate config with certificate material redacted. [intent=direct_read availability=planned]; notes: Blocked on shared direct-read path-variable safety: the current direct-read executor validates path variables with identifier-safe characters only, while the official Gmail path requires an email address value. ETL streams and typed reverse-ETL writes keep supporting these email path values through definition templates.; flags: --user-id, --send-as-email, --id
- Mailbox history
  - history list - List Gmail mailbox history records as ETL records. [intent=etl availability=implemented stream=history]
- Mailbox profile
  - profile get - Read the Gmail mailbox profile as an ETL record. [intent=etl availability=implemented stream=profile]
- Help topics:
  - auth - Use OAuth2 refresh-token credentials from environment variables or stdin; never paste secret values.
  - writes - Use pm reverse plan/preview/run for Gmail mutations; provider commands do not bypass approval.

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

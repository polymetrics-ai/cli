# pm connectors inspect gmail

```text
NAME
  pm connectors inspect gmail - Gmail connector manual

SYNOPSIS
  pm connectors inspect gmail
  pm connectors inspect gmail --json
  pm credentials add <name> --connector gmail [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Gmail messages, threads, drafts, labels, history, filters, send-as aliases, delegates, forwarding addresses, and mailbox profile, and writes approved reverse-ETL mutations (send/insert/import/modify/trash/delete messages and threads; draft and label lifecycle; filter, send-as, delegate, and forwarding-address management; vacation/language/IMAP/POP/auto-forwarding settings) via the Google OAuth 2.0 refresh-token grant.

ICON
  asset: icons/pm-sample.svg
  source: polymetrics
  review_status: polymetrics

CAPABILITIES
  check=true catalog=true read=true write=true query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  base_url
  include_spam_and_trash
  page_size
  scopes
  start_date
  start_history_id
  token_url
  user_id
  client_id (secret)
  client_refresh_token (secret)
  client_secret (secret)

ETL STREAMS
  messages:
    primary key: id
    fields: id(), thread_id()
  threads:
    primary key: id
    fields: history_id(), id(), snippet()
  drafts:
    primary key: id
    fields: id(), message_id(), thread_id()
  labels:
    primary key: id
    fields: id(), label_list_visibility(), message_list_visibility(), messages_total(), messages_unread(), name(), threads_total(), threads_unread(), type()
  history:
    primary key: id
    cursor: id
    fields: id(), labels_added(), labels_removed(), messages_added(), messages_deleted()
  filters:
    primary key: id
    fields: action(), criteria(), id()
  send_as:
    primary key: send_as_email
    fields: display_name(), is_default(), is_primary(), reply_to_address(), send_as_email(), signature(), smtpMsa(), treat_as_alias(), verification_status()
  delegates:
    primary key: delegate_email
    fields: delegate_email(), verification_status()
  forwarding_addresses:
    primary key: forwarding_email
    fields: forwarding_email(), verification_status()
  profile:
    primary key: email_address
    fields: email_address(), history_id(), messages_total(), threads_total()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped, incremental_append, incremental_append_deduped

REVERSE ETL ACTIONS
  send_message:
    endpoint: POST /users/{{ config.user_id }}/messages/send
    risk: sends a real outbound email on behalf of the mailbox owner; irreversible once delivered
  insert_message:
    endpoint: POST /users/{{ config.user_id }}/messages
    risk: inserts a message directly into the mailbox without sending it (no SMTP delivery, no notifications) -- still a real, visible mailbox mutation
  import_message:
    endpoint: POST /users/{{ config.user_id }}/messages/import
    risk: imports a message into the mailbox from an external mail migration source, bypassing spam classification by default
  modify_message:
    endpoint: POST /users/{{ config.user_id }}/messages/{{ record.id }}/modify
    required fields: id
    risk: changes label state on an existing message (e.g. moving in/out of INBOX/TRASH/UNREAD), visible to the mailbox owner
  trash_message:
    endpoint: POST /users/{{ config.user_id }}/messages/{{ record.id }}/trash
    required fields: id
    risk: moves a message to Trash; auto-purged by Gmail after 30 days
  untrash_message:
    endpoint: POST /users/{{ config.user_id }}/messages/{{ record.id }}/untrash
    required fields: id
    risk: restores a trashed message back to its prior labels
  delete_message:
    endpoint: DELETE /users/{{ config.user_id }}/messages/{{ record.id }}
    required fields: id
    risk: permanently deletes a message immediately, bypassing Trash; irreversible
  modify_thread:
    endpoint: POST /users/{{ config.user_id }}/threads/{{ record.id }}/modify
    required fields: id
    risk: changes label state on every message in an existing thread
  trash_thread:
    endpoint: POST /users/{{ config.user_id }}/threads/{{ record.id }}/trash
    required fields: id
    risk: moves an entire thread to Trash; auto-purged by Gmail after 30 days
  untrash_thread:
    endpoint: POST /users/{{ config.user_id }}/threads/{{ record.id }}/untrash
    required fields: id
    risk: restores a trashed thread back to its prior labels
  delete_thread:
    endpoint: DELETE /users/{{ config.user_id }}/threads/{{ record.id }}
    required fields: id
    risk: permanently deletes every message in a thread immediately, bypassing Trash; irreversible
  create_draft:
    endpoint: POST /users/{{ config.user_id }}/drafts
    risk: creates a new unsent draft, visible to the mailbox owner
  update_draft:
    endpoint: PUT /users/{{ config.user_id }}/drafts/{{ record.id }}
    required fields: id
    risk: replaces the entire content of an existing draft
  send_draft:
    endpoint: POST /users/{{ config.user_id }}/drafts/send
    risk: sends a real outbound email from an existing draft on behalf of the mailbox owner; irreversible once delivered
  delete_draft:
    endpoint: DELETE /users/{{ config.user_id }}/drafts/{{ record.id }}
    required fields: id
    risk: permanently deletes a draft; irreversible
  create_label:
    endpoint: POST /users/{{ config.user_id }}/labels
    risk: creates a new custom label visible in the mailbox owner's label list
  update_label:
    endpoint: PUT /users/{{ config.user_id }}/labels/{{ record.id }}
    required fields: id
    risk: replaces the full definition of an existing label (name/visibility/color); a system label's name cannot actually be changed by Gmail even though the request is accepted
  patch_label:
    endpoint: PATCH /users/{{ config.user_id }}/labels/{{ record.id }}
    required fields: id
    risk: partially updates an existing label's fields, leaving unset fields unchanged
  delete_label:
    endpoint: DELETE /users/{{ config.user_id }}/labels/{{ record.id }}
    required fields: id
    risk: removes a user label from the account and from every message/thread that carried it; system labels reject deletion with an error
  create_filter:
    endpoint: POST /users/{{ config.user_id }}/settings/filters
    risk: creates a mail filter that automatically acts on future incoming messages matching its criteria (may auto-forward mail externally)
  delete_filter:
    endpoint: DELETE /users/{{ config.user_id }}/settings/filters/{{ record.id }}
    required fields: id
    risk: removes an existing mail filter; future messages stop being auto-actioned by it
  create_send_as:
    endpoint: POST /users/{{ config.user_id }}/settings/sendAs
    risk: adds a new custom From: alias; Google emails a verification link to the new address before it can send mail
  update_send_as:
    endpoint: PUT /users/{{ config.user_id }}/settings/sendAs/{{ record.sendAsEmail }}
    required fields: sendAsEmail
    risk: replaces the full send-as alias configuration, including which alias is the account default
  patch_send_as:
    endpoint: PATCH /users/{{ config.user_id }}/settings/sendAs/{{ record.sendAsEmail }}
    required fields: sendAsEmail
    risk: partially updates an existing send-as alias, leaving unset fields unchanged
  delete_send_as:
    endpoint: DELETE /users/{{ config.user_id }}/settings/sendAs/{{ record.sendAsEmail }}
    required fields: sendAsEmail
    risk: removes a custom From: alias (the account's primary address cannot be deleted; Gmail rejects that request)
  verify_send_as:
    endpoint: POST /users/{{ config.user_id }}/settings/sendAs/{{ record.sendAsEmail }}/verify
    required fields: sendAsEmail
    risk: re-sends the verification email for a pending custom From: alias
  create_delegate:
    endpoint: POST /users/{{ config.user_id }}/settings/delegates
    risk: grants another account read/send/delete access to this mailbox (Google Workspace accounts only); a significant access-control change
  delete_delegate:
    endpoint: DELETE /users/{{ config.user_id }}/settings/delegates/{{ record.delegateEmail }}
    required fields: delegateEmail
    risk: revokes another account's delegated access to this mailbox
  create_forwarding_address:
    endpoint: POST /users/{{ config.user_id }}/settings/forwardingAddresses
    risk: proposes a new external forwarding address; Google emails a verification link before it can be used by update_auto_forwarding
  delete_forwarding_address:
    endpoint: DELETE /users/{{ config.user_id }}/settings/forwardingAddresses/{{ record.forwardingEmail }}
    required fields: forwardingEmail
    risk: removes a forwarding address; if it is the account's current auto-forwarding target, forwarding stops
  update_auto_forwarding:
    endpoint: PUT /users/{{ config.user_id }}/settings/autoForwarding
    risk: changes the account-wide auto-forwarding singleton; when enabled, silently copies all future incoming mail to an external address
  update_vacation:
    endpoint: PUT /users/{{ config.user_id }}/settings/vacation
    risk: changes the account-wide vacation-responder singleton; when enabled, auto-replies to external senders with the configured message
  update_language:
    endpoint: PUT /users/{{ config.user_id }}/settings/language
    risk: changes the Gmail web interface display language for the account
  update_imap:
    endpoint: PUT /users/{{ config.user_id }}/settings/imap
    risk: changes the account-wide IMAP-access singleton; disabling breaks any external IMAP client currently connected
  update_pop:
    endpoint: PUT /users/{{ config.user_id }}/settings/pop
    risk: changes the account-wide POP-access singleton, including what happens to mail after it is fetched via POP

SECURITY
  read risk: external Gmail API read of message/thread/draft/label/history/filter/send-as/delegate/forwarding-address/profile metadata
  write risk: external Gmail API mutation, including sending real outbound email, permanently deleting messages/threads/drafts, granting mailbox delegation, and changing account-wide forwarding/vacation/IMAP/POP settings
  approval: reverse ETL plan approval required before writes; several actions (send_message, send_draft, delete_message, delete_thread, delete_draft, create_delegate, update_auto_forwarding) warrant elevated operator scrutiny -- see docs.md Write actions & risks
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect gmail

  # Inspect as structured JSON
  pm connectors inspect gmail --json

AGENT WORKFLOW
  - Run pm connectors inspect gmail before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.
  - For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```

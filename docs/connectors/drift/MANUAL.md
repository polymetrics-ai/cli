# pm connectors inspect drift

```text
NAME
  pm connectors inspect drift - Drift connector manual

SYNOPSIS
  pm connectors inspect drift
  pm connectors inspect drift --json
  pm credentials add <name> --connector drift [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Drift users, accounts, conversations, contacts, and teams, and writes contact/account/message/conversation/timeline-event/GDPR mutations through the Drift REST API.

ICON
  asset: icons/drift.svg
  source: upstream_registry
  review_status: upstream_seeded
  review_url: https://devdocs.drift.com/

CAPABILITIES
  check=true catalog=true read=true write=true query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  base_url
  email
  access_token (secret)

ETL STREAMS
  users:
    primary key: id
    cursor: updatedAt
    fields: alias(), availability(), avatarUrl(), bot(), createdAt(), email(), id(), locale(), name(), orgId(), phone(), role(), timeZone(), updatedAt(), verified()
  accounts:
    primary key: account_id
    cursor: updateDateTime
    fields: account_id(), createDateTime(), customProperties(), deleted(), domain(), name(), ownerId(), targeted(), updateDateTime()
  conversations:
    primary key: id
    cursor: updatedAt
    fields: contactId(), conversationTags(), createdAt(), id(), inboxId(), orgId(), participants(), relatedPlaybookId(), status(), updatedAt()
  contacts:
    primary key: id
    cursor: updatedAt
    fields: attributes(), createdAt(), id(), updatedAt()
  teams:
    primary key: id
    cursor: updatedAt
    fields: autoOffline(), id(), main(), members(), name(), orgId(), owner(), responseTimerEnabled(), status(), teamAvailabilityMode(), updatedAt(), workspaceId()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

REVERSE ETL ACTIONS
  create_contact:
    endpoint: POST /contacts
    risk: creates a new Drift contact record; low-risk external mutation, no approval required
  update_contact:
    endpoint: PATCH /contacts/{{ record.id }}
    required fields: id
    risk: mutates an existing Drift contact's attributes, including standard fields (email/name/phone) and any custom attribute; external mutation, approval required
  delete_contact:
    endpoint: DELETE /contacts/{{ record.id }}
    required fields: id
    risk: permanently removes a Drift contact and its conversation history association; destructive, approval required
  post_timeline_event:
    endpoint: POST /contacts/timeline
    risk: posts a custom timeline event onto a contact's record; low-risk external mutation, no approval required
  create_account:
    endpoint: POST /accounts/create
    risk: creates a new Drift account (company) record; low-risk external mutation, no approval required
  update_account:
    endpoint: PATCH /accounts/update
    risk: mutates an existing Drift account's owner/name/domain/targeting/custom properties; external mutation, approval required
  delete_account:
    endpoint: DELETE /accounts/{{ record.account_id }}
    required fields: account_id
    risk: permanently removes a Drift account record; destructive, approval required
  create_message:
    endpoint: POST /conversations/{{ record.conversation_id }}/messages
    required fields: conversation_id
    risk: posts a message into a live Drift conversation, visible to the end customer when type is chat; external mutation, approval required
  create_conversation:
    endpoint: POST /conversations/new
    risk: starts a new Drift conversation for the given contact email; external mutation, approval required
  gdpr_retrieve:
    endpoint: POST /gdpr/retrieve
    risk: triggers Drift to compile and email all data held for the given email address to the account's admin; a data-subject-access-request action, approval required
  gdpr_delete:
    endpoint: POST /gdpr/delete
    risk: permanently erases every contact/user record matching the given email address from Drift; irreversible data-subject-erasure action, approval required

SECURITY
  read risk: external Drift API read of conversational-marketing users, accounts, conversations, contacts, and teams
  write risk: external Drift API mutation of contacts, accounts, conversations, messages, timeline events, and GDPR data-subject requests; delete_contact/delete_account/gdpr_delete are destructive and require approval
  approval: required for delete_contact, delete_account, and gdpr_delete; other create/update writes are lower-risk marketing/support-data mutations
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect drift

  # Inspect as structured JSON
  pm connectors inspect drift --json

AGENT WORKFLOW
  - Run pm connectors inspect drift before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.
  - For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```

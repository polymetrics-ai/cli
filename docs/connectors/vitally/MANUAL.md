# pm connectors inspect vitally

```text
NAME
  pm connectors inspect vitally - Vitally connector manual

SYNOPSIS
  pm connectors inspect vitally
  pm connectors inspect vitally --json
  pm credentials add <name> --connector vitally [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads and writes Vitally customer-success accounts, users, notes, conversations, tasks, and NPS responses via the Vitally REST API.

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
  page_size
  status
  basic_auth_header (secret)

ETL STREAMS
  accounts:
    primary key: id
    fields: id(), name(), traits()
  users:
    primary key: id
    cursor: updatedAt
    fields: accounts(), avatar(), createdAt(), deactivatedAt(), email(), externalId(), firstKnown(), id(), lastInboundMessageTimestamp(), lastOutboundMessageTimestamp(), lastSeenTimestamp(), name(), npsLastFeedback(), npsLastRespondedAt(), npsLastScore(), organizations(), segments(), traits(), unsubscribedFromConversations(), unsubscribedFromConversationsAt(), updatedAt()
  notes:
    primary key: id
    cursor: updated_at
    fields: account_id(), archived_at(), author_id(), category_id(), created_at(), external_id(), id(), note(), note_date(), organization_id(), source(), subject(), tags(), traits(), updated_at()
  conversations:
    primary key: id
    cursor: updated_at
    fields: accounts(), admins(), created_at(), external_id(), id(), rating(), source(), status(), subject(), traits(), updated_at(), users()
  tasks:
    primary key: id
    cursor: updated_at
    fields: account_id(), archived_at(), assigned_to_id(), category_id(), completed_at(), completed_by_id(), created_at(), description(), due_date(), external_id(), id(), meeting_id(), name(), organization_id(), source(), tags(), traits(), updated_at()
  nps_responses:
    primary key: id
    cursor: updated_at
    fields: created_at(), external_id(), feedback(), id(), responded_at(), score(), updated_at(), user_id()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

REVERSE ETL ACTIONS
  create_account:
    endpoint: POST /resources/accounts
    risk: creates a new customer-success account visible to the vendor's CS team; external mutation, approval required
  update_account:
    endpoint: PUT /resources/accounts/{{ record.id }}
    required fields: id
    risk: updates an existing customer-success account's fields/traits, visible to the vendor's CS team; external mutation, approval required
  create_user:
    endpoint: POST /resources/users
    risk: creates a new user record visible to the vendor's CS team; external mutation, approval required
  update_user:
    endpoint: PUT /resources/users/{{ record.id }}
    required fields: id
    risk: updates an existing user's fields/traits, visible to the vendor's CS team; external mutation, approval required
  create_note:
    endpoint: POST /resources/notes
    risk: creates a customer-success note visible to the vendor's CS team; external mutation, approval required
  update_note:
    endpoint: PUT /resources/notes/{{ record.id }}
    required fields: id
    risk: updates an existing customer-success note visible to the vendor's CS team; external mutation, approval required
  delete_note:
    endpoint: DELETE /resources/notes/{{ record.id }}
    required fields: id
    risk: archives/deletes a customer-success note; external mutation, approval required
  create_conversation:
    endpoint: POST /resources/conversations
    risk: creates a historical conversation record visible to the vendor's CS team; does not send outbound messages to real participants (Vitally's own documented behavior); external mutation, approval required
  update_conversation:
    endpoint: PUT /resources/conversations/{{ record.id }}
    required fields: id
    risk: updates an existing conversation record (new messages inserted, existing ones updated by externalId); external mutation, approval required
  delete_conversation:
    endpoint: DELETE /resources/conversations/{{ record.id }}
    required fields: id
    risk: permanently deletes a conversation and all its messages; external mutation, approval required
  create_task:
    endpoint: POST /resources/tasks
    risk: creates a customer-success task visible to the vendor's CS team; external mutation, approval required
  update_task:
    endpoint: PUT /resources/tasks/{{ record.id }}
    required fields: id
    risk: updates an existing customer-success task visible to the vendor's CS team; external mutation, approval required
  create_nps_response:
    endpoint: POST /resources/npsResponses
    risk: creates (or, if externalId already exists, upserts -- Vitally's own documented behavior) an NPS response visible to the vendor's CS team; external mutation, approval required
  update_nps_response:
    endpoint: PUT /resources/npsResponses/{{ record.id }}
    required fields: id
    risk: updates an existing NPS response visible to the vendor's CS team; external mutation, approval required

SECURITY
  read risk: external Vitally API read of customer-success account/user/note/conversation/task/NPS-response data
  write risk: external mutation of Vitally customer-success records (create/update accounts, users, notes, tasks, conversations, NPS responses; delete notes and conversations); approval required
  approval: read: none, read-only sync surface. write: required for all mutating actions (create/update/delete of customer-success records visible to the vendor's CS team).
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect vitally

  # Inspect as structured JSON
  pm connectors inspect vitally --json

AGENT WORKFLOW
  - Run pm connectors inspect vitally before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.
  - For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```

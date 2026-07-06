---
name: pm-boldsign
description: BoldSign connector knowledge and safe action guide.
---

# pm-boldsign

## Purpose

Reads BoldSign documents, templates, teams, contacts, brands, users, contact groups, and sender identities, and writes team/contact-group/document-lifecycle/user-lifecycle mutations, through the BoldSign REST API.

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
- mode
- api_key (secret)

## ETL Streams

- documents:
  - primary key: document_id
  - cursor: created_date
  - fields: created_date(), document_id(), enable_signing_order(), expiry_date(), is_deleted(), labels(), message_title(), sender_detail(), sender_email(), signer_details(), status()
- templates:
  - primary key: document_id
  - cursor: created_date
  - fields: created_date(), document_id(), is_shared_template(), labels(), sender_email(), template_description(), template_name()
- teams:
  - primary key: team_id
  - cursor: created_date
  - fields: created_date(), team_id(), team_name(), users()
- contacts:
  - primary key: id
  - fields: company_name(), email(), id(), name(), phone_number()
- brands:
  - primary key: brand_id
  - fields: background_color(), brand_id(), brand_name(), button_color(), is_default()
- users:
  - primary key: user_id
  - cursor: created_date
  - fields: created_date(), email(), first_name(), last_name(), meta_data(), modified_date(), role(), team_id(), team_name(), user_id(), user_status()
- contact_groups:
  - primary key: group_id
  - fields: contacts(), directories(), group_id(), group_name()
- sender_identities:
  - primary key: id
  - fields: approved_date(), brand_id(), created_by(), email(), id(), meta_data(), name(), notification_settings(), redirect_url(), status()

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

## Reverse ETL Actions

- create_team:
  - endpoint: POST /v1/teams/create
  - risk: external mutation; creates a new BoldSign team; approval required
- update_team:
  - endpoint: PUT /v1/teams/update
  - risk: external mutation; renames an existing BoldSign team; approval required
- update_contact:
  - endpoint: PUT /v1/contacts/update?id={{ record.id }}
  - required fields: id
  - risk: external mutation; overwrites an existing BoldSign contact's details; approval required
- delete_contact:
  - endpoint: DELETE /v1/contacts/delete?id={{ record.id }}
  - required fields: id
  - risk: destructive external mutation; permanently deletes a BoldSign contact; approval required
- create_contact_group:
  - endpoint: POST /v1/contactGroups/create
  - risk: external mutation; creates a new BoldSign contact group; approval required
- update_contact_group:
  - endpoint: PUT /v1/contactGroups/update?groupId={{ record.groupId }}
  - required fields: groupId
  - risk: external mutation; overwrites an existing BoldSign contact group's members/name; approval required
- delete_contact_group:
  - endpoint: DELETE /v1/contactGroups/delete?groupId={{ record.groupId }}
  - required fields: groupId
  - risk: destructive external mutation; permanently deletes a BoldSign contact group; approval required
- revoke_document:
  - endpoint: POST /v1/document/revoke?documentId={{ record.documentId }}
  - required fields: documentId
  - risk: destructive external mutation; revokes a BoldSign document, permanently ending its signature request; approval required
- remind_document:
  - endpoint: POST /v1/document/remind?documentId={{ record.documentId }}
  - required fields: documentId
  - risk: external mutation; sends an email/SMS reminder to a document's pending signers; approval required
- delete_document:
  - endpoint: DELETE /v1/document/delete?documentId={{ record.documentId }}&deletePermanently={{ record.deletePermanently }}
  - required fields: documentId, deletePermanently
  - risk: destructive external mutation; moves a BoldSign document to trash (or permanently deletes it when deletePermanently=true); approval required
- add_document_tags:
  - endpoint: PATCH /v1/document/addTags
  - risk: external mutation; adds label tags to a BoldSign document; approval required
- delete_document_tags:
  - endpoint: DELETE /v1/document/deleteTags
  - risk: external mutation; removes label tags from a BoldSign document; approval required
- update_user:
  - endpoint: PUT /v1/users/update
  - risk: external mutation; changes a BoldSign user's role or active/deactivated status; approval required
- change_user_team:
  - endpoint: PUT /v1/users/changeTeam?userId={{ record.userId }}
  - required fields: userId
  - risk: external mutation; moves a BoldSign user to a different team; approval required

## Security

- read risk: external BoldSign API read of documents, templates, teams, contacts, brands, users, contact groups, and sender identities
- write risk: external mutation of BoldSign teams, contacts, contact groups, document lifecycle state (revoke/remind/delete/tags), and user role/team/status; includes 2 destructive (irreversible-effect) actions (delete_contact, delete_contact_group, delete_document, revoke_document)
- approval: required for every write action; read remains unapproved
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect boldsign
```

### Inspect as structured JSON

```bash
pm connectors inspect boldsign --json
```

## Agent Rules

- Run pm connectors inspect boldsign before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
- For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.

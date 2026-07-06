---
name: pm-dropbox-sign
description: Dropbox Sign connector knowledge and safe action guide.
---

# pm-dropbox-sign

## Purpose

Reads Dropbox Sign (HelloSign) signature requests, templates, team members, and account details, and writes signature-request/template/team/account lifecycle mutations, through the Dropbox Sign REST API.

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

- signature_requests:
  - primary key: signature_request_id
  - cursor: created_at
  - fields: created_at(), has_error(), is_complete(), is_declined(), message(), requester_email_address(), signature_request_id(), subject(), test_mode(), title()
- templates:
  - primary key: template_id
  - cursor: updated_at
  - fields: is_creator(), is_embedded(), is_locked(), message(), template_id(), title(), updated_at()
- team_members:
  - primary key: account_id
  - fields: account_id(), email_address(), role()
- account:
  - primary key: account_id
  - fields: account_id(), email_address(), is_paid_hf(), is_paid_hs(), locale(), role_code()

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

## Reverse ETL Actions

- update_signature_request:
  - endpoint: POST /signature_request/update/{{ record.signature_request_id }}
  - required fields: signature_request_id
  - risk: external mutation; changes a signer's email address or name on an in-progress signature request, redirecting where the next request/reminder is delivered; approval required
- cancel_signature_request:
  - endpoint: POST /signature_request/cancel/{{ record.signature_request_id }}
  - required fields: signature_request_id
  - risk: destructive external mutation; cancels an incomplete signature request, this action is not reversible; approval required
- remind_signature_request:
  - endpoint: POST /signature_request/remind/{{ record.signature_request_id }}
  - required fields: signature_request_id
  - risk: external mutation; sends an email reminder to a signer; cannot be sent again within 1 hour of the last reminder (manual or automatic)
- release_hold_signature_request:
  - endpoint: POST /signature_request/release_hold/{{ record.signature_request_id }}
  - required fields: signature_request_id
  - risk: external mutation; releases a held signature request created from an UnclaimedDraft, immediately sending requests to all signers; approval required
- remove_signature_request:
  - endpoint: POST /signature_request/remove/{{ record.signature_request_id }}
  - required fields: signature_request_id
  - risk: destructive external mutation; removes the caller's access to a completed signature request from the account's list view, this action is not reversible; approval required
- delete_template:
  - endpoint: POST /template/delete/{{ record.template_id }}
  - required fields: template_id
  - risk: destructive external mutation; completely deletes a template from the account, this action is not reversible; approval required
- add_template_user:
  - endpoint: POST /template/add_user/{{ record.template_id }}
  - required fields: template_id
  - risk: external mutation; grants the specified account (which must already be a Team member) access to a template
- remove_template_user:
  - endpoint: POST /template/remove_user/{{ record.template_id }}
  - required fields: template_id
  - risk: external mutation; revokes the specified account's access to a template
- create_team:
  - endpoint: POST /team/create
  - risk: external mutation; creates a new Team and makes the calling account its member; fails if the caller already belongs to a Team
- update_team:
  - endpoint: PUT /team
  - risk: external mutation; renames the caller's own Team
- add_team_member:
  - endpoint: PUT /team/add_member
  - risk: external mutation; invites or moves a user onto the caller's Team, creating a new Dropbox Sign account for the invited email if one does not already exist
- remove_team_member:
  - endpoint: POST /team/remove_member
  - risk: destructive external mutation; removes a user from the caller's Team; optionally transfers the removed account's documents to another account (Enterprise plans only), which is not reversible; approval required
- update_account:
  - endpoint: PUT /account
  - risk: external mutation; updates the caller's account settings (currently limited to the event callback URL and locale)

## Security

- read risk: external Dropbox Sign API read of signature requests, templates, team members, and account data
- write risk: external mutation of signature requests (update/cancel/remind/release_hold/remove), templates (delete/add_user/remove_user), teams (create/update/add_member/remove_member), and account settings; several actions are destructive/not reversible (cancel_signature_request, remove_signature_request, delete_template, remove_team_member) and require approval
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect dropbox-sign
```

### Inspect as structured JSON

```bash
pm connectors inspect dropbox-sign --json
```

## Agent Rules

- Run pm connectors inspect dropbox-sign before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
- For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.

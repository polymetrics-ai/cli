---
name: pm-concord
description: Concord connector knowledge and safe action guide.
---

# pm-concord

## Purpose

Reads and writes Concord contract lifecycle management data: agreements (and their metadata/summary/comments/activities/members/versions/attachments sub-resources), organizations, folders, reports, tags, clauses, approvals, groups, members, events, subscription, branding, and automated templates through the Concord REST API.

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

- agreement_uid
- approval_id
- base_url
- clause_id
- clauses_page_size
- events_end_date
- events_start_date
- folder_id
- mode
- organization_id
- page_size
- report_id
- api_key (secret)

## ETL Streams

- agreements:
  - primary key: uid
  - fields: createdAt(), organizationId(), stage(), status(), title(), uid(), updatedAt()
- user_organizations:
  - primary key: id
  - fields: id(), name(), role(), type()
- folders:
  - primary key: id
  - fields: id(), name(), organizationId(), parentId()
- reports:
  - primary key: id
  - fields: id(), name(), organizationId(), type()
- tags:
  - primary key: id
  - fields: color(), id(), name()
- organization:
  - primary key: id
  - fields: aiEnabled(), allTagsVisible(), askForTags(), canCollaboratorSign(), createdAt(), deleted(), description(), emailDomains(), id(), logo(), name(), parent(), region(), subscription(), subsidiaries()
- folder:
  - primary key: id
  - fields: access(), createdAt(), createdBy(), id(), isBookmarked(), modifiedAt(), name(), parentId()
- folder_agreements:
  - primary key: uuid
  - fields: createdAt(), folderId(), modifiedAt(), organizationId(), status(), title(), uuid()
- report:
  - primary key: id
  - fields: description(), filters(), id(), lastUpdatedAt(), name()
- clauses:
  - primary key: id
  - fields: createdAt(), description(), id(), numberOfTemplatesLinked(), presignedUrl(), title(), version()
- clause:
  - primary key: id
  - fields: createdAt(), description(), id(), numberOfTemplatesLinked(), presignedUrl(), title(), version()
- approvals:
  - primary key: id
  - fields: blockThirdPartySignature(), deletable(), description(), id(), rules(), title()
- approval:
  - primary key: id
  - fields: blockThirdPartySignature(), deletable(), description(), id(), rules(), title()
- groups:
  - primary key: id
  - fields: description(), id(), invitations(), name(), organization(), users()
- members:
  - primary key: userOrganizationId
  - fields: createdAt(), groups(), invitation(), isActive(), job(), organization(), role(), type(), user(), userOrganizationId()
- events:
  - primary key: id
  - fields: actor(), createdAt(), event(), id(), type()
- subscription:
  - primary key: subscriptionId
  - fields: customerId(), featureLevel(), seats(), status(), subscriptionId(), subscriptionName(), type()
- branding:
  - primary key: useForInternalEmails
  - fields: customAgreementView(), customEmailContent(), customEmailSender(), useForInternalEmails()
- automated_templates:
  - primary key: id
  - fields: id(), name(), salesforceReady()
- user_me:
  - primary key: id
  - fields: createdAt(), currentOrganizationId(), email(), fullName(), hasAcceptedTerms(), hasPassword(), hasPicture(), id(), timezone()
- user_preferences:
  - primary key: name
  - fields: dateFormat(), deadlinesNotificationDays(), deadlinesNotificationEnabled(), language(), mobile(), mobileCode(), name(), phone(), phoneCode()
- webhooks_integrations:
  - primary key: id
  - fields: events(), id(), isActive(), url()
- agreement:
  - primary key: uid
  - fields: creation(), folderId(), lastPublicVersion(), lock(), metadata(), permission(), uid()
- agreement_metadata:
  - primary key: agreement_uid
  - fields: agreement_uid(), bookmarked(), description(), inboxed(), lastAccessAt(), organization(), read(), status(), tags(), title(), trashed()
- agreement_summary:
  - primary key: agreementUid
  - fields: agreementCategory(), agreementUid(), clauses(), description(), documentType(), endclauses(), lifecycle(), organizationId(), signedwithlabels(), totalAgreementValue()
- agreement_comments:
  - primary key: comment_uuid
  - fields: agreement_id(), comment_uuid(), commentedText(), createdAt(), createdBy(), reply(), resolved(), text(), uuid(), version(), visibility()
- agreement_activities:
  - primary key: id
  - fields: action(), agreement_id(), createdAt(), id(), organization(), params(), status(), userOrganization(), visibility()
- agreement_members:
  - primary key: agreement_id, member_id
  - fields: agreement_id(), lastAccessAt(), member_id(), permission(), relation(), status(), user(), userSignStatus()
- agreement_versions:
  - primary key: id
  - fields: agreement_id(), comment(), date(), displayVersion(), id(), organization(), type(), user(), version(), visibility()
- agreement_attachments:
  - primary key: id
  - fields: agreement_id(), contentType(), id(), name(), size()

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

## Reverse ETL Actions

- create_folder:
  - endpoint: POST /organizations/{{ config.organization_id }}/folders
  - risk: creates a new Concord folder within the configured organization; low risk, no data destruction
- update_folder:
  - endpoint: PUT /organizations/{{ config.organization_id }}/folders/{{ record.id }}
  - required fields: id
  - risk: renames/moves an existing Concord folder; may change document organization visible to other users
- delete_folder:
  - endpoint: DELETE /organizations/{{ config.organization_id }}/folders/{{ record.id }}
  - required fields: id
  - risk: permanently deletes a Concord folder; destructive, external mutation; approval required
- create_report:
  - endpoint: POST /organizations/{{ config.organization_id }}/reports
  - risk: creates a new saved Concord report within the configured organization; low risk
- update_report:
  - endpoint: PUT /organizations/{{ config.organization_id }}/reports/{{ record.id }}
  - required fields: id
  - risk: replaces an existing Concord saved report's definition; may change what other users see when they run it
- delete_report:
  - endpoint: DELETE /organizations/{{ config.organization_id }}/reports/{{ record.id }}
  - required fields: id
  - risk: permanently deletes a Concord saved report; destructive, external mutation; approval required
- create_clause:
  - endpoint: POST /organizations/{{ config.organization_id }}/clauses
  - risk: creates a new reusable Concord clause template within the configured organization; low risk
- update_clause:
  - endpoint: PUT /organizations/{{ config.organization_id }}/clauses/{{ record.id }}
  - required fields: id
  - risk: updates an existing Concord clause template; may affect future agreements linked to this clause
- delete_clause:
  - endpoint: DELETE /organizations/{{ config.organization_id }}/clauses/{{ record.id }}
  - required fields: id
  - risk: permanently deletes a Concord clause template; destructive, external mutation; approval required
- create_group:
  - endpoint: POST /organizations/{{ config.organization_id }}/groups
  - risk: creates a new Concord user group within the configured organization; low risk
- create_approval:
  - endpoint: POST /organizations/{{ config.organization_id }}/approvals
  - risk: creates a new Concord company approval workflow within the configured organization; affects future agreement signature routing
- update_approval:
  - endpoint: POST /organizations/{{ config.organization_id }}/approvals/{{ record.id }}
  - required fields: id
  - risk: replaces an existing Concord company approval workflow; affects agreements already routed through it
- delete_approval:
  - endpoint: DELETE /organizations/{{ config.organization_id }}/approvals/{{ record.id }}
  - required fields: id
  - risk: permanently deletes a Concord company approval workflow; destructive, external mutation; approval required

## Security

- read risk: external Concord API read of contract lifecycle management data (agreements and sub-resources, organizations, folders, reports, tags, clauses, approvals, groups, members, events, subscription, branding)
- write risk: external mutation of Concord folders, reports, clauses, groups, and company approval workflows (create/update/delete); does not create, sign, or modify agreements themselves
- approval: required for delete_folder/delete_report/delete_clause/delete_approval (destructive); create/update actions are lower risk but still mutate shared organization configuration
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect concord
```

### Inspect as structured JSON

```bash
pm connectors inspect concord --json
```

## Agent Rules

- Run pm connectors inspect concord before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
- For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.

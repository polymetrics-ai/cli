---
name: pm-concord
description: Concord connector knowledge and safe action guide.
---

# pm-concord

## Purpose

Reads and writes Concord contract lifecycle management data: agreements (and their metadata/summary/comments/activities/members/versions/attachments sub-resources), organizations, folders, reports, tags, clauses, approvals, groups, members, events, subscription, branding, and automated templates through the Concord REST API.

## Icon

- id: pm-sample
- asset: icons/pm-sample.svg
- source: polymetrics
- review_status: polymetrics
- review_url: https://github.com/polymetrics-ai/cli

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
- api_key (secret) (required)

## ETL Streams

- agreements:
  - primary key: uid
  - fields: createdAt(string), organizationId(integer), stage(string), status(string), title(string), uid(string), updatedAt(string)
- user_organizations:
  - primary key: id
  - fields: id(integer), name(string), role(string), type(string)
- folders:
  - primary key: id
  - fields: id(integer), name(string), organizationId(integer), parentId(integer)
- reports:
  - primary key: id
  - fields: id(integer), name(string), organizationId(integer), type(string)
- tags:
  - primary key: id
  - fields: color(string), id(integer), name(string)
- organization:
  - primary key: id
  - fields: aiEnabled(boolean), allTagsVisible(boolean), askForTags(boolean), canCollaboratorSign(boolean), createdAt(integer), deleted(boolean), description(string), emailDomains(array), id(integer), logo(string), name(string), parent(object), region(string), subscription(object), subsidiaries(array)
- folder:
  - primary key: id
  - fields: access(object), createdAt(integer), createdBy(object), id(integer), isBookmarked(boolean), modifiedAt(integer), name(string), parentId(integer)
- folder_agreements:
  - primary key: uuid
  - fields: createdAt(integer), folderId(integer), modifiedAt(integer), organizationId(integer), status(string), title(string), uuid(string)
- report:
  - primary key: id
  - fields: description(string), filters(object), id(string), lastUpdatedAt(integer), name(string)
- clauses:
  - primary key: id
  - fields: createdAt(integer), description(string), id(integer), numberOfTemplatesLinked(integer), presignedUrl(string), title(string), version(integer)
- clause:
  - primary key: id
  - fields: createdAt(integer), description(string), id(integer), numberOfTemplatesLinked(integer), presignedUrl(string), title(string), version(integer)
- approvals:
  - primary key: id
  - fields: blockThirdPartySignature(boolean), deletable(boolean), description(string), id(integer), rules(array), title(string)
- approval:
  - primary key: id
  - fields: blockThirdPartySignature(boolean), deletable(boolean), description(string), id(integer), rules(array), title(string)
- groups:
  - primary key: id
  - fields: description(string), id(integer), invitations(array), name(string), organization(object), users(array)
- members:
  - primary key: userOrganizationId
  - fields: createdAt(integer), groups(array), invitation(object), isActive(boolean), job(string), organization(object), role(object), type(string), user(object), userOrganizationId(integer)
- events:
  - primary key: id
  - fields: actor(object), createdAt(integer), event(object), id(integer), type(string)
- subscription:
  - primary key: subscriptionId
  - fields: customerId(string), featureLevel(string), seats(array), status(string), subscriptionId(string), subscriptionName(string), type(string)
- branding:
  - primary key: useForInternalEmails
  - fields: customAgreementView(object), customEmailContent(object), customEmailSender(object), useForInternalEmails(boolean)
- automated_templates:
  - primary key: id
  - fields: id(string), name(string), salesforceReady(boolean)
- user_me:
  - primary key: id
  - fields: createdAt(integer), currentOrganizationId(integer), email(string), fullName(string), hasAcceptedTerms(boolean), hasPassword(boolean), hasPicture(boolean), id(integer), timezone(string)
- user_preferences:
  - primary key: name
  - fields: dateFormat(string), deadlinesNotificationDays(integer), deadlinesNotificationEnabled(boolean), language(string), mobile(string), mobileCode(string), name(string), phone(string), phoneCode(string)
- webhooks_integrations:
  - primary key: id
  - fields: events(array), id(string), isActive(boolean), url(string)
- agreement:
  - primary key: uid
  - fields: creation(object), folderId(integer), lastPublicVersion(object), lock(object), metadata(object), permission(string), uid(string)
- agreement_metadata:
  - primary key: agreement_uid
  - fields: agreement_uid(string), bookmarked(boolean), description(string), inboxed(boolean), lastAccessAt(integer), organization(object), read(boolean), status(string), tags(array), title(string), trashed(boolean)
- agreement_summary:
  - primary key: agreementUid
  - fields: agreementCategory(string), agreementUid(string), clauses(array), description(string), documentType(string), endclauses(array), lifecycle(object), organizationId(integer), signedwithlabels(array), totalAgreementValue(number)
- agreement_comments:
  - primary key: comment_uuid
  - fields: agreement_id(string), comment_uuid(string), commentedText(string), createdAt(integer), createdBy(object), reply(array), resolved(boolean), text(string), uuid(string), version(integer), visibility(string)
- agreement_activities:
  - primary key: id
  - fields: action(string), agreement_id(string), createdAt(integer), id(string), organization(object), params(object), status(string), userOrganization(object), visibility(string)
- agreement_members:
  - primary key: agreement_id, member_id
  - fields: agreement_id(string), lastAccessAt(integer), member_id(integer), permission(string), relation(string), status(string), user(object), userSignStatus(string)
- agreement_versions:
  - primary key: id
  - fields: agreement_id(string), comment(string), date(integer), displayVersion(number), id(string), organization(object), type(string), user(object), version(integer), visibility(string)
- agreement_attachments:
  - primary key: id
  - fields: agreement_id(string), contentType(string), id(string), name(string), size(integer)

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite

## Reverse ETL Actions

- create_folder:
  - endpoint: POST /organizations/{{ config.organization_id }}/folders
  - required fields: name, parentId
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
  - required fields: id, name, description, filters
  - risk: replaces an existing Concord saved report's definition; may change what other users see when they run it
- delete_report:
  - endpoint: DELETE /organizations/{{ config.organization_id }}/reports/{{ record.id }}
  - required fields: id
  - risk: permanently deletes a Concord saved report; destructive, external mutation; approval required
- create_clause:
  - endpoint: POST /organizations/{{ config.organization_id }}/clauses
  - required fields: title, content
  - risk: creates a new reusable Concord clause template within the configured organization; low risk
- update_clause:
  - endpoint: PUT /organizations/{{ config.organization_id }}/clauses/{{ record.id }}
  - required fields: id, title, content
  - risk: updates an existing Concord clause template; may affect future agreements linked to this clause
- delete_clause:
  - endpoint: DELETE /organizations/{{ config.organization_id }}/clauses/{{ record.id }}
  - required fields: id
  - risk: permanently deletes a Concord clause template; destructive, external mutation; approval required
- create_group:
  - endpoint: POST /organizations/{{ config.organization_id }}/groups
  - required fields: name
  - risk: creates a new Concord user group within the configured organization; low risk
- create_approval:
  - endpoint: POST /organizations/{{ config.organization_id }}/approvals
  - required fields: title, description, blockThirdPartySignature
  - risk: creates a new Concord company approval workflow within the configured organization; affects future agreement signature routing
- update_approval:
  - endpoint: POST /organizations/{{ config.organization_id }}/approvals/{{ record.id }}
  - required fields: id, title, description, blockThirdPartySignature
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

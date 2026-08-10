# pm connectors inspect concord

```text
NAME
  pm connectors inspect concord - Concord connector manual

SYNOPSIS
  pm connectors inspect concord
  pm connectors inspect concord --json
  pm credentials add <name> --connector concord [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads and writes Concord contract lifecycle management data: agreements (and their metadata/summary/comments/activities/members/versions/attachments sub-resources), organizations, folders, reports, tags, clauses, approvals, groups, members, events, subscription, branding, and automated templates through the Concord REST API.

ICON
  id: pm-sample
  asset: icons/pm-sample.svg
  source: polymetrics
  review_status: polymetrics
  review_url: https://github.com/polymetrics-ai/cli

CAPABILITIES
  check=true catalog=true read=true write=true query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  agreement_uid
  approval_id
  base_url
  clause_id
  clauses_page_size
  events_end_date
  events_start_date
  folder_id
  mode
  organization_id
  page_size
  report_id
  api_key (secret)

ETL STREAMS
  agreements:
    primary key: uid
    fields: createdAt(string), organizationId(integer), stage(string), status(string), title(string), uid(string), updatedAt(string)
  user_organizations:
    primary key: id
    fields: id(integer), name(string), role(string), type(string)
  folders:
    primary key: id
    fields: id(integer), name(string), organizationId(integer), parentId(integer)
  reports:
    primary key: id
    fields: id(integer), name(string), organizationId(integer), type(string)
  tags:
    primary key: id
    fields: color(string), id(integer), name(string)
  organization:
    primary key: id
    fields: aiEnabled(boolean), allTagsVisible(boolean), askForTags(boolean), canCollaboratorSign(boolean), createdAt(integer), deleted(boolean), description(string), emailDomains(array), id(integer), logo(string), name(string), parent(object), region(string), subscription(object), subsidiaries(array)
  folder:
    primary key: id
    fields: access(object), createdAt(integer), createdBy(object), id(integer), isBookmarked(boolean), modifiedAt(integer), name(string), parentId(integer)
  folder_agreements:
    primary key: uuid
    fields: createdAt(integer), folderId(integer), modifiedAt(integer), organizationId(integer), status(string), title(string), uuid(string)
  report:
    primary key: id
    fields: description(string), filters(object), id(string), lastUpdatedAt(integer), name(string)
  clauses:
    primary key: id
    fields: createdAt(integer), description(string), id(integer), numberOfTemplatesLinked(integer), presignedUrl(string), title(string), version(integer)
  clause:
    primary key: id
    fields: createdAt(integer), description(string), id(integer), numberOfTemplatesLinked(integer), presignedUrl(string), title(string), version(integer)
  approvals:
    primary key: id
    fields: blockThirdPartySignature(boolean), deletable(boolean), description(string), id(integer), rules(array), title(string)
  approval:
    primary key: id
    fields: blockThirdPartySignature(boolean), deletable(boolean), description(string), id(integer), rules(array), title(string)
  groups:
    primary key: id
    fields: description(string), id(integer), invitations(array), name(string), organization(object), users(array)
  members:
    primary key: userOrganizationId
    fields: createdAt(integer), groups(array), invitation(object), isActive(boolean), job(string), organization(object), role(object), type(string), user(object), userOrganizationId(integer)
  events:
    primary key: id
    fields: actor(object), createdAt(integer), event(object), id(integer), type(string)
  subscription:
    primary key: subscriptionId
    fields: customerId(string), featureLevel(string), seats(array), status(string), subscriptionId(string), subscriptionName(string), type(string)
  branding:
    primary key: useForInternalEmails
    fields: customAgreementView(object), customEmailContent(object), customEmailSender(object), useForInternalEmails(boolean)
  automated_templates:
    primary key: id
    fields: id(string), name(string), salesforceReady(boolean)
  user_me:
    primary key: id
    fields: createdAt(integer), currentOrganizationId(integer), email(string), fullName(string), hasAcceptedTerms(boolean), hasPassword(boolean), hasPicture(boolean), id(integer), timezone(string)
  user_preferences:
    primary key: name
    fields: dateFormat(string), deadlinesNotificationDays(integer), deadlinesNotificationEnabled(boolean), language(string), mobile(string), mobileCode(string), name(string), phone(string), phoneCode(string)
  webhooks_integrations:
    primary key: id
    fields: events(array), id(string), isActive(boolean), url(string)
  agreement:
    primary key: uid
    fields: creation(object), folderId(integer), lastPublicVersion(object), lock(object), metadata(object), permission(string), uid(string)
  agreement_metadata:
    primary key: agreement_uid
    fields: agreement_uid(string), bookmarked(boolean), description(string), inboxed(boolean), lastAccessAt(integer), organization(object), read(boolean), status(string), tags(array), title(string), trashed(boolean)
  agreement_summary:
    primary key: agreementUid
    fields: agreementCategory(string), agreementUid(string), clauses(array), description(string), documentType(string), endclauses(array), lifecycle(object), organizationId(integer), signedwithlabels(array), totalAgreementValue(number)
  agreement_comments:
    primary key: comment_uuid
    fields: agreement_id(string), comment_uuid(string), commentedText(string), createdAt(integer), createdBy(object), reply(array), resolved(boolean), text(string), uuid(string), version(integer), visibility(string)
  agreement_activities:
    primary key: id
    fields: action(string), agreement_id(string), createdAt(integer), id(string), organization(object), params(object), status(string), userOrganization(object), visibility(string)
  agreement_members:
    primary key: agreement_id, member_id
    fields: agreement_id(string), lastAccessAt(integer), member_id(integer), permission(string), relation(string), status(string), user(object), userSignStatus(string)
  agreement_versions:
    primary key: id
    fields: agreement_id(string), comment(string), date(integer), displayVersion(number), id(string), organization(object), type(string), user(object), version(integer), visibility(string)
  agreement_attachments:
    primary key: id
    fields: agreement_id(string), contentType(string), id(string), name(string), size(integer)

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

REVERSE ETL ACTIONS
  create_folder:
    endpoint: POST /organizations/{{ config.organization_id }}/folders
    required fields: name, parentId
    risk: creates a new Concord folder within the configured organization; low risk, no data destruction
  update_folder:
    endpoint: PUT /organizations/{{ config.organization_id }}/folders/{{ record.id }}
    required fields: id
    risk: renames/moves an existing Concord folder; may change document organization visible to other users
  delete_folder:
    endpoint: DELETE /organizations/{{ config.organization_id }}/folders/{{ record.id }}
    required fields: id
    risk: permanently deletes a Concord folder; destructive, external mutation; approval required
  create_report:
    endpoint: POST /organizations/{{ config.organization_id }}/reports
    risk: creates a new saved Concord report within the configured organization; low risk
  update_report:
    endpoint: PUT /organizations/{{ config.organization_id }}/reports/{{ record.id }}
    required fields: id, name, description, filters
    risk: replaces an existing Concord saved report's definition; may change what other users see when they run it
  delete_report:
    endpoint: DELETE /organizations/{{ config.organization_id }}/reports/{{ record.id }}
    required fields: id
    risk: permanently deletes a Concord saved report; destructive, external mutation; approval required
  create_clause:
    endpoint: POST /organizations/{{ config.organization_id }}/clauses
    required fields: title, content
    risk: creates a new reusable Concord clause template within the configured organization; low risk
  update_clause:
    endpoint: PUT /organizations/{{ config.organization_id }}/clauses/{{ record.id }}
    required fields: id, title, content
    risk: updates an existing Concord clause template; may affect future agreements linked to this clause
  delete_clause:
    endpoint: DELETE /organizations/{{ config.organization_id }}/clauses/{{ record.id }}
    required fields: id
    risk: permanently deletes a Concord clause template; destructive, external mutation; approval required
  create_group:
    endpoint: POST /organizations/{{ config.organization_id }}/groups
    required fields: name
    risk: creates a new Concord user group within the configured organization; low risk
  create_approval:
    endpoint: POST /organizations/{{ config.organization_id }}/approvals
    required fields: title, description, blockThirdPartySignature
    risk: creates a new Concord company approval workflow within the configured organization; affects future agreement signature routing
  update_approval:
    endpoint: POST /organizations/{{ config.organization_id }}/approvals/{{ record.id }}
    required fields: id, title, description, blockThirdPartySignature
    risk: replaces an existing Concord company approval workflow; affects agreements already routed through it
  delete_approval:
    endpoint: DELETE /organizations/{{ config.organization_id }}/approvals/{{ record.id }}
    required fields: id
    risk: permanently deletes a Concord company approval workflow; destructive, external mutation; approval required

SECURITY
  read risk: external Concord API read of contract lifecycle management data (agreements and sub-resources, organizations, folders, reports, tags, clauses, approvals, groups, members, events, subscription, branding)
  write risk: external mutation of Concord folders, reports, clauses, groups, and company approval workflows (create/update/delete); does not create, sign, or modify agreements themselves
  approval: required for delete_folder/delete_report/delete_clause/delete_approval (destructive); create/update actions are lower risk but still mutate shared organization configuration
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

COMMAND SURFACE
  Run Concord's declared streams and reverse-ETL actions.
  Usage: pm concord <command> [flags]
  Read streams
  Reverse ETL writes
  Other Commands
    agreement activities list - Run the agreement activities ETL stream [intent=etl availability=implemented stream=agreement_activities]
    agreement attachments list - Run the agreement attachments ETL stream [intent=etl availability=implemented stream=agreement_attachments]
    agreement comments list - Run the agreement comments ETL stream [intent=etl availability=implemented stream=agreement_comments]
    agreement list - Run the agreement ETL stream [intent=etl availability=implemented stream=agreement]
    agreement members list - Run the agreement members ETL stream [intent=etl availability=implemented stream=agreement_members]
    agreement metadata list - Run the agreement metadata ETL stream [intent=etl availability=implemented stream=agreement_metadata]
    agreement summary list - Run the agreement summary ETL stream [intent=etl availability=implemented stream=agreement_summary]
    agreement versions list - Run the agreement versions ETL stream [intent=etl availability=implemented stream=agreement_versions]
    agreements list - Run the agreements ETL stream [intent=etl availability=implemented stream=agreements]; notes: discrepancy=present-in-surface-absent-from-artifact
    api delete organizations organizationid agreements agreementuid - Documented DELETE /organizations/{organizationId}/agreements/{agreementUid} (not implemented) [intent=direct_write availability=not_implemented operation=concord.delete.organizations-organizationid-agreements-agreementuid]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api delete organizations organizationid agreements agreementuid approval - Documented DELETE /organizations/{organizationId}/agreements/{agreementUid}/approval (not implemented) [intent=direct_write availability=not_implemented operation=concord.delete.organizations-organizationid-agreements-agreementuid-approval]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api delete organizations organizationid agreements agreementuid attachments id - Documented DELETE /organizations/{organizationId}/agreements/{agreementUid}/attachments/{id} (not implemented) [intent=direct_write availability=not_implemented operation=concord.delete.organizations-organizationid-agreements-agreementuid-attachments-id]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api delete organizations organizationid agreements agreementuid members delayed-invitations delayedinvitationid - Documented DELETE /organizations/{organizationId}/agreements/{agreementUid}/members/delayed-invitations/{delayedInvitationId} (not implemented) [intent=direct_write availability=not_implemented operation=concord.delete.organizations-organizationid-agreements-agreementuid-members-delayed-invitations-delayedinvitationid]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api delete organizations organizationid agreements agreementuid members invitations invitationid - Documented DELETE /organizations/{organizationId}/agreements/{agreementUid}/members/invitations/{invitationId} (not implemented) [intent=direct_write availability=not_implemented operation=concord.delete.organizations-organizationid-agreements-agreementuid-members-invitations-invitationid]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api delete organizations organizationid agreements agreementuid members users me - Documented DELETE /organizations/{organizationId}/agreements/{agreementUid}/members/users/me (not implemented) [intent=direct_write availability=not_implemented operation=concord.delete.organizations-organizationid-agreements-agreementuid-members-users-me]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api delete organizations organizationid agreements agreementuid members users userid - Documented DELETE /organizations/{organizationId}/agreements/{agreementUid}/members/users/{userId} (not implemented) [intent=direct_write availability=not_implemented operation=concord.delete.organizations-organizationid-agreements-agreementuid-members-users-userid]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api delete organizations organizationid agreements agreementuid signature - Documented DELETE /organizations/{organizationId}/agreements/{agreementUid}/signature (not implemented) [intent=direct_write availability=not_implemented operation=concord.delete.organizations-organizationid-agreements-agreementuid-signature]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api delete organizations organizationid agreements agreementuid summary clauses id - Documented DELETE /organizations/{organizationId}/agreements/{agreementUid}/summary/clauses/{id} (not implemented) [intent=direct_write availability=not_implemented operation=concord.delete.organizations-organizationid-agreements-agreementuid-summary-clauses-id]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api delete organizations organizationid agreements agreementuid summary endclauses id - Documented DELETE /organizations/{organizationId}/agreements/{agreementUid}/summary/endclauses/{id} (not implemented) [intent=direct_write availability=not_implemented operation=concord.delete.organizations-organizationid-agreements-agreementuid-summary-endclauses-id]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api delete organizations organizationid agreements agreementuid summary fields id - Documented DELETE /organizations/{organizationId}/agreements/{agreementUid}/summary/fields/{id} (not implemented) [intent=direct_write availability=not_implemented operation=concord.delete.organizations-organizationid-agreements-agreementuid-summary-fields-id]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api delete organizations organizationid agreements agreementuid summary links id - Documented DELETE /organizations/{organizationId}/agreements/{agreementUid}/summary/links/{id} (not implemented) [intent=direct_write availability=not_implemented operation=concord.delete.organizations-organizationid-agreements-agreementuid-summary-links-id]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api delete organizations organizationid branding email - Documented DELETE /organizations/{organizationId}/branding/email (not implemented) [intent=direct_write availability=not_implemented operation=concord.delete.organizations-organizationid-branding-email]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api delete organizations organizationid branding logo - Documented DELETE /organizations/{organizationId}/branding/logo (not implemented) [intent=direct_write availability=not_implemented operation=concord.delete.organizations-organizationid-branding-logo]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api get organizations organizationid agreements agreementuid approval - Documented GET /organizations/{organizationId}/agreements/{agreementUid}/approval (not implemented) [intent=direct_read availability=not_implemented operation=concord.get.organizations-organizationid-agreements-agreementuid-approval]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get organizations organizationid agreements agreementuid attachments id - Documented GET /organizations/{organizationId}/agreements/{agreementUid}/attachments/{id} (not implemented) [intent=direct_read availability=not_implemented operation=concord.get.organizations-organizationid-agreements-agreementuid-attachments-id]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get organizations organizationid agreements agreementuid docx - Documented GET /organizations/{organizationId}/agreements/{agreementUid}.docx (not implemented) [intent=direct_read availability=not_implemented operation=concord.get.organizations-organizationid-agreements-agreementuid-docx]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get organizations organizationid agreements agreementuid path - Documented GET /organizations/{organizationId}/agreements/{agreementUid}/path (not implemented) [intent=direct_read availability=not_implemented operation=concord.get.organizations-organizationid-agreements-agreementuid-path]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get organizations organizationid agreements agreementuid pdf - Documented GET /organizations/{organizationId}/agreements/{agreementUid}.pdf (not implemented) [intent=direct_read availability=not_implemented operation=concord.get.organizations-organizationid-agreements-agreementuid-pdf]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get organizations organizationid agreements agreementuid signature - Documented GET /organizations/{organizationId}/agreements/{agreementUid}/signature (not implemented) [intent=direct_read availability=not_implemented operation=concord.get.organizations-organizationid-agreements-agreementuid-signature]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get organizations organizationid agreements agreementuid summary clauses id - Documented GET /organizations/{organizationId}/agreements/{agreementUid}/summary/clauses/{id} (not implemented) [intent=direct_read availability=not_implemented operation=concord.get.organizations-organizationid-agreements-agreementuid-summary-clauses-id]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get organizations organizationid agreements agreementuid summary endclauses id - Documented GET /organizations/{organizationId}/agreements/{agreementUid}/summary/endclauses/{id} (not implemented) [intent=direct_read availability=not_implemented operation=concord.get.organizations-organizationid-agreements-agreementuid-summary-endclauses-id]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get organizations organizationid agreements agreementuid summary endclauses last - Documented GET /organizations/{organizationId}/agreements/{agreementUid}/summary/endclauses/last (not implemented) [intent=direct_read availability=not_implemented operation=concord.get.organizations-organizationid-agreements-agreementuid-summary-endclauses-last]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get organizations organizationid agreements agreementuid summary fields - Documented GET /organizations/{organizationId}/agreements/{agreementUid}/summary/fields (not implemented) [intent=direct_read availability=not_implemented operation=concord.get.organizations-organizationid-agreements-agreementuid-summary-fields]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get organizations organizationid agreements agreementuid track-changes - Documented GET /organizations/{organizationId}/agreements/{agreementUid}/track-changes (not implemented) [intent=direct_read availability=not_implemented operation=concord.get.organizations-organizationid-agreements-agreementuid-track-changes]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get organizations organizationid agreements agreementuid versions diff - Documented GET /organizations/{organizationId}/agreements/{agreementUid}/versions/diff (not implemented) [intent=direct_read availability=not_implemented operation=concord.get.organizations-organizationid-agreements-agreementuid-versions-diff]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get organizations organizationid agreements agreementuid versions last - Documented GET /organizations/{organizationId}/agreements/{agreementUid}/versions/last (not implemented) [intent=direct_read availability=not_implemented operation=concord.get.organizations-organizationid-agreements-agreementuid-versions-last]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get organizations organizationid agreements agreementuid versions last fields - Documented GET /organizations/{organizationId}/agreements/{agreementUid}/versions/last/fields (not implemented) [intent=direct_read availability=not_implemented operation=concord.get.organizations-organizationid-agreements-agreementuid-versions-last-fields]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get user me organizations organizationid agreements - Documented GET /user/me/organizations/{organizationId}/agreements (not implemented) [intent=direct_read availability=not_implemented operation=concord.get.user-me-organizations-organizationid-agreements]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get user me organizations organizationid folders - Documented GET /user/me/organizations/{organizationId}/folders (not implemented) [intent=direct_read availability=not_implemented operation=concord.get.user-me-organizations-organizationid-folders]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api patch organizations organizationid agreements agreementuid - Documented PATCH /organizations/{organizationId}/agreements/{agreementUid} (not implemented) [intent=direct_write availability=not_implemented operation=concord.patch.organizations-organizationid-agreements-agreementuid]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api patch organizations organizationid agreements agreementuid attachments id - Documented PATCH /organizations/{organizationId}/agreements/{agreementUid}/attachments/{id} (not implemented) [intent=direct_write availability=not_implemented operation=concord.patch.organizations-organizationid-agreements-agreementuid-attachments-id]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api patch organizations organizationid agreements agreementuid metadata - Documented PATCH /organizations/{organizationId}/agreements/{agreementUid}/metadata (not implemented) [intent=direct_write availability=not_implemented operation=concord.patch.organizations-organizationid-agreements-agreementuid-metadata]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api patch organizations organizationid agreements agreementuid rules id - Documented PATCH /organizations/{organizationId}/agreements/{agreementUid}/rules/{id} (not implemented) [intent=direct_write availability=not_implemented operation=concord.patch.organizations-organizationid-agreements-agreementuid-rules-id]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api patch organizations organizationid agreements agreementuid signature - Documented PATCH /organizations/{organizationId}/agreements/{agreementUid}/signature (not implemented) [intent=direct_write availability=not_implemented operation=concord.patch.organizations-organizationid-agreements-agreementuid-signature]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api patch organizations organizationid agreements agreementuid summary clauses id - Documented PATCH /organizations/{organizationId}/agreements/{agreementUid}/summary/clauses/{id} (not implemented) [intent=direct_write availability=not_implemented operation=concord.patch.organizations-organizationid-agreements-agreementuid-summary-clauses-id]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api patch organizations organizationid agreements agreementuid summary endclauses last - Documented PATCH /organizations/{organizationId}/agreements/{agreementUid}/summary/endclauses/last (not implemented) [intent=direct_write availability=not_implemented operation=concord.patch.organizations-organizationid-agreements-agreementuid-summary-endclauses-last]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api patch organizations organizationid agreements agreementuid summary fields id - Documented PATCH /organizations/{organizationId}/agreements/{agreementUid}/summary/fields/{id} (not implemented) [intent=direct_write availability=not_implemented operation=concord.patch.organizations-organizationid-agreements-agreementuid-summary-fields-id]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api patch organizations organizationid agreements agreementuid summary lifecycle - Documented PATCH /organizations/{organizationId}/agreements/{agreementUid}/summary/lifecycle (not implemented) [intent=direct_write availability=not_implemented operation=concord.patch.organizations-organizationid-agreements-agreementuid-summary-lifecycle]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api patch organizations organizationid agreements agreementuid versions last fields - Documented PATCH /organizations/{organizationId}/agreements/{agreementUid}/versions/last/fields (not implemented) [intent=direct_write availability=not_implemented operation=concord.patch.organizations-organizationid-agreements-agreementuid-versions-last-fields]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api patch organizations organizationid agreements agreementuid versions last pdffields - Documented PATCH /organizations/{organizationId}/agreements/{agreementUid}/versions/last/pdffields (not implemented) [intent=direct_write availability=not_implemented operation=concord.patch.organizations-organizationid-agreements-agreementuid-versions-last-pdffields]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api patch organizations organizationid groups - Documented PATCH /organizations/{organizationId}/groups (not implemented) [intent=direct_write availability=not_implemented operation=concord.patch.organizations-organizationid-groups]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post organizations - Documented POST /organizations (not implemented) [intent=direct_write availability=not_implemented operation=concord.post.organizations]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post organizations organizationid agreements - Documented POST /organizations/{organizationId}/agreements (not implemented) [intent=direct_write availability=not_implemented operation=concord.post.organizations-organizationid-agreements]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post organizations organizationid agreements agreementuid approval - Documented POST /organizations/{organizationId}/agreements/{agreementUid}/approval (not implemented) [intent=direct_write availability=not_implemented operation=concord.post.organizations-organizationid-agreements-agreementuid-approval]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post organizations organizationid agreements agreementuid attachments - Documented POST /organizations/{organizationId}/agreements/{agreementUid}/attachments (not implemented) [intent=direct_write availability=not_implemented operation=concord.post.organizations-organizationid-agreements-agreementuid-attachments]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post organizations organizationid agreements agreementuid comments - Documented POST /organizations/{organizationId}/agreements/{agreementUid}/comments (not implemented) [intent=direct_write availability=not_implemented operation=concord.post.organizations-organizationid-agreements-agreementuid-comments]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post organizations organizationid agreements agreementuid integration - Documented POST /organizations/{organizationId}/agreements/{agreementUid}/integration (not implemented) [intent=direct_write availability=not_implemented operation=concord.post.organizations-organizationid-agreements-agreementuid-integration]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post organizations organizationid agreements agreementuid members - Documented POST /organizations/{organizationId}/agreements/{agreementUid}/members (not implemented) [intent=direct_write availability=not_implemented operation=concord.post.organizations-organizationid-agreements-agreementuid-members]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post organizations organizationid agreements agreementuid members delayed-invitations - Documented POST /organizations/{organizationId}/agreements/{agreementUid}/members/delayed-invitations (not implemented) [intent=direct_write availability=not_implemented operation=concord.post.organizations-organizationid-agreements-agreementuid-members-delayed-invitations]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post organizations organizationid agreements agreementuid members delayed-invitations delayedinvitationid send - Documented POST /organizations/{organizationId}/agreements/{agreementUid}/members/delayed-invitations/{delayedInvitationId}/send (not implemented) [intent=direct_write availability=not_implemented operation=concord.post.organizations-organizationid-agreements-agreementuid-members-delayed-invitations-delayedinvitationid-send]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post organizations organizationid agreements agreementuid members me - Documented POST /organizations/{organizationId}/agreements/{agreementUid}/members/me (not implemented) [intent=direct_write availability=not_implemented operation=concord.post.organizations-organizationid-agreements-agreementuid-members-me]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post organizations organizationid agreements agreementuid signature request - Documented POST /organizations/{organizationId}/agreements/{agreementUid}/signature/request (not implemented) [intent=direct_write availability=not_implemented operation=concord.post.organizations-organizationid-agreements-agreementuid-signature-request]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post organizations organizationid agreements agreementuid summary clauses - Documented POST /organizations/{organizationId}/agreements/{agreementUid}/summary/clauses (not implemented) [intent=direct_write availability=not_implemented operation=concord.post.organizations-organizationid-agreements-agreementuid-summary-clauses]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post organizations organizationid agreements agreementuid summary endclauses - Documented POST /organizations/{organizationId}/agreements/{agreementUid}/summary/endclauses (not implemented) [intent=direct_write availability=not_implemented operation=concord.post.organizations-organizationid-agreements-agreementuid-summary-endclauses]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post organizations organizationid agreements agreementuid summary fields - Documented POST /organizations/{organizationId}/agreements/{agreementUid}/summary/fields (not implemented) [intent=direct_write availability=not_implemented operation=concord.post.organizations-organizationid-agreements-agreementuid-summary-fields]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post organizations organizationid agreements agreementuid summary links - Documented POST /organizations/{organizationId}/agreements/{agreementUid}/summary/links (not implemented) [intent=direct_write availability=not_implemented operation=concord.post.organizations-organizationid-agreements-agreementuid-summary-links]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post organizations organizationid agreements agreementuid versions upload - Documented POST /organizations/{organizationId}/agreements/{agreementUid}/versions/upload (not implemented) [intent=direct_write availability=not_implemented operation=concord.post.organizations-organizationid-agreements-agreementuid-versions-upload]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post organizations organizationid auto agreementuid - Documented POST /organizations/{organizationId}/auto/{agreementUid} (not implemented) [intent=direct_write availability=not_implemented operation=concord.post.organizations-organizationid-auto-agreementuid]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post organizations organizationid branding email - Documented POST /organizations/{organizationId}/branding/email (not implemented) [intent=direct_write availability=not_implemented operation=concord.post.organizations-organizationid-branding-email]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post organizations organizationid branding logo - Documented POST /organizations/{organizationId}/branding/logo (not implemented) [intent=direct_write availability=not_implemented operation=concord.post.organizations-organizationid-branding-logo]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post organizations organizationid invitations - Documented POST /organizations/{organizationId}/invitations (not implemented) [intent=direct_write availability=not_implemented operation=concord.post.organizations-organizationid-invitations]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post organizations organizationid reports id - Documented POST /organizations/{organizationId}/reports/{id} (not implemented) [intent=direct_write availability=not_implemented operation=concord.post.organizations-organizationid-reports-id]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post user me organizations organizationid agreements count - Documented POST /user/me/organizations/{organizationId}/agreements/count (not implemented) [intent=direct_write availability=not_implemented operation=concord.post.user-me-organizations-organizationid-agreements-count]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post users me integrations webhooks - Documented POST /users/me/integrations/webhooks (not implemented) [intent=direct_write availability=not_implemented operation=concord.post.users-me-integrations-webhooks]; approval: not implemented: the provider webhook-URL mutation has no dedicated security-reviewed execution contract; risk: high; notes: named_dependency=review.webhook_url_mutation: the provider webhook-URL mutation has no dedicated security-reviewed execution contract
    api put organizations organizationid agreements agreementuid folder - Documented PUT /organizations/{organizationId}/agreements/{agreementUid}/folder (not implemented) [intent=direct_write availability=not_implemented operation=concord.put.organizations-organizationid-agreements-agreementuid-folder]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api put organizations organizationid agreements agreementuid members delayed-invitations delayedinvitationid permission - Documented PUT /organizations/{organizationId}/agreements/{agreementUid}/members/delayed-invitations/{delayedInvitationId}/permission (not implemented) [intent=direct_write availability=not_implemented operation=concord.put.organizations-organizationid-agreements-agreementuid-members-delayed-invitations-delayedinvitationid-permission]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api put organizations organizationid agreements agreementuid members invitations invitationid permission - Documented PUT /organizations/{organizationId}/agreements/{agreementUid}/members/invitations/{invitationId}/permission (not implemented) [intent=direct_write availability=not_implemented operation=concord.put.organizations-organizationid-agreements-agreementuid-members-invitations-invitationid-permission]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api put organizations organizationid agreements agreementuid members users userid permission - Documented PUT /organizations/{organizationId}/agreements/{agreementUid}/members/users/{userId}/permission (not implemented) [intent=direct_write availability=not_implemented operation=concord.put.organizations-organizationid-agreements-agreementuid-members-users-userid-permission]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api put organizations organizationid agreements agreementuid signature - Documented PUT /organizations/{organizationId}/agreements/{agreementUid}/signature (not implemented) [intent=direct_write availability=not_implemented operation=concord.put.organizations-organizationid-agreements-agreementuid-signature]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api put organizations organizationid agreements agreementuid signature slots - Documented PUT /organizations/{organizationId}/agreements/{agreementUid}/signature/slots (not implemented) [intent=direct_write availability=not_implemented operation=concord.put.organizations-organizationid-agreements-agreementuid-signature-slots]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api put organizations organizationid agreements agreementuid summary signedwithlabels - Documented PUT /organizations/{organizationId}/agreements/{agreementUid}/summary/signedwithlabels (not implemented) [intent=direct_write availability=not_implemented operation=concord.put.organizations-organizationid-agreements-agreementuid-summary-signedwithlabels]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api put organizations organizationid branding agreementview - Documented PUT /organizations/{organizationId}/branding/agreementview (not implemented) [intent=direct_write availability=not_implemented operation=concord.put.organizations-organizationid-branding-agreementview]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api put organizations organizationid branding email - Documented PUT /organizations/{organizationId}/branding/email (not implemented) [intent=direct_write availability=not_implemented operation=concord.put.organizations-organizationid-branding-email]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api put organizations organizationid branding logo - Documented PUT /organizations/{organizationId}/branding/logo (not implemented) [intent=direct_write availability=not_implemented operation=concord.put.organizations-organizationid-branding-logo]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api webhook newagreement-post - Documented WEBHOOK newAgreement#POST (not implemented) [intent=docs_only availability=not_implemented operation=concord.webhook.newagreement-post]; approval: not implemented: the runtime has no inbound webhook receiver executor for top-level webhook operations; risk: medium; notes: named_dependency=engine.webhook_receiver_executor: the runtime has no inbound webhook receiver executor for top-level webhook operations
    approval list - Run the approval ETL stream [intent=etl availability=implemented stream=approval]
    approvals list - Run the approvals ETL stream [intent=etl availability=implemented stream=approvals]
    automated templates list - Run the automated templates ETL stream [intent=etl availability=implemented stream=automated_templates]
    branding list - Run the branding ETL stream [intent=etl availability=implemented stream=branding]
    clause list - Run the clause ETL stream [intent=etl availability=implemented stream=clause]
    clauses list - Run the clauses ETL stream [intent=etl availability=implemented stream=clauses]
    create approval apply - Plan and execute the create approval reverse-ETL action [intent=reverse_etl availability=implemented write=create_approval]; approval: requires plan, preview, approval, and execute; risk: creates a new Concord company approval workflow within the configured organization; affects future agreement signature routing; flags: --blockThirdPartySignature (required), --description (required), --title (required)
    create clause apply - Plan and execute the create clause reverse-ETL action [intent=reverse_etl availability=implemented write=create_clause]; approval: requires plan, preview, approval, and execute; risk: creates a new reusable Concord clause template within the configured organization; low risk; flags: --content (required), --title (required)
    create folder apply - Plan and execute the create folder reverse-ETL action [intent=reverse_etl availability=implemented write=create_folder]; approval: requires plan, preview, approval, and execute; risk: creates a new Concord folder within the configured organization; low risk, no data destruction; flags: --name (required), --parentId (required)
    create group apply - Plan and execute the create group reverse-ETL action [intent=reverse_etl availability=implemented write=create_group]; approval: requires plan, preview, approval, and execute; risk: creates a new Concord user group within the configured organization; low risk; flags: --name (required)
    create report apply - Plan and execute the create report reverse-ETL action [intent=reverse_etl availability=implemented write=create_report]; approval: requires plan, preview, approval, and execute; risk: creates a new saved Concord report within the configured organization; low risk
    delete approval apply - Plan and execute the delete approval reverse-ETL action [intent=reverse_etl availability=implemented write=delete_approval]; approval: requires plan, preview, approval, and execute; risk: permanently deletes a Concord company approval workflow; destructive, external mutation; approval required; flags: --id (required)
    delete clause apply - Plan and execute the delete clause reverse-ETL action [intent=reverse_etl availability=implemented write=delete_clause]; approval: requires plan, preview, approval, and execute; risk: permanently deletes a Concord clause template; destructive, external mutation; approval required; flags: --id (required)
    delete folder apply - Plan and execute the delete folder reverse-ETL action [intent=reverse_etl availability=implemented write=delete_folder]; approval: requires plan, preview, approval, and execute; risk: permanently deletes a Concord folder; destructive, external mutation; approval required; flags: --id (required)
    delete report apply - Plan and execute the delete report reverse-ETL action [intent=reverse_etl availability=implemented write=delete_report]; approval: requires plan, preview, approval, and execute; risk: permanently deletes a Concord saved report; destructive, external mutation; approval required; flags: --id (required)
    events list - Run the events ETL stream [intent=etl availability=implemented stream=events]
    folder agreements list - Run the folder agreements ETL stream [intent=etl availability=implemented stream=folder_agreements]
    folder list - Run the folder ETL stream [intent=etl availability=implemented stream=folder]
    folders list - Run the folders ETL stream [intent=etl availability=implemented stream=folders]
    groups list - Run the groups ETL stream [intent=etl availability=implemented stream=groups]
    members list - Run the members ETL stream [intent=etl availability=implemented stream=members]
    organization list - Run the organization ETL stream [intent=etl availability=implemented stream=organization]
    report list - Run the report ETL stream [intent=etl availability=implemented stream=report]
    reports list - Run the reports ETL stream [intent=etl availability=implemented stream=reports]
    subscription list - Run the subscription ETL stream [intent=etl availability=implemented stream=subscription]
    tags list - Run the tags ETL stream [intent=etl availability=implemented stream=tags]
    update approval apply - Plan and execute the update approval reverse-ETL action [intent=reverse_etl availability=implemented write=update_approval]; approval: requires plan, preview, approval, and execute; risk: replaces an existing Concord company approval workflow; affects agreements already routed through it; flags: --blockThirdPartySignature (required), --description (required), --id (required), --title (required)
    update clause apply - Plan and execute the update clause reverse-ETL action [intent=reverse_etl availability=implemented write=update_clause]; approval: requires plan, preview, approval, and execute; risk: updates an existing Concord clause template; may affect future agreements linked to this clause; flags: --content (required), --id (required), --title (required)
    update folder apply - Plan and execute the update folder reverse-ETL action [intent=reverse_etl availability=implemented write=update_folder]; approval: requires plan, preview, approval, and execute; risk: renames/moves an existing Concord folder; may change document organization visible to other users; flags: --id (required)
    update report apply - Plan and execute the update report reverse-ETL action [intent=reverse_etl availability=not_implemented write=update_report]; approval: requires plan, preview, approval, and execute; risk: replaces an existing Concord saved report's definition; may change what other users see when they run it; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    user me list - Run the user me ETL stream [intent=etl availability=implemented stream=user_me]
    user organizations list - Run the user organizations ETL stream [intent=etl availability=implemented stream=user_organizations]
    user preferences list - Run the user preferences ETL stream [intent=etl availability=implemented stream=user_preferences]
    webhooks integrations list - Run the webhooks integrations ETL stream [intent=etl availability=implemented stream=webhooks_integrations]

EXAMPLES
  # Inspect as a manual
  pm connectors inspect concord

  # Inspect as structured JSON
  pm connectors inspect concord --json

AGENT WORKFLOW
  - Run pm connectors inspect concord before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.
  - For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```

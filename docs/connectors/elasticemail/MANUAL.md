# pm connectors inspect elasticemail

```text
NAME
  pm connectors inspect elasticemail - Elastic Email connector manual

SYNOPSIS
  pm connectors inspect elasticemail
  pm connectors inspect elasticemail --json
  pm credentials add <name> --connector elasticemail [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads and writes Elastic Email contacts, campaigns, lists, segments, templates, webhooks, domains, inbound routes, suppressions, and account statistics through the Elastic Email v4 REST API.

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
  base_url
  mode
  api_key (secret)

ETL STREAMS
  contacts:
    primary key: Email
    cursor: DateUpdated
    fields: Activity(object), Consent(object), CustomFields(object), DateAdded(string), DateUpdated(string), Email(string), FirstName(string), LastName(string), Source(string), Status(string), StatusChangeDate(string)
  campaigns:
    primary key: Name
    fields: Content(object), Name(string), Options(object), Recipients(object), Status(string)
  lists:
    primary key: ListName
    fields: AllowUnsubscribe(boolean), DateAdded(string), ListName(string), PublicListID(string)
  segments:
    primary key: Name
    fields: Name(string), Rule(string)
  templates:
    primary key: Name
    fields: Body(object), DateAdded(string), Name(string), Subject(string), TemplateScope(string)
  domains:
    primary key: Domain
    fields: CertificateStatus(string), CustomBouncesDomain(string), DMARC(boolean), DefaultDomain(boolean), Dkim(boolean), Domain(string), IsMarkedForDeletion(boolean), MX(boolean), Spf(boolean), TrackingStatus(string), VERP(boolean), Verify(boolean)
  suppressions:
    primary key: Email
    fields: DateUpdated(string), Email(string), ErrorCode(integer), FriendlyErrorMessage(string)
  suppressions_bounces:
    primary key: Email
    fields: DateUpdated(string), Email(string), ErrorCode(integer), FriendlyErrorMessage(string)
  suppressions_complaints:
    primary key: Email
    fields: DateUpdated(string), Email(string), ErrorCode(integer), FriendlyErrorMessage(string)
  suppressions_unsubscribes:
    primary key: Email
    fields: DateUpdated(string), Email(string), ErrorCode(integer), FriendlyErrorMessage(string)
  webhooks:
    primary key: WebhookID
    fields: DateCreated(string), DateUpdated(string), IsEnabled(boolean), Name(string), NotificationForAbuseReport(boolean), NotificationForClicked(boolean), NotificationForError(boolean), NotificationForOpened(boolean), NotificationForSent(boolean), NotificationForUnsubscribed(boolean), NotifyOncePerEmail(boolean), URL(string), WebhookID(string)
  files:
    primary key: FileName
    fields: ContentType(string), DateAdded(string), ExpirationDate(string), FileName(string), Size(integer)
  inbound_routes:
    primary key: PublicId
    fields: ActionParameter(string), ActionType(string), Filter(string), FilterType(string), Name(string), PublicId(string), SortOrder(integer)
  sub_accounts:
    primary key: PublicAccountID
    fields: ContactsCount(integer), Email(string), EmailCredits(integer), LastActivity(string), PublicAccountID(string), Reputation(number), Status(string), TotalEmailsSent(integer)
  statistics_campaigns:
    primary key: ChannelName
    fields: Bounced(integer), ChannelName(string), Clicked(integer), Complaints(integer), Delivered(integer), EmailTotal(integer), InProgress(integer), Inbound(integer), ManualCancel(integer), NotDelivered(integer), Opened(integer), Recipients(integer), SmsTotal(integer), Unsubscribed(integer)
  statistics_channels:
    primary key: ChannelName
    fields: Bounced(integer), ChannelName(string), Clicked(integer), Complaints(integer), Delivered(integer), EmailTotal(integer), InProgress(integer), Inbound(integer), ManualCancel(integer), NotDelivered(integer), Opened(integer), Recipients(integer), SmsTotal(integer), Unsubscribed(integer)

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped, incremental_append, incremental_append_deduped

REVERSE ETL ACTIONS
  create_contact:
    endpoint: POST /contacts
    required fields: Email
    risk: adds a new contact to the account's overall recipient list; low-risk external mutation, no approval required
  update_contact:
    endpoint: PUT /contacts/{{ record.Email }}
    required fields: Email
    risk: mutates an existing contact's status/name/custom-field data; a Status change (e.g. to Unsubscribed) changes future campaign eligibility for this recipient
  delete_contact:
    endpoint: DELETE /contacts/{{ record.Email }}
    required fields: Email
    risk: permanently removes a contact and its activity/consent history from the account
  create_list:
    endpoint: POST /lists
    required fields: ListName
    risk: creates a new contact list, optionally seeding it from existing contact emails; low-risk external mutation, no approval required
  update_list:
    endpoint: PUT /lists/{{ record.ListName }}
    required fields: ListName
    optional fields: NewListName, AllowUnsubscribe
    risk: renames an existing list or changes its unsubscribe-allowed setting; a rename changes the identifier campaigns/segments reference this list by
  delete_list:
    endpoint: DELETE /lists/{{ record.ListName }}
    required fields: ListName
    risk: permanently removes a contact list; any campaign still targeting this list by name will fail to resolve its recipients
  add_list_contacts:
    endpoint: POST /lists/{{ record.ListName }}/contacts
    required fields: ListName, Emails
    optional fields: Status
    risk: adds existing contacts to a list, making them eligible recipients for any campaign targeting that list
  create_segment:
    endpoint: POST /segments
    required fields: Name, Rule
    risk: creates a new dynamic contact segment from a SQL-like rule; low-risk external mutation, no approval required
  update_segment:
    endpoint: PUT /segments/{{ record.Name }}
    required fields: Name, Rule
    risk: changes the membership rule of an existing segment; immediately changes which contacts any campaign targeting this segment will reach
  delete_segment:
    endpoint: DELETE /segments/{{ record.Name }}
    required fields: Name
    risk: permanently removes a segment; any campaign still targeting this segment by name will fail to resolve its recipients
  create_template:
    endpoint: POST /templates
    required fields: Name
    risk: creates a new email template; low-risk external mutation, no approval required
  update_template:
    endpoint: PUT /templates/{{ record.Name }}
    required fields: Name
    risk: overwrites the subject/body of an existing template; any campaign referencing this template by name sends the new content on its next send
  delete_template:
    endpoint: DELETE /templates/{{ record.Name }}
    required fields: Name
    risk: permanently removes a template; any campaign still referencing this template by name will fail to build its content
  create_campaign:
    endpoint: POST /campaigns
    required fields: Name, Recipients
    risk: creates a new campaign targeting the given lists/segments; depending on Options this may schedule a live send to real recipients, not a preview-only action
  update_campaign:
    endpoint: PUT /campaigns/{{ record.Name }}
    required fields: Name
    risk: mutates an existing campaign's content, recipients, or send options; a campaign already in progress may not accept every field change
  pause_campaign:
    endpoint: PUT /campaigns/{{ record.Name }}/pause
    required fields: Name
    risk: pauses an in-progress campaign send; recipients not yet reached will not receive the email until the campaign is resumed
  delete_campaign:
    endpoint: DELETE /campaigns/{{ record.Name }}
    required fields: Name
    risk: permanently removes a campaign; if it has not finished sending, any remaining scheduled deliveries are cancelled
  create_webhook:
    endpoint: POST /webhook
    required fields: Name, URL
    risk: registers a new outbound webhook that will POST live event data (sent/opened/clicked/bounced) to an external URL of the caller's choosing; verify the target endpoint before enabling
  update_webhook:
    endpoint: PUT /webhook/{{ record.WebhookID }}
    required fields: WebhookID
    risk: mutates an existing webhook's target URL or event subscriptions; a changed URL redirects future event deliveries to a different endpoint
  delete_webhook:
    endpoint: DELETE /webhook/{{ record.WebhookID }}
    required fields: WebhookID
    risk: permanently removes a webhook subscription; event delivery to its target URL stops immediately
  create_domain:
    endpoint: POST /domains
    required fields: Domain
    risk: registers a new sending domain pending DNS verification; low-risk external mutation, no approval required
  delete_domain:
    endpoint: DELETE /domains/{{ record.Domain }}
    required fields: Domain
    risk: permanently removes a verified sending domain; any campaign configured to send from this domain will fail until reconfigured
  create_inbound_route:
    endpoint: POST /inboundroute
    required fields: Name, Filter, FilterType, ActionType
    risk: creates a new inbound-mail routing rule that forwards matching inbound email to an external address or webhook URL of the caller's choosing
  update_inbound_route:
    endpoint: PUT /inboundroute/{{ record.PublicId }}
    required fields: PublicId
    risk: mutates an existing inbound route's match filter or forwarding destination; redirects future matching inbound mail to a different address/URL
  delete_inbound_route:
    endpoint: DELETE /inboundroute/{{ record.PublicId }}
    required fields: PublicId
    risk: permanently removes an inbound-mail routing rule; matching inbound mail is no longer forwarded once removed

SECURITY
  read risk: external Elastic Email API read of contact, campaign, list, segment, template, webhook, domain, inbound-route, suppression, sub-account, and statistics data
  write risk: external Elastic Email API mutations covering contact/list/segment/template/campaign/webhook/domain/inbound-route lifecycle management; create_campaign and pause_campaign can affect a live email send to real recipients, and webhook/inbound-route writes register caller-controlled external destinations for live event/mail forwarding
  approval: standard; no destructive-admin or elevated-scope actions are exposed
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

COMMAND SURFACE
  Run Elastic Email's declared streams and reverse-ETL actions.
  Usage: pm elasticemail <command> [flags]
  Read streams
  Reverse ETL writes
  Other Commands
    add list contacts apply - Plan and execute the add list contacts reverse-ETL action [intent=reverse_etl availability=implemented write=add_list_contacts]; approval: requires plan, preview, approval, and execute; risk: adds existing contacts to a list, making them eligible recipients for any campaign targeting that list; flags: --Emails (required), --ListName (required)
    api delete files name - Documented DELETE /files/{name} (not implemented) [intent=direct_write availability=not_implemented operation=elasticemail.delete.files-name]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api delete security apikeys name - Documented DELETE /security/apikeys/{name} (not implemented) [intent=direct_write availability=not_implemented operation=elasticemail.delete.security-apikeys-name]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api delete security smtp name - Documented DELETE /security/smtp/{name} (not implemented) [intent=direct_write availability=not_implemented operation=elasticemail.delete.security-smtp-name]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api delete subaccounts email - Documented DELETE /subaccounts/{email} (not implemented) [intent=direct_write availability=not_implemented operation=elasticemail.delete.subaccounts-email]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api delete suppressions email - Documented DELETE /suppressions/{email} (not implemented) [intent=direct_write availability=not_implemented operation=elasticemail.delete.suppressions-email]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api delete verifications email - Documented DELETE /verifications/{email} (not implemented) [intent=direct_write availability=not_implemented operation=elasticemail.delete.verifications-email]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api delete verifications files id - Documented DELETE /verifications/files/{id} (not implemented) [intent=direct_write availability=not_implemented operation=elasticemail.delete.verifications-files-id]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api get campaigns name - Documented GET /campaigns/{name} (not implemented) [intent=direct_read availability=not_implemented operation=elasticemail.get.campaigns-name]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get contacts email - Documented GET /contacts/{email} (not implemented) [intent=direct_read availability=not_implemented operation=elasticemail.get.contacts-email]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get contacts export id status - Documented GET /contacts/export/{id}/status (not implemented) [intent=direct_read availability=not_implemented operation=elasticemail.get.contacts-export-id-status]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get domains domain - Documented GET /domains/{domain} (not implemented) [intent=direct_read availability=not_implemented operation=elasticemail.get.domains-domain]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get domains domain restricted - Documented GET /domains/{domain}/restricted (not implemented) [intent=direct_read availability=not_implemented operation=elasticemail.get.domains-domain-restricted]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get emails msgid view - Documented GET /emails/{msgid}/view (not implemented) [intent=direct_read availability=not_implemented operation=elasticemail.get.emails-msgid-view]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get emails transactionid status - Documented GET /emails/{transactionid}/status (not implemented) [intent=direct_read availability=not_implemented operation=elasticemail.get.emails-transactionid-status]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get events - Documented GET /events (not implemented) [intent=direct_read availability=not_implemented operation=elasticemail.get.events]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get events channels export id status - Documented GET /events/channels/export/{id}/status (not implemented) [intent=direct_read availability=not_implemented operation=elasticemail.get.events-channels-export-id-status]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get events channels name - Documented GET /events/channels/{name} (not implemented) [intent=direct_read availability=not_implemented operation=elasticemail.get.events-channels-name]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get events export id status - Documented GET /events/export/{id}/status (not implemented) [intent=direct_read availability=not_implemented operation=elasticemail.get.events-export-id-status]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get events transactionid - Documented GET /events/{transactionid} (not implemented) [intent=direct_read availability=not_implemented operation=elasticemail.get.events-transactionid]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get files name - Documented GET /files/{name} (not implemented) [intent=direct_read availability=not_implemented operation=elasticemail.get.files-name]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get files name info - Documented GET /files/{name}/info (not implemented) [intent=direct_read availability=not_implemented operation=elasticemail.get.files-name-info]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get inboundroute id - Documented GET /inboundroute/{id} (not implemented) [intent=direct_read availability=not_implemented operation=elasticemail.get.inboundroute-id]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get lists listname contacts - Documented GET /lists/{listname}/contacts (not implemented) [intent=direct_read availability=not_implemented operation=elasticemail.get.lists-listname-contacts]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get lists name - Documented GET /lists/{name} (not implemented) [intent=direct_read availability=not_implemented operation=elasticemail.get.lists-name]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get security apikeys - Documented GET /security/apikeys (not implemented) [intent=direct_read availability=not_implemented operation=elasticemail.get.security-apikeys]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get security apikeys name - Documented GET /security/apikeys/{name} (not implemented) [intent=direct_read availability=not_implemented operation=elasticemail.get.security-apikeys-name]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get security smtp - Documented GET /security/smtp (not implemented) [intent=direct_read availability=not_implemented operation=elasticemail.get.security-smtp]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get security smtp name - Documented GET /security/smtp/{name} (not implemented) [intent=direct_read availability=not_implemented operation=elasticemail.get.security-smtp-name]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get segments name - Documented GET /segments/{name} (not implemented) [intent=direct_read availability=not_implemented operation=elasticemail.get.segments-name]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get statistics - Documented GET /statistics (not implemented) [intent=direct_read availability=not_implemented operation=elasticemail.get.statistics]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get statistics campaigns name - Documented GET /statistics/campaigns/{name} (not implemented) [intent=direct_read availability=not_implemented operation=elasticemail.get.statistics-campaigns-name]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get statistics channels name - Documented GET /statistics/channels/{name} (not implemented) [intent=direct_read availability=not_implemented operation=elasticemail.get.statistics-channels-name]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get subaccounts email - Documented GET /subaccounts/{email} (not implemented) [intent=direct_read availability=not_implemented operation=elasticemail.get.subaccounts-email]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get subaccounts email apikey - Documented GET /subaccounts/{email}/apikey (not implemented) [intent=direct_read availability=not_implemented operation=elasticemail.get.subaccounts-email-apikey]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get suppressions email - Documented GET /suppressions/{email} (not implemented) [intent=direct_read availability=not_implemented operation=elasticemail.get.suppressions-email]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get templates name - Documented GET /templates/{name} (not implemented) [intent=direct_read availability=not_implemented operation=elasticemail.get.templates-name]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get verifications - Documented GET /verifications (not implemented) [intent=direct_read availability=not_implemented operation=elasticemail.get.verifications]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get verifications email - Documented GET /verifications/{email} (not implemented) [intent=direct_read availability=not_implemented operation=elasticemail.get.verifications-email]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get verifications files id result - Documented GET /verifications/files/{id}/result (not implemented) [intent=direct_read availability=not_implemented operation=elasticemail.get.verifications-files-id-result]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get verifications files id result download - Documented GET /verifications/files/{id}/result/download (not implemented) [intent=direct_read availability=not_implemented operation=elasticemail.get.verifications-files-id-result-download]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get verifications files result - Documented GET /verifications/files/result (not implemented) [intent=direct_read availability=not_implemented operation=elasticemail.get.verifications-files-result]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get webhook publicid - Documented GET /webhook/{publicid} (not implemented) [intent=direct_read availability=not_implemented operation=elasticemail.get.webhook-publicid]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api patch domains email default - Documented PATCH /domains/{email}/default (not implemented) [intent=direct_write availability=not_implemented operation=elasticemail.patch.domains-email-default]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api patch subaccounts email credits - Documented PATCH /subaccounts/{email}/credits (not implemented) [intent=direct_write availability=not_implemented operation=elasticemail.patch.subaccounts-email-credits]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post campaigns automation name trigger - Documented POST /campaigns/automation/{name}/trigger (not implemented) [intent=direct_write availability=not_implemented operation=elasticemail.post.campaigns-automation-name-trigger]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post contacts delete - Documented POST /contacts/delete (not implemented) [intent=direct_write availability=not_implemented operation=elasticemail.post.contacts-delete]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post contacts export - Documented POST /contacts/export (not implemented) [intent=direct_write availability=not_implemented operation=elasticemail.post.contacts-export]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post contacts import - Documented POST /contacts/import (not implemented) [intent=direct_write availability=not_implemented operation=elasticemail.post.contacts-import]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post emails - Documented POST /emails (not implemented) [intent=direct_write availability=not_implemented operation=elasticemail.post.emails]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post emails mergefile - Documented POST /emails/mergefile (not implemented) [intent=direct_write availability=not_implemented operation=elasticemail.post.emails-mergefile]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post emails transactional - Documented POST /emails/transactional (not implemented) [intent=direct_write availability=not_implemented operation=elasticemail.post.emails-transactional]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post events channels name export - Documented POST /events/channels/{name}/export (not implemented) [intent=direct_write availability=not_implemented operation=elasticemail.post.events-channels-name-export]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post events export - Documented POST /events/export (not implemented) [intent=direct_write availability=not_implemented operation=elasticemail.post.events-export]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post files - Documented POST /files (not implemented) [intent=direct_write availability=not_implemented operation=elasticemail.post.files]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post lists name contacts remove - Documented POST /lists/{name}/contacts/remove (not implemented) [intent=direct_write availability=not_implemented operation=elasticemail.post.lists-name-contacts-remove]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post security apikeys - Documented POST /security/apikeys (not implemented) [intent=direct_write availability=not_implemented operation=elasticemail.post.security-apikeys]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post security smtp - Documented POST /security/smtp (not implemented) [intent=direct_write availability=not_implemented operation=elasticemail.post.security-smtp]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post subaccounts - Documented POST /subaccounts (not implemented) [intent=direct_write availability=not_implemented operation=elasticemail.post.subaccounts]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post suppressions bounces - Documented POST /suppressions/bounces (not implemented) [intent=direct_write availability=not_implemented operation=elasticemail.post.suppressions-bounces]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post suppressions bounces import - Documented POST /suppressions/bounces/import (not implemented) [intent=direct_write availability=not_implemented operation=elasticemail.post.suppressions-bounces-import]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post suppressions complaints - Documented POST /suppressions/complaints (not implemented) [intent=direct_write availability=not_implemented operation=elasticemail.post.suppressions-complaints]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post suppressions complaints import - Documented POST /suppressions/complaints/import (not implemented) [intent=direct_write availability=not_implemented operation=elasticemail.post.suppressions-complaints-import]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post suppressions unsubscribes - Documented POST /suppressions/unsubscribes (not implemented) [intent=direct_write availability=not_implemented operation=elasticemail.post.suppressions-unsubscribes]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post suppressions unsubscribes import - Documented POST /suppressions/unsubscribes/import (not implemented) [intent=direct_write availability=not_implemented operation=elasticemail.post.suppressions-unsubscribes-import]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post verifications email - Documented POST /verifications/{email} (not implemented) [intent=direct_write availability=not_implemented operation=elasticemail.post.verifications-email]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post verifications files - Documented POST /verifications/files (not implemented) [intent=direct_write availability=not_implemented operation=elasticemail.post.verifications-files]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post verifications files id verification - Documented POST /verifications/files/{id}/verification (not implemented) [intent=direct_write availability=not_implemented operation=elasticemail.post.verifications-files-id-verification]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api put domains domain - Documented PUT /domains/{domain} (not implemented) [intent=direct_write availability=not_implemented operation=elasticemail.put.domains-domain]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api put domains domain verification - Documented PUT /domains/{domain}/verification (not implemented) [intent=direct_write availability=not_implemented operation=elasticemail.put.domains-domain-verification]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api put inboundroute order - Documented PUT /inboundroute/order (not implemented) [intent=direct_write availability=not_implemented operation=elasticemail.put.inboundroute-order]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api put security apikeys name - Documented PUT /security/apikeys/{name} (not implemented) [intent=direct_write availability=not_implemented operation=elasticemail.put.security-apikeys-name]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api put security smtp name - Documented PUT /security/smtp/{name} (not implemented) [intent=direct_write availability=not_implemented operation=elasticemail.put.security-smtp-name]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api put subaccounts email settings email - Documented PUT /subaccounts/{email}/settings/email (not implemented) [intent=direct_write availability=not_implemented operation=elasticemail.put.subaccounts-email-settings-email]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    campaigns list - Run the campaigns ETL stream [intent=etl availability=implemented stream=campaigns]
    contacts list - Run the contacts ETL stream [intent=etl availability=implemented stream=contacts]
    create campaign apply - Plan and execute the create campaign reverse-ETL action [intent=reverse_etl availability=not_implemented write=create_campaign]; approval: requires plan, preview, approval, and execute; risk: creates a new campaign targeting the given lists/segments; depending on Options this may schedule a live send to real recipients, not a preview-only action; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    create contact apply - Plan and execute the create contact reverse-ETL action [intent=reverse_etl availability=implemented write=create_contact]; approval: requires plan, preview, approval, and execute; risk: adds a new contact to the account's overall recipient list; low-risk external mutation, no approval required; flags: --Email (required)
    create domain apply - Plan and execute the create domain reverse-ETL action [intent=reverse_etl availability=implemented write=create_domain]; approval: requires plan, preview, approval, and execute; risk: registers a new sending domain pending DNS verification; low-risk external mutation, no approval required; flags: --Domain (required)
    create inbound route apply - Plan and execute the create inbound route reverse-ETL action [intent=reverse_etl availability=implemented write=create_inbound_route]; approval: requires plan, preview, approval, and execute; risk: creates a new inbound-mail routing rule that forwards matching inbound email to an external address or webhook URL of the caller's choosing; flags: --ActionType (required), --Filter (required), --FilterType (required), --Name (required)
    create list apply - Plan and execute the create list reverse-ETL action [intent=reverse_etl availability=implemented write=create_list]; approval: requires plan, preview, approval, and execute; risk: creates a new contact list, optionally seeding it from existing contact emails; low-risk external mutation, no approval required; flags: --ListName (required)
    create segment apply - Plan and execute the create segment reverse-ETL action [intent=reverse_etl availability=implemented write=create_segment]; approval: requires plan, preview, approval, and execute; risk: creates a new dynamic contact segment from a SQL-like rule; low-risk external mutation, no approval required; flags: --Name (required), --Rule (required)
    create template apply - Plan and execute the create template reverse-ETL action [intent=reverse_etl availability=implemented write=create_template]; approval: requires plan, preview, approval, and execute; risk: creates a new email template; low-risk external mutation, no approval required; flags: --Name (required)
    create webhook apply - Plan and execute the create webhook reverse-ETL action [intent=reverse_etl availability=implemented write=create_webhook]; approval: requires plan, preview, approval, and execute; risk: registers a new outbound webhook that will POST live event data (sent/opened/clicked/bounced) to an external URL of the caller's choosing; verify the target endpoint before enabling; flags: --Name (required), --URL (required)
    delete campaign apply - Plan and execute the delete campaign reverse-ETL action [intent=reverse_etl availability=implemented write=delete_campaign]; approval: requires plan, preview, approval, and execute; risk: permanently removes a campaign; if it has not finished sending, any remaining scheduled deliveries are cancelled; flags: --Name (required)
    delete contact apply - Plan and execute the delete contact reverse-ETL action [intent=reverse_etl availability=implemented write=delete_contact]; approval: requires plan, preview, approval, and execute; risk: permanently removes a contact and its activity/consent history from the account; flags: --Email (required)
    delete domain apply - Plan and execute the delete domain reverse-ETL action [intent=reverse_etl availability=implemented write=delete_domain]; approval: requires plan, preview, approval, and execute; risk: permanently removes a verified sending domain; any campaign configured to send from this domain will fail until reconfigured; flags: --Domain (required)
    delete inbound route apply - Plan and execute the delete inbound route reverse-ETL action [intent=reverse_etl availability=implemented write=delete_inbound_route]; approval: requires plan, preview, approval, and execute; risk: permanently removes an inbound-mail routing rule; matching inbound mail is no longer forwarded once removed; flags: --PublicId (required)
    delete list apply - Plan and execute the delete list reverse-ETL action [intent=reverse_etl availability=implemented write=delete_list]; approval: requires plan, preview, approval, and execute; risk: permanently removes a contact list; any campaign still targeting this list by name will fail to resolve its recipients; flags: --ListName (required)
    delete segment apply - Plan and execute the delete segment reverse-ETL action [intent=reverse_etl availability=implemented write=delete_segment]; approval: requires plan, preview, approval, and execute; risk: permanently removes a segment; any campaign still targeting this segment by name will fail to resolve its recipients; flags: --Name (required)
    delete template apply - Plan and execute the delete template reverse-ETL action [intent=reverse_etl availability=implemented write=delete_template]; approval: requires plan, preview, approval, and execute; risk: permanently removes a template; any campaign still referencing this template by name will fail to build its content; flags: --Name (required)
    delete webhook apply - Plan and execute the delete webhook reverse-ETL action [intent=reverse_etl availability=implemented write=delete_webhook]; approval: requires plan, preview, approval, and execute; risk: permanently removes a webhook subscription; event delivery to its target URL stops immediately; flags: --WebhookID (required)
    domains list - Run the domains ETL stream [intent=etl availability=implemented stream=domains]
    files list - Run the files ETL stream [intent=etl availability=implemented stream=files]
    inbound routes list - Run the inbound routes ETL stream [intent=etl availability=implemented stream=inbound_routes]
    lists list - Run the lists ETL stream [intent=etl availability=implemented stream=lists]
    pause campaign apply - Plan and execute the pause campaign reverse-ETL action [intent=reverse_etl availability=implemented write=pause_campaign]; approval: requires plan, preview, approval, and execute; risk: pauses an in-progress campaign send; recipients not yet reached will not receive the email until the campaign is resumed; flags: --Name (required)
    segments list - Run the segments ETL stream [intent=etl availability=implemented stream=segments]
    statistics campaigns list - Run the statistics campaigns ETL stream [intent=etl availability=implemented stream=statistics_campaigns]
    statistics channels list - Run the statistics channels ETL stream [intent=etl availability=implemented stream=statistics_channels]
    sub accounts list - Run the sub accounts ETL stream [intent=etl availability=implemented stream=sub_accounts]
    suppressions bounces list - Run the suppressions bounces ETL stream [intent=etl availability=implemented stream=suppressions_bounces]
    suppressions complaints list - Run the suppressions complaints ETL stream [intent=etl availability=implemented stream=suppressions_complaints]
    suppressions list - Run the suppressions ETL stream [intent=etl availability=implemented stream=suppressions]
    suppressions unsubscribes list - Run the suppressions unsubscribes ETL stream [intent=etl availability=implemented stream=suppressions_unsubscribes]
    templates list - Run the templates ETL stream [intent=etl availability=implemented stream=templates]
    update campaign apply - Plan and execute the update campaign reverse-ETL action [intent=reverse_etl availability=implemented write=update_campaign]; approval: requires plan, preview, approval, and execute; risk: mutates an existing campaign's content, recipients, or send options; a campaign already in progress may not accept every field change; flags: --Name (required)
    update contact apply - Plan and execute the update contact reverse-ETL action [intent=reverse_etl availability=implemented write=update_contact]; approval: requires plan, preview, approval, and execute; risk: mutates an existing contact's status/name/custom-field data; a Status change (e.g. to Unsubscribed) changes future campaign eligibility for this recipient; flags: --Email (required)
    update inbound route apply - Plan and execute the update inbound route reverse-ETL action [intent=reverse_etl availability=implemented write=update_inbound_route]; approval: requires plan, preview, approval, and execute; risk: mutates an existing inbound route's match filter or forwarding destination; redirects future matching inbound mail to a different address/URL; flags: --PublicId (required)
    update list apply - Plan and execute the update list reverse-ETL action [intent=reverse_etl availability=implemented write=update_list]; approval: requires plan, preview, approval, and execute; risk: renames an existing list or changes its unsubscribe-allowed setting; a rename changes the identifier campaigns/segments reference this list by; flags: --ListName (required)
    update segment apply - Plan and execute the update segment reverse-ETL action [intent=reverse_etl availability=implemented write=update_segment]; approval: requires plan, preview, approval, and execute; risk: changes the membership rule of an existing segment; immediately changes which contacts any campaign targeting this segment will reach; flags: --Name (required), --Rule (required)
    update template apply - Plan and execute the update template reverse-ETL action [intent=reverse_etl availability=implemented write=update_template]; approval: requires plan, preview, approval, and execute; risk: overwrites the subject/body of an existing template; any campaign referencing this template by name sends the new content on its next send; flags: --Name (required)
    update webhook apply - Plan and execute the update webhook reverse-ETL action [intent=reverse_etl availability=implemented write=update_webhook]; approval: requires plan, preview, approval, and execute; risk: mutates an existing webhook's target URL or event subscriptions; a changed URL redirects future event deliveries to a different endpoint; flags: --WebhookID (required)
    webhooks list - Run the webhooks ETL stream [intent=etl availability=implemented stream=webhooks]

EXAMPLES
  # Inspect as a manual
  pm connectors inspect elasticemail

  # Inspect as structured JSON
  pm connectors inspect elasticemail --json

AGENT WORKFLOW
  - Run pm connectors inspect elasticemail before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.
  - For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```

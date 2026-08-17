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
  mode
  api_key (secret)

ETL STREAMS
  contacts:
    primary key: Email
    cursor: DateUpdated
    fields: Activity(), Consent(), CustomFields(), DateAdded(), DateUpdated(), Email(), FirstName(), LastName(), Source(), Status(), StatusChangeDate()
  campaigns:
    primary key: Name
    fields: Content(), Name(), Options(), Recipients(), Status()
  lists:
    primary key: ListName
    fields: AllowUnsubscribe(), DateAdded(), ListName(), PublicListID()
  segments:
    primary key: Name
    fields: Name(), Rule()
  templates:
    primary key: Name
    fields: Body(), DateAdded(), Name(), Subject(), TemplateScope()
  domains:
    primary key: Domain
    fields: CertificateStatus(), CustomBouncesDomain(), DMARC(), DefaultDomain(), Dkim(), Domain(), IsMarkedForDeletion(), MX(), Spf(), TrackingStatus(), VERP(), Verify()
  suppressions:
    primary key: Email
    fields: DateUpdated(), Email(), ErrorCode(), FriendlyErrorMessage()
  suppressions_bounces:
    primary key: Email
    fields: DateUpdated(), Email(), ErrorCode(), FriendlyErrorMessage()
  suppressions_complaints:
    primary key: Email
    fields: DateUpdated(), Email(), ErrorCode(), FriendlyErrorMessage()
  suppressions_unsubscribes:
    primary key: Email
    fields: DateUpdated(), Email(), ErrorCode(), FriendlyErrorMessage()
  webhooks:
    primary key: WebhookID
    fields: DateCreated(), DateUpdated(), IsEnabled(), Name(), NotificationForAbuseReport(), NotificationForClicked(), NotificationForError(), NotificationForOpened(), NotificationForSent(), NotificationForUnsubscribed(), NotifyOncePerEmail(), URL(), WebhookID()
  files:
    primary key: FileName
    fields: ContentType(), DateAdded(), ExpirationDate(), FileName(), Size()
  inbound_routes:
    primary key: PublicId
    fields: ActionParameter(), ActionType(), Filter(), FilterType(), Name(), PublicId(), SortOrder()
  sub_accounts:
    primary key: PublicAccountID
    fields: ContactsCount(), Email(), EmailCredits(), LastActivity(), PublicAccountID(), Reputation(), Status(), TotalEmailsSent()
  statistics_campaigns:
    primary key: ChannelName
    fields: Bounced(), ChannelName(), Clicked(), Complaints(), Delivered(), EmailTotal(), InProgress(), Inbound(), ManualCancel(), NotDelivered(), Opened(), Recipients(), SmsTotal(), Unsubscribed()
  statistics_channels:
    primary key: ChannelName
    fields: Bounced(), ChannelName(), Clicked(), Complaints(), Delivered(), EmailTotal(), InProgress(), Inbound(), ManualCancel(), NotDelivered(), Opened(), Recipients(), SmsTotal(), Unsubscribed()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped, incremental_append, incremental_append_deduped

REVERSE ETL ACTIONS
  create_contact:
    endpoint: POST /contacts
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
    required fields: ListName
    optional fields: Emails, Status
    risk: adds existing contacts to a list, making them eligible recipients for any campaign targeting that list
  create_segment:
    endpoint: POST /segments
    risk: creates a new dynamic contact segment from a SQL-like rule; low-risk external mutation, no approval required
  update_segment:
    endpoint: PUT /segments/{{ record.Name }}
    required fields: Name
    risk: changes the membership rule of an existing segment; immediately changes which contacts any campaign targeting this segment will reach
  delete_segment:
    endpoint: DELETE /segments/{{ record.Name }}
    required fields: Name
    risk: permanently removes a segment; any campaign still targeting this segment by name will fail to resolve its recipients
  create_template:
    endpoint: POST /templates
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
    risk: registers a new sending domain pending DNS verification; low-risk external mutation, no approval required
  delete_domain:
    endpoint: DELETE /domains/{{ record.Domain }}
    required fields: Domain
    risk: permanently removes a verified sending domain; any campaign configured to send from this domain will fail until reconfigured
  create_inbound_route:
    endpoint: POST /inboundroute
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

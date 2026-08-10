# pm connectors inspect campaign-monitor

```text
NAME
  pm connectors inspect campaign-monitor - Campaign Monitor connector manual

SYNOPSIS
  pm connectors inspect campaign-monitor
  pm connectors inspect campaign-monitor --json
  pm credentials add <name> --connector campaign-monitor [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads and writes Campaign Monitor clients, campaigns, subscriber lists, subscribers, segments, and templates through the createsend.com v3.3 REST API.

ICON
  id: simple-icons-campaignmonitor
  asset: icons/simple-icons/campaignmonitor.svg
  title: Campaign Monitor
  simple_icon_slug: campaignmonitor
  simple_icon_hex: 111324
  source: simple-icons
  license: CC0-1.0
  review_status: cc0_with_trademark_caveat
  review_url: https://simpleicons.org/?q=Campaign%20Monitor
  match: exact-name-or-slug
  matched_by: campaign-monitor

CAPABILITIES
  check=true catalog=true read=true write=true query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  base_url
  client_id
  username (required)
  password (secret)

ETL STREAMS
  clients:
    primary key: ClientID
    fields: ClientID(string), Name(string)
  campaigns:
    primary key: CampaignID
    cursor: SentDate
    fields: CampaignID(string), FromEmail(string), FromName(string), Name(string), ReplyTo(string), SentDate(string), Subject(string), TotalRecipients(integer), WebVersionTextURL(string), WebVersionURL(string)
  lists:
    primary key: ListID
    fields: ListID(string), Name(string)
  suppressionlist:
    primary key: EmailAddress
    cursor: Date
    fields: Date(string), EmailAddress(string), State(string), SuppressionType(string)
  segments:
    primary key: SegmentID
    fields: ListID(string), OwningClientID(string), SegmentID(string), Title(string)
  templates:
    primary key: TemplateID
    fields: Name(string), OwningClientID(string), PreviewURL(string), ScreenshotURL(string), TemplateID(string)
  list_custom_fields:
    primary key: ListID, Key
    fields: DataType(string), FieldName(string), FieldOptions(array), Key(string), ListID(string), VisibleInPreferenceCenter(boolean)
  list_webhooks:
    primary key: WebhookID
    fields: Events(array), ListID(string), PayloadFormat(string), Status(string), Url(string), WebhookID(string)
  active_subscribers:
    primary key: ListID, EmailAddress
    fields: ConsentToTrack(string), CustomFields(array), Date(string), EmailAddress(string), ListID(string), ListJoinedDate(string), Name(string), ReadsEmailWith(string), State(string)
  unconfirmed_subscribers:
    primary key: ListID, EmailAddress
    fields: ConsentToTrack(string), CustomFields(array), Date(string), EmailAddress(string), ListID(string), ListJoinedDate(string), Name(string), ReadsEmailWith(string), State(string)
  unsubscribed_subscribers:
    primary key: ListID, EmailAddress
    fields: ConsentToTrack(string), CustomFields(array), Date(string), EmailAddress(string), ListID(string), ListJoinedDate(string), Name(string), ReadsEmailWith(string), State(string)
  bounced_subscribers:
    primary key: ListID, EmailAddress
    fields: ConsentToTrack(string), CustomFields(array), Date(string), EmailAddress(string), ListID(string), ListJoinedDate(string), Name(string), ReadsEmailWith(string), State(string)
  deleted_subscribers:
    primary key: ListID, EmailAddress
    fields: ConsentToTrack(string), CustomFields(array), Date(string), EmailAddress(string), ListID(string), ListJoinedDate(string), Name(string), ReadsEmailWith(string), State(string)

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped, incremental_append, incremental_append_deduped

REVERSE ETL ACTIONS
  create_list:
    endpoint: POST /lists/{{ config.client_id }}.json
    required fields: Title
    risk: creates a new subscriber list under the configured client; low-risk external mutation, no approval required
  update_list:
    endpoint: PUT /lists/{{ record.ListID }}.json
    required fields: ListID
    risk: updates list settings; enabling ScrubActiveWithSuppList/AddUnsubscribesToSuppList changes subscriber state for existing contacts; low-risk external mutation, no approval required
  delete_list:
    endpoint: DELETE /lists/{{ record.ListID }}.json
    required fields: ListID
    risk: permanently removes a subscriber list and all of its subscribers/segments; irreversible, approval required
  add_subscriber:
    endpoint: POST /subscribers/{{ record.ListID }}.json
    required fields: ListID, EmailAddress
    risk: adds a new subscriber to a list; low-risk external mutation, no approval required
  update_subscriber:
    endpoint: PUT /subscribers/{{ record.ListID }}.json?email={{ record.CurrentEmailAddress | urlencode }}
    required fields: ListID, CurrentEmailAddress, EmailAddress
    risk: updates an existing subscriber's profile/consent fields on a list, identified by their current email (CurrentEmailAddress, kept out of the body via path_fields since the API takes it as a query param, not a body field); low-risk external mutation, no approval required
  unsubscribe_subscriber:
    endpoint: POST /subscribers/{{ record.ListID }}/unsubscribe.json
    required fields: ListID, EmailAddress
    risk: unsubscribes a contact from a list; low-risk external mutation, no approval required
  delete_subscriber:
    endpoint: DELETE /subscribers/{{ record.ListID }}.json?email={{ record.EmailAddress | urlencode }}
    required fields: ListID, EmailAddress
    risk: permanently removes a subscriber's record from a list (distinct from unsubscribing — this deletes the record entirely); irreversible, approval recommended
  create_segment:
    endpoint: POST /segments/{{ record.ListID }}.json
    required fields: ListID, Title, RuleGroups
    risk: creates a new subscriber segment (a saved rule-based filter) on a list; low-risk external mutation, no approval required
  update_segment:
    endpoint: PUT /segments/{{ record.SegmentID }}.json
    required fields: SegmentID, Title, RuleGroups
    risk: replaces a segment's name and full rule set; low-risk external mutation, no approval required
  delete_segment:
    endpoint: DELETE /segments/{{ record.SegmentID }}.json
    required fields: SegmentID
    risk: permanently removes a segment; any campaign scheduled to send to it loses that targeting; irreversible, low-risk, no approval required
  create_campaign:
    endpoint: POST /campaigns/{{ config.client_id }}.json
    required fields: Name, Subject, FromName, FromEmail, ReplyTo, HtmlUrl, ListIDs
    risk: creates a new DRAFT campaign under the configured client; drafts are not sent until send_campaign is separately invoked, so this alone has no delivery side effect; low-risk, no approval required
  send_campaign:
    endpoint: POST /campaigns/{{ record.CampaignID }}/send.json
    required fields: CampaignID, ConfirmationEmail, SendDate
    risk: delivers a real email campaign to every subscriber on its targeted lists/segments; irreversible once sent, approval required
  unschedule_campaign:
    endpoint: POST /campaigns/{{ record.CampaignID }}/unschedule.json
    required fields: CampaignID
    risk: cancels a scheduled-but-not-yet-sent campaign, reverting it to a draft; low-risk, no approval required
  delete_campaign:
    endpoint: DELETE /campaigns/{{ record.CampaignID }}.json
    required fields: CampaignID
    risk: permanently removes a draft or sent campaign's record from Campaign Monitor; irreversible, approval recommended

SECURITY
  read risk: external Campaign Monitor API read of client, campaign, subscriber, segment, and template data
  write risk: external mutation of Campaign Monitor lists, subscribers (add/update/unsubscribe/delete), segments, and draft campaigns; sending a campaign delivers real email to real recipients
  approval: required for campaign-send and subscriber-delete actions; list/segment/subscriber-add/update mutations are lower risk
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect campaign-monitor

  # Inspect as structured JSON
  pm connectors inspect campaign-monitor --json

AGENT WORKFLOW
  - Run pm connectors inspect campaign-monitor before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.
  - For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```

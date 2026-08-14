# pm connectors inspect sparkpost

```text
NAME
  pm connectors inspect sparkpost - SparkPost connector manual

SYNOPSIS
  pm connectors inspect sparkpost
  pm connectors inspect sparkpost --json
  pm credentials add <name> --connector sparkpost [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads SparkPost recipient lists, templates, sending domains, transmissions, suppression list records, IP pools, webhooks, subaccounts, tracking domains, inbound domains, relay webhooks, sending IPs, and A/B tests; writes email sends, recipient list/template/domain/suppression/IP-pool/webhook/subaccount/relay-webhook lifecycle mutations.

ICON
  id: simple-icons-sparkpost
  asset: icons/simple-icons/sparkpost.svg
  title: SparkPost
  simple_icon_slug: sparkpost
  simple_icon_hex: FA6423
  source: simple-icons
  license: CC0-1.0
  review_status: cc0_with_trademark_caveat
  review_url: https://simpleicons.org/?q=SparkPost
  match: exact-name-or-slug
  matched_by: sparkpost

CAPABILITIES
  check=true catalog=true read=true write=true query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  base_url
  end_date
  start_date
  api_key (secret) (required)

ETL STREAMS
  recipient_lists:
    primary key: id
    fields: attributes(object), description(string), id(string), name(string), total_accepted_recipients(integer)
  templates:
    primary key: id
    fields: has_draft(boolean), has_published(boolean), id(string), last_update_time(string), last_use(string), name(string), published(boolean)
  sending_domains:
    primary key: domain
    fields: domain(string), is_default_bounce_domain(boolean), shared_with_subaccounts(boolean), status(object), tracking_domain(string)
  transmissions:
    primary key: id
    fields: campaign_id(string), description(string), generation_end_time(string), generation_start_time(string), id(string), num_failed_gen(integer), num_generated(integer), num_rcpts(integer), state(string)
  suppression_list:
    primary key: recipient
    fields: created(string), description(string), list_id(string), non_transactional(boolean), recipient(string), source(string), transactional(boolean), type(string), updated(string)
  ip_pools:
    primary key: id
    fields: auto_warmup_overflow_pool(string), fbl_signing_domain(string), id(string), ips(array), name(string), signing_domain(string)
  webhooks:
    primary key: id
    fields: active(boolean), auth_type(string), events(array), id(string), name(string), target(string)
  subaccounts:
    primary key: id
    fields: compliance_status(string), id(integer), ip_pool(string), name(string), status(string)
  tracking_domains:
    primary key: domain
    fields: default(boolean), domain(string), port(integer), secure(boolean), status(object), subaccount_id(integer)
  inbound_domains:
    primary key: domain
    fields: domain(string)
  relay_webhooks:
    primary key: id
    fields: auth_type(string), id(string), match(object), name(string), target(string)
  sending_ips:
    primary key: external_ip
    fields: auto_warmup_enabled(boolean), auto_warmup_stage(integer), customer_provided(boolean), external_ip(string), hostname(string), ip_pool(string)
  ab_tests:
    primary key: id
    fields: created_at(string), id(string), metric(string), name(string), status(string), updated_at(string), version(integer)
  account:
    primary key: customer_id
    fields: company_name(string), country_code(string), created(string), customer_id(integer), status(string), updated(string)

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite

REVERSE ETL ACTIONS
  update_account:
    endpoint: PUT /account
    risk: external mutation; changes account-wide settings (company name, two-factor requirement, default tracking/transactional options) affecting every sender on the account; approval required
  create_transmission:
    endpoint: POST /transmissions
    required fields: recipients, content
    risk: external mutation; sends real email to every listed recipient through the connected SparkPost account; approval required
  create_recipient_list:
    endpoint: POST /recipient-lists
    required fields: recipients
    risk: external mutation; creates a stored recipient list; approval required
  create_template:
    endpoint: POST /templates
    required fields: content
    risk: external mutation; creates a message template (as a draft unless published); approval required
  update_template:
    endpoint: PUT /templates/{{ record.id }}
    required fields: id
    risk: external mutation; updates an existing message template's draft/published content; approval required
  delete_template:
    endpoint: DELETE /templates/{{ record.id }}
    required fields: id
    risk: external mutation; permanently deletes a message template; approval required
  create_sending_domain:
    endpoint: POST /sending-domains
    required fields: domain
    risk: external mutation; registers a new sending domain pending DNS verification; approval required
  update_sending_domain:
    endpoint: PUT /sending-domains/{{ record.domain }}
    required fields: domain
    risk: external mutation; changes an existing sending domain's DKIM/tracking/bounce configuration; approval required
  delete_sending_domain:
    endpoint: DELETE /sending-domains/{{ record.domain }}
    required fields: domain
    risk: external mutation; permanently removes a sending domain; approval required
  create_or_update_suppression:
    endpoint: PUT /suppression-list/{{ record.recipient }}
    required fields: recipient, type
    risk: external mutation; adds or updates a recipient's suppression (opt-out) entry, affecting future deliverability to that address; approval required
  delete_suppression:
    endpoint: DELETE /suppression-list/{{ record.recipient }}
    required fields: recipient
    risk: external mutation; removes a recipient's suppression entry, re-enabling delivery to that address; approval required
  create_ip_pool:
    endpoint: POST /ip-pools
    required fields: name
    risk: external mutation; creates a dedicated IP pool; approval required
  update_ip_pool:
    endpoint: PUT /ip-pools/{{ record.id }}
    required fields: id, name
    risk: external mutation; changes an IP pool's DKIM signing domain / auto-warmup overflow configuration; approval required
  delete_ip_pool:
    endpoint: DELETE /ip-pools/{{ record.id }}
    required fields: id
    risk: external mutation; permanently deletes an IP pool; approval required
  create_webhook:
    endpoint: POST /webhooks
    required fields: name, target, events
    risk: external mutation; creates a webhook that will POST live event batches to an externally-supplied URL; a test POST is sent to target immediately; approval required
  update_webhook:
    endpoint: PUT /webhooks/{{ record.id }}
    required fields: id
    risk: external mutation; changes an existing webhook's target/events/auth; a test POST is sent to a new target immediately; approval required
  delete_webhook:
    endpoint: DELETE /webhooks/{{ record.id }}
    required fields: id
    risk: external mutation; permanently deletes a webhook; approval required
  create_subaccount:
    endpoint: POST /subaccounts
    required fields: name
    risk: external mutation; provisions a new subaccount, optionally with a live API key; approval required
  update_subaccount:
    endpoint: PUT /subaccounts/{{ record.id }}
    required fields: id
    risk: external mutation; changes a subaccount's name/status/ip_pool -- status transitions (e.g. to suspended/terminated) directly affect that subaccount's ability to send mail; approval required
  create_tracking_domain:
    endpoint: POST /tracking-domains
    required fields: domain
    risk: external mutation; registers a new tracking domain pending DNS verification; approval required
  delete_tracking_domain:
    endpoint: DELETE /tracking-domains/{{ record.domain }}
    required fields: domain
    risk: external mutation; permanently removes a tracking domain; approval required
  create_inbound_domain:
    endpoint: POST /inbound-domains
    required fields: domain
    risk: external mutation; registers a new inbound (receiving) domain; approval required
  delete_inbound_domain:
    endpoint: DELETE /inbound-domains/{{ record.domain }}
    required fields: domain
    risk: external mutation; permanently removes an inbound domain, stopping inbound relay of mail addressed to it; approval required
  create_relay_webhook:
    endpoint: POST /relay-webhooks
    required fields: target, match
    risk: external mutation; creates a relay webhook that will POST live inbound-mail batches to an externally-supplied URL; approval required
  update_relay_webhook:
    endpoint: PUT /relay-webhooks/{{ record.id }}
    required fields: id
    risk: external mutation; changes an existing relay webhook's target/match/auth; approval required
  delete_relay_webhook:
    endpoint: DELETE /relay-webhooks/{{ record.id }}
    required fields: id
    risk: external mutation; permanently deletes a relay webhook; approval required

SECURITY
  read risk: external SparkPost API read of recipient list, template, sending domain, transmission, suppression list, IP pool, webhook, subaccount, tracking/inbound domain, relay webhook, sending IP, and A/B test data
  write risk: external SparkPost API mutation including live email sends (create_transmission), suppression/domain/webhook/subaccount lifecycle changes, and IP pool/relay webhook configuration
  approval: required for all write actions; create_transmission sends real, externally-visible email
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect sparkpost

  # Inspect as structured JSON
  pm connectors inspect sparkpost --json

AGENT WORKFLOW
  - Run pm connectors inspect sparkpost before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.
  - For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```

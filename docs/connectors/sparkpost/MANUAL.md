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
  end_date
  start_date
  api_key (secret)

ETL STREAMS
  recipient_lists:
    primary key: id
    fields: attributes(), description(), id(), name(), total_accepted_recipients()
  templates:
    primary key: id
    fields: has_draft(), has_published(), id(), last_update_time(), last_use(), name(), published()
  sending_domains:
    primary key: domain
    fields: domain(), is_default_bounce_domain(), shared_with_subaccounts(), status(), tracking_domain()
  transmissions:
    primary key: id
    fields: campaign_id(), description(), generation_end_time(), generation_start_time(), id(), num_failed_gen(), num_generated(), num_rcpts(), state()
  suppression_list:
    primary key: recipient
    fields: created(), description(), list_id(), non_transactional(), recipient(), source(), transactional(), type(), updated()
  ip_pools:
    primary key: id
    fields: auto_warmup_overflow_pool(), fbl_signing_domain(), id(), ips(), name(), signing_domain()
  webhooks:
    primary key: id
    fields: active(), auth_type(), events(), id(), name(), target()
  subaccounts:
    primary key: id
    fields: compliance_status(), id(), ip_pool(), name(), status()
  tracking_domains:
    primary key: domain
    fields: default(), domain(), port(), secure(), status(), subaccount_id()
  inbound_domains:
    primary key: domain
    fields: domain()
  relay_webhooks:
    primary key: id
    fields: auth_type(), id(), match(), name(), target()
  sending_ips:
    primary key: external_ip
    fields: auto_warmup_enabled(), auto_warmup_stage(), customer_provided(), external_ip(), hostname(), ip_pool()
  ab_tests:
    primary key: id
    fields: created_at(), id(), metric(), name(), status(), updated_at(), version()
  account:
    primary key: customer_id
    fields: company_name(), country_code(), created(), customer_id(), status(), updated()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

REVERSE ETL ACTIONS
  update_account:
    endpoint: PUT /account
    risk: external mutation; changes account-wide settings (company name, two-factor requirement, default tracking/transactional options) affecting every sender on the account; approval required
  create_transmission:
    endpoint: POST /transmissions
    risk: external mutation; sends real email to every listed recipient through the connected SparkPost account; approval required
  create_recipient_list:
    endpoint: POST /recipient-lists
    risk: external mutation; creates a stored recipient list; approval required
  create_template:
    endpoint: POST /templates
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
    required fields: recipient
    risk: external mutation; adds or updates a recipient's suppression (opt-out) entry, affecting future deliverability to that address; approval required
  delete_suppression:
    endpoint: DELETE /suppression-list/{{ record.recipient }}
    required fields: recipient
    risk: external mutation; removes a recipient's suppression entry, re-enabling delivery to that address; approval required
  create_ip_pool:
    endpoint: POST /ip-pools
    risk: external mutation; creates a dedicated IP pool; approval required
  update_ip_pool:
    endpoint: PUT /ip-pools/{{ record.id }}
    required fields: id
    risk: external mutation; changes an IP pool's DKIM signing domain / auto-warmup overflow configuration; approval required
  delete_ip_pool:
    endpoint: DELETE /ip-pools/{{ record.id }}
    required fields: id
    risk: external mutation; permanently deletes an IP pool; approval required
  create_webhook:
    endpoint: POST /webhooks
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
    risk: external mutation; provisions a new subaccount, optionally with a live API key; approval required
  update_subaccount:
    endpoint: PUT /subaccounts/{{ record.id }}
    required fields: id
    risk: external mutation; changes a subaccount's name/status/ip_pool -- status transitions (e.g. to suspended/terminated) directly affect that subaccount's ability to send mail; approval required
  create_tracking_domain:
    endpoint: POST /tracking-domains
    risk: external mutation; registers a new tracking domain pending DNS verification; approval required
  delete_tracking_domain:
    endpoint: DELETE /tracking-domains/{{ record.domain }}
    required fields: domain
    risk: external mutation; permanently removes a tracking domain; approval required
  create_inbound_domain:
    endpoint: POST /inbound-domains
    risk: external mutation; registers a new inbound (receiving) domain; approval required
  delete_inbound_domain:
    endpoint: DELETE /inbound-domains/{{ record.domain }}
    required fields: domain
    risk: external mutation; permanently removes an inbound domain, stopping inbound relay of mail addressed to it; approval required
  create_relay_webhook:
    endpoint: POST /relay-webhooks
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

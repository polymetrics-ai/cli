---
name: pm-bitly
description: Bitly connector knowledge and safe action guide.
---

# pm-bitly

## Purpose

Reads Bitly organizations, groups, campaigns, channels, bitlinks, branded short domains, webhooks, QR codes, and group tags, and writes bitlink/campaign/group/channel/webhook/custom-bitlink/QR-code mutations, through the Bitly v4 REST API.

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
- group_guid
- api_key (secret)

## ETL Streams

- organizations:
  - primary key: guid
  - fields: created(), guid(), is_active(), modified(), name(), role(), tier(), tier_display_name(), tier_family()
- groups:
  - primary key: guid
  - fields: bsds(), created(), guid(), is_active(), modified(), name(), organization_guid(), role()
- campaigns:
  - primary key: guid
  - fields: channel_guids(), created(), description(), group_guid(), guid(), modified(), name()
- channels:
  - primary key: guid
  - fields: bitlinks(), campaign_guid(), created(), group_guid(), guid(), modified(), name()
- bsds:
  - primary key: account
  - fields: account(), bsds()
- webhooks:
  - primary key: guid
  - fields: campaign_guid(), client_id(), created(), event(), group_guid(), guid(), is_active(), modified(), updated_by(), url()
- qr_codes:
  - primary key: qrcode_id
  - fields: archived(), bitlink_id(), created(), created_by(), destination(), expiration_at(), group_guid(), is_customized(), modified(), qr_code_type(), qrcode_id(), tags(), title()
- group_tags:
  - primary key: group_guid
  - fields: group_guid(), tags()
- bitlinks:
  - primary key: id
  - fields: archived(), created_at(), deeplinks(), id(), link(), long_url(), references(), tags(), title()

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

## Reverse ETL Actions

- create_bitlink:
  - endpoint: POST /bitlinks
  - risk: creates a new publicly-resolvable short link; low-risk external mutation, no approval required
- update_bitlink:
  - endpoint: PATCH /bitlinks/{{ record.id }}
  - required fields: id
  - risk: mutates an existing bitlink's metadata or redirect destination; changing long_url redirects all future traffic on that short link and consumes an encode-limit unit
- delete_bitlink:
  - endpoint: DELETE /bitlinks/{{ record.id }}
  - required fields: id
  - risk: permanently removes a bitlink; any traffic still hitting the short URL starts failing to resolve
- update_bitlink_tags:
  - endpoint: PATCH /bitlinks/{{ record.id }}/tags
  - required fields: id
  - optional fields: tags
  - risk: replaces the full tag set on a bitlink; overwrites any tags not included in the submitted list
- delete_bitlink_tags:
  - endpoint: DELETE /bitlinks/{{ record.id }}/tags
  - required fields: id
  - optional fields: tags
  - risk: removes the named tags from a bitlink; irreversible without re-adding them via update_bitlink_tags
- create_campaign:
  - endpoint: POST /campaigns
  - risk: creates a new campaign container in the target group; low-risk external mutation, no approval required
- update_campaign:
  - endpoint: PATCH /campaigns/{{ record.guid }}
  - required fields: guid
  - risk: mutates an existing campaign's name, description, or associated channels
- update_group:
  - endpoint: PATCH /groups/{{ record.guid }}
  - required fields: guid
  - risk: renames or re-parents an existing group; a visible change for every member of that group
- update_group_preferences:
  - endpoint: PATCH /groups/{{ record.group_guid }}/preferences
  - required fields: group_guid
  - optional fields: domain_preference
  - risk: changes the default branded short domain new bitlinks in this group are created with
- create_channel:
  - endpoint: POST /channels
  - risk: creates a new channel container; low-risk external mutation, no approval required
- update_channel:
  - endpoint: PATCH /channels/{{ record.guid }}
  - required fields: guid
  - risk: mutates an existing channel's name or campaign association
- create_webhook:
  - endpoint: POST /webhooks
  - risk: registers a new outbound webhook that will POST live event data (clicks/scans) to an external URL of the caller's choosing; verify the target endpoint before enabling
- update_webhook:
  - endpoint: PATCH /webhooks/{{ record.guid }}
  - required fields: guid
  - risk: mutates an existing webhook's target URL, event type, or active state; a changed url redirects future event deliveries to a different endpoint
- delete_webhook:
  - endpoint: DELETE /webhooks/{{ record.guid }}
  - required fields: guid
  - risk: permanently removes a webhook subscription; event delivery to its target URL stops immediately
- create_custom_bitlink:
  - endpoint: POST /custom_bitlinks
  - risk: claims a custom keyword/back-half on a branded short domain and points it at a bitlink; consumes a finite custom-bitlink allocation on the domain
- update_custom_bitlink:
  - endpoint: PATCH /custom_bitlinks/{{ record.custom_bitlink }}
  - required fields: custom_bitlink
  - optional fields: bitlink_id
  - risk: re-points an existing custom keyword at a different bitlink; redirects all future traffic hitting that custom URL to the new destination
- create_qr_code:
  - endpoint: POST /qr-codes
  - risk: creates a new QR code resource pointed at a bitlink or long_url; low-risk external mutation, no approval required
- update_qr_code:
  - endpoint: PATCH /qr-codes/{{ record.qrcode_id }}
  - required fields: qrcode_id
  - risk: mutates an existing QR code's title or destination; changing destination redirects anyone scanning an already-printed/distributed code
- delete_qr_code:
  - endpoint: DELETE /qr-codes/{{ record.qrcode_id }}
  - required fields: qrcode_id
  - risk: permanently removes a QR code resource; any already-printed/distributed copy of the code stops resolving

## Security

- read risk: external Bitly API read of organization, group, campaign, channel, bitlink, branded-short-domain, webhook, QR code, and tag data
- write risk: external mutation of bitlinks, campaigns, groups, channels, webhooks, custom bitlinks, and QR codes; create_webhook/update_webhook register or repoint an outbound event delivery URL of the caller's choosing and warrant review before use, every write ships with an explicit per-action risk string
- approval: required for destructive actions (delete_bitlink, delete_webhook, delete_qr_code, delete_bitlink_tags) and for webhook URL mutations; creates/updates of bitlinks, campaigns, channels, and QR codes are low-risk
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect bitly
```

### Inspect as structured JSON

```bash
pm connectors inspect bitly --json
```

## Agent Rules

- Run pm connectors inspect bitly before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
- For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.

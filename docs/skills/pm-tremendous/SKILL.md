---
name: pm-tremendous
description: Tremendous connector knowledge and safe action guide.
---

# pm-tremendous

## Purpose

Reads and writes Tremendous campaigns, orders, rewards, funding sources, products, invoices, and members through the Tremendous API.

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

- base_url
- api_key (secret) (required)

## ETL Streams

- campaigns:
  - primary key: id
  - fields: created_at(string), id(string), name(string)
- orders:
  - primary key: id
  - fields: campaign_id(string), created_at(string), id(string), payment_status(string)
- rewards:
  - primary key: id
  - fields: created_at(string), id(string), order_id(string), status(string)
- funding_sources:
  - primary key: id
  - fields: created_at(string), id(string), name(string)
- products:
  - primary key: id
  - fields: category(string), countries(array), currency_codes(array), description(string), disclosure(string), documents(array), id(string), images(array), name(string), skus(array), subcategory(string), usage_instructions(string)
- invoices:
  - primary key: id
  - cursor: created_at
  - fields: amount(number), created_at(string), currency_code(string), id(string), international(boolean), orders(array), paid_at(string), po_number(string), rewards(array), status(string)
- members:
  - primary key: id
  - fields: active(boolean), email(string), id(string), name(string), role(string), status(string)

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

## Reverse ETL Actions

- create_order:
  - endpoint: POST /api/v2/orders
  - required fields: payment, reward
  - risk: spends real funding-source balance to issue a gift card / prepaid card / donation reward to a recipient; external mutation with real financial impact, approval required
- approve_order:
  - endpoint: POST /api/v2/order_approvals/{{ record.id }}/approve
  - required fields: id
  - risk: approves an order pending admin review, releasing its rewards for delivery; real financial impact, approval required
- reject_order:
  - endpoint: POST /api/v2/order_approvals/{{ record.id }}/reject
  - required fields: id
  - risk: rejects an order pending admin review; the order's rewards are never delivered
- cancel_reward:
  - endpoint: POST /api/v2/rewards/{{ record.id }}/cancel
  - required fields: id
  - risk: cancels and refunds a reward; only valid for non-expired rewards with a delivery failure per Tremendous's own API contract
- resend_reward:
  - endpoint: POST /api/v2/rewards/{{ record.id }}/resend
  - required fields: id
  - optional fields: updated_email, updated_phone
  - risk: resends a reward to its recipient (optionally at a new email/phone); only valid for rewards with a previous delivery failure
- generate_reward_link:
  - endpoint: POST /api/v2/rewards/{{ record.id }}/generate_link
  - required fields: id
  - risk: generates a new redemption link for an existing LINK-delivery reward; low-risk, does not move funds
- create_invoice:
  - endpoint: POST /api/v2/invoices
  - required fields: amount
  - risk: creates an invoice that funds the organization's Tremendous balance once paid; low direct risk (a document, not a payment itself), no approval required
- delete_invoice:
  - endpoint: DELETE /api/v2/invoices/{{ record.id }}
  - required fields: id
  - risk: removes an invoice; per Tremendous's own docs this is a cosmetic operation with no further financial consequence (an already-paid invoice's funds are unaffected)
- create_member:
  - endpoint: POST /api/v2/members
  - required fields: email, role
  - risk: invites a new user to manage the Tremendous organization (funding sources, campaigns, orders); grants organization access, approval required
- create_webhook:
  - endpoint: POST /api/v2/webhooks
  - required fields: url
  - risk: registers/replaces the organization's single webhook endpoint; a changed url redirects all future event deliveries to a different endpoint (Tremendous allows exactly one webhook per organization)
- delete_webhook:
  - endpoint: DELETE /api/v2/webhooks/{{ record.id }}
  - required fields: id
  - risk: permanently removes the organization's webhook subscription; event delivery stops immediately

## Security

- read risk: external Tremendous API read of campaign, order, reward, funding source, product, invoice, and member data
- write risk: external mutation with real financial impact: create_order spends funding-source balance to issue rewards; approve_order/reject_order/cancel_reward/resend_reward act on already-issued rewards; create_invoice/delete_invoice/create_member/create_webhook/delete_webhook are organization-administration mutations
- approval: create_order, approve_order, reject_order, cancel_reward, and create_member move money or grant organization access and require approval; create_invoice, delete_invoice, resend_reward, generate_reward_link, create_webhook, and delete_webhook are lower-risk/reversible-adjacent actions that execute without approval
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect tremendous
```

### Inspect as structured JSON

```bash
pm connectors inspect tremendous --json
```

## Agent Rules

- Run pm connectors inspect tremendous before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
- For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.

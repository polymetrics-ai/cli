---
name: pm-tremendous
description: Tremendous connector knowledge and safe action guide.
---

# pm-tremendous

## Purpose

Reads and writes Tremendous campaigns, orders, rewards, funding sources, products, invoices, and members through the Tremendous API.

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
- api_key (secret)

## ETL Streams

- campaigns:
  - primary key: id
  - fields: created_at(), id(), name()
- orders:
  - primary key: id
  - fields: campaign_id(), created_at(), id(), payment_status()
- rewards:
  - primary key: id
  - fields: created_at(), id(), order_id(), status()
- funding_sources:
  - primary key: id
  - fields: created_at(), id(), name()
- products:
  - primary key: id
  - fields: category(), countries(), currency_codes(), description(), disclosure(), documents(), id(), images(), name(), skus(), subcategory(), usage_instructions()
- invoices:
  - primary key: id
  - cursor: created_at
  - fields: amount(), created_at(), currency_code(), id(), international(), orders(), paid_at(), po_number(), rewards(), status()
- members:
  - primary key: id
  - fields: active(), email(), id(), name(), role(), status()

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

## Reverse ETL Actions

- create_order:
  - endpoint: POST /api/v2/orders
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
  - risk: creates an invoice that funds the organization's Tremendous balance once paid; low direct risk (a document, not a payment itself), no approval required
- delete_invoice:
  - endpoint: DELETE /api/v2/invoices/{{ record.id }}
  - required fields: id
  - risk: removes an invoice; per Tremendous's own docs this is a cosmetic operation with no further financial consequence (an already-paid invoice's funds are unaffected)
- create_member:
  - endpoint: POST /api/v2/members
  - risk: invites a new user to manage the Tremendous organization (funding sources, campaigns, orders); grants organization access, approval required
- create_webhook:
  - endpoint: POST /api/v2/webhooks
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

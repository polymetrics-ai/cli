---
name: pm-dwolla
description: Dwolla connector knowledge and safe action guide.
---

# pm-dwolla

## Purpose

Reads Dwolla customers, events, exchange partners, and business classifications, and writes customer/funding-source/transfer/webhook-subscription/beneficial-owner lifecycle mutations, via the Dwolla HAL+JSON REST API using OAuth2 client-credentials.

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
- client_id (secret)
- client_secret (secret)

## ETL Streams

- customers:
  - primary key: id
  - cursor: created
  - fields: businessName(), created(), email(), firstName(), id(), lastName(), status(), type()
- events:
  - primary key: id
  - cursor: created
  - fields: created(), id(), resourceId(), topic()
- exchange_partners:
  - primary key: id
  - cursor: created
  - fields: created(), id(), name(), status()
- business_classifications:
  - primary key: id
  - fields: id(), name()

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

## Reverse ETL Actions

- create_customer:
  - endpoint: POST /customers
  - risk: external mutation; creates a new Dwolla customer (personal, business, receive-only, or unverified), subject to Dwolla's identity-verification rules for the requested type
- update_customer:
  - endpoint: POST /customers/{{ record.id }}
  - required fields: id
  - risk: external mutation; updates a customer's profile fields, or its status (e.g. deactivating/reactivating the customer, which blocks/restores its ability to transact)
- create_funding_source:
  - endpoint: POST /customers/{{ record.customer_id }}/funding-sources
  - required fields: customer_id
  - risk: external mutation; attaches a new bank-account funding source to a customer, either as unverified (routingNumber/accountNumber, requiring later micro-deposit verification) or pre-verified via an open-banking plaidToken/onDemandAuthorizationId
- update_funding_source:
  - endpoint: POST /funding-sources/{{ record.id }}
  - required fields: id
  - risk: external mutation; renames a funding source or replaces its unverified bank-account routing/account numbers
- remove_funding_source:
  - endpoint: POST /funding-sources/{{ record.id }}
  - required fields: id
  - optional fields: removed
  - risk: destructive external mutation; Dwolla has no hard-delete for funding sources, this soft-removes it (POST {removed:true}) so it can no longer send/receive transfers; not reversible via the API
- initiate_micro_deposits:
  - endpoint: POST /funding-sources/{{ record.id }}/initiate-micro-deposits
  - required fields: id
  - risk: external mutation; sends two small trial-deposit ACH transactions to an unverified bank-account funding source, the first step of micro-deposit verification
- verify_micro_deposits:
  - endpoint: POST /funding-sources/{{ record.id }}/verify-micro-deposits
  - required fields: id
  - optional fields: amount1, amount2
  - risk: external mutation; verifies a funding source's micro-deposit amounts, completing bank-account verification; Dwolla locks the funding source after repeated failed attempts
- cancel_transfer:
  - endpoint: POST /transfers/{{ record.id }}
  - required fields: id
  - optional fields: status
  - risk: external mutation; cancels a still-pending transfer before it clears, this action is not reversible and only succeeds while the transfer's status is pending
- create_webhook_subscription:
  - endpoint: POST /webhook-subscriptions
  - risk: external mutation; registers a new webhook subscription; Dwolla enforces a maximum of 10 active subscriptions in Sandbox and 5 in Production
- update_webhook_subscription:
  - endpoint: POST /webhook-subscriptions/{{ record.id }}
  - required fields: id
  - optional fields: paused
  - risk: external mutation; pauses or resumes webhook delivery for a subscription (Dwolla still generates the events while paused, it just withholds delivery)
- delete_webhook_subscription:
  - endpoint: DELETE /webhook-subscriptions/{{ record.id }}
  - required fields: id
  - risk: destructive external mutation; permanently deletes a webhook subscription and stops all future event delivery to it; not reversible
- create_beneficial_owner:
  - endpoint: POST /customers/{{ record.customer_id }}/beneficial-owners
  - required fields: customer_id
  - risk: external mutation; registers a beneficial owner (25%+ equity holder) for a business verified customer; submits sensitive PII (SSN, date of birth, address) to Dwolla for identity verification
- update_beneficial_owner:
  - endpoint: POST /beneficial-owners/{{ record.id }}
  - required fields: id
  - risk: external mutation; updates a beneficial owner's identity/PII fields, which resets its verification status pending re-review
- remove_beneficial_owner:
  - endpoint: DELETE /beneficial-owners/{{ record.id }}
  - required fields: id
  - risk: destructive external mutation; permanently removes a beneficial owner from a business verified customer; not reversible and may affect the customer's beneficial-ownership certification status
- certify_beneficial_ownership:
  - endpoint: POST /customers/{{ record.customer_id }}/beneficial-ownership
  - required fields: customer_id
  - optional fields: status
  - risk: external mutation; the Account Admin attests that a business customer's beneficial-owner information is accurate and complete, which is required before the customer can transact

## Security

- read risk: external Dwolla API read of customer, event, exchange-partner, and business-classification data
- write risk: external mutation of Dwolla customers, funding sources, transfers (cancel only, never create/send), webhook subscriptions, and beneficial owners; several actions are destructive/not reversible (remove_funding_source, cancel_transfer, delete_webhook_subscription, remove_beneficial_owner) and beneficial-owner actions carry SSN/PII; approval required
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect dwolla
```

### Inspect as structured JSON

```bash
pm connectors inspect dwolla --json
```

## Agent Rules

- Run pm connectors inspect dwolla before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
- For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.

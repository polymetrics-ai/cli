---
name: pm-stripe
description: Stripe connector knowledge and safe action guide.
---

# pm-stripe

## Purpose

Reads Stripe customers, charges, invoices, subscriptions, and products, and writes approved reverse ETL customer create, update, and typed destructive delete actions through the Stripe REST API.

## Icon

- asset: icons/stripe.svg
- source: upstream_registry
- review_status: upstream_seeded
- review_url: https://stripe.com/docs/api

## Capabilities

- check=true catalog=true read=true write=true query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- account_id
- base_url
- max_pages
- mode
- page_size
- start_date
- client_secret (secret)

## ETL Streams

- customers:
  - primary key: id
  - cursor: created
  - fields: balance(), created(), currency(), delinquent(), description(), email(), id(), livemode(), name(), object(), phone()
- charges:
  - primary key: id
  - cursor: created
  - fields: amount(), amount_captured(), amount_refunded(), created(), currency(), customer(), id(), livemode(), object(), paid(), refunded(), status()
- invoices:
  - primary key: id
  - cursor: created
  - fields: amount_due(), amount_paid(), amount_remaining(), created(), currency(), customer(), id(), livemode(), object(), paid(), status(), subscription(), total()
- subscriptions:
  - primary key: id
  - cursor: created
  - fields: cancel_at_period_end(), canceled_at(), created(), currency(), current_period_end(), current_period_start(), customer(), id(), livemode(), object(), status()
- products:
  - primary key: id
  - cursor: created
  - fields: active(), created(), description(), id(), livemode(), name(), object(), type(), updated()

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped, incremental_append, incremental_append_deduped

## Reverse ETL Actions

- create_customer:
  - endpoint: POST /customers
  - risk: external mutation; approval required
- update_customer:
  - endpoint: POST /customers/{{ record.id }}
  - required fields: id
  - risk: external mutation; approval required
- delete_customer:
  - endpoint: DELETE /customers/{{ record.id }}
  - required fields: id
  - risk: destructive external mutation; deletes a Stripe customer; reverse ETL approval and typed destructive confirmation required

## Security

- read risk: external Stripe API read of customer and billing data
- write risk: external Stripe API mutation, including typed destructive customer deletion when explicitly confirmed
- approval: reverse ETL plan approval required before writes; destructive actions require typed confirmation
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Command Surface

- Read Stripe billing streams and safely plan customer write actions.
- Usage: pm stripe <command> [flags]
- Source CLI: Stripe API (OpenAPI spec3 2026-07-29.dahlia)
- Global flags:
  - --credential (string): Credential name to use for the Stripe request.
  - --connection (string): Credential name alias used only when --credential is omitted; does not resolve pm connections.
  - --config (string_array): Connector config override as key=value; never pass secret values here.
  - --json (boolean): Emit machine-readable JSON output.
  - --limit (integer): Maximum records to emit from stream commands.
  - --plan (string): Execute an approved reverse-ETL plan by id.
  - --preview (boolean): Preview a reverse-ETL write command without making a network mutation.
  - --approve (string): Approval token required to execute a reverse-ETL plan.
  - --confirm (string): Typed confirmation challenge for destructive reverse-ETL writes.
- Customers
  - customers list - Read Stripe customers through the declared ETL stream. [intent=etl availability=implemented stream=customers]
  - customers create - Plan creation of a Stripe customer through reverse ETL. [intent=reverse_etl availability=implemented write=create_customer]; approval: Plan first, inspect preview output, then run only with the generated approval token.; risk: Creates customer data in Stripe; requires reverse ETL plan, preview, explicit approval, then execute.; flags: --email, --name, --description, --phone
  - customers update - Plan an update to a Stripe customer through reverse ETL. [intent=reverse_etl availability=implemented write=update_customer]; approval: Plan first, inspect preview output, then run only with the generated approval token.; risk: Mutates customer data in Stripe; requires reverse ETL plan, preview, explicit approval, then execute.; flags: --id, --email, --name, --description, --phone
  - customers delete - Plan deletion of a Stripe customer with typed destructive confirmation and customer ID redaction. [intent=reverse_etl availability=implemented write=delete_customer]; approval: Plan first, inspect preview output, then run only with the generated approval token and typed --confirm destructive challenge.; risk: Destructive Stripe customer deletion; redacts the customer ID from previews and write errors; requires reverse ETL plan, preview, explicit approval, and --confirm destructive before execute.; flags: --id
- Billing streams
  - charges list - Read Stripe charges through the declared ETL stream. [intent=etl availability=implemented stream=charges]
  - invoices list - Read Stripe invoices through the declared ETL stream. [intent=etl availability=implemented stream=invoices]
  - subscriptions list - Read Stripe subscriptions through the declared ETL stream. [intent=etl availability=implemented stream=subscriptions]
  - products list - Read Stripe products through the declared ETL stream. [intent=etl availability=implemented stream=products]
- Help topics:
  - destructive-confirmation - Stripe DELETE/destructive operations are in scope only through typed destructive confirmation and the reverse ETL plan → preview → approval → execute path.

## Commands

### Inspect as a manual

```bash
pm connectors inspect stripe
```

### Inspect as structured JSON

```bash
pm connectors inspect stripe --json
```

## Agent Rules

- Run pm connectors inspect stripe before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
- For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.

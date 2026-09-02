---
name: pm-freightview
description: Freightview connector knowledge and safe action guide.
---

# pm-freightview

## Purpose

Reads Freightview shipments, quotes, and tracking events through fixed Freightview v2.0 REST routes using client-credentials authentication.

## Icon

- id: pm-sample
- asset: icons/pm-sample.svg
- source: polymetrics
- review_status: polymetrics
- review_url: https://github.com/polymetrics-ai/cli

## Capabilities

- check=true catalog=true read=true write=false query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- client_id (secret) (required)
- client_secret (secret) (required)

## ETL Streams

- shipments:
  - primary key: shipmentId
  - fields: billTo(object), bol(object), bookedBy(string), bookedDate(string), createdDate(string), direction(string), documents(array), equipment(object), isArchived(boolean), isLiveLoad(boolean), items(array), locations(array), pickup(object), pickupDate(string), quotedBy(string), refNums(array), selectedQuote(object), shipmentId(string), status(string), tracking(object)
- quotes:
  - primary key: quoteId
  - fields: amount(number), carrierId(string), createdDate(string), currency(string), equipmentType(string), method(string), mode(string), paymentTerms(string), pricingMethod(string), pricingType(string), providerCode(string), providerName(string), quoteId(string), quoteNum(string), serviceId(string), source(string), status(string)
- tracking:
  - primary key: createdDate
  - fields: createdDate(string), eventDate(string), eventTime(string), eventType(string), summary(string)

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite

## Security

- read risk: Bounded Freightview v2.0 reads use declared client-credentials authentication and fixed provider routes.
- write risk: unsupported
- approval: none; read-only
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect freightview
```

### Inspect as structured JSON

```bash
pm connectors inspect freightview --json
```

## Agent Rules

- Run pm connectors inspect freightview before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.

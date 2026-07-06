---
name: pm-freightview
description: Freightview connector knowledge and safe action guide.
---

# pm-freightview

## Purpose

Reads Freightview shipments, quotes, and tracking events through the Freightview v2.0 REST API using the client-credentials session-token flow. In architecture v2 this quarantine bundle dispatches live reads through a Tier-2 hook that delegates to the legacy connector until the wave 6 cutover.

## Icon

- asset: icons/pm-sample.svg
- source: polymetrics
- review_status: polymetrics

## Capabilities

- check=true catalog=true read=true write=false query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- base_url
- mode
- client_id (secret)
- client_secret (secret)

## ETL Streams

- shipments:
  - primary key: shipmentId
  - fields: billTo(), bol(), bookedBy(), bookedDate(), createdDate(), direction(), documents(), equipment(), isArchived(), isLiveLoad(), items(), locations(), pickup(), pickupDate(), quotedBy(), refNums(), selectedQuote(), shipmentId(), status(), tracking()
- quotes:
  - primary key: quoteId
  - fields: amount(), carrierId(), createdDate(), currency(), equipmentType(), method(), mode(), paymentTerms(), pricingMethod(), pricingType(), providerCode(), providerName(), quoteId(), quoteNum(), serviceId(), source(), status()
- tracking:
  - primary key: createdDate
  - fields: createdDate(), eventDate(), eventTime(), eventType(), summary()

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

## Security

- read risk: external Freightview API reads performed by the legacy connector via a Tier-2 hook
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

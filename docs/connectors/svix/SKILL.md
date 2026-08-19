---
name: pm-svix
description: Svix connector knowledge and safe action guide.
---

# pm-svix

## Purpose

Reads Svix applications, endpoints, event types, messages, message delivery attempts, background tasks, connectors, and operational webhook endpoints, and writes application/endpoint/event-type/connector/operational-webhook-endpoint lifecycle mutations and outgoing messages, through the Svix REST API.

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

- applications:
  - primary key: id
  - fields: created_at(string), id(string), name(string)
- endpoints:
  - primary key: id
  - cursor: updated_at
  - fields: app_id(string), channels(array), created_at(string), description(string), disabled(boolean), filterTypes(array), id(string), metadata(object), rateLimit(integer), throttleRate(integer), uid(string), updated_at(string), url(string), version(integer)
- event_types:
  - primary key: name
  - cursor: updated_at
  - fields: archived(boolean), created_at(string), deprecated(boolean), description(string), featureFlags(array), groupName(string), name(string), schemas(object), updated_at(string)
- messages:
  - primary key: id
  - cursor: timestamp
  - fields: app_id(string), channels(array), eventId(string), eventType(string), id(string), payload(object), tags(array), timestamp(string)
- background_tasks:
  - primary key: id
  - cursor: updated_at
  - fields: data(object), id(string), status(string), task(string), updated_at(string)
- connectors:
  - primary key: id
  - cursor: updated_at
  - fields: allowedEventTypes(array), created_at(string), description(string), featureFlags(array), id(string), instructions(string), kind(string), logo(string), name(string), orgId(string), productType(string), transformation(string), uid(string), updated_at(string)
- operational_webhook_endpoints:
  - primary key: id
  - cursor: updated_at
  - fields: created_at(string), description(string), disabled(boolean), filterTypes(array), id(string), metadata(object), rateLimit(integer), throttleRate(integer), uid(string), updated_at(string), url(string)

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

## Reverse ETL Actions

- create_application:
  - endpoint: POST /app
  - required fields: name
  - risk: creates a new Svix application (a webhook-sending namespace); low-risk external mutation, no approval required
- update_application:
  - endpoint: PUT /app/{{ record.id }}
  - required fields: id, name
  - risk: replaces an existing application's metadata/name/throttle rate; external mutation, no approval required
- delete_application:
  - endpoint: DELETE /app/{{ record.id }}
  - required fields: id
  - risk: irreversibly deletes an application and all its endpoints, messages, and delivery history; approval required
- create_endpoint:
  - endpoint: POST /app/{{ record.app_id }}/endpoint
  - required fields: app_id, url
  - risk: creates a new webhook delivery endpoint on an application; the endpoint immediately starts receiving future events; low-risk external mutation, no approval required
- update_endpoint:
  - endpoint: PUT /app/{{ record.app_id }}/endpoint/{{ record.id }}
  - required fields: app_id, id, url
  - risk: replaces an existing endpoint's delivery URL/filters/disabled state; changing url redirects all future webhook deliveries for that endpoint; external mutation, no approval required
- delete_endpoint:
  - endpoint: DELETE /app/{{ record.app_id }}/endpoint/{{ record.id }}
  - required fields: app_id, id
  - risk: irreversibly deletes a webhook delivery endpoint and stops all future deliveries to it; approval required
- create_event_type:
  - endpoint: POST /event-type
  - required fields: name, description
  - risk: creates a new event type definition; low-risk external mutation, no approval required
- update_event_type:
  - endpoint: PUT /event-type/{{ record.name }}
  - required fields: name, description
  - risk: replaces an existing event type's description/schema/archived state; external mutation, no approval required
- delete_event_type:
  - endpoint: DELETE /event-type/{{ record.name }}
  - required fields: name
  - risk: archives (soft-deletes) an event type definition; approval required
- send_message:
  - endpoint: POST /app/{{ record.app_id }}/msg
  - required fields: app_id, eventType, payload
  - risk: sends a real outgoing webhook message that Svix immediately attempts to deliver to every matching endpoint on the application; approval required
- create_connector:
  - endpoint: POST /connector
  - required fields: name, transformation
  - risk: creates a new outgoing-webhook payload-transformation connector template; low-risk external mutation, no approval required
- update_connector:
  - endpoint: PUT /connector/{{ record.id }}
  - required fields: id, name, transformation
  - risk: replaces an existing connector's transformation JS/description; changes the payload shape delivered to every endpoint using this connector; external mutation, no approval required
- delete_connector:
  - endpoint: DELETE /connector/{{ record.id }}
  - required fields: id
  - risk: irreversibly deletes a connector transformation template; approval required
- create_operational_webhook_endpoint:
  - endpoint: POST /operational-webhook/endpoint
  - required fields: url
  - risk: creates a new operational webhook endpoint (Svix account-level events, e.g. message.attempt.exhausted); low-risk external mutation, no approval required
- update_operational_webhook_endpoint:
  - endpoint: PUT /operational-webhook/endpoint/{{ record.id }}
  - required fields: id, url
  - risk: replaces an existing operational webhook endpoint's delivery URL/filters/disabled state; external mutation, no approval required
- delete_operational_webhook_endpoint:
  - endpoint: DELETE /operational-webhook/endpoint/{{ record.id }}
  - required fields: id
  - risk: irreversibly deletes an operational webhook endpoint and stops all future account-level event deliveries to it; approval required

## Security

- read risk: external Svix API read of application, endpoint, message, and delivery-attempt data
- write risk: external Svix API mutation (application/endpoint/event-type/connector/operational-webhook-endpoint lifecycle, outgoing message creation)
- approval: reverse ETL plan approval required before writes
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect svix
```

### Inspect as structured JSON

```bash
pm connectors inspect svix --json
```

## Agent Rules

- Run pm connectors inspect svix before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
- For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.

# Overview

RD Station Marketing reads and writes the OAuth bearer Platform API under
`https://api.rd.services/platform`. The bundle keeps the 5 legacy streams from
`internal/connectors/rd-station-marketing` (`contacts`, `segmentations`, `events`,
`landing_pages`, `email_templates`) and adds documented Platform API streams for contact detail,
segmentation contacts, contact events, contact funnels, contact fields, analytics, and product
catalog feeds.

The legacy Go package remains read-only until the wave6 registry flip. The write actions here cover
documented JSON REST mutations that the engine can express without hooks.

## Auth setup

Provide an RD Station Marketing OAuth access token via the `access_token` secret. It is sent as
`Authorization: Bearer <access_token>`, matching legacy's `connsdk.Bearer(secret)`. `base_url`
defaults to `https://api.rd.services/platform`; keep that `/platform` base when overriding it for
tests or proxies.

Detail/fan-out streams use explicit config keys: `contact_identifier`,
`contact_identifier_value`, `contact_uuid`, `segmentation_id`, and `catalog_feed_id`.

## Streams notes

The legacy-parity list streams keep the legacy page-number shape (`page`, `page_size`, default
125). `contacts` and `events` retain catalog cursor fields without server-side incremental request
params, matching legacy's full-refresh reads with cursor metadata only.

New documented streams:

- `contact_detail`: `GET /contacts/{identifier}:{value}`.
- `segmentation_contacts`: `GET /segmentations/{id}/contacts`.
- `contact_conversion_events` and `contact_opportunity_events`: `GET /contacts/{uuid}/events`
  with the required event type fixed to `CONVERSION` or `OPPORTUNITY`.
- `contact_funnel`: `GET /contacts/{identifier}:{value}/funnels/default`.
- `contact_fields`: `GET /contacts/fields`.
- `analytics_conversions`, `analytics_emails`, `analytics_funnel`,
  `analytics_workflow_emails`: documented analytics GET endpoints.
- `catalog_feeds` and `catalog_feed`: product catalog feed list/detail.

Legacy fallback mapping is narrowed to the documented primary fields the bundle fixtures exercise:
`uuid` for list-stream IDs and `event_identifier` for contact-event IDs. The engine has no multi-key
fallback filter, so alternate legacy defensive keys such as raw `id`, `title`, and `event` are
documented as a known limit instead of being represented with unsupported template syntax.

## Write actions & risks

Write actions require the standard reverse-ETL plan, preview, approval, execute flow.

- `create_contact`, `update_contact`, `delete_contact`: create, mutate, or delete contacts.
- `add_contact_tags`: appends tags to an existing contact.
- `update_contact_funnel`: mutates lifecycle/opportunity metadata in the default contact funnel.
- `insert_workflow_leads`: inserts lead identifiers into a workflow.
- `create_contact_field`: creates custom contact fields.
- `create_catalog_feed`, `update_catalog_feed`, `delete_catalog_feed`: manage product catalog feeds.

Deletes are marked destructive. `create_contact` uses `minProperties: 1` because the docs express
email-or-phone semantics, while the engine's draft-07 subset has no `anyOf`.

## Known limits

- The API-key-only conversion endpoint `POST /conversions?api_key=...` is excluded. It uses a
  separate API-key auth path while this bundle is OAuth bearer scoped and the engine auth is
  connector-wide, not per-action.
- Webhook Service endpoints under `/integrations/webhooks` are excluded. Supporting them without
  breaking legacy `base_url=https://api.rd.services/platform` behavior would require per-stream or
  per-action base URLs, or a hook.
- OAuth event writes use `POST /events?event_type=...`. `writes.json` has no query parameter field,
  and embedding a query string in `path` is not preserved by write replay/request shape, so those
  writes are excluded rather than exposing an action that would omit `event_type`.
- Optional analytics filters such as date ranges and asset IDs are not declared because conformance
  supplies synthetic values for every spec key. Adding those filters cleanly would require either
  always-valid defaults or a fixture override mechanism.
- Legacy alternate-key fallback is narrowed for the expanded bundle. The legacy Go mapper tried
  multiple defensive raw keys for a few list streams; the declarative bundle uses the documented
  `uuid`, `name`, and `event_type` fields for those streams because the engine does not support a
  multi-key fallback filter.

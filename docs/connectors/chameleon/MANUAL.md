# pm connectors inspect chameleon

```text
NAME
  pm connectors inspect chameleon - Chameleon connector manual

SYNOPSIS
  pm connectors inspect chameleon
  pm connectors inspect chameleon --json
  pm credentials add <name> --connector chameleon [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Chameleon surveys, tours, launchers, tooltips, and segments through the Chameleon v3 REST API.

ICON
  id: pm-sample
  asset: icons/pm-sample.svg
  source: polymetrics
  review_status: polymetrics
  review_url: https://github.com/polymetrics-ai/cli

CAPABILITIES
  check=true catalog=true read=true write=true query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  base_url
  mode
  api_key (secret) (required)

ETL STREAMS
  surveys:
    primary key: id
    cursor: updated_at
    fields: created_at(string), id(string), is_live(boolean), state(string), title(string), type(string), updated_at(string)
  tours:
    primary key: id
    cursor: updated_at
    fields: created_at(string), id(string), is_live(boolean), state(string), title(string), type(string), updated_at(string)
  launchers:
    primary key: id
    cursor: updated_at
    fields: created_at(string), id(string), is_live(boolean), state(string), title(string), type(string), updated_at(string)
  tooltips:
    primary key: id
    cursor: updated_at
    fields: created_at(string), id(string), is_live(boolean), state(string), title(string), type(string), updated_at(string)
  segments:
    primary key: id
    cursor: updated_at
    fields: created_at(string), description(string), id(string), name(string), updated_at(string)
  embeds:
    primary key: id
    cursor: updated_at
    fields: archived_at(string), created_at(string), dashboard_url(string), description(string), id(string), name(string), position(integer), published_at(string), segment_ids(array), style(string), tag_ids(array), updated_at(string)
  event_names:
    primary key: id
    cursor: updated_at
    fields: created_at(string), dashboard_url(string), description(string), id(string), kind(string), last_seen_at(string), name(string), published_at(string), source(string), uid(string), updated_at(string)
  tags:
    primary key: id
    cursor: updated_at
    fields: created_at(string), description(string), disabled_at(string), id(string), last_seen_at(string), models_count(integer), name(string), uid(string), updated_at(string)
  deliveries:
    primary key: id
    cursor: updated_at
    fields: at(string), at_href(string), created_at(string), from(string), group_kind(string), id(string), idempotency_key(string), interaction_id(string), model_id(string), model_kind(string), once(boolean), options(object), profile_id(string), until(string), updated_at(string), use_segmentation(boolean)
  webhooks:
    primary key: id
    fields: id(string), last_item_at(string), last_item_error(string), last_item_state(string), name(string), uid(string)
  companies:
    primary key: id
    fields: clv(number), created_at(string), domain(string), id(string), plan(string), uid(string)

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped, incremental_append, incremental_append_deduped

REVERSE ETL ACTIONS
  publish_survey:
    endpoint: PATCH /edit/surveys/{{ record.id }}
    required fields: id, published_at
    risk: external mutation publishing/unpublishing a live in-product Microsurvey to end-users; approval required
  publish_tour:
    endpoint: PATCH /edit/tours/{{ record.id }}
    required fields: id, published_at
    risk: external mutation publishing/unpublishing a live in-product Tour to end-users; approval required
  publish_launcher:
    endpoint: PATCH /edit/launchers/{{ record.id }}
    required fields: id, published_at
    risk: external mutation publishing/unpublishing a live in-product Launcher to end-users; approval required
  publish_tooltip:
    endpoint: PATCH /edit/tooltips/{{ record.id }}
    required fields: id, published_at
    risk: external mutation publishing/unpublishing a live in-product Tooltip to end-users; approval required
  publish_embed:
    endpoint: PATCH /edit/embeds/{{ record.id }}
    required fields: id, published_at
    risk: external mutation publishing/unpublishing a live in-product Embeddable to end-users; approval required
  create_delivery:
    endpoint: POST /edit/deliveries
    required fields: model_kind, model_id
    risk: external mutation directly triggering a Tour or Microsurvey experience for one specific end-user; approval required
  delete_delivery:
    endpoint: DELETE /edit/deliveries/{{ record.id }}
    required fields: id
    risk: cancels a not-yet-triggered Delivery; irreversible once the target has already been shown, approval required
  create_webhook:
    endpoint: POST /edit/webhooks
    required fields: url, topics
    risk: external mutation creating a new outbound webhook subscription that will POST Chameleon event data to a third-party URL; approval required
  delete_webhook:
    endpoint: DELETE /edit/webhooks/{{ record.id }}
    required fields: id
    risk: irreversible removal of an outbound webhook subscription; approval required

SECURITY
  read risk: external Chameleon API read of in-product experience, segment, tag, event, delivery, webhook, and company data
  write risk: external mutations publishing/unpublishing in-product experiences, triggering/cancelling user-targeted Deliveries, and creating/deleting outbound Webhook subscriptions; every write action requires approval
  approval: read: none; write: required for every action
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect chameleon

  # Inspect as structured JSON
  pm connectors inspect chameleon --json

AGENT WORKFLOW
  - Run pm connectors inspect chameleon before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.
  - For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```

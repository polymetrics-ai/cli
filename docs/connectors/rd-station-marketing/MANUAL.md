# pm connectors inspect rd-station-marketing

```text
NAME
  pm connectors inspect rd-station-marketing - RD Station Marketing connector manual

SYNOPSIS
  pm connectors inspect rd-station-marketing
  pm connectors inspect rd-station-marketing --json
  pm credentials add <name> --connector rd-station-marketing [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads and writes RD Station Marketing platform contacts, segmentation contacts, analytics, contact fields, product catalog feeds, and workflows.

ICON
  id: rdstation
  asset: icons/rdstation.svg
  source: official
  review_status: official_verified
  review_url: https://developers.rdstation.com/

CAPABILITIES
  check=true catalog=true read=true write=true query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  base_url
  catalog_feed_id
  contact_identifier
  contact_identifier_value
  contact_uuid
  segmentation_id
  access_token (secret) (required)

ETL STREAMS
  contacts:
    primary key: id
    cursor: updated_at
    fields: created_at(string), email(string), id(string), name(string), updated_at(string)
  segmentations:
    primary key: id
    fields: created_at(string), id(string), name(string)
  events:
    primary key: id
    cursor: created_at
    fields: created_at(string), email(string), event_type(string), id(string)
  landing_pages:
    primary key: id
    fields: created_at(string), id(string), name(string)
  email_templates:
    primary key: id
    fields: created_at(string), id(string), name(string)
  contact_detail:
    primary key: id
    fields: bio(string), birthdate(string), city(string), country(string), created_at(string), email(string), facebook(string), id(string), job_title(string), legal_bases(array), linkedin(string), mobile_phone(string), name(string), personal_phone(string), phone(string), state(string), tags(array), twitter(string), updated_at(string), uuid(string), website(string)
  segmentation_contacts:
    primary key: id
    fields: created_at(string), email(string), id(string), last_conversion_date(string), links(array), name(string), phone(string), uuid(string)
  contact_conversion_events:
    primary key: id
    fields: email(string), event_family(string), event_identifier(string), event_timestamp(string), event_type(string), id(string), name(string), payload(object), phone(string)
  contact_opportunity_events:
    primary key: id
    fields: email(string), event_family(string), event_identifier(string), event_timestamp(string), event_type(string), id(string), name(string), payload(object), phone(string)
  contact_funnel:
    primary key: id
    fields: contact_owner_email(string), fit(integer), id(string), interest(integer), lifecycle_stage(string), opportunity(boolean)
  contact_fields:
    primary key: id
    fields: api_identifier(string), custom_field(boolean), data_type(string), id(string), label(object), name(object), presentation_type(string), uuid(string), validation_rules(object)
  analytics_conversions:
    primary key: asset_id
    fields: asset_created_at(string), asset_id(integer), asset_identifier(string), asset_type(string), asset_updated_at(string), conversion_rate(number), conversions_count(integer), visits_count(integer)
  analytics_emails:
    primary key: campaign_id
    fields: campaign_id(integer), campaign_name(string), contacts_count(integer), email_bounced_count(integer), email_clicked_count(integer), email_clicked_rate(number), email_delivered_count(integer), email_delivered_rate(number), email_dropped_count(integer), email_opened_count(integer), email_opened_rate(number), email_spam_reported_count(integer), email_spam_reported_rate(number), email_unsubscribed_count(integer), send_at(string)
  analytics_funnel:
    primary key: reference_day
    fields: contacts_count(integer), opportunities_count(integer), qualified_contacts_count(integer), reference_day(string), sales_count(integer), visitors_count(integer)
  analytics_workflow_emails:
    primary key: workflow_action_id
    fields: email_bounced_unique_count(integer), email_clicked_rate(number), email_clicked_unique_count(integer), email_delivered_count(integer), email_delivered_rate(number), email_dropped_count(integer), email_name(string), email_opened_rate(number), email_opened_unique_count(integer), email_spam_reported_count(integer), email_spam_reported_rate(number), email_unsubscribed_count(integer), workflow_action_id(string), workflow_created_at(string), workflow_id(string), workflow_name(string), workflow_updated_at(string)
  catalog_feeds:
    primary key: id
    fields: created_at(string), credentials(object), format(string), id(integer), name(string), status(string), updated_at(string), url(string)
  catalog_feed:
    primary key: id
    fields: created_at(string), credentials(object), format(string), id(integer), name(string), status(string), updated_at(string), url(string)

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped, incremental_append, incremental_append_deduped

REVERSE ETL ACTIONS
  create_contact:
    endpoint: POST /contacts
    risk: creates a contact in the RD Station Marketing lead base
  update_contact:
    endpoint: PATCH /contacts/{{ record.identifier }}:{{ record.value }}
    required fields: identifier, value
    risk: updates an existing RD Station Marketing contact
  delete_contact:
    endpoint: DELETE /contacts/{{ record.identifier }}:{{ record.value }}
    required fields: identifier, value
    risk: deletes a contact from the RD Station Marketing lead base
  add_contact_tags:
    endpoint: POST /contacts/{{ record.identifier }}:{{ record.value }}/tag
    required fields: identifier, value, tags
    risk: adds tags to an existing RD Station Marketing contact
  update_contact_funnel:
    endpoint: PUT /contacts/{{ record.identifier }}:{{ record.value }}/funnels/default
    required fields: identifier, value
    risk: updates lifecycle/opportunity ownership fields in the default contact funnel
  insert_workflow_leads:
    endpoint: POST /workflows/{{ record.workflow_id }}/leads
    required fields: workflow_id, leads
    risk: inserts one or more leads into a marketing automation workflow
  create_contact_field:
    endpoint: POST /contacts/fields
    required fields: api_identifier, data_type, label, name, presentation_type
    risk: creates a custom contact field in the RD Station Marketing account
  create_catalog_feed:
    endpoint: POST /catalog_feeds
    required fields: name, url, format
    risk: creates a product catalog feed configuration
  update_catalog_feed:
    endpoint: PATCH /catalog_feeds/{{ record.catalog_feed_id }}
    required fields: catalog_feed_id, name, url, format
    risk: updates a product catalog feed configuration
  delete_catalog_feed:
    endpoint: DELETE /catalog_feeds/{{ record.catalog_feed_id }}
    required fields: catalog_feed_id
    risk: deletes a product catalog feed configuration

SECURITY
  read risk: external RD Station Marketing API read of contact, campaign, analytics, field, workflow, and catalog-feed data
  write risk: creates, updates, and deletes RD Station Marketing contacts, contact fields, and catalog feeds; mutates contact funnels, tags, and workflow membership
  approval: reverse ETL writes require plan preview and approval token
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect rd-station-marketing

  # Inspect as structured JSON
  pm connectors inspect rd-station-marketing --json

AGENT WORKFLOW
  - Run pm connectors inspect rd-station-marketing before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.
  - For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```

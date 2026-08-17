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
  access_token (secret)

ETL STREAMS
  contacts:
    primary key: id
    cursor: updated_at
    fields: created_at(), email(), id(), name(), updated_at()
  segmentations:
    primary key: id
    fields: created_at(), id(), name()
  events:
    primary key: id
    cursor: created_at
    fields: created_at(), email(), event_type(), id()
  landing_pages:
    primary key: id
    fields: created_at(), id(), name()
  email_templates:
    primary key: id
    fields: created_at(), id(), name()
  contact_detail:
    primary key: id
    fields: bio(), birthdate(), city(), country(), created_at(), email(), facebook(), id(), job_title(), legal_bases(), linkedin(), mobile_phone(), name(), personal_phone(), phone(), state(), tags(), twitter(), updated_at(), uuid(), website()
  segmentation_contacts:
    primary key: id
    fields: created_at(), email(), id(), last_conversion_date(), links(), name(), phone(), uuid()
  contact_conversion_events:
    primary key: id
    fields: email(), event_family(), event_identifier(), event_timestamp(), event_type(), id(), name(), payload(), phone()
  contact_opportunity_events:
    primary key: id
    fields: email(), event_family(), event_identifier(), event_timestamp(), event_type(), id(), name(), payload(), phone()
  contact_funnel:
    primary key: id
    fields: contact_owner_email(), fit(), id(), interest(), lifecycle_stage(), opportunity()
  contact_fields:
    primary key: id
    fields: api_identifier(), custom_field(), data_type(), id(), label(), name(), presentation_type(), uuid(), validation_rules()
  analytics_conversions:
    primary key: asset_id
    fields: asset_created_at(), asset_id(), asset_identifier(), asset_type(), asset_updated_at(), conversion_rate(), conversions_count(), visits_count()
  analytics_emails:
    primary key: campaign_id
    fields: campaign_id(), campaign_name(), contacts_count(), email_bounced_count(), email_clicked_count(), email_clicked_rate(), email_delivered_count(), email_delivered_rate(), email_dropped_count(), email_opened_count(), email_opened_rate(), email_spam_reported_count(), email_spam_reported_rate(), email_unsubscribed_count(), send_at()
  analytics_funnel:
    primary key: reference_day
    fields: contacts_count(), opportunities_count(), qualified_contacts_count(), reference_day(), sales_count(), visitors_count()
  analytics_workflow_emails:
    primary key: workflow_action_id
    fields: email_bounced_unique_count(), email_clicked_rate(), email_clicked_unique_count(), email_delivered_count(), email_delivered_rate(), email_dropped_count(), email_name(), email_opened_rate(), email_opened_unique_count(), email_spam_reported_count(), email_spam_reported_rate(), email_unsubscribed_count(), workflow_action_id(), workflow_created_at(), workflow_id(), workflow_name(), workflow_updated_at()
  catalog_feeds:
    primary key: id
    fields: created_at(), credentials(), format(), id(), name(), status(), updated_at(), url()
  catalog_feed:
    primary key: id
    fields: created_at(), credentials(), format(), id(), name(), status(), updated_at(), url()

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
    required fields: identifier, value
    risk: adds tags to an existing RD Station Marketing contact
  update_contact_funnel:
    endpoint: PUT /contacts/{{ record.identifier }}:{{ record.value }}/funnels/default
    required fields: identifier, value
    risk: updates lifecycle/opportunity ownership fields in the default contact funnel
  insert_workflow_leads:
    endpoint: POST /workflows/{{ record.workflow_id }}/leads
    required fields: workflow_id
    risk: inserts one or more leads into a marketing automation workflow
  create_contact_field:
    endpoint: POST /contacts/fields
    risk: creates a custom contact field in the RD Station Marketing account
  create_catalog_feed:
    endpoint: POST /catalog_feeds
    risk: creates a product catalog feed configuration
  update_catalog_feed:
    endpoint: PATCH /catalog_feeds/{{ record.catalog_feed_id }}
    required fields: catalog_feed_id
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

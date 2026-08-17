# pm connectors inspect getgist

```text
NAME
  pm connectors inspect getgist - GetGist connector manual

SYNOPSIS
  pm connectors inspect getgist
  pm connectors inspect getgist --json
  pm credentials add <name> --connector getgist [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Gist contacts, tags, segments, campaigns, forms, teammates, articles, collections, conversations, teams, workspace metadata, and e-commerce resources through the Gist REST API; writes regular JSON Gist resources and relationship actions.

ICON
  asset: icons/pm-sample.svg
  source: polymetrics
  review_status: polymetrics

CAPABILITIES
  check=true catalog=true read=true write=true query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  article_id
  article_query
  base_url
  batch_id
  campaign_id
  category_id
  collection_id
  contact_email
  contact_id
  contact_user_id
  conversation_id
  customer_id
  event_type
  form_id
  include_count
  mode
  order_id
  product_id
  product_variant_id
  segment_id
  store_id
  subscription_type_id
  team_id
  teammate_id
  variant_id
  api_key (secret)

ETL STREAMS
  contacts:
    primary key: id
    cursor: updated_at
    fields: created_at(), email(), id(), last_contacted_at(), last_seen_at(), name(), phone(), session_count(), signed_up_at(), type(), unsubscribed_from_emails(), updated_at(), user_id()
  contact_details:
    primary key: id
    cursor: updated_at
    fields: created_at(), email(), id(), last_contacted_at(), last_seen_at(), name(), phone(), session_count(), signed_up_at(), type(), unsubscribed_from_emails(), updated_at(), user_id()
  contact_batch_status:
    primary key: id
    fields: created_at(), description(), id(), name(), status(), title(), type(), updated_at()
  articles:
    primary key: id
    fields: created_at(), description(), id(), name(), status(), title(), type(), updated_at()
  article_details:
    primary key: id
    fields: created_at(), description(), id(), name(), status(), title(), type(), updated_at()
  article_search:
    primary key: id
    fields: created_at(), description(), id(), name(), status(), title(), type(), updated_at()
  article_settings:
    primary key: id
    fields: created_at(), description(), id(), name(), status(), title(), type(), updated_at()
  collections:
    primary key: id
    fields: created_at(), description(), id(), name(), status(), title(), type(), updated_at()
  collection_details:
    primary key: id
    fields: created_at(), description(), id(), name(), status(), title(), type(), updated_at()
  events:
    primary key: id
    fields: created_at(), description(), id(), name(), status(), title(), type(), updated_at()
  tags:
    primary key: id
    fields: id(), name(), type()
  segments:
    primary key: id
    fields: count(), created_at(), id(), name(), person_type(), type(), updated_at()
  segment_details:
    primary key: id
    fields: count(), created_at(), id(), name(), person_type(), type(), updated_at()
  forms:
    primary key: id
    fields: created_at(), id(), name(), type(), updated_at()
  form_details:
    primary key: id
    fields: created_at(), id(), name(), type(), updated_at()
  form_submissions:
    primary key: id
    fields: created_at(), description(), id(), name(), status(), title(), type(), updated_at()
  campaigns:
    primary key: id
    fields: created_at(), id(), name(), status(), subject(), type(), updated_at()
  campaign_details:
    primary key: id
    fields: created_at(), id(), name(), status(), subject(), type(), updated_at()
  subscription_types:
    primary key: id
    fields: created_at(), description(), id(), name(), status(), title(), type(), updated_at()
  subscription_type_details:
    primary key: id
    fields: created_at(), description(), id(), name(), status(), title(), type(), updated_at()
  conversations:
    primary key: id
    fields: created_at(), description(), id(), name(), status(), title(), type(), updated_at()
  conversation_details:
    primary key: id
    fields: created_at(), description(), id(), name(), status(), title(), type(), updated_at()
  conversation_messages:
    primary key: id
    fields: created_at(), description(), id(), name(), status(), title(), type(), updated_at()
  conversation_global_counts:
    primary key: scope
    fields: closed(), count(), id(), open(), scope(), team_id(), teammate_id()
  conversation_team_counts:
    primary key: scope
    fields: closed(), count(), id(), open(), scope(), team_id(), teammate_id()
  conversation_teammate_counts:
    primary key: scope
    fields: closed(), count(), id(), open(), scope(), team_id(), teammate_id()
  teams:
    primary key: id
    fields: created_at(), description(), id(), name(), status(), title(), type(), updated_at()
  team_details:
    primary key: id
    fields: created_at(), description(), id(), name(), status(), title(), type(), updated_at()
  teammates:
    primary key: id
    fields: email(), id(), name(), type()
  teammate_details:
    primary key: id
    fields: email(), id(), name(), type()
  token_info:
    primary key: id
    fields: created_at(), description(), id(), name(), status(), title(), type(), updated_at()
  ecommerce_stores:
    primary key: id
    fields: created_at(), description(), id(), name(), status(), title(), type(), updated_at()
  ecommerce_store_details:
    primary key: id
    fields: created_at(), description(), id(), name(), status(), title(), type(), updated_at()
  ecommerce_customer_details:
    primary key: id
    fields: created_at(), description(), id(), name(), status(), title(), type(), updated_at()
  ecommerce_product_details:
    primary key: id
    fields: created_at(), description(), id(), name(), status(), title(), type(), updated_at()
  ecommerce_product_variant_details:
    primary key: id
    fields: created_at(), description(), id(), name(), status(), title(), type(), updated_at()
  ecommerce_product_category_details:
    primary key: id
    fields: created_at(), description(), id(), name(), status(), title(), type(), updated_at()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

REVERSE ETL ACTIONS
  create_article:
    endpoint: POST /articles
    risk: create through the Gist API
  update_article:
    endpoint: PATCH /articles/{{ record.article_id }}
    required fields: article_id
    risk: update through the Gist API
  delete_article:
    endpoint: DELETE /articles/{{ record.article_id }}
    required fields: article_id
    risk: delete through the Gist API
  create_collection:
    endpoint: POST /collections
    risk: create through the Gist API
  update_collection:
    endpoint: POST /collections/{{ record.collection_id }}
    required fields: collection_id
    risk: update through the Gist API
  delete_collection:
    endpoint: DELETE /collections/{{ record.collection_id }}
    required fields: collection_id
    risk: delete through the Gist API
  upsert_contact:
    endpoint: POST /contacts
    risk: upsert through the Gist API
  upsert_contacts_batch:
    endpoint: POST /contacts/batch
    risk: upsert through the Gist API
  delete_contact:
    endpoint: DELETE /contacts/{{ record.contact_id }}
    required fields: contact_id
    risk: delete through the Gist API
  track_event:
    endpoint: POST /events
    risk: create through the Gist API
  upsert_tag:
    endpoint: POST /tags
    risk: upsert through the Gist API
  delete_tag:
    endpoint: DELETE /tags/{{ record.tag_id }}
    required fields: tag_id
    risk: delete through the Gist API
  add_tag_to_contacts:
    endpoint: POST /tags
    risk: update through the Gist API
  remove_tag_from_contacts:
    endpoint: POST /tags
    risk: update through the Gist API
  subscribe_contact_to_form:
    endpoint: POST /forms/{{ record.form_id }}/subscribe
    required fields: form_id
    risk: update through the Gist API
  subscribe_contact_to_campaign:
    endpoint: POST /campaigns
    risk: update through the Gist API
  unsubscribe_contact_from_campaign:
    endpoint: POST /campaigns
    risk: update through the Gist API
  attach_contact_to_subscription_type:
    endpoint: POST /subscription_types/{{ record.subscription_type_id }}
    required fields: subscription_type_id
    risk: update through the Gist API
  detach_contact_from_subscription_type:
    endpoint: POST /subscription_types/{{ record.subscription_type_id }}
    required fields: subscription_type_id
    risk: update through the Gist API
  create_conversation:
    endpoint: POST /conversations
    risk: create through the Gist API
  update_conversation:
    endpoint: PATCH /conversations/{{ record.conversation_id }}
    required fields: conversation_id
    risk: update through the Gist API
  reply_to_conversation:
    endpoint: POST /conversations/{{ record.conversation_id }}/messages
    required fields: conversation_id
    risk: create through the Gist API
  delete_conversation:
    endpoint: DELETE /conversations/{{ record.conversation_id }}
    required fields: conversation_id
    risk: delete through the Gist API
  unassign_conversation:
    endpoint: PATCH /conversations/{{ record.conversation_id }}/assign
    required fields: conversation_id
    risk: update through the Gist API
  assign_conversation:
    endpoint: PATCH /conversations/{{ record.conversation_id }}/assign
    required fields: conversation_id
    risk: update through the Gist API
  snooze_conversation:
    endpoint: PATCH /conversations/{{ record.conversation_id }}
    required fields: conversation_id
    risk: update through the Gist API
  unsnooze_conversation:
    endpoint: PATCH /conversations/{{ record.conversation_id }}
    required fields: conversation_id
    risk: update through the Gist API
  close_conversation:
    endpoint: PATCH /conversations/{{ record.conversation_id }}
    required fields: conversation_id
    risk: update through the Gist API
  prioritize_conversation:
    endpoint: PATCH /conversations/{{ record.conversation_id }}/priority
    required fields: conversation_id
    risk: update through the Gist API
  tag_conversation:
    endpoint: POST /conversations/{{ record.conversation_id }}/tags
    required fields: conversation_id
    risk: update through the Gist API
  untag_conversation:
    endpoint: DELETE /conversations/{{ record.conversation_id }}/tags
    required fields: conversation_id
    risk: delete through the Gist API
  create_store:
    endpoint: POST /ecommerce/stores
    risk: create through the Gist API
  update_store:
    endpoint: PATCH /ecommerce/stores/{{ record.store_id }}
    required fields: store_id
    risk: update through the Gist API
  create_customer:
    endpoint: POST /ecommerce/stores/{{ record.store_id }}/customers
    required fields: store_id
    risk: create through the Gist API
  update_customer:
    endpoint: PATCH /ecommerce/stores/{{ record.store_id }}/customers/{{ record.customer_id }}
    required fields: store_id, customer_id
    risk: update through the Gist API
  create_product:
    endpoint: POST /ecommerce/stores/{{ record.store_id }}/products
    required fields: store_id
    risk: create through the Gist API
  update_product:
    endpoint: PATCH /ecommerce/stores/{{ record.store_id }}/products/{{ record.product_id }}
    required fields: store_id, product_id
    risk: update through the Gist API
  create_product_variant:
    endpoint: POST /ecommerce/stores/{{ record.store_id }}/products/{{ record.product_id }}/variants
    required fields: store_id, product_id
    risk: create through the Gist API
  update_product_variant:
    endpoint: PATCH /ecommerce/stores/{{ record.store_id }}/products/{{ record.product_id }}/variants/{{ record.product_variant_id }}
    required fields: store_id, product_id, product_variant_id
    risk: update through the Gist API
  create_product_category:
    endpoint: POST /ecommerce/stores/{{ record.store_id }}/categories
    required fields: store_id
    risk: create through the Gist API
  update_product_category:
    endpoint: PATCH /ecommerce/stores/{{ record.store_id }}/categories/{{ record.category_id }}
    required fields: store_id, category_id
    risk: update through the Gist API
  upsert_cart:
    endpoint: POST /ecommerce/stores/{{ record.store_id }}/customers/{{ record.customer_id }}/cart
    required fields: store_id, customer_id
    risk: upsert through the Gist API
  delete_cart:
    endpoint: DELETE /ecommerce/stores/{{ record.store_id }}/customers/{{ record.customer_id }}/cart
    required fields: store_id, customer_id
    risk: delete through the Gist API
  create_order:
    endpoint: POST /ecommerce/stores/{{ record.store_id }}/orders
    required fields: store_id
    risk: create through the Gist API
  update_order:
    endpoint: PATCH /ecommerce/stores/{{ record.store_id }}/orders/{{ record.order_id }}
    required fields: store_id, order_id
    risk: update through the Gist API

SECURITY
  read risk: external Gist API reads of workspace, contact, marketing, conversation, team, knowledge base, and e-commerce data
  write risk: creates, updates, deletes, tags, subscribes, assigns, replies to, and otherwise mutates Gist workspace resources
  approval: reverse ETL writes require plan preview and approval token
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect getgist

  # Inspect as structured JSON
  pm connectors inspect getgist --json

AGENT WORKFLOW
  - Run pm connectors inspect getgist before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.
  - For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```

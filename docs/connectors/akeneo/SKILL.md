---
name: pm-akeneo
description: Akeneo connector knowledge and safe action guide.
---

# pm-akeneo

## Purpose

Reads Akeneo PIM products, categories, families, attributes, channels, product models, family variants, attribute groups, association types, locales, currencies, and measure families, and writes create-or-update upserts for the 9 catalog-structure resources, through the Akeneo REST API (OAuth2 password grant).

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

- api_username (required)
- base_url (required)
- client_id (required)
- page_size
- password (secret)
- secret (secret)

## ETL Streams

- products:
  - primary key: id
  - fields: categories(array), created(string), enabled(boolean), family(string), groups(array), id(string), parent(string), updated(string), uuid(string), values(object)
- categories:
  - primary key: id
  - fields: id(string), labels(object), parent(string), updated(string)
- families:
  - primary key: id
  - fields: attribute_as_image(string), attribute_as_label(string), attributes(array), id(string), labels(object)
- attributes:
  - primary key: id
  - fields: group(string), id(string), labels(object), localizable(boolean), scopable(boolean), type(string)
- channels:
  - primary key: id
  - fields: category_tree(string), currencies(array), id(string), labels(object), locales(array)
- product_models:
  - primary key: id
  - fields: categories(array), created(string), family_variant(string), id(string), parent(string), updated(string), values(object)
- family_variants:
  - primary key: id
  - fields: attributes(array), id(string), labels(object), variant_attribute_sets(array)
- attribute_groups:
  - primary key: id
  - fields: attributes(array), id(string), labels(object), sort_order(integer)
- association_types:
  - primary key: id
  - fields: id(string), is_two_way(boolean), labels(object)
- locales:
  - primary key: id
  - fields: enabled(boolean), id(string)
- currencies:
  - primary key: id
  - fields: enabled(boolean), id(string)
- measure_families:
  - primary key: id
  - fields: id(string), standard_unit_code(string), units(object)

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite

## Reverse ETL Actions

- create_or_update_product:
  - endpoint: PATCH /api/rest/v1/products/{{ record.id }}
  - required fields: id
  - risk: creates a new product (201) or updates an existing one (204) in the connected Akeneo PIM catalog; visible to every downstream channel the product is enabled/categorized for
- create_or_update_category:
  - endpoint: PATCH /api/rest/v1/categories/{{ record.id }}
  - required fields: id
  - risk: creates or updates a category node; re-parenting an existing category changes the catalog tree for every product classified under it
- create_or_update_family:
  - endpoint: PATCH /api/rest/v1/families/{{ record.id }}
  - required fields: id
  - risk: creates or updates a product family definition; changes the required/optional attribute set for every product assigned to this family
- create_or_update_attribute:
  - endpoint: PATCH /api/rest/v1/attributes/{{ record.id }}
  - required fields: id
  - risk: creates a new attribute (schema mutation, affects every family/product referencing it) or updates an existing one's non-structural properties (labels/group); some attribute properties are immutable after creation per Akeneo's own API rules
- create_or_update_channel:
  - endpoint: PATCH /api/rest/v1/channels/{{ record.id }}
  - required fields: id
  - risk: creates or updates a distribution channel; changes which locales/currencies/category tree every product exported to this channel uses
- create_or_update_product_model:
  - endpoint: PATCH /api/rest/v1/product-models/{{ record.id }}
  - required fields: id
  - risk: creates or updates a product model; a shared parent for variant products, changes propagate to every variant beneath it
- create_or_update_family_variant:
  - endpoint: PATCH /api/rest/v1/family-variants/{{ record.id }}
  - required fields: id
  - risk: creates or updates a family variant's axis/attribute-set structure; changes which attributes distinguish variant products under this family
- create_or_update_attribute_group:
  - endpoint: PATCH /api/rest/v1/attribute-groups/{{ record.id }}
  - required fields: id
  - risk: creates or updates an attribute group; reorganizes attribute grouping in the PIM's data-entry UI, a low-risk organizational mutation
- create_or_update_association_type:
  - endpoint: PATCH /api/rest/v1/association-types/{{ record.id }}
  - required fields: id
  - risk: creates or updates an association type (e.g. cross-sell/up-sell relationship definition); low-risk organizational mutation, no product data changes on its own

## Security

- read risk: external Akeneo PIM API read of product, category, family, attribute, channel, product-model, family-variant, attribute-group, association-type, locale, currency, and measure-family data
- write risk: external Akeneo PIM API upsert (create-or-update, PATCH-based) of products, categories, families, attributes, channels, product models, family variants, attribute groups, and association types; schema-shaping mutations (family/attribute/attribute-group) affect every product referencing them, approval required
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect akeneo
```

### Inspect as structured JSON

```bash
pm connectors inspect akeneo --json
```

## Agent Rules

- Run pm connectors inspect akeneo before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
- For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.

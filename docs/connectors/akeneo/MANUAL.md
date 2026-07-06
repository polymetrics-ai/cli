# pm connectors inspect akeneo

```text
NAME
  pm connectors inspect akeneo - Akeneo connector manual

SYNOPSIS
  pm connectors inspect akeneo
  pm connectors inspect akeneo --json
  pm credentials add <name> --connector akeneo [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Akeneo PIM products, categories, families, attributes, channels, product models, family variants, attribute groups, association types, locales, currencies, and measure families, and writes create-or-update upserts for the 9 catalog-structure resources, through the Akeneo REST API (OAuth2 password grant).

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
  api_username
  base_url
  client_id
  page_size
  password (secret)
  secret (secret)

ETL STREAMS
  products:
    primary key: id
    fields: categories(), created(), enabled(), family(), groups(), id(), parent(), updated(), uuid(), values()
  categories:
    primary key: id
    fields: id(), labels(), parent(), updated()
  families:
    primary key: id
    fields: attribute_as_image(), attribute_as_label(), attributes(), id(), labels()
  attributes:
    primary key: id
    fields: group(), id(), labels(), localizable(), scopable(), type()
  channels:
    primary key: id
    fields: category_tree(), currencies(), id(), labels(), locales()
  product_models:
    primary key: id
    fields: categories(), created(), family_variant(), id(), parent(), updated(), values()
  family_variants:
    primary key: id
    fields: attributes(), id(), labels(), variant_attribute_sets()
  attribute_groups:
    primary key: id
    fields: attributes(), id(), labels(), sort_order()
  association_types:
    primary key: id
    fields: id(), is_two_way(), labels()
  locales:
    primary key: id
    fields: enabled(), id()
  currencies:
    primary key: id
    fields: enabled(), id()
  measure_families:
    primary key: id
    fields: id(), standard_unit_code(), units()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

REVERSE ETL ACTIONS
  create_or_update_product:
    endpoint: PATCH /api/rest/v1/products/{{ record.id }}
    required fields: id
    risk: creates a new product (201) or updates an existing one (204) in the connected Akeneo PIM catalog; visible to every downstream channel the product is enabled/categorized for
  create_or_update_category:
    endpoint: PATCH /api/rest/v1/categories/{{ record.id }}
    required fields: id
    risk: creates or updates a category node; re-parenting an existing category changes the catalog tree for every product classified under it
  create_or_update_family:
    endpoint: PATCH /api/rest/v1/families/{{ record.id }}
    required fields: id
    risk: creates or updates a product family definition; changes the required/optional attribute set for every product assigned to this family
  create_or_update_attribute:
    endpoint: PATCH /api/rest/v1/attributes/{{ record.id }}
    required fields: id
    risk: creates a new attribute (schema mutation, affects every family/product referencing it) or updates an existing one's non-structural properties (labels/group); some attribute properties are immutable after creation per Akeneo's own API rules
  create_or_update_channel:
    endpoint: PATCH /api/rest/v1/channels/{{ record.id }}
    required fields: id
    risk: creates or updates a distribution channel; changes which locales/currencies/category tree every product exported to this channel uses
  create_or_update_product_model:
    endpoint: PATCH /api/rest/v1/product-models/{{ record.id }}
    required fields: id
    risk: creates or updates a product model; a shared parent for variant products, changes propagate to every variant beneath it
  create_or_update_family_variant:
    endpoint: PATCH /api/rest/v1/family-variants/{{ record.id }}
    required fields: id
    risk: creates or updates a family variant's axis/attribute-set structure; changes which attributes distinguish variant products under this family
  create_or_update_attribute_group:
    endpoint: PATCH /api/rest/v1/attribute-groups/{{ record.id }}
    required fields: id
    risk: creates or updates an attribute group; reorganizes attribute grouping in the PIM's data-entry UI, a low-risk organizational mutation
  create_or_update_association_type:
    endpoint: PATCH /api/rest/v1/association-types/{{ record.id }}
    required fields: id
    risk: creates or updates an association type (e.g. cross-sell/up-sell relationship definition); low-risk organizational mutation, no product data changes on its own

SECURITY
  read risk: external Akeneo PIM API read of product, category, family, attribute, channel, product-model, family-variant, attribute-group, association-type, locale, currency, and measure-family data
  write risk: external Akeneo PIM API upsert (create-or-update, PATCH-based) of products, categories, families, attributes, channels, product models, family variants, attribute groups, and association types; schema-shaping mutations (family/attribute/attribute-group) affect every product referencing them, approval required
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect akeneo

  # Inspect as structured JSON
  pm connectors inspect akeneo --json

AGENT WORKFLOW
  - Run pm connectors inspect akeneo before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.
  - For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```

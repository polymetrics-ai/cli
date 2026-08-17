# pm connectors inspect clazar

```text
NAME
  pm connectors inspect clazar - Clazar connector manual

SYNOPSIS
  pm connectors inspect clazar
  pm connectors inspect clazar --json
  pm credentials add <name> --connector clazar [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Clazar cloud GTM data (buyers, listings, contracts, opportunities, private offers, reseller offers, contacts, and metering records) and writes buyer/opportunity/contract/private-offer/contact/metering mutations, contract activation, and metering-record submission, through the Clazar REST API using OAuth2 client credentials.

ICON
  asset: icons/clazar.svg
  source: upstream_registry
  review_status: upstream_seeded
  review_url: https://clazar.io/api-docs

CAPABILITIES
  check=true catalog=true read=true write=true query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  base_url
  mode
  start_date
  client_id (secret)
  client_secret (secret)

ETL STREAMS
  buyers:
    primary key: id
    cursor: last_modified_at
    fields: cloud(), cloud_account_id(), domain(), id(), last_modified_at(), latest_contract_id(), listing_id(), name(), status()
  listings:
    primary key: id
    cursor: last_modified_at
    fields: cloud(), cloud_id(), cloud_url(), eula_type(), id(), last_modified_at(), long_description(), short_description(), status(), title()
  contracts:
    primary key: id
    cursor: last_modified_at
    fields: accepted_at(), auto_renew(), buyer_id(), cloud(), cloud_id(), duration(), end_at(), id(), last_modified_at(), latest_offer_id(), listing_id(), offer_type(), start_at(), status()
  opportunities:
    primary key: id
    cursor: last_modified_at
    fields: accept_by(), cloud(), cloud_id(), created_at(), customer_company(), customer_website(), id(), last_modified_at(), stage(), status(), target_close_date(), title()
  private_offers:
    primary key: id
    cursor: last_modified_at
    fields: accepted_at(), archived(), cloud(), cloud_id(), duration(), eula_type(), expiration_at(), id(), last_modified_at(), listing_id(), name(), offer_type(), published_at(), status()
  reseller_offers:
    primary key: id
    fields: accepted_at(), archived(), cloud(), cloud_id(), cloud_url(), eula_type(), expiration_at(), id(), listing_id(), name(), published_at(), status()
  contacts:
    primary key: id
    fields: created_at(), email(), full_name(), id(), is_editable(), phone_number(), updated_at(), uuid()
  metering:
    primary key: id
    fields: cloud(), contract_id(), custom_properties(), dimension(), end_time(), id(), quantity(), start_time(), status()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped, incremental_append, incremental_append_deduped

REVERSE ETL ACTIONS
  update_buyer:
    endpoint: PATCH /buyers/{{ record.id }}/
    required fields: id
    risk: external mutation of a buyer's custom properties / external-system associations; approval required
  update_opportunity:
    endpoint: PATCH /opportunities/{{ record.id }}/
    required fields: id
    risk: external mutation of an opportunity's custom properties / external-system associations; approval required
  update_private_offer:
    endpoint: PATCH /private_offers/{{ record.id }}/
    required fields: id
    risk: external mutation of a private offer's custom properties / external-system associations; approval required
  update_contract:
    endpoint: PATCH /contracts/{{ record.id }}/
    required fields: id
    risk: external mutation of a contract's custom properties / external-system associations; approval required
  activate_contract:
    endpoint: POST /contracts/{{ record.id }}/activate/
    required fields: id
    risk: irreversibly activates a pending Clazar contract in the underlying cloud marketplace; approval required (destructive/high-impact state transition)
  create_contact:
    endpoint: POST /contacts/
    risk: creates a new Clazar contact record; low-risk (no external marketplace side effects)
  update_contact:
    endpoint: PATCH /contacts/{{ record.id }}/
    required fields: id
    risk: updates a Clazar contact record; low-risk (no external marketplace side effects)
  delete_contact:
    endpoint: DELETE /contacts/{{ record.id }}/
    required fields: id
    risk: permanently deletes a Clazar contact record; approval required (destructive, irreversible)
  update_metering_record:
    endpoint: PATCH /metering/{{ record.id }}/
    required fields: id
    risk: updates only the custom_properties of a submitted metering record; low-risk
  create_metering_records:
    endpoint: POST /metering/
    optional fields: request
    risk: submits usage-based billing metering records that drive cloud marketplace invoicing for the buyer's contract; approval required (financial impact, effectively irreversible once billed)

SECURITY
  read risk: external Clazar API read of cloud marketplace GTM data
  write risk: external mutation of Clazar buyer/opportunity/contract/private-offer/contact/metering-record data; activate_contract irreversibly transitions a contract's state in the underlying cloud marketplace, and create_metering_records submits usage data that drives marketplace billing — every write ships with an explicit per-action risk string
  approval: required for activate_contract, delete_contact, and create_metering_records (financial/state-transition impact); custom_properties/external_object_associations updates and contact create/update are low-risk
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect clazar

  # Inspect as structured JSON
  pm connectors inspect clazar --json

AGENT WORKFLOW
  - Run pm connectors inspect clazar before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.
  - For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```

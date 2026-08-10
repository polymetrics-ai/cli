# pm connectors inspect ip2whois

```text
NAME
  pm connectors inspect ip2whois - IP2WHOIS connector manual

SYNOPSIS
  pm connectors inspect ip2whois
  pm connectors inspect ip2whois --json
  pm credentials add <name> --connector ip2whois [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Looks up WHOIS records for configured domains via the IP2WHOIS API, exposing a flattened whois stream and per-role contact streams (registrant, admin, tech, billing). The nameservers stream is not migrated; see docs.md Known limits.

ICON
  id: ip2whois
  asset: icons/ip2whois.svg
  source: upstream_registry
  review_status: upstream_seeded
  review_url: https://www.ip2whois.com/developers-api

CAPABILITIES
  check=true catalog=true read=true write=false query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  base_url
  domains (required)
  mode
  api_key (secret) (required)

ETL STREAMS
  whois:
    primary key: domain
    cursor: update_date
    fields: admin_email(string), admin_name(string), billing_email(string), billing_name(string), create_date(string), domain(string), domain_age(integer), domain_id(string), expire_date(string), nameservers(string), registrant_country(string), registrant_email(string), registrant_name(string), registrant_organization(string), registrar_iana_id(string), registrar_name(string), registrar_url(string), status(string), tech_email(string), tech_name(string), update_date(string), whois_server(string)
  contacts_registrant:
    primary key: domain, role
    fields: city(string), country(string), domain(string), email(string), fax(string), name(string), organization(string), phone(string), region(string), role(string), street_address(string), zip_code(string)
  contacts_admin:
    primary key: domain, role
    fields: city(string), country(string), domain(string), email(string), fax(string), name(string), organization(string), phone(string), region(string), role(string), street_address(string), zip_code(string)
  contacts_tech:
    primary key: domain, role
    fields: city(string), country(string), domain(string), email(string), fax(string), name(string), organization(string), phone(string), region(string), role(string), street_address(string), zip_code(string)
  contacts_billing:
    primary key: domain, role
    fields: city(string), country(string), domain(string), email(string), fax(string), name(string), organization(string), phone(string), region(string), role(string), street_address(string), zip_code(string)

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

SECURITY
  read risk: external IP2WHOIS API read of WHOIS records for the configured domain set
  approval: none; read-only, no reverse-ETL write surface
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect ip2whois

  # Inspect as structured JSON
  pm connectors inspect ip2whois --json

AGENT WORKFLOW
  - Run pm connectors inspect ip2whois before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```

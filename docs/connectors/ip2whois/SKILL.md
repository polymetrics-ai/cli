---
name: pm-ip2whois
description: IP2WHOIS connector knowledge and safe action guide.
---

# pm-ip2whois

## Purpose

Looks up WHOIS records for configured domains via the IP2WHOIS API, exposing a flattened whois stream and per-role contact streams (registrant, admin, tech, billing). The nameservers stream is not migrated; see docs.md Known limits.

## Icon

- id: ip2whois
- asset: icons/ip2whois.svg
- source: upstream_registry
- review_status: upstream_seeded
- review_url: https://www.ip2whois.com/developers-api

## Capabilities

- check=true catalog=true read=true write=false query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- base_url
- domains (required)
- mode
- api_key (secret) (required)

## ETL Streams

- whois:
  - primary key: domain
  - cursor: update_date
  - fields: admin_email(string), admin_name(string), billing_email(string), billing_name(string), create_date(string), domain(string), domain_age(integer), domain_id(string), expire_date(string), nameservers(string), registrant_country(string), registrant_email(string), registrant_name(string), registrant_organization(string), registrar_iana_id(string), registrar_name(string), registrar_url(string), status(string), tech_email(string), tech_name(string), update_date(string), whois_server(string)
- contacts_registrant:
  - primary key: domain, role
  - fields: city(string), country(string), domain(string), email(string), fax(string), name(string), organization(string), phone(string), region(string), role(string), street_address(string), zip_code(string)
- contacts_admin:
  - primary key: domain, role
  - fields: city(string), country(string), domain(string), email(string), fax(string), name(string), organization(string), phone(string), region(string), role(string), street_address(string), zip_code(string)
- contacts_tech:
  - primary key: domain, role
  - fields: city(string), country(string), domain(string), email(string), fax(string), name(string), organization(string), phone(string), region(string), role(string), street_address(string), zip_code(string)
- contacts_billing:
  - primary key: domain, role
  - fields: city(string), country(string), domain(string), email(string), fax(string), name(string), organization(string), phone(string), region(string), role(string), street_address(string), zip_code(string)

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

## Security

- read risk: external IP2WHOIS API read of WHOIS records for the configured domain set
- approval: none; read-only, no reverse-ETL write surface
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect ip2whois
```

### Inspect as structured JSON

```bash
pm connectors inspect ip2whois --json
```

## Agent Rules

- Run pm connectors inspect ip2whois before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.

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
- domains
- mode
- api_key (secret)

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

## Command Surface

- Run IP2WHOIS's declared streams and reverse-ETL actions.
- Usage: pm ip2whois <command> [flags]
- Read streams
- Other Commands
  - api get v2 - Documented GET /v2 (not implemented) [intent=direct_read availability=not_implemented operation=ip2whois.get.v2]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get v2-domain-lookup-nameservers-array - Documented GET /v2 (domain lookup, nameservers[] array) (not implemented) [intent=direct_read availability=not_implemented operation=ip2whois.get.v2-domain-lookup-nameservers-array]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - contacts admin list - Run the contacts admin ETL stream [intent=etl availability=implemented stream=contacts_admin]; notes: discrepancy=present-in-surface-absent-from-artifact
  - contacts billing list - Run the contacts billing ETL stream [intent=etl availability=implemented stream=contacts_billing]; notes: discrepancy=present-in-surface-absent-from-artifact
  - contacts registrant list - Run the contacts registrant ETL stream [intent=etl availability=implemented stream=contacts_registrant]; notes: discrepancy=present-in-surface-absent-from-artifact
  - contacts tech list - Run the contacts tech ETL stream [intent=etl availability=implemented stream=contacts_tech]; notes: discrepancy=present-in-surface-absent-from-artifact
  - whois list - Run the whois ETL stream [intent=etl availability=implemented stream=whois]; notes: discrepancy=present-in-surface-absent-from-artifact

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

---
name: pm-aviationstack
description: Aviationstack connector knowledge and safe action guide.
---

# pm-aviationstack

## Purpose

Reads aviationstack flights and aviation reference data (airlines, airports, airplanes, countries) through the aviationstack REST API. Read-only.

## Icon

- asset: icons/pm-sample.svg
- source: polymetrics
- review_status: polymetrics

## Capabilities

- check=true catalog=true read=true write=false query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- base_url
- mode
- access_key (secret)

## ETL Streams

- flights:
  - primary key: flight_date, flight_iata
  - cursor: flight_date
  - fields: airline_iata(), airline_name(), arrival_airport(), arrival_iata(), arrival_scheduled(), departure_airport(), departure_iata(), departure_scheduled(), flight_date(), flight_iata(), flight_icao(), flight_number(), flight_status()
- airlines:
  - primary key: id
  - fields: airline_name(), callsign(), country_iso2(), country_name(), date_founded(), fleet_size(), iata_code(), icao_code(), id(), status(), type()
- airports:
  - primary key: id
  - fields: airport_name(), city_iata_code(), country_iso2(), country_name(), gmt(), iata_code(), icao_code(), id(), latitude(), longitude(), timezone()
- airplanes:
  - primary key: id
  - fields: airline_iata_code(), first_flight_date(), iata_type(), icao_code_hex(), id(), model_code(), model_name(), plane_owner(), plane_status(), production_line(), registration_number()
- countries:
  - primary key: id
  - fields: capital(), continent(), country_iso2(), country_iso3(), country_iso_numeric(), country_name(), currency_code(), id(), phone_prefix(), population()

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped, incremental_append, incremental_append_deduped

## Security

- read risk: external aviationstack API read of flight and aviation reference data
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect aviationstack
```

### Inspect as structured JSON

```bash
pm connectors inspect aviationstack --json
```

## Agent Rules

- Run pm connectors inspect aviationstack before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.

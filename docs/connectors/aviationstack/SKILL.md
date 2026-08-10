---
name: pm-aviationstack
description: Aviationstack connector knowledge and safe action guide.
---

# pm-aviationstack

## Purpose

Reads aviationstack flights and aviation reference data (airlines, airports, airplanes, countries) through the aviationstack REST API. Read-only.

## Icon

- id: pm-sample
- asset: icons/pm-sample.svg
- source: polymetrics
- review_status: polymetrics
- review_url: https://github.com/polymetrics-ai/cli

## Capabilities

- check=true catalog=true read=true write=false query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- base_url
- mode
- access_key (secret) (required)

## ETL Streams

- flights:
  - primary key: flight_date, flight_iata
  - cursor: flight_date
  - fields: airline_iata(string), airline_name(string), arrival_airport(string), arrival_iata(string), arrival_scheduled(string), departure_airport(string), departure_iata(string), departure_scheduled(string), flight_date(string), flight_iata(string), flight_icao(string), flight_number(string), flight_status(string)
- airlines:
  - primary key: id
  - fields: airline_name(string), callsign(string), country_iso2(string), country_name(string), date_founded(string), fleet_size(string), iata_code(string), icao_code(string), id(string), status(string), type(string)
- airports:
  - primary key: id
  - fields: airport_name(string), city_iata_code(string), country_iso2(string), country_name(string), gmt(string), iata_code(string), icao_code(string), id(string), latitude(string), longitude(string), timezone(string)
- airplanes:
  - primary key: id
  - fields: airline_iata_code(string), first_flight_date(string), iata_type(string), icao_code_hex(string), id(string), model_code(string), model_name(string), plane_owner(string), plane_status(string), production_line(string), registration_number(string)
- countries:
  - primary key: id
  - fields: capital(string), continent(string), country_iso2(string), country_iso3(string), country_iso_numeric(string), country_name(string), currency_code(string), id(string), phone_prefix(string), population(string)

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

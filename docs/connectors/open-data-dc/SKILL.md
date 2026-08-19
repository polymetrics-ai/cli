---
name: pm-open-data-dc
description: Open Data DC connector knowledge and safe action guide.
---

# pm-open-data-dc

## Purpose

Reads District of Columbia Master Address Repository (MAR 2) locations, units, and SSL parcel records via the Open Data DC API. Read-only.

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
- location
- marid
- mode
- api_key (secret) (required)

## ETL Streams

- locations:
  - primary key: MarId
  - fields: AddrNum(string), Anc(string), CensusTract(string), FullAddress(string), Latitude(number), Longitude(number), MarId(string), Quadrant(string), ResidenceType(string), SSL(string), StName(string), Status(string), Ward(string), Xcoord(number), Ycoord(number), Zipcode(string), distance(number)
- units:
  - primary key: UnitNum
  - fields: FullAddress(string), MarId(string), Status(string), UnitNum(string), UnitSSL(string), UnitType(string)
- ssls:
  - primary key: SSL
  - fields: Col(string), FullAddress(string), Lot(string), Lot_type(string), MarId(string), SSL(string), Square(string)

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite

## Security

- read risk: external Open Data DC (MAR 2) API read of public address/parcel data
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect open-data-dc
```

### Inspect as structured JSON

```bash
pm connectors inspect open-data-dc --json
```

## Agent Rules

- Run pm connectors inspect open-data-dc before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.

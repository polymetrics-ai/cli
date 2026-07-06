---
name: pm-open-data-dc
description: Open Data DC connector knowledge and safe action guide.
---

# pm-open-data-dc

## Purpose

Reads District of Columbia Master Address Repository (MAR 2) locations, units, and SSL parcel records via the Open Data DC API. Read-only.

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
- location
- marid
- mode
- api_key (secret)

## ETL Streams

- locations:
  - primary key: MarId
  - fields: AddrNum(), Anc(), CensusTract(), FullAddress(), Latitude(), Longitude(), MarId(), Quadrant(), ResidenceType(), SSL(), StName(), Status(), Ward(), Xcoord(), Ycoord(), Zipcode(), distance()
- units:
  - primary key: UnitNum
  - fields: FullAddress(), MarId(), Status(), UnitNum(), UnitSSL(), UnitType()
- ssls:
  - primary key: SSL
  - fields: Col(), FullAddress(), Lot(), Lot_type(), MarId(), SSL(), Square()

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

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

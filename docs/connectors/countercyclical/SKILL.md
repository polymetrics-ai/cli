---
name: pm-countercyclical
description: Countercyclical connector knowledge and safe action guide.
---

# pm-countercyclical

## Purpose

Reads Countercyclical investments, valuations, research memos, teams, assumptions, and pipelines, and creates investments, through the Countercyclical REST API.

## Icon

- asset: icons/pm-sample.svg
- source: polymetrics
- review_status: polymetrics

## Capabilities

- check=true catalog=true read=true write=true query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- base_url
- mode
- api_key (secret)

## ETL Streams

- investments:
  - primary key: id
  - fields: cik(), country(), createdAt(), description(), editedName(), employees(), exchange(), figi(), financingType(), id(), industry(), isArchived(), isFavorite(), isLocked(), lei(), marketType(), name(), sector(), tickerSymbol(), updatedAt(), visibility(), website()
- valuations:
  - primary key: id
  - fields: createdAt(), delineation(), description(), discountRate(), endingQuarter(), endingYear(), growthMetric(), growthRate(), id(), isFavorite(), name(), shareToken(), startingQuarter(), startingYear(), status(), terminalPeriod(), terminalRate(), updatedAt()
- memos:
  - primary key: id
  - fields: archived(), backgroundColor(), bannerVisible(), body(), createdAt(), documentType(), emoji(), favorited(), foregroundColor(), id(), locked(), publiclyVisible(), sourcesVisible(), title(), tocVisible(), updatedAt(), views()
- teams:
  - primary key: id
  - fields: id(), title()
- assumptions:
  - primary key: id
  - fields: discountRate(), id(), name()
- pipelines:
  - primary key: id
  - fields: id(), name()

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

## Reverse ETL Actions

- create_investment:
  - endpoint: POST /integrations/make/actions/investments
  - risk: creates a new Investment in the caller's Countercyclical workspace via the Make-integration action endpoint (the only documented general-purpose creation endpoint; the functionally-identical Zapier-integration endpoint is not separately exposed, see api_surface.json); external mutation, no approval required

## Security

- read risk: external Countercyclical API read of investment and valuation data
- write risk: external mutation: creates a new Investment record in the caller's workspace; no update/delete actions are exposed
- approval: required for the create_investment write action; read-only otherwise
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect countercyclical
```

### Inspect as structured JSON

```bash
pm connectors inspect countercyclical --json
```

## Agent Rules

- Run pm connectors inspect countercyclical before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
- For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.

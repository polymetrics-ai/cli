---
name: pm-alpaca-broker-api
description: Alpaca Broker API connector knowledge and safe action guide.
---

# pm-alpaca-broker-api

## Purpose

Reads Alpaca Broker API accounts, assets, market calendar, clock, country info, account activities, journals, and per-account positions/watchlists/orders/documents over the Broker REST API (read-only).

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
- limit
- username
- password (secret)

## ETL Streams

- accounts:
  - primary key: id
  - fields: account_number(), account_type(), created_at(), crypto_status(), currency(), enabled_assets(), id(), kyc_results(), last_equity(), status()
- assets:
  - primary key: id
  - fields: class(), easy_to_borrow(), exchange(), fractionable(), id(), marginable(), name(), shortable(), status(), symbol(), tradable()
- calendar:
  - primary key: date
  - fields: close(), date(), open(), session_close(), session_open()
- clock:
  - primary key: timestamp
  - fields: is_open(), next_close(), next_open(), timestamp()
- country_info:
  - primary key: country_code
  - fields: country_code(), country_name(), phone_calling_code()
- account_activities:
  - primary key: id
  - fields: account_id(), activity_sub_type(), activity_type(), cum_qty(), cusip(), date(), description(), id(), leaves_qty(), net_amount(), order_id(), per_share_amount(), price(), qty(), side(), status(), symbol(), transaction_time(), type()
- journals:
  - primary key: id
  - fields: created_at(), description(), entry_type(), from_account(), id(), net_amount(), price(), qty(), settle_date(), status(), symbol(), system_date(), to_account()
- positions:
  - primary key: id, account_id
  - fields: account_id(), asset_class(), asset_id(), avg_entry_price(), change_today(), cost_basis(), current_price(), exchange(), id(), lastday_price(), market_value(), qty(), qty_available(), side(), symbol(), unrealized_pl(), unrealized_plpc()
- watchlists:
  - primary key: id
  - fields: account_id(), created_at(), id(), name(), updated_at()
- orders:
  - primary key: id
  - fields: account_id(), asset_class(), canceled_at(), created_at(), filled_at(), filled_avg_price(), filled_qty(), id(), limit_price(), notional(), order_class(), order_type(), qty(), side(), status(), stop_price(), submitted_at(), symbol(), time_in_force(), type(), updated_at()
- documents:
  - primary key: id, account_id
  - fields: account_id(), date(), id(), name(), sub_type(), type()

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

## Security

- read risk: external Alpaca Broker API read of account/asset/market metadata, plus per-account trading positions, orders, watchlists, and document metadata (financial PII adjacent; no document content is downloaded, only listing metadata)
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect alpaca-broker-api
```

### Inspect as structured JSON

```bash
pm connectors inspect alpaca-broker-api --json
```

## Agent Rules

- Run pm connectors inspect alpaca-broker-api before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.

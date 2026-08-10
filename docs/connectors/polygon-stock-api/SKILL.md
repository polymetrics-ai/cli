---
name: pm-polygon-stock-api
description: Polygon Stock API connector knowledge and safe action guide.
---

# pm-polygon-stock-api

## Purpose

Reads Polygon.io stock tickers, dividends, and splits through the Polygon.io reference REST API.

## Icon

- id: polygon
- asset: icons/polygon.svg
- source: upstream_registry
- review_status: upstream_seeded
- review_url: https://polygon.io/docs/stocks/getting-started

## Capabilities

- check=true catalog=true read=true write=false query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- active
- base_url
- ex_dividend_date
- execution_date
- locale
- market
- mode
- order
- page_size
- sort
- ticker
- type
- api_key (secret)

## ETL Streams

- tickers:
  - primary key: ticker
  - fields: active(boolean), currency_name(string), locale(string), market(string), name(string), primary_exchange(string), ticker(string)
- dividends:
  - primary key: id
  - cursor: ex_dividend_date
  - fields: cash_amount(number), currency(string), ex_dividend_date(string), id(string), ticker(string)
- splits:
  - primary key: id
  - cursor: execution_date
  - fields: execution_date(string), id(string), split_from(number), split_to(number), ticker(string)

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

## Security

- read risk: external Polygon.io API read of stock reference data (tickers, dividends, splits)
- approval: none; read-only, no obviously-safe reverse-ETL writes
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Command Surface

- Run Polygon Stock API's declared streams and reverse-ETL actions.
- Usage: pm polygon-stock-api <command> [flags]
- Read streams
- Other Commands
  - api get stocks filings 10-k vx sections - Documented GET /stocks/filings/10-K/vX/sections (not implemented) [intent=direct_read availability=not_implemented operation=polygon-stock-api.get.stocks-filings-10-k-vx-sections]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get stocks filings 8-k vx disclosures - Documented GET /stocks/filings/8-K/vX/disclosures (not implemented) [intent=direct_read availability=not_implemented operation=polygon-stock-api.get.stocks-filings-8-k-vx-disclosures]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get stocks filings 8-k vx text - Documented GET /stocks/filings/8-K/vX/text (not implemented) [intent=direct_read availability=not_implemented operation=polygon-stock-api.get.stocks-filings-8-k-vx-text]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get stocks filings vx 13-f - Documented GET /stocks/filings/vX/13-F (not implemented) [intent=direct_read availability=not_implemented operation=polygon-stock-api.get.stocks-filings-vx-13-f]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get stocks filings vx form-3 - Documented GET /stocks/filings/vX/form-3 (not implemented) [intent=direct_read availability=not_implemented operation=polygon-stock-api.get.stocks-filings-vx-form-3]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get stocks filings vx form-4 - Documented GET /stocks/filings/vX/form-4 (not implemented) [intent=direct_read availability=not_implemented operation=polygon-stock-api.get.stocks-filings-vx-form-4]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get stocks filings vx index - Documented GET /stocks/filings/vX/index (not implemented) [intent=direct_read availability=not_implemented operation=polygon-stock-api.get.stocks-filings-vx-index]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get stocks filings vx risk-factors - Documented GET /stocks/filings/vX/risk-factors (not implemented) [intent=direct_read availability=not_implemented operation=polygon-stock-api.get.stocks-filings-vx-risk-factors]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get stocks financials v1 balance-sheets - Documented GET /stocks/financials/v1/balance-sheets (not implemented) [intent=direct_read availability=not_implemented operation=polygon-stock-api.get.stocks-financials-v1-balance-sheets]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get stocks financials v1 cash-flow-statements - Documented GET /stocks/financials/v1/cash-flow-statements (not implemented) [intent=direct_read availability=not_implemented operation=polygon-stock-api.get.stocks-financials-v1-cash-flow-statements]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get stocks financials v1 income-statements - Documented GET /stocks/financials/v1/income-statements (not implemented) [intent=direct_read availability=not_implemented operation=polygon-stock-api.get.stocks-financials-v1-income-statements]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get stocks financials v1 ratios - Documented GET /stocks/financials/v1/ratios (not implemented) [intent=direct_read availability=not_implemented operation=polygon-stock-api.get.stocks-financials-v1-ratios]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get stocks taxonomies vx disclosures - Documented GET /stocks/taxonomies/vX/disclosures (not implemented) [intent=direct_read availability=not_implemented operation=polygon-stock-api.get.stocks-taxonomies-vx-disclosures]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get stocks taxonomies vx risk-factors - Documented GET /stocks/taxonomies/vX/risk-factors (not implemented) [intent=direct_read availability=not_implemented operation=polygon-stock-api.get.stocks-taxonomies-vx-risk-factors]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get stocks v1 dividends - Documented GET /stocks/v1/dividends (not implemented) [intent=direct_read availability=not_implemented operation=polygon-stock-api.get.stocks-v1-dividends]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get stocks v1 short-interest - Documented GET /stocks/v1/short-interest (not implemented) [intent=direct_read availability=not_implemented operation=polygon-stock-api.get.stocks-v1-short-interest]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get stocks v1 short-volume - Documented GET /stocks/v1/short-volume (not implemented) [intent=direct_read availability=not_implemented operation=polygon-stock-api.get.stocks-v1-short-volume]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get stocks v1 splits - Documented GET /stocks/v1/splits (not implemented) [intent=direct_read availability=not_implemented operation=polygon-stock-api.get.stocks-v1-splits]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get stocks vx float - Documented GET /stocks/vX/float (not implemented) [intent=direct_read availability=not_implemented operation=polygon-stock-api.get.stocks-vx-float]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get v1 indicators ema stockticker - Documented GET /v1/indicators/ema/{stockTicker} (not implemented) [intent=direct_read availability=not_implemented operation=polygon-stock-api.get.v1-indicators-ema-stockticker]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get v1 indicators macd stockticker - Documented GET /v1/indicators/macd/{stockTicker} (not implemented) [intent=direct_read availability=not_implemented operation=polygon-stock-api.get.v1-indicators-macd-stockticker]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get v1 indicators rsi stockticker - Documented GET /v1/indicators/rsi/{stockTicker} (not implemented) [intent=direct_read availability=not_implemented operation=polygon-stock-api.get.v1-indicators-rsi-stockticker]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get v1 indicators sma stockticker - Documented GET /v1/indicators/sma/{stockTicker} (not implemented) [intent=direct_read availability=not_implemented operation=polygon-stock-api.get.v1-indicators-sma-stockticker]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get v1 marketstatus now - Documented GET /v1/marketstatus/now (not implemented) [intent=direct_read availability=not_implemented operation=polygon-stock-api.get.v1-marketstatus-now]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get v1 marketstatus upcoming - Documented GET /v1/marketstatus/upcoming (not implemented) [intent=direct_read availability=not_implemented operation=polygon-stock-api.get.v1-marketstatus-upcoming]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get v1 open-close stocksticker date - Documented GET /v1/open-close/{stocksTicker}/{date} (not implemented) [intent=direct_read availability=not_implemented operation=polygon-stock-api.get.v1-open-close-stocksticker-date]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get v1 related-companies ticker - Documented GET /v1/related-companies/{ticker} (not implemented) [intent=direct_read availability=not_implemented operation=polygon-stock-api.get.v1-related-companies-ticker]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get v2 aggs grouped locale us market stocks date - Documented GET /v2/aggs/grouped/locale/us/market/stocks/{date} (not implemented) [intent=direct_read availability=not_implemented operation=polygon-stock-api.get.v2-aggs-grouped-locale-us-market-stocks-date]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get v2 aggs ticker stocksticker prev - Documented GET /v2/aggs/ticker/{stocksTicker}/prev (not implemented) [intent=direct_read availability=not_implemented operation=polygon-stock-api.get.v2-aggs-ticker-stocksticker-prev]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get v2 aggs ticker stocksticker range multiplier timespan from to - Documented GET /v2/aggs/ticker/{stocksTicker}/range/{multiplier}/{timespan}/{from}/{to} (not implemented) [intent=direct_read availability=not_implemented operation=polygon-stock-api.get.v2-aggs-ticker-stocksticker-range-multiplier-timespan-from-to]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get v2 aggs ticker ticker range multiplier timespan from to - Documented GET /v2/aggs/ticker/{ticker}/range/{multiplier}/{timespan}/{from}/{to} (not implemented) [intent=direct_read availability=not_implemented operation=polygon-stock-api.get.v2-aggs-ticker-ticker-range-multiplier-timespan-from-to]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get v2 last nbbo stocksticker - Documented GET /v2/last/nbbo/{stocksTicker} (not implemented) [intent=direct_read availability=not_implemented operation=polygon-stock-api.get.v2-last-nbbo-stocksticker]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get v2 last trade stocksticker - Documented GET /v2/last/trade/{stocksTicker} (not implemented) [intent=direct_read availability=not_implemented operation=polygon-stock-api.get.v2-last-trade-stocksticker]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get v2 reference news - Documented GET /v2/reference/news (not implemented) [intent=direct_read availability=not_implemented operation=polygon-stock-api.get.v2-reference-news]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get v2 snapshot locale us markets stocks direction - Documented GET /v2/snapshot/locale/us/markets/stocks/{direction} (not implemented) [intent=direct_read availability=not_implemented operation=polygon-stock-api.get.v2-snapshot-locale-us-markets-stocks-direction]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get v2 snapshot locale us markets stocks tickers - Documented GET /v2/snapshot/locale/us/markets/stocks/tickers (not implemented) [intent=direct_read availability=not_implemented operation=polygon-stock-api.get.v2-snapshot-locale-us-markets-stocks-tickers]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get v2 snapshot locale us markets stocks tickers stocksticker - Documented GET /v2/snapshot/locale/us/markets/stocks/tickers/{stocksTicker} (not implemented) [intent=direct_read availability=not_implemented operation=polygon-stock-api.get.v2-snapshot-locale-us-markets-stocks-tickers-stocksticker]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get v3 quotes stockticker - Documented GET /v3/quotes/{stockTicker} (not implemented) [intent=direct_read availability=not_implemented operation=polygon-stock-api.get.v3-quotes-stockticker]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get v3 reference conditions - Documented GET /v3/reference/conditions (not implemented) [intent=direct_read availability=not_implemented operation=polygon-stock-api.get.v3-reference-conditions]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get v3 reference exchanges - Documented GET /v3/reference/exchanges (not implemented) [intent=direct_read availability=not_implemented operation=polygon-stock-api.get.v3-reference-exchanges]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get v3 reference tickers ticker - Documented GET /v3/reference/tickers/{ticker} (not implemented) [intent=direct_read availability=not_implemented operation=polygon-stock-api.get.v3-reference-tickers-ticker]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get v3 reference tickers types - Documented GET /v3/reference/tickers/types (not implemented) [intent=direct_read availability=not_implemented operation=polygon-stock-api.get.v3-reference-tickers-types]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get v3 snapshot - Documented GET /v3/snapshot (not implemented) [intent=direct_read availability=not_implemented operation=polygon-stock-api.get.v3-snapshot]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get v3 trades stockticker - Documented GET /v3/trades/{stockTicker} (not implemented) [intent=direct_read availability=not_implemented operation=polygon-stock-api.get.v3-trades-stockticker]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get vx reference financials - Documented GET /vX/reference/financials (not implemented) [intent=direct_read availability=not_implemented operation=polygon-stock-api.get.vx-reference-financials]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get vx reference ipos - Documented GET /vX/reference/ipos (not implemented) [intent=direct_read availability=not_implemented operation=polygon-stock-api.get.vx-reference-ipos]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get vx reference tickers id events - Documented GET /vX/reference/tickers/{id}/events (not implemented) [intent=direct_read availability=not_implemented operation=polygon-stock-api.get.vx-reference-tickers-id-events]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - dividends list - Run the dividends ETL stream [intent=etl availability=implemented stream=dividends]
  - splits list - Run the splits ETL stream [intent=etl availability=implemented stream=splits]
  - tickers list - Run the tickers ETL stream [intent=etl availability=implemented stream=tickers]

## Commands

### Inspect as a manual

```bash
pm connectors inspect polygon-stock-api
```

### Inspect as structured JSON

```bash
pm connectors inspect polygon-stock-api --json
```

## Agent Rules

- Run pm connectors inspect polygon-stock-api before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.

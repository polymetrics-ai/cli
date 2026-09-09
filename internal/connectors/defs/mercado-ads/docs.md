# Mercado Ads Connector

## Overview

Reads Mercado Ads advertiser and campaign metrics through fixed Advertising API routes.

Readable streams: `brand_advertisers`, `display_advertisers`, `product_advertisers`, `brand_campaigns_metrics`, `display_campaigns_metrics`, `product_campaigns_metrics`.

Service API documentation: https://developers.mercadolibre.com.ar/en_us/advertising.

## Auth setup

Connection fields:

- `advertiser_id` (optional, string);
- `campaign_id` (optional, string);
- `client_id` (required, secret, string);
- `client_refresh_token` (required, secret, string);
- `client_secret` (required, secret, string);
- `end_date` (optional, string);
- `lookback_days` (required, string);
- `start_date` (optional, string);

Authentication uses declared mode(s): `oauth2_refresh_token`.

## Execution contract

Default stream pagination: `offset_limit`.

Connection check: `GET /advertising/advertisers`
Check query: `product_id`=`BADS`.

## Streams notes

- `brand_advertisers`: `GET /advertising/advertisers`; records `advertisers`
  - Query: `product_id`=`BADS`.
- `display_advertisers`: `GET /advertising/advertisers`; records `advertisers`
  - Query: `product_id`=`DISPLAY`.
- `product_advertisers`: `GET /advertising/advertisers`; records `advertisers`
  - Query: `product_id`=`PADS`.
- `brand_campaigns_metrics`: `GET /advertising/advertisers/{{ config.advertiser_id }}/brand_ads/campaigns/{{ config.campaign_id }}/metrics`; records `metrics`
  - Incremental cursor: `date`.
- `display_campaigns_metrics`: `GET /advertising/advertisers/{{ config.advertiser_id }}/display_ads/campaigns/{{ config.campaign_id }}/metrics`; records `metrics`
  - Incremental cursor: `date`.
- `product_campaigns_metrics`: `GET /advertising/advertisers/{{ config.advertiser_id }}/product_ads/campaigns/{{ config.campaign_id }}/metrics`; records `metrics`
  - Incremental cursor: `date`.

## Write actions & risks

This connector's write surface is declared separately in the rendered execution bundle.

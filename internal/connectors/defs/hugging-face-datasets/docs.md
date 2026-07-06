# Overview

Reads dataset splits and per-split sizes from the Hugging Face dataset-viewer REST API. Read-only;
an optional user access token unlocks gated and private datasets.

Readable streams: `splits`, `sizes`.

This connector is read-only; no write actions are declared.

Service API documentation: https://huggingface.co/docs/datasets/.

## Auth setup

Connection fields:

- `access_token` (optional, secret, string); Never logged.
- `api_token` (optional, secret, string); Optional Hugging Face user access token, sent as a Bearer
  token (Authorization: Bearer <api_token>). Public datasets do not need one; gated/private datasets
  do. Never logged.
- `base_url` (optional, string); default `https://datasets-server.huggingface.co`; format `uri`;
  Hugging Face dataset-viewer API base URL override for tests or proxies.
- `dataset_name` (required, string); Hugging Face dataset identifier (e.g. 'ibm/duorc'); sent as the
  dataset query param on every request.
- `token` (optional, secret, string); Never logged.

Secret fields are redacted in logs and write previews: `access_token`, `api_token`, `token`.

Default configuration values: `base_url=https://datasets-server.huggingface.co`.

Authentication behavior:

- Bearer token authentication using `secrets.api_token` when `{{ secrets.api_token }}`.
- Bearer token authentication using `secrets.access_token` when `{{ secrets.access_token }}`.
- Bearer token authentication using `secrets.token` when `{{ secrets.token }}`.
- No authentication.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/is-valid` with query `dataset`=`{{ config.dataset_name }}`.

## Streams notes

Default pagination: single request; no pagination.

- `splits`: GET `/splits` - records path `splits`; query `dataset`=`{{ config.dataset_name }}`.
- `sizes`: GET `/size` - records path `size.splits`; query `dataset`=`{{ config.dataset_name }}`.

## Write actions & risks

This connector is read-only. Read behavior: external Hugging Face dataset-viewer API read of dataset
split/size metadata; an optional access token unlocks gated/private dataset reads.

## Known limits

- Batch defaults: read_page_size=100.
- API coverage includes 2 stream-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  non_data_endpoint=1, out_of_scope=1.

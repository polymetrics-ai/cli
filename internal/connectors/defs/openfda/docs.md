# Overview

Reads documented FDA drug, device, food, animal/veterinary, cosmetics, tobacco, transparency, and
other public datasets from the openFDA REST API.

Readable streams: `drug_event`, `drug_label`, `drug_enforcement`, `device_event`,
`food_enforcement`, `animalandveterinary_event`, `cosmetic_event`, `food_event`, `drug_ndc`,
`drug_drugsfda`, `drug_shortages`, `drug_orangebook`, `device_510k`, `device_pma`, `device_udi`,
`device_enforcement`, `device_recall`, `device_classification`, `device_registrationlisting`,
`device_covid19serology`, `tobacco_problem`, `tobacco_researchdigitalads`,
`tobacco_researchpreventionads`, `tobacco_researchsmokefree`, `transparency_crl`,
`other_historicaldocument`, `other_nsde`, `other_substance`, `other_unii`.

This connector is read-only; no write actions are declared.

Service API documentation: https://open.fda.gov/apis/.

## Auth setup

Connection fields:

- `api_key` (optional, secret, string); Optional openFDA API key, sent as the api_key query
  parameter to raise rate limits. openFDA works credential-free; when unset, requests are sent
  unauthenticated (public tier).
- `base_url` (optional, string); default `https://api.fda.gov`; format `uri`; openFDA API base URL
  override for tests or proxies.
- `mode` (optional, string).
- `search` (optional, string); Optional openFDA search query string, applied to every stream via the
  search query parameter. Absent when unset (unfiltered).

Secret fields are redacted in logs and write previews: `api_key`.

Default configuration values: `base_url=https://api.fda.gov`.

Authentication behavior:

- API key authentication in query parameter `api_key` using `secrets.api_key` when `{{
  secrets.api_key }}`.
- No authentication.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/drug/event.json` with query `limit`=`1`.

## Streams notes

Default pagination: offset/limit pagination; offset parameter `skip`; limit parameter `limit`; page
size 100.

- `drug_event`: GET `/drug/event.json` - records path `results`; query `search` from template `{{
  config.search }}`, omitted when absent; offset/limit pagination; offset parameter `skip`; limit
  parameter `limit`; page size 100.
- `drug_label`: GET `/drug/label.json` - records path `results`; query `search` from template `{{
  config.search }}`, omitted when absent; offset/limit pagination; offset parameter `skip`; limit
  parameter `limit`; page size 100.
- `drug_enforcement`: GET `/drug/enforcement.json` - records path `results`; query `search` from
  template `{{ config.search }}`, omitted when absent; offset/limit pagination; offset parameter
  `skip`; limit parameter `limit`; page size 100.
- `device_event`: GET `/device/event.json` - records path `results`; query `search` from template
  `{{ config.search }}`, omitted when absent; offset/limit pagination; offset parameter `skip`;
  limit parameter `limit`; page size 100.
- `food_enforcement`: GET `/food/enforcement.json` - records path `results`; query `search` from
  template `{{ config.search }}`, omitted when absent; offset/limit pagination; offset parameter
  `skip`; limit parameter `limit`; page size 100.
- `animalandveterinary_event`: GET `/animalandveterinary/event.json` - records path `results`; query
  `search` from template `{{ config.search }}`, omitted when absent; offset/limit pagination; offset
  parameter `skip`; limit parameter `limit`; page size 100; emits passthrough records.
- `cosmetic_event`: GET `/cosmetic/event.json` - records path `results`; query `search` from
  template `{{ config.search }}`, omitted when absent; offset/limit pagination; offset parameter
  `skip`; limit parameter `limit`; page size 100; emits passthrough records.
- `food_event`: GET `/food/event.json` - records path `results`; query `search` from template `{{
  config.search }}`, omitted when absent; offset/limit pagination; offset parameter `skip`; limit
  parameter `limit`; page size 100; emits passthrough records.
- `drug_ndc`: GET `/drug/ndc.json` - records path `results`; query `search` from template `{{
  config.search }}`, omitted when absent; offset/limit pagination; offset parameter `skip`; limit
  parameter `limit`; page size 100; emits passthrough records.
- `drug_drugsfda`: GET `/drug/drugsfda.json` - records path `results`; query `search` from template
  `{{ config.search }}`, omitted when absent; offset/limit pagination; offset parameter `skip`;
  limit parameter `limit`; page size 100; emits passthrough records.
- `drug_shortages`: GET `/drug/shortages.json` - records path `results`; query `search` from
  template `{{ config.search }}`, omitted when absent; offset/limit pagination; offset parameter
  `skip`; limit parameter `limit`; page size 100; emits passthrough records.
- `drug_orangebook`: GET `/drug/orangebook.json` - records path `results`; query `search` from
  template `{{ config.search }}`, omitted when absent; offset/limit pagination; offset parameter
  `skip`; limit parameter `limit`; page size 100; emits passthrough records.
- `device_510k`: GET `/device/510k.json` - records path `results`; query `search` from template `{{
  config.search }}`, omitted when absent; offset/limit pagination; offset parameter `skip`; limit
  parameter `limit`; page size 100; emits passthrough records.
- `device_pma`: GET `/device/pma.json` - records path `results`; query `search` from template `{{
  config.search }}`, omitted when absent; offset/limit pagination; offset parameter `skip`; limit
  parameter `limit`; page size 100; emits passthrough records.
- `device_udi`: GET `/device/udi.json` - records path `results`; query `search` from template `{{
  config.search }}`, omitted when absent; offset/limit pagination; offset parameter `skip`; limit
  parameter `limit`; page size 100; emits passthrough records.
- `device_enforcement`: GET `/device/enforcement.json` - records path `results`; query `search` from
  template `{{ config.search }}`, omitted when absent; offset/limit pagination; offset parameter
  `skip`; limit parameter `limit`; page size 100; emits passthrough records.
- `device_recall`: GET `/device/recall.json` - records path `results`; query `search` from template
  `{{ config.search }}`, omitted when absent; offset/limit pagination; offset parameter `skip`;
  limit parameter `limit`; page size 100; emits passthrough records.
- `device_classification`: GET `/device/classification.json` - records path `results`; query
  `search` from template `{{ config.search }}`, omitted when absent; offset/limit pagination; offset
  parameter `skip`; limit parameter `limit`; page size 100; emits passthrough records.
- `device_registrationlisting`: GET `/device/registrationlisting.json` - records path `results`;
  query `search` from template `{{ config.search }}`, omitted when absent; offset/limit pagination;
  offset parameter `skip`; limit parameter `limit`; page size 100; emits passthrough records.
- `device_covid19serology`: GET `/device/covid19serology.json` - records path `results`; query
  `search` from template `{{ config.search }}`, omitted when absent; offset/limit pagination; offset
  parameter `skip`; limit parameter `limit`; page size 100; emits passthrough records.
- `tobacco_problem`: GET `/tobacco/problem.json` - records path `results`; query `search` from
  template `{{ config.search }}`, omitted when absent; offset/limit pagination; offset parameter
  `skip`; limit parameter `limit`; page size 100; emits passthrough records.
- `tobacco_researchdigitalads`: GET `/tobacco/researchdigitalads.json` - records path `results`;
  query `search` from template `{{ config.search }}`, omitted when absent; offset/limit pagination;
  offset parameter `skip`; limit parameter `limit`; page size 100; emits passthrough records.
- `tobacco_researchpreventionads`: GET `/tobacco/researchpreventionads.json` - records path
  `results`; query `search` from template `{{ config.search }}`, omitted when absent; offset/limit
  pagination; offset parameter `skip`; limit parameter `limit`; page size 100; emits passthrough
  records.
- `tobacco_researchsmokefree`: GET `/tobacco/researchsmokefree.json` - records path `results`; query
  `search` from template `{{ config.search }}`, omitted when absent; offset/limit pagination; offset
  parameter `skip`; limit parameter `limit`; page size 100; emits passthrough records.
- `transparency_crl`: GET `/transparency/crl.json` - records path `results`; query `search` from
  template `{{ config.search }}`, omitted when absent; offset/limit pagination; offset parameter
  `skip`; limit parameter `limit`; page size 100; emits passthrough records.
- `other_historicaldocument`: GET `/other/historicaldocument.json` - records path `results`; query
  `search` from template `{{ config.search }}`, omitted when absent; offset/limit pagination; offset
  parameter `skip`; limit parameter `limit`; page size 100; emits passthrough records.
- `other_nsde`: GET `/other/nsde.json` - records path `results`; query `search` from template `{{
  config.search }}`, omitted when absent; offset/limit pagination; offset parameter `skip`; limit
  parameter `limit`; page size 100; emits passthrough records.
- `other_substance`: GET `/other/substance.json` - records path `results`; query `search` from
  template `{{ config.search }}`, omitted when absent; offset/limit pagination; offset parameter
  `skip`; limit parameter `limit`; page size 100; emits passthrough records.
- `other_unii`: GET `/other/unii.json` - records path `results`; query `search` from template `{{
  config.search }}`, omitted when absent; offset/limit pagination; offset parameter `skip`; limit
  parameter `limit`; page size 100; emits passthrough records.

## Write actions & risks

This connector is read-only. Read behavior: external openFDA API read of public FDA regulatory
datasets.

## Known limits

- Batch defaults: read_page_size=100.
- API coverage includes 29 stream-backed endpoint group(s).

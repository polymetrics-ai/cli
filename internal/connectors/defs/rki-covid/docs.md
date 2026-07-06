# Overview

Reads public Germany COVID case, state, district, and history data derived from RKI reports via the
corona-zahlen.org JSON API. Read-only, credential-free.

Readable streams: `germany`, `states`, `districts`, `cases_history`, `deaths_history`,
`germany_incidence_history`, `germany_recovered_history`, `germany_r_value_history`,
`germany_hospitalization_history`, `germany_frozen_incidence_history`, `germany_age_groups`,
`states_cases_history`, `states_deaths_history`, `states_incidence_history`,
`states_recovered_history`, `states_frozen_incidence_history`, `states_hospitalization_history`,
`states_age_groups`, `districts_cases_history`, `districts_deaths_history`,
`districts_incidence_history`, `districts_recovered_history`, `districts_frozen_incidence_history`,
`districts_age_groups`, `testing_history`, `vaccinations`, `vaccinations_states`,
`vaccinations_history`.

This connector is read-only; no write actions are declared.

Service API documentation: https://api.corona-zahlen.org.

## Auth setup

Connection fields:

- `base_url` (optional, string); default `https://api.corona-zahlen.org`; format `uri`;
  corona-zahlen.org API base URL override for tests or proxies. Defaults to
  https://api.corona-zahlen.org.
- `days` (optional, string); Omitted entirely when unset.

Default configuration values: `base_url=https://api.corona-zahlen.org`.

Authentication behavior:

- No authentication.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/germany`.

## Streams notes

Default pagination: single request; no pagination.

Incremental streams use their declared cursor fields and send lower-bound parameters only when a
lower bound is available.

- `germany`: GET `/germany` - records at response root; query `days` from template `{{ config.days
  }}`, omitted when absent; computed output fields `id`, `stream`; emits passthrough records.
- `states`: GET `/states` - records path `data`; query `days` from template `{{ config.days }}`,
  omitted when absent; computed output fields `id`, `stream`; emits passthrough records.
- `districts`: GET `/districts` - records path `data`; query `days` from template `{{ config.days
  }}`, omitted when absent; computed output fields `id`, `stream`; emits passthrough records.
- `cases_history`: GET `/germany/history/cases` - records path `data`; query `days` from template
  `{{ config.days }}`, omitted when absent; incremental cursor `date`; formatted as `rfc3339`;
  computed output fields `id`, `stream`; emits passthrough records.
- `deaths_history`: GET `/germany/history/deaths` - records path `data`; query `days` from template
  `{{ config.days }}`, omitted when absent; incremental cursor `date`; formatted as `rfc3339`;
  computed output fields `id`, `stream`; emits passthrough records.
- `germany_incidence_history`: GET `/germany/history/incidence` - records path `data`; query `days`
  from template `{{ config.days }}`, omitted when absent; incremental cursor `date`; formatted as
  `rfc3339`; computed output fields `id`, `stream`; emits passthrough records.
- `germany_recovered_history`: GET `/germany/history/recovered` - records path `data`; query `days`
  from template `{{ config.days }}`, omitted when absent; incremental cursor `date`; formatted as
  `rfc3339`; computed output fields `id`, `stream`; emits passthrough records.
- `germany_r_value_history`: GET `/germany/history/rValue` - records path `data`; query `days` from
  template `{{ config.days }}`, omitted when absent; incremental cursor `date`; formatted as
  `rfc3339`; computed output fields `id`, `stream`; emits passthrough records.
- `germany_hospitalization_history`: GET `/germany/history/hospitalization` - records path `data`;
  query `days` from template `{{ config.days }}`, omitted when absent; incremental cursor `date`;
  formatted as `rfc3339`; computed output fields `id`, `stream`; emits passthrough records.
- `germany_frozen_incidence_history`: GET `/germany/history/frozen-incidence` - records path
  `data.history`; query `days` from template `{{ config.days }}`, omitted when absent; incremental
  cursor `date`; formatted as `rfc3339`; computed output fields `id`, `stream`; emits passthrough
  records.
- `germany_age_groups`: GET `/germany/age-groups` - records path `data`; flattens keyed objects; key
  field `age_group`; query `days` from template `{{ config.days }}`, omitted when absent; computed
  output fields `id`, `stream`; emits passthrough records.
- `states_cases_history`: GET `/states/history/cases` - records path `data`; flattens keyed objects;
  key field `abbreviation`; query `days` from template `{{ config.days }}`, omitted when absent;
  computed output fields `id`, `stream`; emits passthrough records.
- `states_deaths_history`: GET `/states/history/deaths` - records path `data`; flattens keyed
  objects; key field `abbreviation`; query `days` from template `{{ config.days }}`, omitted when
  absent; computed output fields `id`, `stream`; emits passthrough records.
- `states_incidence_history`: GET `/states/history/incidence` - records path `data`; flattens keyed
  objects; key field `abbreviation`; query `days` from template `{{ config.days }}`, omitted when
  absent; computed output fields `id`, `stream`; emits passthrough records.
- `states_recovered_history`: GET `/states/history/recovered` - records path `data`; flattens keyed
  objects; key field `abbreviation`; query `days` from template `{{ config.days }}`, omitted when
  absent; computed output fields `id`, `stream`; emits passthrough records.
- `states_frozen_incidence_history`: GET `/states/history/frozen-incidence` - records path `data`;
  flattens keyed objects; key field `abbreviation`; query `days` from template `{{ config.days }}`,
  omitted when absent; computed output fields `id`, `stream`; emits passthrough records.
- `states_hospitalization_history`: GET `/states/history/hospitalization` - records path `data`;
  flattens keyed objects; key field `abbreviation`; query `days` from template `{{ config.days }}`,
  omitted when absent; computed output fields `id`, `stream`; emits passthrough records.
- `states_age_groups`: GET `/states/age-groups` - records path `data`; flattens keyed objects; key
  field `abbreviation`; query `days` from template `{{ config.days }}`, omitted when absent;
  computed output fields `id`, `stream`; emits passthrough records.
- `districts_cases_history`: GET `/districts/history/cases` - records path `data`; flattens keyed
  objects; key field `ags`; query `days` from template `{{ config.days }}`, omitted when absent;
  computed output fields `id`, `stream`; emits passthrough records.
- `districts_deaths_history`: GET `/districts/history/deaths` - records path `data`; flattens keyed
  objects; key field `ags`; query `days` from template `{{ config.days }}`, omitted when absent;
  computed output fields `id`, `stream`; emits passthrough records.
- `districts_incidence_history`: GET `/districts/history/incidence` - records path `data`; flattens
  keyed objects; key field `ags`; query `days` from template `{{ config.days }}`, omitted when
  absent; computed output fields `id`, `stream`; emits passthrough records.
- `districts_recovered_history`: GET `/districts/history/recovered` - records path `data`; flattens
  keyed objects; key field `ags`; query `days` from template `{{ config.days }}`, omitted when
  absent; computed output fields `id`, `stream`; emits passthrough records.
- `districts_frozen_incidence_history`: GET `/districts/history/frozen-incidence` - records path
  `data`; flattens keyed objects; key field `ags`; query `days` from template `{{ config.days }}`,
  omitted when absent; computed output fields `id`, `stream`; emits passthrough records.
- `districts_age_groups`: GET `/districts/age-groups` - records path `data`; flattens keyed objects;
  key field `ags`; query `days` from template `{{ config.days }}`, omitted when absent; computed
  output fields `id`, `stream`; emits passthrough records.
- `testing_history`: GET `/testing/history` - records path `data.history`; query `days` from
  template `{{ config.days }}`, omitted when absent; computed output fields `id`, `stream`; emits
  passthrough records.
- `vaccinations`: GET `/vaccinations` - records path `data`; query `days` from template `{{
  config.days }}`, omitted when absent; computed output fields `id`, `stream`; emits passthrough
  records.
- `vaccinations_states`: GET `/vaccinations/states` - records path `data`; flattens keyed objects;
  key field `abbreviation`; query `days` from template `{{ config.days }}`, omitted when absent;
  computed output fields `id`, `stream`; emits passthrough records.
- `vaccinations_history`: GET `/vaccinations/history` - records path `data.history`; query `days`
  from template `{{ config.days }}`, omitted when absent; incremental cursor `date`; formatted as
  `rfc3339`; computed output fields `id`, `stream`; emits passthrough records.

## Write actions & risks

This connector is read-only. Read behavior: external corona-zahlen.org public JSON API read of
Germany COVID metrics.

## Known limits

- Batch defaults: read_page_size=100.
- API coverage includes 28 stream-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  duplicate_of=51, non_data_endpoint=14.

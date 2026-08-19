---
name: pm-rki-covid
description: RKI COVID connector knowledge and safe action guide.
---

# pm-rki-covid

## Purpose

Reads public Germany COVID case, state, district, and history data derived from RKI reports via the corona-zahlen.org JSON API. Read-only, credential-free.

## Icon

- id: rki
- asset: icons/rki.svg
- source: upstream_registry
- review_status: upstream_seeded

## Capabilities

- check=true catalog=true read=true write=false query=false
- Integration type: api

## Authentication

- No secret authentication is required for this connector.

## Configuration

- base_url
- days

## ETL Streams

- germany:
  - primary key: id
  - fields: abbreviation(string), administeredVaccinations(integer), age_group(string), ags(string), calendarWeek(string), cases(integer), cases7Days(integer), dataSource(string), date(string), deaths(integer), delta(integer), history(array), id(string), incidence7Days(number), laboratoryCount(integer), name(string), performedTests(integer), positiveTests(integer), positivityRate(number), quote(number), rValue4Days(number), rValue7Days(number), recovered(integer), stream(string), vaccinated(integer), weekIncidence(number)
- states:
  - primary key: id
  - fields: abbreviation(string), administeredVaccinations(integer), age_group(string), ags(string), calendarWeek(string), cases(integer), cases7Days(integer), dataSource(string), date(string), deaths(integer), delta(integer), history(array), id(string), incidence7Days(number), laboratoryCount(integer), name(string), performedTests(integer), positiveTests(integer), positivityRate(number), quote(number), rValue4Days(number), rValue7Days(number), recovered(integer), stream(string), vaccinated(integer), weekIncidence(number)
- districts:
  - primary key: id
  - fields: abbreviation(string), administeredVaccinations(integer), age_group(string), ags(string), calendarWeek(string), cases(integer), cases7Days(integer), dataSource(string), date(string), deaths(integer), delta(integer), history(array), id(string), incidence7Days(number), laboratoryCount(integer), name(string), performedTests(integer), positiveTests(integer), positivityRate(number), quote(number), rValue4Days(number), rValue7Days(number), recovered(integer), stream(string), vaccinated(integer), weekIncidence(number)
- cases_history:
  - primary key: id
  - cursor: date
  - fields: abbreviation(string), administeredVaccinations(integer), age_group(string), ags(string), calendarWeek(string), cases(integer), cases7Days(integer), dataSource(string), date(string), deaths(integer), delta(integer), history(array), id(string), incidence7Days(number), laboratoryCount(integer), name(string), performedTests(integer), positiveTests(integer), positivityRate(number), quote(number), rValue4Days(number), rValue7Days(number), recovered(integer), stream(string), vaccinated(integer), weekIncidence(number)
- deaths_history:
  - primary key: id
  - cursor: date
  - fields: abbreviation(string), administeredVaccinations(integer), age_group(string), ags(string), calendarWeek(string), cases(integer), cases7Days(integer), dataSource(string), date(string), deaths(integer), delta(integer), history(array), id(string), incidence7Days(number), laboratoryCount(integer), name(string), performedTests(integer), positiveTests(integer), positivityRate(number), quote(number), rValue4Days(number), rValue7Days(number), recovered(integer), stream(string), vaccinated(integer), weekIncidence(number)
- germany_incidence_history:
  - primary key: id
  - cursor: date
  - fields: abbreviation(string), administeredVaccinations(integer), age_group(string), ags(string), calendarWeek(string), cases(integer), cases7Days(integer), dataSource(string), date(string), deaths(integer), delta(integer), history(array), id(string), incidence7Days(number), laboratoryCount(integer), name(string), performedTests(integer), positiveTests(integer), positivityRate(number), quote(number), rValue4Days(number), rValue7Days(number), recovered(integer), stream(string), vaccinated(integer), weekIncidence(number)
- germany_recovered_history:
  - primary key: id
  - cursor: date
  - fields: abbreviation(string), administeredVaccinations(integer), age_group(string), ags(string), calendarWeek(string), cases(integer), cases7Days(integer), dataSource(string), date(string), deaths(integer), delta(integer), history(array), id(string), incidence7Days(number), laboratoryCount(integer), name(string), performedTests(integer), positiveTests(integer), positivityRate(number), quote(number), rValue4Days(number), rValue7Days(number), recovered(integer), stream(string), vaccinated(integer), weekIncidence(number)
- germany_r_value_history:
  - primary key: id
  - cursor: date
  - fields: abbreviation(string), administeredVaccinations(integer), age_group(string), ags(string), calendarWeek(string), cases(integer), cases7Days(integer), dataSource(string), date(string), deaths(integer), delta(integer), history(array), id(string), incidence7Days(number), laboratoryCount(integer), name(string), performedTests(integer), positiveTests(integer), positivityRate(number), quote(number), rValue4Days(number), rValue7Days(number), recovered(integer), stream(string), vaccinated(integer), weekIncidence(number)
- germany_hospitalization_history:
  - primary key: id
  - cursor: date
  - fields: abbreviation(string), administeredVaccinations(integer), age_group(string), ags(string), calendarWeek(string), cases(integer), cases7Days(integer), dataSource(string), date(string), deaths(integer), delta(integer), history(array), id(string), incidence7Days(number), laboratoryCount(integer), name(string), performedTests(integer), positiveTests(integer), positivityRate(number), quote(number), rValue4Days(number), rValue7Days(number), recovered(integer), stream(string), vaccinated(integer), weekIncidence(number)
- germany_frozen_incidence_history:
  - primary key: id
  - cursor: date
  - fields: abbreviation(string), administeredVaccinations(integer), age_group(string), ags(string), calendarWeek(string), cases(integer), cases7Days(integer), dataSource(string), date(string), deaths(integer), delta(integer), history(array), id(string), incidence7Days(number), laboratoryCount(integer), name(string), performedTests(integer), positiveTests(integer), positivityRate(number), quote(number), rValue4Days(number), rValue7Days(number), recovered(integer), stream(string), vaccinated(integer), weekIncidence(number)
- germany_age_groups:
  - primary key: id
  - fields: abbreviation(string), administeredVaccinations(integer), age_group(string), ags(string), calendarWeek(string), cases(integer), cases7Days(integer), dataSource(string), date(string), deaths(integer), delta(integer), history(array), id(string), incidence7Days(number), laboratoryCount(integer), name(string), performedTests(integer), positiveTests(integer), positivityRate(number), quote(number), rValue4Days(number), rValue7Days(number), recovered(integer), stream(string), vaccinated(integer), weekIncidence(number)
- states_cases_history:
  - primary key: id
  - fields: abbreviation(string), administeredVaccinations(integer), age_group(string), ags(string), calendarWeek(string), cases(integer), cases7Days(integer), dataSource(string), date(string), deaths(integer), delta(integer), history(array), id(string), incidence7Days(number), laboratoryCount(integer), name(string), performedTests(integer), positiveTests(integer), positivityRate(number), quote(number), rValue4Days(number), rValue7Days(number), recovered(integer), stream(string), vaccinated(integer), weekIncidence(number)
- states_deaths_history:
  - primary key: id
  - fields: abbreviation(string), administeredVaccinations(integer), age_group(string), ags(string), calendarWeek(string), cases(integer), cases7Days(integer), dataSource(string), date(string), deaths(integer), delta(integer), history(array), id(string), incidence7Days(number), laboratoryCount(integer), name(string), performedTests(integer), positiveTests(integer), positivityRate(number), quote(number), rValue4Days(number), rValue7Days(number), recovered(integer), stream(string), vaccinated(integer), weekIncidence(number)
- states_incidence_history:
  - primary key: id
  - fields: abbreviation(string), administeredVaccinations(integer), age_group(string), ags(string), calendarWeek(string), cases(integer), cases7Days(integer), dataSource(string), date(string), deaths(integer), delta(integer), history(array), id(string), incidence7Days(number), laboratoryCount(integer), name(string), performedTests(integer), positiveTests(integer), positivityRate(number), quote(number), rValue4Days(number), rValue7Days(number), recovered(integer), stream(string), vaccinated(integer), weekIncidence(number)
- states_recovered_history:
  - primary key: id
  - fields: abbreviation(string), administeredVaccinations(integer), age_group(string), ags(string), calendarWeek(string), cases(integer), cases7Days(integer), dataSource(string), date(string), deaths(integer), delta(integer), history(array), id(string), incidence7Days(number), laboratoryCount(integer), name(string), performedTests(integer), positiveTests(integer), positivityRate(number), quote(number), rValue4Days(number), rValue7Days(number), recovered(integer), stream(string), vaccinated(integer), weekIncidence(number)
- states_frozen_incidence_history:
  - primary key: id
  - fields: abbreviation(string), administeredVaccinations(integer), age_group(string), ags(string), calendarWeek(string), cases(integer), cases7Days(integer), dataSource(string), date(string), deaths(integer), delta(integer), history(array), id(string), incidence7Days(number), laboratoryCount(integer), name(string), performedTests(integer), positiveTests(integer), positivityRate(number), quote(number), rValue4Days(number), rValue7Days(number), recovered(integer), stream(string), vaccinated(integer), weekIncidence(number)
- states_hospitalization_history:
  - primary key: id
  - fields: abbreviation(string), administeredVaccinations(integer), age_group(string), ags(string), calendarWeek(string), cases(integer), cases7Days(integer), dataSource(string), date(string), deaths(integer), delta(integer), history(array), id(string), incidence7Days(number), laboratoryCount(integer), name(string), performedTests(integer), positiveTests(integer), positivityRate(number), quote(number), rValue4Days(number), rValue7Days(number), recovered(integer), stream(string), vaccinated(integer), weekIncidence(number)
- states_age_groups:
  - primary key: id
  - fields: abbreviation(string), administeredVaccinations(integer), age_group(string), ags(string), calendarWeek(string), cases(integer), cases7Days(integer), dataSource(string), date(string), deaths(integer), delta(integer), history(array), id(string), incidence7Days(number), laboratoryCount(integer), name(string), performedTests(integer), positiveTests(integer), positivityRate(number), quote(number), rValue4Days(number), rValue7Days(number), recovered(integer), stream(string), vaccinated(integer), weekIncidence(number)
- districts_cases_history:
  - primary key: id
  - fields: abbreviation(string), administeredVaccinations(integer), age_group(string), ags(string), calendarWeek(string), cases(integer), cases7Days(integer), dataSource(string), date(string), deaths(integer), delta(integer), history(array), id(string), incidence7Days(number), laboratoryCount(integer), name(string), performedTests(integer), positiveTests(integer), positivityRate(number), quote(number), rValue4Days(number), rValue7Days(number), recovered(integer), stream(string), vaccinated(integer), weekIncidence(number)
- districts_deaths_history:
  - primary key: id
  - fields: abbreviation(string), administeredVaccinations(integer), age_group(string), ags(string), calendarWeek(string), cases(integer), cases7Days(integer), dataSource(string), date(string), deaths(integer), delta(integer), history(array), id(string), incidence7Days(number), laboratoryCount(integer), name(string), performedTests(integer), positiveTests(integer), positivityRate(number), quote(number), rValue4Days(number), rValue7Days(number), recovered(integer), stream(string), vaccinated(integer), weekIncidence(number)
- districts_incidence_history:
  - primary key: id
  - fields: abbreviation(string), administeredVaccinations(integer), age_group(string), ags(string), calendarWeek(string), cases(integer), cases7Days(integer), dataSource(string), date(string), deaths(integer), delta(integer), history(array), id(string), incidence7Days(number), laboratoryCount(integer), name(string), performedTests(integer), positiveTests(integer), positivityRate(number), quote(number), rValue4Days(number), rValue7Days(number), recovered(integer), stream(string), vaccinated(integer), weekIncidence(number)
- districts_recovered_history:
  - primary key: id
  - fields: abbreviation(string), administeredVaccinations(integer), age_group(string), ags(string), calendarWeek(string), cases(integer), cases7Days(integer), dataSource(string), date(string), deaths(integer), delta(integer), history(array), id(string), incidence7Days(number), laboratoryCount(integer), name(string), performedTests(integer), positiveTests(integer), positivityRate(number), quote(number), rValue4Days(number), rValue7Days(number), recovered(integer), stream(string), vaccinated(integer), weekIncidence(number)
- districts_frozen_incidence_history:
  - primary key: id
  - fields: abbreviation(string), administeredVaccinations(integer), age_group(string), ags(string), calendarWeek(string), cases(integer), cases7Days(integer), dataSource(string), date(string), deaths(integer), delta(integer), history(array), id(string), incidence7Days(number), laboratoryCount(integer), name(string), performedTests(integer), positiveTests(integer), positivityRate(number), quote(number), rValue4Days(number), rValue7Days(number), recovered(integer), stream(string), vaccinated(integer), weekIncidence(number)
- districts_age_groups:
  - primary key: id
  - fields: abbreviation(string), administeredVaccinations(integer), age_group(string), ags(string), calendarWeek(string), cases(integer), cases7Days(integer), dataSource(string), date(string), deaths(integer), delta(integer), history(array), id(string), incidence7Days(number), laboratoryCount(integer), name(string), performedTests(integer), positiveTests(integer), positivityRate(number), quote(number), rValue4Days(number), rValue7Days(number), recovered(integer), stream(string), vaccinated(integer), weekIncidence(number)
- testing_history:
  - primary key: id
  - fields: abbreviation(string), administeredVaccinations(integer), age_group(string), ags(string), calendarWeek(string), cases(integer), cases7Days(integer), dataSource(string), date(string), deaths(integer), delta(integer), history(array), id(string), incidence7Days(number), laboratoryCount(integer), name(string), performedTests(integer), positiveTests(integer), positivityRate(number), quote(number), rValue4Days(number), rValue7Days(number), recovered(integer), stream(string), vaccinated(integer), weekIncidence(number)
- vaccinations:
  - primary key: id
  - fields: abbreviation(string), administeredVaccinations(integer), age_group(string), ags(string), calendarWeek(string), cases(integer), cases7Days(integer), dataSource(string), date(string), deaths(integer), delta(integer), history(array), id(string), incidence7Days(number), laboratoryCount(integer), name(string), performedTests(integer), positiveTests(integer), positivityRate(number), quote(number), rValue4Days(number), rValue7Days(number), recovered(integer), stream(string), vaccinated(integer), weekIncidence(number)
- vaccinations_states:
  - primary key: id
  - fields: abbreviation(string), administeredVaccinations(integer), age_group(string), ags(string), calendarWeek(string), cases(integer), cases7Days(integer), dataSource(string), date(string), deaths(integer), delta(integer), history(array), id(string), incidence7Days(number), laboratoryCount(integer), name(string), performedTests(integer), positiveTests(integer), positivityRate(number), quote(number), rValue4Days(number), rValue7Days(number), recovered(integer), stream(string), vaccinated(integer), weekIncidence(number)
- vaccinations_history:
  - primary key: id
  - cursor: date
  - fields: abbreviation(string), administeredVaccinations(integer), age_group(string), ags(string), calendarWeek(string), cases(integer), cases7Days(integer), dataSource(string), date(string), deaths(integer), delta(integer), history(array), id(string), incidence7Days(number), laboratoryCount(integer), name(string), performedTests(integer), positiveTests(integer), positivityRate(number), quote(number), rValue4Days(number), rValue7Days(number), recovered(integer), stream(string), vaccinated(integer), weekIncidence(number)

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped, incremental_append, incremental_append_deduped

## Security

- read risk: external corona-zahlen.org public JSON API read of Germany COVID metrics
- approval: none; read-only public data API, no credentials
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect rki-covid
```

### Inspect as structured JSON

```bash
pm connectors inspect rki-covid --json
```

## Agent Rules

- Run pm connectors inspect rki-covid before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.

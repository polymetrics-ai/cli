---
name: pm-rki-covid
description: RKI COVID connector knowledge and safe action guide.
---

# pm-rki-covid

## Purpose

Reads public Germany COVID case, state, district, and history data derived from RKI reports via the corona-zahlen.org JSON API. Read-only, credential-free.

## Icon

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
  - fields: abbreviation(), administeredVaccinations(), age_group(), ags(), calendarWeek(), cases(), cases7Days(), dataSource(), date(), deaths(), delta(), history(), id(), incidence7Days(), laboratoryCount(), name(), performedTests(), positiveTests(), positivityRate(), quote(), rValue4Days(), rValue7Days(), recovered(), stream(), vaccinated(), weekIncidence()
- states:
  - primary key: id
  - fields: abbreviation(), administeredVaccinations(), age_group(), ags(), calendarWeek(), cases(), cases7Days(), dataSource(), date(), deaths(), delta(), history(), id(), incidence7Days(), laboratoryCount(), name(), performedTests(), positiveTests(), positivityRate(), quote(), rValue4Days(), rValue7Days(), recovered(), stream(), vaccinated(), weekIncidence()
- districts:
  - primary key: id
  - fields: abbreviation(), administeredVaccinations(), age_group(), ags(), calendarWeek(), cases(), cases7Days(), dataSource(), date(), deaths(), delta(), history(), id(), incidence7Days(), laboratoryCount(), name(), performedTests(), positiveTests(), positivityRate(), quote(), rValue4Days(), rValue7Days(), recovered(), stream(), vaccinated(), weekIncidence()
- cases_history:
  - primary key: id
  - cursor: date
  - fields: abbreviation(), administeredVaccinations(), age_group(), ags(), calendarWeek(), cases(), cases7Days(), dataSource(), date(), deaths(), delta(), history(), id(), incidence7Days(), laboratoryCount(), name(), performedTests(), positiveTests(), positivityRate(), quote(), rValue4Days(), rValue7Days(), recovered(), stream(), vaccinated(), weekIncidence()
- deaths_history:
  - primary key: id
  - cursor: date
  - fields: abbreviation(), administeredVaccinations(), age_group(), ags(), calendarWeek(), cases(), cases7Days(), dataSource(), date(), deaths(), delta(), history(), id(), incidence7Days(), laboratoryCount(), name(), performedTests(), positiveTests(), positivityRate(), quote(), rValue4Days(), rValue7Days(), recovered(), stream(), vaccinated(), weekIncidence()
- germany_incidence_history:
  - primary key: id
  - cursor: date
  - fields: abbreviation(), administeredVaccinations(), age_group(), ags(), calendarWeek(), cases(), cases7Days(), dataSource(), date(), deaths(), delta(), history(), id(), incidence7Days(), laboratoryCount(), name(), performedTests(), positiveTests(), positivityRate(), quote(), rValue4Days(), rValue7Days(), recovered(), stream(), vaccinated(), weekIncidence()
- germany_recovered_history:
  - primary key: id
  - cursor: date
  - fields: abbreviation(), administeredVaccinations(), age_group(), ags(), calendarWeek(), cases(), cases7Days(), dataSource(), date(), deaths(), delta(), history(), id(), incidence7Days(), laboratoryCount(), name(), performedTests(), positiveTests(), positivityRate(), quote(), rValue4Days(), rValue7Days(), recovered(), stream(), vaccinated(), weekIncidence()
- germany_r_value_history:
  - primary key: id
  - cursor: date
  - fields: abbreviation(), administeredVaccinations(), age_group(), ags(), calendarWeek(), cases(), cases7Days(), dataSource(), date(), deaths(), delta(), history(), id(), incidence7Days(), laboratoryCount(), name(), performedTests(), positiveTests(), positivityRate(), quote(), rValue4Days(), rValue7Days(), recovered(), stream(), vaccinated(), weekIncidence()
- germany_hospitalization_history:
  - primary key: id
  - cursor: date
  - fields: abbreviation(), administeredVaccinations(), age_group(), ags(), calendarWeek(), cases(), cases7Days(), dataSource(), date(), deaths(), delta(), history(), id(), incidence7Days(), laboratoryCount(), name(), performedTests(), positiveTests(), positivityRate(), quote(), rValue4Days(), rValue7Days(), recovered(), stream(), vaccinated(), weekIncidence()
- germany_frozen_incidence_history:
  - primary key: id
  - cursor: date
  - fields: abbreviation(), administeredVaccinations(), age_group(), ags(), calendarWeek(), cases(), cases7Days(), dataSource(), date(), deaths(), delta(), history(), id(), incidence7Days(), laboratoryCount(), name(), performedTests(), positiveTests(), positivityRate(), quote(), rValue4Days(), rValue7Days(), recovered(), stream(), vaccinated(), weekIncidence()
- germany_age_groups:
  - primary key: id
  - fields: abbreviation(), administeredVaccinations(), age_group(), ags(), calendarWeek(), cases(), cases7Days(), dataSource(), date(), deaths(), delta(), history(), id(), incidence7Days(), laboratoryCount(), name(), performedTests(), positiveTests(), positivityRate(), quote(), rValue4Days(), rValue7Days(), recovered(), stream(), vaccinated(), weekIncidence()
- states_cases_history:
  - primary key: id
  - fields: abbreviation(), administeredVaccinations(), age_group(), ags(), calendarWeek(), cases(), cases7Days(), dataSource(), date(), deaths(), delta(), history(), id(), incidence7Days(), laboratoryCount(), name(), performedTests(), positiveTests(), positivityRate(), quote(), rValue4Days(), rValue7Days(), recovered(), stream(), vaccinated(), weekIncidence()
- states_deaths_history:
  - primary key: id
  - fields: abbreviation(), administeredVaccinations(), age_group(), ags(), calendarWeek(), cases(), cases7Days(), dataSource(), date(), deaths(), delta(), history(), id(), incidence7Days(), laboratoryCount(), name(), performedTests(), positiveTests(), positivityRate(), quote(), rValue4Days(), rValue7Days(), recovered(), stream(), vaccinated(), weekIncidence()
- states_incidence_history:
  - primary key: id
  - fields: abbreviation(), administeredVaccinations(), age_group(), ags(), calendarWeek(), cases(), cases7Days(), dataSource(), date(), deaths(), delta(), history(), id(), incidence7Days(), laboratoryCount(), name(), performedTests(), positiveTests(), positivityRate(), quote(), rValue4Days(), rValue7Days(), recovered(), stream(), vaccinated(), weekIncidence()
- states_recovered_history:
  - primary key: id
  - fields: abbreviation(), administeredVaccinations(), age_group(), ags(), calendarWeek(), cases(), cases7Days(), dataSource(), date(), deaths(), delta(), history(), id(), incidence7Days(), laboratoryCount(), name(), performedTests(), positiveTests(), positivityRate(), quote(), rValue4Days(), rValue7Days(), recovered(), stream(), vaccinated(), weekIncidence()
- states_frozen_incidence_history:
  - primary key: id
  - fields: abbreviation(), administeredVaccinations(), age_group(), ags(), calendarWeek(), cases(), cases7Days(), dataSource(), date(), deaths(), delta(), history(), id(), incidence7Days(), laboratoryCount(), name(), performedTests(), positiveTests(), positivityRate(), quote(), rValue4Days(), rValue7Days(), recovered(), stream(), vaccinated(), weekIncidence()
- states_hospitalization_history:
  - primary key: id
  - fields: abbreviation(), administeredVaccinations(), age_group(), ags(), calendarWeek(), cases(), cases7Days(), dataSource(), date(), deaths(), delta(), history(), id(), incidence7Days(), laboratoryCount(), name(), performedTests(), positiveTests(), positivityRate(), quote(), rValue4Days(), rValue7Days(), recovered(), stream(), vaccinated(), weekIncidence()
- states_age_groups:
  - primary key: id
  - fields: abbreviation(), administeredVaccinations(), age_group(), ags(), calendarWeek(), cases(), cases7Days(), dataSource(), date(), deaths(), delta(), history(), id(), incidence7Days(), laboratoryCount(), name(), performedTests(), positiveTests(), positivityRate(), quote(), rValue4Days(), rValue7Days(), recovered(), stream(), vaccinated(), weekIncidence()
- districts_cases_history:
  - primary key: id
  - fields: abbreviation(), administeredVaccinations(), age_group(), ags(), calendarWeek(), cases(), cases7Days(), dataSource(), date(), deaths(), delta(), history(), id(), incidence7Days(), laboratoryCount(), name(), performedTests(), positiveTests(), positivityRate(), quote(), rValue4Days(), rValue7Days(), recovered(), stream(), vaccinated(), weekIncidence()
- districts_deaths_history:
  - primary key: id
  - fields: abbreviation(), administeredVaccinations(), age_group(), ags(), calendarWeek(), cases(), cases7Days(), dataSource(), date(), deaths(), delta(), history(), id(), incidence7Days(), laboratoryCount(), name(), performedTests(), positiveTests(), positivityRate(), quote(), rValue4Days(), rValue7Days(), recovered(), stream(), vaccinated(), weekIncidence()
- districts_incidence_history:
  - primary key: id
  - fields: abbreviation(), administeredVaccinations(), age_group(), ags(), calendarWeek(), cases(), cases7Days(), dataSource(), date(), deaths(), delta(), history(), id(), incidence7Days(), laboratoryCount(), name(), performedTests(), positiveTests(), positivityRate(), quote(), rValue4Days(), rValue7Days(), recovered(), stream(), vaccinated(), weekIncidence()
- districts_recovered_history:
  - primary key: id
  - fields: abbreviation(), administeredVaccinations(), age_group(), ags(), calendarWeek(), cases(), cases7Days(), dataSource(), date(), deaths(), delta(), history(), id(), incidence7Days(), laboratoryCount(), name(), performedTests(), positiveTests(), positivityRate(), quote(), rValue4Days(), rValue7Days(), recovered(), stream(), vaccinated(), weekIncidence()
- districts_frozen_incidence_history:
  - primary key: id
  - fields: abbreviation(), administeredVaccinations(), age_group(), ags(), calendarWeek(), cases(), cases7Days(), dataSource(), date(), deaths(), delta(), history(), id(), incidence7Days(), laboratoryCount(), name(), performedTests(), positiveTests(), positivityRate(), quote(), rValue4Days(), rValue7Days(), recovered(), stream(), vaccinated(), weekIncidence()
- districts_age_groups:
  - primary key: id
  - fields: abbreviation(), administeredVaccinations(), age_group(), ags(), calendarWeek(), cases(), cases7Days(), dataSource(), date(), deaths(), delta(), history(), id(), incidence7Days(), laboratoryCount(), name(), performedTests(), positiveTests(), positivityRate(), quote(), rValue4Days(), rValue7Days(), recovered(), stream(), vaccinated(), weekIncidence()
- testing_history:
  - primary key: id
  - fields: abbreviation(), administeredVaccinations(), age_group(), ags(), calendarWeek(), cases(), cases7Days(), dataSource(), date(), deaths(), delta(), history(), id(), incidence7Days(), laboratoryCount(), name(), performedTests(), positiveTests(), positivityRate(), quote(), rValue4Days(), rValue7Days(), recovered(), stream(), vaccinated(), weekIncidence()
- vaccinations:
  - primary key: id
  - fields: abbreviation(), administeredVaccinations(), age_group(), ags(), calendarWeek(), cases(), cases7Days(), dataSource(), date(), deaths(), delta(), history(), id(), incidence7Days(), laboratoryCount(), name(), performedTests(), positiveTests(), positivityRate(), quote(), rValue4Days(), rValue7Days(), recovered(), stream(), vaccinated(), weekIncidence()
- vaccinations_states:
  - primary key: id
  - fields: abbreviation(), administeredVaccinations(), age_group(), ags(), calendarWeek(), cases(), cases7Days(), dataSource(), date(), deaths(), delta(), history(), id(), incidence7Days(), laboratoryCount(), name(), performedTests(), positiveTests(), positivityRate(), quote(), rValue4Days(), rValue7Days(), recovered(), stream(), vaccinated(), weekIncidence()
- vaccinations_history:
  - primary key: id
  - cursor: date
  - fields: abbreviation(), administeredVaccinations(), age_group(), ags(), calendarWeek(), cases(), cases7Days(), dataSource(), date(), deaths(), delta(), history(), id(), incidence7Days(), laboratoryCount(), name(), performedTests(), positiveTests(), positivityRate(), quote(), rValue4Days(), rValue7Days(), recovered(), stream(), vaccinated(), weekIncidence()

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

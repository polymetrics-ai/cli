# pm connectors inspect rki-covid

```text
NAME
  pm connectors inspect rki-covid - RKI COVID connector manual

SYNOPSIS
  pm connectors inspect rki-covid
  pm connectors inspect rki-covid --json
  pm credentials add <name> --connector rki-covid [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads public Germany COVID case, state, district, and history data derived from RKI reports via the corona-zahlen.org JSON API. Read-only, credential-free.

ICON
  id: rki
  asset: icons/rki.svg
  source: upstream_registry
  review_status: upstream_seeded

CAPABILITIES
  check=true catalog=true read=true write=false query=false
  Integration type: api

AUTHENTICATION
  No secret authentication is required for this connector.

CONFIGURATION
  base_url
  days

ETL STREAMS
  germany:
    primary key: id
    fields: abbreviation(string), administeredVaccinations(integer), age_group(string), ags(string), calendarWeek(string), cases(integer), cases7Days(integer), dataSource(string), date(string), deaths(integer), delta(integer), history(array), id(string), incidence7Days(number), laboratoryCount(integer), name(string), performedTests(integer), positiveTests(integer), positivityRate(number), quote(number), rValue4Days(number), rValue7Days(number), recovered(integer), stream(string), vaccinated(integer), weekIncidence(number)
  states:
    primary key: id
    fields: abbreviation(string), administeredVaccinations(integer), age_group(string), ags(string), calendarWeek(string), cases(integer), cases7Days(integer), dataSource(string), date(string), deaths(integer), delta(integer), history(array), id(string), incidence7Days(number), laboratoryCount(integer), name(string), performedTests(integer), positiveTests(integer), positivityRate(number), quote(number), rValue4Days(number), rValue7Days(number), recovered(integer), stream(string), vaccinated(integer), weekIncidence(number)
  districts:
    primary key: id
    fields: abbreviation(string), administeredVaccinations(integer), age_group(string), ags(string), calendarWeek(string), cases(integer), cases7Days(integer), dataSource(string), date(string), deaths(integer), delta(integer), history(array), id(string), incidence7Days(number), laboratoryCount(integer), name(string), performedTests(integer), positiveTests(integer), positivityRate(number), quote(number), rValue4Days(number), rValue7Days(number), recovered(integer), stream(string), vaccinated(integer), weekIncidence(number)
  cases_history:
    primary key: id
    cursor: date
    fields: abbreviation(string), administeredVaccinations(integer), age_group(string), ags(string), calendarWeek(string), cases(integer), cases7Days(integer), dataSource(string), date(string), deaths(integer), delta(integer), history(array), id(string), incidence7Days(number), laboratoryCount(integer), name(string), performedTests(integer), positiveTests(integer), positivityRate(number), quote(number), rValue4Days(number), rValue7Days(number), recovered(integer), stream(string), vaccinated(integer), weekIncidence(number)
  deaths_history:
    primary key: id
    cursor: date
    fields: abbreviation(string), administeredVaccinations(integer), age_group(string), ags(string), calendarWeek(string), cases(integer), cases7Days(integer), dataSource(string), date(string), deaths(integer), delta(integer), history(array), id(string), incidence7Days(number), laboratoryCount(integer), name(string), performedTests(integer), positiveTests(integer), positivityRate(number), quote(number), rValue4Days(number), rValue7Days(number), recovered(integer), stream(string), vaccinated(integer), weekIncidence(number)
  germany_incidence_history:
    primary key: id
    cursor: date
    fields: abbreviation(string), administeredVaccinations(integer), age_group(string), ags(string), calendarWeek(string), cases(integer), cases7Days(integer), dataSource(string), date(string), deaths(integer), delta(integer), history(array), id(string), incidence7Days(number), laboratoryCount(integer), name(string), performedTests(integer), positiveTests(integer), positivityRate(number), quote(number), rValue4Days(number), rValue7Days(number), recovered(integer), stream(string), vaccinated(integer), weekIncidence(number)
  germany_recovered_history:
    primary key: id
    cursor: date
    fields: abbreviation(string), administeredVaccinations(integer), age_group(string), ags(string), calendarWeek(string), cases(integer), cases7Days(integer), dataSource(string), date(string), deaths(integer), delta(integer), history(array), id(string), incidence7Days(number), laboratoryCount(integer), name(string), performedTests(integer), positiveTests(integer), positivityRate(number), quote(number), rValue4Days(number), rValue7Days(number), recovered(integer), stream(string), vaccinated(integer), weekIncidence(number)
  germany_r_value_history:
    primary key: id
    cursor: date
    fields: abbreviation(string), administeredVaccinations(integer), age_group(string), ags(string), calendarWeek(string), cases(integer), cases7Days(integer), dataSource(string), date(string), deaths(integer), delta(integer), history(array), id(string), incidence7Days(number), laboratoryCount(integer), name(string), performedTests(integer), positiveTests(integer), positivityRate(number), quote(number), rValue4Days(number), rValue7Days(number), recovered(integer), stream(string), vaccinated(integer), weekIncidence(number)
  germany_hospitalization_history:
    primary key: id
    cursor: date
    fields: abbreviation(string), administeredVaccinations(integer), age_group(string), ags(string), calendarWeek(string), cases(integer), cases7Days(integer), dataSource(string), date(string), deaths(integer), delta(integer), history(array), id(string), incidence7Days(number), laboratoryCount(integer), name(string), performedTests(integer), positiveTests(integer), positivityRate(number), quote(number), rValue4Days(number), rValue7Days(number), recovered(integer), stream(string), vaccinated(integer), weekIncidence(number)
  germany_frozen_incidence_history:
    primary key: id
    cursor: date
    fields: abbreviation(string), administeredVaccinations(integer), age_group(string), ags(string), calendarWeek(string), cases(integer), cases7Days(integer), dataSource(string), date(string), deaths(integer), delta(integer), history(array), id(string), incidence7Days(number), laboratoryCount(integer), name(string), performedTests(integer), positiveTests(integer), positivityRate(number), quote(number), rValue4Days(number), rValue7Days(number), recovered(integer), stream(string), vaccinated(integer), weekIncidence(number)
  germany_age_groups:
    primary key: id
    fields: abbreviation(string), administeredVaccinations(integer), age_group(string), ags(string), calendarWeek(string), cases(integer), cases7Days(integer), dataSource(string), date(string), deaths(integer), delta(integer), history(array), id(string), incidence7Days(number), laboratoryCount(integer), name(string), performedTests(integer), positiveTests(integer), positivityRate(number), quote(number), rValue4Days(number), rValue7Days(number), recovered(integer), stream(string), vaccinated(integer), weekIncidence(number)
  states_cases_history:
    primary key: id
    fields: abbreviation(string), administeredVaccinations(integer), age_group(string), ags(string), calendarWeek(string), cases(integer), cases7Days(integer), dataSource(string), date(string), deaths(integer), delta(integer), history(array), id(string), incidence7Days(number), laboratoryCount(integer), name(string), performedTests(integer), positiveTests(integer), positivityRate(number), quote(number), rValue4Days(number), rValue7Days(number), recovered(integer), stream(string), vaccinated(integer), weekIncidence(number)
  states_deaths_history:
    primary key: id
    fields: abbreviation(string), administeredVaccinations(integer), age_group(string), ags(string), calendarWeek(string), cases(integer), cases7Days(integer), dataSource(string), date(string), deaths(integer), delta(integer), history(array), id(string), incidence7Days(number), laboratoryCount(integer), name(string), performedTests(integer), positiveTests(integer), positivityRate(number), quote(number), rValue4Days(number), rValue7Days(number), recovered(integer), stream(string), vaccinated(integer), weekIncidence(number)
  states_incidence_history:
    primary key: id
    fields: abbreviation(string), administeredVaccinations(integer), age_group(string), ags(string), calendarWeek(string), cases(integer), cases7Days(integer), dataSource(string), date(string), deaths(integer), delta(integer), history(array), id(string), incidence7Days(number), laboratoryCount(integer), name(string), performedTests(integer), positiveTests(integer), positivityRate(number), quote(number), rValue4Days(number), rValue7Days(number), recovered(integer), stream(string), vaccinated(integer), weekIncidence(number)
  states_recovered_history:
    primary key: id
    fields: abbreviation(string), administeredVaccinations(integer), age_group(string), ags(string), calendarWeek(string), cases(integer), cases7Days(integer), dataSource(string), date(string), deaths(integer), delta(integer), history(array), id(string), incidence7Days(number), laboratoryCount(integer), name(string), performedTests(integer), positiveTests(integer), positivityRate(number), quote(number), rValue4Days(number), rValue7Days(number), recovered(integer), stream(string), vaccinated(integer), weekIncidence(number)
  states_frozen_incidence_history:
    primary key: id
    fields: abbreviation(string), administeredVaccinations(integer), age_group(string), ags(string), calendarWeek(string), cases(integer), cases7Days(integer), dataSource(string), date(string), deaths(integer), delta(integer), history(array), id(string), incidence7Days(number), laboratoryCount(integer), name(string), performedTests(integer), positiveTests(integer), positivityRate(number), quote(number), rValue4Days(number), rValue7Days(number), recovered(integer), stream(string), vaccinated(integer), weekIncidence(number)
  states_hospitalization_history:
    primary key: id
    fields: abbreviation(string), administeredVaccinations(integer), age_group(string), ags(string), calendarWeek(string), cases(integer), cases7Days(integer), dataSource(string), date(string), deaths(integer), delta(integer), history(array), id(string), incidence7Days(number), laboratoryCount(integer), name(string), performedTests(integer), positiveTests(integer), positivityRate(number), quote(number), rValue4Days(number), rValue7Days(number), recovered(integer), stream(string), vaccinated(integer), weekIncidence(number)
  states_age_groups:
    primary key: id
    fields: abbreviation(string), administeredVaccinations(integer), age_group(string), ags(string), calendarWeek(string), cases(integer), cases7Days(integer), dataSource(string), date(string), deaths(integer), delta(integer), history(array), id(string), incidence7Days(number), laboratoryCount(integer), name(string), performedTests(integer), positiveTests(integer), positivityRate(number), quote(number), rValue4Days(number), rValue7Days(number), recovered(integer), stream(string), vaccinated(integer), weekIncidence(number)
  districts_cases_history:
    primary key: id
    fields: abbreviation(string), administeredVaccinations(integer), age_group(string), ags(string), calendarWeek(string), cases(integer), cases7Days(integer), dataSource(string), date(string), deaths(integer), delta(integer), history(array), id(string), incidence7Days(number), laboratoryCount(integer), name(string), performedTests(integer), positiveTests(integer), positivityRate(number), quote(number), rValue4Days(number), rValue7Days(number), recovered(integer), stream(string), vaccinated(integer), weekIncidence(number)
  districts_deaths_history:
    primary key: id
    fields: abbreviation(string), administeredVaccinations(integer), age_group(string), ags(string), calendarWeek(string), cases(integer), cases7Days(integer), dataSource(string), date(string), deaths(integer), delta(integer), history(array), id(string), incidence7Days(number), laboratoryCount(integer), name(string), performedTests(integer), positiveTests(integer), positivityRate(number), quote(number), rValue4Days(number), rValue7Days(number), recovered(integer), stream(string), vaccinated(integer), weekIncidence(number)
  districts_incidence_history:
    primary key: id
    fields: abbreviation(string), administeredVaccinations(integer), age_group(string), ags(string), calendarWeek(string), cases(integer), cases7Days(integer), dataSource(string), date(string), deaths(integer), delta(integer), history(array), id(string), incidence7Days(number), laboratoryCount(integer), name(string), performedTests(integer), positiveTests(integer), positivityRate(number), quote(number), rValue4Days(number), rValue7Days(number), recovered(integer), stream(string), vaccinated(integer), weekIncidence(number)
  districts_recovered_history:
    primary key: id
    fields: abbreviation(string), administeredVaccinations(integer), age_group(string), ags(string), calendarWeek(string), cases(integer), cases7Days(integer), dataSource(string), date(string), deaths(integer), delta(integer), history(array), id(string), incidence7Days(number), laboratoryCount(integer), name(string), performedTests(integer), positiveTests(integer), positivityRate(number), quote(number), rValue4Days(number), rValue7Days(number), recovered(integer), stream(string), vaccinated(integer), weekIncidence(number)
  districts_frozen_incidence_history:
    primary key: id
    fields: abbreviation(string), administeredVaccinations(integer), age_group(string), ags(string), calendarWeek(string), cases(integer), cases7Days(integer), dataSource(string), date(string), deaths(integer), delta(integer), history(array), id(string), incidence7Days(number), laboratoryCount(integer), name(string), performedTests(integer), positiveTests(integer), positivityRate(number), quote(number), rValue4Days(number), rValue7Days(number), recovered(integer), stream(string), vaccinated(integer), weekIncidence(number)
  districts_age_groups:
    primary key: id
    fields: abbreviation(string), administeredVaccinations(integer), age_group(string), ags(string), calendarWeek(string), cases(integer), cases7Days(integer), dataSource(string), date(string), deaths(integer), delta(integer), history(array), id(string), incidence7Days(number), laboratoryCount(integer), name(string), performedTests(integer), positiveTests(integer), positivityRate(number), quote(number), rValue4Days(number), rValue7Days(number), recovered(integer), stream(string), vaccinated(integer), weekIncidence(number)
  testing_history:
    primary key: id
    fields: abbreviation(string), administeredVaccinations(integer), age_group(string), ags(string), calendarWeek(string), cases(integer), cases7Days(integer), dataSource(string), date(string), deaths(integer), delta(integer), history(array), id(string), incidence7Days(number), laboratoryCount(integer), name(string), performedTests(integer), positiveTests(integer), positivityRate(number), quote(number), rValue4Days(number), rValue7Days(number), recovered(integer), stream(string), vaccinated(integer), weekIncidence(number)
  vaccinations:
    primary key: id
    fields: abbreviation(string), administeredVaccinations(integer), age_group(string), ags(string), calendarWeek(string), cases(integer), cases7Days(integer), dataSource(string), date(string), deaths(integer), delta(integer), history(array), id(string), incidence7Days(number), laboratoryCount(integer), name(string), performedTests(integer), positiveTests(integer), positivityRate(number), quote(number), rValue4Days(number), rValue7Days(number), recovered(integer), stream(string), vaccinated(integer), weekIncidence(number)
  vaccinations_states:
    primary key: id
    fields: abbreviation(string), administeredVaccinations(integer), age_group(string), ags(string), calendarWeek(string), cases(integer), cases7Days(integer), dataSource(string), date(string), deaths(integer), delta(integer), history(array), id(string), incidence7Days(number), laboratoryCount(integer), name(string), performedTests(integer), positiveTests(integer), positivityRate(number), quote(number), rValue4Days(number), rValue7Days(number), recovered(integer), stream(string), vaccinated(integer), weekIncidence(number)
  vaccinations_history:
    primary key: id
    cursor: date
    fields: abbreviation(string), administeredVaccinations(integer), age_group(string), ags(string), calendarWeek(string), cases(integer), cases7Days(integer), dataSource(string), date(string), deaths(integer), delta(integer), history(array), id(string), incidence7Days(number), laboratoryCount(integer), name(string), performedTests(integer), positiveTests(integer), positivityRate(number), quote(number), rValue4Days(number), rValue7Days(number), recovered(integer), stream(string), vaccinated(integer), weekIncidence(number)

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped, incremental_append, incremental_append_deduped

SECURITY
  read risk: external corona-zahlen.org public JSON API read of Germany COVID metrics
  approval: none; read-only public data API, no credentials
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect rki-covid

  # Inspect as structured JSON
  pm connectors inspect rki-covid --json

AGENT WORKFLOW
  - Run pm connectors inspect rki-covid before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```

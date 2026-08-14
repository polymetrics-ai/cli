# pm connectors inspect zenefits

```text
NAME
  pm connectors inspect zenefits - Zenefits connector manual

SYNOPSIS
  pm connectors inspect zenefits
  pm connectors inspect zenefits --json
  pm credentials add <name> --connector zenefits [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Zenefits people, companies, departments, locations, employments, custom fields/values, bank accounts, labor groups, and time-off data.

ICON
  id: pm-sample
  asset: icons/pm-sample.svg
  source: polymetrics
  review_status: polymetrics
  review_url: https://github.com/polymetrics-ai/cli

CAPABILITIES
  check=true catalog=true read=true write=false query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  base_url
  token (secret) (required)

ETL STREAMS
  people:
    primary key: id
    fields: first_name(string), id(string), last_name(string), status(string)
  companies:
    primary key: id
    fields: id(string), name(string)
  departments:
    primary key: id
    fields: id(string), name(string)
  locations:
    primary key: id
    fields: city(string), company(object), country(string), id(string), name(string), people(object), phone(string), state(string), street1(string), street2(string), zip(string)
  employments:
    primary key: id
    fields: annual_salary(string), comp_type(string), employment_type(string), hire_date(string), id(string), is_active(boolean), pay_rate(string), person(object), termination_date(string), termination_type(string), working_hours_per_week(string)
  custom_fields:
    primary key: id
    fields: can_manager_view_field(boolean), can_person_edit_field(boolean), can_person_view_field(boolean), custom_field_type(string), custom_field_values(object), help_text(string), help_url(string), id(string), is_field_required(boolean), is_sensitive(boolean), name(string)
  custom_field_values:
    primary key: id
    fields: custom_field(object), id(string), person(object), value(string)
  company_banks:
    primary key: id
    fields: account_number(string), account_type(string), bank_name(string), company(object), id(string), routing_number(string)
  employee_banks:
    primary key: id
    fields: account_number(string), account_type(string), amount_per_paycheck(string), bank_name(string), id(string), is_primary_account(boolean), is_salary_account(boolean), percentage_per_paycheck(string), person(object), priority(string), routing_number(string)
  labor_group_types:
    primary key: id
    fields: id(string), labor_groups(object), name(string)
  labor_groups:
    primary key: id
    fields: assigned_members(object), code(string), id(string), labor_group_type(object), name(string)
  vacation_types:
    primary key: id
    fields: company(object), counts_as(string), id(string), name(string), status(string), vacation_requests(object)
  vacation_requests:
    primary key: id
    fields: approved_date(string), created_date(string), creator(object), deny_reason(string), end_date(string), hours(string), id(string), person(object), reason(string), start_date(string), status(string), vacation_type(object)

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite

SECURITY
  read risk: external Zenefits account read of people, companies, departments, locations, employments, custom field definitions/values, company and employee bank account details, labor groups, and time-off vacation types/requests
  approval: none; read-only bearer token access. The entire documented Zenefits API is read-only (no write endpoint exists), so there is no write risk to assess
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect zenefits

  # Inspect as structured JSON
  pm connectors inspect zenefits --json

AGENT WORKFLOW
  - Run pm connectors inspect zenefits before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```

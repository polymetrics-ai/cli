# pm connectors inspect openfda

```text
NAME
  pm connectors inspect openfda - OpenFDA connector manual

SYNOPSIS
  pm connectors inspect openfda
  pm connectors inspect openfda --json
  pm credentials add <name> --connector openfda [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads documented FDA drug, device, food, animal/veterinary, cosmetics, tobacco, transparency, and other public datasets from the openFDA REST API.

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
  mode
  search
  api_key (secret)

ETL STREAMS
  drug_event:
    primary key: safetyreportid
    fields: fulfillexpeditecriteria(string), occurcountry(string), primarysourcecountry(string), receiptdate(string), receivedate(string), safetyreportid(string), safetyreportversion(string), serious(string), seriousnessdeath(string), transmissiondate(string)
  drug_label:
    primary key: id
    fields: effective_time(string), id(string), indications_and_usage(array), openfda(object), purpose(array), set_id(string), version(string), warnings(array)
  drug_enforcement:
    primary key: recall_number
    fields: classification(string), country(string), distribution_pattern(string), product_type(string), reason_for_recall(string), recall_initiation_date(string), recall_number(string), recalling_firm(string), report_date(string), state(string), status(string), voluntary_mandated(string)
  device_event:
    primary key: mdr_report_key
    fields: adverse_event_flag(string), date_of_event(string), date_received(string), event_type(string), manufacturer_name(string), mdr_report_key(string), product_problem_flag(string), report_number(string), report_source_code(string)
  food_enforcement:
    primary key: recall_number
    fields: classification(string), country(string), distribution_pattern(string), product_type(string), reason_for_recall(string), recall_initiation_date(string), recall_number(string), recalling_firm(string), report_date(string), state(string), status(string), voluntary_mandated(string)
  animalandveterinary_event:
    fields: id(string), openfda(object)
  cosmetic_event:
    fields: id(string), openfda(object)
  food_event:
    fields: id(string), openfda(object)
  drug_ndc:
    fields: id(string), openfda(object)
  drug_drugsfda:
    fields: id(string), openfda(object)
  drug_shortages:
    fields: id(string), openfda(object)
  drug_orangebook:
    fields: id(string), openfda(object)
  device_510k:
    fields: id(string), openfda(object)
  device_pma:
    fields: id(string), openfda(object)
  device_udi:
    fields: id(string), openfda(object)
  device_enforcement:
    fields: id(string), openfda(object)
  device_recall:
    fields: id(string), openfda(object)
  device_classification:
    fields: id(string), openfda(object)
  device_registrationlisting:
    fields: id(string), openfda(object)
  device_covid19serology:
    fields: id(string), openfda(object)
  tobacco_problem:
    fields: id(string), openfda(object)
  tobacco_researchdigitalads:
    fields: id(string), openfda(object)
  tobacco_researchpreventionads:
    fields: id(string), openfda(object)
  tobacco_researchsmokefree:
    fields: id(string), openfda(object)
  transparency_crl:
    fields: id(string), openfda(object)
  other_historicaldocument:
    fields: id(string), openfda(object)
  other_nsde:
    fields: id(string), openfda(object)
  other_substance:
    fields: id(string), openfda(object)
  other_unii:
    fields: id(string), openfda(object)

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite

SECURITY
  read risk: external openFDA API read of public FDA regulatory datasets
  approval: none; read-only public reference API
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect openfda

  # Inspect as structured JSON
  pm connectors inspect openfda --json

AGENT WORKFLOW
  - Run pm connectors inspect openfda before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```

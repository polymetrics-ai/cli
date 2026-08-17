---
name: pm-openfda
description: OpenFDA connector knowledge and safe action guide.
---

# pm-openfda

## Purpose

Reads documented FDA drug, device, food, animal/veterinary, cosmetics, tobacco, transparency, and other public datasets from the openFDA REST API.

## Icon

- asset: icons/pm-sample.svg
- source: polymetrics
- review_status: polymetrics

## Capabilities

- check=true catalog=true read=true write=false query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- base_url
- mode
- search
- api_key (secret)

## ETL Streams

- drug_event:
  - primary key: safetyreportid
  - fields: fulfillexpeditecriteria(), occurcountry(), primarysourcecountry(), receiptdate(), receivedate(), safetyreportid(), safetyreportversion(), serious(), seriousnessdeath(), transmissiondate()
- drug_label:
  - primary key: id
  - fields: effective_time(), id(), indications_and_usage(), openfda(), purpose(), set_id(), version(), warnings()
- drug_enforcement:
  - primary key: recall_number
  - fields: classification(), country(), distribution_pattern(), product_type(), reason_for_recall(), recall_initiation_date(), recall_number(), recalling_firm(), report_date(), state(), status(), voluntary_mandated()
- device_event:
  - primary key: mdr_report_key
  - fields: adverse_event_flag(), date_of_event(), date_received(), event_type(), manufacturer_name(), mdr_report_key(), product_problem_flag(), report_number(), report_source_code()
- food_enforcement:
  - primary key: recall_number
  - fields: classification(), country(), distribution_pattern(), product_type(), reason_for_recall(), recall_initiation_date(), recall_number(), recalling_firm(), report_date(), state(), status(), voluntary_mandated()
- animalandveterinary_event:
  - fields: id(), openfda()
- cosmetic_event:
  - fields: id(), openfda()
- food_event:
  - fields: id(), openfda()
- drug_ndc:
  - fields: id(), openfda()
- drug_drugsfda:
  - fields: id(), openfda()
- drug_shortages:
  - fields: id(), openfda()
- drug_orangebook:
  - fields: id(), openfda()
- device_510k:
  - fields: id(), openfda()
- device_pma:
  - fields: id(), openfda()
- device_udi:
  - fields: id(), openfda()
- device_enforcement:
  - fields: id(), openfda()
- device_recall:
  - fields: id(), openfda()
- device_classification:
  - fields: id(), openfda()
- device_registrationlisting:
  - fields: id(), openfda()
- device_covid19serology:
  - fields: id(), openfda()
- tobacco_problem:
  - fields: id(), openfda()
- tobacco_researchdigitalads:
  - fields: id(), openfda()
- tobacco_researchpreventionads:
  - fields: id(), openfda()
- tobacco_researchsmokefree:
  - fields: id(), openfda()
- transparency_crl:
  - fields: id(), openfda()
- other_historicaldocument:
  - fields: id(), openfda()
- other_nsde:
  - fields: id(), openfda()
- other_substance:
  - fields: id(), openfda()
- other_unii:
  - fields: id(), openfda()

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

## Security

- read risk: external openFDA API read of public FDA regulatory datasets
- approval: none; read-only public reference API
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect openfda
```

### Inspect as structured JSON

```bash
pm connectors inspect openfda --json
```

## Agent Rules

- Run pm connectors inspect openfda before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.

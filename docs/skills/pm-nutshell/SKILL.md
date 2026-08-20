---
name: pm-nutshell
description: Nutshell connector knowledge and safe action guide.
---

# pm-nutshell

## Purpose

Reads and writes documented Nutshell CRM REST resources through the Nutshell REST API.

## Icon

- id: pm-sample
- asset: icons/pm-sample.svg
- source: polymetrics
- review_status: polymetrics
- review_url: https://github.com/polymetrics-ai/cli

## Capabilities

- check=true catalog=true read=true write=true query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- account_id
- activity_id
- base_url
- competitor_id
- competitor_map_id
- contact_id
- edition_id
- email_id
- form_field_id
- form_id
- industry_id
- invoice_id
- lead_id
- market_id
- mode
- note_id
- outcome_id
- page_size
- product_category_id
- product_id
- product_map_id
- quote_id
- task_id
- user_id
- username (required)
- password (secret) (required)

## ETL Streams

- accounts:
  - primary key: id
  - cursor: modifiedTime
  - fields: accountTypeId(integer), createdTime(string), entityType(string), id(integer), industryId(integer), isHotLead(boolean), modifiedTime(string), name(string), url(string)
- contacts:
  - primary key: id
  - cursor: modifiedTime
  - fields: createdTime(string), description(string), entityType(string), htmlUrl(string), id(integer), modifiedTime(string), name(string)
- leads:
  - primary key: id
  - cursor: modifiedTime
  - fields: closedTime(string), confidence(integer), createdTime(string), entityType(string), id(integer), isOverdue(boolean), modifiedTime(string), name(string), status(integer), value(string)
- activities:
  - primary key: id
  - cursor: modifiedTime
  - fields: activityTypeId(integer), createdTime(string), description(string), entityType(string), id(integer), isFlagged(boolean), logNote(string), modifiedTime(string), name(string), status(integer)
- users:
  - primary key: id
  - fields: createdTime(string), emails(string), entityType(string), id(integer), isAdministrator(boolean), isEnabled(boolean), modifiedTime(string), name(string)
- account:
  - primary key: id
  - fields: createdTime(string), deletedTime(string), href(string), htmlUrl(string), id(string), modifiedTime(string), name(string)
- account_custom_fields:
  - primary key: id
  - fields: createdTime(string), deletedTime(string), href(string), htmlUrl(string), id(string), modifiedTime(string), name(string)
- account_custom_field_attributes:
  - primary key: id
  - fields: createdTime(string), deletedTime(string), href(string), htmlUrl(string), id(string), modifiedTime(string), name(string)
- account_list_items:
  - primary key: id
  - fields: createdTime(string), deletedTime(string), href(string), htmlUrl(string), id(string), modifiedTime(string), name(string)
- account_list_fields:
  - primary key: id
  - fields: createdTime(string), deletedTime(string), href(string), htmlUrl(string), id(string), modifiedTime(string), name(string)
- account_types:
  - primary key: id
  - fields: createdTime(string), deletedTime(string), href(string), htmlUrl(string), id(string), modifiedTime(string), name(string)
- activity:
  - primary key: id
  - fields: createdTime(string), deletedTime(string), href(string), htmlUrl(string), id(string), modifiedTime(string), name(string)
- activity_types:
  - primary key: id
  - fields: createdTime(string), deletedTime(string), href(string), htmlUrl(string), id(string), modifiedTime(string), name(string)
- audiences:
  - primary key: id
  - fields: createdTime(string), deletedTime(string), href(string), htmlUrl(string), id(string), modifiedTime(string), name(string)
- competitor_maps:
  - primary key: id
  - fields: createdTime(string), deletedTime(string), href(string), htmlUrl(string), id(string), modifiedTime(string), name(string)
- competitor_map:
  - primary key: id
  - fields: createdTime(string), deletedTime(string), href(string), htmlUrl(string), id(string), modifiedTime(string), name(string)
- competitors:
  - primary key: id
  - fields: createdTime(string), deletedTime(string), href(string), htmlUrl(string), id(string), modifiedTime(string), name(string)
- competitor:
  - primary key: id
  - fields: createdTime(string), deletedTime(string), href(string), htmlUrl(string), id(string), modifiedTime(string), name(string)
- contact:
  - primary key: id
  - fields: createdTime(string), deletedTime(string), href(string), htmlUrl(string), id(string), modifiedTime(string), name(string)
- contact_custom_fields:
  - primary key: id
  - fields: createdTime(string), deletedTime(string), href(string), htmlUrl(string), id(string), modifiedTime(string), name(string)
- contact_custom_field_attributes:
  - primary key: id
  - fields: createdTime(string), deletedTime(string), href(string), htmlUrl(string), id(string), modifiedTime(string), name(string)
- contact_list_items:
  - primary key: id
  - fields: createdTime(string), deletedTime(string), href(string), htmlUrl(string), id(string), modifiedTime(string), name(string)
- contact_list_fields:
  - primary key: id
  - fields: createdTime(string), deletedTime(string), href(string), htmlUrl(string), id(string), modifiedTime(string), name(string)
- editions:
  - primary key: id
  - fields: createdTime(string), deletedTime(string), href(string), htmlUrl(string), id(string), modifiedTime(string), name(string)
- edition:
  - primary key: id
  - fields: createdTime(string), deletedTime(string), href(string), htmlUrl(string), id(string), modifiedTime(string), name(string)
- email:
  - primary key: id
  - fields: createdTime(string), deletedTime(string), href(string), htmlUrl(string), id(string), modifiedTime(string), name(string)
- events:
  - primary key: id
  - fields: createdTime(string), deletedTime(string), href(string), htmlUrl(string), id(string), modifiedTime(string), name(string)
- deleted_events:
  - primary key: id
  - fields: createdTime(string), deletedTime(string), href(string), htmlUrl(string), id(string), modifiedTime(string), name(string)
- saved_filters:
  - primary key: id
  - fields: createdTime(string), deletedTime(string), href(string), htmlUrl(string), id(string), modifiedTime(string), name(string)
- forms:
  - primary key: id
  - fields: createdTime(string), deletedTime(string), href(string), htmlUrl(string), id(string), modifiedTime(string), name(string)
- form_field:
  - primary key: id
  - fields: createdTime(string), deletedTime(string), href(string), htmlUrl(string), id(string), modifiedTime(string), name(string)
- form:
  - primary key: id
  - fields: createdTime(string), deletedTime(string), href(string), htmlUrl(string), id(string), modifiedTime(string), name(string)
- industries:
  - primary key: id
  - fields: createdTime(string), deletedTime(string), href(string), htmlUrl(string), id(string), modifiedTime(string), name(string)
- industry:
  - primary key: id
  - fields: createdTime(string), deletedTime(string), href(string), htmlUrl(string), id(string), modifiedTime(string), name(string)
- invoices:
  - primary key: id
  - fields: createdTime(string), deletedTime(string), href(string), htmlUrl(string), id(string), modifiedTime(string), name(string)
- invoice:
  - primary key: id
  - fields: createdTime(string), deletedTime(string), href(string), htmlUrl(string), id(string), modifiedTime(string), name(string)
- lead:
  - primary key: id
  - fields: createdTime(string), deletedTime(string), href(string), htmlUrl(string), id(string), modifiedTime(string), name(string)
- lead_custom_fields:
  - primary key: id
  - fields: createdTime(string), deletedTime(string), href(string), htmlUrl(string), id(string), modifiedTime(string), name(string)
- lead_installments:
  - primary key: id
  - fields: createdTime(string), deletedTime(string), href(string), htmlUrl(string), id(string), modifiedTime(string), name(string)
- lead_stages:
  - primary key: id
  - fields: createdTime(string), deletedTime(string), href(string), htmlUrl(string), id(string), modifiedTime(string), name(string)
- lead_custom_field_attributes:
  - primary key: id
  - fields: createdTime(string), deletedTime(string), href(string), htmlUrl(string), id(string), modifiedTime(string), name(string)
- lead_list_items:
  - primary key: id
  - fields: createdTime(string), deletedTime(string), href(string), htmlUrl(string), id(string), modifiedTime(string), name(string)
- lead_list_fields:
  - primary key: id
  - fields: createdTime(string), deletedTime(string), href(string), htmlUrl(string), id(string), modifiedTime(string), name(string)
- lead_reports:
  - primary key: id
  - fields: createdTime(string), deletedTime(string), href(string), htmlUrl(string), id(string), modifiedTime(string), name(string)
- markets:
  - primary key: id
  - fields: createdTime(string), deletedTime(string), href(string), htmlUrl(string), id(string), modifiedTime(string), name(string)
- market:
  - primary key: id
  - fields: createdTime(string), deletedTime(string), href(string), htmlUrl(string), id(string), modifiedTime(string), name(string)
- notes:
  - primary key: id
  - fields: createdTime(string), deletedTime(string), href(string), htmlUrl(string), id(string), modifiedTime(string), name(string)
- note:
  - primary key: id
  - fields: createdTime(string), deletedTime(string), href(string), htmlUrl(string), id(string), modifiedTime(string), name(string)
- lead_outcomes:
  - primary key: id
  - fields: createdTime(string), deletedTime(string), href(string), htmlUrl(string), id(string), modifiedTime(string), name(string)
- lead_outcome:
  - primary key: id
  - fields: createdTime(string), deletedTime(string), href(string), htmlUrl(string), id(string), modifiedTime(string), name(string)
- product_categories:
  - primary key: id
  - fields: createdTime(string), deletedTime(string), href(string), htmlUrl(string), id(string), modifiedTime(string), name(string)
- product_category:
  - primary key: id
  - fields: createdTime(string), deletedTime(string), href(string), htmlUrl(string), id(string), modifiedTime(string), name(string)
- product_maps:
  - primary key: id
  - fields: createdTime(string), deletedTime(string), href(string), htmlUrl(string), id(string), modifiedTime(string), name(string)
- product_map:
  - primary key: id
  - fields: createdTime(string), deletedTime(string), href(string), htmlUrl(string), id(string), modifiedTime(string), name(string)
- products:
  - primary key: id
  - fields: createdTime(string), deletedTime(string), href(string), htmlUrl(string), id(string), modifiedTime(string), name(string)
- product:
  - primary key: id
  - fields: createdTime(string), deletedTime(string), href(string), htmlUrl(string), id(string), modifiedTime(string), name(string)
- quotes:
  - primary key: id
  - fields: createdTime(string), deletedTime(string), href(string), htmlUrl(string), id(string), modifiedTime(string), name(string)
- quote:
  - primary key: id
  - fields: createdTime(string), deletedTime(string), href(string), htmlUrl(string), id(string), modifiedTime(string), name(string)
- sources:
  - primary key: id
  - fields: createdTime(string), deletedTime(string), href(string), htmlUrl(string), id(string), modifiedTime(string), name(string)
- stages:
  - primary key: id
  - fields: createdTime(string), deletedTime(string), href(string), htmlUrl(string), id(string), modifiedTime(string), name(string)
- pipelines:
  - primary key: id
  - fields: createdTime(string), deletedTime(string), href(string), htmlUrl(string), id(string), modifiedTime(string), name(string)
- tags:
  - primary key: id
  - fields: createdTime(string), deletedTime(string), href(string), htmlUrl(string), id(string), modifiedTime(string), name(string)
- tasks:
  - primary key: id
  - fields: createdTime(string), deletedTime(string), href(string), htmlUrl(string), id(string), modifiedTime(string), name(string)
- task:
  - primary key: id
  - fields: createdTime(string), deletedTime(string), href(string), htmlUrl(string), id(string), modifiedTime(string), name(string)
- territories:
  - primary key: id
  - fields: createdTime(string), deletedTime(string), href(string), htmlUrl(string), id(string), modifiedTime(string), name(string)
- user:
  - primary key: id
  - fields: createdTime(string), deletedTime(string), href(string), htmlUrl(string), id(string), modifiedTime(string), name(string)

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

## Reverse ETL Actions

- create_account:
  - endpoint: POST /accounts
  - required fields: accounts
  - risk: creates company/account records in Nutshell
- delete_account:
  - endpoint: DELETE /accounts/{{ record.id }}
  - required fields: id
  - risk: deletes a Nutshell account/company; recoverable only through the undelete endpoint for a limited period
- undelete_account:
  - endpoint: POST /accounts/{{ record.id }}/undelete
  - required fields: id
  - risk: restores a deleted Nutshell account/company
- create_account_custom_field:
  - endpoint: POST /accounts/customfield
  - required fields: name, type
  - risk: creates an account custom field definition
- create_activity:
  - endpoint: POST /activities
  - required fields: activities
  - risk: creates Nutshell activity records
- update_activity:
  - endpoint: PUT /activities/{{ record.id }}
  - required fields: id, activities
  - risk: updates an existing Nutshell activity
- create_audience:
  - endpoint: POST /audiences
  - required fields: emAudiences
  - risk: creates a Nutshell email marketing audience
- delete_competitor_map:
  - endpoint: DELETE /competitormaps/{{ record.id }}
  - required fields: id
  - risk: deletes a lead-competitor relationship
- create_contact:
  - endpoint: POST /contacts
  - required fields: contacts
  - risk: creates person/contact records in Nutshell
- delete_contact:
  - endpoint: DELETE /contacts/{{ record.id }}
  - required fields: id
  - risk: deletes a Nutshell contact/person; recoverable only through the undelete endpoint for a limited period
- undelete_contact:
  - endpoint: POST /contacts/{{ record.id }}/undelete
  - required fields: id
  - risk: restores a deleted Nutshell contact/person
- create_contact_custom_field:
  - endpoint: POST /contacts/customfield
  - required fields: name, type
  - risk: creates a contact custom field definition
- create_lead:
  - endpoint: POST /leads
  - required fields: leads
  - risk: creates Nutshell lead records
- delete_lead:
  - endpoint: DELETE /leads/{{ record.id }}
  - required fields: id
  - risk: deletes a Nutshell lead; recoverable only through the undelete endpoint for a limited period
- reopen_lead:
  - endpoint: POST /leads/{{ record.id }}/reopen
  - required fields: id
  - risk: reopens a previously closed Nutshell lead
- set_lead_pipeline:
  - endpoint: POST /leads/{{ record.id }}/stageset
  - required fields: id, stageset
  - risk: changes the pipeline/stageset assigned to a lead
- update_lead_status:
  - endpoint: POST /leads/{{ record.id }}/status
  - required fields: id
  - risk: updates a lead status/outcome and optional competitor/product maps
- undelete_lead:
  - endpoint: POST /leads/{{ record.id }}/undelete
  - required fields: id
  - risk: restores a deleted Nutshell lead
- watch_lead:
  - endpoint: POST /leads/{{ record.id }}/watch
  - required fields: id
  - risk: toggles watch notifications for the authenticated user on a lead
- create_lead_custom_field:
  - endpoint: POST /leads/customfield
  - required fields: name, type
  - risk: creates a lead custom field definition
- create_note:
  - endpoint: POST /notes
  - required fields: data
  - risk: creates a note attached to a Nutshell entity
- delete_note:
  - endpoint: DELETE /notes/{{ record.id }}
  - required fields: id
  - risk: deletes a Nutshell note; recoverable only through the undelete endpoint for a limited period
- undelete_note:
  - endpoint: POST /notes/{{ record.id }}/undelete
  - required fields: id
  - risk: restores a deleted Nutshell note
- create_product_category:
  - endpoint: POST /productcategories
  - required fields: productCategories
  - risk: creates a Nutshell product category
- delete_product_map:
  - endpoint: DELETE /productMaps/{{ record.id }}
  - required fields: id
  - risk: deletes a product mapping from a lead
- delete_product:
  - endpoint: DELETE /products/{{ record.id }}
  - required fields: id
  - risk: deletes a Nutshell product; recoverable only through the undelete endpoint for a limited period
- undelete_product:
  - endpoint: POST /products/{{ record.id }}/undelete
  - required fields: id
  - risk: restores a deleted Nutshell product
- create_source:
  - endpoint: POST /sources
  - required fields: sources
  - risk: creates a lead source in Nutshell
- delete_source:
  - endpoint: DELETE /sources/{{ record.id }}
  - required fields: id
  - risk: deletes a lead source; recoverable only through the undelete endpoint for a limited period
- undelete_source:
  - endpoint: POST /sources/{{ record.id }}/undelete
  - required fields: id
  - risk: restores a deleted lead source
- create_tag:
  - endpoint: POST /tags
  - required fields: tags
  - risk: creates a Nutshell tag and optionally links entities
- delete_tag:
  - endpoint: DELETE /tags/{{ record.id }}
  - required fields: id
  - risk: deletes a Nutshell tag; recoverable only through the undelete endpoint for a limited period
- undelete_tag:
  - endpoint: POST /tags/{{ record.id }}/undelete
  - required fields: id
  - risk: restores a deleted Nutshell tag
- create_task:
  - endpoint: POST /tasks
  - required fields: title
  - risk: creates a task in Nutshell
- delete_task:
  - endpoint: DELETE /tasks/{{ record.id }}
  - required fields: id
  - risk: deletes a Nutshell task

## Security

- read risk: external Nutshell CRM read of account/contact/lead/activity/user and related CRM data
- write risk: external Nutshell CRM mutations including creates, updates, undeletes, watches, and destructive deletes
- approval: required for write actions; destructive delete actions carry destructive confirmation metadata
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect nutshell
```

### Inspect as structured JSON

```bash
pm connectors inspect nutshell --json
```

## Agent Rules

- Run pm connectors inspect nutshell before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
- For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.

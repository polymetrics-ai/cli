# pm connectors inspect hubplanner

```text
NAME
  pm connectors inspect hubplanner - Hubplanner connector manual

SYNOPSIS
  pm connectors inspect hubplanner
  pm connectors inspect hubplanner --json
  pm credentials add <name> --connector hubplanner [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Hubplanner scheduling, project, resource, client, billing, time, vacation, custom-field, and webhook-subscription data and exposes typed reverse-ETL writes for documented Hubplanner REST resources.

ICON
  id: hubplanner
  asset: icons/hubplanner.svg
  source: upstream_registry
  review_status: upstream_seeded

CAPABILITIES
  check=true catalog=true read=true write=true query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  base_url
  api_key (secret) (required)

ETL STREAMS
  resources:
    primary key: _id
    fields: _id(string), createdDate(string), email(string), firstName(string), lastName(string), note(string), role(string), status(string), type(string)
  projects:
    primary key: _id
    fields: _id(string), budgetCashAmount(number), budgetCurrency(string), budgetHours(number), createdDate(string), name(string), note(string), projectCode(string), status(string), updatedDate(string)
  clients:
    primary key: _id
    fields: _id(string), createdDate(string), email(string), name(string), note(string), phone(string)
  events:
    primary key: _id
    fields: _id(string), backgroundColor(string), createdDate(string), eventCode(string), metadata(string), name(string), updatedDate(string)
  holidays:
    primary key: _id
    fields: _id(string), date(string), end(string), holidayGroup(string), name(string), start(string)
  bookings:
    primary key: _id
    fields: _id(string), category(string), end(string), note(string), project(string), resource(string), start(string), state(string)
  billing_rates:
    primary key: _id
    fields: _id(string), createdDate(string), currency(string), label(string), metadata(string), rate(number), updatedDate(string)
  booking_categories:
    primary key: _id
    fields: _id(string), categoryGroupId(string), categoryGroupName(string), createdDate(string), gridColor(string), group(string), name(string), type(string), updatedDate(string)
  resource_custom_field_templates:
    primary key: _id
    fields: _id(string), allowMultipleValues(boolean), canResourceEdit(boolean), category(string), characterLimit(integer), choices(array), createdDate(string), defaultRadioId(string), defaultValue(string), filterGrid(boolean), instructions(string), isChoicesSortedAlphabetically(boolean), isRequired(boolean), label(string), maxValue(number), minValue(number), placeholderText(string), status(string), stepValue(number), type(string), updatedDate(string), weekStartOn(integer)
  project_custom_field_templates:
    primary key: _id
    fields: _id(string), allowMultipleValues(boolean), canResourceEdit(boolean), category(string), characterLimit(integer), choices(array), createdDate(string), defaultRadioId(string), defaultValue(string), filterGrid(boolean), instructions(string), isChoicesSortedAlphabetically(boolean), isRequired(boolean), label(string), maxValue(number), minValue(number), placeholderText(string), status(string), stepValue(number), type(string), updatedDate(string), weekStartOn(integer)
  project_groups:
    primary key: _id
    fields: _id(string), createdDate(string), metadata(string), name(string), parentGroupId(string), projects(array), updatedDate(string)
  resource_groups:
    primary key: _id
    fields: _id(string), approvers(array), createdDate(string), metadata(string), name(string), parentGroupId(string), resources(array), updatedDate(string)
  cost_categories:
    primary key: _id
    fields: _id(string), createdDate(string), name(string), updatedDate(string)
  project_managers:
    primary key: _id
    fields: _id(string), createdDate(string), email(string), firstName(string), isProjectManager(boolean), lastName(string), links(object), metadata(string), note(string), role(string), status(string), updatedDate(string)
  project_tags:
    primary key: _id
    fields: _id(string), category(string), value(string)
  resource_tags:
    primary key: _id
    fields: _id(string), category(string), value(string)
  time_entries:
    primary key: _id
    fields: _id(string), categoryName(string), categoryTemplateId(string), createdDate(string), creator(string), date(string), locked(boolean), metadata(string), minutes(integer), note(string), project(string), projectName(string), projectStatus(string), projectType(string), resource(string), status(string), updatedDate(string)
  unassigned_work:
    primary key: _id
    fields: _id(string), value(string)
  vacations:
    primary key: _id
    fields: _id(string), approvalInfo(object), creatorId(string), end(string), metadata(string), minutesPerDay(integer), percentAllocation(string), resource(string), resourceType(string), start(string), state(string), title(string), type(string)
  webhook_subscriptions:
    primary key: _id
    fields: _id(string), companyId(string), creationDate(string), event(string), target_url(string)

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite

REVERSE ETL ACTIONS
  create_billing_rate:
    endpoint: POST /billingRate
    required fields: label, currency, rate
    risk: creates a Hubplanner billing rate
  update_billing_rate:
    endpoint: PUT /billingRate/{{ record.id }}
    required fields: id, label, currency, rate
    risk: updates a Hubplanner billing rate
  delete_billing_rate:
    endpoint: DELETE /billingRate/{{ record.id }}
    required fields: id
    risk: deletes a Hubplanner billing rate
  create_booking_category:
    endpoint: POST /categories
    required fields: name
    risk: creates a Hubplanner booking category
  update_booking_category:
    endpoint: PUT /categories/{{ record.id }}
    required fields: id, name
    risk: updates a Hubplanner booking category
  create_booking:
    endpoint: POST /booking
    required fields: resource, project, start, end
    risk: creates a Hubplanner booking
  update_booking:
    endpoint: PUT /booking/{{ record.id }}
    required fields: id, resource, project, start, end
    risk: updates a Hubplanner booking
  patch_booking:
    endpoint: PATCH /booking/{{ record.id }}
    required fields: id
    risk: patches selected Hubplanner booking fields
  delete_booking:
    endpoint: DELETE /booking/{{ record.id }}
    required fields: id
    risk: deletes a Hubplanner booking
  delete_bookings:
    endpoint: DELETE /booking?ids={{ record.ids }}
    required fields: ids
    risk: deletes multiple Hubplanner bookings by ids
  create_client:
    endpoint: POST /client
    required fields: name
    risk: creates a Hubplanner client
  update_client:
    endpoint: PUT /client/{{ record.id }}
    required fields: id, name
    risk: updates a Hubplanner client
  delete_client:
    endpoint: DELETE /client/{{ record.id }}
    required fields: id
    risk: deletes a Hubplanner client
  create_resource_custom_field_template:
    endpoint: POST /resource/customField/template
    required fields: label, type
    risk: creates a Hubplanner resource custom field template
  update_resource_custom_field_template:
    endpoint: PUT /resource/customField/template/{{ record.id }}
    required fields: id, label, type
    risk: updates a Hubplanner resource custom field template
  delete_resource_custom_field_template:
    endpoint: DELETE /resource/customField/template/{{ record.id }}
    required fields: id
    risk: deletes a Hubplanner resource custom field template
  create_project_custom_field_template:
    endpoint: POST /project/customField/template
    required fields: label, type
    risk: creates a Hubplanner project custom field template
  update_project_custom_field_template:
    endpoint: PUT /project/customField/template/{{ record.id }}
    required fields: id, label, type
    risk: updates a Hubplanner project custom field template
  delete_project_custom_field_template:
    endpoint: DELETE /project/customField/template/{{ record.id }}
    required fields: id
    risk: deletes a Hubplanner project custom field template
  create_event:
    endpoint: POST /event
    required fields: name
    risk: creates a Hubplanner event
  update_event:
    endpoint: PUT /event/{{ record.id }}
    required fields: id, name
    risk: updates a Hubplanner event
  delete_event:
    endpoint: DELETE /event/{{ record.id }}
    required fields: id
    risk: deletes a Hubplanner event
  create_project_group:
    endpoint: POST /projectgroup
    required fields: name
    risk: creates a Hubplanner project group
  update_project_group:
    endpoint: PUT /projectgroup/{{ record.id }}
    required fields: id, name
    risk: updates a Hubplanner project group
  delete_project_group:
    endpoint: DELETE /projectgroup/{{ record.id }}
    required fields: id
    risk: deletes a Hubplanner project group
  create_holiday:
    endpoint: POST /holiday
    required fields: name, date
    risk: creates a Hubplanner holiday
  update_holiday:
    endpoint: PUT /holiday/{{ record.id }}
    required fields: id, name, date
    risk: updates a Hubplanner holiday
  delete_holiday:
    endpoint: DELETE /holiday/{{ record.id }}
    required fields: id
    risk: deletes a Hubplanner holiday
  create_milestone:
    endpoint: POST /milestone
    required fields: name, date, project
    risk: creates a Hubplanner milestone
  update_milestone:
    endpoint: PUT /milestone/{{ record.id }}
    required fields: id, name, date, project
    risk: updates a Hubplanner milestone
  delete_milestone:
    endpoint: DELETE /milestone/{{ record.id }}
    required fields: id
    risk: deletes a Hubplanner milestone
  create_cost_category:
    endpoint: POST /costCategories
    required fields: name
    risk: creates a Hubplanner project cost category
  update_cost_category:
    endpoint: PUT /costCategories/{{ record.id }}
    required fields: id, name
    risk: updates a Hubplanner project cost category
  delete_cost_category:
    endpoint: DELETE /costCategories/{{ record.id }}
    required fields: id
    risk: deletes a Hubplanner project cost category
  create_project_tag:
    endpoint: POST /project-tag
    required fields: value
    risk: creates a Hubplanner project tag
  update_project_tag:
    endpoint: PUT /project-tag/{{ record.id }}
    required fields: id, value
    risk: updates a Hubplanner project tag
  delete_project_tag:
    endpoint: DELETE /project-tag/{{ record.id }}
    required fields: id
    risk: deletes a Hubplanner project tag
  create_resource_tag:
    endpoint: POST /resource-tag
    required fields: value
    risk: creates a Hubplanner resource tag
  update_resource_tag:
    endpoint: PUT /resource-tag/{{ record.id }}
    required fields: id, value
    risk: updates a Hubplanner resource tag
  delete_resource_tag:
    endpoint: DELETE /resource-tag/{{ record.id }}
    required fields: id
    risk: deletes a Hubplanner resource tag
  create_project:
    endpoint: POST /project
    required fields: name
    risk: creates a Hubplanner project
  patch_project:
    endpoint: PATCH /project
    required fields: _id
    risk: patches a Hubplanner project by _id in the typed body
  delete_project:
    endpoint: DELETE /project/{{ record.id }}
    required fields: id
    risk: deletes a Hubplanner project
  delete_projects:
    endpoint: DELETE /project
    required fields: ids
    risk: deletes multiple Hubplanner projects by ids
  create_resource:
    endpoint: POST /resource
    required fields: firstName, lastName
    risk: creates a Hubplanner resource
  patch_resource:
    endpoint: PATCH /resource
    required fields: _id
    risk: patches a Hubplanner resource by _id in the typed body
  delete_resource:
    endpoint: DELETE /resource/{{ record.id }}
    required fields: id
    risk: deletes a Hubplanner resource
  delete_resources:
    endpoint: DELETE /resource
    required fields: ids
    risk: deletes multiple Hubplanner resources by ids
  create_time_entry:
    endpoint: POST /timeentry
    required fields: resource, project, date, minutes
    risk: creates a Hubplanner time entry
  update_time_entry:
    endpoint: PUT /timeentry/{{ record.id }}
    required fields: id, resource, project, date, minutes
    risk: updates a Hubplanner time entry
  delete_time_entry:
    endpoint: DELETE /timeentry/{{ record.id }}
    required fields: id
    risk: deletes a Hubplanner time entry
  delete_time_entries:
    endpoint: DELETE /timeentry
    required fields: ids
    risk: deletes multiple Hubplanner time entries by ids
  create_unassigned_work:
    endpoint: POST /unassigned-work
    required fields: value
    risk: creates Hubplanner unassigned work
  update_unassigned_work:
    endpoint: PUT /unassigned-work/{{ record.id }}
    required fields: id, value
    risk: updates Hubplanner unassigned work
  delete_unassigned_work:
    endpoint: DELETE /unassigned-work/{{ record.id }}
    required fields: id
    risk: deletes Hubplanner unassigned work
  create_vacation:
    endpoint: POST /vacation
    required fields: resource, start, end
    risk: creates a Hubplanner vacation
  update_vacation:
    endpoint: PUT /vacation/{{ record.id }}
    required fields: id, resource, start, end
    risk: updates a Hubplanner vacation
  patch_vacation:
    endpoint: PATCH /vacation/{{ record.id }}
    required fields: id
    risk: patches selected Hubplanner vacation fields
  delete_vacation:
    endpoint: DELETE /vacation/{{ record.id }}
    required fields: id
    risk: deletes a Hubplanner vacation
  create_webhook_subscription:
    endpoint: POST /subscription
    required fields: target_url, event
    risk: creates a Hubplanner webhook subscription that can deliver future event payloads to the approved HTTPS target URL
  delete_webhook_subscription:
    endpoint: DELETE /subscription/{{ record.id }}
    required fields: id
    risk: deletes a Hubplanner webhook subscription

SECURITY
  read risk: external Hubplanner API reads of scheduling, project, resource, client, time, vacation, billing, and webhook-subscription data
  write risk: typed reverse-ETL actions create, update, patch, and delete Hubplanner records using fixed documented REST paths only
  approval: reverse ETL writes require plan, preview, explicit approval, and destructive confirmation for deletes
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

COMMAND SURFACE
  Inspect, sync, directly read, and safely plan typed Hubplanner operations.
  Usage: pm hubplanner <command> [flags]
  Source CLI: Hubplanner API (Provider-owned Sections/*.md endpoint reference)
  Global flags:
    --credential (string): Credential name to use for the Hubplanner request.
    --connection (string): Alias for --credential.
    --config (string_array): Connector config override as key=value.
    --json (boolean): Emit machine-readable JSON output.
    --limit (integer): Maximum ETL records to emit.
    --max-bytes (integer): Maximum direct-read response bytes, capped by each typed operation.
    --preview (boolean): Preview a reverse-ETL write command without making a network mutation.
    --approval-token-stdin (boolean): Read the approval token as one bounded line from standard input.
    --confirm (string): Typed confirmation challenge for destructive reverse-ETL writes.
  Billing Rates
  Booking Categories
  Bookings
  Clients
  Cost Categories
  Events
  Holidays
  Milestones
  Project Custom Field Templates
  Project Custom Fields
  Project Groups
  Project Managers
  Project Tags
  Projects
  Resource Custom Field Templates
  Resource Custom Fields
  Resource Groups
  Resource Tags
  Resources
  Time Entries
  Unassigned Work
  Vacations
  Webhook Events
  Webhook Subscriptions
  Other Commands
    resources list - List Hubplanner resources as ETL records. [intent=etl availability=implemented stream=resources]
    projects list - List Hubplanner projects as ETL records. [intent=etl availability=implemented stream=projects]
    clients list - List Hubplanner clients as ETL records. [intent=etl availability=implemented stream=clients]
    events list - List Hubplanner events as ETL records. [intent=etl availability=implemented stream=events]
    holidays list - List Hubplanner holidays as ETL records. [intent=etl availability=implemented stream=holidays]
    bookings list - List Hubplanner bookings as ETL records. [intent=etl availability=implemented stream=bookings]
    billing-rates list - List Hubplanner billing rates as ETL records. [intent=etl availability=implemented stream=billing_rates]
    booking-categories list - List Hubplanner booking categories as ETL records. [intent=etl availability=implemented stream=booking_categories]
    resource-custom-field-templates list - List Hubplanner resource custom field templates as ETL records. [intent=etl availability=implemented stream=resource_custom_field_templates]
    project-custom-field-templates list - List Hubplanner project custom field templates as ETL records. [intent=etl availability=implemented stream=project_custom_field_templates]
    project-groups list - List Hubplanner project groups as ETL records. [intent=etl availability=implemented stream=project_groups]
    resource-groups list - List Hubplanner resource groups as ETL records. [intent=etl availability=implemented stream=resource_groups]
    cost-categories list - List Hubplanner project cost categories as ETL records. [intent=etl availability=implemented stream=cost_categories]
    project-managers list - List Hubplanner project managers as ETL records. [intent=etl availability=implemented stream=project_managers]
    project-tags list - List Hubplanner project tags as ETL records. [intent=etl availability=implemented stream=project_tags]
    resource-tags list - List Hubplanner resource tags as ETL records. [intent=etl availability=implemented stream=resource_tags]
    time-entries list - List Hubplanner time entries as ETL records. [intent=etl availability=implemented stream=time_entries]
    unassigned-work list - List Hubplanner unassigned work as ETL records. [intent=etl availability=implemented stream=unassigned_work]
    vacations list - List Hubplanner vacations as ETL records. [intent=etl availability=implemented stream=vacations]
    webhook-subscriptions list - List Hubplanner webhook subscriptions as ETL records. [intent=etl availability=implemented stream=webhook_subscriptions]
    billing-rates get - Get a Hubplanner billing rate by id. [intent=direct_read availability=implemented operation=hubplanner.billing_rates_get]; flags: --id, --page, --page-cursor
    booking-categories get - Get a Hubplanner booking category by id. [intent=direct_read availability=implemented operation=hubplanner.booking_categories_get]; flags: --id, --page, --page-cursor
    bookings get - Get a Hubplanner booking by id. [intent=direct_read availability=implemented operation=hubplanner.bookings_get]; flags: --id, --page, --page-cursor
    clients get - Get a Hubplanner client by id. [intent=direct_read availability=implemented operation=hubplanner.clients_get]; flags: --id, --page, --page-cursor
    events get - Get a Hubplanner event by id. [intent=direct_read availability=implemented operation=hubplanner.events_get]; flags: --id, --page, --page-cursor
    project-groups get - Get a Hubplanner project group by id. [intent=direct_read availability=implemented operation=hubplanner.project_groups_get]; flags: --id, --page, --page-cursor
    holidays get - Get a Hubplanner holiday by id. [intent=direct_read availability=implemented operation=hubplanner.holidays_get]; flags: --id, --page, --page-cursor
    milestones get - Get a Hubplanner milestone by id. [intent=direct_read availability=implemented operation=hubplanner.milestones_get]; flags: --id, --page, --page-cursor
    cost-categories get - Get a Hubplanner project cost category by id. [intent=direct_read availability=implemented operation=hubplanner.cost_categories_get]; flags: --id, --page, --page-cursor
    projects get - Get a Hubplanner project by id. [intent=direct_read availability=implemented operation=hubplanner.projects_get]; flags: --id, --page, --page-cursor
    resources get - Get a Hubplanner resource by id. [intent=direct_read availability=implemented operation=hubplanner.resources_get]; flags: --id, --page, --page-cursor
    time-entries get - Get a Hubplanner time entry by id. [intent=direct_read availability=implemented operation=hubplanner.time_entries_get]; flags: --id, --page, --page-cursor
    unassigned-work get - Get a Hubplanner unassigned-work record by id. [intent=direct_read availability=implemented operation=hubplanner.unassigned_work_get]; flags: --id, --page, --page-cursor
    vacations get - Get a Hubplanner vacation by id. [intent=direct_read availability=implemented operation=hubplanner.vacations_get]; flags: --id, --page, --page-cursor
    resource-custom-fields search - Search Hubplanner resource custom field templates. [intent=direct_read availability=implemented operation=hubplanner.resource_custom_field_template_search]; flags: --type, --label, --required, --page, --page-cursor
    project-custom-fields search - Search Hubplanner project custom field templates. [intent=direct_read availability=implemented operation=hubplanner.project_custom_field_template_search]; flags: --type, --label, --required, --page, --page-cursor
    billing-rates create - Plan or preview Hubplanner write action `create_billing_rate`. [intent=reverse_etl availability=implemented write=create_billing_rate]; approval: reverse ETL plan -> preview -> approval -> execute; risk: creates a Hubplanner billing rate; flags: --label, --rate, --currency
    billing-rates update - Plan or preview Hubplanner write action `update_billing_rate`. [intent=reverse_etl availability=implemented write=update_billing_rate]; approval: reverse ETL plan -> preview -> approval -> execute; risk: updates a Hubplanner billing rate; flags: --id, --label, --rate, --currency
    billing-rates delete - Plan or preview Hubplanner write action `delete_billing_rate`. [intent=reverse_etl availability=implemented write=delete_billing_rate]; approval: reverse ETL plan -> preview -> approval -> execute with destructive confirmation; risk: deletes a Hubplanner billing rate; flags: --id
    booking-categories create - Plan or preview Hubplanner write action `create_booking_category`. [intent=reverse_etl availability=implemented write=create_booking_category]; approval: reverse ETL plan -> preview -> approval -> execute; risk: creates a Hubplanner booking category; flags: --name, --type, --gridColor, --group
    booking-categories update - Plan or preview Hubplanner write action `update_booking_category`. [intent=reverse_etl availability=implemented write=update_booking_category]; approval: reverse ETL plan -> preview -> approval -> execute; risk: updates a Hubplanner booking category; flags: --id, --name, --type, --gridColor, --group
    bookings create - Plan or preview Hubplanner write action `create_booking`. [intent=reverse_etl availability=implemented write=create_booking]; approval: reverse ETL plan -> preview -> approval -> execute; risk: creates a Hubplanner booking; flags: --id, --resource, --project, --start, --end, --note, --state, --category
    bookings update - Plan or preview Hubplanner write action `update_booking`. [intent=reverse_etl availability=implemented write=update_booking]; approval: reverse ETL plan -> preview -> approval -> execute; risk: updates a Hubplanner booking; flags: --id, --resource, --project, --start, --end, --note, --state, --category
    bookings patch - Plan or preview Hubplanner write action `patch_booking`. [intent=reverse_etl availability=implemented write=patch_booking]; approval: reverse ETL plan -> preview -> approval -> execute; risk: patches selected Hubplanner booking fields; flags: --id, --resource, --project, --start, --end, --note, --state, --category
    bookings delete - Plan or preview Hubplanner write action `delete_booking`. [intent=reverse_etl availability=implemented write=delete_booking]; approval: reverse ETL plan -> preview -> approval -> execute with destructive confirmation; risk: deletes a Hubplanner booking; flags: --id
    bookings delete-many - Plan or preview Hubplanner write action `delete_bookings` (bulk). [intent=reverse_etl availability=implemented write=delete_bookings]; approval: reverse ETL plan -> preview -> approval -> execute with destructive confirmation; risk: deletes multiple Hubplanner bookings by ids; flags: --ids
    clients create - Plan or preview Hubplanner write action `create_client`. [intent=reverse_etl availability=implemented write=create_client]; approval: reverse ETL plan -> preview -> approval -> execute; risk: creates a Hubplanner client; flags: --id, --name, --email, --phone, --note
    clients update - Plan or preview Hubplanner write action `update_client`. [intent=reverse_etl availability=implemented write=update_client]; approval: reverse ETL plan -> preview -> approval -> execute; risk: updates a Hubplanner client; flags: --id, --name, --email, --phone, --note
    clients delete - Plan or preview Hubplanner write action `delete_client`. [intent=reverse_etl availability=implemented write=delete_client]; approval: reverse ETL plan -> preview -> approval -> execute with destructive confirmation; risk: deletes a Hubplanner client; flags: --id
    resource-custom-field-templates create - Plan or preview Hubplanner write action `create_resource_custom_field_template`. [intent=reverse_etl availability=implemented write=create_resource_custom_field_template]; approval: reverse ETL plan -> preview -> approval -> execute; risk: creates a Hubplanner resource custom field template; flags: --id, --label, --type, --status, --instructions, --choices
    resource-custom-field-templates update - Plan or preview Hubplanner write action `update_resource_custom_field_template`. [intent=reverse_etl availability=implemented write=update_resource_custom_field_template]; approval: reverse ETL plan -> preview -> approval -> execute; risk: updates a Hubplanner resource custom field template; flags: --id, --label, --type, --status, --instructions, --choices
    resource-custom-field-templates delete - Plan or preview Hubplanner write action `delete_resource_custom_field_template`. [intent=reverse_etl availability=implemented write=delete_resource_custom_field_template]; approval: reverse ETL plan -> preview -> approval -> execute with destructive confirmation; risk: deletes a Hubplanner resource custom field template; flags: --id
    project-custom-field-templates create - Plan or preview Hubplanner write action `create_project_custom_field_template`. [intent=reverse_etl availability=implemented write=create_project_custom_field_template]; approval: reverse ETL plan -> preview -> approval -> execute; risk: creates a Hubplanner project custom field template; flags: --id, --label, --type, --status, --instructions, --choices
    project-custom-field-templates update - Plan or preview Hubplanner write action `update_project_custom_field_template`. [intent=reverse_etl availability=implemented write=update_project_custom_field_template]; approval: reverse ETL plan -> preview -> approval -> execute; risk: updates a Hubplanner project custom field template; flags: --id, --label, --type, --status, --instructions, --choices
    project-custom-field-templates delete - Plan or preview Hubplanner write action `delete_project_custom_field_template`. [intent=reverse_etl availability=implemented write=delete_project_custom_field_template]; approval: reverse ETL plan -> preview -> approval -> execute with destructive confirmation; risk: deletes a Hubplanner project custom field template; flags: --id
    events create - Plan or preview Hubplanner write action `create_event`. [intent=reverse_etl availability=implemented write=create_event]; approval: reverse ETL plan -> preview -> approval -> execute; risk: creates a Hubplanner event; flags: --name, --eventCode, --backgroundColor, --metadata
    events update - Plan or preview Hubplanner write action `update_event`. [intent=reverse_etl availability=implemented write=update_event]; approval: reverse ETL plan -> preview -> approval -> execute; risk: updates a Hubplanner event; flags: --id, --name, --eventCode, --backgroundColor, --metadata
    events delete - Plan or preview Hubplanner write action `delete_event`. [intent=reverse_etl availability=implemented write=delete_event]; approval: reverse ETL plan -> preview -> approval -> execute with destructive confirmation; risk: deletes a Hubplanner event; flags: --id
    project-groups create - Plan or preview Hubplanner write action `create_project_group`. [intent=reverse_etl availability=implemented write=create_project_group]; approval: reverse ETL plan -> preview -> approval -> execute; risk: creates a Hubplanner project group; flags: --name, --parentGroupId, --projects
    project-groups update - Plan or preview Hubplanner write action `update_project_group`. [intent=reverse_etl availability=implemented write=update_project_group]; approval: reverse ETL plan -> preview -> approval -> execute; risk: updates a Hubplanner project group; flags: --id, --name, --parentGroupId, --projects
    project-groups delete - Plan or preview Hubplanner write action `delete_project_group`. [intent=reverse_etl availability=implemented write=delete_project_group]; approval: reverse ETL plan -> preview -> approval -> execute with destructive confirmation; risk: deletes a Hubplanner project group; flags: --id
    holidays create - Plan or preview Hubplanner write action `create_holiday`. [intent=reverse_etl availability=implemented write=create_holiday]; approval: reverse ETL plan -> preview -> approval -> execute; risk: creates a Hubplanner holiday; flags: --id, --name, --date, --start, --end, --holidayGroup
    holidays update - Plan or preview Hubplanner write action `update_holiday`. [intent=reverse_etl availability=implemented write=update_holiday]; approval: reverse ETL plan -> preview -> approval -> execute; risk: updates a Hubplanner holiday; flags: --id, --name, --date, --start, --end, --holidayGroup
    holidays delete - Plan or preview Hubplanner write action `delete_holiday`. [intent=reverse_etl availability=implemented write=delete_holiday]; approval: reverse ETL plan -> preview -> approval -> execute with destructive confirmation; risk: deletes a Hubplanner holiday; flags: --id
    milestones create - Plan or preview Hubplanner write action `create_milestone`. [intent=reverse_etl availability=implemented write=create_milestone]; approval: reverse ETL plan -> preview -> approval -> execute; risk: creates a Hubplanner milestone; flags: --name, --date, --project
    milestones update - Plan or preview Hubplanner write action `update_milestone`. [intent=reverse_etl availability=implemented write=update_milestone]; approval: reverse ETL plan -> preview -> approval -> execute; risk: updates a Hubplanner milestone; flags: --id, --name, --date, --project
    milestones delete - Plan or preview Hubplanner write action `delete_milestone`. [intent=reverse_etl availability=implemented write=delete_milestone]; approval: reverse ETL plan -> preview -> approval -> execute with destructive confirmation; risk: deletes a Hubplanner milestone; flags: --id
    cost-categories create - Plan or preview Hubplanner write action `create_cost_category`. [intent=reverse_etl availability=implemented write=create_cost_category]; approval: reverse ETL plan -> preview -> approval -> execute; risk: creates a Hubplanner project cost category; flags: --name
    cost-categories update - Plan or preview Hubplanner write action `update_cost_category`. [intent=reverse_etl availability=implemented write=update_cost_category]; approval: reverse ETL plan -> preview -> approval -> execute; risk: updates a Hubplanner project cost category; flags: --id, --name
    cost-categories delete - Plan or preview Hubplanner write action `delete_cost_category`. [intent=reverse_etl availability=implemented write=delete_cost_category]; approval: reverse ETL plan -> preview -> approval -> execute with destructive confirmation; risk: deletes a Hubplanner project cost category; flags: --id
    project-tags create - Plan or preview Hubplanner write action `create_project_tag`. [intent=reverse_etl availability=implemented write=create_project_tag]; approval: reverse ETL plan -> preview -> approval -> execute; risk: creates a Hubplanner project tag; flags: --value
    project-tags update - Plan or preview Hubplanner write action `update_project_tag`. [intent=reverse_etl availability=implemented write=update_project_tag]; approval: reverse ETL plan -> preview -> approval -> execute; risk: updates a Hubplanner project tag; flags: --id, --value
    project-tags delete - Plan or preview Hubplanner write action `delete_project_tag`. [intent=reverse_etl availability=implemented write=delete_project_tag]; approval: reverse ETL plan -> preview -> approval -> execute with destructive confirmation; risk: deletes a Hubplanner project tag; flags: --id
    resource-tags create - Plan or preview Hubplanner write action `create_resource_tag`. [intent=reverse_etl availability=implemented write=create_resource_tag]; approval: reverse ETL plan -> preview -> approval -> execute; risk: creates a Hubplanner resource tag; flags: --value
    resource-tags update - Plan or preview Hubplanner write action `update_resource_tag`. [intent=reverse_etl availability=implemented write=update_resource_tag]; approval: reverse ETL plan -> preview -> approval -> execute; risk: updates a Hubplanner resource tag; flags: --id, --value
    resource-tags delete - Plan or preview Hubplanner write action `delete_resource_tag`. [intent=reverse_etl availability=implemented write=delete_resource_tag]; approval: reverse ETL plan -> preview -> approval -> execute with destructive confirmation; risk: deletes a Hubplanner resource tag; flags: --id
    projects create - Plan or preview Hubplanner write action `create_project`. [intent=reverse_etl availability=implemented write=create_project]; approval: reverse ETL plan -> preview -> approval -> execute; risk: creates a Hubplanner project; flags: --id, --name, --status, --projectCode
    projects patch - Plan or preview Hubplanner write action `patch_project`. [intent=reverse_etl availability=implemented write=patch_project]; approval: reverse ETL plan -> preview -> approval -> execute; risk: patches a Hubplanner project by _id in the typed body; flags: --id, --name, --status, --projectCode
    projects delete - Plan or preview Hubplanner write action `delete_project`. [intent=reverse_etl availability=implemented write=delete_project]; approval: reverse ETL plan -> preview -> approval -> execute with destructive confirmation; risk: deletes a Hubplanner project; flags: --id
    projects delete-many - Plan or preview Hubplanner write action `delete_projects` (bulk). [intent=reverse_etl availability=implemented write=delete_projects]; approval: reverse ETL plan -> preview -> approval -> execute with destructive confirmation; risk: deletes multiple Hubplanner projects by ids; flags: --ids
    resources create - Plan or preview Hubplanner write action `create_resource`. [intent=reverse_etl availability=implemented write=create_resource]; approval: reverse ETL plan -> preview -> approval -> execute; risk: creates a Hubplanner resource; flags: --id, --firstName, --lastName, --email, --status
    resources patch - Plan or preview Hubplanner write action `patch_resource`. [intent=reverse_etl availability=implemented write=patch_resource]; approval: reverse ETL plan -> preview -> approval -> execute; risk: patches a Hubplanner resource by _id in the typed body; flags: --id, --firstName, --lastName, --email, --status
    resources delete - Plan or preview Hubplanner write action `delete_resource`. [intent=reverse_etl availability=implemented write=delete_resource]; approval: reverse ETL plan -> preview -> approval -> execute with destructive confirmation; risk: deletes a Hubplanner resource; flags: --id
    resources delete-many - Plan or preview Hubplanner write action `delete_resources` (bulk). [intent=reverse_etl availability=implemented write=delete_resources]; approval: reverse ETL plan -> preview -> approval -> execute with destructive confirmation; risk: deletes multiple Hubplanner resources by ids; flags: --ids
    time-entries create - Plan or preview Hubplanner write action `create_time_entry`. [intent=reverse_etl availability=implemented write=create_time_entry]; approval: reverse ETL plan -> preview -> approval -> execute; risk: creates a Hubplanner time entry; flags: --resource, --project, --date, --minutes, --note, --categoryTemplateId
    time-entries update - Plan or preview Hubplanner write action `update_time_entry`. [intent=reverse_etl availability=implemented write=update_time_entry]; approval: reverse ETL plan -> preview -> approval -> execute; risk: updates a Hubplanner time entry; flags: --id, --resource, --project, --date, --minutes, --note, --categoryTemplateId
    time-entries delete - Plan or preview Hubplanner write action `delete_time_entry`. [intent=reverse_etl availability=implemented write=delete_time_entry]; approval: reverse ETL plan -> preview -> approval -> execute with destructive confirmation; risk: deletes a Hubplanner time entry; flags: --id
    time-entries delete-many - Plan or preview Hubplanner write action `delete_time_entries` (bulk). [intent=reverse_etl availability=implemented write=delete_time_entries]; approval: reverse ETL plan -> preview -> approval -> execute with destructive confirmation; risk: deletes multiple Hubplanner time entries by ids; flags: --ids
    unassigned-work create - Plan or preview Hubplanner write action `create_unassigned_work`. [intent=reverse_etl availability=implemented write=create_unassigned_work]; approval: reverse ETL plan -> preview -> approval -> execute; risk: creates Hubplanner unassigned work; flags: --value
    unassigned-work update - Plan or preview Hubplanner write action `update_unassigned_work`. [intent=reverse_etl availability=implemented write=update_unassigned_work]; approval: reverse ETL plan -> preview -> approval -> execute; risk: updates Hubplanner unassigned work; flags: --id, --value
    unassigned-work delete - Plan or preview Hubplanner write action `delete_unassigned_work`. [intent=reverse_etl availability=implemented write=delete_unassigned_work]; approval: reverse ETL plan -> preview -> approval -> execute with destructive confirmation; risk: deletes Hubplanner unassigned work; flags: --id
    vacations create - Plan or preview Hubplanner write action `create_vacation`. [intent=reverse_etl availability=implemented write=create_vacation]; approval: reverse ETL plan -> preview -> approval -> execute; risk: creates a Hubplanner vacation; flags: --id, --resource, --start, --end, --note, --status
    vacations update - Plan or preview Hubplanner write action `update_vacation`. [intent=reverse_etl availability=implemented write=update_vacation]; approval: reverse ETL plan -> preview -> approval -> execute; risk: updates a Hubplanner vacation; flags: --id, --resource, --start, --end, --note, --status
    vacations patch - Plan or preview Hubplanner write action `patch_vacation`. [intent=reverse_etl availability=implemented write=patch_vacation]; approval: reverse ETL plan -> preview -> approval -> execute; risk: patches selected Hubplanner vacation fields; flags: --id, --resource, --start, --end, --note, --status
    vacations delete - Plan or preview Hubplanner write action `delete_vacation`. [intent=reverse_etl availability=implemented write=delete_vacation]; approval: reverse ETL plan -> preview -> approval -> execute with destructive confirmation; risk: deletes a Hubplanner vacation; flags: --id
    webhook-subscriptions create - Plan or preview Hubplanner write action `create_webhook_subscription`. [intent=reverse_etl availability=implemented write=create_webhook_subscription]; approval: reverse ETL plan -> preview -> approval -> execute with destructive confirmation for future webhook data egress; risk: creates a Hubplanner webhook subscription that can deliver future event payloads to the approved HTTPS target URL; flags: --id, --target-url, --event
    webhook-subscriptions delete - Plan or preview Hubplanner write action `delete_webhook_subscription`. [intent=reverse_etl availability=implemented write=delete_webhook_subscription]; approval: reverse ETL plan -> preview -> approval -> execute with destructive confirmation; risk: deletes a Hubplanner webhook subscription; flags: --id
    webhook-events project-update - Webhook event delivery surface `project.update` is documented but blocked. [intent=docs_only availability=planned]; notes: Hubplanner sends this provider callback to an external service; the current connector contract cannot receive it as CDC without the shared CDC/changefeed runtime.
    webhook-events resource-update - Webhook event delivery surface `resource.update` is documented but blocked. [intent=docs_only availability=planned]; notes: Hubplanner sends this provider callback to an external service; the current connector contract cannot receive it as CDC without the shared CDC/changefeed runtime.
    webhook-events booking-create - Webhook event delivery surface `booking.create` is documented but blocked. [intent=docs_only availability=planned]; notes: Hubplanner sends this provider callback to an external service; the current connector contract cannot receive it as CDC without the shared CDC/changefeed runtime.
    webhook-events timeEntry-create - Webhook event delivery surface `timeEntry.create` is documented but blocked. [intent=docs_only availability=planned]; notes: Hubplanner sends this provider callback to an external service; the current connector contract cannot receive it as CDC without the shared CDC/changefeed runtime.
    webhook-events timeEntry-update - Webhook event delivery surface `timeEntry.update` is documented but blocked. [intent=docs_only availability=planned]; notes: Hubplanner sends this provider callback to an external service; the current connector contract cannot receive it as CDC without the shared CDC/changefeed runtime.
    webhook-events timeEntry-create-update - Webhook event delivery surface `timeEntry.create.update` is documented but blocked. [intent=docs_only availability=planned]; notes: Hubplanner sends this provider callback to an external service; the current connector contract cannot receive it as CDC without the shared CDC/changefeed runtime.
    webhook-events timeEntry-delete - Webhook event delivery surface `timeEntry.delete` is documented but blocked. [intent=docs_only availability=planned]; notes: Hubplanner sends this provider callback to an external service; the current connector contract cannot receive it as CDC without the shared CDC/changefeed runtime.
    webhook-events booking-update - Webhook event delivery surface `booking.update` is documented but blocked. [intent=docs_only availability=planned]; notes: Hubplanner sends this provider callback to an external service; the current connector contract cannot receive it as CDC without the shared CDC/changefeed runtime.
    webhook-events booking-delete - Webhook event delivery surface `booking.delete` is documented but blocked. [intent=docs_only availability=planned]; notes: Hubplanner sends this provider callback to an external service; the current connector contract cannot receive it as CDC without the shared CDC/changefeed runtime.
    webhook-events booking-delete-multiple - Webhook event delivery surface `booking.delete.multiple` is documented but blocked. [intent=docs_only availability=planned]; notes: Hubplanner sends this provider callback to an external service; the current connector contract cannot receive it as CDC without the shared CDC/changefeed runtime.
  Help topics:
    hubplanner safety - Hubplanner writes require reverse-ETL plan, preview, approval, and destructive confirmation for deletes.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect hubplanner

  # Inspect as structured JSON
  pm connectors inspect hubplanner --json

AGENT WORKFLOW
  - Run pm connectors inspect hubplanner before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.
  - For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```

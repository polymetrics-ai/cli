# pm connectors inspect surveymonkey

```text
NAME
  pm connectors inspect surveymonkey - SurveyMonkey connector manual

SYNOPSIS
  pm connectors inspect surveymonkey
  pm connectors inspect surveymonkey --json
  pm credentials add <name> --connector surveymonkey [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads and writes SurveyMonkey REST v3 and SCIM v2 resources through the documented API surface.

ICON
  id: surveymonkey
  asset: icons/surveymonkey.svg
  source: upstream_registry
  review_status: upstream_seeded
  review_url: https://developer.surveymonkey.com/api/v3/

CAPABILITIES
  check=true catalog=true read=true write=true query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  action_type
  base_url
  benchmark_bundle_id
  collector_id
  contact_field_id
  contact_id
  contact_list_id
  error_id
  group_id
  language_code
  member_id
  message_id
  organization_id
  page_id
  question_id
  recipient_id
  response_id
  scim_resource_type_id
  scim_schema_id
  scim_user_id
  share_id
  survey_id
  user_id
  webhook_id
  workgroup_id
  access_token (secret) (required)

ETL STREAMS
  surveys:
    primary key: id
    fields: date_created(string), date_modified(string), id(string), title(string)
  survey_search:
    primary key: id
    fields: collector_id(string), created_at(string), data(object), date_created(string), date_modified(string), email(string), externalId(string), first_name(string), group_id(string), href(string), id(string), last_name(string), links(object), meta(object), name(string), page_id(string), schemas(array), status(string), subtype(string), survey_id(string), title(string), total(integer), type(string), updated_at(string), userName(string), username(string), workgroup_id(string)
  survey:
    primary key: id
    fields: collector_id(string), created_at(string), data(object), date_created(string), date_modified(string), email(string), externalId(string), first_name(string), group_id(string), href(string), id(string), last_name(string), links(object), meta(object), name(string), page_id(string), schemas(array), status(string), subtype(string), survey_id(string), title(string), total(integer), type(string), updated_at(string), userName(string), username(string), workgroup_id(string)
  survey_details:
    primary key: id
    fields: collector_id(string), created_at(string), data(object), date_created(string), date_modified(string), email(string), externalId(string), first_name(string), group_id(string), href(string), id(string), last_name(string), links(object), meta(object), name(string), page_id(string), schemas(array), status(string), subtype(string), survey_id(string), title(string), total(integer), type(string), updated_at(string), userName(string), username(string), workgroup_id(string)
  survey_share_list:
    primary key: id
    fields: collector_id(string), created_at(string), data(object), date_created(string), date_modified(string), email(string), externalId(string), first_name(string), group_id(string), href(string), id(string), last_name(string), links(object), meta(object), name(string), page_id(string), schemas(array), status(string), subtype(string), survey_id(string), title(string), total(integer), type(string), updated_at(string), userName(string), username(string), workgroup_id(string)
  survey_categories:
    primary key: id
    fields: collector_id(string), created_at(string), data(object), date_created(string), date_modified(string), email(string), externalId(string), first_name(string), group_id(string), href(string), id(string), last_name(string), links(object), meta(object), name(string), page_id(string), schemas(array), status(string), subtype(string), survey_id(string), title(string), total(integer), type(string), updated_at(string), userName(string), username(string), workgroup_id(string)
  survey_templates:
    primary key: id
    fields: collector_id(string), created_at(string), data(object), date_created(string), date_modified(string), email(string), externalId(string), first_name(string), group_id(string), href(string), id(string), last_name(string), links(object), meta(object), name(string), page_id(string), schemas(array), status(string), subtype(string), survey_id(string), title(string), total(integer), type(string), updated_at(string), userName(string), username(string), workgroup_id(string)
  team_survey_templates:
    primary key: id
    fields: collector_id(string), created_at(string), data(object), date_created(string), date_modified(string), email(string), externalId(string), first_name(string), group_id(string), href(string), id(string), last_name(string), links(object), meta(object), name(string), page_id(string), schemas(array), status(string), subtype(string), survey_id(string), title(string), total(integer), type(string), updated_at(string), userName(string), username(string), workgroup_id(string)
  survey_languages_catalog:
    primary key: id
    fields: collector_id(string), created_at(string), data(object), date_created(string), date_modified(string), email(string), externalId(string), first_name(string), group_id(string), href(string), id(string), last_name(string), links(object), meta(object), name(string), page_id(string), schemas(array), status(string), subtype(string), survey_id(string), title(string), total(integer), type(string), updated_at(string), userName(string), username(string), workgroup_id(string)
  collectors:
    primary key: id
    fields: date_created(string), date_modified(string), id(string), name(string)
  survey_pages:
    primary key: id
    fields: collector_id(string), created_at(string), data(object), date_created(string), date_modified(string), email(string), externalId(string), first_name(string), group_id(string), href(string), id(string), last_name(string), links(object), meta(object), name(string), page_id(string), schemas(array), status(string), subtype(string), survey_id(string), title(string), total(integer), type(string), updated_at(string), userName(string), username(string), workgroup_id(string)
  survey_page:
    primary key: id
    fields: collector_id(string), created_at(string), data(object), date_created(string), date_modified(string), email(string), externalId(string), first_name(string), group_id(string), href(string), id(string), last_name(string), links(object), meta(object), name(string), page_id(string), schemas(array), status(string), subtype(string), survey_id(string), title(string), total(integer), type(string), updated_at(string), userName(string), username(string), workgroup_id(string)
  survey_questions:
    primary key: id
    fields: collector_id(string), created_at(string), data(object), date_created(string), date_modified(string), email(string), externalId(string), first_name(string), group_id(string), href(string), id(string), last_name(string), links(object), meta(object), name(string), page_id(string), schemas(array), status(string), subtype(string), survey_id(string), title(string), total(integer), type(string), updated_at(string), userName(string), username(string), workgroup_id(string)
  survey_question:
    primary key: id
    fields: collector_id(string), created_at(string), data(object), date_created(string), date_modified(string), email(string), externalId(string), first_name(string), group_id(string), href(string), id(string), last_name(string), links(object), meta(object), name(string), page_id(string), schemas(array), status(string), subtype(string), survey_id(string), title(string), total(integer), type(string), updated_at(string), userName(string), username(string), workgroup_id(string)
  question_benchmark:
    primary key: id
    fields: collector_id(string), created_at(string), data(object), date_created(string), date_modified(string), email(string), externalId(string), first_name(string), group_id(string), href(string), id(string), last_name(string), links(object), meta(object), name(string), page_id(string), schemas(array), status(string), subtype(string), survey_id(string), title(string), total(integer), type(string), updated_at(string), userName(string), username(string), workgroup_id(string)
  question_rollups:
    primary key: id
    fields: collector_id(string), created_at(string), data(object), date_created(string), date_modified(string), email(string), externalId(string), first_name(string), group_id(string), href(string), id(string), last_name(string), links(object), meta(object), name(string), page_id(string), schemas(array), status(string), subtype(string), survey_id(string), title(string), total(integer), type(string), updated_at(string), userName(string), username(string), workgroup_id(string)
  question_trends:
    primary key: id
    fields: collector_id(string), created_at(string), data(object), date_created(string), date_modified(string), email(string), externalId(string), first_name(string), group_id(string), href(string), id(string), last_name(string), links(object), meta(object), name(string), page_id(string), schemas(array), status(string), subtype(string), survey_id(string), title(string), total(integer), type(string), updated_at(string), userName(string), username(string), workgroup_id(string)
  page_rollups:
    primary key: id
    fields: collector_id(string), created_at(string), data(object), date_created(string), date_modified(string), email(string), externalId(string), first_name(string), group_id(string), href(string), id(string), last_name(string), links(object), meta(object), name(string), page_id(string), schemas(array), status(string), subtype(string), survey_id(string), title(string), total(integer), type(string), updated_at(string), userName(string), username(string), workgroup_id(string)
  page_trends:
    primary key: id
    fields: collector_id(string), created_at(string), data(object), date_created(string), date_modified(string), email(string), externalId(string), first_name(string), group_id(string), href(string), id(string), last_name(string), links(object), meta(object), name(string), page_id(string), schemas(array), status(string), subtype(string), survey_id(string), title(string), total(integer), type(string), updated_at(string), userName(string), username(string), workgroup_id(string)
  survey_response_summaries:
    primary key: id
    fields: collector_id(string), created_at(string), data(object), date_created(string), date_modified(string), email(string), externalId(string), first_name(string), group_id(string), href(string), id(string), last_name(string), links(object), meta(object), name(string), page_id(string), schemas(array), status(string), subtype(string), survey_id(string), title(string), total(integer), type(string), updated_at(string), userName(string), username(string), workgroup_id(string)
  survey_response:
    primary key: id
    fields: collector_id(string), created_at(string), data(object), date_created(string), date_modified(string), email(string), externalId(string), first_name(string), group_id(string), href(string), id(string), last_name(string), links(object), meta(object), name(string), page_id(string), schemas(array), status(string), subtype(string), survey_id(string), title(string), total(integer), type(string), updated_at(string), userName(string), username(string), workgroup_id(string)
  survey_response_details:
    primary key: id
    fields: collector_id(string), created_at(string), data(object), date_created(string), date_modified(string), email(string), externalId(string), first_name(string), group_id(string), href(string), id(string), last_name(string), links(object), meta(object), name(string), page_id(string), schemas(array), status(string), subtype(string), survey_id(string), title(string), total(integer), type(string), updated_at(string), userName(string), username(string), workgroup_id(string)
  survey_responses:
    primary key: id
    fields: analyze_url(string), date_created(string), date_modified(string), id(string)
  survey_rollups:
    primary key: id
    fields: collector_id(string), created_at(string), data(object), date_created(string), date_modified(string), email(string), externalId(string), first_name(string), group_id(string), href(string), id(string), last_name(string), links(object), meta(object), name(string), page_id(string), schemas(array), status(string), subtype(string), survey_id(string), title(string), total(integer), type(string), updated_at(string), userName(string), username(string), workgroup_id(string)
  survey_trends:
    primary key: id
    fields: collector_id(string), created_at(string), data(object), date_created(string), date_modified(string), email(string), externalId(string), first_name(string), group_id(string), href(string), id(string), last_name(string), links(object), meta(object), name(string), page_id(string), schemas(array), status(string), subtype(string), survey_id(string), title(string), total(integer), type(string), updated_at(string), userName(string), username(string), workgroup_id(string)
  survey_languages:
    primary key: id
    fields: collector_id(string), created_at(string), data(object), date_created(string), date_modified(string), email(string), externalId(string), first_name(string), group_id(string), href(string), id(string), last_name(string), links(object), meta(object), name(string), page_id(string), schemas(array), status(string), subtype(string), survey_id(string), title(string), total(integer), type(string), updated_at(string), userName(string), username(string), workgroup_id(string)
  survey_language:
    primary key: id
    fields: collector_id(string), created_at(string), data(object), date_created(string), date_modified(string), email(string), externalId(string), first_name(string), group_id(string), href(string), id(string), last_name(string), links(object), meta(object), name(string), page_id(string), schemas(array), status(string), subtype(string), survey_id(string), title(string), total(integer), type(string), updated_at(string), userName(string), username(string), workgroup_id(string)
  question_bank_questions:
    primary key: id
    fields: collector_id(string), created_at(string), data(object), date_created(string), date_modified(string), email(string), externalId(string), first_name(string), group_id(string), href(string), id(string), last_name(string), links(object), meta(object), name(string), page_id(string), schemas(array), status(string), subtype(string), survey_id(string), title(string), total(integer), type(string), updated_at(string), userName(string), username(string), workgroup_id(string)
  survey_folders:
    primary key: id
    fields: collector_id(string), created_at(string), data(object), date_created(string), date_modified(string), email(string), externalId(string), first_name(string), group_id(string), href(string), id(string), last_name(string), links(object), meta(object), name(string), page_id(string), schemas(array), status(string), subtype(string), survey_id(string), title(string), total(integer), type(string), updated_at(string), userName(string), username(string), workgroup_id(string)
  users_me:
    primary key: id
    fields: collector_id(string), created_at(string), data(object), date_created(string), date_modified(string), email(string), externalId(string), first_name(string), group_id(string), href(string), id(string), last_name(string), links(object), meta(object), name(string), page_id(string), schemas(array), status(string), subtype(string), survey_id(string), title(string), total(integer), type(string), updated_at(string), userName(string), username(string), workgroup_id(string)
  user_workgroups:
    primary key: id
    fields: collector_id(string), created_at(string), data(object), date_created(string), date_modified(string), email(string), externalId(string), first_name(string), group_id(string), href(string), id(string), last_name(string), links(object), meta(object), name(string), page_id(string), schemas(array), status(string), subtype(string), survey_id(string), title(string), total(integer), type(string), updated_at(string), userName(string), username(string), workgroup_id(string)
  user_shared:
    primary key: id
    fields: collector_id(string), created_at(string), data(object), date_created(string), date_modified(string), email(string), externalId(string), first_name(string), group_id(string), href(string), id(string), last_name(string), links(object), meta(object), name(string), page_id(string), schemas(array), status(string), subtype(string), survey_id(string), title(string), total(integer), type(string), updated_at(string), userName(string), username(string), workgroup_id(string)
  groups:
    primary key: id
    fields: collector_id(string), created_at(string), data(object), date_created(string), date_modified(string), email(string), externalId(string), first_name(string), group_id(string), href(string), id(string), last_name(string), links(object), meta(object), name(string), page_id(string), schemas(array), status(string), subtype(string), survey_id(string), title(string), total(integer), type(string), updated_at(string), userName(string), username(string), workgroup_id(string)
  group:
    primary key: id
    fields: collector_id(string), created_at(string), data(object), date_created(string), date_modified(string), email(string), externalId(string), first_name(string), group_id(string), href(string), id(string), last_name(string), links(object), meta(object), name(string), page_id(string), schemas(array), status(string), subtype(string), survey_id(string), title(string), total(integer), type(string), updated_at(string), userName(string), username(string), workgroup_id(string)
  group_activities:
    primary key: id
    fields: collector_id(string), created_at(string), data(object), date_created(string), date_modified(string), email(string), externalId(string), first_name(string), group_id(string), href(string), id(string), last_name(string), links(object), meta(object), name(string), page_id(string), schemas(array), status(string), subtype(string), survey_id(string), title(string), total(integer), type(string), updated_at(string), userName(string), username(string), workgroup_id(string)
  group_activities_by_action:
    primary key: id
    fields: collector_id(string), created_at(string), data(object), date_created(string), date_modified(string), email(string), externalId(string), first_name(string), group_id(string), href(string), id(string), last_name(string), links(object), meta(object), name(string), page_id(string), schemas(array), status(string), subtype(string), survey_id(string), title(string), total(integer), type(string), updated_at(string), userName(string), username(string), workgroup_id(string)
  group_members:
    primary key: id
    fields: collector_id(string), created_at(string), data(object), date_created(string), date_modified(string), email(string), externalId(string), first_name(string), group_id(string), href(string), id(string), last_name(string), links(object), meta(object), name(string), page_id(string), schemas(array), status(string), subtype(string), survey_id(string), title(string), total(integer), type(string), updated_at(string), userName(string), username(string), workgroup_id(string)
  group_member:
    primary key: id
    fields: collector_id(string), created_at(string), data(object), date_created(string), date_modified(string), email(string), externalId(string), first_name(string), group_id(string), href(string), id(string), last_name(string), links(object), meta(object), name(string), page_id(string), schemas(array), status(string), subtype(string), survey_id(string), title(string), total(integer), type(string), updated_at(string), userName(string), username(string), workgroup_id(string)
  workgroups:
    primary key: id
    fields: collector_id(string), created_at(string), data(object), date_created(string), date_modified(string), email(string), externalId(string), first_name(string), group_id(string), href(string), id(string), last_name(string), links(object), meta(object), name(string), page_id(string), schemas(array), status(string), subtype(string), survey_id(string), title(string), total(integer), type(string), updated_at(string), userName(string), username(string), workgroup_id(string)
  workgroup:
    primary key: id
    fields: collector_id(string), created_at(string), data(object), date_created(string), date_modified(string), email(string), externalId(string), first_name(string), group_id(string), href(string), id(string), last_name(string), links(object), meta(object), name(string), page_id(string), schemas(array), status(string), subtype(string), survey_id(string), title(string), total(integer), type(string), updated_at(string), userName(string), username(string), workgroup_id(string)
  workgroup_shares:
    primary key: id
    fields: collector_id(string), created_at(string), data(object), date_created(string), date_modified(string), email(string), externalId(string), first_name(string), group_id(string), href(string), id(string), last_name(string), links(object), meta(object), name(string), page_id(string), schemas(array), status(string), subtype(string), survey_id(string), title(string), total(integer), type(string), updated_at(string), userName(string), username(string), workgroup_id(string)
  workgroup_share:
    primary key: id
    fields: collector_id(string), created_at(string), data(object), date_created(string), date_modified(string), email(string), externalId(string), first_name(string), group_id(string), href(string), id(string), last_name(string), links(object), meta(object), name(string), page_id(string), schemas(array), status(string), subtype(string), survey_id(string), title(string), total(integer), type(string), updated_at(string), userName(string), username(string), workgroup_id(string)
  workgroup_members:
    primary key: id
    fields: collector_id(string), created_at(string), data(object), date_created(string), date_modified(string), email(string), externalId(string), first_name(string), group_id(string), href(string), id(string), last_name(string), links(object), meta(object), name(string), page_id(string), schemas(array), status(string), subtype(string), survey_id(string), title(string), total(integer), type(string), updated_at(string), userName(string), username(string), workgroup_id(string)
  workgroup_member:
    primary key: id
    fields: collector_id(string), created_at(string), data(object), date_created(string), date_modified(string), email(string), externalId(string), first_name(string), group_id(string), href(string), id(string), last_name(string), links(object), meta(object), name(string), page_id(string), schemas(array), status(string), subtype(string), survey_id(string), title(string), total(integer), type(string), updated_at(string), userName(string), username(string), workgroup_id(string)
  contact_lists:
    primary key: id
    fields: collector_id(string), created_at(string), data(object), date_created(string), date_modified(string), email(string), externalId(string), first_name(string), group_id(string), href(string), id(string), last_name(string), links(object), meta(object), name(string), page_id(string), schemas(array), status(string), subtype(string), survey_id(string), title(string), total(integer), type(string), updated_at(string), userName(string), username(string), workgroup_id(string)
  contact_list:
    primary key: id
    fields: collector_id(string), created_at(string), data(object), date_created(string), date_modified(string), email(string), externalId(string), first_name(string), group_id(string), href(string), id(string), last_name(string), links(object), meta(object), name(string), page_id(string), schemas(array), status(string), subtype(string), survey_id(string), title(string), total(integer), type(string), updated_at(string), userName(string), username(string), workgroup_id(string)
  contact_list_contacts:
    primary key: id
    fields: collector_id(string), created_at(string), data(object), date_created(string), date_modified(string), email(string), externalId(string), first_name(string), group_id(string), href(string), id(string), last_name(string), links(object), meta(object), name(string), page_id(string), schemas(array), status(string), subtype(string), survey_id(string), title(string), total(integer), type(string), updated_at(string), userName(string), username(string), workgroup_id(string)
  contact_list_contacts_bulk:
    primary key: id
    fields: collector_id(string), created_at(string), data(object), date_created(string), date_modified(string), email(string), externalId(string), first_name(string), group_id(string), href(string), id(string), last_name(string), links(object), meta(object), name(string), page_id(string), schemas(array), status(string), subtype(string), survey_id(string), title(string), total(integer), type(string), updated_at(string), userName(string), username(string), workgroup_id(string)
  contacts:
    primary key: id
    fields: collector_id(string), created_at(string), data(object), date_created(string), date_modified(string), email(string), externalId(string), first_name(string), group_id(string), href(string), id(string), last_name(string), links(object), meta(object), name(string), page_id(string), schemas(array), status(string), subtype(string), survey_id(string), title(string), total(integer), type(string), updated_at(string), userName(string), username(string), workgroup_id(string)
  contact:
    primary key: id
    fields: collector_id(string), created_at(string), data(object), date_created(string), date_modified(string), email(string), externalId(string), first_name(string), group_id(string), href(string), id(string), last_name(string), links(object), meta(object), name(string), page_id(string), schemas(array), status(string), subtype(string), survey_id(string), title(string), total(integer), type(string), updated_at(string), userName(string), username(string), workgroup_id(string)
  contacts_bulk:
    primary key: id
    fields: collector_id(string), created_at(string), data(object), date_created(string), date_modified(string), email(string), externalId(string), first_name(string), group_id(string), href(string), id(string), last_name(string), links(object), meta(object), name(string), page_id(string), schemas(array), status(string), subtype(string), survey_id(string), title(string), total(integer), type(string), updated_at(string), userName(string), username(string), workgroup_id(string)
  contact_fields:
    primary key: id
    fields: collector_id(string), created_at(string), data(object), date_created(string), date_modified(string), email(string), externalId(string), first_name(string), group_id(string), href(string), id(string), last_name(string), links(object), meta(object), name(string), page_id(string), schemas(array), status(string), subtype(string), survey_id(string), title(string), total(integer), type(string), updated_at(string), userName(string), username(string), workgroup_id(string)
  contact_field:
    primary key: id
    fields: collector_id(string), created_at(string), data(object), date_created(string), date_modified(string), email(string), externalId(string), first_name(string), group_id(string), href(string), id(string), last_name(string), links(object), meta(object), name(string), page_id(string), schemas(array), status(string), subtype(string), survey_id(string), title(string), total(integer), type(string), updated_at(string), userName(string), username(string), workgroup_id(string)
  collector:
    primary key: id
    fields: collector_id(string), created_at(string), data(object), date_created(string), date_modified(string), email(string), externalId(string), first_name(string), group_id(string), href(string), id(string), last_name(string), links(object), meta(object), name(string), page_id(string), schemas(array), status(string), subtype(string), survey_id(string), title(string), total(integer), type(string), updated_at(string), userName(string), username(string), workgroup_id(string)
  collector_messages:
    primary key: id
    fields: collector_id(string), created_at(string), data(object), date_created(string), date_modified(string), email(string), externalId(string), first_name(string), group_id(string), href(string), id(string), last_name(string), links(object), meta(object), name(string), page_id(string), schemas(array), status(string), subtype(string), survey_id(string), title(string), total(integer), type(string), updated_at(string), userName(string), username(string), workgroup_id(string)
  collector_message:
    primary key: id
    fields: collector_id(string), created_at(string), data(object), date_created(string), date_modified(string), email(string), externalId(string), first_name(string), group_id(string), href(string), id(string), last_name(string), links(object), meta(object), name(string), page_id(string), schemas(array), status(string), subtype(string), survey_id(string), title(string), total(integer), type(string), updated_at(string), userName(string), username(string), workgroup_id(string)
  collector_message_recipients:
    primary key: id
    fields: collector_id(string), created_at(string), data(object), date_created(string), date_modified(string), email(string), externalId(string), first_name(string), group_id(string), href(string), id(string), last_name(string), links(object), meta(object), name(string), page_id(string), schemas(array), status(string), subtype(string), survey_id(string), title(string), total(integer), type(string), updated_at(string), userName(string), username(string), workgroup_id(string)
  collector_recipients:
    primary key: id
    fields: collector_id(string), created_at(string), data(object), date_created(string), date_modified(string), email(string), externalId(string), first_name(string), group_id(string), href(string), id(string), last_name(string), links(object), meta(object), name(string), page_id(string), schemas(array), status(string), subtype(string), survey_id(string), title(string), total(integer), type(string), updated_at(string), userName(string), username(string), workgroup_id(string)
  survey_collector_recipient:
    primary key: id
    fields: collector_id(string), created_at(string), data(object), date_created(string), date_modified(string), email(string), externalId(string), first_name(string), group_id(string), href(string), id(string), last_name(string), links(object), meta(object), name(string), page_id(string), schemas(array), status(string), subtype(string), survey_id(string), title(string), total(integer), type(string), updated_at(string), userName(string), username(string), workgroup_id(string)
  survey_collector_message_stats:
    primary key: id
    fields: collector_id(string), created_at(string), data(object), date_created(string), date_modified(string), email(string), externalId(string), first_name(string), group_id(string), href(string), id(string), last_name(string), links(object), meta(object), name(string), page_id(string), schemas(array), status(string), subtype(string), survey_id(string), title(string), total(integer), type(string), updated_at(string), userName(string), username(string), workgroup_id(string)
  collector_responses:
    primary key: id
    fields: collector_id(string), created_at(string), data(object), date_created(string), date_modified(string), email(string), externalId(string), first_name(string), group_id(string), href(string), id(string), last_name(string), links(object), meta(object), name(string), page_id(string), schemas(array), status(string), subtype(string), survey_id(string), title(string), total(integer), type(string), updated_at(string), userName(string), username(string), workgroup_id(string)
  collector_response:
    primary key: id
    fields: collector_id(string), created_at(string), data(object), date_created(string), date_modified(string), email(string), externalId(string), first_name(string), group_id(string), href(string), id(string), last_name(string), links(object), meta(object), name(string), page_id(string), schemas(array), status(string), subtype(string), survey_id(string), title(string), total(integer), type(string), updated_at(string), userName(string), username(string), workgroup_id(string)
  collector_response_details:
    primary key: id
    fields: collector_id(string), created_at(string), data(object), date_created(string), date_modified(string), email(string), externalId(string), first_name(string), group_id(string), href(string), id(string), last_name(string), links(object), meta(object), name(string), page_id(string), schemas(array), status(string), subtype(string), survey_id(string), title(string), total(integer), type(string), updated_at(string), userName(string), username(string), workgroup_id(string)
  collector_bulk_responses:
    primary key: id
    fields: collector_id(string), created_at(string), data(object), date_created(string), date_modified(string), email(string), externalId(string), first_name(string), group_id(string), href(string), id(string), last_name(string), links(object), meta(object), name(string), page_id(string), schemas(array), status(string), subtype(string), survey_id(string), title(string), total(integer), type(string), updated_at(string), userName(string), username(string), workgroup_id(string)
  collector_stats:
    primary key: id
    fields: collector_id(string), created_at(string), data(object), date_created(string), date_modified(string), email(string), externalId(string), first_name(string), group_id(string), href(string), id(string), last_name(string), links(object), meta(object), name(string), page_id(string), schemas(array), status(string), subtype(string), survey_id(string), title(string), total(integer), type(string), updated_at(string), userName(string), username(string), workgroup_id(string)
  webhooks:
    primary key: id
    fields: collector_id(string), created_at(string), data(object), date_created(string), date_modified(string), email(string), externalId(string), first_name(string), group_id(string), href(string), id(string), last_name(string), links(object), meta(object), name(string), page_id(string), schemas(array), status(string), subtype(string), survey_id(string), title(string), total(integer), type(string), updated_at(string), userName(string), username(string), workgroup_id(string)
  webhook:
    primary key: id
    fields: collector_id(string), created_at(string), data(object), date_created(string), date_modified(string), email(string), externalId(string), first_name(string), group_id(string), href(string), id(string), last_name(string), links(object), meta(object), name(string), page_id(string), schemas(array), status(string), subtype(string), survey_id(string), title(string), total(integer), type(string), updated_at(string), userName(string), username(string), workgroup_id(string)
  benchmark_bundles:
    primary key: id
    fields: collector_id(string), created_at(string), data(object), date_created(string), date_modified(string), email(string), externalId(string), first_name(string), group_id(string), href(string), id(string), last_name(string), links(object), meta(object), name(string), page_id(string), schemas(array), status(string), subtype(string), survey_id(string), title(string), total(integer), type(string), updated_at(string), userName(string), username(string), workgroup_id(string)
  benchmark_bundle:
    primary key: id
    fields: collector_id(string), created_at(string), data(object), date_created(string), date_modified(string), email(string), externalId(string), first_name(string), group_id(string), href(string), id(string), last_name(string), links(object), meta(object), name(string), page_id(string), schemas(array), status(string), subtype(string), survey_id(string), title(string), total(integer), type(string), updated_at(string), userName(string), username(string), workgroup_id(string)
  benchmark_bundle_analysis:
    primary key: id
    fields: collector_id(string), created_at(string), data(object), date_created(string), date_modified(string), email(string), externalId(string), first_name(string), group_id(string), href(string), id(string), last_name(string), links(object), meta(object), name(string), page_id(string), schemas(array), status(string), subtype(string), survey_id(string), title(string), total(integer), type(string), updated_at(string), userName(string), username(string), workgroup_id(string)
  organizations:
    primary key: id
    fields: collector_id(string), created_at(string), data(object), date_created(string), date_modified(string), email(string), externalId(string), first_name(string), group_id(string), href(string), id(string), last_name(string), links(object), meta(object), name(string), page_id(string), schemas(array), status(string), subtype(string), survey_id(string), title(string), total(integer), type(string), updated_at(string), userName(string), username(string), workgroup_id(string)
  organization:
    primary key: id
    fields: collector_id(string), created_at(string), data(object), date_created(string), date_modified(string), email(string), externalId(string), first_name(string), group_id(string), href(string), id(string), last_name(string), links(object), meta(object), name(string), page_id(string), schemas(array), status(string), subtype(string), survey_id(string), title(string), total(integer), type(string), updated_at(string), userName(string), username(string), workgroup_id(string)
  roles:
    primary key: id
    fields: collector_id(string), created_at(string), data(object), date_created(string), date_modified(string), email(string), externalId(string), first_name(string), group_id(string), href(string), id(string), last_name(string), links(object), meta(object), name(string), page_id(string), schemas(array), status(string), subtype(string), survey_id(string), title(string), total(integer), type(string), updated_at(string), userName(string), username(string), workgroup_id(string)
  errors:
    primary key: id
    fields: collector_id(string), created_at(string), data(object), date_created(string), date_modified(string), email(string), externalId(string), first_name(string), group_id(string), href(string), id(string), last_name(string), links(object), meta(object), name(string), page_id(string), schemas(array), status(string), subtype(string), survey_id(string), title(string), total(integer), type(string), updated_at(string), userName(string), username(string), workgroup_id(string)
  error:
    primary key: id
    fields: collector_id(string), created_at(string), data(object), date_created(string), date_modified(string), email(string), externalId(string), first_name(string), group_id(string), href(string), id(string), last_name(string), links(object), meta(object), name(string), page_id(string), schemas(array), status(string), subtype(string), survey_id(string), title(string), total(integer), type(string), updated_at(string), userName(string), username(string), workgroup_id(string)
  scim_users:
    primary key: id
    fields: collector_id(string), created_at(string), data(object), date_created(string), date_modified(string), email(string), externalId(string), first_name(string), group_id(string), href(string), id(string), last_name(string), links(object), meta(object), name(string), page_id(string), schemas(array), status(string), subtype(string), survey_id(string), title(string), total(integer), type(string), updated_at(string), userName(string), username(string), workgroup_id(string)
  scim_user:
    primary key: id
    fields: collector_id(string), created_at(string), data(object), date_created(string), date_modified(string), email(string), externalId(string), first_name(string), group_id(string), href(string), id(string), last_name(string), links(object), meta(object), name(string), page_id(string), schemas(array), status(string), subtype(string), survey_id(string), title(string), total(integer), type(string), updated_at(string), userName(string), username(string), workgroup_id(string)
  scim_schemas:
    primary key: id
    fields: collector_id(string), created_at(string), data(object), date_created(string), date_modified(string), email(string), externalId(string), first_name(string), group_id(string), href(string), id(string), last_name(string), links(object), meta(object), name(string), page_id(string), schemas(array), status(string), subtype(string), survey_id(string), title(string), total(integer), type(string), updated_at(string), userName(string), username(string), workgroup_id(string)
  scim_schema:
    primary key: id
    fields: collector_id(string), created_at(string), data(object), date_created(string), date_modified(string), email(string), externalId(string), first_name(string), group_id(string), href(string), id(string), last_name(string), links(object), meta(object), name(string), page_id(string), schemas(array), status(string), subtype(string), survey_id(string), title(string), total(integer), type(string), updated_at(string), userName(string), username(string), workgroup_id(string)
  scim_resource_types:
    primary key: id
    fields: collector_id(string), created_at(string), data(object), date_created(string), date_modified(string), email(string), externalId(string), first_name(string), group_id(string), href(string), id(string), last_name(string), links(object), meta(object), name(string), page_id(string), schemas(array), status(string), subtype(string), survey_id(string), title(string), total(integer), type(string), updated_at(string), userName(string), username(string), workgroup_id(string)
  scim_resource_type:
    primary key: id
    fields: collector_id(string), created_at(string), data(object), date_created(string), date_modified(string), email(string), externalId(string), first_name(string), group_id(string), href(string), id(string), last_name(string), links(object), meta(object), name(string), page_id(string), schemas(array), status(string), subtype(string), survey_id(string), title(string), total(integer), type(string), updated_at(string), userName(string), username(string), workgroup_id(string)
  scim_service_provider_config:
    primary key: id
    fields: collector_id(string), created_at(string), data(object), date_created(string), date_modified(string), email(string), externalId(string), first_name(string), group_id(string), href(string), id(string), last_name(string), links(object), meta(object), name(string), page_id(string), schemas(array), status(string), subtype(string), survey_id(string), title(string), total(integer), type(string), updated_at(string), userName(string), username(string), workgroup_id(string)

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite

REVERSE ETL ACTIONS
  update_collector:
    endpoint: PATCH /v3/collectors/{{ record.collector_id }}
    required fields: collector_id
    risk: updates SurveyMonkey collector data in the connected account; approval required
  replace_collector:
    endpoint: PUT /v3/collectors/{{ record.collector_id }}
    required fields: collector_id
    risk: replaces SurveyMonkey collector data in the connected account; approval required
  create_collector_message:
    endpoint: POST /v3/collectors/{{ record.collector_id }}/messages
    required fields: collector_id
    risk: creates SurveyMonkey collector message data in the connected account; approval required
  update_collector_message:
    endpoint: PATCH /v3/collectors/{{ record.collector_id }}/messages/{{ record.message_id }}
    required fields: collector_id, message_id
    risk: updates SurveyMonkey collector message data in the connected account; approval required
  replace_collector_message:
    endpoint: PUT /v3/collectors/{{ record.collector_id }}/messages/{{ record.message_id }}
    required fields: collector_id, message_id
    risk: replaces SurveyMonkey collector message data in the connected account; approval required
  create_message_recipient:
    endpoint: POST /v3/collectors/{{ record.collector_id }}/messages/{{ record.message_id }}/recipients
    required fields: collector_id, message_id
    risk: creates SurveyMonkey message recipient data in the connected account; approval required
  create_message_recipients_bulk:
    endpoint: POST /v3/collectors/{{ record.collector_id }}/messages/{{ record.message_id }}/recipients/bulk
    required fields: collector_id, message_id
    risk: creates SurveyMonkey bulk message recipients data in the connected account; approval required
  send_collector_message:
    endpoint: POST /v3/collectors/{{ record.collector_id }}/messages/{{ record.message_id }}/send
    required fields: collector_id, message_id
    risk: sends or shares SurveyMonkey collector message; may notify real recipients or change access for real users; approval required
  create_collector_response:
    endpoint: POST /v3/collectors/{{ record.collector_id }}/responses
    required fields: collector_id
    risk: creates SurveyMonkey collector response data in the connected account; approval required
  update_collector_response:
    endpoint: PATCH /v3/collectors/{{ record.collector_id }}/responses/{{ record.response_id }}
    required fields: collector_id, response_id
    risk: updates SurveyMonkey collector response data in the connected account; approval required
  replace_collector_response:
    endpoint: PUT /v3/collectors/{{ record.collector_id }}/responses/{{ record.response_id }}
    required fields: collector_id, response_id
    risk: replaces SurveyMonkey collector response data in the connected account; approval required
  create_contact_list:
    endpoint: POST /v3/contact_lists
    risk: creates SurveyMonkey contact list data in the connected account; approval required
  update_contact_list:
    endpoint: PATCH /v3/contact_lists/{{ record.contact_list_id }}
    required fields: contact_list_id
    risk: updates SurveyMonkey contact list data in the connected account; approval required
  replace_contact_list:
    endpoint: PUT /v3/contact_lists/{{ record.contact_list_id }}
    required fields: contact_list_id
    risk: replaces SurveyMonkey contact list data in the connected account; approval required
  copy_contact_list:
    endpoint: POST /v3/contact_lists/{{ record.contact_list_id }}/copy
    required fields: contact_list_id
    risk: creates SurveyMonkey copied contact list data in the connected account; approval required
  merge_contact_list:
    endpoint: POST /v3/contact_lists/{{ record.contact_list_id }}/merge
    required fields: contact_list_id
    risk: updates SurveyMonkey contact lists through merge data in the connected account; approval required
  add_contact_to_contact_list:
    endpoint: POST /v3/contact_lists/{{ record.contact_list_id }}/contacts
    required fields: contact_list_id
    risk: creates SurveyMonkey contact-list membership data in the connected account; approval required
  add_contacts_to_contact_list_bulk:
    endpoint: POST /v3/contact_lists/{{ record.contact_list_id }}/contacts/bulk
    required fields: contact_list_id
    risk: creates SurveyMonkey bulk contact-list memberships data in the connected account; approval required
  create_contact:
    endpoint: POST /v3/contacts
    risk: creates SurveyMonkey contact data in the connected account; approval required
  create_contacts_bulk:
    endpoint: POST /v3/contacts/bulk
    risk: creates SurveyMonkey bulk contacts data in the connected account; approval required
  update_contact:
    endpoint: PATCH /v3/contacts/{{ record.contact_id }}
    required fields: contact_id
    risk: updates SurveyMonkey contact data in the connected account; approval required
  replace_contact:
    endpoint: PUT /v3/contacts/{{ record.contact_id }}
    required fields: contact_id
    risk: replaces SurveyMonkey contact data in the connected account; approval required
  update_contact_field:
    endpoint: PATCH /v3/contact_fields/{{ record.contact_field_id }}
    required fields: contact_field_id
    risk: updates SurveyMonkey contact field data in the connected account; approval required
  create_organization:
    endpoint: POST /v3/organizations
    risk: changes SurveyMonkey organization administration/provisioning state; approval required
  update_organization:
    endpoint: PATCH /v3/organization/{{ record.organization_id }}
    required fields: organization_id
    risk: changes SurveyMonkey organization administration/provisioning state; approval required
  create_survey_folder:
    endpoint: POST /v3/survey_folders
    risk: creates SurveyMonkey survey folder data in the connected account; approval required
  create_survey:
    endpoint: POST /v3/surveys
    risk: creates SurveyMonkey survey data in the connected account; approval required
  update_survey:
    endpoint: PATCH /v3/surveys/{{ record.survey_id }}
    required fields: survey_id
    risk: updates SurveyMonkey survey data in the connected account; approval required
  replace_survey:
    endpoint: PUT /v3/surveys/{{ record.survey_id }}
    required fields: survey_id
    risk: replaces SurveyMonkey survey data in the connected account; approval required
  share_survey:
    endpoint: POST /v3/surveys/{{ record.survey_id }}/share
    required fields: survey_id
    risk: sends or shares SurveyMonkey survey access; may notify real recipients or change access for real users; approval required
  create_survey_page:
    endpoint: POST /v3/surveys/{{ record.survey_id }}/pages
    required fields: survey_id
    risk: creates SurveyMonkey survey page data in the connected account; approval required
  update_survey_page:
    endpoint: PATCH /v3/surveys/{{ record.survey_id }}/pages/{{ record.page_id }}
    required fields: survey_id, page_id
    risk: updates SurveyMonkey survey page data in the connected account; approval required
  replace_survey_page:
    endpoint: PUT /v3/surveys/{{ record.survey_id }}/pages/{{ record.page_id }}
    required fields: survey_id, page_id
    risk: replaces SurveyMonkey survey page data in the connected account; approval required
  create_survey_question:
    endpoint: POST /v3/surveys/{{ record.survey_id }}/pages/{{ record.page_id }}/questions
    required fields: survey_id, page_id
    risk: creates SurveyMonkey survey question data in the connected account; approval required
  create_survey_language:
    endpoint: POST /v3/surveys/{{ record.survey_id }}/languages/{{ record.language_code }}
    required fields: survey_id, language_code
    risk: creates SurveyMonkey survey language/translation data in the connected account; approval required
  update_survey_language:
    endpoint: PATCH /v3/surveys/{{ record.survey_id }}/languages/{{ record.language_code }}
    required fields: survey_id, language_code
    risk: updates SurveyMonkey survey language/translation data in the connected account; approval required
  create_survey_collector:
    endpoint: POST /v3/surveys/{{ record.survey_id }}/collectors
    required fields: survey_id
    risk: creates SurveyMonkey survey collector data in the connected account; approval required
  update_survey_response:
    endpoint: PATCH /v3/surveys/{{ record.survey_id }}/responses/{{ record.response_id }}
    required fields: survey_id, response_id
    risk: updates SurveyMonkey survey response data in the connected account; approval required
  replace_survey_response:
    endpoint: PUT /v3/surveys/{{ record.survey_id }}/responses/{{ record.response_id }}
    required fields: survey_id, response_id
    risk: replaces SurveyMonkey survey response data in the connected account; approval required
  create_webhook:
    endpoint: POST /v3/webhooks
    risk: creates SurveyMonkey webhook data in the connected account; approval required
  update_webhook:
    endpoint: PATCH /v3/webhooks/{{ record.webhook_id }}
    required fields: webhook_id
    risk: updates SurveyMonkey webhook data in the connected account; approval required
  replace_webhook:
    endpoint: PUT /v3/webhooks/{{ record.webhook_id }}
    required fields: webhook_id
    risk: replaces SurveyMonkey webhook data in the connected account; approval required
  update_group_member:
    endpoint: PATCH /v3/groups/{{ record.group_id }}/members/{{ record.member_id }}
    required fields: group_id, member_id
    risk: changes SurveyMonkey group membership administration/provisioning state; approval required
  create_workgroup:
    endpoint: POST /v3/workgroups
    risk: changes SurveyMonkey workgroup administration/provisioning state; approval required
  update_workgroup:
    endpoint: PATCH /v3/workgroups/{{ record.workgroup_id }}
    required fields: workgroup_id
    risk: changes SurveyMonkey workgroup administration/provisioning state; approval required
  create_workgroup_share:
    endpoint: POST /v3/workgroups/{{ record.workgroup_id }}/shares
    required fields: workgroup_id
    risk: sends or shares SurveyMonkey workgroup share; may notify real recipients or change access for real users; approval required
  create_workgroup_shares_bulk:
    endpoint: POST /v3/workgroups/{{ record.workgroup_id }}/shares/bulk
    required fields: workgroup_id
    risk: sends or shares SurveyMonkey bulk workgroup shares; may notify real recipients or change access for real users; approval required
  create_workgroup_member:
    endpoint: POST /v3/workgroups/{{ record.workgroup_id }}/members
    required fields: workgroup_id
    risk: changes SurveyMonkey workgroup membership administration/provisioning state; approval required
  create_workgroup_members_bulk:
    endpoint: POST /v3/workgroups/{{ record.workgroup_id }}/members/bulk
    required fields: workgroup_id
    risk: changes SurveyMonkey bulk workgroup membership administration/provisioning state; approval required
  update_workgroup_member:
    endpoint: PATCH /v3/workgroups/{{ record.workgroup_id }}/members/{{ record.member_id }}
    required fields: workgroup_id, member_id
    risk: changes SurveyMonkey workgroup membership administration/provisioning state; approval required
  create_scim_user:
    endpoint: POST /scim/v2/Users
    risk: changes SurveyMonkey SCIM user provisioning administration/provisioning state; approval required
  replace_scim_user:
    endpoint: PUT /scim/v2/Users/{{ record.scim_user_id }}
    required fields: scim_user_id
    risk: changes SurveyMonkey SCIM user provisioning administration/provisioning state; approval required
  update_scim_user:
    endpoint: PATCH /scim/v2/Users/{{ record.scim_user_id }}
    required fields: scim_user_id
    risk: changes SurveyMonkey SCIM user provisioning administration/provisioning state; approval required

SECURITY
  read risk: external SurveyMonkey API read of survey, collector, contact, team, webhook, benchmark, error, organization, and SCIM data
  write risk: external SurveyMonkey API mutations can create or update surveys, collectors, messages, contacts, contact lists, workgroups, organizations, webhooks, and SCIM users; message send/share actions can notify real recipients
  approval: reverse ETL writes require plan preview and approval token before any SurveyMonkey mutation is executed
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect surveymonkey

  # Inspect as structured JSON
  pm connectors inspect surveymonkey --json

AGENT WORKFLOW
  - Run pm connectors inspect surveymonkey before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.
  - For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```

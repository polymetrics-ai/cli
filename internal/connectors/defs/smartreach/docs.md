# Overview

Reads SmartReach teams, campaigns, prospects, email settings, do-not-contact records, users, and
accounts; creates/updates prospects and accounts, manages campaign membership/status, do-not-contact
entries, and task status.

Readable streams: `campaigns`, `prospects`, `teams`, `email_settings`, `do_not_contact`, `users`,
`accounts`.

Write actions: `add_or_update_prospects`, `add_prospects_to_campaign`,
`unassign_prospects_from_campaign`, `update_prospect_campaign_status`, `update_campaign_status`,
`remove_from_do_not_contact`, `add_emails_to_do_not_contact`, `add_domains_to_do_not_contact`,
`create_or_update_account`, `update_task_status`.

Service API documentation: https://help.smartreach.io/reference/getprospects.

## Auth setup

Connection fields:

- `api_key` (required, secret, string); SmartReach API key, sent as the X-API-KEY header. Never
  logged.
- `base_url` (optional, string); default `https://api.smartreach.io/api/v3`; format `uri`;
  SmartReach API base URL override for tests or proxies.
- `newer_than` (optional, string); Optional newer_than query filter passed through verbatim.
- `older_than` (optional, string); Optional older_than query filter passed through verbatim.
- `team_id` (optional, string); Optional team_id query filter, applied to
  campaigns/prospects/email_settings/do_not_contact streams (not teams).

Secret fields are redacted in logs and write previews: `api_key`.

Default configuration values: `base_url=https://api.smartreach.io/api/v3`.

Authentication behavior:

- API key authentication in `X-API-KEY` using `secrets.api_key`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `teams`.

## Streams notes

Default pagination: single request; no pagination.

- `campaigns`: GET `campaigns` - records path `campaigns`; query `newer_than` from template `{{
  config.newer_than }}`, omitted when absent; `older_than` from template `{{ config.older_than }}`,
  omitted when absent; `team_id` from template `{{ config.team_id }}`, omitted when absent; emits
  passthrough records.
- `prospects`: GET `prospects` - records path `prospects`; query `newer_than` from template `{{
  config.newer_than }}`, omitted when absent; `older_than` from template `{{ config.older_than }}`,
  omitted when absent; `team_id` from template `{{ config.team_id }}`, omitted when absent; emits
  passthrough records.
- `teams`: GET `teams` - records path `teams`; query `newer_than` from template `{{
  config.newer_than }}`, omitted when absent; `older_than` from template `{{ config.older_than }}`,
  omitted when absent; emits passthrough records.
- `email_settings`: GET `email-settings` - records path `email_settings`; query `newer_than` from
  template `{{ config.newer_than }}`, omitted when absent; `older_than` from template `{{
  config.older_than }}`, omitted when absent; `team_id` from template `{{ config.team_id }}`,
  omitted when absent; emits passthrough records.
- `do_not_contact`: GET `do-not-contact` - records path `dnc`; query `newer_than` from template `{{
  config.newer_than }}`, omitted when absent; `older_than` from template `{{ config.older_than }}`,
  omitted when absent; `team_id` from template `{{ config.team_id }}`, omitted when absent; emits
  passthrough records.
- `users`: GET `users` - records path `users`; query `newer_than` from template `{{
  config.newer_than }}`, omitted when absent; `older_than` from template `{{ config.older_than }}`,
  omitted when absent; `team_id` from template `{{ config.team_id }}`, omitted when absent; emits
  passthrough records.
- `accounts`: GET `search/accounts` - records path `accounts`; query `newer_than` from template `{{
  config.newer_than }}`, omitted when absent; `older_than` from template `{{ config.older_than }}`,
  omitted when absent; `team_id` from template `{{ config.team_id }}`, omitted when absent; emits
  passthrough records.

## Write actions & risks

Overall write risk: creates/updates prospects and CRM accounts, enrolls/unenrolls prospects in
outbound campaigns, starts/stops entire campaigns, changes prospect engagement status, blacklists
emails/domains from outreach, and changes sales task status.

Reverse ETL writes should be planned, previewed, approved, and then executed. Declared actions:

- `add_or_update_prospects`: POST `prospects?team_id={{ config.team_id }}` - kind `upsert`; body
  type `json`; accepted fields `city`, `company`, `country`, `custom_fields`, `email`, `first_name`,
  `last_name`, `linkedin_url`, `list`, `owner_id`, `phone_number`, `state`, `tags`, `timezone`;
  risk: external mutation; creates or updates a prospect (deduped by
  email/phone/linkedin_url/company+name per SmartReach's unique_identifier_columns rule) on the
  connected SmartReach account; approval required.
- `add_prospects_to_campaign`: POST `campaigns/{{ record.campaign_id }}/prospects?team_id={{
  config.team_id }}` - kind `custom`; body type `json`; path fields `campaign_id`; required record
  fields `campaign_id`, `prospect_ids`; accepted fields `campaign_id`,
  `ignore_prospects_in_other_campaigns`, `prospect_ids`; risk: external mutation; enrolls prospects
  into an outbound campaign, triggering scheduled sequence messages to real recipients; approval
  required.
- `unassign_prospects_from_campaign`: PUT `campaigns/{{ record.campaign_id }}/prospects?team_id={{
  config.team_id }}` - kind `custom`; body type `json`; path fields `campaign_id`; required record
  fields `campaign_id`, `prospect_ids`; accepted fields `campaign_id`, `prospect_ids`; risk:
  external mutation; removes prospects from an outbound campaign, stopping any further scheduled
  sequence messages to them; approval required.
- `update_prospect_campaign_status`: PUT `prospects/prospect_status_change?team_id={{ config.team_id
  }}` - kind `update`; body type `json`; required record fields `prospect_ids`, `prospect_status`,
  `campaign_ids`; accepted fields `campaign_ids`, `prospect_ids`, `prospect_status`,
  `will_resume_at`; risk: external mutation; changes a prospect's engagement status for a campaign
  (e.g. pausing/resuming/marking replied), affecting whether further outreach messages are sent;
  approval required.
- `update_campaign_status`: PUT `campaigns/{{ record.campaign_id }}/status?team_id={{ config.team_id
  }}` - kind `update`; body type `json`; path fields `campaign_id`; required record fields
  `campaign_id`, `status`; accepted fields `campaign_id`, `schedule_start_at`, `status`; risk:
  external mutation; starts, schedules, or stops an entire outbound campaign, directly controlling
  whether real outreach messages are sent to prospects; approval required.
- `remove_from_do_not_contact`: DELETE `do-not-contact?team_id={{ config.team_id }}` - kind
  `delete`; body type `json`; required record fields `dnc_ids`; accepted fields `dnc_ids`; risk:
  external mutation; removes entries from the do-not-contact suppression list, which re-enables
  outreach to those emails/domains; approval required.
- `add_emails_to_do_not_contact`: POST `do-not-contact/email?team_id={{ config.team_id }}` - kind
  `create`; body type `json`; required record fields `emails`; accepted fields `emails`; risk:
  external mutation; blacklists emails from all future outreach on the connected SmartReach account;
  approval required.
- `add_domains_to_do_not_contact`: POST `do-not-contact/domain?team_id={{ config.team_id }}` - kind
  `create`; body type `json`; required record fields `domains`; accepted fields `domains`; risk:
  external mutation; blacklists entire domains from all future outreach on the connected SmartReach
  account; approval required.
- `create_or_update_account`: POST `accounts?team_id={{ config.team_id }}` - kind `upsert`; body
  type `json`; required record fields `name`; accepted fields `custom_fields`, `custom_id`,
  `description`, `industry`, `linkedin_url`, `name`, `update_account`, `website`; risk: external
  mutation; creates or updates a CRM-style account (company) record on the connected SmartReach
  account; approval required.
- `update_task_status`: PUT `tasks/{{ record.task_id }}/status` - kind `update`; body type `json`;
  path fields `task_id`; body fields `status_type`, `due_at`, `snoozed_till`; required record fields
  `task_id`, `status_type`; accepted fields `due_at`, `snoozed_till`, `status_type`, `task_id`;
  risk: external mutation; changes a sales task's status (due/snoozed/done/skipped) on the connected
  SmartReach account; approval required.

## Known limits

- API coverage includes 7 stream-backed endpoint group(s), 10 write-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  duplicate_of=6, out_of_scope=2.

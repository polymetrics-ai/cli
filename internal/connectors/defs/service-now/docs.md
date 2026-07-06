# Overview

Reads and writes ServiceNow incident, user, and group table data through the ServiceNow Table API.

Readable streams: `incidents`, `users`, `groups`.

Write actions: `create_incident`, `create_user`, `create_group`, `update_incident`, `update_user`,
`update_group`.

Service API documentation:
https://www.servicenow.com/docs/bundle/zurich-api-reference/page/integrate/inbound-rest/concept/c_TableAPI.html.

## Auth setup

Connection fields:

- `base_url` (required, string); format `uri`; Your ServiceNow instance root, e.g.
  https://<instance>.service-now.com.
- `mode` (optional, string).
- `password` (required, secret, string); ServiceNow password, sent as the HTTP Basic password. Never
  logged.
- `username` (required, string); ServiceNow username, sent as the HTTP Basic username.

Secret fields are redacted in logs and write previews: `password`.

Authentication behavior:

- HTTP Basic authentication using `config.username`, `secrets.password`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/api/now/table/sys_user`.

## Streams notes

Default pagination: offset/limit pagination; offset parameter `sysparm_offset`; limit parameter
`sysparm_limit`; page size 100; maximum 1 page(s).

- `incidents`: GET `/api/now/table/incident` - records path `result`; offset/limit pagination;
  offset parameter `sysparm_offset`; limit parameter `sysparm_limit`; page size 100; maximum 1
  page(s); emits passthrough records.
- `users`: GET `/api/now/table/sys_user` - records path `result`; offset/limit pagination; offset
  parameter `sysparm_offset`; limit parameter `sysparm_limit`; page size 100; maximum 1 page(s);
  emits passthrough records.
- `groups`: GET `/api/now/table/sys_user_group` - records path `result`; offset/limit pagination;
  offset parameter `sysparm_offset`; limit parameter `sysparm_limit`; page size 100; maximum 1
  page(s); emits passthrough records.

## Write actions & risks

Overall write risk: creates incident/user/group records and updates their fields by sys_id
(ServiceNow Table API PATCH, which modifies only submitted fields); creating/deactivating a user
account is a higher-scrutiny mutation than incident/group create-update.

Reverse ETL writes should be planned, previewed, approved, and then executed. Declared actions:

- `create_incident`: POST `/api/now/table/incident` - kind `create`; body type `json`; accepted
  fields `assigned_to`, `assignment_group`, `caller_id`, `category`, `description`, `impact`,
  `priority`, `short_description`, `state`, `urgency`; risk: creates a new incident record; low-risk
  external mutation (a new ticket), no approval required.
- `create_user`: POST `/api/now/table/sys_user` - kind `create`; body type `json`; required record
  fields `user_name`; accepted fields `active`, `department`, `email`, `first_name`, `last_name`,
  `name`, `user_name`; risk: creates a new ServiceNow user account record; a new user account
  granted whatever role/ACL defaults the instance applies is a higher-scrutiny mutation than an
  incident/group create.
- `create_group`: POST `/api/now/table/sys_user_group` - kind `create`; body type `json`; required
  record fields `name`; accepted fields `active`, `description`, `manager`, `name`; risk: creates a
  new user group record; low-risk external mutation, no approval required.
- `update_incident`: PATCH `/api/now/table/incident/{{ record.sys_id }}` - kind `update`; body type
  `json`; path fields `sys_id`; required record fields `sys_id`; accepted fields `assigned_to`,
  `assignment_group`, `category`, `description`, `impact`, `priority`, `short_description`, `state`,
  `sys_id`, `urgency`; risk: mutates an existing incident's recorded fields (only fields present in
  the submitted record are changed; ServiceNow's Table API PATCH/PUT both modify only the submitted
  fields, never the whole record) by sys_id.
- `update_user`: PATCH `/api/now/table/sys_user/{{ record.sys_id }}` - kind `update`; body type
  `json`; path fields `sys_id`; required record fields `sys_id`; accepted fields `active`,
  `department`, `email`, `first_name`, `last_name`, `name`, `sys_id`; risk: mutates an existing user
  account's profile fields by sys_id, including active (deactivating a user's account revokes their
  instance access); higher-scrutiny than incident/group updates.
- `update_group`: PATCH `/api/now/table/sys_user_group/{{ record.sys_id }}` - kind `update`; body
  type `json`; path fields `sys_id`; required record fields `sys_id`; accepted fields `active`,
  `description`, `manager`, `name`, `sys_id`; risk: mutates an existing group's recorded fields by
  sys_id, including active/manager; can change who is considered the group's membership owner.

## Known limits

- Batch defaults: read_page_size=100.
- API coverage includes 3 stream-backed endpoint group(s), 6 write-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  binary_payload=1, destructive_admin=3, duplicate_of=6, out_of_scope=3.

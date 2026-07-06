# Overview

Reads Eventee event agenda, attendee, registration, group, review, and partner data; writes
documented Eventee agenda, attendee, registration, partner, speaker, and track mutations through the
public REST API.

Readable streams: `lectures`, `speakers`, `days`, `halls`, `tracks`, `workshops`, `pauses`,
`partners`, `reviews`, `groups`, `participants`, `registrations`.

Write actions: `clear_test_content`, `create_hall`, `update_hall`, `delete_hall`, `create_lecture`,
`update_lecture`, `delete_lecture`, `invite_attendees`, `update_attendee_checkin`,
`remove_attendee`, `create_partner`, `update_partner`, `delete_partner`, `create_pause`,
`update_pause`, `delete_pause`, `invite_registrations`, `remove_registration`, `create_speaker`,
`update_speaker`, `delete_speaker`, `create_track`, `update_track`, `delete_track`.

Service API documentation: https://publiceventeeapi.docs.apiary.io/.

## Auth setup

Connection fields:

- `api_token` (required, secret, string); Eventee public API token, sent as a Bearer token
  (Authorization: Bearer <api_token>). Never logged.
- `base_url` (optional, string); default `https://api.eventee.co/public/v1`; format `uri`; Eventee
  API base URL override for tests or proxies.

Secret fields are redacted in logs and write previews: `api_token`.

Default configuration values: `base_url=https://api.eventee.co/public/v1`.

Authentication behavior:

- Bearer token authentication using `secrets.api_token`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/content`.

## Streams notes

Default pagination: single request; no pagination.

- `lectures`: GET `/content` - records path `lectures`.
- `speakers`: GET `/content` - records path `speakers`.
- `days`: GET `/content` - records path `days`.
- `halls`: GET `/content` - records path `halls`.
- `tracks`: GET `/content` - records path `tracks`.
- `workshops`: GET `/content` - records path `workshops`.
- `pauses`: GET `/content` - records path `pauses`.
- `partners`: GET `/partners` - records at response root.
- `reviews`: GET `/reviews` - records at response root.
- `groups`: GET `/groups` - records at response root.
- `participants`: GET `/participants` - records at response root.
- `registrations`: GET `/registrations` - records at response root.

## Write actions & risks

Overall write risk: creates, updates, invites, checks in, removes, or deletes Eventee event content
and attendees/registrants; destructive deletes require approval.

Reverse ETL writes should be planned, previewed, approved, and then executed. Declared actions:

- `clear_test_content`: DELETE `/test/content` - kind `delete`; body type `none`; confirmation
  `destructive`; risk: deletes all tracks, pauses, speakers, workshops, lectures, and halls from the
  configured test event.
- `create_hall`: POST `/hall` - kind `create`; body type `json`; required record fields `name`;
  accepted fields `name`; risk: creates a hall in the configured event.
- `update_hall`: PATCH `/hall/{{ record.id }}` - kind `update`; body type `json`; path fields `id`;
  required record fields `id`, `name`; accepted fields `id`, `name`; risk: updates a hall in the
  configured event.
- `delete_hall`: DELETE `/hall/{{ record.id }}` - kind `delete`; body type `none`; path fields `id`;
  required record fields `id`; accepted fields `id`; confirmation `destructive`; risk: deletes a
  hall from the configured event.
- `create_lecture`: POST `/lecture` - kind `create`; body type `json`; required record fields
  `name`, `start`, `end`, `hall_id`, `speakers`, `type`, `tracks`; accepted fields `capacity`,
  `description`, `discussion`, `end`, `hall_id`, `moderating`, `name`, `polling`, `speakers`,
  `start`, `stream`, `tracks`, `type`, `virtual_meeting`; risk: creates a lecture or session in the
  configured event.
- `update_lecture`: PATCH `/lecture/{{ record.id }}` - kind `update`; body type `json`; path fields
  `id`; required record fields `id`, `name`, `start`, `end`, `hall_id`, `speakers`, `type`,
  `tracks`; accepted fields `capacity`, `description`, `discussion`, `end`, `hall_id`, `id`,
  `moderating`, `name`, `polling`, `speakers`, `start`, `stream`, `tracks`, `type`,
  `virtual_meeting`; risk: updates an existing lecture or session in the configured event.
- `delete_lecture`: DELETE `/lecture/{{ record.id }}` - kind `delete`; body type `none`; path fields
  `id`; required record fields `id`; accepted fields `id`; confirmation `destructive`; risk: deletes
  a lecture or session from the configured event.
- `invite_attendees`: PUT `/attendee/invite` - kind `create`; body type `json`; required record
  fields `users`; accepted fields `users`; risk: invites one or more attendees to the configured
  event.
- `update_attendee_checkin`: PUT `/attendee/{{ record.id }}/checkin` - kind `update`; body type
  `json`; path fields `id`; required record fields `id`, `checkin`; accepted fields `checkin`, `id`;
  risk: sets the check-in state for an attendee.
- `remove_attendee`: DELETE `/attendee` - kind `delete`; body type `json`; required record fields
  `email`; accepted fields `email`; confirmation `destructive`; risk: removes an invited attendee
  and may remove their access and event-linked information.
- `create_partner`: POST `/partner` - kind `create`; body type `json`; required record fields
  `company`; accepted fields `address`, `booth_number`, `company`, `description`, `email`,
  `logo_url`, `phone`, `photo_url`, `web`; risk: creates a partner, sponsor, or exhibitor profile in
  the configured event.
- `update_partner`: PATCH `/partner/{{ record.id }}` - kind `update`; body type `json`; path fields
  `id`; required record fields `id`, `company`; accepted fields `address`, `booth_number`,
  `company`, `description`, `email`, `id`, `logo_url`, `phone`, `photo_url`, `web`; risk: updates an
  existing partner, sponsor, or exhibitor profile in the configured event.
- `delete_partner`: DELETE `/partner/{{ record.id }}` - kind `delete`; body type `none`; path fields
  `id`; required record fields `id`; accepted fields `id`; confirmation `destructive`; risk: deletes
  a partner, sponsor, or exhibitor profile from the configured event.
- `create_pause`: POST `/pause` - kind `create`; body type `json`; required record fields `name`,
  `start`, `end`; accepted fields `description`, `end`, `name`, `start`; risk: creates a pause or
  break in the configured event agenda.
- `update_pause`: PATCH `/pause/{{ record.id }}` - kind `update`; body type `json`; path fields
  `id`; required record fields `id`, `name`, `start`, `end`; accepted fields `description`, `end`,
  `id`, `name`, `start`; risk: updates an existing pause or break in the configured event agenda.
- `delete_pause`: DELETE `/pause/{{ record.id }}` - kind `delete`; body type `none`; path fields
  `id`; required record fields `id`; accepted fields `id`; confirmation `destructive`; risk: deletes
  a pause or break from the configured event agenda.
- `invite_registrations`: PUT `/registration/invite` - kind `create`; body type `json`; required
  record fields `registrations`; accepted fields `registrations`; risk: invites one or more
  registrants to the configured event.
- `remove_registration`: DELETE `/registration` - kind `delete`; body type `json`; required record
  fields `email`; accepted fields `email`; confirmation `destructive`; risk: removes an invited
  registrant from the configured event.
- `create_speaker`: POST `/speaker` - kind `create`; body type `json`; required record fields
  `name`, `phone`; accepted fields `bio`, `company`, `country`, `email`, `facebook`, `language`,
  `linkedIn`, `name`, `phone`, `photo`, `position`, `twitter`, `web`; risk: creates a speaker
  profile in the configured event.
- `update_speaker`: PATCH `/speaker/{{ record.id }}` - kind `update`; body type `json`; path fields
  `id`; required record fields `id`, `name`, `phone`; accepted fields `bio`, `company`, `country`,
  `email`, `facebook`, `id`, `language`, `linkedIn`, `name`, `phone`, `photo`, `position`,
  `twitter`, `web`; risk: updates an existing speaker profile in the configured event.
- `delete_speaker`: DELETE `/speaker/{{ record.id }}` - kind `delete`; body type `none`; path fields
  `id`; required record fields `id`; accepted fields `id`; confirmation `destructive`; risk: deletes
  a speaker profile from the configured event.
- `create_track`: POST `/label` - kind `create`; body type `json`; accepted fields `color`, `name`;
  risk: creates a track label in the configured event.
- `update_track`: PATCH `/label/{{ record.id }}` - kind `update`; body type `json`; path fields
  `id`; required record fields `id`; accepted fields `color`, `id`, `name`; risk: updates an
  existing track label in the configured event.
- `delete_track`: DELETE `/label/{{ record.id }}` - kind `delete`; body type `none`; path fields
  `id`; required record fields `id`; accepted fields `id`; confirmation `destructive`; risk: deletes
  a track label from the configured event.

## Known limits

- Batch defaults: read_page_size=100, write_batch_size=1.
- API coverage includes 12 stream-backed endpoint group(s), 24 write-backed endpoint group(s).

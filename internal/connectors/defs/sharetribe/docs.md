# Overview

Reads and writes Sharetribe listings, users, transactions, availability, stock, and marketplace data
through the Sharetribe Integration API.

Readable streams: `listings`, `users`, `transactions`, `events`, `marketplace`, `files`,
`file_attachments`.

Write actions: `create_listing`, `update_listing`, `close_listing`, `open_listing`,
`approve_listing`, `approve_user`, `update_user_profile`, `update_user_permissions`,
`verify_user_email`, `transition_transaction`, `transition_transaction_speculative`,
`update_transaction_metadata`, `create_availability_exception`, `delete_availability_exception`,
`set_listing_stock`, `create_stock_adjustment`.

Service API documentation: https://www.sharetribe.com/api-reference/integration.html.

## Auth setup

Connection fields:

- `base_url` (optional, string); default `https://flex-api.sharetribe.com/v1`; format `uri`;
  Sharetribe Integration API base URL override for tests or proxies.
- `mode` (optional, string).
- `oauth_access_token` (required, secret, string); Sharetribe Integration API OAuth2 access token,
  sent as a Bearer token. Never logged.

Secret fields are redacted in logs and write previews: `oauth_access_token`.

Default configuration values: `base_url=https://flex-api.sharetribe.com/v1`.

Authentication behavior:

- Bearer token authentication using `secrets.oauth_access_token`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/integration_api/listings/query`.

## Streams notes

Default pagination: page-number pagination; page parameter `page`; size parameter `per_page`; starts
at 1; page size 100; maximum 1 page(s).

Pagination by stream: none: `marketplace`; page_number: `listings`, `users`, `transactions`,
`events`, `files`, `file_attachments`.

- `listings`: GET `/integration_api/listings/query` - records path `data`; page-number pagination;
  page parameter `page`; size parameter `per_page`; starts at 1; page size 100; maximum 1 page(s);
  emits passthrough records.
- `users`: GET `/integration_api/users/query` - records path `data`; page-number pagination; page
  parameter `page`; size parameter `per_page`; starts at 1; page size 100; maximum 1 page(s); emits
  passthrough records.
- `transactions`: GET `/integration_api/transactions/query` - records path `data`; page-number
  pagination; page parameter `page`; size parameter `per_page`; starts at 1; page size 100; maximum
  1 page(s); emits passthrough records.
- `events`: GET `/integration_api/events/query` - records path `data`; page-number pagination; page
  parameter `page`; size parameter `per_page`; starts at 1; page size 100; maximum 1 page(s); emits
  passthrough records.
- `marketplace`: GET `/integration_api/marketplace/show` - records path `data`; emits passthrough
  records.
- `files`: GET `/integration_api/files/query` - records path `data`; page-number pagination; page
  parameter `page`; size parameter `per_page`; starts at 1; page size 100; maximum 1 page(s); emits
  passthrough records.
- `file_attachments`: GET `/integration_api/file_attachments/query` - records path `data`;
  page-number pagination; page parameter `page`; size parameter `per_page`; starts at 1; page size
  100; maximum 1 page(s); emits passthrough records.

## Write actions & risks

Overall write risk: creates/updates/closes/opens/approves listings, approves/updates users and their
permissions, transitions transactions (including real payment/payout process actions), manages
availability exceptions and listing stock.

Reverse ETL writes should be planned, previewed, approved, and then executed. Declared actions:

- `create_listing`: POST `/integration_api/listings/create` - kind `create`; body type `json`;
  required record fields `title`, `authorId`; accepted fields `authorId`, `availabilityPlan`,
  `description`, `geolocation`, `images`, `metadata`, `price`, `privateData`, `publicData`, `state`,
  `title`; risk: creates a new marketplace listing; low-risk external mutation, no approval
  required.
- `update_listing`: POST `/integration_api/listings/update` - kind `update`; body type `json`;
  required record fields `id`; accepted fields `availabilityPlan`, `description`, `geolocation`,
  `id`, `images`, `metadata`, `price`, `privateData`, `publicData`, `title`; risk: mutates an
  existing listing's details by id; cannot be used to change listing state (use
  close_listing/open_listing/approve_listing for that); publicData/privateData/metadata are merged
  with the existing object on the top level, not deep-merged.
- `close_listing`: POST `/integration_api/listings/close` - kind `custom`; body type `json`;
  required record fields `id`; accepted fields `id`; risk: sets the listing's state to closed; it
  stops being discoverable via the public Marketplace API listings/query endpoint but remains
  reachable by id or through related transactions.
- `open_listing`: POST `/integration_api/listings/open` - kind `custom`; body type `json`; required
  record fields `id`; accepted fields `id`; risk: sets a closed listing's state back to published,
  making it publicly discoverable again; low-risk, no approval required.
- `approve_listing`: POST `/integration_api/listings/approve` - kind `custom`; body type `json`;
  required record fields `id`; accepted fields `id`; risk: approves a listing currently in
  pendingApproval state, setting it to published and making it publicly visible; review before
  enabling in a caller with untrusted input if the marketplace relies on manual listing moderation.
- `approve_user`: POST `/integration_api/users/approve` - kind `custom`; body type `json`; required
  record fields `id`; accepted fields `id`; risk: approves a pending user account, setting its state
  to active and granting it marketplace access; higher-scrutiny than listing writes since it grants
  account access.
- `update_user_profile`: POST `/integration_api/users/update_profile` - kind `update`; body type
  `json`; required record fields `id`; accepted fields `bio`, `displayName`, `firstName`, `id`,
  `lastName`, `metadata`, `privateData`, `profileImageId`, `protectedData`, `publicData`; risk:
  mutates an existing user's profile fields by id; publicData/protectedData/privateData/metadata are
  merged with the existing object on the top level, not deep-merged.
- `update_user_permissions`: POST `/integration_api/users/update_permissions` - kind `update`; body
  type `json`; required record fields `id`; accepted fields `id`, `initiateTransactions`,
  `postListings`, `read`; risk: changes what a user is permitted to do on the marketplace (post
  listings, initiate transactions, read); a higher-scrutiny access-control mutation, review before
  enabling in a caller with untrusted input.
- `verify_user_email`: POST `/integration_api/users/verify_email` - kind `custom`; body type `json`;
  required record fields `id`, `email`; accepted fields `email`, `id`; risk: marks the given email
  address as verified for the user; low-risk account-state mutation.
- `transition_transaction`: POST `/integration_api/transactions/transition` - kind `custom`; body
  type `json`; required record fields `id`, `transition`, `params`; accepted fields `id`, `params`,
  `transition`; risk: transitions a real transaction to a new state via the marketplace's
  transaction process (e.g. accepting/declining a booking, marking payment); only operator-actor
  transitions are permitted; a maximum of 100 transitions per transaction; can trigger real payment
  capture/payout actions depending on the process definition.
- `transition_transaction_speculative`: POST `/integration_api/transactions/transition_speculative`
  - kind `custom`; body type `json`; required record fields `id`, `transition`, `params`; accepted
  fields `id`, `params`, `transition`; risk: simulates a transaction transition to validate
  parameters or preview the resulting price breakdown; the transaction state is NOT actually changed
  - safe to call freely, no approval required.
- `update_transaction_metadata`: POST `/integration_api/transactions/update_metadata` - kind
  `update`; body type `json`; required record fields `id`, `metadata`; accepted fields `id`,
  `metadata`; risk: mutates an existing transaction's metadata (merged with the existing object on
  the top level, not deep-merged); low-risk, does not affect payment/process state.
- `create_availability_exception`: POST `/integration_api/availability_exceptions/create` - kind
  `create`; body type `json`; required record fields `listingId`, `seats`, `start`, `end`; accepted
  fields `end`, `listingId`, `seats`, `start`; risk: overrides a listing's availability plan for a
  given time range (e.g. blocking it out or opening extra seats); low-risk external mutation, no
  approval required.
- `delete_availability_exception`: POST `/integration_api/availability_exceptions/delete` - kind
  `delete`; body type `json`; required record fields `id`; accepted fields `id`; risk: permanently
  removes an availability exception by id, restoring the listing's default availability plan for
  that time range.
- `set_listing_stock`: POST `/integration_api/stock/compare_and_set` - kind `update`; body type
  `json`; required record fields `listingId`, `newTotal`; accepted fields `listingId`, `newTotal`,
  `oldTotal`; risk: sets a listing's total available stock via a compare-and-set (only applied if
  the listing's current stock matches oldTotal); low-risk external mutation, no approval required.
- `create_stock_adjustment`: POST `/integration_api/stock_adjustments/create` - kind `create`; body
  type `json`; required record fields `listingId`, `quantity`; accepted fields `listingId`,
  `quantity`; risk: creates an immutable stock adjustment for a listing (increases or decreases
  available stock by quantity); low-risk external mutation, no approval required.

## Known limits

- Batch defaults: read_page_size=100.
- API coverage includes 7 stream-backed endpoint group(s), 16 write-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  binary_payload=1, duplicate_of=3, out_of_scope=4.

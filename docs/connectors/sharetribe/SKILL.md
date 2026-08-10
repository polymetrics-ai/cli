---
name: pm-sharetribe
description: Sharetribe connector knowledge and safe action guide.
---

# pm-sharetribe

## Purpose

Reads and writes Sharetribe listings, users, transactions, availability, stock, and marketplace data through the Sharetribe Integration API.

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

- base_url
- mode
- oauth_access_token (secret)

## ETL Streams

- listings:
  - primary key: id
  - fields: attributes(object), id(string), type(string), updated_at(string)
- users:
  - primary key: id
  - fields: attributes(object), id(string), type(string), updated_at(string)
- transactions:
  - primary key: id
  - fields: attributes(object), id(string), type(string), updated_at(string)
- events:
  - primary key: id
  - fields: attributes(object), id(string), type(string), updated_at(string)
- marketplace:
  - primary key: id
  - fields: attributes(object), id(string), type(string)
- files:
  - primary key: id
  - fields: attributes(object), id(string), relationships(object), type(string)
- file_attachments:
  - primary key: id
  - fields: attributes(object), id(string), relationships(object), type(string)

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

## Reverse ETL Actions

- create_listing:
  - endpoint: POST /integration_api/listings/create
  - required fields: title, authorId
  - risk: creates a new marketplace listing; low-risk external mutation, no approval required
- update_listing:
  - endpoint: POST /integration_api/listings/update
  - required fields: id
  - risk: mutates an existing listing's details by id; cannot be used to change listing state (use close_listing/open_listing/approve_listing for that); publicData/privateData/metadata are merged with the existing object on the top level, not deep-merged
- close_listing:
  - endpoint: POST /integration_api/listings/close
  - required fields: id
  - risk: sets the listing's state to closed; it stops being discoverable via the public Marketplace API listings/query endpoint but remains reachable by id or through related transactions
- open_listing:
  - endpoint: POST /integration_api/listings/open
  - required fields: id
  - risk: sets a closed listing's state back to published, making it publicly discoverable again; low-risk, no approval required
- approve_listing:
  - endpoint: POST /integration_api/listings/approve
  - required fields: id
  - risk: approves a listing currently in pendingApproval state, setting it to published and making it publicly visible; review before enabling in a caller with untrusted input if the marketplace relies on manual listing moderation
- approve_user:
  - endpoint: POST /integration_api/users/approve
  - required fields: id
  - risk: approves a pending user account, setting its state to active and granting it marketplace access; higher-scrutiny than listing writes since it grants account access
- update_user_profile:
  - endpoint: POST /integration_api/users/update_profile
  - required fields: id
  - risk: mutates an existing user's profile fields by id; publicData/protectedData/privateData/metadata are merged with the existing object on the top level, not deep-merged
- update_user_permissions:
  - endpoint: POST /integration_api/users/update_permissions
  - required fields: id
  - risk: changes what a user is permitted to do on the marketplace (post listings, initiate transactions, read); a higher-scrutiny access-control mutation, review before enabling in a caller with untrusted input
- verify_user_email:
  - endpoint: POST /integration_api/users/verify_email
  - required fields: id, email
  - risk: marks the given email address as verified for the user; low-risk account-state mutation
- transition_transaction:
  - endpoint: POST /integration_api/transactions/transition
  - required fields: id, transition, params
  - risk: transitions a real transaction to a new state via the marketplace's transaction process (e.g. accepting/declining a booking, marking payment); only operator-actor transitions are permitted; a maximum of 100 transitions per transaction; can trigger real payment capture/payout actions depending on the process definition
- transition_transaction_speculative:
  - endpoint: POST /integration_api/transactions/transition_speculative
  - required fields: id, transition, params
  - risk: simulates a transaction transition to validate parameters or preview the resulting price breakdown; the transaction state is NOT actually changed — safe to call freely, no approval required
- update_transaction_metadata:
  - endpoint: POST /integration_api/transactions/update_metadata
  - required fields: id, metadata
  - risk: mutates an existing transaction's metadata (merged with the existing object on the top level, not deep-merged); low-risk, does not affect payment/process state
- create_availability_exception:
  - endpoint: POST /integration_api/availability_exceptions/create
  - required fields: listingId, seats, start, end
  - risk: overrides a listing's availability plan for a given time range (e.g. blocking it out or opening extra seats); low-risk external mutation, no approval required
- delete_availability_exception:
  - endpoint: POST /integration_api/availability_exceptions/delete
  - required fields: id
  - risk: permanently removes an availability exception by id, restoring the listing's default availability plan for that time range
- set_listing_stock:
  - endpoint: POST /integration_api/stock/compare_and_set
  - required fields: listingId, newTotal
  - risk: sets a listing's total available stock via a compare-and-set (only applied if the listing's current stock matches oldTotal); low-risk external mutation, no approval required
- create_stock_adjustment:
  - endpoint: POST /integration_api/stock_adjustments/create
  - required fields: listingId, quantity
  - risk: creates an immutable stock adjustment for a listing (increases or decreases available stock by quantity); low-risk external mutation, no approval required

## Security

- read risk: external Sharetribe Integration API read of listing, user, transaction, event, marketplace, file, and file-attachment data
- write risk: creates/updates/closes/opens/approves listings, approves/updates users and their permissions, transitions transactions (including real payment/payout process actions), manages availability exceptions and listing stock
- approval: none for listing/user-profile/stock/availability-exception mutations (reversible, low-risk marketplace-operator actions); review transition_transaction and update_user_permissions before enabling in a caller with untrusted input, since transitions can trigger real payment capture/payout and permission changes affect account access
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Command Surface

- Run Sharetribe's declared streams and reverse-ETL actions.
- Usage: pm sharetribe <command> [flags]
- Read streams
- Reverse ETL writes
- Other Commands
  - api get integration-api availability-exceptions query - Documented GET /integration_api/availability_exceptions/query (not implemented) [intent=direct_read availability=not_implemented operation=sharetribe.get.integration-api-availability-exceptions-query]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get integration-api listings show - Documented GET /integration_api/listings/show (not implemented) [intent=direct_read availability=not_implemented operation=sharetribe.get.integration-api-listings-show]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get integration-api messages query - Documented GET /integration_api/messages/query (not implemented) [intent=direct_read availability=not_implemented operation=sharetribe.get.integration-api-messages-query]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get integration-api stock-adjustments query - Documented GET /integration_api/stock_adjustments/query (not implemented) [intent=direct_read availability=not_implemented operation=sharetribe.get.integration-api-stock-adjustments-query]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get integration-api stock-reservations show - Documented GET /integration_api/stock_reservations/show (not implemented) [intent=direct_read availability=not_implemented operation=sharetribe.get.integration-api-stock-reservations-show]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get integration-api transactions show - Documented GET /integration_api/transactions/show (not implemented) [intent=direct_read availability=not_implemented operation=sharetribe.get.integration-api-transactions-show]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get integration-api users show - Documented GET /integration_api/users/show (not implemented) [intent=direct_read availability=not_implemented operation=sharetribe.get.integration-api-users-show]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get listings query - Documented GET /listings/query (not implemented) [intent=direct_read availability=not_implemented operation=sharetribe.get.listings-query]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get listings show - Documented GET /listings/show (not implemented) [intent=direct_read availability=not_implemented operation=sharetribe.get.listings-show]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get reviews query - Documented GET /reviews/query (not implemented) [intent=direct_read availability=not_implemented operation=sharetribe.get.reviews-query]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get reviews show - Documented GET /reviews/show (not implemented) [intent=direct_read availability=not_implemented operation=sharetribe.get.reviews-show]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get timeslots query - Documented GET /timeslots/query (not implemented) [intent=direct_read availability=not_implemented operation=sharetribe.get.timeslots-query]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get users show - Documented GET /users/show (not implemented) [intent=direct_read availability=not_implemented operation=sharetribe.get.users-show]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api post integration-api images upload - Documented POST /integration_api/images/upload (not implemented) [intent=direct_write availability=not_implemented operation=sharetribe.post.integration-api-images-upload]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api post own-listings create - Documented POST /own_listings/create (not implemented) [intent=direct_write availability=not_implemented operation=sharetribe.post.own-listings-create]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api post own-listings create-draft - Documented POST /own_listings/create_draft (not implemented) [intent=direct_write availability=not_implemented operation=sharetribe.post.own-listings-create-draft]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api post own-listings open - Documented POST /own_listings/open (not implemented) [intent=direct_write availability=not_implemented operation=sharetribe.post.own-listings-open]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api post own-listings publish-draft - Documented POST /own_listings/publish_draft (not implemented) [intent=direct_write availability=not_implemented operation=sharetribe.post.own-listings-publish-draft]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api post transactions initiate - Documented POST /transactions/initiate (not implemented) [intent=direct_write availability=not_implemented operation=sharetribe.post.transactions-initiate]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api post v1 integration-api file-downloads create - Documented POST /v1/integration_api/file_downloads/create (not implemented) [intent=direct_write availability=not_implemented operation=sharetribe.post.v1-integration-api-file-downloads-create]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api post v1 integration-api file-uploads create - Documented POST /v1/integration_api/file_uploads/create (not implemented) [intent=direct_write availability=not_implemented operation=sharetribe.post.v1-integration-api-file-uploads-create]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api post v1 integration-api files create - Documented POST /v1/integration_api/files/create (not implemented) [intent=direct_write availability=not_implemented operation=sharetribe.post.v1-integration-api-files-create]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - approve listing apply - Plan and execute the approve listing reverse-ETL action [intent=reverse_etl availability=implemented write=approve_listing]; approval: requires plan, preview, approval, and execute; risk: approves a listing currently in pendingApproval state, setting it to published and making it publicly visible; review before enabling in a caller with untrusted input if the marketplace relies on manual listing moderation; flags: --id (required)
  - approve user apply - Plan and execute the approve user reverse-ETL action [intent=reverse_etl availability=implemented write=approve_user]; approval: requires plan, preview, approval, and execute; risk: approves a pending user account, setting its state to active and granting it marketplace access; higher-scrutiny than listing writes since it grants account access; flags: --id (required)
  - close listing apply - Plan and execute the close listing reverse-ETL action [intent=reverse_etl availability=implemented write=close_listing]; approval: requires plan, preview, approval, and execute; risk: sets the listing's state to closed; it stops being discoverable via the public Marketplace API listings/query endpoint but remains reachable by id or through related transactions; flags: --id (required)
  - create availability exception apply - Plan and execute the create availability exception reverse-ETL action [intent=reverse_etl availability=implemented write=create_availability_exception]; approval: requires plan, preview, approval, and execute; risk: overrides a listing's availability plan for a given time range (e.g. blocking it out or opening extra seats); low-risk external mutation, no approval required; flags: --end (required), --listingId (required), --seats (required), --start (required)
  - create listing apply - Plan and execute the create listing reverse-ETL action [intent=reverse_etl availability=implemented write=create_listing]; approval: requires plan, preview, approval, and execute; risk: creates a new marketplace listing; low-risk external mutation, no approval required; flags: --authorId (required), --title (required)
  - create stock adjustment apply - Plan and execute the create stock adjustment reverse-ETL action [intent=reverse_etl availability=implemented write=create_stock_adjustment]; approval: requires plan, preview, approval, and execute; risk: creates an immutable stock adjustment for a listing (increases or decreases available stock by quantity); low-risk external mutation, no approval required; flags: --listingId (required), --quantity (required)
  - delete availability exception apply - Plan and execute the delete availability exception reverse-ETL action [intent=reverse_etl availability=implemented write=delete_availability_exception]; approval: requires plan, preview, approval, and execute; risk: permanently removes an availability exception by id, restoring the listing's default availability plan for that time range; flags: --id (required)
  - events list - Run the events ETL stream [intent=etl availability=implemented stream=events]
  - file attachments list - Run the file attachments ETL stream [intent=etl availability=implemented stream=file_attachments]
  - files list - Run the files ETL stream [intent=etl availability=implemented stream=files]
  - listings list - Run the listings ETL stream [intent=etl availability=implemented stream=listings]
  - marketplace list - Run the marketplace ETL stream [intent=etl availability=implemented stream=marketplace]
  - open listing apply - Plan and execute the open listing reverse-ETL action [intent=reverse_etl availability=implemented write=open_listing]; approval: requires plan, preview, approval, and execute; risk: sets a closed listing's state back to published, making it publicly discoverable again; low-risk, no approval required; flags: --id (required)
  - set listing stock apply - Plan and execute the set listing stock reverse-ETL action [intent=reverse_etl availability=implemented write=set_listing_stock]; approval: requires plan, preview, approval, and execute; risk: sets a listing's total available stock via a compare-and-set (only applied if the listing's current stock matches oldTotal); low-risk external mutation, no approval required; flags: --listingId (required), --newTotal (required)
  - transactions list - Run the transactions ETL stream [intent=etl availability=implemented stream=transactions]
  - transition transaction apply - Plan and execute the transition transaction reverse-ETL action [intent=reverse_etl availability=not_implemented write=transition_transaction]; approval: requires plan, preview, approval, and execute; risk: transitions a real transaction to a new state via the marketplace's transaction process (e.g. accepting/declining a booking, marking payment); only operator-actor transitions are permitted; a maximum of 100 transitions per transaction; can trigger real payment capture/payout actions depending on the process definition; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
  - transition transaction speculative apply - Plan and execute the transition transaction speculative reverse-ETL action [intent=reverse_etl availability=not_implemented write=transition_transaction_speculative]; approval: requires plan, preview, approval, and execute; risk: simulates a transaction transition to validate parameters or preview the resulting price breakdown; the transaction state is NOT actually changed — safe to call freely, no approval required; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
  - update listing apply - Plan and execute the update listing reverse-ETL action [intent=reverse_etl availability=implemented write=update_listing]; approval: requires plan, preview, approval, and execute; risk: mutates an existing listing's details by id; cannot be used to change listing state (use close_listing/open_listing/approve_listing for that); publicData/privateData/metadata are merged with the existing object on the top level, not deep-merged; flags: --id (required)
  - update transaction metadata apply - Plan and execute the update transaction metadata reverse-ETL action [intent=reverse_etl availability=not_implemented write=update_transaction_metadata]; approval: requires plan, preview, approval, and execute; risk: mutates an existing transaction's metadata (merged with the existing object on the top level, not deep-merged); low-risk, does not affect payment/process state; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
  - update user permissions apply - Plan and execute the update user permissions reverse-ETL action [intent=reverse_etl availability=implemented write=update_user_permissions]; approval: requires plan, preview, approval, and execute; risk: changes what a user is permitted to do on the marketplace (post listings, initiate transactions, read); a higher-scrutiny access-control mutation, review before enabling in a caller with untrusted input; flags: --id (required)
  - update user profile apply - Plan and execute the update user profile reverse-ETL action [intent=reverse_etl availability=implemented write=update_user_profile]; approval: requires plan, preview, approval, and execute; risk: mutates an existing user's profile fields by id; publicData/protectedData/privateData/metadata are merged with the existing object on the top level, not deep-merged; flags: --id (required)
  - users list - Run the users ETL stream [intent=etl availability=implemented stream=users]
  - verify user email apply - Plan and execute the verify user email reverse-ETL action [intent=reverse_etl availability=implemented write=verify_user_email]; approval: requires plan, preview, approval, and execute; risk: marks the given email address as verified for the user; low-risk account-state mutation; flags: --email (required), --id (required)

## Commands

### Inspect as a manual

```bash
pm connectors inspect sharetribe
```

### Inspect as structured JSON

```bash
pm connectors inspect sharetribe --json
```

## Agent Rules

- Run pm connectors inspect sharetribe before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
- For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.

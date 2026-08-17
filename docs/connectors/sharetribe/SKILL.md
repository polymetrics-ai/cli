---
name: pm-sharetribe
description: Sharetribe connector knowledge and safe action guide.
---

# pm-sharetribe

## Purpose

Reads and writes Sharetribe listings, users, transactions, availability, stock, and marketplace data through the Sharetribe Integration API.

## Icon

- asset: icons/pm-sample.svg
- source: polymetrics
- review_status: polymetrics

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
  - fields: attributes(), id(), type(), updated_at()
- users:
  - primary key: id
  - fields: attributes(), id(), type(), updated_at()
- transactions:
  - primary key: id
  - fields: attributes(), id(), type(), updated_at()
- events:
  - primary key: id
  - fields: attributes(), id(), type(), updated_at()
- marketplace:
  - primary key: id
  - fields: attributes(), id(), type()
- files:
  - primary key: id
  - fields: attributes(), id(), relationships(), type()
- file_attachments:
  - primary key: id
  - fields: attributes(), id(), relationships(), type()

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

## Reverse ETL Actions

- create_listing:
  - endpoint: POST /integration_api/listings/create
  - risk: creates a new marketplace listing; low-risk external mutation, no approval required
- update_listing:
  - endpoint: POST /integration_api/listings/update
  - risk: mutates an existing listing's details by id; cannot be used to change listing state (use close_listing/open_listing/approve_listing for that); publicData/privateData/metadata are merged with the existing object on the top level, not deep-merged
- close_listing:
  - endpoint: POST /integration_api/listings/close
  - risk: sets the listing's state to closed; it stops being discoverable via the public Marketplace API listings/query endpoint but remains reachable by id or through related transactions
- open_listing:
  - endpoint: POST /integration_api/listings/open
  - risk: sets a closed listing's state back to published, making it publicly discoverable again; low-risk, no approval required
- approve_listing:
  - endpoint: POST /integration_api/listings/approve
  - risk: approves a listing currently in pendingApproval state, setting it to published and making it publicly visible; review before enabling in a caller with untrusted input if the marketplace relies on manual listing moderation
- approve_user:
  - endpoint: POST /integration_api/users/approve
  - risk: approves a pending user account, setting its state to active and granting it marketplace access; higher-scrutiny than listing writes since it grants account access
- update_user_profile:
  - endpoint: POST /integration_api/users/update_profile
  - risk: mutates an existing user's profile fields by id; publicData/protectedData/privateData/metadata are merged with the existing object on the top level, not deep-merged
- update_user_permissions:
  - endpoint: POST /integration_api/users/update_permissions
  - risk: changes what a user is permitted to do on the marketplace (post listings, initiate transactions, read); a higher-scrutiny access-control mutation, review before enabling in a caller with untrusted input
- verify_user_email:
  - endpoint: POST /integration_api/users/verify_email
  - risk: marks the given email address as verified for the user; low-risk account-state mutation
- transition_transaction:
  - endpoint: POST /integration_api/transactions/transition
  - risk: transitions a real transaction to a new state via the marketplace's transaction process (e.g. accepting/declining a booking, marking payment); only operator-actor transitions are permitted; a maximum of 100 transitions per transaction; can trigger real payment capture/payout actions depending on the process definition
- transition_transaction_speculative:
  - endpoint: POST /integration_api/transactions/transition_speculative
  - risk: simulates a transaction transition to validate parameters or preview the resulting price breakdown; the transaction state is NOT actually changed — safe to call freely, no approval required
- update_transaction_metadata:
  - endpoint: POST /integration_api/transactions/update_metadata
  - risk: mutates an existing transaction's metadata (merged with the existing object on the top level, not deep-merged); low-risk, does not affect payment/process state
- create_availability_exception:
  - endpoint: POST /integration_api/availability_exceptions/create
  - risk: overrides a listing's availability plan for a given time range (e.g. blocking it out or opening extra seats); low-risk external mutation, no approval required
- delete_availability_exception:
  - endpoint: POST /integration_api/availability_exceptions/delete
  - risk: permanently removes an availability exception by id, restoring the listing's default availability plan for that time range
- set_listing_stock:
  - endpoint: POST /integration_api/stock/compare_and_set
  - risk: sets a listing's total available stock via a compare-and-set (only applied if the listing's current stock matches oldTotal); low-risk external mutation, no approval required
- create_stock_adjustment:
  - endpoint: POST /integration_api/stock_adjustments/create
  - risk: creates an immutable stock adjustment for a listing (increases or decreases available stock by quantity); low-risk external mutation, no approval required

## Security

- read risk: external Sharetribe Integration API read of listing, user, transaction, event, marketplace, file, and file-attachment data
- write risk: creates/updates/closes/opens/approves listings, approves/updates users and their permissions, transitions transactions (including real payment/payout process actions), manages availability exceptions and listing stock
- approval: none for listing/user-profile/stock/availability-exception mutations (reversible, low-risk marketplace-operator actions); review transition_transaction and update_user_permissions before enabling in a caller with untrusted input, since transitions can trigger real payment capture/payout and permission changes affect account access
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

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

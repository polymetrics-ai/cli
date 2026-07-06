# Overview

Reads Cloudbeds guests, hotels, rooms, reservations, transactions, rate plans, room types, items,
taxes/fees, sources, groups, house accounts, housekeeping, custom fields, payment methods, webhooks,
allotment blocks, and guest/reservation notes, and writes guest, reservation, payment, folio,
housekeeping, house-account, webhook, and room-block actions, through the Cloudbeds v1.2 REST API.

Readable streams: `guests`, `hotels`, `rooms`, `reservations`, `transactions`, `rate_plans`,
`room_types`, `items`, `item_categories`, `taxes_and_fees`, `sources`, `groups`, `house_accounts`,
`housekeepers`, `housekeeping_status`, `custom_fields`, `payment_methods`, `webhooks`,
`allotment_blocks`, `guest_notes`, `reservation_notes`.

Write actions: `create_guest`, `update_guest`, `create_guest_note`, `update_guest_note`,
`delete_guest_note`, `create_reservation`, `update_reservation`, `create_reservation_note`,
`update_reservation_note`, `delete_reservation_note`, `check_in_room`, `check_out_room`,
`assign_room`, `post_payment`, `void_payment`, `post_adjustment`, `delete_adjustment`, `post_item`,
`void_item`, `create_item_category`, `create_group_note`, `update_housekeeping_status`,
`create_housekeeper`, `update_housekeeper`, `assign_housekeeping`, `create_house_account`,
`update_house_account_status`, `create_webhook`, `delete_webhook`, `create_room_block`,
`update_room_block`, `delete_room_block`.

Service API documentation: https://hotels.cloudbeds.com/api/docs/.

## Auth setup

Connection fields:

- `api_key` (required, secret, string); Cloudbeds API access token, sent as a Bearer token. Used
  only for auth; never logged.
- `base_url` (optional, string); default `https://api.cloudbeds.com/api/v1.2`; format `uri`;
  Cloudbeds API base URL override for tests or proxies.
- `property_id` (optional, string); Cloudbeds property (hotel) ID. Required for the groups and
  allotment_blocks streams specifically (Cloudbeds itself requires propertyID for GET /getGroups and
  /getAllotmentBlocks); ignored by every other stream, which reads across all accessible properties.
- `rate_plans_end_date` (optional, string); format `date`; Exclusive end date (YYYY-MM-DD) for the
  rate_plans stream's rate-availability window. Required by the rate_plans stream specifically;
  ignored by every other stream.
- `rate_plans_start_date` (optional, string); format `date`; Inclusive start date (YYYY-MM-DD) for
  the rate_plans stream's rate-availability window. Required by the rate_plans stream specifically
  (Cloudbeds itself requires startDate/endDate for GET /getRatePlans); ignored by every other
  stream.

Secret fields are redacted in logs and write previews: `api_key`.

Default configuration values: `base_url=https://api.cloudbeds.com/api/v1.2`.

Authentication behavior:

- Bearer token authentication using `secrets.api_key`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/getHotels`.

## Streams notes

Default pagination: page-number pagination; page parameter `pageNumber`; size parameter `pageSize`;
starts at 1; page size 100.

Pagination by stream: none: `rate_plans`, `items`, `item_categories`, `taxes_and_fees`, `sources`,
`house_accounts`, `custom_fields`, `payment_methods`, `webhooks`, `guest_notes`,
`reservation_notes`; page_number: `guests`, `hotels`, `rooms`, `reservations`, `transactions`,
`room_types`, `groups`, `housekeepers`, `housekeeping_status`, `allotment_blocks`.

- `guests`: GET `/getGuestList` - records path `data`; page-number pagination; page parameter
  `pageNumber`; size parameter `pageSize`; starts at 1; page size 100.
- `hotels`: GET `/getHotels` - records path `data`; page-number pagination; page parameter
  `pageNumber`; size parameter `pageSize`; starts at 1; page size 100.
- `rooms`: GET `/getRoomBlocks` - records path `data`; page-number pagination; page parameter
  `pageNumber`; size parameter `pageSize`; starts at 1; page size 100.
- `reservations`: GET `/getReservations` - records path `data`; page-number pagination; page
  parameter `pageNumber`; size parameter `pageSize`; starts at 1; page size 100.
- `transactions`: GET `/getTransactions` - records path `data`; page-number pagination; page
  parameter `pageNumber`; size parameter `pageSize`; starts at 1; page size 100.
- `rate_plans`: GET `/getRatePlans` - records path `data`; query `endDate`=`{{
  config.rate_plans_end_date }}`; `startDate`=`{{ config.rate_plans_start_date }}`.
- `room_types`: GET `/getRoomTypes` - records path `data`; page-number pagination; page parameter
  `pageNumber`; size parameter `pageSize`; starts at 1; page size 100.
- `items`: GET `/getItems` - records path `data`.
- `item_categories`: GET `/getItemCategories` - records path `data`.
- `taxes_and_fees`: GET `/getTaxesAndFees` - records path `data`.
- `sources`: GET `/getSources` - records path `data`.
- `groups`: GET `/getGroups` - records path `data`; query `propertyID`=`{{ config.property_id }}`;
  page-number pagination; page parameter `pageNumber`; size parameter `pageSize`; starts at 1; page
  size 100.
- `house_accounts`: GET `/getHouseAccountList` - records path `data`.
- `housekeepers`: GET `/getHousekeepers` - records path `data`; page-number pagination; page
  parameter `pageNumber`; size parameter `pageSize`; starts at 1; page size 100.
- `housekeeping_status`: GET `/getHousekeepingStatus` - records path `data`; page-number pagination;
  page parameter `pageNumber`; size parameter `pageSize`; starts at 1; page size 100.
- `custom_fields`: GET `/getCustomFields` - records path `data`.
- `payment_methods`: GET `/getPaymentMethods` - records path `data`.
- `webhooks`: GET `/getWebhooks` - records path `data`.
- `allotment_blocks`: GET `/getAllotmentBlocks` - records path `data`; query `propertyID`=`{{
  config.property_id }}`; page-number pagination; page parameter `pageNumber`; size parameter
  `pageSize`; starts at 1; page size 100.
- `guest_notes`: GET `/getGuestNotes` - records path `data`; fan-out; ids from request
  `/getGuestList`; id-list records path `data`; id field `guestID`; id sent as query parameter
  `guestID`; stamps `guestID`.
- `reservation_notes`: GET `/getReservationNotes` - records path `data`; fan-out; ids from request
  `/getReservations`; id-list records path `data`; id field `reservationID`; id sent as query
  parameter `reservationID`; stamps `reservationID`.

## Write actions & risks

Overall write risk: external mutation of Cloudbeds guests, reservations, payments/folio items, house
accounts, housekeeping state, groups, webhooks, and room blocks - several actions are real financial
transactions (post_payment, void_payment, post_adjustment, post_item, void_item) or irreversible
deletions.

Reverse ETL writes should be planned, previewed, approved, and then executed. Declared actions:

- `create_guest`: POST `/postGuest` - kind `create`; body type `form`; required record fields
  `propertyID`, `reservationID`; accepted fields `guestAddress1`, `guestCity`, `guestCountry`,
  `guestEmail`, `guestFirstName`, `guestLastName`, `guestPhone`, `guestZip`, `propertyID`,
  `reservationID`; risk: adds an additional guest to an existing reservation; no approval required.
- `update_guest`: PUT `/putGuest` - kind `update`; body type `form`; required record fields
  `propertyID`, `guestID`; accepted fields `guestEmail`, `guestFirstName`, `guestID`,
  `guestLastName`, `guestPhone`, `propertyID`; risk: overwrites a guest's profile fields; no
  approval required.
- `create_guest_note`: POST `/postGuestNote` - kind `create`; body type `form`; required record
  fields `propertyID`, `guestID`, `guestNote`; accepted fields `guestID`, `guestNote`, `propertyID`,
  `userID`; risk: adds a free-text note to a guest profile; no approval required.
- `update_guest_note`: PUT `/putGuestNote` - kind `update`; body type `form`; required record fields
  `propertyID`, `guestID`, `noteID`, `guestNote`; accepted fields `guestID`, `guestNote`, `noteID`,
  `propertyID`; risk: overwrites an existing guest note's text; no approval required.
- `delete_guest_note`: DELETE `/deleteGuestNote?propertyID={{ record.propertyID }}&guestID={{
  record.guestID }}&noteID={{ record.noteID }}` - kind `delete`; body type `none`; path fields
  `propertyID`, `guestID`, `noteID`; required record fields `propertyID`, `guestID`, `noteID`;
  accepted fields `guestID`, `noteID`, `propertyID`; missing records treated as success for status
  `404`; risk: irreversibly removes a guest note; approval required.
- `create_reservation`: POST `/postReservation` - kind `create`; body type `form`; required record
  fields `propertyID`, `startDate`, `endDate`, `guestFirstName`, `guestLastName`, `rooms`; accepted
  fields `adults`, `children`, `endDate`, `guestEmail`, `guestFirstName`, `guestLastName`,
  `paymentMethod`, `propertyID`, `rooms`, `sendEmailConfirmation`, `sourceID`, `startDate`; risk:
  creates a new, billable reservation (books rooms and may authorize/charge a card depending on
  paymentMethod); external mutation, no approval required for the booking itself but review
  payment-capturing config carefully.
- `update_reservation`: PUT `/putReservation` - kind `update`; body type `form`; required record
  fields `propertyID`, `reservationID`; accepted fields `checkoutDate`, `estimatedArrivalTime`,
  `propertyID`, `reservationID`, `rooms`, `sendStatusChangeEmail`, `status`; risk: mutates an
  existing reservation's status/dates/room assignment; a status change to a cancelled/no-show state
  has real revenue and inventory-release consequences; no approval required (Cloudbeds' own
  confirmation flow, e.g. sendStatusChangeEmail, is the operator-facing guardrail).
- `create_reservation_note`: POST `/postReservationNote` - kind `create`; body type `form`; required
  record fields `propertyID`, `reservationID`, `reservationNote`; accepted fields `propertyID`,
  `reservationID`, `reservationNote`, `userID`; risk: adds a free-text note to a reservation; no
  approval required.
- `update_reservation_note`: PUT `/putReservationNote` - kind `update`; body type `form`; required
  record fields `propertyID`, `reservationID`, `reservationNoteID`, `reservationNote`; accepted
  fields `propertyID`, `reservationID`, `reservationNote`, `reservationNoteID`; risk: overwrites an
  existing reservation note's text; no approval required.
- `delete_reservation_note`: DELETE `/deleteReservationNote?propertyID={{ record.propertyID
  }}&reservationID={{ record.reservationID }}&reservationNoteID={{ record.reservationNoteID }}` -
  kind `delete`; body type `none`; path fields `propertyID`, `reservationID`, `reservationNoteID`;
  required record fields `propertyID`, `reservationID`, `reservationNoteID`; accepted fields
  `propertyID`, `reservationID`, `reservationNoteID`; missing records treated as success for status
  `404`; risk: irreversibly removes a reservation note; approval required.
- `check_in_room`: POST `/postRoomCheckIn` - kind `update`; body type `form`; required record fields
  `propertyID`, `reservationID`, `roomID`; accepted fields `propertyID`, `reservationID`, `roomID`,
  `subReservationID`; risk: checks a guest into a room, changing reservation/room status; no
  approval required.
- `check_out_room`: POST `/postRoomCheckOut` - kind `update`; body type `form`; required record
  fields `propertyID`, `reservationID`, `roomID`; accepted fields `propertyID`, `reservationID`,
  `roomID`, `subReservationID`; risk: checks a guest out of a room, changing reservation/room status
  and folio; no approval required.
- `assign_room`: POST `/postRoomAssign` - kind `update`; body type `form`; required record fields
  `propertyID`, `reservationID`, `reservationRoomID`, `newRoomID`; accepted fields `adjustPrice`,
  `newRoomID`, `oldRoomID`, `overrideRates`, `propertyID`, `reservationID`, `reservationRoomID`,
  `roomTypeID`, `subReservationID`; risk: reassigns a reservation to a different physical room,
  optionally repricing; no approval required.
- `post_payment`: POST `/postPayment` - kind `create`; body type `form`; required record fields
  `propertyID`, `type`, `amount`; accepted fields `amount`, `cardType`, `description`, `groupCode`,
  `houseAccountID`, `isDeposit`, `propertyID`, `reservationID`, `subReservationID`, `type`; risk:
  records a payment/deposit against a reservation or house account; a real financial transaction -
  approval required.
- `void_payment`: POST `/postVoidPayment` - kind `delete`; body type `form`; required record fields
  `propertyID`, `paymentID`; accepted fields `houseAccountID`, `paymentID`, `propertyID`,
  `reservationID`; risk: irreversibly voids a recorded payment; a real financial reversal - approval
  required.
- `post_adjustment`: POST `/postAdjustment` - kind `create`; body type `form`; required record
  fields `propertyID`, `reservationID`, `type`, `amount`; accepted fields `amount`, `itemID`,
  `notes`, `propertyID`, `reservationID`, `type`; risk: posts a manual folio adjustment (credit or
  charge) to a reservation; a real financial transaction - approval required.
- `delete_adjustment`: DELETE `/deleteAdjustment?reservationID={{ record.reservationID
  }}&adjustmentID={{ record.adjustmentID }}` - kind `delete`; body type `none`; path fields
  `reservationID`, `adjustmentID`; required record fields `reservationID`, `adjustmentID`; accepted
  fields `adjustmentID`, `reservationID`; missing records treated as success for status `404`; risk:
  irreversibly removes a folio adjustment; a real financial reversal - approval required.
- `post_item`: POST `/postItem` - kind `create`; body type `form`; required record fields
  `propertyID`, `itemID`, `itemQuantity`; accepted fields `groupCode`, `houseAccountID`, `itemID`,
  `itemNote`, `itemPaid`, `itemPrice`, `itemQuantity`, `propertyID`, `reservationID`, `saleDate`,
  `subReservationID`; risk: posts a catalog item sale to a reservation/house account folio; a real
  financial transaction - approval required.
- `void_item`: POST `/postVoidItem` - kind `delete`; body type `form`; required record fields
  `propertyID`, `soldProductID`; accepted fields `groupCode`, `houseAccountID`, `propertyID`,
  `reservationID`, `soldProductID`; risk: irreversibly voids a previously-sold item line; a real
  financial reversal - approval required.
- `create_item_category`: POST `/postItemCategory` - kind `create`; body type `form`; required
  record fields `propertyID`, `categoryName`; accepted fields `categoryCode`, `categoryColor`,
  `categoryName`, `itemID`, `propertyID`; risk: creates a new item catalog category; no approval
  required.
- `create_group_note`: POST `/postGroupNote` - kind `create`; body type `form`; required record
  fields `propertyID`, `groupCode`, `groupNote`; accepted fields `groupCode`, `groupNote`,
  `propertyID`; risk: adds a free-text note to a group profile; no approval required.
- `update_housekeeping_status`: POST `/postHousekeepingStatus` - kind `update`; body type `form`;
  required record fields `propertyID`, `roomID`; accepted fields `doNotDisturb`, `propertyID`,
  `refusedService`, `roomComments`, `roomCondition`, `roomID`, `vacantPickup`; risk: changes a
  room's live housekeeping/cleaning status; no approval required.
- `create_housekeeper`: POST `/postHousekeeper` - kind `create`; body type `form`; required record
  fields `propertyID`, `name`; accepted fields `name`, `propertyID`; risk: creates a new housekeeper
  staff record; no approval required.
- `update_housekeeper`: PUT `/putHousekeeper` - kind `update`; body type `form`; required record
  fields `propertyID`, `housekeeperID`, `name`; accepted fields `housekeeperID`, `name`,
  `propertyID`; risk: renames an existing housekeeper staff record; no approval required.
- `assign_housekeeping`: POST `/postHousekeepingAssignment` - kind `update`; body type `form`;
  required record fields `propertyID`, `roomIDs`, `housekeeperID`; accepted fields `housekeeperID`,
  `propertyID`, `roomIDs`; risk: assigns rooms to a housekeeper for cleaning; no approval required.
- `create_house_account`: POST `/postNewHouseAccount` - kind `create`; body type `form`; required
  record fields `propertyID`, `accountName`; accepted fields `accountName`, `isPrivate`,
  `propertyID`; risk: creates a new house (folio) account for non-guest billing; no approval
  required.
- `update_house_account_status`: PUT `/putHouseAccountStatus` - kind `update`; body type `form`;
  required record fields `propertyID`, `houseAccountID`, `status`; accepted fields `houseAccountID`,
  `propertyID`, `status`; risk: opens or closes a house account; closing prevents further postings
  against it; no approval required.
- `create_webhook`: POST `/postWebhook` - kind `create`; body type `form`; required record fields
  `propertyID`, `object`, `action`, `endpointUrl`; accepted fields `action`, `endpointUrl`,
  `object`, `propertyID`; risk: subscribes an external endpoint URL to receive Cloudbeds event
  notifications; low-risk external mutation, no approval required.
- `delete_webhook`: DELETE `/deleteWebhook?subscriptionID={{ record.subscriptionID }}` - kind
  `delete`; body type `none`; path fields `subscriptionID`; required record fields `subscriptionID`;
  accepted fields `subscriptionID`; missing records treated as success for status `404`; risk:
  removes a webhook subscription; approval required (stops event delivery to the subscribed
  endpoint).
- `create_room_block`: POST `/postRoomBlock` - kind `create`; body type `form`; required record
  fields `propertyID`, `roomBlockType`, `startDate`, `endDate`, `rooms`; accepted fields `email`,
  `endDate`, `firstName`, `lastName`, `lengthOfHoldInHours`, `phone`, `propertyID`,
  `roomBlockReason`, `roomBlockType`, `rooms`, `startDate`; risk: holds rooms out of general sale
  for a block; no approval required.
- `update_room_block`: PUT `/putRoomBlock` - kind `update`; body type `form`; required record fields
  `propertyID`, `roomBlockID`; accepted fields `email`, `endDate`, `firstName`, `lastName`,
  `lengthOfHoldInHours`, `phone`, `propertyID`, `roomBlockID`, `roomBlockReason`, `rooms`,
  `startDate`; risk: changes an existing room block's held rooms/dates; no approval required.
- `delete_room_block`: POST `/deleteRoomBlock` - kind `delete`; body type `form`; required record
  fields `propertyID`, `roomBlockID`; accepted fields `propertyID`, `roomBlockID`; missing records
  treated as success for status `404`; risk: releases a room block back to general sale;
  irreversible for any rooms already picked up under it - approval required.

## Known limits

- Batch defaults: read_page_size=100.
- API coverage includes 21 stream-backed endpoint group(s), 32 write-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  binary_payload=5, destructive_admin=1, duplicate_of=14, non_data_endpoint=4, out_of_scope=28,
  requires_elevated_scope=11.

---
name: pm-cloudbeds
description: Cloudbeds connector knowledge and safe action guide.
---

# pm-cloudbeds

## Purpose

Reads Cloudbeds guests, hotels, rooms, reservations, transactions, rate plans, room types, items, taxes/fees, sources, groups, house accounts, housekeeping, custom fields, payment methods, webhooks, allotment blocks, and guest/reservation notes, and writes guest, reservation, payment, folio, housekeeping, house-account, webhook, and room-block actions, through the Cloudbeds v1.2 REST API.

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
- property_id
- rate_plans_end_date
- rate_plans_start_date
- api_key (secret)

## ETL Streams

- guests:
  - primary key: guestID
  - fields: dateCreated(), dateModified(), guestEmail(), guestID(), guestName(), isAnonymized(), isMainGuest(), propertyID(), reservationID()
- hotels:
  - primary key: propertyID
  - fields: organizationID(), propertyCurrency(), propertyDescription(), propertyID(), propertyImage(), propertyName(), propertyTimezone()
- rooms:
  - primary key: propertyID
  - fields: propertyID(), roomBlocks()
- reservations:
  - primary key: reservationID
  - fields: adults(), balance(), children(), dateCreated(), dateModified(), endDate(), guestID(), guestName(), origin(), propertyID(), reservationID(), sourceName(), startDate(), status()
- transactions:
  - primary key: transactionID
  - fields: amount(), category(), currency(), description(), guestID(), guestName(), propertyID(), reservationID(), transactionCategory(), transactionCode(), transactionDateTime(), transactionDateTimeUTC(), transactionID(), transactionType()
- rate_plans:
  - primary key: ratePlanID
  - fields: addOns(), baseRate(), daysOfWeek(), derivedType(), derivedValue(), isDerived(), parentRateID(), parentRatePlanID(), parentRatePlanNamePrivate(), parentRatePlanNamePublic(), promoCode(), propertyID(), rateID(), ratePlanID(), ratePlanNamePrivate(), ratePlanNamePublic(), roomRate(), roomRateDetailed(), roomTypeID(), roomTypeName(), roomsAvailable(), totalRate()
- room_types:
  - primary key: roomTypeID
  - fields: adultsIncluded(), childrenIncluded(), isPrivate(), maxGuests(), propertyID(), roomTypeDescription(), roomTypeID(), roomTypeName(), roomTypeNameShort()
- items:
  - primary key: itemID
  - fields: categoryID(), categoryName(), description(), fees(), grandTotal(), itemCode(), itemID(), itemQuantity(), itemType(), name(), price(), priceWithoutFeesAndTaxes(), reorderNeeded(), reorderThreshold(), sku(), stockInventory(), stopSell(), stopSellMet(), taxes(), totalFees(), totalTaxes()
- item_categories:
  - primary key: categoryID
  - fields: categoryCode(), categoryColor(), categoryID(), categoryName()
- taxes_and_fees:
  - primary key: type, feeID, taxID
  - fields: amount(), amountAdult(), amountChild(), amountRateBased(), amountType(), availableFor(), childId(), code(), createdAt(), dateRanges(), expiredAt(), feeID(), feesCharged(), inclusiveOrExclusive(), isDeleted(), kind(), lengthOfStaySettings(), name(), roomTypes(), taxID(), taxesCharged(), type()
- sources:
  - primary key: propertyID, sourceID
  - fields: commission(), fees(), isThirdParty(), paymentCollect(), propertyID(), sourceID(), sourceName(), status(), taxes()
- groups:
  - primary key: groupCode
  - fields: address1(), address2(), city(), commissionType(), contacts(), countryCode(), created(), groupCode(), legalName(), name(), sourceID(), sourceName(), state(), status(), taxDocumentType(), taxIdNumber(), type(), zip()
- house_accounts:
  - primary key: accountID
  - fields: accountID(), accountName(), accountStatus(), isPrivate(), propertyID(), userName()
- housekeepers:
  - primary key: housekeeperID
  - fields: housekeeperID(), name(), propertyID()
- housekeeping_status:
  - primary key: roomID
  - fields: date(), doNotDisturb(), frontdeskStatus(), housekeeper(), housekeeperID(), refusedService(), roomBlocked(), roomComments(), roomCondition(), roomID(), roomName(), roomOccupied(), roomTypeID(), roomTypeName(), vacantPickup()
- custom_fields:
  - primary key: propertyID, shortcode
  - fields: applyTo(), displayed(), isPersonal(), maxCharacters(), name(), propertyID(), required(), shortcode(), type()
- payment_methods:
  - primary key: code
  - fields: cardTypes(), code(), method(), name()
- webhooks:
  - primary key: id
  - fields: event(), id(), key(), owner(), subscriptionData(), subscriptionType()
- allotment_blocks:
  - primary key: allotmentBlockId
  - fields: allotmentBlockCode(), allotmentBlockId(), allotmentBlockName(), allotmentBlockStatus(), allotmentIntervals(), allotmentType(), autoRelease(), bookingCodeUrl(), eventCode(), eventId(), groupCode(), groupId(), isAutoRelease(), propertyID(), ratePlan(), ratePlanId(), rateType(), releaseDate(), releaseScheduleStatus(), releaseScheduleType(), releaseStatus(), reservationsCount(), resources(), roomsHeld(), roomsPickedUp(), roomsRemaining()
- guest_notes:
  - primary key: guestNoteID
  - fields: dateCreated(), dateModified(), guestID(), guestNote(), guestNoteID(), userName()
- reservation_notes:
  - primary key: reservationNoteID
  - fields: dateCreated(), dateModified(), reservationID(), reservationNote(), reservationNoteID(), userName()

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

## Reverse ETL Actions

- create_guest:
  - endpoint: POST /postGuest
  - risk: adds an additional guest to an existing reservation; no approval required
- update_guest:
  - endpoint: PUT /putGuest
  - risk: overwrites a guest's profile fields; no approval required
- create_guest_note:
  - endpoint: POST /postGuestNote
  - risk: adds a free-text note to a guest profile; no approval required
- update_guest_note:
  - endpoint: PUT /putGuestNote
  - risk: overwrites an existing guest note's text; no approval required
- delete_guest_note:
  - endpoint: DELETE /deleteGuestNote?propertyID={{ record.propertyID }}&guestID={{ record.guestID }}&noteID={{ record.noteID }}
  - required fields: propertyID, guestID, noteID
  - risk: irreversibly removes a guest note; approval required
- create_reservation:
  - endpoint: POST /postReservation
  - risk: creates a new, billable reservation (books rooms and may authorize/charge a card depending on paymentMethod); external mutation, no approval required for the booking itself but review payment-capturing config carefully
- update_reservation:
  - endpoint: PUT /putReservation
  - risk: mutates an existing reservation's status/dates/room assignment; a status change to a cancelled/no-show state has real revenue and inventory-release consequences; no approval required (Cloudbeds' own confirmation flow, e.g. sendStatusChangeEmail, is the operator-facing guardrail)
- create_reservation_note:
  - endpoint: POST /postReservationNote
  - risk: adds a free-text note to a reservation; no approval required
- update_reservation_note:
  - endpoint: PUT /putReservationNote
  - risk: overwrites an existing reservation note's text; no approval required
- delete_reservation_note:
  - endpoint: DELETE /deleteReservationNote?propertyID={{ record.propertyID }}&reservationID={{ record.reservationID }}&reservationNoteID={{ record.reservationNoteID }}
  - required fields: propertyID, reservationID, reservationNoteID
  - risk: irreversibly removes a reservation note; approval required
- check_in_room:
  - endpoint: POST /postRoomCheckIn
  - risk: checks a guest into a room, changing reservation/room status; no approval required
- check_out_room:
  - endpoint: POST /postRoomCheckOut
  - risk: checks a guest out of a room, changing reservation/room status and folio; no approval required
- assign_room:
  - endpoint: POST /postRoomAssign
  - risk: reassigns a reservation to a different physical room, optionally repricing; no approval required
- post_payment:
  - endpoint: POST /postPayment
  - risk: records a payment/deposit against a reservation or house account; a real financial transaction — approval required
- void_payment:
  - endpoint: POST /postVoidPayment
  - risk: irreversibly voids a recorded payment; a real financial reversal — approval required
- post_adjustment:
  - endpoint: POST /postAdjustment
  - risk: posts a manual folio adjustment (credit or charge) to a reservation; a real financial transaction — approval required
- delete_adjustment:
  - endpoint: DELETE /deleteAdjustment?reservationID={{ record.reservationID }}&adjustmentID={{ record.adjustmentID }}
  - required fields: reservationID, adjustmentID
  - risk: irreversibly removes a folio adjustment; a real financial reversal — approval required
- post_item:
  - endpoint: POST /postItem
  - risk: posts a catalog item sale to a reservation/house account folio; a real financial transaction — approval required
- void_item:
  - endpoint: POST /postVoidItem
  - risk: irreversibly voids a previously-sold item line; a real financial reversal — approval required
- create_item_category:
  - endpoint: POST /postItemCategory
  - risk: creates a new item catalog category; no approval required
- create_group_note:
  - endpoint: POST /postGroupNote
  - risk: adds a free-text note to a group profile; no approval required
- update_housekeeping_status:
  - endpoint: POST /postHousekeepingStatus
  - risk: changes a room's live housekeeping/cleaning status; no approval required
- create_housekeeper:
  - endpoint: POST /postHousekeeper
  - risk: creates a new housekeeper staff record; no approval required
- update_housekeeper:
  - endpoint: PUT /putHousekeeper
  - risk: renames an existing housekeeper staff record; no approval required
- assign_housekeeping:
  - endpoint: POST /postHousekeepingAssignment
  - risk: assigns rooms to a housekeeper for cleaning; no approval required
- create_house_account:
  - endpoint: POST /postNewHouseAccount
  - risk: creates a new house (folio) account for non-guest billing; no approval required
- update_house_account_status:
  - endpoint: PUT /putHouseAccountStatus
  - risk: opens or closes a house account; closing prevents further postings against it; no approval required
- create_webhook:
  - endpoint: POST /postWebhook
  - risk: subscribes an external endpoint URL to receive Cloudbeds event notifications; low-risk external mutation, no approval required
- delete_webhook:
  - endpoint: DELETE /deleteWebhook?subscriptionID={{ record.subscriptionID }}
  - required fields: subscriptionID
  - risk: removes a webhook subscription; approval required (stops event delivery to the subscribed endpoint)
- create_room_block:
  - endpoint: POST /postRoomBlock
  - risk: holds rooms out of general sale for a block; no approval required
- update_room_block:
  - endpoint: PUT /putRoomBlock
  - risk: changes an existing room block's held rooms/dates; no approval required
- delete_room_block:
  - endpoint: POST /deleteRoomBlock
  - risk: releases a room block back to general sale; irreversible for any rooms already picked up under it — approval required

## Security

- read risk: external Cloudbeds API read of guest, hotel, room, reservation, transaction, rate, inventory, and operational (housekeeping/webhook/notes) data
- write risk: external mutation of Cloudbeds guests, reservations, payments/folio items, house accounts, housekeeping state, groups, webhooks, and room blocks — several actions are real financial transactions (post_payment, void_payment, post_adjustment, post_item, void_item) or irreversible deletions
- approval: financial-transaction and delete-kind actions (post_payment/void_payment/post_adjustment/delete_adjustment/post_item/void_item/delete_guest_note/delete_reservation_note/delete_webhook/delete_room_block) require approval; profile/note/status/assignment/create-kind actions do not
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect cloudbeds
```

### Inspect as structured JSON

```bash
pm connectors inspect cloudbeds --json
```

## Agent Rules

- Run pm connectors inspect cloudbeds before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
- For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.

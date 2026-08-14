# pm connectors inspect cloudbeds

```text
NAME
  pm connectors inspect cloudbeds - Cloudbeds connector manual

SYNOPSIS
  pm connectors inspect cloudbeds
  pm connectors inspect cloudbeds --json
  pm credentials add <name> --connector cloudbeds [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Cloudbeds guests, hotels, rooms, reservations, transactions, rate plans, room types, items, taxes/fees, sources, groups, house accounts, housekeeping, custom fields, payment methods, webhooks, allotment blocks, and guest/reservation notes, and writes guest, reservation, payment, folio, housekeeping, house-account, webhook, and room-block actions, through the Cloudbeds v1.2 REST API.

ICON
  id: pm-sample
  asset: icons/pm-sample.svg
  source: polymetrics
  review_status: polymetrics
  review_url: https://github.com/polymetrics-ai/cli

CAPABILITIES
  check=true catalog=true read=true write=true query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  base_url
  property_id
  rate_plans_end_date
  rate_plans_start_date
  api_key (secret) (required)

ETL STREAMS
  guests:
    primary key: guestID
    fields: dateCreated(string), dateModified(string), guestEmail(string), guestID(string), guestName(string), isAnonymized(boolean), isMainGuest(boolean), propertyID(string), reservationID(string)
  hotels:
    primary key: propertyID
    fields: organizationID(string), propertyCurrency(string), propertyDescription(string), propertyID(string), propertyImage(string), propertyName(string), propertyTimezone(string)
  rooms:
    primary key: propertyID
    fields: propertyID(string), roomBlocks(array)
  reservations:
    primary key: reservationID
    fields: adults(integer), balance(number), children(integer), dateCreated(string), dateModified(string), endDate(string), guestID(string), guestName(string), origin(string), propertyID(string), reservationID(string), sourceName(string), startDate(string), status(string)
  transactions:
    primary key: transactionID
    fields: amount(number), category(string), currency(string), description(string), guestID(string), guestName(string), propertyID(string), reservationID(string), transactionCategory(string), transactionCode(string), transactionDateTime(string), transactionDateTimeUTC(string), transactionID(string), transactionType(string)
  rate_plans:
    primary key: ratePlanID
    fields: addOns(array), baseRate(number), daysOfWeek(object), derivedType(string), derivedValue(number), isDerived(boolean), parentRateID(string), parentRatePlanID(string), parentRatePlanNamePrivate(string), parentRatePlanNamePublic(string), promoCode(string), propertyID(string), rateID(string), ratePlanID(string), ratePlanNamePrivate(string), ratePlanNamePublic(string), roomRate(number), roomRateDetailed(array), roomTypeID(string), roomTypeName(string), roomsAvailable(integer), totalRate(number)
  room_types:
    primary key: roomTypeID
    fields: adultsIncluded(integer), childrenIncluded(integer), isPrivate(boolean), maxGuests(integer), propertyID(string), roomTypeDescription(string), roomTypeID(string), roomTypeName(string), roomTypeNameShort(string)
  items:
    primary key: itemID
    fields: categoryID(string), categoryName(string), description(string), fees(array), grandTotal(number), itemCode(string), itemID(string), itemQuantity(integer), itemType(string), name(string), price(number), priceWithoutFeesAndTaxes(number), reorderNeeded(boolean), reorderThreshold(integer), sku(string), stockInventory(boolean), stopSell(boolean), stopSellMet(boolean), taxes(array), totalFees(number), totalTaxes(number)
  item_categories:
    primary key: categoryID
    fields: categoryCode(string), categoryColor(string), categoryID(string), categoryName(string)
  taxes_and_fees:
    primary key: type, feeID, taxID
    fields: amount(number), amountAdult(number), amountChild(number), amountRateBased(boolean), amountType(string), availableFor(array), childId(string), code(string), createdAt(string), dateRanges(array), expiredAt(string), feeID(string), feesCharged(array), inclusiveOrExclusive(string), isDeleted(boolean), kind(string), lengthOfStaySettings(object), name(string), roomTypes(array), taxID(string), taxesCharged(array), type(string)
  sources:
    primary key: propertyID, sourceID
    fields: commission(number), fees(array), isThirdParty(boolean), paymentCollect(string), propertyID(string), sourceID(string), sourceName(string), status(string), taxes(array)
  groups:
    primary key: groupCode
    fields: address1(string), address2(string), city(string), commissionType(string), contacts(array), countryCode(string), created(string), groupCode(string), legalName(string), name(string), sourceID(string), sourceName(string), state(string), status(string), taxDocumentType(string), taxIdNumber(string), type(string), zip(string)
  house_accounts:
    primary key: accountID
    fields: accountID(string), accountName(string), accountStatus(string), isPrivate(boolean), propertyID(string), userName(string)
  housekeepers:
    primary key: housekeeperID
    fields: housekeeperID(string), name(string), propertyID(string)
  housekeeping_status:
    primary key: roomID
    fields: date(string), doNotDisturb(boolean), frontdeskStatus(string), housekeeper(string), housekeeperID(string), refusedService(boolean), roomBlocked(boolean), roomComments(string), roomCondition(string), roomID(string), roomName(string), roomOccupied(boolean), roomTypeID(string), roomTypeName(string), vacantPickup(boolean)
  custom_fields:
    primary key: propertyID, shortcode
    fields: applyTo(string), displayed(boolean), isPersonal(boolean), maxCharacters(integer), name(string), propertyID(string), required(boolean), shortcode(string), type(string)
  payment_methods:
    primary key: code
    fields: cardTypes(array), code(string), method(string), name(string)
  webhooks:
    primary key: id
    fields: event(object), id(string), key(object), owner(object), subscriptionData(object), subscriptionType(string)
  allotment_blocks:
    primary key: allotmentBlockId
    fields: allotmentBlockCode(string), allotmentBlockId(string), allotmentBlockName(string), allotmentBlockStatus(string), allotmentIntervals(array), allotmentType(string), autoRelease(object), bookingCodeUrl(string), eventCode(string), eventId(string), groupCode(string), groupId(string), isAutoRelease(boolean), propertyID(string), ratePlan(string), ratePlanId(string), rateType(string), releaseDate(string), releaseScheduleStatus(string), releaseScheduleType(string), releaseStatus(string), reservationsCount(integer), resources(array), roomsHeld(integer), roomsPickedUp(integer), roomsRemaining(integer)
  guest_notes:
    primary key: guestNoteID
    fields: dateCreated(string), dateModified(string), guestID(string), guestNote(string), guestNoteID(string), userName(string)
  reservation_notes:
    primary key: reservationNoteID
    fields: dateCreated(string), dateModified(string), reservationID(string), reservationNote(string), reservationNoteID(string), userName(string)

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite

REVERSE ETL ACTIONS
  create_guest:
    endpoint: POST /postGuest
    required fields: propertyID, reservationID
    risk: adds an additional guest to an existing reservation; no approval required
  update_guest:
    endpoint: PUT /putGuest
    required fields: propertyID, guestID
    risk: overwrites a guest's profile fields; no approval required
  create_guest_note:
    endpoint: POST /postGuestNote
    required fields: propertyID, guestID, guestNote
    risk: adds a free-text note to a guest profile; no approval required
  update_guest_note:
    endpoint: PUT /putGuestNote
    required fields: propertyID, guestID, noteID, guestNote
    risk: overwrites an existing guest note's text; no approval required
  delete_guest_note:
    endpoint: DELETE /deleteGuestNote?propertyID={{ record.propertyID }}&guestID={{ record.guestID }}&noteID={{ record.noteID }}
    required fields: propertyID, guestID, noteID
    risk: irreversibly removes a guest note; approval required
  create_reservation:
    endpoint: POST /postReservation
    required fields: propertyID, startDate, endDate, guestFirstName, guestLastName, rooms
    risk: creates a new, billable reservation (books rooms and may authorize/charge a card depending on paymentMethod); external mutation, no approval required for the booking itself but review payment-capturing config carefully
  update_reservation:
    endpoint: PUT /putReservation
    required fields: propertyID, reservationID
    risk: mutates an existing reservation's status/dates/room assignment; a status change to a cancelled/no-show state has real revenue and inventory-release consequences; no approval required (Cloudbeds' own confirmation flow, e.g. sendStatusChangeEmail, is the operator-facing guardrail)
  create_reservation_note:
    endpoint: POST /postReservationNote
    required fields: propertyID, reservationID, reservationNote
    risk: adds a free-text note to a reservation; no approval required
  update_reservation_note:
    endpoint: PUT /putReservationNote
    required fields: propertyID, reservationID, reservationNoteID, reservationNote
    risk: overwrites an existing reservation note's text; no approval required
  delete_reservation_note:
    endpoint: DELETE /deleteReservationNote?propertyID={{ record.propertyID }}&reservationID={{ record.reservationID }}&reservationNoteID={{ record.reservationNoteID }}
    required fields: propertyID, reservationID, reservationNoteID
    risk: irreversibly removes a reservation note; approval required
  check_in_room:
    endpoint: POST /postRoomCheckIn
    required fields: propertyID, reservationID, roomID
    risk: checks a guest into a room, changing reservation/room status; no approval required
  check_out_room:
    endpoint: POST /postRoomCheckOut
    required fields: propertyID, reservationID, roomID
    risk: checks a guest out of a room, changing reservation/room status and folio; no approval required
  assign_room:
    endpoint: POST /postRoomAssign
    required fields: propertyID, reservationID, reservationRoomID, newRoomID
    risk: reassigns a reservation to a different physical room, optionally repricing; no approval required
  post_payment:
    endpoint: POST /postPayment
    required fields: propertyID, type, amount
    risk: records a payment/deposit against a reservation or house account; a real financial transaction — approval required
  void_payment:
    endpoint: POST /postVoidPayment
    required fields: propertyID, paymentID
    risk: irreversibly voids a recorded payment; a real financial reversal — approval required
  post_adjustment:
    endpoint: POST /postAdjustment
    required fields: propertyID, reservationID, type, amount
    risk: posts a manual folio adjustment (credit or charge) to a reservation; a real financial transaction — approval required
  delete_adjustment:
    endpoint: DELETE /deleteAdjustment?reservationID={{ record.reservationID }}&adjustmentID={{ record.adjustmentID }}
    required fields: reservationID, adjustmentID
    risk: irreversibly removes a folio adjustment; a real financial reversal — approval required
  post_item:
    endpoint: POST /postItem
    required fields: propertyID, itemID, itemQuantity
    risk: posts a catalog item sale to a reservation/house account folio; a real financial transaction — approval required
  void_item:
    endpoint: POST /postVoidItem
    required fields: propertyID, soldProductID
    risk: irreversibly voids a previously-sold item line; a real financial reversal — approval required
  create_item_category:
    endpoint: POST /postItemCategory
    required fields: propertyID, categoryName
    risk: creates a new item catalog category; no approval required
  create_group_note:
    endpoint: POST /postGroupNote
    required fields: propertyID, groupCode, groupNote
    risk: adds a free-text note to a group profile; no approval required
  update_housekeeping_status:
    endpoint: POST /postHousekeepingStatus
    required fields: propertyID, roomID
    risk: changes a room's live housekeeping/cleaning status; no approval required
  create_housekeeper:
    endpoint: POST /postHousekeeper
    required fields: propertyID, name
    risk: creates a new housekeeper staff record; no approval required
  update_housekeeper:
    endpoint: PUT /putHousekeeper
    required fields: propertyID, housekeeperID, name
    risk: renames an existing housekeeper staff record; no approval required
  assign_housekeeping:
    endpoint: POST /postHousekeepingAssignment
    required fields: propertyID, roomIDs, housekeeperID
    risk: assigns rooms to a housekeeper for cleaning; no approval required
  create_house_account:
    endpoint: POST /postNewHouseAccount
    required fields: propertyID, accountName
    risk: creates a new house (folio) account for non-guest billing; no approval required
  update_house_account_status:
    endpoint: PUT /putHouseAccountStatus
    required fields: propertyID, houseAccountID, status
    risk: opens or closes a house account; closing prevents further postings against it; no approval required
  create_webhook:
    endpoint: POST /postWebhook
    required fields: propertyID, object, action, endpointUrl
    risk: subscribes an external endpoint URL to receive Cloudbeds event notifications; low-risk external mutation, no approval required
  delete_webhook:
    endpoint: DELETE /deleteWebhook?subscriptionID={{ record.subscriptionID }}
    required fields: subscriptionID
    risk: removes a webhook subscription; approval required (stops event delivery to the subscribed endpoint)
  create_room_block:
    endpoint: POST /postRoomBlock
    required fields: propertyID, roomBlockType, startDate, endDate, rooms
    risk: holds rooms out of general sale for a block; no approval required
  update_room_block:
    endpoint: PUT /putRoomBlock
    required fields: propertyID, roomBlockID
    risk: changes an existing room block's held rooms/dates; no approval required
  delete_room_block:
    endpoint: POST /deleteRoomBlock
    required fields: propertyID, roomBlockID
    risk: releases a room block back to general sale; irreversible for any rooms already picked up under it — approval required

SECURITY
  read risk: external Cloudbeds API read of guest, hotel, room, reservation, transaction, rate, inventory, and operational (housekeeping/webhook/notes) data
  write risk: external mutation of Cloudbeds guests, reservations, payments/folio items, house accounts, housekeeping state, groups, webhooks, and room blocks — several actions are real financial transactions (post_payment, void_payment, post_adjustment, post_item, void_item) or irreversible deletions
  approval: financial-transaction and delete-kind actions (post_payment/void_payment/post_adjustment/delete_adjustment/post_item/void_item/delete_guest_note/delete_reservation_note/delete_webhook/delete_room_block) require approval; profile/note/status/assignment/create-kind actions do not
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect cloudbeds

  # Inspect as structured JSON
  pm connectors inspect cloudbeds --json

AGENT WORKFLOW
  - Run pm connectors inspect cloudbeds before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.
  - For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```

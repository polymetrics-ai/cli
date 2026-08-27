# Discovery — issue 4361

## Immutable starting point

- Linked issue: [#4361](https://github.com/polymetrics-ai/cli/issues/4361)
- Initial base: `origin/main` at `2165619ec8f5f9d4141b491b7a5a64bc460d0c71`
- Worktree: `/Users/karthiksivadas/.treehouse/cli-83d592/58/cli` (verified isolated before branching)
- Branch: `fm/cli-structured-scalar-union-foundation-r1`
- Source provenance preserved by the Batch 2–3 evidence lane, not modified here:
  - Twilio: `https://raw.githubusercontent.com/twilio/twilio-oai/main/spec/json/twilio_api_v2010.json`, SHA-256 `6b6ffccef14cd55fd6d2fd38c3fde68585e1937eead4c099a92c5efe696b882c`
  - Xero: `https://raw.githubusercontent.com/XeroAPI/Xero-OpenAPI/master/xero_accounting.yaml`, SHA-256 `f980dbcfc317b9ea60c5095f1861f9e60fe05e5f8059781c4f2c2a816e918746`

## Ownership boundary and reproduction

- Shared declaration gate: `engine.ValidateStructuredJSONRecordField` currently accepts object/array and multi-non-null scalar unions, but rejects an exact `["string","null"]` union as not object/array.
- Shared source projection: `sourceProjectionFlagType` collapses `["string","null"]` to `string`; that cannot distinguish the literal command text `null` from a JSON null.
- Runtime already has the correct closed mechanics for a `type:"json"` flag without `allow_bare_string`: it decodes exactly one JSON value and validates it through the named record schema before any I/O.
- Parallel operation-body helpers have the same object/array admission rule and are reviewed as the same declaration-bound foundation. This does not create a generic body input: the fixed operation and exact `body.<field>` mapping remain required.

## Source-backed affected rows

All rows are retained as `generated_cli_command, reverse_etl` in the Batch 2–3 ledger. The listed field is the source-cited `["string","null"]` field that raised this foundation gap; operation mapping and certification stay independent.

| # | Connector | Provider source ID | Method | Path | Intended command action | Field |
| --- | --- | --- | --- | --- | --- | --- |
| 1 | twilio | twilio.rest.CreateAddress | POST | /Accounts/{AccountSid}/Addresses.json | create_address | City |
| 2 | twilio | twilio.rest.CreateCall | POST | /Accounts/{AccountSid}/Calls.json | create_call | From |
| 3 | twilio | twilio.rest.CreateIncomingPhoneNumberAssignedAddOn | POST | /Accounts/{AccountSid}/IncomingPhoneNumbers/{ResourceSid}/AssignedAddOns.json | create_incoming_phone_number_assigned_add_on | InstalledAddOnSid |
| 4 | twilio | twilio.rest.CreateIncomingPhoneNumberLocal | POST | /Accounts/{AccountSid}/IncomingPhoneNumbers/Local.json | create_incoming_phone_number_local | PhoneNumber |
| 5 | twilio | twilio.rest.CreateIncomingPhoneNumberMobile | POST | /Accounts/{AccountSid}/IncomingPhoneNumbers/Mobile.json | create_incoming_phone_number_mobile | PhoneNumber |
| 6 | twilio | twilio.rest.CreateIncomingPhoneNumberTollFree | POST | /Accounts/{AccountSid}/IncomingPhoneNumbers/TollFree.json | create_incoming_phone_number_toll_free | PhoneNumber |
| 7 | twilio | twilio.rest.CreateMessage | POST | /Accounts/{AccountSid}/Messages.json | create_message | To |
| 8 | twilio | twilio.rest.CreateParticipant | POST | /Accounts/{AccountSid}/Conferences/{ConferenceSid}/Participants.json | create_participant | From |
| 9 | twilio | twilio.rest.CreatePayments | POST | /Accounts/{AccountSid}/Calls/{CallSid}/Payments.json | create_payments | IdempotencyKey |
| 10 | twilio | twilio.rest.CreateQueue | POST | /Accounts/{AccountSid}/Queues.json | create_queue | FriendlyName |
| 11 | twilio | twilio.rest.CreateSipAuthCallsCredentialListMapping | POST | /Accounts/{AccountSid}/SIP/Domains/{DomainSid}/Auth/Calls/CredentialListMappings.json | create_sip_auth_calls_credential_list_mapping | CredentialListSid |
| 12 | twilio | twilio.rest.CreateSipAuthCallsIpAccessControlListMapping | POST | /Accounts/{AccountSid}/SIP/Domains/{DomainSid}/Auth/Calls/IpAccessControlListMappings.json | create_sip_auth_calls_ip_access_control_list_mapping | IpAccessControlListSid |
| 13 | twilio | twilio.rest.CreateSipAuthRegistrationsCredentialListMapping | POST | /Accounts/{AccountSid}/SIP/Domains/{DomainSid}/Auth/Registrations/CredentialListMappings.json | create_sip_auth_registrations_credential_list_mapping | CredentialListSid |
| 14 | twilio | twilio.rest.CreateSipCredential | POST | /Accounts/{AccountSid}/SIP/CredentialLists/{CredentialListSid}/Credentials.json | create_sip_credential | Password |
| 15 | twilio | twilio.rest.CreateSipCredentialList | POST | /Accounts/{AccountSid}/SIP/CredentialLists.json | create_sip_credential_list | FriendlyName |
| 16 | twilio | twilio.rest.CreateSipCredentialListMapping | POST | /Accounts/{AccountSid}/SIP/Domains/{DomainSid}/CredentialListMappings.json | create_sip_credential_list_mapping | CredentialListSid |
| 17 | twilio | twilio.rest.CreateSipDomain | POST | /Accounts/{AccountSid}/SIP/Domains.json | create_sip_domain | DomainName |
| 18 | twilio | twilio.rest.CreateSipIpAccessControlList | POST | /Accounts/{AccountSid}/SIP/IpAccessControlLists.json | create_sip_ip_access_control_list | FriendlyName |
| 19 | twilio | twilio.rest.CreateSipIpAccessControlListMapping | POST | /Accounts/{AccountSid}/SIP/Domains/{DomainSid}/IpAccessControlListMappings.json | create_sip_ip_access_control_list_mapping | IpAccessControlListSid |
| 20 | twilio | twilio.rest.CreateSipIpAddress | POST | /Accounts/{AccountSid}/SIP/IpAccessControlLists/{IpAccessControlListSid}/IpAddresses.json | create_sip_ip_address | FriendlyName |
| 21 | twilio | twilio.rest.CreateStream | POST | /Accounts/{AccountSid}/Calls/{CallSid}/Streams.json | create_stream | Url |
| 22 | twilio | twilio.rest.CreateUsageTrigger | POST | /Accounts/{AccountSid}/Usage/Triggers.json | create_usage_trigger | CallbackUrl |
| 23 | twilio | twilio.rest.CreateUserDefinedMessage | POST | /Accounts/{AccountSid}/Calls/{CallSid}/UserDefinedMessages.json | create_user_defined_message | Content |
| 24 | twilio | twilio.rest.CreateUserDefinedMessageSubscription | POST | /Accounts/{AccountSid}/Calls/{CallSid}/UserDefinedMessageSubscriptions.json | create_user_defined_message_subscription | Callback |
| 25 | twilio | twilio.rest.CreateValidationRequest | POST | /Accounts/{AccountSid}/OutgoingCallerIds.json | create_outgoing_caller_id_validation_request | PhoneNumber |
| 26 | twilio | twilio.rest.UpdateCallRecording | POST | /Accounts/{AccountSid}/Calls/{CallSid}/Recordings/{Sid}.json | update_call_recording | Status |
| 27 | twilio | twilio.rest.UpdateConferenceRecording | POST | /Accounts/{AccountSid}/Conferences/{ConferenceSid}/Recordings/{Sid}.json | update_conference_recording | Status |
| 28 | twilio | twilio.rest.UpdateMember | POST | /Accounts/{AccountSid}/Queues/{QueueSid}/Members/{CallSid}.json | update_member | Url |
| 29 | twilio | twilio.rest.UpdatePayments | POST | /Accounts/{AccountSid}/Calls/{CallSid}/Payments/{Sid}.json | update_payments | IdempotencyKey |
| 30 | twilio | twilio.rest.UpdateRealtimeTranscription | POST | /Accounts/{AccountSid}/Calls/{CallSid}/Transcriptions/{Sid}.json | update_realtime_transcription | Status |
| 31 | twilio | twilio.rest.UpdateSipCredentialList | POST | /Accounts/{AccountSid}/SIP/CredentialLists/{Sid}.json | update_sip_credential_list | FriendlyName |
| 32 | twilio | twilio.rest.UpdateSipIpAccessControlList | POST | /Accounts/{AccountSid}/SIP/IpAccessControlLists/{Sid}.json | update_sip_ip_access_control_list | FriendlyName |
| 33 | twilio | twilio.rest.UpdateSiprec | POST | /Accounts/{AccountSid}/Calls/{CallSid}/Siprec/{Sid}.json | update_siprec | Status |
| 34 | twilio | twilio.rest.UpdateStream | POST | /Accounts/{AccountSid}/Calls/{CallSid}/Streams/{Sid}.json | update_stream | Status |
| 35 | xero | xero.rest.post.batchpayments_batchpaymentid.15 | POST | /BatchPayments/{BatchPaymentID} | delete_batch_payment_by_url_param | Status |
| 36 | xero | xero.rest.post.batchpayments.13 | POST | /BatchPayments | delete_batch_payment | BatchPaymentID |
| 37 | xero | xero.rest.post.payments_paymentid.149 | POST | /Payments/{PaymentID} | delete_payment | Status |

## Command-surface and lane review

| Lane | Effect of this foundation | Truthful current status |
| --- | --- | --- |
| Direct read | Allows only a fixed operation body field that declares exact `string|null`; no route/method/body escape hatch. | No affected current-main command materialized. |
| Direct write | Same fixed-operation constraint; does not make a provider operation implemented. | No affected current-main command materialized. |
| Binary download | None. | Independent executor/operation contract remains required. |
| Binary upload | Named record JSON fields can use the shared record preflight, but file/source controls remain independent. | No claim changed. |
| ETL | None. | Stream schemas and warehouse materialization remain independent. |
| Reverse ETL | Enables a future generated named strict JSON flag to preserve declared string and null arms. | Batch 2–3 actions remain partial until their separate materialization and approval/execution foundations close. DELETE/reverse-ETL classification is unchanged. |

Current `main` has no Twilio `cli_surface.json`; Xero's current surface has no cited commands. Therefore the direct usable-surface delta is **zero**. The consumer work is the preserved Batch 2–3 lane: rerun its source-backed materialization after this foundation is merged, retain partial/block reasons until its independent reverse-ETL and command reachability gates are green, and then prove credential-free `missing --credential` without provider access.

---
name: pm-whatsapp
description: WhatsApp connector knowledge and safe action guide.
---

# pm-whatsapp

## Purpose

Reads WhatsApp Business Platform (Cloud API + Business Management API) phone numbers, message templates, WhatsApp Business Account (WABA) metadata, subscribed apps, and template analytics; executes bounded direct reads and a typed template-analytics read-query (messaging/conversation/pricing analytics stay ledger-only because Graph exposes them solely as field expansions); and models WhatsApp sends (all message types), template and phone-number administration, business-profile updates, QR/subscription management, and bounded media upload/download as typed reverse-ETL actions. The WhatsApp Web multidevice (whatsmeow) access mode from vicentereig/whatsapp-cli is modeled as a documented, config-scoped op set. Framed for HMS/healthcare patient messaging alongside the Bahmni EMR connector; recipient numbers and message content are treated as patient PHI.

## Icon

- asset: icons/pm-sample.svg
- source: polymetrics
- review_status: polymetrics
- review_url: https://github.com/polymetrics-ai/cli

## Capabilities

- check=true catalog=true read=true write=true query=true
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- base_url
- max_pages
- mode
- page_size
- phone_number_id
- start_date
- store_dir
- waba_id
- access_token (secret)

## ETL Streams

- phone_numbers:
  - primary key: id
  - fields: code_verification_status(), display_phone_number(), id(), name_status(), platform_type(), quality_rating(), throughput(), verified_name()
- message_templates:
  - primary key: id
  - fields: category(), components(), id(), language(), name(), quality_score(), status()
- subscribed_apps:
  - primary key: id
  - fields: id(), link(), name()
- waba:
  - primary key: id
  - fields: account_review_status(), currency(), id(), message_template_namespace(), name(), timezone_id()

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

## Reverse ETL Actions

- send_text_message:
  - endpoint: POST /{{ config.phone_number_id }}/messages
  - required fields: messaging_product, to, type, text
  - risk: sends a WhatsApp text message to a patient number (PHI); requires reverse ETL approval and consent
- send_image_message:
  - endpoint: POST /{{ config.phone_number_id }}/messages
  - required fields: messaging_product, to, type, image
  - risk: sends a WhatsApp image message to a patient number (PHI); requires reverse ETL approval and consent
- send_audio_message:
  - endpoint: POST /{{ config.phone_number_id }}/messages
  - required fields: messaging_product, to, type, audio
  - risk: sends a WhatsApp audio message to a patient number (PHI); requires reverse ETL approval and consent
- send_video_message:
  - endpoint: POST /{{ config.phone_number_id }}/messages
  - required fields: messaging_product, to, type, video
  - risk: sends a WhatsApp video message to a patient number (PHI); requires reverse ETL approval and consent
- send_document_message:
  - endpoint: POST /{{ config.phone_number_id }}/messages
  - required fields: messaging_product, to, type, document
  - risk: sends a WhatsApp document message to a patient number (PHI); requires reverse ETL approval and consent
- send_sticker_message:
  - endpoint: POST /{{ config.phone_number_id }}/messages
  - required fields: messaging_product, to, type, sticker
  - risk: sends a WhatsApp sticker message to a patient number (PHI); requires reverse ETL approval and consent
- send_location_message:
  - endpoint: POST /{{ config.phone_number_id }}/messages
  - required fields: messaging_product, to, type, location
  - risk: sends a WhatsApp location message to a patient number (PHI); requires reverse ETL approval and consent
- send_contacts_message:
  - endpoint: POST /{{ config.phone_number_id }}/messages
  - required fields: messaging_product, to, type, contacts
  - risk: sends a WhatsApp contacts message to a patient number (PHI); requires reverse ETL approval and consent
- send_interactive_message:
  - endpoint: POST /{{ config.phone_number_id }}/messages
  - required fields: messaging_product, to, type, interactive
  - risk: sends a WhatsApp interactive message to a patient number (PHI); requires reverse ETL approval and consent
- send_template_message:
  - endpoint: POST /{{ config.phone_number_id }}/messages
  - required fields: messaging_product, to, type, template
  - risk: sends a WhatsApp template message to a patient number (PHI); requires reverse ETL approval and consent
- send_reaction_message:
  - endpoint: POST /{{ config.phone_number_id }}/messages
  - required fields: messaging_product, to, type, reaction
  - risk: sends a WhatsApp reaction message to a patient number (PHI); requires reverse ETL approval and consent
- mark_message_read:
  - endpoint: POST /{{ config.phone_number_id }}/messages
  - required fields: messaging_product, status, message_id
  - risk: marks an inbound patient message as read
- send_typing_indicator:
  - endpoint: POST /{{ config.phone_number_id }}/messages
  - required fields: messaging_product, status, message_id, typing_indicator
  - risk: sends a typing indicator on an inbound patient conversation
- create_message_template:
  - endpoint: POST /{{ config.waba_id }}/message_templates
  - required fields: name, language, category
  - optional fields: components
  - risk: creates a WhatsApp message template pending Meta review
- update_message_template:
  - endpoint: POST /{{ record.template_id }}
  - required fields: template_id
  - optional fields: category, components
  - risk: edits an existing message template
- delete_message_template:
  - endpoint: DELETE /{{ config.waba_id }}/message_templates?name={{ record.name }}
  - required fields: name
  - risk: permanently deletes a message template and every language version registered under its name
- update_business_profile:
  - endpoint: POST /{{ config.phone_number_id }}/whatsapp_business_profile
  - required fields: messaging_product
  - optional fields: about, address, description, email, vertical, websites
  - risk: updates the WhatsApp business profile for the number
- register_phone_number:
  - endpoint: POST /{{ config.phone_number_id }}/register
  - required fields: messaging_product, pin
  - risk: registers the phone number on Cloud API
- deregister_phone_number:
  - endpoint: POST /{{ config.phone_number_id }}/deregister
  - required fields: messaging_product
  - risk: deregisters the phone number from Cloud API
- request_verification_code:
  - endpoint: POST /{{ config.phone_number_id }}/request_code
  - required fields: code_method
  - optional fields: language
  - risk: requests a phone-number verification code
- verify_phone_number:
  - endpoint: POST /{{ config.phone_number_id }}/verify_code
  - required fields: code
  - risk: submits the phone-number verification code
- set_two_step_pin:
  - endpoint: POST /{{ config.phone_number_id }}
  - required fields: pin
  - risk: sets or changes the two-step verification PIN
- subscribe_waba_app:
  - endpoint: POST /{{ config.waba_id }}/subscribed_apps
  - required fields: messaging_product
  - risk: subscribes the app to WABA webhooks
- unsubscribe_waba_app:
  - endpoint: DELETE /{{ config.waba_id }}/subscribed_apps
  - required fields: messaging_product
  - risk: unsubscribes the app from WABA webhooks
- create_qr_code:
  - endpoint: POST /{{ config.phone_number_id }}/message_qrdls
  - required fields: prefilled_message
  - optional fields: generate_qr_image
  - risk: creates a QR code / short link for the number
- delete_qr_code:
  - endpoint: DELETE /{{ config.phone_number_id }}/message_qrdls/{{ record.code }}
  - required fields: code
  - risk: deletes a QR code / short link
- upload_media:
  - endpoint: POST /{{ config.phone_number_id }}/media
  - required fields: messaging_product, media_file, type
  - risk: uploads bounded media to WhatsApp and returns a media id (patient content; PHI)
- delete_media:
  - endpoint: DELETE /{{ record.media_id }}
  - required fields: media_id
  - risk: permanently deletes an uploaded media object

## Security

- read risk: external Meta Graph API read of WABA metadata, phone numbers, templates, subscribed apps, and template analytics; direct reads are size-bounded and secret-shaped response fields are redacted; media get-url intentionally returns the short-TTL media URL it was asked to resolve
- write risk: typed WhatsApp reverse ETL: patient message sends (PHI), template and phone-number administration, business-profile updates, QR/subscription management, and bounded media upload; message bodies and recipient numbers are redacted in plans
- approval: reverse ETL requires plan, preview, approval, execute; message sends, media upload, destructive deletes, and phone-number registration/PIN actions require --confirm destructive plus healthcare consent/template pre-approval
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Command Surface

- Inspect, read, and safely plan typed WhatsApp patient-messaging operations.
- Usage: pm whatsapp <command> [flags]
- Source CLI: WhatsApp Business Platform + vicentereig/whatsapp-cli (Meta Graph API v25.0; whatsapp-cli v1.3.2 (whatsmeow))
- Global flags:
  - --credential (string): Credential name to use for the WhatsApp request.
  - --connection (string): Alias for --credential.
  - --config (string_array): Connector config override as key=value.
  - --json (boolean): Emit machine-readable JSON.
  - --limit (integer): Bound the number of records for stream reads.
  - --confirm (string): Typed confirmation for destructive/admin reverse-ETL actions.: values=destructive
- Messages
  - messages send-text - Send a WhatsApp text message to a patient number (PHI). [intent=reverse_etl availability=partial write=send_text_message]; approval: reverse ETL plan -> preview -> approval -> execute; PHI redacted in plans; risk: sends a WhatsApp text message to a patient number (PHI); requires reverse ETL approval and consent; notes: Typed reverse-ETL action; no raw Graph method/path/body escape hatch is exposed.; flags: --to
  - messages send-image - Send a WhatsApp image message (PHI). [intent=reverse_etl availability=partial write=send_image_message]; approval: reverse ETL plan -> preview -> approval -> execute; PHI redacted in plans; risk: sends a WhatsApp image message to a patient number (PHI); requires reverse ETL approval and consent; notes: Typed reverse-ETL action; no raw Graph method/path/body escape hatch is exposed.; flags: --to
  - messages send-audio - Send a WhatsApp audio message (PHI). [intent=reverse_etl availability=partial write=send_audio_message]; approval: reverse ETL plan -> preview -> approval -> execute; PHI redacted in plans; risk: sends a WhatsApp audio message to a patient number (PHI); requires reverse ETL approval and consent; notes: Typed reverse-ETL action; no raw Graph method/path/body escape hatch is exposed.; flags: --to
  - messages send-video - Send a WhatsApp video message (PHI). [intent=reverse_etl availability=partial write=send_video_message]; approval: reverse ETL plan -> preview -> approval -> execute; PHI redacted in plans; risk: sends a WhatsApp video message to a patient number (PHI); requires reverse ETL approval and consent; notes: Typed reverse-ETL action; no raw Graph method/path/body escape hatch is exposed.; flags: --to
  - messages send-document - Send a WhatsApp document message, e.g. a lab-result PDF (PHI). [intent=reverse_etl availability=partial write=send_document_message]; approval: reverse ETL plan -> preview -> approval -> execute; PHI redacted in plans; risk: sends a WhatsApp document message to a patient number (PHI); requires reverse ETL approval and consent; notes: Typed reverse-ETL action; no raw Graph method/path/body escape hatch is exposed.; flags: --to
  - messages send-sticker - Send a WhatsApp sticker message. [intent=reverse_etl availability=partial write=send_sticker_message]; approval: reverse ETL plan -> preview -> approval -> execute; PHI redacted in plans; risk: sends a WhatsApp sticker message to a patient number (PHI); requires reverse ETL approval and consent; notes: Typed reverse-ETL action; no raw Graph method/path/body escape hatch is exposed.; flags: --to
  - messages send-location - Send a WhatsApp location message. [intent=reverse_etl availability=partial write=send_location_message]; approval: reverse ETL plan -> preview -> approval -> execute; PHI redacted in plans; risk: sends a WhatsApp location message to a patient number (PHI); requires reverse ETL approval and consent; notes: Typed reverse-ETL action; no raw Graph method/path/body escape hatch is exposed.; flags: --to
  - messages send-contacts - Send a WhatsApp contacts message (top-level contacts array). [intent=reverse_etl availability=partial write=send_contacts_message]; approval: reverse ETL plan -> preview -> approval -> execute; PHI redacted in plans; risk: sends a WhatsApp contacts message to a patient number (PHI); requires reverse ETL approval and consent; notes: Typed reverse-ETL action; no raw Graph method/path/body escape hatch is exposed.; flags: --to
  - messages send-interactive - Send a WhatsApp interactive message (action rows/buttons arrays). [intent=reverse_etl availability=partial write=send_interactive_message]; approval: reverse ETL plan -> preview -> approval -> execute; PHI redacted in plans; risk: sends a WhatsApp interactive message to a patient number (PHI); requires reverse ETL approval and consent; notes: Typed reverse-ETL action; no raw Graph method/path/body escape hatch is exposed.; flags: --to
  - messages send-template - Send a WhatsApp template message (appointment/OTP/utility). [intent=reverse_etl availability=partial write=send_template_message]; approval: reverse ETL plan -> preview -> approval -> execute; PHI redacted in plans; risk: sends a WhatsApp template message to a patient number (PHI); requires reverse ETL approval and consent; notes: Typed reverse-ETL action; no raw Graph method/path/body escape hatch is exposed.; flags: --to
  - messages send-reaction - Send a WhatsApp reaction to a message. [intent=reverse_etl availability=partial write=send_reaction_message]; approval: reverse ETL plan -> preview -> approval -> execute; PHI redacted in plans; risk: sends a WhatsApp reaction message to a patient number (PHI); requires reverse ETL approval and consent; notes: Typed reverse-ETL action; no raw Graph method/path/body escape hatch is exposed.; flags: --to
  - messages mark-read - Mark an inbound patient message as read. [intent=reverse_etl availability=partial write=mark_message_read]; approval: reverse ETL plan -> preview -> approval -> execute; PHI redacted in plans; risk: marks an inbound patient message as read; notes: Typed reverse-ETL action; no raw Graph method/path/body escape hatch is exposed.; flags: --message-id
  - messages typing - Send a typing indicator on an inbound conversation. [intent=reverse_etl availability=partial write=send_typing_indicator]; approval: reverse ETL plan -> preview -> approval -> execute; PHI redacted in plans; risk: sends a typing indicator on an inbound patient conversation; notes: Typed reverse-ETL action; no raw Graph method/path/body escape hatch is exposed.; flags: --message-id
- Templates
  - templates list - List WhatsApp message templates. [intent=etl availability=implemented stream=message_templates]
  - templates get - Retrieve a single message template's detail. [intent=direct_read availability=implemented]; risk: bounded Graph JSON read; response is size-capped and secret-shaped fields are redacted; flags: --template-id
  - templates create - Create a WhatsApp message template pending Meta review. [intent=reverse_etl availability=partial write=create_message_template]; approval: reverse ETL plan -> preview -> approval -> execute; PHI redacted in plans; risk: creates a WhatsApp message template pending Meta review; notes: Typed reverse-ETL action; no raw Graph method/path/body escape hatch is exposed.; flags: --name, --language, --category
  - templates update - Edit an existing message template. [intent=reverse_etl availability=partial write=update_message_template]; approval: reverse ETL plan -> preview -> approval -> execute; PHI redacted in plans; risk: edits an existing message template; notes: Typed reverse-ETL action; no raw Graph method/path/body escape hatch is exposed.; flags: --template-id
  - templates delete - Permanently delete a message template. [intent=reverse_etl availability=partial write=delete_message_template]; approval: reverse ETL plan -> preview -> approval -> execute; --confirm destructive required; PHI (message body + recipient number) redacted in plans; risk: high: permanently deletes a message template; notes: Typed reverse-ETL action; no raw Graph method/path/body escape hatch is exposed.; flags: --name
- Phone numbers
  - phone-numbers list - List WhatsApp phone numbers on the WABA. [intent=etl availability=implemented stream=phone_numbers]
  - phone-numbers get - Retrieve a single phone number's detail. [intent=direct_read availability=implemented]; risk: bounded Graph JSON read; response is size-capped and secret-shaped fields are redacted; flags: --phone-number-id
  - phone-numbers register - Register the phone number on Cloud API. [intent=reverse_etl availability=partial write=register_phone_number]; approval: reverse ETL plan -> preview -> approval -> execute; --confirm destructive required; PHI (message body + recipient number) redacted in plans; risk: high: registers the phone number on Cloud API; notes: Typed reverse-ETL action; no raw Graph method/path/body escape hatch is exposed.; flags: --messaging-product, --pin
  - phone-numbers deregister - Deregister the phone number from Cloud API. [intent=reverse_etl availability=partial write=deregister_phone_number]; approval: reverse ETL plan -> preview -> approval -> execute; --confirm destructive required; PHI (message body + recipient number) redacted in plans; risk: high: deregisters the phone number from Cloud API; notes: Typed reverse-ETL action; no raw Graph method/path/body escape hatch is exposed.; flags: --messaging-product
  - phone-numbers request-code - Request a phone-number verification code. [intent=reverse_etl availability=partial write=request_verification_code]; approval: reverse ETL plan -> preview -> approval -> execute; PHI redacted in plans; risk: requests a phone-number verification code; notes: Typed reverse-ETL action; no raw Graph method/path/body escape hatch is exposed.; flags: --code-method
  - phone-numbers verify-code - Submit the phone-number verification code. [intent=reverse_etl availability=partial write=verify_phone_number]; approval: reverse ETL plan -> preview -> approval -> execute; PHI redacted in plans; risk: submits the phone-number verification code; notes: Typed reverse-ETL action; no raw Graph method/path/body escape hatch is exposed.; flags: --code
  - phone-numbers set-pin - Set or change the two-step verification PIN. [intent=reverse_etl availability=partial write=set_two_step_pin]; approval: reverse ETL plan -> preview -> approval -> execute; --confirm destructive required; PHI (message body + recipient number) redacted in plans; risk: high: sets or changes the two-step verification PIN; notes: Typed reverse-ETL action; no raw Graph method/path/body escape hatch is exposed.; flags: --pin
- Business profile
  - profile get - Retrieve the WhatsApp business profile. [intent=direct_read availability=implemented]; risk: bounded Graph JSON read; response is size-capped and secret-shaped fields are redacted
  - profile update - Update the WhatsApp business profile. [intent=reverse_etl availability=partial write=update_business_profile]; approval: reverse ETL plan -> preview -> approval -> execute; PHI redacted in plans; risk: updates the WhatsApp business profile for the number; notes: Typed reverse-ETL action; no raw Graph method/path/body escape hatch is exposed.; flags: --messaging-product
- Media
  - media get-url - Resolve a media id to a 5-minute TTL media URL. [intent=direct_read availability=implemented]; risk: bounded Graph JSON read; response is size-capped and secret-shaped fields are redacted; flags: --media-id
  - media upload - Upload bounded media to WhatsApp and return a media id (PHI). [intent=reverse_etl availability=partial write=upload_media]; approval: reverse ETL plan -> preview -> approval -> execute; PHI redacted in plans; risk: uploads bounded media to WhatsApp and returns a media id (patient content; PHI); notes: Typed reverse-ETL action; no raw Graph method/path/body escape hatch is exposed.; flags: --media-file, --type
  - media delete - Permanently delete an uploaded media object. [intent=reverse_etl availability=partial write=delete_media]; approval: reverse ETL plan -> preview -> approval -> execute; --confirm destructive required; PHI (message body + recipient number) redacted in plans; risk: high: permanently deletes an uploaded media object; notes: Typed reverse-ETL action; no raw Graph method/path/body escape hatch is exposed.; flags: --media-id
- Analytics
  - analytics messaging - Messaging analytics on the WABA node (not executable in this engine). [intent=direct_read availability=unsupported_api]; notes: Graph exposes messaging analytics only as a field expansion (GET /{waba_id}?fields=analytics.start(..).end(..).granularity(..)). The declarative direct-read engine resolves path variables and flat query parameters, and this connector exposes no raw Graph `fields` flag, so the windowed read is not authorable without a generic escape hatch.
  - analytics conversation - Conversation analytics on the WABA node (not executable in this engine). [intent=direct_read availability=unsupported_api]; notes: Graph exposes conversation analytics only as a field expansion (GET /{waba_id}?fields=conversation_analytics.start(..).end(..).granularity(..)), which is not authorable without a raw Graph `fields` flag.
  - analytics pricing - Pricing analytics on the WABA node (not executable in this engine). [intent=direct_read availability=unsupported_api]; notes: Graph exposes pricing analytics only as a field expansion (GET /{waba_id}?fields=pricing_analytics.start(..).end(..).granularity(..)), which is not authorable without a raw Graph `fields` flag.
  - analytics template - Bounded typed WhatsApp template analytics read-query. [intent=direct_read availability=implemented]; approval: none: read-only analytics query with connector-authored query parameters; risk: bounded typed analytics read-query; response is size-capped and secret/PHI fields are redacted; notes: Executes through the typed operation direct-read engine over the template_analytics edge, which takes flat query parameters; no raw Graph fields string or generic HTTP flag is exposed.; flags: --start, --end, --granularity, --template-ids, --metric-types
- Subscribed apps
  - apps list - List apps subscribed to the WABA webhooks. [intent=etl availability=implemented stream=subscribed_apps]
  - apps subscribe - Subscribe the app to WABA webhooks. [intent=reverse_etl availability=partial write=subscribe_waba_app]; approval: reverse ETL plan -> preview -> approval -> execute; PHI redacted in plans; risk: subscribes the app to WABA webhooks; notes: Typed reverse-ETL action; no raw Graph method/path/body escape hatch is exposed.; flags: --messaging-product
  - apps unsubscribe - Unsubscribe the app from WABA webhooks. [intent=reverse_etl availability=partial write=unsubscribe_waba_app]; approval: reverse ETL plan -> preview -> approval -> execute; --confirm destructive required; PHI (message body + recipient number) redacted in plans; risk: high: unsubscribes the app from WABA webhooks; notes: Typed reverse-ETL action; no raw Graph method/path/body escape hatch is exposed.; flags: --messaging-product
- QR / short links
  - qr create - Create a QR code / short link for the number. [intent=reverse_etl availability=partial write=create_qr_code]; approval: reverse ETL plan -> preview -> approval -> execute; PHI redacted in plans; risk: creates a QR code / short link for the number; notes: Typed reverse-ETL action; no raw Graph method/path/body escape hatch is exposed.; flags: --prefilled-message
  - qr delete - Delete a QR code / short link. [intent=reverse_etl availability=partial write=delete_qr_code]; approval: reverse ETL plan -> preview -> approval -> execute; --confirm destructive required; PHI (message body + recipient number) redacted in plans; risk: high: deletes a QR code / short link; notes: Typed reverse-ETL action; no raw Graph method/path/body escape hatch is exposed.; flags: --code
- WABA metadata
  - waba get - Read WABA metadata as a single ETL record. [intent=etl availability=implemented stream=waba]
- WhatsApp Web (whatsmeow)
  - web sync - WhatsApp Web history backfill + realtime capture + media sync (whatsmeow). [intent=etl availability=unsupported_local unsupported local workflow]; notes: WhatsApp Web multidevice (whatsmeow) op modeled from vicentereig/whatsapp-cli. Requires a local QR session (spec.mode=web, store_dir); not executed by the declarative Graph HTTP engine. Patient content is PHI.
  - web chats list - List WhatsApp Web chats. [intent=etl availability=unsupported_local unsupported local workflow]; notes: WhatsApp Web multidevice (whatsmeow) op modeled from vicentereig/whatsapp-cli. Requires a local QR session (spec.mode=web, store_dir); not executed by the declarative Graph HTTP engine. Patient content is PHI.
  - web messages list - List WhatsApp Web messages for a chat. [intent=etl availability=unsupported_local unsupported local workflow]; notes: WhatsApp Web multidevice (whatsmeow) op modeled from vicentereig/whatsapp-cli. Requires a local QR session (spec.mode=web, store_dir); not executed by the declarative Graph HTTP engine. Patient content is PHI.
  - web contacts search - List/search WhatsApp Web contacts. [intent=etl availability=unsupported_local unsupported local workflow]; notes: WhatsApp Web multidevice (whatsmeow) op modeled from vicentereig/whatsapp-cli. Requires a local QR session (spec.mode=web, store_dir); not executed by the declarative Graph HTTP engine. Patient content is PHI.
  - web messages search - Search WhatsApp Web message content. [intent=direct_read availability=unsupported_local unsupported local workflow]; notes: WhatsApp Web multidevice (whatsmeow) op modeled from vicentereig/whatsapp-cli. Requires a local QR session (spec.mode=web, store_dir); not executed by the declarative Graph HTTP engine. Patient content is PHI.
  - web chats query - Parameterized WhatsApp Web chat lookup. [intent=direct_read availability=unsupported_local unsupported local workflow]; notes: WhatsApp Web multidevice (whatsmeow) op modeled from vicentereig/whatsapp-cli. Requires a local QR session (spec.mode=web, store_dir); not executed by the declarative Graph HTTP engine. Patient content is PHI.
  - web send - Send a WhatsApp Web text message (whatsmeow; PHI). [intent=reverse_etl availability=unsupported_local unsupported local workflow]; notes: WhatsApp Web multidevice (whatsmeow) op modeled from vicentereig/whatsapp-cli. Requires a local QR session (spec.mode=web, store_dir); not executed by the declarative Graph HTTP engine. Patient content is PHI.
  - web media download - Download a WhatsApp Web attachment (bounded local output). [intent=direct_read availability=unsupported_local unsupported local workflow]; notes: WhatsApp Web multidevice (whatsmeow) op modeled from vicentereig/whatsapp-cli. Requires a local QR session (spec.mode=web, store_dir); not executed by the declarative Graph HTTP engine. Patient content is PHI.
  - web auth - WhatsApp Web QR login and local session lifecycle. [intent=auth availability=unsupported_local unsupported local workflow]; notes: WhatsApp Web multidevice (whatsmeow) op modeled from vicentereig/whatsapp-cli. Requires a local QR session (spec.mode=web, store_dir); not executed by the declarative Graph HTTP engine. Patient content is PHI.
- Help topics:
  - whatsapp-auth - Cloud mode uses a Graph access_token via credentials (Bearer); web mode uses a local whatsmeow QR session and store_dir. Never pass secrets in command text; the two modes' credentials are never conflated.
  - whatsapp-phi - Recipient numbers and message bodies are patient PHI; they are redacted by default in plans and never printed at inspection. Sends require consent and template pre-approval.
  - whatsapp-tiers - Cloud API messaging tiers, template categories (utility/marketing/authentication), and Graph rate limits bound send throughput; the connector retries 429s with backoff.

## Commands

### Inspect as a manual

```bash
pm connectors inspect whatsapp
```

### Inspect as structured JSON

```bash
pm connectors inspect whatsapp --json
```

## Agent Rules

- Run pm connectors inspect whatsapp before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
- For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.

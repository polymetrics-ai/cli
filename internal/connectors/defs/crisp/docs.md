# Overview

Reads bounded pages and individual resources from Crisp conversations through the Crisp REST API
V1. Wave 1 exposes 21 provider-owned conversation read operations through individually reachable
commands. It does not make provider write operations executable.

The provider inventory is based on Crisp's [official REST API reference](https://docs.crisp.chat/references/rest-api/v1/)
and its [official Postman Collection v2.1.0](https://docs.crisp.chat/static/data/collections/rest-api-v1.postman),
retrieved 2026-08-05 (ETag `6a6b6306-32de7`). The ledger contains all 234 documented
operations: 21 Wave 1 reads are stream-backed; 213 retain an explicit blocked disposition.

## Auth setup

Configure a Crisp identifier and key through an environment-backed credential profile or standard
input. Both fields are secret values and must not be passed as command flags or configuration
overrides.

Required connection configuration:

- `website_id` (string): target Crisp website identifier.
- `identifier` (secret string): Crisp token identifier for HTTP Basic authentication.
- `key` (secret string): Crisp token key for HTTP Basic authentication.
- `token_tier` (string): Crisp token tier sent as `X-Crisp-Tier`; defaults to `website` and also
  accepts `plugin` or `user` when the configured credential is issued for that tier.
- `page_number` (string): positive page number for page-addressed provider resources; defaults to
  `1`. Wave 1 page-addressed commands reject values below `1` supplied with `--page-number` or
  present in their effective credential/profile configuration before dispatch.
- `base_url` (URI): Crisp API base URL; defaults to `https://api.crisp.chat` and can be overridden
  only for test fixtures or an approved proxy.

## Streams notes

Each command performs one bounded provider request. Page-addressed resources use `page_number`;
the `conversations list` command additionally accepts Crisp's named filter and ordering flags and
enforces the provider's `per_page` range of 20 through 50.

Wave 1 streams are:

- `list_conversations`, `suggested_conversation_segments`, `suggested_conversation_data`,
  `spam_conversations`, and `spam_conversation_content`.
- `conversation`, `conversation_messages`, `conversation_message`, `conversation_routing`,
  `conversation_meta`, and `conversation_original_message`.
- `conversation_pages`, `conversation_events`, `conversation_files`, `conversation_state`,
  `conversation_relations`, `conversation_participants`, `conversation_block_status`,
  `conversation_verify_status`, `conversation_browsing`, and `conversation_call`.

Provider responses are projected from their `data` envelope without connector-defined field
redaction. No generic raw API or arbitrary query escape hatch is exposed.

## Write actions & risks

No Crisp write action is declared executable in this wave. All 129 documented provider writes
remain explicitly blocked on later connector-local implementation waves; destructive write paths
must use the shared plan, preview, explicit approval, and confirm-gate flow when they are added.

## Known limits

- The provider ledger contains 234 operations. This wave makes 21 reads executable; the remaining
  213 operations are individually dispositioned in `api_surface.json`.
- Fourteen provider HEAD checks are blocked on a shared typed HEAD direct-read/check executor.
- `GET /v1/media/animation/list/{page_number}` is blocked on provider RTM result correlation.
- `POST /v1/bucket/url/generate` is blocked on provider RTM correlation plus a bounded signed-upload
  workflow.
- Other read and write operations are deferred to later Crisp connector waves; they are not exposed
  as commands in this release.

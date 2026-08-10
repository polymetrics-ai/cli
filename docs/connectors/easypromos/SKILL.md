---
name: pm-easypromos
description: Easypromos connector knowledge and safe action guide.
---

# pm-easypromos

## Purpose

Reads Easypromos promotions, organizing brands, stages, users, participations, and prizes through the Easypromos REST API.

## Icon

- id: pm-sample
- asset: icons/pm-sample.svg
- source: polymetrics
- review_status: polymetrics
- review_url: https://github.com/polymetrics-ai/cli

## Capabilities

- check=true catalog=true read=true write=false query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- base_url
- mode
- promotion_id
- bearer_token (secret)

## ETL Streams

- promotions:
  - primary key: id
  - fields: created(string), default_language(string), description(string), end_date(string), id(string), organizing_brand_id(string), organizing_brand_name(string), promotion_type(string), start_date(string), status(string), timezone(string), title(string), url(string)
- organizing_brands:
  - primary key: id
  - fields: id(string), name(string)
- stages:
  - primary key: id
  - fields: end_date(string), id(string), name(string), start_date(string), type(string), visible(boolean)
- users:
  - primary key: id
  - fields: country(string), created(string), email(string), external_id(string), first_name(string), id(string), language(string), last_name(string), login_type(string), nickname(string), promotion_id(string), status(string)
- participations:
  - primary key: id
  - fields: created(string), id(string), ip(string), promotion_id(string), stage_id(string), user_agent(string), user_id(string)
- prizes:
  - primary key: id
  - fields: code(string), created(string), download_url(string), id(string), participation_id(string), prize_type_id(string), prize_type_name(string), redeem_url(string), stage_id(string)

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

## Security

- read risk: external Easypromos API read of promotion, user, participation, and prize data
- approval: none; read-only, no reverse-ETL writes
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Command Surface

- Run Easypromos's declared streams and reverse-ETL actions.
- Usage: pm easypromos <command> [flags]
- Read streams
- Reverse ETL writes
- Other Commands
  - api delete users segments promotion-id - Documented DELETE /users/segments/{promotion_id} (not implemented) [intent=direct_write availability=not_implemented operation=easypromos.delete.users-segments-promotion-id]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api get coin-transactions promotion-id - Documented GET /coin_transactions/{promotion_id} (not implemented) [intent=direct_read availability=not_implemented operation=easypromos.get.coin-transactions-promotion-id]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get coin-transactions promotion-id users user-id - Documented GET /coin_transactions/{promotion_id}/users/{user_id} (not implemented) [intent=direct_read availability=not_implemented operation=easypromos.get.coin-transactions-promotion-id-users-user-id]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get coins promotion-id users user-id - Documented GET /coins/{promotion_id}/users/{user_id} (not implemented) [intent=direct_read availability=not_implemented operation=easypromos.get.coins-promotion-id-users-user-id]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get organizing-brands organizing-brand-id - Documented GET /organizing_brands/{organizing_brand_id} (not implemented) [intent=direct_read availability=not_implemented operation=easypromos.get.organizing-brands-organizing-brand-id]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get organizing-brands organizing-brand-id promotions - Documented GET /organizing_brands/{organizing_brand_id}/promotions (not implemented) [intent=direct_read availability=not_implemented operation=easypromos.get.organizing-brands-organizing-brand-id-promotions]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get participations promotion-id participation-id - Documented GET /participations/{promotion_id}/{participation_id} (not implemented) [intent=direct_read availability=not_implemented operation=easypromos.get.participations-promotion-id-participation-id]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get participations promotion-id remaining stage-id user-id - Documented GET /participations/{promotion_id}/remaining/{stage_id}/{user_id} (not implemented) [intent=direct_read availability=not_implemented operation=easypromos.get.participations-promotion-id-remaining-stage-id-user-id]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get participations promotion-id users user-id - Documented GET /participations/{promotion_id}/users/{user_id} (not implemented) [intent=direct_read availability=not_implemented operation=easypromos.get.participations-promotion-id-users-user-id]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get pos promotion-id - Documented GET /pos/{promotion_id} (not implemented) [intent=direct_read availability=not_implemented operation=easypromos.get.pos-promotion-id]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get prizes inventory promotion-id - Documented GET /prizes/inventory/{promotion_id} (not implemented) [intent=direct_read availability=not_implemented operation=easypromos.get.prizes-inventory-promotion-id]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get prizes promotion-id users user-id - Documented GET /prizes/{promotion_id}/users/{user_id} (not implemented) [intent=direct_read availability=not_implemented operation=easypromos.get.prizes-promotion-id-users-user-id]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get promotions promotion-id - Documented GET /promotions/{promotion_id} (not implemented) [intent=direct_read availability=not_implemented operation=easypromos.get.promotions-promotion-id]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get promotions promotion-id registration-fields - Documented GET /promotions/{promotion_id}/registration_fields (not implemented) [intent=direct_read availability=not_implemented operation=easypromos.get.promotions-promotion-id-registration-fields]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get ranking promotion-id - Documented GET /ranking/{promotion_id} (not implemented) [intent=direct_read availability=not_implemented operation=easypromos.get.ranking-promotion-id]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get stages promotion-id stage-id - Documented GET /stages/{promotion_id}/{stage_id} (not implemented) [intent=direct_read availability=not_implemented operation=easypromos.get.stages-promotion-id-stage-id]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get users logintoken promotion-id user-id - Documented GET /users/logintoken/{promotion_id}/{user_id} (not implemented) [intent=direct_read availability=not_implemented operation=easypromos.get.users-logintoken-promotion-id-user-id]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get users promotion-id user-id - Documented GET /users/{promotion_id}/{user_id} (not implemented) [intent=direct_read availability=not_implemented operation=easypromos.get.users-promotion-id-user-id]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api post coin-transactions promotion-id - Documented POST /coin_transactions/{promotion_id} (not implemented) [intent=direct_write availability=not_implemented operation=easypromos.post.coin-transactions-promotion-id]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api post participations promotion-id - Documented POST /participations/{promotion_id} (not implemented) [intent=direct_write availability=not_implemented operation=easypromos.post.participations-promotion-id]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api post participations promotion-id check-requirement stage-id - Documented POST /participations/{promotion_id}/check_requirement/{stage_id} (not implemented) [intent=direct_write availability=not_implemented operation=easypromos.post.participations-promotion-id-check-requirement-stage-id]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api post participations promotion-id participate stage-id - Documented POST /participations/{promotion_id}/participate/{stage_id} (not implemented) [intent=direct_write availability=not_implemented operation=easypromos.post.participations-promotion-id-participate-stage-id]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api post participations promotion-id validate-code stage-id - Documented POST /participations/{promotion_id}/validate_code/{stage_id} (not implemented) [intent=direct_write availability=not_implemented operation=easypromos.post.participations-promotion-id-validate-code-stage-id]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api post users autologin promotion-id - Documented POST /users/autologin/{promotion_id} (not implemented) [intent=direct_write availability=not_implemented operation=easypromos.post.users-autologin-promotion-id]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api post users segments promotion-id - Documented POST /users/segments/{promotion_id} (not implemented) [intent=direct_write availability=not_implemented operation=easypromos.post.users-segments-promotion-id]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api put coins promotion-id - Documented PUT /coins/{promotion_id} (not implemented) [intent=direct_write availability=not_implemented operation=easypromos.put.coins-promotion-id]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - organizing brands list - Run the organizing brands ETL stream [intent=etl availability=implemented stream=organizing_brands]
  - participations list - Run the participations ETL stream [intent=etl availability=implemented stream=participations]
  - prizes list - Run the prizes ETL stream [intent=etl availability=implemented stream=prizes]
  - promotions list - Run the promotions ETL stream [intent=etl availability=implemented stream=promotions]
  - stages list - Run the stages ETL stream [intent=etl availability=implemented stream=stages]
  - users list - Run the users ETL stream [intent=etl availability=implemented stream=users]

## Commands

### Inspect as a manual

```bash
pm connectors inspect easypromos
```

### Inspect as structured JSON

```bash
pm connectors inspect easypromos --json
```

## Agent Rules

- Run pm connectors inspect easypromos before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.

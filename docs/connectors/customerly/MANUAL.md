# pm connectors inspect customerly

```text
NAME
  pm connectors inspect customerly - Customerly connector manual

SYNOPSIS
  pm connectors inspect customerly
  pm connectors inspect customerly --json
  pm credentials add <name> --connector customerly [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Customerly users, leads, and accounts, and writes user/lead/tag/message/attribute/company mutations through the Customerly v1 REST API.

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
  mode
  page_size
  api_key (secret)

ETL STREAMS
  users:
    primary key: user_id, email
    cursor: last_update
    fields: city(string), country(string), create_date(string), crmhero_user_id(string), email(string), first_seen_at(string), last_activity(string), last_update(string), name(string), role(string), sub_active(boolean), sub_status(string), timezone(string), user_id(integer), username(string)
  leads:
    primary key: crmhero_user_id
    cursor: last_update
    fields: city(string), country(string), create_date(string), crmhero_user_id(string), email(string), last_update(string), name(string), role(string), sub_active(boolean), sub_status(string), timezone(string), username(string)
  accounts:
    primary key: account_id
    fields: account_id(integer), email(string)

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

REVERSE ETL ACTIONS
  delete_user:
    endpoint: DELETE /users?user_id={{ record.user_id }}
    required fields: user_id
    risk: external mutation; irreversibly deletes a live Customerly user and every conversation/survey/campaign record tied to them; approval required
  delete_lead:
    endpoint: DELETE /leads?email={{ record.email }}
    required fields: email
    risk: external mutation; irreversibly deletes a live Customerly lead and every associated record; approval required
  unsubscribe_user:
    endpoint: POST /users/unsubscribe/{{ record.user_id }}
    required fields: user_id
    risk: external mutation; unsubscribes a live user from Customerly messaging; approval required
  add_tag:
    endpoint: POST /tags
    required fields: tag
    risk: external mutation; adds or removes a tag on one or more live users/leads
  delete_tag:
    endpoint: DELETE /tags
    required fields: tag
    risk: external mutation; permanently removes a tag definition from the app; it is un-applied from every contact that carried it; approval required
  send_message:
    endpoint: POST /messages
    required fields: from, to, content
    risk: sends a user-visible message from Customerly on the sender's behalf; may notify the recipient
  add_user_attributes:
    endpoint: POST /users/add-attributes/{{ record.user_id }}
    required fields: user_id, attributes
    risk: external mutation; adds/overwrites custom attribute values on a live user
  add_company_attributes:
    endpoint: POST /company/add-attributes/{{ record.company_id }}
    required fields: company_id
    risk: external mutation; adds/overwrites custom attribute values (and optionally renames) a live company
  add_user_to_company:
    endpoint: POST /users/add-to-company
    required fields: company_id
    risk: external mutation; links a live user to a company, creating the company if it does not already exist

SECURITY
  read risk: external Customerly API read of user, lead, and account contact data
  write risk: external mutation of live Customerly users/leads/tags/messages/attributes/companies, including irreversible user and lead deletion; approval required
  approval: read: none; write: required
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

COMMAND SURFACE
  Run Customerly's declared streams and reverse-ETL actions.
  Usage: pm customerly <command> [flags]
  Read streams
  Reverse ETL writes
  Other Commands
    accounts list - Run the accounts ETL stream [intent=etl availability=implemented stream=accounts]
    add company attributes apply - Plan and execute the add company attributes reverse-ETL action [intent=reverse_etl availability=implemented write=add_company_attributes]; approval: requires plan, preview, approval, and execute; risk: external mutation; adds/overwrites custom attribute values (and optionally renames) a live company; flags: --company_id (required)
    add tag apply - Plan and execute the add tag reverse-ETL action [intent=reverse_etl availability=implemented write=add_tag]; approval: requires plan, preview, approval, and execute; risk: external mutation; adds or removes a tag on one or more live users/leads; flags: --tag (required)
    add user attributes apply - Plan and execute the add user attributes reverse-ETL action [intent=reverse_etl availability=not_implemented write=add_user_attributes]; approval: requires plan, preview, approval, and execute; risk: external mutation; adds/overwrites custom attribute values on a live user; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    add user to company apply - Plan and execute the add user to company reverse-ETL action [intent=reverse_etl availability=implemented write=add_user_to_company]; approval: requires plan, preview, approval, and execute; risk: external mutation; links a live user to a company, creating the company if it does not already exist; flags: --company_id (required)
    api delete v1 knowledge articles knowledge-base-article-id - Documented DELETE /v1/knowledge/articles/{knowledge_base_article_id} (not implemented) [intent=direct_write availability=not_implemented operation=customerly.delete.v1-knowledge-articles-knowledge-base-article-id]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api delete v1 knowledge collections knowledge-base-collection-id - Documented DELETE /v1/knowledge/collections/{knowledge_base_collection_id} (not implemented) [intent=direct_write availability=not_implemented operation=customerly.delete.v1-knowledge-collections-knowledge-base-collection-id]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api delete v1 users-email email - Documented DELETE /v1/users?email={email} (not implemented) [intent=direct_write availability=not_implemented operation=customerly.delete.v1-users-email-email]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api get v1 knowledge articles - Documented GET /v1/knowledge/articles (not implemented) [intent=direct_read availability=not_implemented operation=customerly.get.v1-knowledge-articles]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get v1 knowledge articles knowledge-base-article-id - Documented GET /v1/knowledge/articles/{knowledge_base_article_id} (not implemented) [intent=direct_read availability=not_implemented operation=customerly.get.v1-knowledge-articles-knowledge-base-article-id]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get v1 knowledge collections - Documented GET /v1/knowledge/collections (not implemented) [intent=direct_read availability=not_implemented operation=customerly.get.v1-knowledge-collections]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get v1 knowledge collections knowledge-base-collection-id - Documented GET /v1/knowledge/collections/{knowledge_base_collection_id} (not implemented) [intent=direct_read availability=not_implemented operation=customerly.get.v1-knowledge-collections-knowledge-base-collection-id]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get v1 knowledge writers - Documented GET /v1/knowledge/writers (not implemented) [intent=direct_read availability=not_implemented operation=customerly.get.v1-knowledge-writers]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get v1 leads-email email - Documented GET /v1/leads?email={email} (not implemented) [intent=direct_read availability=not_implemented operation=customerly.get.v1-leads-email-email]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get v1 tags - Documented GET /v1/tags (not implemented) [intent=direct_read availability=not_implemented operation=customerly.get.v1-tags]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get v1 users-email email - Documented GET /v1/users?email={email} (not implemented) [intent=direct_read availability=not_implemented operation=customerly.get.v1-users-email-email]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get v1 users-user-id user-id - Documented GET /v1/users?user_id={user_id} (not implemented) [intent=direct_read availability=not_implemented operation=customerly.get.v1-users-user-id-user-id]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api post v1 knowledge articles - Documented POST /v1/knowledge/articles (not implemented) [intent=direct_write availability=not_implemented operation=customerly.post.v1-knowledge-articles]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post v1 knowledge articles knowledge-base-article-id - Documented POST /v1/knowledge/articles/{knowledge_base_article_id} (not implemented) [intent=direct_write availability=not_implemented operation=customerly.post.v1-knowledge-articles-knowledge-base-article-id]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post v1 knowledge collections - Documented POST /v1/knowledge/collections (not implemented) [intent=direct_write availability=not_implemented operation=customerly.post.v1-knowledge-collections]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post v1 knowledge collections knowledge-base-collection-id - Documented POST /v1/knowledge/collections/{knowledge_base_collection_id} (not implemented) [intent=direct_write availability=not_implemented operation=customerly.post.v1-knowledge-collections-knowledge-base-collection-id]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post v1 leads - Documented POST /v1/leads (not implemented) [intent=direct_write availability=not_implemented operation=customerly.post.v1-leads]; approval: not implemented: the REST write executor lacks the provider-specific top-level body envelope required by this operation; risk: high; notes: named_dependency=engine.rest_write_body_envelope: the REST write executor lacks the provider-specific top-level body envelope required by this operation
    api post v1 users - Documented POST /v1/users (not implemented) [intent=direct_write availability=not_implemented operation=customerly.post.v1-users]; approval: not implemented: the REST write executor lacks the provider-specific top-level body envelope required by this operation; risk: high; notes: named_dependency=engine.rest_write_body_envelope: the REST write executor lacks the provider-specific top-level body envelope required by this operation
    delete lead apply - Plan and execute the delete lead reverse-ETL action [intent=reverse_etl availability=implemented write=delete_lead]; approval: requires plan, preview, approval, and execute; risk: external mutation; irreversibly deletes a live Customerly lead and every associated record; approval required; flags: --email (required)
    delete tag apply - Plan and execute the delete tag reverse-ETL action [intent=reverse_etl availability=implemented write=delete_tag]; approval: requires plan, preview, approval, and execute; risk: external mutation; permanently removes a tag definition from the app; it is un-applied from every contact that carried it; approval required; flags: --tag (required)
    delete user apply - Plan and execute the delete user reverse-ETL action [intent=reverse_etl availability=implemented write=delete_user]; approval: requires plan, preview, approval, and execute; risk: external mutation; irreversibly deletes a live Customerly user and every conversation/survey/campaign record tied to them; approval required; flags: --user_id (required)
    leads list - Run the leads ETL stream [intent=etl availability=implemented stream=leads]
    send message apply - Plan and execute the send message reverse-ETL action [intent=reverse_etl availability=not_implemented write=send_message]; approval: requires plan, preview, approval, and execute; risk: sends a user-visible message from Customerly on the sender's behalf; may notify the recipient; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    unsubscribe user apply - Plan and execute the unsubscribe user reverse-ETL action [intent=reverse_etl availability=implemented write=unsubscribe_user]; approval: requires plan, preview, approval, and execute; risk: external mutation; unsubscribes a live user from Customerly messaging; approval required; flags: --user_id (required)
    users list - Run the users ETL stream [intent=etl availability=implemented stream=users]

EXAMPLES
  # Inspect as a manual
  pm connectors inspect customerly

  # Inspect as structured JSON
  pm connectors inspect customerly --json

AGENT WORKFLOW
  - Run pm connectors inspect customerly before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.
  - For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```

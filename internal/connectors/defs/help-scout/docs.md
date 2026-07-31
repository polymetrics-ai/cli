# Overview

Help Scout connector parity for the official Inbox/Mailbox API endpoint set. The connector uses the Help Scout API origin in `base_url` (default `https://api.helpscout.net`) and explicit `/v2` or `/v3` operation paths.

Operation ledger reviewed from the official endpoint index on 2026-07-31: 144 canonical operations. Executable connector-local coverage is 45 ETL streams and 65 reverse-ETL write actions. The remaining 34 operations are represented as blocked/planned rows: 32 report/provider-query reads behind #2985 and 2 attachment binary downloads behind a bounded binary/file foundation.

No live provider calls, credentials, writes, or certification were used to author this bundle.

## Auth setup

Connection fields:

- `client_id` (required, secret): Help Scout OAuth2 application client id.
- `client_secret` (required, secret): Help Scout OAuth2 application client secret.
- `base_url` (optional): API origin, default `https://api.helpscout.net`.
- `token_url` (optional): OAuth2 token endpoint, default `https://api.helpscout.net/v2/oauth2/token`.
- `start_date` (optional): RFC3339 lower bound for list streams that support `modifiedSince`.
- Parameterized streams use optional typed config identifiers such as `conversation_id`, `customer_id`, `mailbox_id`, `user_id`, `webhook_id`, and `property_slug`.

Secrets are declared with `x-secret` and must be supplied from credential storage, environment, or stdin flows. Never place secret values in task text, docs, issue bodies, or fixtures.

Connection checks call `GET /v2/mailboxes`.

## Streams notes

All stream requests are bounded by page-number pagination (`page`, `size`, page size 50, max pages 20). Representative streams include: conversations, customers, get_address, get_conversation, get_conversation_v3, get_customer, get_thread_original_source, list_chats_handles, list_customers_v3, list_emails, list_threads, list_threads_v3. Additional parameterized streams cover single-resource and nested list endpoints such as conversation threads, customer emails/phones/websites, inbox folders/fields/saved replies, organizations, teams, tags, users, webhooks, workflows, and v3 system-user resources.

Report endpoints under `/v2/reports/**` are not exposed as ETL streams in this slice because they are provider-query/direct-read surfaces. They remain blocked/planned until #2985 supplies the shared typed bounded provider-query contract distinct from warehouse `pm query`.

Attachment binary download endpoints (`GET /v2/conversations/{conversationId}/attachments/{attachmentId}/file` and `/data`) remain blocked/planned until a bounded binary/file download contract defines destination paths, byte limits, overwrite policy, and redaction behavior.

## Write actions & risks

The bundle declares 65 Help Scout reverse-ETL actions, including create/update/delete operations for conversations, threads, customers, contact methods, inbox saved replies/routing, organizations/properties, users, teams, workflows, webhooks, and attachment upload/delete. Representative actions: upload_attachment, delete_attachment, create_conversation, update_custom_fields, delete_conversation, delete_snooze, update_snooze, update_tags, create_chat_thread, create_customer_thread, create_note, create_phone_thread.

Every Help Scout write action is conservative by policy:

- `confirm: "destructive"` is required, including admin and create/update actions.
- Reverse ETL must follow plan -> preview -> explicit approval token -> execute.
- DELETE actions are idempotent with `missing_ok_status: [404]` and redact path identifiers in previews/errors.
- Attachment upload uses a multipart action with a project-local file path, byte cap, preview approval, and payload digest support from the existing engine.
- No generic HTTP method/path/body, raw provider query, shell, or file write escape hatch is exposed.

## Known limits

- Dynamic conformance is skipped because OAuth2 client-credentials token exchange cannot run with synthetic credentials; static validation, sanitized fixtures, and Help Scout-owned full-surface tests provide local evidence.
- `metadata.capabilities.cdc` remains `false`. Webhook CRUD endpoints are typed admin reverse-ETL/configuration actions and webhook list/get streams; this is not a CDC/changefeed executor or checkpoint/resume contract. CDC truth/state foundations #2986 and #2988 remain external dependencies.
- Report/provider-query reads are blocked/planned behind #2985.
- Attachment binary downloads are blocked/planned until a bounded binary/file download foundation is available. Attachment upload/delete are reverse-ETL actions with typed confirmation.
- Write schemas are intentionally bounded to reviewed path fields plus provider body payloads supplied by reverse-ETL records; unsupported provider validation errors are surfaced without logging secrets.
- Fixture pages are sanitized synthetic shapes and are not live certification evidence.

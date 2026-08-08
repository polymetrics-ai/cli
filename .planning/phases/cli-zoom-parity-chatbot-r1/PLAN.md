# Plan — Zoom Chatbot documented-operation parity, R1

## Delivery record

- Parent: [#3915](https://github.com/polymetrics-ai/cli/issues/3915); provider-owned slice:
  [#3944](https://github.com/polymetrics-ai/cli/issues/3944).
- Scope: Zoom's **Chatbot** module only: its four documented mutation operations, their exact
  typed direct-write CLI routes, declared client-credentials transport, generated Zoom docs/site
  catalog output, focused fixtures/tests, reusable foundations required by those contracts, and
  this phase evidence.
- Required skills carried by the parent delivery: `golang-how-to`, `golang-cli`,
  `golang-testing`, `golang-error-handling`, `golang-security`, `golang-safety`,
  `golang-design-patterns`, `golang-structs-interfaces`, and `golang-documentation`.
- GSD provenance: `scripts/gsd sources` resolved `discuss-phase`, `plan-phase`,
  `execute-phase`, `verify-work`, and `code-review`. This provider-category phase is not
  registered by the official runtime, and the parent contract forbids role spawning; this is the
  documented inline manual-GSD fallback with explicit discussion/plan/RED/GREEN/verify/review
  evidence.

## Live artifact audit — completed before RED

The source is Zoom's current provider artifact, not the inherited endpoint ledger.

| Item | Evidence |
| --- | --- |
| API URL | `https://developers.zoom.us/docs/api/chatbot.md` |
| Retrieval | `2026-08-08T12:36:23Z` |
| HTTP / bytes | `200` / `14,740` |
| SHA-256 | `1faa9f8a419c91703ea9dec1a0ee70e027c446c3c6bf41d395e0923c19b750b6` |
| Artifact | OpenAPI `3.1.1`, API `2`, server `https://api.zoom.us/v2` |
| Ledger delta | `0` — exactly four `provider_module=chatbot` rows already match method, path, title, and source URL |

The live `## Operations` section contains exactly these four actions:

| Method | Path | Provider title |
| --- | --- | --- |
| POST | `/im/chat/messages` | Send Chatbot messages |
| PUT | `/im/chat/messages/{message_id}` | Edit a Chatbot message |
| DELETE | `/im/chat/messages/{message_id}` | Delete a Chatbot message |
| POST | `/im/chat/users/{userId}/unfurls/{triggerId}` | Link Unfurls |

All four document the `imchat:bot` scope and the Client Credentials Flow. The artifact explicitly
requires a `client_id:client_secret` HTTP Basic authorization header at its OAuth token endpoint;
the access token then becomes a Bearer credential for the API action. The `content` request member
is a provider-defined JSON object for send/edit and a string for link unfurls. Link Unfurls returns
`204 No Content`, so executable verification must assert status rather than inventing a response
body.

## Locked decisions

- Implement all four actions as `rest_write` / `direct_write` operations. No row is
  `unsafe_or_disallowed`, no duplicate is excluded, and no generic HTTP or shell escape hatch is
  introduced.
- All actions retain the typed plan → no-network preview → explicit single-use approval → execute
  lifecycle. DELETE additionally uses the existing destructive typed confirmation gate. Link
  Unfurls is a status-bearing action with `output_policy: none`; its `204` is success evidence.
- Add a reusable `oauth2_client_credentials` client-auth-style foundation. Existing declarations
  retain safe form-post behavior by default; `client_auth: basic` sends only `grant_type`, scope,
  and declared extra form parameters while placing the declared client ID/secret in HTTP Basic.
  It is used only by the Chatbot operation-scoped transport and unblocks other documented OAuth
  providers that require `client_secret_basic` wire behavior.
- Add a reusable `json_object` CLI flag foundation. It accepts exactly one JSON object, rejects
  arrays/scalars/trailing data, and remains legal only for an operation-declared object-typed body
  member that the operation's closed body schema validates. The old generic `json` flag remains
  rejected. This makes the provider's named `content` object executable without accepting a raw
  arbitrary HTTP body.
- Add Chatbot-only secret credential fields `chatbot_client_id` and `chatbot_client_secret` plus a
  documented `chatbot_token_url` configuration default. The token URL is independently declared so
  fixture tests can prove the exact Basic exchange; no ordinary Zoom OAuth bearer is selected for
  Chatbot actions. Each operation declares the paired operation base URL/auth override to retain
  the previously reviewed host/header isolation rule.
- Command paths mirror provider resources: `chatbot messages send`, `chatbot messages edit`,
  `chatbot messages delete`, and `chatbot link-unfurls create`. Required input flags map only to
  documented path/body fields; no `page`, `per_page`, `limit`, cursor, or other paging flag is
  hand-authored.
- Mark message/JID/account/content values as declared redaction fields for plan previews, errors,
  and JSON results. Synthetic fixtures only are used. No credential, token, or token-derived value
  is printed or recorded.
- Run `connectorgen surface-sync`; never hand-author fields it derives. The docs generator may
  alter unrelated aggregate indexes, so restore whole unrelated generated files and retain only
  Zoom-specific generated output.

## TDD execution

1. **Plan checkpoint** — commit this source audit, foundation decision, target counts, and
   verification plan before test or production changes.
2. **RED checkpoint** — add tests only. They must prove: `23 → 27` executable rows,
   `1819 → 1815` locally blocked rows, `1 → 5` direct writes; all four command paths are unknown
   before declaration; generic JSON remains rejected while the required named object flag is not
   supported; and a raw bundle declaring OAuth `client_auth: basic` is rejected before its token
   exchange can use HTTP Basic. Run and capture this failure verbatim, then commit and push it
   before any production JSON/engine change.
3. **GREEN foundations** — separately implement and test the basic-client-auth and named JSON
   object contracts (engine/connsdk/schema/validator/CLI). Retain and strengthen direct-write test
   assertions so required path bindings must correspond to declared URL placeholders rather than
   incorrectly demanding every write flag be a body field.
4. **GREEN connector** — declare the four operations, command groups/routes, closed body schemas,
   secret-safe fixtures, provider endpoint coverage, configuration, metadata, generated docs, and
   website catalog. Reconcile only Chatbot rows to `covered_by.direct_write`.
5. **Verify/review** — build `pm`; run base/group/each-command help and the full plan lifecycle
   for every command against isolated token/API loopback fixtures. Assert HTTP Basic token exchange,
   Bearer action requests, method/path/body, no output of synthetic sensitive fields, DELETE typed
   confirmation, and Link Unfurls `204` success. Review the foundation and Zoom changes, then
   record issue handoff.

## Execution deviation — required no-body foundation

The initial Chatbot lifecycle GREEN run exposed a pre-existing executor gap: an empty typed map
passed through an interface to the JSON requester becomes the literal `null`, so a declared
no-body action would not actually send an empty request. This was not a scope expansion or
deferral: Link Unfurls documents `204 No Content`, and the delivery contract requires all
status-only actions to be real actions. The work therefore added a separate RED test commit
(`b81cefb78`) and a separate engine GREEN commit (`acbf7405c`) before completing the connector
declaration. The foundation applies to every typed no-body `rest_write` and is documented in the
TDD ledger and summary.

Review then found that a declared sensitive path value can appear in an HTTP error URL even when
the operation uses `json_redacted`. The same red-first rule produced a separate test checkpoint
(`c9c89c707`) and generic path-error redaction fix (`070432f40`). It redacts typed path values in
terminal-facing direct-write transport diagnostics while preserving complete non-sensitive provider
diagnostics. This is required for Chatbot message/JID paths and future declared writes alike.

## Target accounting

| Measure | Before | After |
| --- | ---: | ---: |
| Zoom executable operations | 23 | 27 |
| Zoom-local implementable rows | 1,819 | 1,815 |
| Direct reads | 17 | 17 |
| Direct writes | 1 | 5 |
| Reverse-ETL writes | 2 | 2 |
| `unsafe_or_disallowed` Zoom rows | 0 | 0 |

## Verification plan

- RED/GREEN focused tests for `connsdk`, engine auth/direct-write, commandrunner, Zoom bundle,
  connectorgen, conformance/certify, and the app lifecycle.
- `go run ./cmd/connectorgen surface-sync --check`, full validation, and scoped
  `surface-reconcile --check --notes-contains provider_module=chatbot`.
- Fresh binary help: `pm help zoom`, bare `pm zoom`, bare `pm zoom chatbot`, and every exact
  Chatbot command `--help`.
- Fresh binary plan/preview/approval/execute for every action through an isolated local fixture;
  no real provider mutation or real credential is used.
- Scoped CI-equivalent gates from `AGENTS.md`: vet, lint, docs, website typecheck, CLI golden
  transcripts, contract/surface/boundary/release checks. The full suite remains CI-owned.

## Canonical references

- `AGENTS.md`
- `docs/migration/HANDOFF-CODEX.md`
- `docs/migration/conventions.md`
- `docs/architecture/connector-architecture-v2-design.md`
- `.agents/agentic-delivery/contracts/issue-agent-contract.md`
- `.agents/agentic-delivery/contracts/parent-orchestrator-contract.md`
- `.agents/agentic-delivery/references/gsd-pi-adapter.md`
- `.agents/agentic-delivery/references/cli-help-docs-website-parity.md`
- `https://developers.zoom.us/docs/api/chatbot.md`

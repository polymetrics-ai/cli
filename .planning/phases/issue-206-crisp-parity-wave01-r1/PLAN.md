# PLAN — Crisp provider parity Wave 1

## Scope

Deliver one Crisp declarative connector bundle in two reviewable commits:

1. an inventory-only `api_surface.json` commit containing the R1 report's complete, provider-owned 234-operation ledger; then
2. a Wave 1 connector bundle that makes exactly 21 core conversation GET operations individually reachable as `pm crisp <command>`.

Allowed implementation paths:

- `internal/connectors/defs/crisp/**`
- Crisp-owned connector tests/fixtures and generated CLI/docs artifacts when registration requires them
- `.planning/phases/issue-206-crisp-parity-wave01-r1/**`

Hard stops: shared runtime/engine/command-runner changes, live provider activity, credentials, non-Crisp connector edits, later waves, pushes, PRs, or merges.

## Wave 1 contract

The bundle uses HTTP Basic credentials, a safe read-only check request, and one stream/CLI surface per operation:

1. `GET /v1/website/{website_id}/conversations/{page_number}`
2. `GET /v1/website/{website_id}/conversations/suggest/segments/{page_number}`
3. `GET /v1/website/{website_id}/conversations/suggest/data/{page_number}`
4. `GET /v1/website/{website_id}/conversations/spams/{page_number}`
5. `GET /v1/website/{website_id}/conversations/spam/{spam_id}/content`
6. `GET /v1/website/{website_id}/conversation/{session_id}`
7. `GET /v1/website/{website_id}/conversation/{session_id}/messages`
8. `GET /v1/website/{website_id}/conversation/{session_id}/message/{fingerprint}`
9. `GET /v1/website/{website_id}/conversation/{session_id}/routing`
10. `GET /v1/website/{website_id}/conversation/{session_id}/meta`
11. `GET /v1/website/{website_id}/conversation/{session_id}/original/{original_id}`
12. `GET /v1/website/{website_id}/conversation/{session_id}/pages/{page_number}`
13. `GET /v1/website/{website_id}/conversation/{session_id}/events/{page_number}`
14. `GET /v1/website/{website_id}/conversation/{session_id}/files/{page_number}`
15. `GET /v1/website/{website_id}/conversation/{session_id}/state`
16. `GET /v1/website/{website_id}/conversation/{session_id}/relations`
17. `GET /v1/website/{website_id}/conversation/{session_id}/participants`
18. `GET /v1/website/{website_id}/conversation/{session_id}/block`
19. `GET /v1/website/{website_id}/conversation/{session_id}/verify`
20. `GET /v1/website/{website_id}/conversation/{session_id}/browsing`
21. `GET /v1/website/{website_id}/conversation/{session_id}/call`

All list-style operations have bounded page-number handling. Singleton operations require typed non-empty configuration for their documented path variables. Every CLI command includes its matching `{method,path}` api-surface reference.

## TDD execution slices

1. **Inventory checkpoint:** extract and JSON-validate the R1 ledger, commit only `api_surface.json`, and verify all 234 endpoint rows retain source citations and blocked disposition.
2. **Red check:** add a connector-local parity test that requires the bundle, Basic-auth spec/check, 21 stream coverage rows, and 21 reachable CLI command declarations; run it before implementation and capture its failure.
3. **Implementation:** add metadata, credential spec, streams, schemas, fixtures, `cli_surface.json`, docs, and change the matching ledger rows to `covered_by.stream`.
4. **Green verification:** validate the whole defs root, run Crisp-focused conformance/parity tests and relevant CLI tests, build `./cmd/pm`, and run all 21 `pm crisp <command> --help` paths without credentials.
5. **Review:** run the required GSD verify and code-review prompts plus scoped static checks; fix connector-local findings only.

## Honest remaining scope

After this wave, the ledger must report 21 executable provider operations and 213 non-executable operations. The later Wave 2 people/message/website reads, Wave 3 helpdesk reads, Wave 4 remaining safe reads, and Wave 5 writes are connector-local implementation backlog, not completed parity. The report's 16 named shared-foundation blockers remain unchanged.

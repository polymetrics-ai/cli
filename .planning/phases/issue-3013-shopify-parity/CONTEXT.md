# CONTEXT — issue-3013-shopify-parity resume

## Locked decisions

- Captain decision `admin-api-host-allowlist` (2026-08-06): the Admin API host is only lowercase canonical `*.myshopify.com`; no custom-domain allow-list is authorized.
- `shop_domain` enters an HTTPS Admin API base URL with the access token header, so validation must reject unsafe hosts at credential acceptance before vault/state persistence.
- The required fixture host `fixture-shop.myshopify.com` remains valid.
- Destructive/delete operation policy is already idempotently recorded on #3013 and #3014–#3020. Do not append duplicate markers or change the reconciled counts.

## Scope fences

- No shared engine, command runner, direct-read/direct-write path, dependency, live provider call, credential, write execution, certification, merge, or default-branch push.
- The sole shared-path exception is a connector-specific row and acceptance test in the existing credential-boundary test table; it changes no runtime behavior and proves the connector declaration is enforced.
- Regenerate only the connector documentation artifacts affected by the updated spec, and inspect the diff for unrelated generated drift.

## GSD execution mode

`discuss-phase --auto` and `plan-phase --tdd --skip-research --auto` were run through `scripts/gsd prompt`. Existing issue context and the captain decision answer every material question, so this resume proceeds inline without a user question or a mutating subagent.

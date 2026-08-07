# Verification checklist — help-scout documented-operation parity

Phase `help-scout-parity-sweep-r1`. GSD lifecycle and fallback recorded in
`.planning/traces/gsd-top50-sweep-continue-r1.md`.

## Derivation

- [x] The ledger's recorded `artifact_url` (`…/mailbox-api/endpoints/`) **404s** — confirmed, it is a
      section prefix, not a page.
- [x] No machine-readable spec exists (no `.json`/`.yaml`/openapi/swagger linked from the docs).
- [x] Inventory taken from the shared left-nav: **146 unique endpoint page URLs** (183 raw hrefs
      before dedup — the nav repeats the current-page entry).
- [x] All 146 pages fetched individually; **each renders exactly one `METHOD path` request line** —
      no page with zero, none with two.
- [x] Deduped on the **templated** path each page publishes in its own `Path Parameters` block →
      **144** operations, `GET 79 / POST 21 / PUT 20 / PATCH 6 / DELETE 18`.
- [x] Both deltas explained with evidence rather than asserted: 146→145 (Accept-header pair),
      145→144 (`?async=true` query variant).
- [x] Full inventory committed as `DERIVED-OPERATIONS.json` so the count is auditable, not a claim.

## Red before green

- [x] `cmd/connectorgen/help_scout_api_surface_test.go` written **first**, against the unmodified
      bundle.
- [x] **Red was authored before it could be run** (corrupted shared Go build cache). It was committed
      with `red_confirmed: false` and an explicit blocker note rather than claiming unobserved
      evidence; **no authoring happened in that window**.
- [x] Once the cache recovered, the run was made and the verbatim failure captured in a separate
      commit. `red_confirmed` false → true.
- [x] Finding F5 check: no pre-existing `cmd/connectorgen` test referenced help-scout.
- [x] No test weakened, skipped, narrowed, or deleted.

## Gates

- [x] `go run ./cmd/connectorgen validate internal/connectors/defs/help-scout` → **0 findings**.
- [x] `go test ./cmd/connectorgen` — the **whole package** → `ok … 12.059s`.
- [x] `go test ./internal/connectors/commandrunner -run
      TestEveryImplementedCommandPassesRuntimePreflight` → **PASS**.
- [x] `go run ./cmd/connectorgen surface-sync --check` → 551 scanned, **0 filled / 0 corrected**.
- [x] `go build ./cmd/pm` succeeds.

## Reachability — run, not assumed

- [x] **All 139 commands invoked through the built binary**; 0 unreachable.
- [x] `pm help-scout delete-customer plan --help` shows `--customer-id` **and** `--async`
      (`enum true|false`), proving the count collapse did not drop the async behaviour.
- [x] `pm help-scout attachments download --help` renders the bounded binary download.

## Confinement — the delta is help-scout only

- [x] `operation_endpoint_ledger.json` **unchanged**. The only operation is a `binary_download`, and
      `deriveOperationDirectReadEndpointLedger` only ledgers `rest_read`/`provider_search` kinds.
- [x] No other connector's bundle touched.
- [x] Counts recomputed **from the files themselves**: 144 rows, 144 unique keys, 139 covered, 5
      blocked, 0 excluded, 0 blank, **0 rows containing `?`**, **0 rows containing `*`**, 0 dangling
      `covered_by`, every stream/write/direct-read command reference resolving, every declared schema
      file present.

## Two defects the gates caught — both mine, both recorded

- [x] **The dedup survivor named the write action**, producing `delete_customer_asynchronously` on a
      path that performs the sync delete. Fixed by naming the action for the endpoint and exposing
      `--async` as a query flag.
- [x] **An assumed primary key.** New stream schemas were given `x-primary-key: ["id"]`; Help Scout's
      user statuses key on `userId`. `connectorgen validate` rejected it as `primary_key_missing`;
      primary keys now come from the record the provider actually returns.

## Standing constraints

- [x] **No hand-authored paging flags.** Checked with the standing regex → empty.
- [x] Every blocked row carries a machine-checkable `Named dependency:` marker → 5/5.
- [x] No webhook EVENT rows. The 26 events live at `developer.helpscout.com/webhooks/` and are
      excluded by policy; the 5 `/v2/webhooks` **management** endpoints are counted as operations.
- [x] `capabilities.write` flipped false → true, as `connectorgen validate` requires once a
      non-excluded mutation exists.

## Known-unmet — recorded, not skipped

- [ ] **CLI help/docs/website parity overlay.** `docs/cli/**`, `website/**`, generated help/manual
      artifacts and golden transcripts are **not** regenerated in this phase; the consolidated-sweep
      model regenerates shared artifacts **once at the end**.
- [ ] `TestGoldenTranscripts/root_bare_manual` fails on this branch — **verified pre-existing**
      (reproduced with this phase's changes stashed out) and resolved by that regeneration.

Both are sweep-level obligations to be discharged before the PR merges.

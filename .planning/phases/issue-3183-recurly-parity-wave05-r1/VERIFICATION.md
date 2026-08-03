# Verification checklist — Recurly parity wave05 r1

## Final gate rerun after resume

- [x] Isolation confirmed: `pwd -P` and `git rev-parse --show-toplevel` both point at `/Users/karthiksivadas/.treehouse/cli-83d592/48/cli`; branch is `fm/cli-recurly-parity-wave05-r1`.
- [x] Inventory preserved at `/tmp/recurly-wave05-r1-snapshot-20260801023416` before final reruns.
- [x] Scope guard: modified/untracked files are limited to Recurly defs, Recurly generated connector docs, generated website connector catalog surfaces, CLI golden transcripts, and GSD planning artifacts. No shared runtime files are modified.
- [x] Official source count reproduced from Recurly OpenAPI v2021-02-25: 197 operations (GET 97, POST 42, PUT 35, DELETE 23).
- [x] Focused connectorgen validation for Recurly via single-connector temp defs root: `exit 0`, 0 findings, 0 warnings, 1 connector checked.
- [x] Full connectorgen validation: `go run ./cmd/connectorgen validate internal/connectors/defs --json` checked 549 connectors with 0 findings and 0 warnings.
- [x] Focused conformance: `go test ./internal/connectors/conformance -run 'TestConformance/recurly' -count=1` passed (`ok`, about 3.6s).
- [x] Focused CLI/dynamic/golden/docs tests passed: `go test -timeout 10m ./internal/cli -run '<focused dynamic/golden/docs regex>' -count=1` (`ok`, about 120s).
- [x] `go build ./cmd/pm` passed.
- [x] Connector docs validation: `./pm docs validate --connectors-dir docs/connectors` passed.
- [x] Boundary check: `make connector-boundary` passed with `outcome: clean`.
- [x] `git diff --check` passed after final edits.
- [x] Issue addendum marker previously verified exactly once on #3183-#3190.

## Schema / CLI correction

- [x] Direct-read CLI flags map only concrete required body leaves, not whole JSON objects or arrays.
- [x] `pm recurly gift cards preview --help` shows `--unit-amount (integer)` mapped to `body.unit_amount`, avoiding object/array passthrough and keeping direct read request construction typed.

## Make verify note

`make verify` was attempted before the final focused/full reruns. It repeatedly reached the `go test -timeout 20m ./...` target and timed out in unrelated timeout-heavy packages (`internal/cli` and/or `internal/connectors/certify`) while many other worktrees were concurrently running `make verify` / `go test` processes on this host. Current process evidence showed concurrent `/tmp/fm-cli-*` test binaries and other `make verify` / `go test -timeout 20m ./...` runs. This is recorded as an unrelated local resource-contention timeout, not a Recurly validation failure. The Recurly-specific and full connector validation/build/docs/boundary/diff gates above are green.

## Not run by design

- Live provider checks, credentialed Recurly calls, certification claims, pushes, PR updates, VPS, Thaalam changes, `/no-mistakes`, and shared-runtime edits were intentionally not run for this wave.

## Resume session — CLI command-surface close (fresh worker)

- [x] Isolation reconfirmed on `fm/cli-recurly-parity-wave05-r1` at `459bf2781`.
- [x] Full 197-command surface built via `scripts/gen-recurly-cli-surface.py`: etl=93, reverse_etl=96, direct_read=8; availability implemented=194, planned=3 (bounded binary), no partial.
- [x] Every stream and every write has its own `pm recurly <command>` entry; required record fields map to typed leaf flags (kebab-case flags over `record.*` targets).
- [x] `go run ./cmd/connectorgen validate` → 549 connectors, 0 findings (no unknown targets).
- [x] `recurly_full_surface_test.go` added and passing (coverage stream=93, write=96, direct_read=5, operation=3).
- [x] Conformance: `go test ./internal/connectors/conformance -run 'TestConformance/recurly' -count=1` PASS.
- [x] Test suites PASS: internal/cli (338s), connectorgen, engine, commandrunner, defs, bundleregistry, conformance.
- [x] `pm docs validate --connectors-dir docs/connectors` PASS; MANUAL/SKILL/catalog (Recurly 93/96 read+write) and website connector data regenerated; diff scope limited to Recurly-owned files.
- [x] The 3 official binary/export endpoints remain honest `planned` (bounded metadata) dispositions; no certification claims.

### Kept out of scope (unchanged)

- No live/credentialed Recurly calls; no shared-runtime edits; no other connectors touched; no pushes/merges; PR stays draft for firstmate.

## Captain review-fix session (authorized fixes for complete parity)

- [x] Guarded custody recover attempted first: `no-mistakes axi sync --recover --keep-local` refused again (`blocked_recover_gate_diverged`: gate branch 459bf2781 != preserved head 3f8ce5a5, no files/refs changed). Per captain instruction, no unguarded reset was run; custody escalation recorded.
- [x] Fix 1 — required-body writes now send real JSON bodies (body_type json): update_account (AccountUpdate), create_billing_info (BillingInfoCreate), create_usage/update_usage (UsageCreate), update_subscription (SubscriptionUpdate). Fixtures updated to prove request bodies serialize (expect.body).
- [x] Fix 2 — create_account and update_account record_schemas now accept billing_info, address, custom_fields, company (full AccountUpdate body fields).
- [x] Fix 3 — schemas/get_account_balance.json models the real AccountBalance shape (object/account/past_due/balances); removed the fabricated id primary key and fake code/state/created_at/updated_at fields; fixture rewritten to the real shape.
- [x] Fix 4 — refund_invoice is confirm:"destructive" and its record_schema includes amount/percentage/line_items/refund_method/credit_customer_notes/external_refund.
- [x] Regression: TestRecurlyReviewFixFindings added to recurly_full_surface_test.go locking all four fixes.
- [x] Gates: connectorgen validate exit 0, 0 findings; recurly conformance PASS; engine/commandrunner/defs/bundleregistry/connectorgen PASS; `go test ./internal/cli` PASS (348s); docs/catalog/website regenerated (Recurly-only scope).

### Blocked on custody (escalated)

- /no-mistakes re-run is blocked on custody recovery that the guarded path refuses (`blocked_recover_gate_diverged`). This requires firstmate/captain resolution; no unguarded reset was run.

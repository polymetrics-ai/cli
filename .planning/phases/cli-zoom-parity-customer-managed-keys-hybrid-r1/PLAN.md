# Plan — Zoom Customer Managed Keys Hybrid documented-operation parity, R1

## Delivery record

- Parent: [#3915](https://github.com/polymetrics-ai/cli/issues/3915); slice:
  [#3950](https://github.com/polymetrics-ai/cli/issues/3950).
- Scope owner: Zoom's provider-owned **Customer Managed Keys Hybrid** category only: its one
  key-connector archival API, the narrowly required direct-write redaction/confirmation
  foundation, Zoom-local test fixture, generated CLI/website docs, and this phase evidence.
- Required skills in scope: `golang-how-to`, `golang-cli`, `golang-testing`,
  `golang-error-handling`, `golang-security`, `golang-safety`, `golang-design-patterns`,
  `golang-structs-interfaces`, `golang-documentation`, and `no-mistakes` (loaded for the
  parent run and recorded here).
- GSD provenance: `scripts/gsd doctor`; `scripts/gsd sources` for `discuss-phase`,
  `plan-phase`, `execute-phase`, `verify-work`, and `code-review`; generated command prompts
  were inspected. This provider-category phase is not registered by the official runtime, so it
  follows the documented inline manual-GSD fallback: discussion, plan, RED, execute, verify,
  and review evidence remain here, and no forbidden role worker is spawned.

## Source audit — completed before RED

The source of truth is Zoom's current provider artifact, not the inherited endpoint ledger.

| Item | Evidence |
| --- | --- |
| API URL | `https://developers.zoom.us/docs/api/customer-managed-keys-hybrid.md` |
| Retrieval | `2026-08-08T10:54:00Z` |
| HTTP / bytes | `200` / `1,890` |
| Artifact | OpenAPI `3.1.1`, API `2`, server `https://{keyConnectorLb}/api/v2` |
| Reconciled operation | POST `/api/v2/kms/cse/archival/datakey/decrypt` — Key connector archival API |
| Ledger delta | `0` — the inherited row matches method, origin-relative path, title, source URL, and `provider_module=customer-managed-keys-hybrid` |

The linked authentication reference is separately audited at
`https://developers.zoom.us/docs/api/references/key-connector-archival-auth/` on
`2026-08-08T10:55:19Z` (HTTP `200`, `24,583` bytes). It documents a customer-managed,
JWKS-signed bearer JWT with the archival scope, sent only to the customer-hosted key connector;
it is not a standard Zoom OAuth bearer token. No token value, certificate, or key material is
recorded in this repository.

The operation requires JSON `encrypt_context` and `key_id`, and a successful JSON response can
contain `key_id` and a base64 plaintext key. It has no page, per-page, limit, cursor, or other
pagination input. The OpenAPI server includes `/api/v2`; the operation declaration preserves its
ledger-exact `/api/v2/...` path while the existing base-URL normalizer sends the correct relative
`/kms/...` request beneath a customer host configured as `https://<host>/api/v2`.

## Locked implementation decisions

- Implement the sole documented POST. There are zero `unsafe_or_disallowed` rows and no
  exclusions.
- Use the typed `rest_write` / `direct_write` executor, never a generic HTTP mechanism and never
  `writes.json`: this response must be returned through the declared redacted output policy.
  It remains single-request, non-batchable, plan → no-network preview → explicit single-use
  approval → typed confirmation → execute.
- Add the narrowly needed `json_redacted` direct-write foundation in its own commits. It must
  redact generic secret-shaped response fields and declared sensitive fields, redact declared
  request literals from direct-write error text, and render a direct-write plan sample from its
  redacted record while preserving the private execution record. This unblocks all connectors
  with legitimate secret-returning declared `rest_write` operations; it is not Zoom-specific.
- The operation is `secret_sensitive` with `sensitive_policy` input mode `env_or_stdin`, transform
  `none`, fields `encrypt_context`, `key_id`, and `plainkey`, and typed confirmation. The declared
  operation—not a hand-authored flag—owns sensitive output behavior.
- Add separately stored `key_connector_jwt` and `key_connector_base_url` credential fields.
  The Key Connector base URL is explicitly provisioned and must end in `/api/v2`. The one
  operation declares its own paired `rest.base_url` and `rest.auth`, so it cannot inherit or
  send the normal Zoom OAuth bearer to that customer-hosted origin; ordinary Zoom streams retain
  their existing base URL and bearer declaration.
- Expose only `--encrypt-context` and `--key-id`, both required typed JSON body inputs. Do not
  hand-author `page`, `per_page`, `limit`, cursor, or other undocumented flags.
- `surface-sync` owns derivable command metadata and the endpoint-ledger projection. Regenerate
  docs and website catalog, then mechanically retain only Zoom/catalog deltas; never hand-merge
  generated files.

## TDD execution slices

1. **Plan checkpoint** — this source audit, configuration/auth decision, and foundation scope
   exist before any production declaration.
2. **RED checkpoint** — change only direct-write safety tests, the Zoom command-surface test, and
   synthetic fixture/evidence. It asserts `22 → 23` covered, `1820 → 1819` locally blocked,
   direct reads stay `17`, direct writes become `1`, reverse-ETL writes stay `2`, and the new path
   is unknown before declarations. It also proves current `json_redacted` direct-write output
   leaks both generic and explicitly named sensitive fields. Run and capture the failure; commit
   and push it before any production engine or Zoom JSON changes.
3. **GREEN foundation checkpoint** — in a separate commit, make redacted direct-write response,
   error, and plan-preview handling honor the declared policy and sensitive fields; make a
   secret-sensitive direct write require the existing typed confirmation grant. Run targeted
   engine/commandrunner/app tests. The Zoom surface remains intentionally RED until its following
   declaration commit.
4. **GREEN operation-origin/auth foundation checkpoint** — add a typed `rest.base_url` and
   `rest.auth` override for declared `rest_write` operations only. It must bind preview and
   execution to the same customer-hosted origin, retain the shared approval/rate-limit/transport
   controls, and prove the ordinary Zoom OAuth bearer is never sent to that origin. Commit it
   separately before authoring Zoom JSON.
5. **GREEN connector checkpoint** — declare the one `rest_write`, CLI group/command/body flags,
   operation-scoped auth profile, endpoint projection, metadata, fixture, and generated docs. Test
   exact normalized request path, key-connector bearer selection, required fields, response/error
   redaction, no pagination flags, and approval-gated execution through an isolated fixture server.
6. **Verify/review checkpoint** — build `pm`; run base, group, and command help plus plan-lifecycle
   reachability from the binary. A live customer key connector is not discoverable without an
   operator-owned deployment, so run the exact command only against an isolated loopback fixture
   with an environment-only synthetic JWT and prove it reaches the declared POST/approval gate,
   never unknown-command/unknown-flag and never outputting the synthetic value. Record the
   bounded external-runtime limitation rather than inventing a host.

## Verification plan

- RED/GREEN: focused `go test -count=1` for direct-write engine, commandrunner/app lifecycle, and
  `./internal/connectors/defs/zoom/...`.
- Surface: `go run ./cmd/connectorgen surface-sync --check`, Zoom validation, and full validation.
- Runtime: Zoom conformance, commandrunner, `go vet ./...`, and a built temporary `pm` binary.
- Binary: `pm help zoom`, bare `pm zoom`, bare Customer Managed Keys Hybrid namespace, the exact
  command `--help`, and its plan/preview lifecycle against an isolated fixture credential created
  from environment-only synthetic input.
- Docs: `pm docs validate --connectors-dir docs/connectors`, generated website catalog plus
  typecheck, docs/website scope comparisons, golden checks, and the required local non-full-suite
  gates listed in `AGENTS.md`.

## Canonical references

- `AGENTS.md`
- `docs/migration/HANDOFF-CODEX.md`
- `docs/migration/conventions.md`
- `docs/architecture/connector-architecture-v2-design.md`
- `.agents/agentic-delivery/contracts/issue-agent-contract.md`
- `.agents/agentic-delivery/contracts/parent-orchestrator-contract.md`
- `.agents/agentic-delivery/references/gsd-pi-adapter.md`
- `.agents/agentic-delivery/references/cli-help-docs-website-parity.md`
- `https://developers.zoom.us/docs/api/customer-managed-keys-hybrid.md`
- `https://developers.zoom.us/docs/api/references/key-connector-archival-auth/`

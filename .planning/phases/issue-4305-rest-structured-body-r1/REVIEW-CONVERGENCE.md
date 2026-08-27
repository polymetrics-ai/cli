# Review convergence: Issue #4305

## Frozen review target

- **Head SHA:** `af241838624b8ee4b372b9b652f414c65f35f34f`
- **PR/base:** #4313, `fm/cli-rest-structured-body-r1` → `main`
- **Base SHA:** `060bb7864e3419e09ab10e000bb14ac1ea3724ec`
- **Freeze rule:** every review pass reads this SHA or a subsequently named fix SHA; no lens may edit production files. The coordinator alone records ledgers and applies one coordinated accepted-finding wave.
- **GSD route:** `scripts/gsd sources code-review` and `scripts/gsd prompt code-review` were run. The official command expects a numbered roadmap phase and one reviewer; this issue phase uses the documented inline/manual fallback so Firstmate's explicit ten-lens, frozen-SHA procedure can be recorded faithfully.
- **Skills loaded for this review:** `gsd-ns-review`, `golang-how-to`, `golang-security`, `golang-safety`, `golang-error-handling`, `golang-lint`, `golang-testing`, plus the existing CLI/documentation skills recorded in the phase plan.

## Required acceptance evidence

1. Provider-owned structured request values remain declaration-bound and fail before I/O for schema, metadata, and action/command boundary violations.
2. No provider base URL or any other credential/configured secret reaches plaintext `state.json` on any direct-write success, application-error, transport-error, or mixed-error path.
3. Provider receipts, headers, and bodies are still returned verbatim to the caller. Persistence safety must **not** redact, truncate, or otherwise alter provider output.
4. Existing scalar, form, SCIM, binary, specialized GitHub, typed-action, and approval/confirmation outcomes remain compatible.

## Convergence protocol

1. Run the ten read-only lenses below against the frozen target, collecting severity, file/line, exploit/regression path, evidence, and required test.
2. Merge the ten lens ledgers into the finding ledger below. Duplicate reports collapse into one canonical finding; every report receives a disposition.
3. Run exactly one audit-worker pass over the frozen target and merged ledger using the requested `--harness claude` route. If the local runner cannot provide that harness, record the truthful external-harness fallback and do not label a Codex pass as Claude.
4. Apply one coordinated fix wave for accepted in-scope findings, using red-green TDD and targeted validation.
5. Spawn a fresh Codex context that did not write the fixes to review the exact resulting SHA. No finding cap, round cap, or forced-green rule applies.
6. Publish the final SHA only after local validation; record CI rollup and the fresh-context outcome against the same SHA.

## Ten role-separated lenses

| ID | Lens | Required focus | Status |
| --- | --- | --- | --- |
| L1 | Secret taint and plaintext persistence | Trace configured `base_url`, credentials, headers, body, errors, `ReverseRun`, and every `state.json` write/read path; prove no secret persistence while the returned receipt is verbatim. | complete: F1, F4; notes the broader credential-config storage contract |
| L2 | Provider receipt/output contract | Verify caller-visible success/error result, HTTP error, header, body, and GraphQL paths preserve exact provider values and avoid output redaction/truncation. | complete: F4; normal path verified verbatim |
| L3 | Structured-schema closure | Review schema normalization/compiler closure, nested object/array constraints, type fidelity, limits, and source-schema disagreement. | complete: F2, F3 |
| L4 | CLI and generated surface | Review command flags/help/manual/schema, raw-body escape rejection, command/action identity, and input field isolation. | complete: clean |
| L5 | Typed action and reverse-ETL integration | Trace the canonical materialization path, approval/confirmation payload binding, mutation resistance, and record/action isolation. | complete: clean; output-redaction suggestion declined by captain policy |
| L6 | Request transport binding | Trace path/query/header/body request assembly, fixed provider metadata, source parameter validation, and pre-I/O rejection. | complete: clean |
| L7 | Declaration/generator validation | Review provider declaration preflight, source-lock agreement, generated artifacts, and downstream composition contract. | complete: clean; F3 is a runtime-preflight disagreement |
| L8 | Regression and test adversary | Assess happy/bad/edge coverage, zero-I/O assertions, compatibility coverage, and missing deterministic tests. | complete: F1 coverage gap |
| L9 | Durability and compatibility | Review state persistence behavior, error lifecycle, serialization, and scalar/form/SCIM/binary/special GitHub compatibility. | complete: F4 |
| L10 | Go/API maintainability | Review error handling, API boundaries, normalization, concurrency/lifetime, code quality, and reviewable failure modes. | complete: F1, F4 |

## Merged finding ledger

| ID | Severity | Lens reports | File/line | Disposition | Evidence / fix / test |
| --- | --- | --- | --- | --- | --- |
| F1 | High | L1, L8, L10 | `internal/app/app.go:3430-3439`; `internal/connectors/connectors.go:1086-1149` | accepted | Persisted direct-write receipt is sanitized only on errors and only from runtime secrets. Always produce a persistence-only clone that masks runtime/config/request-sensitive variants on every completion path; keep the caller result verbatim. Add HTTP and GraphQL success/error coverage for credential/config/body-sensitive echoes in headers, raw/decoded bodies, and GraphQL messages. |
| F2 | High | L3 | `internal/connectors/engine/structured_rest_body.go:694-711,799-814` | accepted | `json.Unmarshal` turns declaration numeric constraints into `float64` and rounds values beyond IEEE-754 precision. Parse declaration schemas with `UseNumber` and preserve exact source numeric lexemes. Add enum/minimum/maximum high-precision regression coverage. |
| F3 | Medium | L3 | `internal/connectors/engine/structured_rest_body.go:1157-1201,1513-1629` | accepted | `patternProperties` can pass structured-field preflight but canonicalization refuses all matching keys. Reject unsupported dynamic schema forms while validating the declaration, before CLI reachability/I/O. |
| F4 | Medium | L1, L2, L9, L10 | `internal/app/app.go:3451-3465` | accepted | On an indeterminate post-rename failure, finalization reloads and returns the persistence-sanitized copy. Preserve durable recovery/error semantics but return the original in-memory provider receipt to the caller. Add REST/GraphQL post-rename sync-failure tests. |
| F5 | Info | L5 | `internal/app/app.go:3184,3465`; `internal/cli/cli.go:1881,2229` | declined | The suggested public-output masking would violate captain ruling #4337 and this review's fixed contract. Provider receipts stay verbatim at every caller-facing boundary; only the durable clone is sanitized. |
| F6 | Info | L2 | `internal/connectors/engine/direct_write.go:351-380` | deferred | Typed HTTP-error reconstruction omits non-UTF-8 `RawBody` although the primary `OperationDirectWriteResult` retains it. This is pre-existing and not needed for the plaintext-state fix; retain as residual compatibility risk, no scope expansion. |
| F7 | High | audit | `internal/app/app.go:3409`; `internal/connectors/connectors.go:1037`; `internal/connectors/engine/write.go:456` | accepted, scope confirmation required | Generic non-`OperationDirectWrite` provider receipts are cloned but not secret-sanitized before `DestinationResult` reaches plaintext state. The same persistence-only policy must cover supported JSON/form/SCIM/binary generic writes on both success and error while keeping the caller result verbatim. |

### Lens evidence summary

- L1 verified ordinary failed runtime-secret echoes are already persistence-sanitized, but successful receipt echoes, config-bound echoes, and request-sensitive echoes are not comprehensively covered. It also identified that credential metadata serializes `Config`; this is a broader credential-storage contract rather than a receipt persistence path and remains explicitly out of this narrow fix unless the audit proves it is required by the accepted Firstmate scope.
- L2 verified ordinary direct CLI and reverse-run JSON output serializes the original run and preserves HTTP/GraphQL status, multi-value headers, raw body, decoded body, and typed cause.
- L3 verified canonical structured-body materialization is shared and bounded before transport, while F2/F3 violate source fidelity/reachable declaration preflight.
- L4/L6/L7 found closed definition-owned command, request, and generator binding with no raw-body or metadata override route.
- L5 verified re-materialization and approval digest checks reject action/body mutation; its only proposal was declined as F5.
- L8/L9/L10 independently confirmed F1/F4 and found no scalar/form/SCIM/binary/special GitHub regression outside those concerns.

## Audit and fresh-context records

| Stage | Target SHA | Worker/harness | Result | Evidence |
| --- | --- | --- | --- | --- |
| Ten-lens merge | `af241838624b8ee4b372b9b652f414c65f35f34f` | role-separated read-only lenses | 4 accepted findings, 1 declined policy conflict, 1 deferred pre-existing residual | this record |
| Audit | `af241838624b8ee4b372b9b652f414c65f35f34f` | requested `--harness claude` | blockers F1–F4 and F7 verified; F5 declined, F6 deferred | frozen read-only audit worker report |
| Fresh-context re-review | pending fix SHA | Codex context with no fix authorship | pending | pending |

## Audit result and required scope decision

The audit independently verified F1–F4, found no reason to reopen F5/F6, and added F7. Its F1/F7 acceptance bar is: no configured credential, config-bound request value, or request-sensitive value may be retained in any persisted provider receipt; caller-visible receipt remains exact.

**Open decision: `state-config-base-url-persistence`.** The literal requirement that `base_url` never reaches `state.json` on any path is not met independently of any provider receipt: `CredentialMeta.Config` is serialized to plaintext state (`internal/app/types.go`) and credential creation copies `req.Config` into it (`internal/app/app.go`). Removing or masking it there can prevent a configured connector from being restored and is a credential-storage/schema migration, while the prior accepted scope explicitly excluded generic state-file encryption and narrowed the change to receipt persistence. Firstmate must choose one current contract before the fix wave:

1. **Receipt boundary only:** repair F1–F4/F7 for all persisted provider receipts; credential metadata configuration remains the existing separate contract.
2. **Broader credential-storage redesign:** authorize a durable, compatible way to persist or restore secret-bearing `base_url`/config outside plaintext state, including migration and verification scope.
3. **Classify `base_url` as non-secret metadata:** explicitly narrow “no base_url” to echoed receipt values, not credential configuration storage.

No production fix is started until this decision resolves, because the three choices have incompatible persistence and compatibility consequences.

### Authorized partial safety follow-up — 2026-08-24

Firstmate authorized a separate, decision-independent input hardening slice: reject a declared `base_url` containing URL user-info or a query component before credential storage or request assembly. The red/green evidence is recorded in `TDD-LEDGER.md`; it covers all source declarations, including the 29 that omit `format: uri`, and keeps caller-visible provider receipts verbatim.

This does **not** resolve F1/F4/F7 or the open `state-config-base-url-persistence` decision. The unchanged plaintext GraphQL receipt test uses an ordinary endpoint URI, so the prior persistence-only projection—not URL admission rejection—is still what makes that test pass. No persistence test was weakened or removed.

## Inbox 008 research: provider-owned `base_url` classification

**Question:** does any connector provider in this repository declare an endpoint `base_url` that itself contains a credential, token, signing key, or other secret?

**Answer:** **no.** The complete declaration scan found 532 `spec.json` files with a `properties.base_url`; zero mark that field `x-secret: true`. Of the 477 declarations with defaults, zero defaults contain URL user-info (`@`) and zero contain a query delimiter (`?`). The remaining 55 fields are required/override endpoint addresses rather than defaults. Of all 532 fields, 503 have `format: uri`, 29 have no format, and none use a pattern or enum which encodes a credential-bearing URL shape. A scan of declared `url`/`base_url` template values found no static user-info or query-credential value. The sole template which mentions both endpoint configuration and a secret is Breezy HR's `{{ config.base_url }}/company/{{ secrets.company_id }}`: the secret is a separately declared path value, not part of `base_url`.

Representative provider/source evidence:

| Provider | Declaration-owned `base_url` shape | Provider source |
| --- | --- | --- |
| GitHub | `https://api.github.com` (or a GitHub Enterprise/test endpoint) | Source lock: pinned [GitHub REST OpenAPI description](https://raw.githubusercontent.com/github/rest-api-description/b26c240ded1c8b79cb0fb09dee4a21239061fa23/descriptions/api.github.com/api.github.com.json). The connector has distinct `x-secret` token/private-key properties; GitHub's [authentication documentation](https://docs.github.com/en/rest/authentication/authenticating-to-the-rest-api) sends the token in `Authorization`, not in the endpoint. |
| WooCommerce | `https://example.com/wp-json/wc/v3` | [Official REST reference](https://woocommerce.github.io/woocommerce-rest-api-docs/) is source-locked in `sources/woocommerce-parity-source-lock.json`. It documents the endpoint address separately from consumer-key/secret authentication. |
| Freshdesk | `https://&lt;domain&gt;/api/v2` | The source-locked declaration requires an endpoint host plus version path; its credential is a separate secret field. |
| Chargebee | `https://{site}.chargebee.com/api/v2` | The source-locked declaration requires a site endpoint plus version path; credentials are separate. |
| Stripe | `https://api.stripe.com/v1` | The declaration defines an endpoint/proxy override; credentials are separately declared. |
| Microsoft Teams/Graph | `https://graph.microsoft.com/v1.0` | The declaration defines the Microsoft Graph endpoint; authentication is separately declared. |

This establishes the provider-sourced fact needed for Firstmate's decision: **in-scope `base_url` is endpoint metadata, not provider credential material.** No production code or test was changed by this research.

### Runtime caveat (not a contrary provider finding)

The current generic `format: uri` validator accepts any absolute URI (`configuration_validation.go` and `schema.go`), and URL assembly preserves user-info/query components (`direct_write.go` and `connsdk/http.go`). Therefore, a caller can presently provide a non-provider-shaped credential-bearing URI even though no source-locked connector does. Classifying declared provider endpoint metadata as non-secret is supported by the scan; treating arbitrary caller-supplied URI text as categorically non-secret would require an explicit validation policy and is **not** authorized by this inbox instruction. The open decision remains unresolved.

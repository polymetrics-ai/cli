# Issue #4307: Declaration-owned headers and bounded transfer operations

## Task Delivery Header

- Issue: Refs #4307 — feat(engine): enforce declaration-owned headers and bounded transfer operations
- Base branch: `main`
- Merges into: `main`
- Delivery: Committed implementation and delivery evidence on `fm/cli-closed-operation-runtime-r1`, ready for Firstmate's later no-mistakes/PR stage; no push or merge in this lane.
- Working branch: `fm/cli-closed-operation-runtime-r1`
- Task: Build the connector-agnostic F2/F4 operation runtime: provider-declared typed non-auth request headers and bounded binary download/upload/multipart, fixed status, and text export. Extend the existing main contracts without connector-specific code, generic transports, raw operation controls, or production connector-definition edits.
- Verification: TDD evidence using two synthetic connector identities; focused and affected package tests; generator validation/goldens/docs checks; `go vet ./...`; `go build ./cmd/pm`; `git diff --check`; tracked `make connector-boundary`; `make verify`; generated CLI/manual help assessment and deep source review.

## Evidence Table

| Acceptance criterion | Evidence | Observable assertion or fake reason |
| --- | --- | --- |
| Typed headers are operation-owned and preserve request fidelity. | live | Two synthetic identities expose distinct declared headers; a provider double observes only the operation's exact path/query/header/body request. |
| Unsafe and malformed headers never reach transport. | live | A request counter remains zero for unknown, duplicate, cross-operation, protected/case-variant, CR/LF, malformed, oversized, invalid-schema, and runtime-owned header inputs. |
| Header namespaces remain isolated. | live | Path/query/body values and a same-named field on a different operation are absent from the observed header request. |
| Binary and multipart transfer stays bounded and declaration-owned. | live | Provider doubles observe exact endpoint/method/part names/media; cap, unknown part, redirect, unsafe path, and input failure assertions observe zero I/O or no completed artifact as applicable. |
| Download has atomic completion. | live | A successful bounded response creates exactly the requested output; over-cap, media, redirect, or stream failure leaves no final file or partial-success result. |
| Status/text behaviour is closed and bounded. | live | Status returns only declared bounded metadata with no decoded body; text observes declared media/charset and cap, and rejected declarations send zero requests. |
| Declared output preserves ordinary provider data. | live | Synthetic providers return unusual-but-declared status/header/body fields; the result retains all of them, while a credential/transport-secret canary remains present with the established explicit masking marker. |
| Mutating operations bind approval to exact bytes. | live | A changed typed header or multipart input after preview invalidates the approval digest before a provider double sees an execute request. |
| Existing typed paths are unchanged. | live | Existing GraphQL, scalar/form/SCIM, structured-body, credential/auth, and no-credential installed-command preflight regressions pass unchanged. |
| Generated help and adoption docs have no escape hatch. | live | Generator/golden/docs checks show typed declared inputs only; repository assertions find no raw header/body/URL/operation control exposed by the changed surface. |

## Foundation Check

| Need | Existing main evidence | Planned extension |
| --- | --- | --- |
| Typed direct-operation field maps | `operationDirectReadOverrides` handles path/query/body only. | Add a validated header namespace and a single canonical construction point that can reject protected/runtimed-owned keys. |
| Bounded binary/text/status execution | #4297 introduced closed executors and binary safe-output primitives. | Register all named kinds through loader/validator/runner mirrors; use rather than duplicate their bounded I/O and policy seams. |
| Typed structured REST body | #4305 is the active owner. | Consume the shared declared body boundary when it lands; retain headers and multipart as independent declaration-owned contracts. |
| Reverse-ETL typed destination dispatch | #4304/#4303 own its application transport. | Preserve plan → preview → approval → execute and keep binary transfers outside record writes. |

## TDD Slices

1. **Header declaration and validation (red → green → refactor).** Add synthetic-bundle and installed-command failures for missing/unknown headers, invalid name/schema/requiredness, mapping disagreement, and two-operation cross-binding. Implement the closed schema and exact mapping admission.
2. **Protected-header and isolation hardening (red → green → refactor).** Add table-driven zero-I/O tests for case/normalization variants of credential, proxy, cookie, host, connection, forwarding, and transport metadata, including CR/LF and byte bounds. Centralize canonical rejection before the request builder.
3. **F4 declaration/dispatch mirrors (red → green → refactor).** Add failing loader/validation/generator/commandrunner tests that declare the named binary/multipart/status/text kinds from two synthetic bundles. Register the existing executor contracts without a generic fallback.
4. **Bounded files and multipart (red → green → refactor).** Add request-fidelity, exact part/media, cap, unknown-part, explicit safe output, redirect/media, stream-error, and partial-file tests. Connect declaration-owned input/output validation and atomic completion to the existing approval/confirmation path.
5. **Status/text/output/approval regressions (red → green → refactor).** Prove bodyless status output, text media/charset/cap stdout/file rules, complete declared ordinary response preservation, explicit presence-preserving credential/transport masking, changed-after-preview digest rejection, no-credential preflight, and unchanged GraphQL/scalar/form/SCIM/structured-body/credential flows.
6. **Documentation and generated CLI (red → green → refactor).** Update the declaration/adoption contract and regenerate/help-test the typed CLI surface; record why bare namespaces are unaffected if no namespace command shape changes.

## Lifecycle and Skills

- Resolved and generated inline GSD prompts: `discuss-phase 4307 --auto`, `plan-phase 4307 --tdd --skip-research`, `execute-phase 4307 --interactive`, `verify-work 4307`, and `code-review 4307 --depth=standard`.
- Inline/manual fallback: #4307 is absent from `.planning/ROADMAP.md`; the firstmate brief requires autonomous single-worker execution and forbids delivery-role spawning. This preserves the lifecycle artifacts, TDD, verification, and review gates.
- Loaded skills: `golang-how-to`, `golang-cli`, `golang-testing`, `golang-error-handling`, `golang-security`, `golang-safety`, `golang-design-patterns`, `golang-structs-interfaces`, and `golang-documentation`.
- CLI parity: generated connector commands/help may change. Assess runtime help, generated manual/goldens, `docs/cli/**`, and `website/**`; record explicit not-applicable evidence for hand-authored namespace help only if none changes.

## Commit Checkpoints

1. GSD context, plan, ledger, and verification checklist.
2. Header red/green checkpoint.
3. F4 loader/dispatch and bounded-file red/green checkpoint.
4. Documentation/generator and regression checkpoint.
5. Full verification and review-fix checkpoint.

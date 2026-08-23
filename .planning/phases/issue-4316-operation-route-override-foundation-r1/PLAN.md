# Issue 4316 — Declaration-owned per-operation routing

## Task Delivery Header

- Issue: Refs #4316 — feat(engine): add declaration-owned per-operation routing
- Base branch: main
- Merges into: main
- Delivery: Pull request open against `main`, API-confirmed with its available checks green; captain retains merge authority.
- Working branch: fm/cli-operation-route-override-foundation-r1
- Task: Make API base, version, and route selection closed and declaration-owned for direct read/write, binary download/upload, ETL, and reverse ETL.  Resolve the five source-locked Help Scout v3 direct-read gaps through the shared engine, report unresolved or conflicting routing before provider I/O, and preserve source trace and generated canonical surfaces.
- Verification: Red/green/refactor targeted engine tests; Help Scout real-definition fixture tests; connector generator validation and surface-sync; affected package tests; `go vet`; build; individual local `make verify` gates; GSD verify and review prompts.

## Evidence Table

| Acceptance criterion | Evidence | Observable assertion or fake reason |
| --- | --- | --- |
| One declared operation route reaches the expected base/version/path on every executor surface | live | Engine tests run real definitions and capture the actual request URLs/methods for direct read/write, binary download/upload, ETL, and reverse ETL. |
| Five Help Scout direct-read declarations become executable through the shared engine | live | Real Help Scout bundle tests assert the five source-locked operations issue their declared v3 URL, including page navigation where declared. |
| Missing route fails before provider I/O | live | A server records zero requests while the real engine returns the source-traced missing-foundation diagnostic. |
| Conflicting bases cannot be silently chosen | live | A single operation with incompatible declaration-owned bases returns the same missing-foundation diagnostic and causes zero requests. |
| Existing CLI and documentation surface remains canonical | live | `surface-sync --check`, generator validation, and docs checks assert no hand-authored routing escape hatch or generated drift. |

## Discussion decisions

- The connector bundle, not credential/configuration or caller input, owns the final operation base, API version, and route.
- A resolver is shared by every engine surface. Surface-specific executors can supply an operation identity but cannot invent a URL or quietly retain a legacy base.
- Bad or conflicting definition data is a foundation error. It is returned before HTTP client construction/provider I/O and includes the operation and declaration-source trace used by command-runner blocked diagnostics.
- Only declaration-backed test servers may supply an execution base in fixtures; no production command gains a URL flag or generic transport escape hatch.
- This lane changes no command arguments or help text. CLI/manual/website parity is therefore verified as not-applicable-to-output-change plus the normal generated-surface and docs checks.
- Inline/manual GSD fallback: this runtime has no compatible isolated Pi lifecycle workers and the repository's canonical single-worker rule forbids role spawning. The generated GSD prompts are executed and recorded inline.

## TDD slices

1. **Red — resolver contract.** Write a real-engine test using an operation fixture with a declaration-owned base/version/route. Assert direct read requests the composed URL. Add missing and conflict cases that assert zero provider requests and the source-traced foundation error.
2. **Green — common resolver hook.** Introduce the smallest declarative operation-routing type/validation and use it at the common request-construction seam. Rerun the red test.
3. **Red — cross-surface route propagation.** Add a fixture that uses the same operation routing on direct write, binary transfer, ETL, and reverse ETL and asserts captured method/path/base; include declared pagination.
4. **Green — all engine entrypoints.** Thread the resolved route through every executor without connector-local special cases. Rerun the targeted suite.
5. **Red — Help Scout acceptance.** Add the five real Help Scout v3 direct-read behavioral cases, each asserting its exact declaration-derived request URL and no arbitrary fallback.
6. **Green/refactor — definition and canonical outputs.** Add only necessary Help Scout declaration source/mapping data, regenerate through the project tool, remove duplication, run generator/snapshot checks, and record tests/results.

## Planned checks

- `go test -timeout 20m ./internal/connectors/engine/...`
- `go test -timeout 20m ./internal/connectors/commandrunner/...`
- `go test -timeout 20m ./internal/cli/...`
- `go vet ./...`
- `go build ./cmd/pm`
- `go run ./cmd/connectorgen validate`
- `go run ./cmd/connectorgen surface-sync --check`
- Repository individual `make verify` gates named by its Makefile (never a single timeout-prone `make verify`).
- `scripts/gsd prompt verify-work 4316`; `scripts/gsd prompt code-review 4316` with inline evidence.

## CLI help/manual/website parity

- Runtime help: no flag/command changes planned; verify the changed Help Scout surface has no generated CLI diff except canonical declaration support.
- Bare namespace and `pm help`: not applicable unless the generated Help surface changes; run their project check if code generation changes output.
- `docs/cli/**`, `website/**`: no documentation change planned; run docs check and search for affected Help Scout route language if declarations change.
- Generated help/manual: `surface-sync --check` and generator validation are required.

## Skills loaded

`golang-how-to`, `golang-cli`, `golang-testing`, `golang-error-handling`, `golang-security`, `golang-safety`, `golang-design-patterns`, `golang-structs-interfaces`, `golang-context`, `golang-concurrency`, and `golang-documentation`.

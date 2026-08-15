Refs #3860

## Summary

- surface PostgreSQL's declaration-owned `polling_watermark` status as
  **planned**, with its blocking reason, through runtime help, bare namespace
  help, inspect JSON, catalog JSON, generated manuals, docs, and website data;
- document the bounded keyset mechanism, source ordering/checkpoint scope,
  at-least-once replay, snapshot restriction, hard-delete boundary,
  cursor-advancing soft-delete requirement, and explicit rebootstrap outcome;
- preserve the distinction from PostgreSQL logical-replication CDC and prevent
  a non-implemented declaration from serializing fake `source`/`target`
  bindings or a fabricated REST endpoint.

## TDD and observable proof

- **Red:** `pm connectors inspect postgres --json` omitted the polling status.
  The CLI test failed against real output.
- **Green:** the real inspect/catalog output observes `status=planned`, a
  non-empty reason, and no `implemented`, `change_capture`, `source`, or
  `target` contract. `TestPostgresNativeAPISurfaceHasNoFabricatedRESTEndpoints`
  reads the actual PostgreSQL bundle and observes `endpoints=[]`.
- **Refactor:** planned/unsupported descriptor JSON now emits only status and
  reason; the focused test proves the zero-value binding fields are absent.

## Validation

Passed:

- focused CLI, connector, and engine surface tests, including unsafe-mode,
  valid-preflight, soft-delete, source-identity, and no-fabricated-REST proofs;
- `go vet ./internal/cli ./internal/connectors/...`;
- `go run ./cmd/connectorgen validate internal/connectors/defs` (552 bundles,
  zero findings) and `surface-sync --check` (zero corrections);
- fresh `go build -o ./bin/pm ./cmd/pm`, sanctioned CLI docs generation,
  sanctioned website-data generation, connector-doc validation, exact command
  parity, generated diff review, and normalized hash proof that no website
  connector entry other than PostgreSQL changed;
- `git diff --check`, inline manual review, and GSD prompt/evidence lifecycle.

The supplied live PostgreSQL command was attempted once. Docker stalled in the
shared harness's image-store-capacity probe before PostgreSQL startup; a
read-only `docker info` probe also stalled. The supervisor identified machine
saturation and waived this non-required documentation lane; it was not retried.
The aggregate CLI/connector sweeps were also launched under the saturated host;
CI remains the authoritative full-suite gate.

## Parity and scope

- Checked `pm help connectors`, bare `pm connectors`, `pm connectors --help`,
  PostgreSQL inspect JSON, and CDC catalog JSON.
- No credential, provider call, generic write tool, dependency, or change to
  #4125, #4136, #4090, or #4154.
- Skills: `golang-how-to`, `golang-cli`, `golang-testing`,
  `golang-error-handling`, `golang-security`, `golang-documentation`, and
  `vercel-react-best-practices`. The routing-required frontend design skills
  are not installed in this runner; existing website patterns were followed.

No automated reviewer was requested by this task; CI is the delivery gate.

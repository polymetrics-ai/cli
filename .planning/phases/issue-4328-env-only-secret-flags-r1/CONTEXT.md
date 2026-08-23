# Context: Issue #4328

## Task Delivery Header

- Issue: Refs #4328 — fix(connectorgen): keep OpenAPI write secrets out of reverse-ETL CLI flags
- Base branch: `main`
- Merges into: `main`
- Delivery: Pull request open against `main`, with the local full `make verify` gate green and the API-reported PR base read back as `main`.
- Working branch: `fm/cli-env-only-secret-flag-generalization-r1`
- Task: Generalize `env_only` generation so every declaration-owned request secret is environment-only independent of protocol, intent, flag type, or mapping depth; preserve established GraphQL behavior; prove CircleCI webhook signing-secret is protected; and report the full connector-definition blast-radius count.
- Verification: Red then green behavioral tests through the actual `connectorgen validate` path against definitions, a deterministic all-definition sweep, GitHub source/descriptor byte-and-SHA measurements, `go test -timeout 20m ./cmd/connectorgen/...`, `make verify`, `git diff --check`, and a direct PR base API read-back.

## Discussion conclusion

The launch brief resolves the only implementation choice: sensitivity is declaration-owned and must be selected from the repository's request-side declaration mechanism, never inferred from protocol, field name, or CLI mapping. The earlier issue text presents a source-importer alternative for a broader reverse-ETL execution capability; this bounded validator correction does not add a write channel or expose a new flag. It applies the existing `env_only` channel policy to every declared request secret and retains current GraphQL mutation protection.

The task runs non-interactively under the Firstmate brief. Generated GSD prompts are executed inline because this direct-PR task forbids role spawning; that is the documented manual-GSD fallback, not a lifecycle waiver.

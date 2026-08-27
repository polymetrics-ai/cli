# Review — Stripe provider-dialect tolerance foundation

## Scope reviewed

- `cmd/connectorgen/sourceimport.go`
- `cmd/connectorgen/sourceimport_test.go`
- `cmd/connectorgen/sourceprojection_test.go`
- retained Stripe source lock, artifact, crosswalk, disposition, and generated
  source projection/evidence
- `docs/migration/conventions.md`
- Phase evidence under this directory

## Review result

No critical, warning, or informational findings remain after the final
current-main review.

The review checked that a retained source lock can only narrow the reference
depth budget; it cannot loosen document resource, grammar, or traversal
limits. It checked source-local `$ref` normalization, memoization, cycle
detection, target-kind checks, and bounded structural validation before a
depth-overrun component is recorded as unreferenced. Hidden malformed,
external, dynamic, cyclic, ambiguous, and resource-exhausting references
remain terminal rather than becoming a source gap.

Only a typed finite-depth response error becomes an operation-local
`source_descriptor` missing-foundation condition. It retains the exact source
operation and citation, invents no request/response/output fields, and stops
at registry/commandrunner preflight before credentials, transport, or an
executor can be reached. The retained Stripe regression confirms all 589
locked operations are unique and source-cited, and its focused simple GET,
DELETE, and nested-reference cases cover the intended behaviour.

The final regression and generator gates recorded in `VERIFICATION.md`,
including `make lint` and `make connector-boundary`, passed cleanly.

The manual GSD review fallback is necessary because this runtime cannot run the
canonical isolated review role and the project contract forbids spawning it.

# TDD ledger — cli-engine-secret-input-r1

Phase uses the manual GSD programming-loop fallback recorded in `PLAN.md`. Every behavior task
must have a real failing focused test run before its production edit.

## secret-input-parse

Status: red-confirmed → green-confirmed

Red command:

```text
go test ./internal/connectors/commandrunner -run 'TestResolveSecretInputs' -count=1
```

Red result: the focused test failed to compile because `SecretInputRequest`, `SecretInputField`,
and `ResolveSecretInputs` did not exist. This was run before production code was added.

The test covers source references that do not carry secret values in argv-derived data, an
undeclared target, forbidden inline input mode, a missing environment value, and an apply failure
that attempts to echo its input.

Green command:

```text
go test ./internal/connectors/commandrunner -run 'TestResolveSecretInputs' -count=1
ok  \tpolymetrics.ai/internal/connectors/commandrunner
```

## secret-input-typing

Status: red-confirmed → green-confirmed

Red command:

```text
go test ./internal/connectors/commandrunner -run 'TestResolveSecretInputsOnlyAllowsDeclaredStringBodyTargets' -count=1
```

Red result: the resolver accepted a declared `config.*` target and called its sink. The test proves
that a typed secret input must be constrained to a declared request-body field before it can be
materialized.

Green result: the resolver rejects non-`body.*`, duplicate, malformed, and non-string declarations
before resolving an environment source or invoking its sink.

## secret-input-leak-safety

Status: red-confirmed → green-confirmed

The initial red compile run covered the leak test before `ResolveSecretInputs` existed. Green tests
prove that an apply callback and a stdin reader may both attempt to return text containing the
sentinel, while the public error remains fixed and value-free. The argv-derived `FromEnv` slice is
also asserted to contain only the field/environment reference, never the resolved value.

Mutation check: temporarily returning the downstream apply error directly made
`TestResolveSecretInputsUsesNonInlineSourcesAndNeverLeaksOnApplyFailure` fail with the fixed
diagnostic `ResolveSecretInputs error leaked the secret`. The safe error boundary was restored and
the focused suite returned green.

## zendesk-nine-op-metadata

Status: planned

Planned red tests: exactly the nine named Zendesk operations expose the typed source surface; an
ordinary command does not.

## storage-seam-binding

Status: deferred-by-decision

Deferred by firstmate key `secret-storage-seam-collision` until connector mechanism foundations
lands. No production storage implementation may start before that rebase.

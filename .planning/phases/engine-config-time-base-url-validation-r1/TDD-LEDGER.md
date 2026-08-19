# TDD LEDGER — engine configuration-time spec-constraint validation

Manual GSD fallback; the adapter does not expose `programming-loop`. Every
behaviour change begins with an executed red test before the corresponding
production edit.

## Planned red-to-green cases

| ID | Behaviour | Red evidence required | Green evidence required |
| --- | --- | --- | --- |
| R1 | Invalid `base_url` is rejected when a credential is added | `AddCredential` accepts `base_url=not-a-uri` despite GitHub's `format: uri` spec declaration | error names `base_url` and `format uri`; no credential persists |
| R2 | Patterns are configuration-time constraints | a config field violating an actual bundle pattern is accepted | pattern name/text and field appear in the error; valid value persists |
| R3 | Enums are configuration-time constraints | a config field outside an actual bundle enum is accepted | enum and field appear in the error; a declared value persists |
| R4 | All declared format siblings use the same engine path | invalid date/date-time/URI-shaped values are accepted | every actual format family is refused at add time |
| R5 | Constraint-free inputs are unchanged | not applicable — this guards a non-regression | a connector/property with no relevant constraint accepts the same value as before |
| R6 | Rejection is pre-persistence | current app flow could reach vault/state after validation failure | a failing validation leaves `ListCredentials()` empty and no stored credential |
| R7 | Promoted native definition forwarder can honour constraints | a constrained `definitionConnector` does not expose the optional validator | the wrapper delegates to its `engine.Base` and rejects the declared enum |

## Run log

### R1 — real configuration boundary reproduction

Status: red-confirmed

No production code had been edited when this test was added and run:

```text
$ go test ./internal/app -run '^TestAddCredentialRejectsInvalidGitHubBaseURLAtConfigurationTime$' -count=1
--- FAIL: TestAddCredentialRejectsInvalidGitHubBaseURLAtConfigurationTime (0.76s)
    app_test.go:153: AddCredential() accepted GitHub base_url that violates spec format uri
FAIL
FAIL    polymetrics.ai/internal/app    1.230s
FAIL
```

The red test uses the real bundled GitHub definition, whose `base_url` has
`"format": "uri"`; it confirms the failure happens at `AddCredential`, before
any runtime request exists.

### Constraint survey after R1

The read-only inventory found 673 top-level constrained configuration fields:
554 `uri`, 81 `date-time`, 20 `date`, 2 patterns, and 16 enums. All constrained
fields are non-secret. The recursive inventory found no declared numeric,
string, array, or object bounds anywhere in the current connector specs.

The next red test group will use real bundle fields for every supported family:
GitHub `base_url` (URI) and `since` (date-time), Google Search Console
`start_date` (date), AgileCRM `domain` and Docker Hub `docker_username`
(patterns), CoinAPI `environment` and Tier-3 Postgres `mode` (enums). It will
also prove an unconstrained GitHub field remains accepted despite unrelated
`required` declarations, guarding against accidentally calling full-schema
validation on the flat credential map.

### R2–R6 — surveyed constraint families and compatibility guard

Status: red-confirmed

Before production edits, the focused app test showed every currently declared
constraint family was accepted at the credential boundary. The focused engine
test did not compile because the deliberately absent configuration-validation
API was named by the tests first.

```text
$ go test ./internal/connectors/engine -run '^(TestSchemaValidateConfigurationAppliesDeclaredConstraintsOnly|TestSchemaWithoutConfigurationConstraintsIsNotAdvertised)$' -count=1
# polymetrics.ai/internal/connectors/engine [polymetrics.ai/internal/connectors/engine.test]
configuration_validation_test.go:25:10: sch.HasConfigurationConstraints undefined
configuration_validation_test.go:55:15: sch.ValidateConfiguration undefined
FAIL    polymetrics.ai/internal/connectors/engine [build failed]

$ go test ./internal/app -run '^(TestAddCredentialRejectsInvalidGitHubBaseURLAtConfigurationTime|TestAddCredentialRejectsDeclaredConfigurationConstraintsAtConfigurationTime|TestAddCredentialLeavesConstraintFreeConnectorUnconstrained)$' -count=1
--- FAIL: TestAddCredentialRejectsInvalidGitHubBaseURLAtConfigurationTime
    AddCredential() accepted GitHub base_url that violates spec format uri
--- FAIL: TestAddCredentialRejectsDeclaredConfigurationConstraintsAtConfigurationTime
    date-time: AddCredential(github.since) accepted "not-a-date-time"
    date: AddCredential(google-search-console.start_date) accepted "2026-02-30"
    agilecrm pattern: AddCredential(agilecrm.domain) accepted "not.allowed"
    docker hub pattern: AddCredential(dockerhub.docker_username) accepted "Uppercase"
    engine connector enum: AddCredential(coin-api.environment) accepted "preview"
    tier three base enum: AddCredential(postgres.mode) accepted "preview"
FAIL    polymetrics.ai/internal/app
```

`TestAddCredentialLeavesConstraintFreeConnectorUnconstrained` was green in
the same run, preserving the control case before the new engine seam exists.

### Green transition — focused matrix

Status: green-confirmed

The implementation compiles declared formats into the engine schema, applies
format/pattern/enum checks only to supplied configuration fields, and reaches
that validator through the optional connector contract before vault or state
mutation. The focused red tests now pass:

```text
$ go test ./internal/connectors/engine -run '^(TestSchemaValidateConfigurationAppliesDeclaredConstraintsOnly|TestSchemaWithoutConfigurationConstraintsIsNotAdvertised|TestConnectorConfigurationConstraintContractReflectsDeclaration)$' -count=1
ok      polymetrics.ai/internal/connectors/engine

$ go test ./internal/app -run '^(TestAddCredentialRejectsInvalidGitHubBaseURLAtConfigurationTime|TestAddCredentialRejectsDeclaredConfigurationConstraintsAtConfigurationTime|TestAddCredentialLeavesConstraintFreeConnectorUnconstrained)$' -count=1
ok      polymetrics.ai/internal/app
```

The engine contract test proves an unconstrained bundle advertises
`HasConfigurationConstraints=false`; the application control test proves the
real constraint-free Faker connector remains accepted without coercing its
flat configuration strings.

### R7 — promoted-native forwarder audit

Status: red-confirmed

The registry registers promoted native connectors through
`native/nativeset.definitionConnector`, which embeds the original connector and
stores an `engine.Base` only for the bundled definition. Before a forwarding
method exists, a constrained wrapped bundle cannot advertise the optional
validator:

```text
$ go test ./internal/connectors/native/nativeset -run '^TestDefinitionConnectorForwardsDeclaredConfigurationConstraints$' -count=1
--- FAIL: TestDefinitionConnectorForwardsDeclaredConfigurationConstraints
    definitionConnector does not expose ConfigurationConstraintValidator
FAIL    polymetrics.ai/internal/connectors/native/nativeset
```

The fix must delegate to `Base`'s actual declaration signal, preserving the
constraint-free current promoted connectors as genuinely unconstrained.

### R7 — green transition

```text
$ go test ./internal/connectors/native/nativeset -run '^TestDefinitionConnectorForwardsDeclaredConfigurationConstraints$' -count=1
ok      polymetrics.ai/internal/connectors/native/nativeset

$ go test ./internal/app -run '^(TestAddCredentialRejectsInvalidGitHubBaseURLAtConfigurationTime|TestAddCredentialRejectsDeclaredConfigurationConstraintsAtConfigurationTime|TestAddCredentialLeavesConstraintFreeConnectorUnconstrained)$' -count=1
ok      polymetrics.ai/internal/app
```

The promoted-native forwarder now delegates both the actual declaration signal
and validation to `engine.Base`; its constraint-free bundles return false from
that signal instead of receiving a synthetic success.

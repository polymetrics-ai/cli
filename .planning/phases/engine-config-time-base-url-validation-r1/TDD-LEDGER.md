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

## Run log

Pending R1. This ledger is created before the red test and production work.

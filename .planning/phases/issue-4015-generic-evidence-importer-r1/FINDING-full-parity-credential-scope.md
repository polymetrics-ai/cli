# Finding — Full-parity credential scope is asserted, not verified

Refs #4015

## Observation

The accepted-evidence writer stamps `credential_scope: "full_parity"` and the
matching required note unconditionally. The acceptance validator then accepts
those same constants; it does not receive a distinct proof that the credential
was verified as full parity.

A GitHub `--direct-read-only` certification returns before the full-run tail
stages, so its report has no full-parity stage and `FullParityVerified()` is
false. It can nevertheless reach the proof/evidence writer under a caller
attestation. That would make an accepted record claim a verified full-parity
credential although this bounded read-only report did not prove that fact.

## Impact on #4015

The fresh GitHub App read-only runs completed all 25 currently
`eligible_pending_live` commands (23 REST, 2 GraphQL), but no GitHub accepted
evidence is published by this change while the scope claim is unverified.
Potential records for `operation:rest_read` and `operation:graphql_query` are
the affected records; neither is committed.

This is not unique to the withheld GitHub slice. The twelve PostgreSQL
transport records were created by the live built-binary test with `--full
--write`, rather than `--full-parity`. The runner therefore reached the normal
tail stages but did not append `full_parity`; `FullParityVerified()` is false
for that report. The transport evidence writer nevertheless passed
`CredentialFullParity: true` unconditionally. The two subsequently published
PostgreSQL CDC records use the same unconditional evidence writer and have no
`FullParityVerified()` report input at all. The runs and their independent
read-backs are real; the record-level full-parity credential assertion is not
verified by the present contract.

## Required follow-up

Create a separately reviewed credential-scope proof contract that derives the
accepted scope from a real verification, or supplies a distinct, validated
bounded-scope identity. Do not weaken the existing accepted-evidence validator
or convert a caller assertion into proof. Re-enable publication only after the
record can state a scope that has actually been verified.

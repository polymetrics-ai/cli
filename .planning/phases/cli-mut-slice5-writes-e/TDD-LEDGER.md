# TDD Ledger — GitHub mutation certification slice 5 writes-e

## Manual GSD fallback

The repository-local Pi adapter prompts were generated and reviewed. This isolated
terminal runtime cannot run Pi subagents, so the required lifecycle was completed
inline. No production connector source was changed.

## Red / green contract

**Red:** A command exit status is not mutation proof. Issue #4221 demonstrates that
a delete can report success while leaving the provider object present.

**Green:** A `certified` command has a real plan/preview/run exchange, an independent
provider read-back of the produced value, terminal disposal by the provider API, an
independent containment read-back, and a schema-v2 record accepted by
`certification-matrix --check`.

## Final corrected attempt ledger

The ordinal lists below are disjoint and exhaustive over `1..145`. They supersede
the provisional 50-command checkpoint totals. The certified set is exactly the 26
schema-v2 evidence records added by this branch over
`origin/integration/4015-mvp-flat-r1`.

| Bucket | Count | Ordinals |
| --- | ---: | --- |
| `certified` | 26 | 25, 27, 28, 50, 98, 99, 104, 105, 106, 109, 111, 112, 113, 114, 124, 126, 127, 128, 129, 133, 134, 135, 136, 137, 140, 141 |
| `no_object` | 13 | 7, 12, 19, 43, 46, 47, 48, 49, 88, 93, 110, 115, 142 |
| `wrong_credential` | 0 | — |
| `entitlement` | 19 | 3, 6, 8, 9, 10, 11, 13, 14, 17, 18, 20, 22, 23, 24, 77, 83, 117, 120, 144 |
| `not_implemented` | 0 | — |
| `product_defect` | 82 | 1, 2, 15, 16, 26, 29, 30, 31, 32, 33, 34, 35, 36, 37, 38, 39, 40, 41, 42, 44, 45, 51, 52, 53, 54, 55, 56, 57, 58, 59, 60, 61, 62, 63, 64, 65, 66, 67, 68, 69, 70, 71, 72, 73, 74, 75, 76, 78, 79, 80, 81, 82, 84, 85, 86, 87, 89, 90, 91, 95, 96, 97, 100, 101, 102, 103, 107, 108, 116, 118, 119, 121, 122, 123, 125, 130, 131, 132, 138, 139, 143, 145 |
| `escape_needs_captain` | 5 | 4, 5, 21, 92, 94 |
| **Total attempted** | **145** | **145** |

## Standing-correction audit

### `no_object`

Every retained `no_object` row first read the parent collection. Where the
collection was empty, fixture creation was attempted. The retained rows are limited
to object classes the provider would not let this contained certification identity
create:

- 7, 12, 19: user Codespaces secret fixture creation was rejected because GitHub
  requires a provider-key sealed value; the deliberately non-secret fixture was
  rejected and the collection remained empty.
- 43, 46: GitHub exposes no REST creation endpoint for review-suggestion objects;
  the real parent issue existed, but no disposable suggestion could be minted.
- 47–49: the repository has no issue-field definition, and the repository REST
  surface cannot create that configuration; collection/fixture attempts returned
  404 or 422.
- 88, 93: the organization Dependabot-secret collection was read, but GitHub
  rejected the non-secret encrypted fixture and the collection remained empty.
- 110: GitHub returned 405 for both read and mutation of a repository-level pull
  request cap, so no provider object could be created.
- 115: the bypass endpoint requires a real detected-secret placeholder. The parent
  collection was empty and GitHub rejected the tagged placeholder fixture.
- 142: GitHub rejected the tagged attestation fixture because a trusted Sigstore
  bundle cannot be minted by this non-OIDC local identity.

All earlier Copilot-space, team, ordinary repository-secret, environment-variable,
and other fixture-backed provisional `no_object` rows were converted because their
parent object class was creatable. A failure to complete independent command proof
after that recovery attempt is recorded as `product_defect`, never banked as an
absent object.

### `wrong_credential`

No final row remains `wrong_credential`. Every provisional credential failure was
checked against the measured alternate route in
`github-credential-matrix-2026-08-18.tsv`. Endpoints unavailable to both measured
routes are `entitlement`; command failures with a provider control are
`product_defect`.

## Product-defect controls

Raw `api.github.com` controls distinguish provider behavior from PM behavior. The
repeated classes include:

- `class=integer_id_scientific_notation`: ordinals 1, 2, 15, 16, 45, 51, 52,
  54, 56, 57, 60, 62, 64, 67, 68, 70, 72, 74, 87, 90, 132, 138, and 139. The
  fleet control using the exact integer succeeded, read back the effect, and was
  directly disposed; each subsequent instance was tagged once and not re-derived.
- `class=delete_false_success`: delete commands that reported a successful PM run
  without establishing the required provider state change. Raw controls and
  independent reads were used for the command family; these are never certified
  from PM exit status.
- Missing or unusable request fields: 44, 53, 55, 58, 59, 65, 95, 96, 102, 107,
  108, 116, 122, 123, 125, 130, 131, and 145. In each class, PM omitted or rejected
  a provider-required value while the raw control reached GitHub successfully and
  was independently contained.
- `class=ghsa_id_typed_integer`: 118, 119, and 121 reject canonical GHSA string
  identifiers before request dispatch; the raw string-path controls reached the
  provider.
- `class=anonymous_endpoint_authenticated`: 143 always authenticates an endpoint
  that GitHub requires to be anonymous; the anonymous raw control returned 202
  using only fake credentials.

## Containment

Direct provider DELETE plus an independent 404/empty read-back was used wherever
GitHub offers deletion. For immutable or non-deletable resources the provider's
actual terminal disposal was used and recorded (`contained_closed`, disabled,
inactive, restored preference, or immutable terminal status). Cleanup alone never
upgraded command correctness.

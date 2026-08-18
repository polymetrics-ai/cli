# GitHub command certification defect fixes — delivery summary

## Census

The pre-implementation census covers every one of the 277 supplied command
entries; see `CENSUS.md` for command-level membership.

| Root-cause class | Commands | Share |
| --- | ---: | ---: |
| integer ID scientific notation | 147 | 53.1% |
| missing request payload | 56 | 20.2% |
| wrong API path | 4 | 1.4% |
| false success | 11 | 4.0% |
| other | 59 | 21.3% |
| **Total** | **277** | **100.0%** |

The measured 77% scientific-notation rate among the 39 reason-bearing entries
was therefore not the population rate. The full inventory classifies 147/277
(53.1%) in that class.

## Source fixes

- State JSON uses lossless `json.Number` decoding for interface-backed command
  records, and template interpolation renders all numeric types in decimal form.
- Blob, commit, tree, check-run, branch-protection, and commit-status actions now
  declare provider-required fields and their commands expose typed body flags.
- Three mislabelled Agent commands reuse the corresponding Actions write actions
  and `/actions/` paths. The singular user-project draft route now accepts only a
  numeric user ID, never a login.
- An allow-listed missing delete is accounted as `RecordsUnchanged`, not
  `RecordsWritten`. A connector command that promised a mutation rejects that
  incomplete acknowledgement, exits non-zero, and persists zero succeeded / one
  failed. Closed idempotent cleanup workflows may explicitly accept already
  absent state without claiming a provider write.

## Sanitized live before/after proof

All after-runs used the disposable Keychain identity and the fixed
`Polymetrics-Cert/pm-cert-3993-20260810-wz0fru` scope. No credential, approval
token, public-key material, ciphertext, or raw command transcript is recorded.

| Class | Command | Recorded before | Fixed binary after | Independent read-back |
| --- | --- | --- | --- | --- |
| exact integer | `orgs update-webhook` | hook ID serialized in scientific notation; provider rejected the path | exit 0, completed, 1 succeeded / 0 failed with exact ID `667291475` | GET 200 matched `active=false`, events `issues` |
| exact integer | `orgs update-webhook-config-for-org` | hook ID serialized in scientific notation; provider rejected the path | exit 0, completed, 1/0 with exact ID | GET 200 matched tagged URL, JSON content type, strict SSL |
| exact integer | `orgs delete-webhook` | PM could report success while exact-ID resource remained | exit 0, completed, 1/0 with exact ID | GET returned 404 |
| required body | `git blobs create` | no content flag; empty body returned 422 | exit 0, completed, 1/0 | replayed and read back the pre-existing 30-byte blob SHA |
| required body | `git trees create` | no tree flag; empty body returned 422 | exit 0, completed, 1/0 | replayed and read back the pre-existing one-entry root tree SHA |
| required body | `git commits create` | no message/tree flags; empty body returned 422 | exit 0, completed, 1/0 | replayed and read back a pre-existing tagged unsigned commit SHA |
| corrected path | `agents update-org-variable` | `/agents/variables/{name}` returned 404; raw `/actions/` control returned 204 | exit 0, completed, 1/0 via `/actions/` | Actions getter matched updated tagged value and `selected` visibility |
| corrected path | `agents set-selected-repos-for-org-variable` | `/agents/.../repositories` returned 404 | exit 0, completed, 1/0 via `/actions/` | collection matched the fixture repository ID |
| corrected path | `agents set-selected-repos-for-org-secret` | `/agents/.../repositories` returned 404 | exit 0, completed, 1/0 via `/actions/` | collection matched the fixture repository ID |
| honest result | `orgs create-webhook` | PM reported success while the hook collection remained empty | exit 0, completed, 1/0 | collection contained the tagged hook |
| honest result | `orgs delete-webhook` | PM reported success while GET remained 200 | real delete: exit 0, completed, 1/0; repeated absent delete: exit 1 | persisted failed run had 0 succeeded / 1 failed; GET returned 404 |
| honest result | `actions delete-org-secret` | allow-listed 404s could be counted as writes | real delete: exit 0, completed, 1/0; repeated absent delete: exit 1 with `wrote 0, unchanged 1` | persisted failed run had 0 succeeded / 1 failed; GET returned 404 |

The Git object write proofs created no new cleanup obligation: GitHub Git
objects are content-addressed and have no delete API, so each command replayed
the exact definition of a provider object whose SHA existed before the call.
No ref changed. This avoids inventing an undeletable fixture merely to prove a
body reached the real endpoint.

## Fixture cleanup

- Organization webhook: deleted through `pm`; independent getter returned 404.
- Actions organization variable: deleted through `pm`; independent getter returned 404.
- Actions organization secret: deleted through `pm`; independent getter returned 404.
- Local encrypted certification project: removed after the provider absence checks.

## Not fixed in this PR

All four requested common root-cause classes are fixed. The 59-command `other`
census bucket is **not fixed** here: its members have individualized causes or
insufficient evidence for one of the four source fixes, and changing them under
an inferred shared cause would exceed the measured task. No declared command
was removed, disabled, or marked unsupported.

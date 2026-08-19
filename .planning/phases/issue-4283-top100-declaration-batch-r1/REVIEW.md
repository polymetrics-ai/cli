# Increment 001 review

Manual GSD code-review fallback: the project-local Pi adapter generated the `code-review` command, but this environment cannot run its incompatible reviewer role. The review was performed inline as required by the delivery contract.

## Scope reviewed

- Ten connector-local source locks, empty ledgers where no safe executor/action exists, generated non-live sweep artifacts, and four source/API-surface drift reconciliations.
- Connector documentation updates limited to refreshed source provenance and corrected coverage/hash statements.
- The progress ledger, 3,405 exact rejection records, and the transport foundation-gap record.

## Findings

No Critical, Warning, or Info findings. In particular, review confirmed:

- no files under `defs/github/` or `defs/zoom/` changed;
- no engine/foundation code, credentials, live provider calls, or generic transport declarations were added;
- each newly surfaced GitLab, Notion, Vercel, and Sentry endpoint is disabled with source provenance;
- GitHub-source-lock aggregate schema parity was restored with `counts`, while per-method counts remain recorded; and
- the source-to-surface denominator is 4,378 / 4,378, checked again after the review.

## Transport follow-up review

Path (b) was reviewed after the increment checkpoint. `go test -timeout 20m ./internal/connectors/certify -run '^TestCertificationDeclaredTransportPairFailsWhenRegistrationIsMissing$'` and `go test -timeout 20m ./internal/synctransport -run 'TestRegisterDeclaredTransports'` pass. The ten transport rejections are recoverable and do not weaken the runtime's factory/evidence admission.

## Docker Hub full-parity review

Reviewed the Docker Hub source-contract inventory, exact crosswalk, and 54
per-operation dispositions against the pinned OpenAPI source.

- The 49 `operations.json` entries bind exactly to existing API-surface
  method/path entries but do not claim a terminal CLI route; `surface-sync`
  therefore correctly makes no command-surface change.
- Every DELETE is retained as a `rest_write` inventory contract with
  `mutation_class: delete` and destructive confirmation, while its actual
  terminal route remains disabled with evidence.
- The CSV export is not mislabeled as a JSON direct read; all three HEAD
  operations are explicit, recoverable response-less executor gaps; and the
  source-deprecated login is `provider-does-not-expose` rather than a
  misleading scope failure.
- No capability was promoted, no write action or transport descriptor was
  fabricated, and no file outside Docker Hub definitions and issue #4283
  evidence changed.

No Critical, Warning, or Info findings remain for the Docker Hub proof slice.

## Docker Hub secret-policy retrofit review

Reviewed the captain's credential-endpoint correction against the pinned Docker
Hub OpenAPI and the runnable command contracts.

- The eight personal/organization token list, detail, update, and delete routes
  are enabled only because their pinned response schemas expose metadata; the
  two destructive routes use typed delete actions and the shared reverse-ETL
  lifecycle.
- The two token-create responses expose `token`, while login/2FA/auth-token
  expose `token` or `access_token`. All five are declared source contracts with
  `secret_sensitive`, redacting `sensitive_policy`, source secret-field marks,
  and recoverable `foundation-gap` records. No secret response executor was
  invented.
- `spec.json` marks each source-named credential ingress field `x-secret`.
  Full source request schemas remain in `operations.json`; `writes.json`
  projects only the engine-supported schema subset, as conformance requires.
- Static validation, sweep generation, fixture-backed conformance, commandrunner
  preflight, generated documentation, and all 41 no-credential binary dispatch
  checks pass. `connector-boundary` is explicitly pending CI after two worker
  detached-capture attempts terminated before an exit record.

No new Critical, Warning, or Info finding was identified. The pending boundary
gate is preserved as a required merge check, not treated as a pass.

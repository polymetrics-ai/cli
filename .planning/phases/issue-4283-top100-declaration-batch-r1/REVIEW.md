# Increment 001 review

Manual GSD code-review fallback: the project-local Pi adapter generated the `code-review` command, but this environment cannot run its incompatible reviewer role. The review was performed inline as required by the delivery contract.

## Scope reviewed

- Ten connector-local source locks, empty ledgers where no safe executor/action exists, generated non-live sweep artifacts, and four source/API-surface drift reconciliations.
- Connector documentation updates limited to refreshed source provenance and corrected coverage/hash statements.
- The progress ledger, 3,405 exact rejection records, and the transport foundation-gap record.

## Findings

No Critical, Warning, or Info findings. In particular, review confirmed:

- no files under `defs/github/` or `defs/zoom/` changed;
- no engine/foundation code, credentials, live provider calls, generic writer, or invented destination transport declaration was added;
- each newly surfaced GitLab, Notion, Vercel, and Sentry endpoint is disabled with source provenance;
- GitHub-source-lock aggregate schema parity was restored with `counts`, while per-method counts remain recorded; and
- the source-to-surface denominator is 4,378 / 4,378, checked again after the review.

## Transport follow-up review

After PR #4286, the ten source declarations were reviewed against the real
definition-owned composition. `go test -timeout 20m ./internal/app -run
'^TestOpenRegistersDefinitionOwnedProductionTransports$'` passes with the
loaded declarations. The reverse leg is correctly retained as the one
recoverable `generic-typed-destination-executor` gap: the only declarative
destination is still the closed issue-label adapter. No action binding or
generic writer was fabricated.

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

## Complete six-class map review

Reviewed the nine new source-lock crosswalk/disposition pairs and the Docker
Hub reference summary against the captain's map contract.

- Every pinned source operation is present once in its connector map; the
  primary class totals equal the source-lock count for all nine bundles.
- `ENABLED%` counts only operations with an implemented CLI binding. It does
  not promote an operation inventory, elevated scope, stream, write action or
  source URL into a runnable command.
- Source transport is now declared through the definition-owned #4286 contract.
  Destination transport is the precise `generic-typed-destination-executor`
  gap; no GitHub evidence, action binding, or generic write transport was
  invented.
- The maps retain every documented DELETE and report enabled/documented delete
  counts. No provider call, credential, engine edit, GitHub bundle or Zoom
  bundle change was made.

No Critical, Warning, or Info finding was identified in the complete-map
artifact. Connector-boundary remains pending CI because detached worker capture
still has no observable exit record.

## Vocabulary correction review

The complete-map dispositions were re-reviewed to keep the foundation lane
honest. All 3,889 non-enabled rows from the nine new maps are
`declaration-pending`; their retained minimal change is connector declaration,
not engine work. Docker Hub has been normalized to the same primary-class row
shape. Five source-operation rows remain engine gaps; the ten reverse legs add
one shared typed-destination foundation-gap ID, cited in
`FOUNDATION-GAP-REASONS.json`.

## Classification correction review

The preceding “six primary classes” language is superseded. Review found that
250 ordinary typed write endpoints had been classified as `reverse_etl`, which
incorrectly hid 118 enabled direct-write bindings. The maps now classify those
endpoints as `direct_write` (2,370 total, 118 enabled) and retain reverse ETL
only as a zero-eligible attribute on every direct-write row. The one shared,
recoverable reason is `generic-typed-destination-executor`, with the existing
destination-factory evidence and minimum recovery. No Critical, Warning, or
Info finding remains from this correction.

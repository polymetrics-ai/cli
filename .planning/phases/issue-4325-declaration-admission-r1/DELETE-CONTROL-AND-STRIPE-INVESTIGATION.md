# GitHub delete control and Stripe account-delete investigation

## Scope and result

This follow-up examines `origin/main` at
`060bb7864e3419e09ab10e000bb14ac1ea3724ec` plus this shared
declaration-admission PR. It adds only a GitHub implemented-delete regression
control. It does not convert Stripe or perform any provider I/O.

GitHub demonstrates that DELETE is an executable operation kind when its
endpoint has a typed action and command contract. Stripe's account deletion is
currently an admitted *mapping concern* only: its API surface declares the
official endpoint as blocked, but no action or CLI binding exists to execute
it. The difference is endpoint-specific, not a generic delete restriction.

## GitHub: implemented delete control

| Contract layer | Exact evidence |
| --- | --- |
| Provider source | `internal/connectors/defs/github/sources/github-operation-source-lock.json` cites `https://raw.githubusercontent.com/github/rest-api-description/b26c240ded1c8b79cb0fb09dee4a21239061fa23/descriptions/api.github.com/api.github.com.json`; source ID `github.rest.issues.delete-label` identifies `paths["/repos/{owner}/{repo}/labels/{name}"].delete`, operation ID `issues/delete-label`, and `DELETE /repos/{owner}/{repo}/labels/{name}`. |
| Typed action and schema | `internal/connectors/defs/github/writes.json` action `delete_label` is `DELETE /repos/{{ config.owner }}/{{ config.repo }}/labels/{{ record.name }}` with `path_fields: ["name"]`, `body_type: "none"`, and a closed record schema requiring `name` as a string with `maxLength: 8192`; it declares delete risk text. |
| Generated command/surface | `internal/connectors/defs/github/cli_surface.json` maps `label delete` to `write: "delete_label"`, intent `reverse_etl`, availability `implemented`, source CLI path `gh label delete`, and API surface `DELETE /repos/{owner}/{repo}/labels/{name}`. |
| Behavioral evidence | `internal/connectors/defs/github/fixtures/writes/delete_label.json` asserts a `DELETE /repos/synthetic-conformance-value/synthetic-conformance-value/labels/bug` request. `internal/connectors/operation-evidence.json` records `write:delete_label` and passed `write_request_shape:delete_label` conformance. `internal/connectors/defs/github/certification.json` pairs `create_label` with `delete_label` cleanup. |
| Credential boundary | In a freshly initialized local project with no credentials, `pm github label delete --json` resolves the command then refuses with `missing --credential` (exit 1). This is dispatch evidence only; it sends no provider request. |
| This PR's regression | `TestDeclarationAdmissionAdmitsGitHubImplementedDeleteControl` builds the exact cited declaration, admits it, then calls `commandrunner.Preflight(engine.New(bundle, nil), []string{"label", "delete"})`. The test proves an implemented delete has a canonical source declaration and an executable runtime preflight without credentials or provider I/O. |

`internal/connectors/defs/github/certification-sweep.json` still records the
REST command's live certification state as `fixture_required`; that is not a
credential-bound or live-provider proof. The regression control deliberately
does not claim either certificate.

## Stripe: account delete is an endpoint-specific gap

| Contract layer | Exact evidence |
| --- | --- |
| Provider mapping | `internal/connectors/defs/stripe/api_surface.json` has `DELETE /v1/accounts/{account}` with `operation: "destructive_action"`, `status: "blocked"`, `risk: "high"`, `blocked_by_default: true`, `source_url: "https://stripe.com/docs/api"`, and source notes for `operation_id=DeleteAccountsAccount`, `summary=Delete an account`, `source=https://raw.githubusercontent.com/stripe/openapi/master/openapi/spec3.json`, `spec_version=2026-07-29.dahlia`. |
| Explicit missing contract | That same endpoint's `reason` says it is blocked until a typed destructive reverse-ETL action with `confirm=destructive`, plan/preview/approval/execute evidence, idempotency notes, and fixtures is authored. |
| Missing runtime/action/CLI binding | `internal/connectors/defs/stripe/writes.json` has no `delete_account` action. Its only DELETE action is `delete_customer` at `/customers/{{ record.id }}`, with a closed schema requiring an ID matching `^cus_[A-Za-z0-9_]+$`, an explicit destructive confirmation, and 404 idempotency handling. `internal/connectors/defs/stripe/cli_surface.json` exposes `customers delete` for that action but no `accounts delete` command or binding to `/v1/accounts/{account}`. |
| Safe executable check | `./pm stripe accounts delete --json` returns `error: unknown command "accounts delete"` with the structured `usage_error` classification. It performs no provider I/O. |

Stripe cannot execute the account endpoint as a generic DELETE today. The CLI
does not expose a generic HTTP write/delete executor; `commandrunner` resolves
only declared command paths. Adding a raw delete route would bypass the
endpoint's missing typed record/path contract, destructive confirmation,
idempotency policy, fixture, and reverse-ETL plan/preview/approval/execute
gate. The existing mapping therefore remains valuable for declaration
completeness, while implementation must wait for the named endpoint-specific
runtime contract.

## Certificate boundary

The new admission checker asks whether a cited source operation has one
canonical declaration and whether an `implemented` row has the runtime binding
it claims. It does not promote Stripe's blocked API-surface mapping to an
implemented command, and it does not treat a DELETE method as deferred by
default. Runtime preflight, credential-bound proof, and live certification
remain separate certificates.

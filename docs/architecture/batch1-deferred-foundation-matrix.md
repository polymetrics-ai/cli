# Batch 1 deferred foundation matrix

This is a source-ledger projection for the ten Batch-1 connectors. It classifies
every **genuinely blocked** provider operation exactly once; it is not a claim
that a deferred operation is executable, certified, or safe to invoke. The
machine-readable authority, including every operation record and every retained
source URL/location, is
[batch1-deferred-foundation-matrix.json](batch1-deferred-foundation-matrix.json).

## Result

The current source denominator is 4,341 operations. The existing three-way
accounting is 767 runnable, 1,666 declarable now, and **1,908 genuinely
blocked**. This matrix contains all 1,908 blocked rows, no duplicate record
keys, and **zero unclassifiable rows**.

| Connector | Deferred operations |
| --- | ---: |
| Docker Hub | 50 |
| Notion | 7 |
| Stripe | 581 |
| Bitbucket | 111 |
| GitLab | 835 |
| CircleCI | 20 |
| Sentry | 76 |
| Vercel | 139 |
| Asana | 37 |
| Jira | 52 |
| **Total** | **1,908** |

## Selection and uniqueness rule

An operation is included if either its provider source import stops before a
descriptor exists, or its retained descriptor has a foundation other than
`closed-source-operation-execution-foundation-r1` and
`source-cited-non-executable-mutation-foundation-r1`. Exact source bindings to
currently `implemented` commands are excluded. The exception is Docker Hub's
two crosswalk-bound direct reads that still fail runtime preflight: they remain
deferred, leaving its four working stream bindings as the only exclusions.

For descriptor-backed connectors, the join is source method/path, after the
documented GitLab `/api/v4` base-path normalization. This avoids confusing a
stale or generated source ID with the provider operation identity. Each JSON
record preserves both the descriptor source ID and the joined declaration
ledger evidence.

Each record has one `primary_group`, chosen in this precedence order:

1. provider source/importer foundation;
2. CLI discovery/generation foundation;
3. typed request contract/body/schema foundation;
4. other source foundation.

All source gap entries remain in the record's `missing_foundations` array. The
precedence makes the aggregate mutually exclusive; it does not hide a
secondary problem on a single operation.

This is the same definition and denominator recorded in
`.planning/phases/issue-4325-batch1-connectors-repair-r1/TDD-LEDGER.md:379-398`.
It deliberately does not classify a row as blocked merely because it is a
write, a delete, binary, or sensitive. Those are operation traits and may
affect the contract needed to make a command runnable, but they are not a
generic foundation reason.

## Primary categories

| Category | Exact count | Meaning |
| --- | ---: | --- |
| Provider source/importer | 740 | The provider shape cannot yet become a complete descriptor without the cited importer contract. |
| CLI discovery/generation | 1 | A provider route cannot safely become a generated command binding. |
| Typed contract/body/schema | 1,167 | The source descriptor exists, but a request/body/media shape cannot become a closed operation contract. |
| Action/write/transport | 0 | Rows with only the declaration-only action dispositions are counted in the 1,666 declarable bucket, not here. |
| Binary policy | 0 | Binary is retained as a per-record trait; no selected row has an independent binary-policy foundation. |
| Sensitive input | 0 | Sensitivity is retained as source evidence; no selected descriptor has a sensitive-input foundation as its primary gap. |
| Other | 0 | No operation was left without a normalized group. |

### Normalized foundation groups

The lanes below are the source-ledger `parity_class` counts. Representative
links are the provider's pinned official source URL; the JSON file records the
complete provider location and every operation, not just these examples.

| Group | Count / connector breakdown | Lanes | Current owner evidence | Representative source records |
| --- | --- | --- | --- | --- |
| Docker Hub unresolved provider reference | 50 — Docker Hub 50 | binary read 1; direct read 22; direct write 27 | Connectorgen source importer: `cmd/connectorgen/sourceimport.go:5576-5578`, `:5888-5901` | [AuditLogs_ListAuditActions](https://docs.docker.com/reference/api/hub/latest.yaml) `GET /v2/auditlogs/{account}/actions`; [AuthCreateAccessToken](https://docs.docker.com/reference/api/hub/latest.yaml) `POST /v2/auth/token` |
| Notion schema-depth refusal | 7 — Notion 7 | binary write 1; direct write 6 | Connectorgen source importer: `cmd/connectorgen/sourceimport.go:4673-4675` | [create-a-database](https://developers.notion.com/openapi.json) `POST /v1/data_sources`; [introspect-token](https://developers.notion.com/openapi.json) `POST /v1/oauth/introspect` |
| Stripe reference-depth refusal | 581 — Stripe 581 | binary read 4; direct read 254; direct write 323 | Connectorgen source importer: `cmd/connectorgen/sourceimport.go:5572-5578` | [DeleteAccountsAccount](https://raw.githubusercontent.com/stripe/openapi/master/openapi/spec3.json) `DELETE /v1/accounts/{account}`; [DeleteAccountsAccountBankAccountsId](https://raw.githubusercontent.com/stripe/openapi/master/openapi/spec3.json) `DELETE /v1/accounts/{account}/bank_accounts/{id}` |
| `cli-recursive-schema-foundation-r1` | 98 — Bitbucket 77; Jira 11; Vercel 10 | binary read 1; direct read 19; direct write 30; ETL 48 | Connectorgen source importer: `cmd/connectorgen/sourceimport.go:2360-2365`, `:5570-5574` | [Bitbucket repositories workspace](https://developer.atlassian.com/cloud/bitbucket/swagger.v3.json) `GET /repositories/{workspace}`; [Vercel operation source](https://openapi.vercel.sh/) (see JSON source IDs) |
| `cli-openapi30-reference-sibling-foundation-r1` | 2 — Asana 2 | direct write 1; ETL 1 | Connectorgen source importer: `cmd/connectorgen/sourceimport.go:5164-5167` | [createMembership](https://raw.githubusercontent.com/Asana/openapi/56796a67a3c093eedf55fd9682357957a2ebfd85/defs/asana_oas.yaml) `POST /memberships`; [getMembership](https://raw.githubusercontent.com/Asana/openapi/56796a67a3c093eedf55fd9682357957a2ebfd85/defs/asana_oas.yaml) `GET /memberships/{membership_gid}` |
| `cli-malformed-path-parameter-foundation-r1` | 2 — GitLab 2 | binary read 1; direct write 1 | Connectorgen source importer: `cmd/connectorgen/sourceimport.go:6636-6646` | [getApiV4JobsIdSbomScansSbomScanId](https://gitlab.com/gitlab-org/gitlab/-/raw/master/doc/api/openapi/openapi_v3.yaml) `GET /api/v4/jobs/{id}/sbom_scans/{sbom_digest}` |
| `cli-operation-route-override-foundation-r1` | 1 — Sentry 1 | direct read 1 | Connectorgen source importer: `cmd/connectorgen/sourceimport.go:6152-6164` | [listSeerModels](https://raw.githubusercontent.com/getsentry/sentry-api-schema/main/openapi-derefed.json) `GET /api/0/seer/models/` |
| `cli-request-encoding-foundation-r1` | 51 — Asana 1; GitLab 47; Jira 1; Sentry 2 | binary write 1; direct write 50 | Connectorgen source importer: `cmd/connectorgen/sourceimport.go:6791-6798` | [createAttachmentForObject](https://raw.githubusercontent.com/Asana/openapi/56796a67a3c093eedf55fd9682357957a2ebfd85/defs/asana_oas.yaml) `POST /attachments`; [postApiV4BulkImports](https://gitlab.com/gitlab-org/gitlab/-/raw/master/doc/api/openapi/openapi_v3.yaml) `POST /api/v4/bulk_imports` |
| `cli-request-media-selection-foundation-r1` | 1 — Jira 1 | direct write 1 | Connectorgen source importer: `cmd/connectorgen/sourceimport.go:6791-6795` | [setPreference](https://dac-static.atlassian.com/cloud/jira/platform/swagger-v3.v3.json) `PUT /rest/api/3/mypreferences` |
| `cli-request-schema-foundation-r1` | 1,115 — Asana 34; Bitbucket 34; CircleCI 20; GitLab 786; Jira 39; Sentry 73; Vercel 129 | direct read 25; direct write 1,090 | Connectorgen source importer: `cmd/connectorgen/sourceimport.go:6809-6826` | [createAccessRequest](https://raw.githubusercontent.com/Asana/openapi/56796a67a3c093eedf55fd9682357957a2ebfd85/defs/asana_oas.yaml) `POST /access_requests`; [deleteRedirects](https://openapi.vercel.sh/) `DELETE /v1/bulk-redirects` |
| Action/write/transport declaration-only | 0 | — | Connectorgen source projection: `cmd/connectorgen/sourceprojection.go:213-240` | None: these are declarable-now rows, so excluding them preserves the measured split. |
| Binary policy | 0 | — | Connectorgen CLI validation: `cmd/connectorgen/validate.go:2813-2842` | None: response media remains in the complete JSON records rather than being mislabelled as a separate foundation. |
| Sensitive input | 0 | — | CLI validation: `cmd/connectorgen/validate.go:903-928`; engine bundle validator: `internal/connectors/engine/bundle.go:3177-3213` | None: source-declared sensitive traits are recorded per operation; no generic sensitive-input deferral is invented. |

## Delete is not a deferred category

GitHub is the positive control. `pm github repo delete` is an implemented,
behaviorally tested DELETE workflow: its local test asserts no HTTP request at
plan, preview, or unconfirmed execution; exactly one DELETE after human preview,
approval-token stdin, and `--confirm destructive`; and no replay after the
grant is spent. See `internal/cli/github_repo_delete_cli_test.go:14-152` and
`internal/connectors/commandrunner/github_write_contract_test.go:181-203`.

Therefore, this matrix never treats a delete as deferred simply because it is a
delete. A deferred delete is grouped only by its concrete source/importer,
route, typed action/body/schema, or other exact foundation. For example, Stripe
account deletion remains under the Stripe reference-depth importer group here;
it still needs a typed operation-specific destructive action and command binding
after that importer can produce a complete descriptor.

## Complete operation data

`batch1-deferred-foundation-matrix.json` contains 1,908 `operations` records,
one per `(connector, source.id)`. Every record includes:

- the provider source ID, method, path, official URL, source location, source
  lock path, and descriptor path when available;
- the matched declaration-ledger lane, operation model, covered-by reference,
  and declaration state;
- its mutually exclusive primary group, all retained source gaps, and source
  evidence for mutation/delete, binary-policy, and sensitive-input traits; and
- the current rejection/foundation object from the declaration ledger.

The data is read-only documentation. It does not rewrite a source lock,
descriptor, connector definition, or command surface.

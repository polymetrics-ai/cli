# Slice 2 manual direct run

All commands in this receipt were invoked directly through the built `pm` binary against the disposable GitHub repository.  For each accepted direct read, the agent-derived assertion was: the response is a non-null JSON object or array (and therefore rejects a plausible wrong answer such as `null`, a scalar, or a malformed envelope).  Its schema-v2 evidence record was written immediately and accepted by `go run ./cmd/connectorgen certification-matrix --check`.

| Command | Outcome | Assertion / provider result |
| --- | --- | --- |
| actions artifact-and-log-retention view | certified | agent_derived: non-null object; response reported integer `days` and `maximum_allowed_days`. |
| actions concurrency_groups view | certified | agent_derived: non-null object; response reported `concurrency_groups` array and `total_count`. |
| actions fork-pr-workflows-private-repos view | certified | agent_derived: non-null object. |
| actions organization-secrets view | certified | agent_derived: non-null object. |
| actions organization-variables view | certified | agent_derived: non-null object. |
| actions public-key view | certified | agent_derived: non-null object. |
| actions sub view | certified | agent_derived: non-null object. |
| actions usage view | certified | agent_derived: non-null object; response reported cache usage fields. |
| actions retention-limit view | entitlement | GitHub: `http 402` for `/actions/cache/retention-limit`. |
| actions selected-actions view | entitlement | GitHub: `http 409` for `/actions/permissions/selected-actions`. |
| actions storage-limit view | entitlement | GitHub: `http 402` for `/actions/cache/storage-limit`. |
| interactions get-pull-request-creation-cap-for-repo | product_defect | GitHub: `http 405` for `/interaction-limits/pulls/creation-cap`; endpoint/method surface requires correction. |
| import large_files view | product_defect | GitHub: `http 404`; provider states the endpoint is deprecated and directs callers to GitHub Importer. |
| import view | product_defect | GitHub: `http 404`; provider states the endpoint is deprecated and directs callers to GitHub Importer. |
| installation view | wrong_credential | Runtime: `provider rejected the credential` with both classic and fine-grained credentials. |

Pre-request `missing path variable` and `unresolved path template` failures from the retired runner are superseded by direct calls with the supplied path constants; they are not `no_object` evidence.

The bounded structural exception supplied by the fleet audit is product work, not a provider result: commands requiring `{enterprise-team}` with no matching flag are uninvokable.  The affected GitHub command family is `enterprise-team-memberships` (add, bulk-add, bulk-remove, get, list, remove) and `enterprise-team-organizations` (add, bulk-add, bulk-remove, delete, get-assignment, get-assignments).  No retry is appropriate until the command surface exposes that path input.

`actions create-org-variable` is also a product defect in the contained write path: the real command created `rplan_b6ec67ad91b44496`, and preview resolved `POST /orgs/Polymetrics-Cert/actions/variables`, but it remained `planned` with zero preview time and emitted no `Approval token:` line.  It therefore cannot reach the required approval-and-execute stage.

`dependabot list-alerts-for-org` is a controlled product defect: at the same time, with the same classic credential and `Polymetrics-Cert` organization, the direct provider GET `/orgs/Polymetrics-Cert/dependabot/alerts` returned HTTP 200 with `[]`, while `pm github dependabot list-alerts-for-org --org Polymetrics-Cert` returned HTTP 400.  GitHub accepts the request; the connector command malforms it.

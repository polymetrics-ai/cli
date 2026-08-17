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

`actions create-org-variable` is not a product defect.  Its JSON preview correctly produced a no-network dry-run for `POST /orgs/Polymetrics-Cert/actions/variables`, including an execution digest and approval target.  JSON intentionally omits the approval token and does not itself authorize a provider write.  Per the repository's approval contract, no agent execution was attempted; a human approval token is required before the disposable provider mutation can be performed and subsequently deleted/read back.

`dependabot list-alerts-for-org` is a controlled product defect: at the same time, with the same classic credential and `Polymetrics-Cert` organization, the direct provider GET `/orgs/Polymetrics-Cert/dependabot/alerts` returned HTTP 200 with `[]`, while `pm github dependabot list-alerts-for-org --org Polymetrics-Cert` returned HTTP 400.  GitHub accepts the request; the connector command malforms it.

## Continuing direct batch

The subsequent direct batch used the same disposable credential and asserted `status: 200` plus a non-null JSON object or array, rejecting `null`, scalar values, malformed envelopes, and any non-200 response.  Evidence for each passing row was schema-v2 validated immediately.

| Command group | Outcome |
| --- | --- |
| Actions organization permission, fork-workflow, self-hosted runner, custom-image, and runner-application reads | certified where GitHub returned `200`; individual schema-v2 records carry the captured exchange. |
| Agents organization/repository secret, variable, and public-key reads | certified. |
| Billing organization and disposable-user usage report reads | certified where GitHub returned `200`. |
| Interactions organization and authenticated-user restriction/capability reads | certified. |
| GitHub-hosted runner images, limits, machine sizes, platforms, and list reads | entitlement: GitHub returned `404` with `GitHub hosted runners are not supported for this organization`. |
| Actions selected-repository permission reads | entitlement: GitHub returned `409` for this organization configuration. |
| Actions allowed-actions organization read | entitlement: GitHub returned `409` for `/actions/permissions/selected-actions`. |
| Billing budgets read | entitlement: GitHub returned `400` for the organization billing-budgets endpoint. |
| Code-quality findings read | entitlement: GitHub returned `403` and stated `Code quality is not enabled for this repository.` |

`compare view --basehead main...main` is a controlled product defect.  The command rejects the required `basehead` form before sending a request (`path variable basehead must not contain path traversal`).  At the same time, a raw HTTPS `curl --http1.1` request directly to `https://api.github.com/repos/Polymetrics-Cert/pm-cert-3993-20260810-wz0fru/compare/main...main`, using the same classic credential through stdin-only curl configuration, returned HTTP `200` with `status: identical`.  This bypasses every Polymetrics layer.  GitHub accepts self-comparison on the repository's default `main` branch; the connector's path safety rule incorrectly rejects GitHub's documented comparison delimiter.

# Bounded live results — generated GitHub direct-read candidates

These are real, serial invocations by the disposable certification identity
against the approved `Polymetrics-Cert/pm-cert-3993-20260810-wz0fru` fixture.
Credentials were supplied only through `--from-env`; no secret, successful
response body, or accepted-evidence record is retained here.

## Scope accounting

| Trial family | Declared | Generated direct reads executed | Reverse-ETL deferred |
| --- | ---: | ---: | ---: |
| Advanced Security | 50 | 31 | 19 |
| Copilot | 50 | 23 | 27 |
| Enterprise | 46 | 21 | 25 |
| Codespaces | 46 | 22 | 24 |
| **Total** | **192** | **97** | **95** |

The 95 reverse-ETL commands are deliberately not candidates in this task. They
need the separate mutation-fixture lifecycle: owned fixture, plan, preview,
approval, execution, independent read-back, cleanup, and evidence import.

## Candidate provenance

The generator wrote **97** candidates. The existing **23 hand-authored** direct
read candidates remain deliberately more specific overrides for these named
commands: `repo read-file`, `readme view`, `branches view`, `git ref view`,
`commits view`, `commits status view`, `commits statuses view`, `activity view`,
`assignees view-2`, `events view`, `topics view`, `community profile view`,
`hash-algorithm view`, `teams view`, `actions access view`, `actions caches
view`, `actions downloads view`, `actions permissions view`, `actions workflow
view`, `actions runners view`, `actions secrets view`, `actions variables
view`, and `rulesets rule-suites view`.

Each override asserts a fixture-specific semantic value that `cli_surface.json`
does not declare (for example a known README, ref, or named resource). The
generic generator can express their structural object-or-array assertion, but
cannot honestly synthesize those produced values from the declared surface;
the merge rule preserves their stronger manual assertions. There is no unnamed
escape hatch: a manual candidate shadows only its exact declared command.

## One terminal result per executed read

| Cohort | Produced-value pass | Product defect | Provider / missing-fixture non-pass | Total |
| --- | ---: | ---: | ---: | ---: |
| `trial_advanced_security` | 15 | 9 | 7 | 31 |
| `trial_copilot` | 6 | 0 | 17 | 23 |
| `trial_enterprise` | 5 | 0 | 16 | 21 |
| `trial_codespaces` | 8 | 1 | 13 | 22 |
| **Total** | **34** | **10** | **53** | **97** |

“Produced-value pass” means the operation was executed live and its generated
`/response` object-or-array assertion passed. It is **not certification**:
the accepted evidence importer is concurrently owned elsewhere, so none of the
34 is represented as `live_tested` or published by this delivery.

Every remaining executed read has a concrete non-pass outcome. Provider 404s
for an intentionally absent sub-resource are missing-fixture results, not
passes and not entitlement conclusions. Provider policies/refusals remain
provider results rather than product defects.

## Prominent product-defect findings (10)

The sweep marks these separately as `product_defect`; each declaration maps a
required REST path parameter to a CLI flag that is incorrectly optional:

- `code-scanning analyses view-2` — `analysis_id` / `--analysis-id`
- `code-scanning autofix view` — `alert_number` / `--alert-number`
- `code-scanning databases view-2` — `language` / `--language`
- `code-scanning instances view` — `alert_number` / `--alert-number`
- `code-scanning repos view` — `codeql_variant_analysis_id` /
  `--codeql-variant-analysis-id`
- `code-scanning sarifs view` — `sarif_id` / `--sarif-id`
- `code-scanning variant-analyses view` — `codeql_variant_analysis_id` /
  `--codeql-variant-analysis-id`
- `dependabot secrets view-2` — `secret_name` / `--secret-name`
- `secret-scanning locations view` — `alert_number` / `--alert-number`
- `codespaces secrets view-2` — `secret_name` / `--secret-name`

These are explicitly not folded into the provider non-pass total.

## Measured provider evidence

- `code-scanning analyses view` executed and GitHub returned HTTP 404 with
  `{"message":"no analysis found","documentation_url":"https://docs.github.com/rest/code-scanning/code-scanning","status":"404"}`.
  Code scanning is enabled and reachable; this is a missing-analysis fixture,
  not an entitlement block.
- `code-scanning default-setup view` executed and passed its produced-value
  assertion. `secret-scanning list-alerts-for-org` also executed and passed.
- Copilot organization metrics returned HTTP 403 with
  `{"message":"The 'Copilot usage metrics' policy must be enabled to use this API","documentation_url":"https://docs.github.com/copilot/concepts/copilot-metrics","status":"403"}`.
  That is a provider policy refusal, not a product defect.
- `copilot get-copilot-seat-details-for-user` returned HTTP 404 with
  `{"message":"No seat found for this user in this organization.","documentation_url":"https://docs.github.com/copilot/managing-copilot/managing-github-copilot-in-your-organization/managing-access-to-github-copilot-for-members-of-your-organization/granting-access-to-copilot-for-members-of-your-organization","status":"404"}`.
  This is a missing-seat fixture result.
- Copilot Spaces item and resource reads returned HTTP 404 `Not Found` for the
  declared placeholder IDs. The list endpoints executed; item/resource
  fixtures have not been created by this read-only delivery.

## Failure demonstration

After candidate generation, the generated assertion for the known live-passing
`copilot configuration view` was deliberately changed from `object_or_array`
to `string` and the real certification runner was executed. Its own named
stage went red with:

```
generated_direct_read_copilot_configuration_view: declared output at /response has the wrong type
```

The generator was restored, both generated artifacts were regenerated, and the
same named stage was executed again with `passed: true`.

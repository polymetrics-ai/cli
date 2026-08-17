# GitHub direct-read certification throughput — live results

## Result

Throughput is **flat at 100-operation batches**. Three fresh 100-operation runs averaged **154.831 seconds / 1.548 seconds per target operation** (range 150.351–158.521 seconds). That is below the fresh, setup-heavy 10-operation cohort (25.528 seconds / 2.553 seconds per operation), not degradation.

Execution is therefore not the present throughput bottleneck. Publication is still blocked solely by #4211 / PR #4215's unverified credential-scope contract; this lane staged no accepted certification record.

## Method

- A disposable \`polymetrics-ai-certification\` identity was supplied only by environment-variable reference. Its value was never in argv, output, a file, or this report; the ambient \`gh\` login was not used.
- The temporary candidate source was unmerged PR #4214 at \`7306b9ec3e079b51ac9c70a674605a3a27f6e09b\`. The 10-operation cohort used its first ten generated direct reads. The 100-operation cohort used all 97 generated candidates plus three existing direct-read overrides.
- Each timer ran from a fresh \`pm init\` through credential validation, serial direct reads, checkpoint/report persistence, and teardown. Binary build was measured separately. A pass required the candidate's \`/response\` \`object_or_array\` assertion; process exit status alone never counted.
- The three 100-operation repeats used fresh projects and no checkpoint reuse. Their failed command sets matched exactly, so their safe replay taxonomy is the same as the first 100-operation batch.

## Measured curve

| Cohort | Target reads | Wall incl. setup/teardown | Mean/op | Pass | Product defect | Missing fixture | Provider refusal | Resumed |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| 10 | 10 | 25.528 s | 2.553 s | 2 | 4 | 0 | 4 | 0 |
| 100 fresh | 100 | 158.521 s | 1.585 s | 33 | 14 | 6 | 47 | 0 |
| 100 repeat 1 | 100 | 155.621 s | 1.556 s | 33 | 14 | 6 | 47 | 0 |
| 100 repeat 2 | 100 | 150.351 s | 1.504 s | 33 | 14 | 6 | 47 | 0 |
| 100 mean (n=3) | 100 | 154.831 s | **1.548 s** | 33 | 14 | 6 | 47 | 0 |

The original runner knew ten existing sweep product defects. Safe replays proved four additional \`enterprise-team\` failures are local unresolved path templates, so the final 100-operation taxonomy is \`33 + 14 + 6 + 47 = 100\`. An ambiguous GitHub \`404 Not Found\` remains a provider refusal, not proof that a fixture is absent.

Separate source-build times were 4.660 seconds for the 10-operation cohort and 55.371, 27.235, and 25.071 seconds for the three 100-operation cohorts. A normal campaign reuses a built binary; rebuilding every 100 operations adds a measured mean 35.892 seconds per cohort and is setup cost, not read throughput.

## Rate-limit evidence

Certification reports intentionally retain safe, derived rate facts—not raw headers, URLs, bodies, or secrets. Each \`reset_at\` comes from a real provider reset-header observation in \`internal/connectors/connsdk/http.go\`.

| Cohort | Attempt events | Reset-header observations | Limiter waits | Limiter not-sent | Retry attempts above first |
| --- | ---: | ---: | ---: | ---: | ---: |
| 10 | 9 | 11 | 0 | 0 | 0 |
| each 100 | 101 | 93 | 0 | 0 | 12 |

No timed run emitted a limiter \`wait\` or \`not_sent\` event, or a rate-limit reason. The 12 extra attempts in each 100-operation batch were requester retries on provider failures. The safe audit artifact cannot prove an absent raw \`Retry-After\` header, but it proves that the limiter did not wait or refuse a send. Therefore the honest conclusion is: **no GitHub throttle point was reached in 310 timed target stages (10 + 3×100); no threshold can be named from this run.** The small-batch premium is fixed project/validation/report setup, not a measured rate-limit mechanism.

## Resumability probe

An interrupted 100-operation attempt resumed from its checkpoint with 17 resumed stages and 34 checkpoint-completed terminal stages rather than restarting at operation one. Its 135.793-second wall is excluded from the fresh curve because it contains pre-resume work. This proves the required restart behavior without contaminating the curve.

## Failure evidence

The six provider-evidenced missing fixtures said: two \`No alert found for alert number 1\`, \`no analysis found\`, two \`No code security configuration found for id 1\`, and \`No seat found for this user in this organization.\` The 47 provider refusals comprise 30 ambiguous 404s and 17 non-404 responses (two 402, nine 403, one 409, one 500, two 400, two 503). Where PM printed \`[redacted]\`, this report preserves that redaction rather than inventing provider text.

Product defects:

| Command | Defect |
| --- | --- |
| \`code-scanning analyses view-2\` | missing \`analysis_id\` |
| \`code-scanning autofix view\` | missing \`alert_number\` |
| \`code-scanning databases view-2\` | missing \`language\` |
| \`code-scanning instances view\` | missing \`alert_number\` |
| \`code-scanning repos view\` | missing \`codeql_variant_analysis_id\` |
| \`code-scanning sarifs view\` | missing \`sarif_id\` |
| \`code-scanning variant-analyses view\` | missing \`codeql_variant_analysis_id\` |
| \`codespaces secrets view-2\` | missing \`secret_name\` |
| \`dependabot secrets view-2\` | missing \`secret_name\` |
| \`secret-scanning locations view\` | missing \`alert_number\` |
| \`enterprise-team-memberships get\` | unresolved \`{enterprise-team}\` template |
| \`enterprise-team-memberships list\` | unresolved \`{enterprise-team}\` template |
| \`enterprise-team-organizations get-assignment\` | unresolved \`{enterprise-team}\` template |
| \`enterprise-team-organizations get-assignments\` | unresolved \`{enterprise-team}\` template |

## Projection from the live curve

The re-read current GitHub sweep ledger has **639 implemented direct reads**; it has 655 direct-read rows total, with 15 \`unsupported_api\` and one \`unsupported_local\`. This lane does not reduce publication remaining count, because it withheld evidence.

- Implemented surface: \`639 × 1.548311 s = 989.871 s\` = **16 minutes 30 seconds**.
- All 655 rows if the unsupported 16 become runnable: \`1,014.144 s\` = **16 minutes 54 seconds**.
- Seven 100-sized cohorts with a deliberate rebuild each time add \`7 × 35.892 s\` = **4 minutes 11 seconds**, an avoidable setup allowance.

These are measured-seconds-per-operation projections, not a percentage guess.

## Rules for the rulebook

1. Use fresh, serial 100-operation direct-read cohorts. Include setup, validation, checkpoint/report I/O, and teardown in the wall; use ten only as a configuration smoke, not a campaign forecast.
2. Use the connector's declared rate-limit configuration and a disposable environment-provided credential. Never use the ambient CLI login.
3. Persist checkpoint state before treating a terminal stage as done. On an interruption or limiter wait, resume the same project and report resumed count separately; do not mix its wall into a fresh-batch curve.
4. Honor a provider \`Retry-After\` exactly. Save only event type, duration, reason, and derived reset time—never raw credentials or response bodies.
5. Treat a limiter \`wait\` or \`not_sent\` event as a new throughput regime: stop the fresh curve there, wait/resume, and report that regime separately.
6. Count only a passed produced-value assertion. Separately classify provider refusal, provider-evidenced missing fixture, and product defect; ambiguous 404s are provider refusals.

## Publication status

\`STAGED-LIVE-REPORT.json\` holds the safe, non-accepted payload and deliberately omits a \`credential_scope\` claim. Nothing was published to the matrix. Import it through PR #4216 only after #4211 / PR #4215 fixes and verifies the scope contract.


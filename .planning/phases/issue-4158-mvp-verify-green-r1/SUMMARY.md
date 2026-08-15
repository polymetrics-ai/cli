# Investigation Summary — Production MVP verify blocker

## 1. Reproduction

At `ef3c71caf`, run:

```bash
go test -timeout 20m ./internal/cli -run '^TestFreshBinaryDeclarativeGitHubWarehouseFlowRoundTrip$' -count=2 -v
```

The test builds a fresh 152,132,226-byte `pm` binary for each run, assembles a
real project with the declarative GitHub definition, durable warehouse, reverse
plan, approval, and local faithful GitHub HTTPS API, then invokes `pm flow run`.
Both runs exit **3**; the result is repeatable. The test's token checks passed
and no credential value is retained here.

## 2. Trigger, mask, and symptom

- **Trigger:** #4170's action step has no `job` but serializes the full inline
  action scope.
- **Masking condition:** before #4168, direct flow runs did not resolve action
  jobs. #4168 added `resolveManifestJobs`, which requires a `rplan_…` job and
  derives all action scope from the already-approved reverse plan. The fixture
  was added later and still uses the former shape.
- **Visible symptom:** the freshly built external binary fails before flow
  dispatch with exit 3 and `validation/flow_job_reference_refused`; the reason
  is the empty action job being malformed. There is no provider write or
  durable flow receipt.

## 3. Divergent path against a proven path

The failing path is:

`flow run` → `resolveManifestJobs` → action `step.Job == ""` → typed malformed
job refusal.

The proven path is identical until the action step: a valid `rplan_…` job is
resolved, its approved source/destination/mappings/authorization are derived,
and the real action runner reaches acknowledgement, independent read-back, and
receipt persistence. `git log -S'resolveManifestJobs'` and blame locate the
new path in `5c12fb536` (#4168). `#4150` and `#4155` both precede it.

## 4. Smallest counterfactual

Only change the temporary fixture's action step to set `job` to the already
created reverse-plan ID and leave `action_cfg` with `read_back_stream` alone.
The fresh-binary test passes in 42.85 seconds: one record is synced, one action
is acknowledged, one warehouse row is observed, a checkpoint and receipt are
durable, and the replay/unapproved/auth/unsafe refusal checks remain true. The
temporary change was reverted.

## 5. Disconfirming evidence

The stated shared-cause explanation is false. The failing flow has GitHub and
warehouse endpoints only; it does not construct a PostgreSQL managed-target
driver, database write plan, or history route. The PostgreSQL #4158 live test
was separately invoked with its build tag but correctly skipped because the
required explicit container opt-in and endpoint are not available. Its outcome
is therefore unproven here, but cannot explain this fresh-binary refusal.

## Decision required

This lane must not silently update the fixture or weaken its assertion. Firstmate
must select one:

1. **Fixture migration:** update #4170's manifest to the job-only #4168
   contract and add its happy/bad/edge tests.
2. **Compatibility restoration:** change the public flow resolver to accept
   legacy inline action scope, which needs a new non-PostgreSQL ownership and
   safety plan.

# Discussion log — certification batch scaling

## Fixed constraints from the captain brief

- Read operations only; no GitHub mutation or fixture creation. Publication is allowed only after the verified scope contract lands and through its supported importer.
- Use the disposable certification identity through an environment variable only; do not use ambient `gh` authentication.
- Serialize requests and preserve checkpoint state so a throttle can resume without replaying completed reads.
- A successful process is not a pass: a candidate must assert a produced `/response` value.
- Provider refusal, missing fixture, and product defect are separate terminal outcomes.
- The throughput curve must include setup, credential validation, checkpoint/report work, and teardown.

## Resolved implementation choices

- PR #4214 is open but its head is fetchable. The measurement will execute its candidate-projection source unchanged in a disposable copy, pinning the exact SHA in every result. This avoids both waiting for the merge and absorbing its owned changes into this PR.
- The projection holds 97 generated direct-read candidates. A one-run 100-operation manifest supplements them with three existing declaration-owned direct-read overrides; all remain read-only and assertion-bearing.
- There will be one 10-operation run, one 100-operation run, then at least three identical 100-operation runs. More batches are allowed only if a rate event needs a resumed continuation.
- The report will distinguish an observed throttle from a non-observation. It will not invent a throttle threshold merely from GitHub documentation or a projected request count.

## Publication gate

PR #4215 merged while this lane was running. The rebased base now verifies the
scope construction, but the landed `connectorgen certification-evidence`
command exposes only `transport` and `change-capture`; the generic importer is
still owned by open PR #4216. The lane will retain a sanitized, bounded staged
input rather than hand-author accepted evidence or present a partial
direct-read run as `full_parity`.

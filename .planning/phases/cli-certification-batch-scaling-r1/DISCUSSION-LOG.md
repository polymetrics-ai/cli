# Discussion log — certification batch scaling

## Fixed constraints from the captain brief

- Read operations only; no GitHub mutation, fixture creation, or publication.
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

`credential_scope: full_parity` is unverified on the integration base. This lane will retain only a sanitized, non-accepted live-report input. It will neither call an evidence importer nor generate a matrix/evidence artifact until the #4211 contract is actually present and verified.

# TDD ledger — certification batch scaling

## Cycle 1 — accountable rate-limit configuration

- **Red:** The first harness invocation was refused before sending a provider request: `rate-limit policy "authenticated-user" requires non-secret config "rate_limit_account" for its declared scope`. This was a safe control, not a credential failure.
- **Green:** The harness supplied only the declared non-secret account label, plus the disposable environment credential reference. The 10-operation run and three fresh 100-operation runs then completed and emitted safe response-derived rate events.
- **Refactor/result:** No connector code, bundle declaration, credential value, or accepted evidence record changed. The harness keeps the credential as an environment-variable *name* and never logs or writes its value.

## Cycle 2 — bounded manifests

- **Red:** The source candidate projection would reject an under-specified candidate; its generated candidate check was run against pinned PR #4214 source before live use.
- **Green:** The deterministic manifests contained exactly 10 and 100 direct reads, respectively. Generated candidates had exactly one `/response: object_or_array` assertion; the 100-operation manifest was 97 generated candidates plus three existing direct-read overrides.

## Cycle 3 — live scaling curve

- **Red:** A `connectors certify` process returned exit code 2 for non-passing stages. The harness did not count the process exit as success; it counted only terminal stages that passed their `/response` assertion.
- **Green:** The 10-stage result sums to `2 produced-value + 4 product defects + 0 fixture-missing + 4 provider refusals`. The fixed 100-stage manifest was independently run three times and, after replaying its failed direct reads for safe provider text, sums to `33 + 14 + 6 + 47 = 100` each time. The repeated manifests have identical failure command sets. Timed walls include project setup, credential validation, report/checkpoint I/O, and teardown.

No existing test is weakened, skipped, or mode-excluded.

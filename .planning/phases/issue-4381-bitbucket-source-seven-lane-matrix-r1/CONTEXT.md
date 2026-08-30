# Context — Bitbucket Track A

- Parent issue: #4325. Child issue: #4381.
- Base parent branch and commit: `fm/cli-top100-declaration-batch-r1` / `dc481bac1a8b78d60ac0b4a2b0dfd1a9068ce8db`.
- Start comment was posted to #4381 before project changes, naming branch `feat/4381-bitbucket-track-a-matrix-r1`.
- Immutable source denominator: 297 locked Bitbucket REST rows.
- Crosswalk boundary: 331 method/path identities; 34 crosswalk-only and zero lock-only after method/path reconciliation.
- Source-authoritative ETL signal: a successful response schema resolved through the retained source contract that declares both a string `next` continuation link and an array `values`, yielding 73 retained rows; no general GET-to-ETL inference. This intentionally includes `search_result_page` for `searchTeam`, `searchAccount`, and `searchWorkspace` without relying on schema spelling.
- Source-authoritative direct/write signal: the first action in the locked provider summary classifies a bounded read or mutation. HTTP method and source ID remain cited facts, but neither is the lane selector.
- Current definitions are not provider-fact authority and not execution proof for this Track A matrix.

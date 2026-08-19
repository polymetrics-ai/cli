# #3897 Delivery Summary

**Status:** Local implementation and verification complete; no-mistakes,
push, draft PR, and CI observation remain.

## Deliverables

1. Query manifests use the existing `connection` field to scope every bare
   DuckDB warehouse view to one owner.
2. Action manifests use `action_cfg.source_connection`; action source rows use
   the action-specific uncapped `ActionSourceReadRequest.Connection` before
   the existing local action runner.
3. `_unattributed` selects root-owned tables only. Omitted duplicate sources
   fail with `*warehouse.AmbiguousTableError`, decorated only with a manifest
   field that flow users can supply.
4. JSON round-trips and the runner boundary retain selected identity. This
   path owns no query/action preview digest; #3994 remains responsible for the
   later provider-action lifecycle.
5. Runtime help, generated manual, website documentation, and golden help
   transcripts describe the new manifest syntax without adding a CLI flag.

## Coverage

coverage:
  automated:
    - focused real-Parquet flow query/action tests for separate acme/globex rows
    - correction regression for 101 selected action rows and failed-checkpoint safety
    - ambiguity and root-owned selector tests
    - manifest JSON and action-runner identity tests
    - app/flow/CLI package tests and app/flow race tests
    - fresh built-binary query proof asserting returned rows
  manual:
    - source diff review against #3897 exclusions
    - runtime help, bare namespace, and command-help parity checks
  human_judgment: false

## Constraints honored

- No connector action dispatch, provider mutation, reverse lifecycle,
  scheduling, rate policy, transport work, generic HTTP/SQL write, or live
  provider call was added.
- Both temporary proof roots were moved to Trash and their original paths were
  verified absent.
- Correction rounds: **1 / 5**.

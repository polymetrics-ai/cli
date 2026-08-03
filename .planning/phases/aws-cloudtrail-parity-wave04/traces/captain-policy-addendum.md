<!-- pm-aws-cloudtrail-wave04-r1-captain-policy -->
<!-- pm-aws-cloudtrail-wave04-r1-scope-correction -->
Captain policy addendum for `fm/cli-aws-cloudtrail-parity-wave04-r1`:

- Official AWS CloudTrail API action ledger inventories 60 documented operations exactly once.
- Scope-corrected implemented classification: 19 ETL/read streams, 0 bounded direct/provider query commands, 0 typed reverse-ETL write/admin actions, 0 binary operations, 0 CDC operations, 41 blocked/planned operations.
- Blocked/planned breakdown: 10 provider query/direct-read actions and 31 write/admin actions require typed operation/write metadata plus shared promoted-native command-surface, write-validation, dry-run, and operation-direct-read forwarding outside this connector-local branch.
- Event record contents page remains schema evidence only: current event record version `1.11` with 31 top-level event fields; those fields are not counted as API operations or CDC feeds.
- Safety contract: no raw AWS action, path, header, body, shell, file, SQL, or generic HTTP escape hatch; implemented read streams map declared stream names to fixed CloudTrail `X-Amz-Target` values only.
- Reverse-ETL CloudTrail writes are not exposed in this corrective head. Any future write/admin enablement must preserve plan -> preview -> explicit approval -> execute, destructive confirmation metadata, and fixed typed JSON-RPC actions.
- Local verification is fixture/replay only; no live AWS provider calls or credentialed connector checks were run for this branch.

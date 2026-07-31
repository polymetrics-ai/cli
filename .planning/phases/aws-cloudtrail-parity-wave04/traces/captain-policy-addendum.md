<!-- pm-aws-cloudtrail-wave04-r1-captain-policy -->
Captain policy addendum for `fm/cli-aws-cloudtrail-parity-wave04-r1`:

- Official AWS CloudTrail API action ledger is implemented as 60 documented operations exactly once.
- Classification: 19 ETL/read streams, 10 bounded direct/provider query or lookup operations, 31 typed reverse-ETL write/admin operations, 0 binary operations, 0 CDC operations, 0 excluded operations.
- Event record contents page remains schema evidence only: current event record version `1.11` with 31 top-level event fields; those fields are not counted as API operations or CDC feeds.
- Safety contract: no raw AWS action, path, header, body, shell, file, SQL, or generic HTTP escape hatch; native runtime maps declared names to fixed CloudTrail `X-Amz-Target` values only.
- Reverse-ETL writes use closed top-level schemas and the existing plan -> preview -> explicit approval -> execute flow; destructive/admin actions declare confirmation metadata and provider delete idempotency where applicable.
- Local verification is fixture/replay only; no live AWS provider calls or credentialed connector checks were run for this branch.

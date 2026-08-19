# Zoom direct-write/delete — Wave #4268 summary

No direct-write command can be honestly delivered on this foundation revision. The source operations are valid typed REST declarations, but the endpoint ledger cannot cover an operation-backed REST direct write: coverage currently accepts a reverse-ETL write action or fixed GraphQL operation only. Creating synthetic reverse-ETL actions would be a false model.

The connector-local rejection record preserves all 61 exact affected REST operation IDs, including 18 deletes, as recoverable `foundation-gap` work. Eight upload/multipart operations are separately rejected as `schema-incompatible` because their provider-required redirect/media constraints have no valid post-foundation declaration. The foundation-gap log names the smallest required shared changes; this wave does not implement them.

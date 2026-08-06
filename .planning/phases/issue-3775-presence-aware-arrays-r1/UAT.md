# UAT — issue #3775 presence-aware required string arrays

Automated UAT passed; no human judgment is required because each acceptance criterion is a
deterministic runtime representation or rejection behavior.

- D1: The common required-flag path accepts explicit blank zero-minimum arrays but rejects absent,
  raw-empty, scalar-blank, and cardinality-invalid inputs.
- D2: The public operation direct-read path sends `body.items` as typed `[]string{}` and serialized
  `{"items":[]}`.
- D3: The public reverse-ETL planning path returns `record.items` as typed `[]string{}` and
  serialized `{"items":[]}`, remains approval-required, and leaves `Write` uncalled.

No UAT action used a live provider, credential, or reverse-ETL execution.

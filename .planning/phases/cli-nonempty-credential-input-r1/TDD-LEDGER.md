# TDD ledger — provider-neutral non-empty credentials

| Slice | Red | Green | Refactor / evidence |
| --- | --- | --- | --- |
| Stdin transport | Pending: CLI accepts empty/newline-only stdin or broadly trims terminal bytes before App persistence. | Pending: one LF/CRLF transport delimiter is removed; empty normalized input returns a typed validation error before persistence. | Pending: assertions expose only length and SHA-256 fingerprint. |
| App persistence | Pending: direct `App.AddCredential` can pass an empty map value to `vault.Put`. | Pending: every provided secret field is non-empty before encrypted persistence. | Pending: existing secretless connector credentials remain allowed. |
| Auth emission | Pending: selected blank credentials can form Bearer/basic/API-key header/query authentication. | Pending: selected required auth refuses the blank material before request mutation. | Pending: optional declaration-selected `none` auth remains valid. |

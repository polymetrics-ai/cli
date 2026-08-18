# TDD ledger — GitHub slice 4 live certification

| Step | Evidence |
| --- | --- |
| Red: evidence contract | A record without the required schema-v2 proof fields is rejected by `certification-matrix --check`; no incomplete record was retained. |
| Green: live read evidence | Each certified read executed through `pm github ... --connection pm-cert-classic --json`, asserted a response property, and was persisted as a uniquely named schema-v2 record. |
| Green: matrix admission | `go run ./cmd/connectorgen certification-matrix --check` passed after the evidence records were written. |
| Green: cleanup | Direct GitHub read-backs confirmed zero user keys, signing keys, GPG keys, social accounts, blocks, and follows after contained probes. |

No production behavior changed; this is a live-evidence delivery record.

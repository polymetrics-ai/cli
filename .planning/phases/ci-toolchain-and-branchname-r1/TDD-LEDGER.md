# TDD ledger — CI toolchain and integration branch guard

| ID | Guarantee | Red evidence | Green proof |
| --- | --- | --- | --- |
| R1 | The security scan uses a scanner-clean standard library. | Go 1.25.12 scan reports seven reachable standard-library vulnerabilities, all fixed in 1.25.13. | Go 1.25.13 scan completes with `No vulnerabilities found.` |
| R2 | There is one project toolchain across explicit and `go.mod`-derived workflows. | Active workflows and `go.mod` currently pin 1.25.12. | Every active pin and `toolchain` directive reads 1.25.13; the two `go-version-file: go.mod` callers inherit it. |
| R3 | The integration parent remains guarded by issue-linked structure. | `integration/4015-mvp-flat-r1` does not match the conventional branch rule. | An explicit `^integration/[1-9][0-9]*-[a-z0-9][a-z0-9._-]*$` arm accepts it and rejects zero/missing/non-numeric issue numbers and invalid slugs. |
| R4 | Existing valid and invalid branch classifications do not change. | Existing `fm/*` exceptions and conventional grammar are the baseline. | Existing exception case and conventional pattern are left byte-for-byte intact; only the new anchored integration arm is added. |

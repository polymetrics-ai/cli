# TDD ledger — GitHub certification suite r1

| Slice | Red evidence | Green evidence | Refactor / result |
| --- | --- | --- | --- |
| Surface-driven sweep generation | Pending: add a `cmd/connectorgen` test for the absent `certification-sweep` command and its expected 1,571-record, exact-one accounting artifact. | Pending. | The command must load the bundle surface, not a second hand-authored command list. |
| Product-defect classification | Pending: the same test expects `releases assets view` to appear under product defects because required REST `asset_id` maps to a non-required CLI `asset-id`. | Pending. | The comparison is generic over loaded operation metadata; shared Go never names GitHub. |
| Non-pass/provider-refusal contract | Pending: validator tests reject an unexecuted `pass`, a missing reason, a duplicate path, and a provider refusal without concrete provider status. | Pending. | All non-executed rows remain non-pass with machine-readable reasons. |
| Certified assertion failure proof | Pending after generator green: scratch-change one declaration-owned `/response/...` assertion after bundle validation and run that candidate red. | Pending: restore exact source and rerun that candidate green. | The scratch edit is neither committed nor evidence of a provider result. |
| Evidence integration | Blocked: #4198 is open and `http_exchanges` is not available. | Deferred until its capture contract lands. | No accepted evidence record is fabricated. |

No existing test will be weakened, skipped, or mode-excluded.

# Discussion log — issue #3792

| Topic | Decision | Evidence |
| --- | --- | --- |
| Can #3792 begin before #3788? | Yes. The runtime has a loaded `provider_search` kind and executor already; #3788 is declaration/evidence work, not a prerequisite for no-network operation admission. | `engine/bundle.go:1499-1538`; `engine/direct_read.go:32-138`; #3792 scope. |
| Can it avoid #3740's active conflict? | Yes, if metadata admission is implemented in `engine/direct_read.go` and consumed only by `validateOperationDirectReadCommand`. | PR #3740 changed-path inventory includes `connectors.go` and `engine/connector.go`, not `engine/direct_read.go`; task ownership. |
| How is operation data obtained? | From the loaded bundle through a closed, no-network preflight method; no generic raw operation access is added. | Existing write analogue: `commandrunner/runner.go:545-579`; `engine/direct_write.go:155-184`. |
| How is output content handled? | Deferred to #3771/#3852. This phase compares an operation's declared policy with the command policy without adding or invoking stripping. | task ownership; #3771; #3852. |

The generated `scripts/gsd prompt discuss-phase 3792` was executed inline as a manual fallback;
the fixed decisions above are the required discussion artifact.

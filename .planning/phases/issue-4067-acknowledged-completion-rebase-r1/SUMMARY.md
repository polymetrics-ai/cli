---
phase: issue-4067-acknowledged-completion-rebase-r1
status: local_delivery_ready
coverage:
  - id: D1
    description: An acknowledged transport run completes after an unrelated writer without replaying work, and the returned completed run matches the durable reopened run.
    verification:
      - kind: unit
        ref: "internal/app/transport_dispatch_test.go: TestRunETLTransportAcknowledgedCompletionRebasesUnrelatedStateForAllModes"
        status: pass
      - kind: unit
        ref: go test -count=3 -timeout 20m ./internal/app focused completion set
        status: pass
    human_judgment: false
  - id: D2
    description: A changed, missing, or terminal target cannot be overwritten; its ordinary revision-conflict error remains detectable.
    verification:
      - kind: unit
        ref: "internal/app/transport_dispatch_test.go: TestRunETLTransportAcknowledgedCompletionFailsClosedWhenTargetChanges"
        status: pass
    human_judgment: false
  - id: D3
    description: Cancellation and state-store outcomes remain truthful across the acknowledged-checkpoint boundary.
    verification:
      - kind: unit
        ref: "internal/app/transport_dispatch_test.go: TestRunETLTransportCancellationAfterAcknowledgedCheckpointForAllModes"
        status: pass
      - kind: unit
        ref: "internal/app/transport_dispatch_test.go: TestRunETLTransportAcknowledgedCompletionReturnsTruthfulPersistenceOutcome"
        status: pass
    human_judgment: false
  - id: D4
    description: #4046 typed-conflict finalization and R7/R8 source identity/per-stream CAS are unchanged.
    verification:
      - kind: unit
        ref: internal/app transport R7/R8 and #4046 focused regression command
        status: pass
      - kind: unit
        ref: go test -race -count=3 -timeout 20m ./internal/app focused completion set
        status: pass
    human_judgment: false
---

# #4067 summary

The focused transport correction and both Sol-reported generated outputs are locally reviewed and verified. Independent Sol audit finding F1 rejected immutable candidate `883a86cf0040d559edcd4777413d1c2de20cd94a`: ordinary successful final completion leaves an acknowledged transport run durably `running` when an unrelated writer advances the whole project revision. The all-seven-mode real JSON-store witness first reproduced that leak, then passed after the transport branch captured its own acknowledged target stream and allowed a latest-state completion only if the still-running target exactly matched it. Changed/missing/terminal targets fail closed with a detectable revision conflict; cancellation remains durable in all modes; definite versus committed/indeterminate persistence results return truthfully; focused race and #4046/R7/R8 regressions pass. Canonical generators refreshed only the candidate-owned website transport-eligibility output and certification source locations. The manual standard-depth review found no unresolved behavior, scope, test, or generated-artifact issue. The first no-mistakes run reached `checks-passed`, but its automatic delivery opened forbidden duplicate #4068 and its document step expanded into an unrelated warehouse rewrite; #4068 was closed unmerged and the rewrite was removed. The second 2/5 local no-mistakes run passed review/test/document/lint with no finding and returned custody at `cd5c90400c8e2781af59570cd42394e2b5c30162`. Ordinary #4059-only push/update and #4059 exact-head CI remain pending; no certification or merge is implied.

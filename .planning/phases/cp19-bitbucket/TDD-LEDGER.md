# CP19 TDD ledger

Baseline: 45c235dda353dc9e30838e860acac0aa5ebd2e5c. Retained source SHA256: 0cd529630a96b1c9fd2a75081d83302dcd651c7d665f3c780719ba4eed80b627. No production edits preceded the regression.

Red: `go test -json -count=1 -timeout 20m ./internal/cli -run '^TestBatch1CP19BitbucketSource(CommandJoin|JoinOracle)$'` exited1: source join selected297 primary source keys, resolved5 current targets (3 implemented/2 planned), reported292 missing exact targets. Three independent healthy current targets resolve. Five oracle cases PASS: complete, omitted, same-count wrong path, same-count wrong method, present-but-false. This is declarative source-binding evidence through the real engine consumer, not actual provider behavior.

Green: pending complete source-backed admitted generation and full production-path proofs. No Green claim from the passing controls.

Controls: `go test -json -count=1 -timeout 20m ./cmd/connectorgen -run '^TestVNextGeneratedParameterPaths173$'` PASS; actual renderer/publisher/loader/preflight works for fully authored aliases and mappings. `go test -json -count=1 -timeout 20m ./internal/cli -run '^TestBitbucketPublicCommandsStayCredentialBound$'` PASS for3 current implemented commands with0 provider sends.

Exact raw output/source/test identities retained in the allocated CP19 private baseline-receipt.json. Authoring dependency routed to Firstmate before creating forbidden derived mappings or a private importer. CP16 carry custody remains unchanged.

## Firstmate196 independent evidence checkpoint

Red: `go test -json -count=1 -timeout 20m ./internal/cli -run '^TestBatch1CP19BitbucketPaginationWire$'` exits1. Both anonymous/follow_next and bearer/follow_next return IDs repo-a,repo-b (count2) rather than repo-a through repo-e (count5); one physical send rather than three. CLI exit0 does not establish complete retrieval. Both terminal controls pass. The retained next-link contract requires following next even when page/size are omitted and the page is short. Current page_number declaration ignores next. No production edit or repaired Green exists.

Controls: `go test -json -count=1 -timeout 20m ./cmd/connectorgen -run '^TestBatch1CP19Bitbucket(RetainedFacts|FactOracle)$'` passes all297 exact source-operation subtests and seven oracle cases, with no skips. The enhanced check compares the complete source_operation and uses json.Number to reject changed large integers. This proves retained-source normalization, not canonical execution lowering. Original and enhanced outputs are separately retained.

Controls: `go test -race -json -count=1 -timeout 20m ./internal/cli -run '^TestBatch1CP19BitbucketPaginationWire$/^(anonymous|bearer)$/^terminal$'` exits0, exactly two terminal subtests, no skips. It checks actual IDs, count, physical sends, authentication and request constraints with independent local project state. It does not reach a continuation or prove protected-endpoint authorization enforcement.

Green: still pending shared source-to-schema4 lowering and declared pagination repair. The source-join RED was not rerun unchanged. Raw results, selected members and file hashes are in private independent-evidence-198.json. SOURCE-OBSERVATIONS.md retains source coordinates for73 pageable operations,13 downloads and two uploads; these are source facts and fixture obligations, not binary or warehouse proof.

## Firstmate236 shared coordinate walk

Behavior-preserving correction at merged parent ed32e3c, preserving a0f7ead3 ancestry. Red: controlled Go overlay omitting original terminal validation fails21 subtests; no defective source committed. Passing original baseline is compatibility evidence, not RED. Expanded permanent suite has53 parent/subtest events, independently asserting coordinates/reasons, nil/empty, cache visits,100000 budget, adjacent bounded reference depths and early zero-resolution refusal. Separate final-source overlays with same-count wrong coordinates and wrong reason are rejected.

Green: final focused `TestCoordinate236` suite passes with native Go1.26.6, GOTOOLCHAIN=local, -count=1 -timeout20m. Existing source-lane normal selection995 parent/subtest events has14 failing parents/84 failure events; exact same failed members reproduce with original production helper via overlay. These are inherited saved-write fixture admission failures, not fixed here or hidden as successes. Native vet passes; package lint retains15 unchanged findings, diff lint zero. Raw bytes/commands/output and member comparison are in private shared-coordinate-236 receipts; final race disposition follows VERIFICATION.md. Parent owns coherent-unit review and actual hosted CodeQL outcome.

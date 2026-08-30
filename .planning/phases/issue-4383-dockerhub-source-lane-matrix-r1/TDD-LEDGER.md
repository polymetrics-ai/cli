# TDD ledger — #4383 Docker Hub source-to-seven-lane matrix

| Stage | Evidence | Result |
| --- | --- | --- |
| Red | Add the focused Docker Hub matrix contract test before `dockerhub-source-lane-matrix.json` exists, then run it with `-count=1`. | Observed: the test compiled and failed every contract case because the matrix file did not yet exist; no runtime or source-import failure was involved. |
| Green | Materialize the source-lock-bound matrix and validate every source row, source fact, lane state/reason, count, and backlink. | Observed: focused package test passes with 54 rows, 378 cells, and all four connector artifact backlink sets reconciled. |
| Edge | Mutate decoded matrix in memory for hidden and duplicate rows, invalid/source-less backlinks, omitted source facts, missing ETL/sync, missing direct-write/reverse-ETL, and count mismatch. | Observed: all adversarial subtests pass by rejecting each mutation before any execution path is involved. |
| Refactor | Normalize source citations/reasons, include exact source URL and location on every cell, then gofmt. | Observed: gofmt, JSON syntax checks, package vet, package test, and focused race test pass. |

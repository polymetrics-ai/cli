# Plan

## Task Delivery Header

Issue #4407 on `feat/4407-notion-track-a-matrix-r1`; source snapshot `dc481bac1a8b78d60ac0b4a2b0dfd1a9068ce8db`; Track A only; no merge.

## 1. Preserve frozen source input

Materialize only the Notion operation lock and crosswalk missing from `origin/main` by Git archive from the fixed Batch R1 snapshot. Verify each destination blob hash and byte count against the source Git objects.

## 2. Build explicit seven-lane matrix

Create one source-operation row per retained lock ID, retaining method, path, operation ID, citation, scope parameters, media, pagination, and callback/event facts. Each row has all seven cells with an allowed disposition. Keep source-only restrictions visible; use `mapped_unproven`, never `implemented`, because this track has no runtime proof.

## 3. Add reconciliation test

Validate lock denominator, exact source IDs, every lane cell, source backlink, crosswalk-only boundary records, manual semantic-read/mutation/ETL/binary cohorts, lane-count summary, and no promotion to `implemented`. Include deliberate red cases for hidden rows, a missing cell, a missing backlink, an invalid disposition, and a dropped boundary identity.

## 4. Verification and review

Run source JSON validity, the focused Notion package test and vet, source/projection check modes that do not change files, `agentcontractgen check`, staged whitespace checks, and scoped manual review. Record the two expected mapping-control restrictions separately: v2 embedded `source_operation` payload parsing and the absent canonical descriptor. Neither is a runtime-foundation claim.

## Skills applied

`connector-lane-build-order`, `golang-how-to`, `golang-testing`, `golang-error-handling`, `golang-safety`, `golang-security`, `golang-design-patterns`, and `golang-structs-interfaces`.

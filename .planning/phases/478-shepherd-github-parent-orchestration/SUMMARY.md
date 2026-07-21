# Summary: #478

Status: implementation locally verified; stacked PR and parent-owned review pending.

The plan-first checkpoint fixes the immutable base, owned file boundary, strict RED→GREEN→REFACTOR
sequence, fake-only transport policy, exact-head review policy, human gates, and coordinator-bounded
verification matrix. The test-only RED checkpoint is now captured: all three matching test files
fail because their production modules are intentionally absent.

Minimal GREEN completed at 21 focused passes. A subsequent test-only adversarial RED checkpoint
captured 10 receipt, restart, planned-review binding, parent handoff, hostile-array, duplicate
finding, and disposition failures against unchanged production. The matching correction is pushed
at `40ce66d4b5010b92089895a05709687143d15a05` and passes 27/27 focused tests.

The delivered boundary now provides exact-shape authoritative evidence validation, a declarative
independent Codex-only review route, reconcile-before-mutate issue/PR/roster publication, exact
generation/marker/base/head integration receipts, merged-PR restart recovery, upstream child and
parent handoff capture, and a broker-consumed exact-generation/head parent ready gate. It never
wires sessions or performs a parent merge.

All parent-authorized local gates pass at the implementation head: the serialized Shepherd suite
reports 290 pass, 0 fail, and one intentional sandbox skip; strict owned and all-production
TypeScript pass with TypeScript 5.9.3 against cached Pi 0.80.6; pinned offline Pi RPC returns `true`
for `pm-shepherd`; and exact diff, immutable-base, and owned-path checks pass. No reviewer was
started; stable-head independent review remains parent-owned.

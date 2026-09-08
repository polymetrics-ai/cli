# Shared correction verification through Firstmate198

CR-173-01 implementation and WR-173-01 oracle corrections have focused normal/race evidence. Neither correction is independently accepted. The affected physical database witness remains incomplete; no complete implementation or hosted-green claim.

## Original receipts

Original command, source/input snapshot, output and receipt bytes remain under the private task receipts-cp13-141 directory. The compact index shared-corrections-handoff-198-receipts.json binds original receipt hashes and selected commands. Counts include Go parent/child events, overlap and are not added. Vet/lint/generator/check invocations do not imply test events. The focused race command's commandrunner selection had zero tests; its full normal package ran separately.

| Run | Exit | Pass/run events | Raw SHA256 |
|---|---:|---:|---|
| shared-command-token-red-194-01 | 1 | 53/78 | 116c53d838638c48d9ee3bd63af862b744087802857f97fe281f7602503a0413 |
| shared-command-token-green-194-01 | 0 | 82/82 | 1203306c764e7cb9cdf9d70ff8f7b594fcc8ef59349a819e77009e268003751b |
| shared-checkpoint-red-194-01 | 1 | 17/29 | 6f5e843a20ec4cf9515baeaa245df98851bc1b3139de463a18f13604856fd1c8 |
| shared-checkpoint-green-194-01 | 0 | 29/29 | 62b15093a2b2fd00102c2178dbcbf99f48d811d2820054c935aa4a44d0762de9 |
| shared-corrections-race-194-01 | 0 | 111/111 | 7f928f0e2a81d9faeb383924d5a1f9d8792218cdc094d61e7e97f375334c85b8 |
| shared-commandrunner-normal-194-01 | 0 | 309/309 | cc18b16064ca8ef670b12f74693cac1ce1943ef8d748a85ae20b44647f0ebcdb |
| shared-flag-grammar-continuity-194-01 | 0 | 12/12 | 0bdd4c64097491696b7e010325e3655cc26115cf9f32caf6ee85997c0970e818 |
| shared-vet-194-01 | 0 | 0/0 | e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855 |
| shared-lint-194-01 | 0 | 0/0 | e92606b0bf483111dff0a120c315ea165821348f31365020e2468a0059095c47 |
| shared-source-lanes-check-194-01 | 1 | 0/0 | ee971d3c3043dd3d34f33e1c2e4897fbe8b79312a4fc5132922a64ce2596a922 |
| shared-source-lanes-generate-194-01 | 0 | 0/0 | 0a6e9b9ac60222289b509a65cf48808c03cdbd167ee18ae9bd960de315a419df |
| shared-source-demands-generate-194-01 | 1 | 0/0 | 4ee34600e66acf4ffe120bf9d2fe46aab9f0fd2bed1aeda2448b15ce6b004898 |
| shared-source-demands-generate-194-02 | 0 | 0/0 | 1af4f5ef512c1babc494c1f1546b6fecc0a61b9c0c8eb379f53e656e5ffe8ea9 |
| shared-generated-index-194-01 | 0 | 0/0 | 0ba4d7d836e45b54a827c1daa82f4c8af8c3c0dbd413530cfa404994890ec694 |
| shared-checkpoint-physical-194-01 | 1 | 0/1 | dad4e4cba5923e74b8c474fcae152eaa75beb8484a2450eed8e75f65e4534ff0 |
| shared-source-demands-check-194-01 | 0 | 0/0 | e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855 |
| shared-source-lanes-check-197-01 | 0 | 0/0 | e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855 |
| shared-source-visibility-197-01 | 0 | 36/36 | 04fd2f601a4e0e175e7636e8802484482645870c8a741dec7a06cc1720e8b27f |

## Infrastructure and generated propagation

The source-lanes initial check rejected stale inputs. Regeneration changed only manifest input pins, retaining all source operations/lane semantics/totals. Atlas revision53 changed existing authoring/transport constraints and proof-test references. Exactly27 authored assessment Atlas observations were rebound without changing assessment dispositions; the first demand generation rejected old observations before the successful second generation. Demand register baseline, coverage, known obligations and requirement/source-fit dispositions remain unchanged. The ten embedded cohort payloads differ only by decoded atlas_sha256; other543 entries remain identical. No connector execution/source-lock changes or capability promotion occurred. Exact mechanical deltas are in shared-source-manifest-delta-194.json, shared-assessment-rebinding-194.json, shared-demand-register-delta-194.json and shared-generated-delta-194.json.

Two parallel receipt snapshots failed with FileExistsError BEFORE their commands began (shared-source-lanes-check-194-02 and shared-source-visibility-194-01). Original directories/diagnostics remain immutable. Firstmate197 explicitly resolved that infrastructure obstacle; serial fresh-name197 runs passed. The original runner was not modified.

The physical194 run waited in Docker pull/credential helper before a database container existed, then terminated through owned cleanup: exit1, one selected event/zero pass. It is setup failure, not product RED. Firstmate197 diagnosis proves the exact postgres:16.10 image is currently cached on the same daemon reached by /var/run/docker.sock. Harness Start unconditionally pulls even when cached, without supported cache-only selection. Proposed test-only policy is pending under shared-db-cache-197; no helper, credential, daemon or security settings changed. Prior successful physical173 proof remains historical and cannot verify the changed oracle.

## Remaining gates and custody

Fresh affected physical proof, final shared correction handoff and Firstmate-bound independent correction review remain. Hosted CodeQL annotation is still under196B assessment, not dismissed or accepted as a bug. CR-173-02 remains OPEN and has been accepted into cli-batch1-cp23-notion custody under196. No Notion files changed. Two read-only196 architecture children run under exact Astra/xhigh native bindings in196-binding.json; no new representation/runtime implementation is authorized before Firstmate reconciliation. Shared source is frozen during their reads.

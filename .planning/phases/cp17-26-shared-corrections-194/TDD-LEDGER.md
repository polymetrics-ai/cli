# Shared correction TDD194

CR-173-01 Red: original reviewer173-P01 reached canonical publication and actual parser failure; new sibling RED pending before guard edit. Green: pending.

WR-173-01 Red: original reviewer173-P02 native backward LSN accepted; expanded oracle RED pending. Green: pending; affected physical witness pending.

CR-173-01 expanded Red: shared-command-token-red-194-01 exit1,78run/53pass; raw116c53d838638c48d9ee3bd63af862b744087802857f97fe281f7602503a0413. All16 availability/position admission failures plus actual publication and6parser siblings reached intended assertions. No input drift. Green follows same test bytes after shared guard prefix refusal only.

CR-173-01 Green: shared-command-token-green-194-01 exit0,82/82 including prior exact generated parameter mappings; raw1203306c764e7cb9cdf9d70ff8f7b594fcc8ef59349a819e77009e268003751b, no input drift.

WR-173-01 Red: shared-checkpoint-red-194-01 exit1,29run/17pass; raw6f5e843a20ec4cf9515baeaa245df98851bc1b3139de463a18f13604856fd1c8. Actual backward/equal/malformed LSN and changed version/protocol/generation/barrier assertions fail; healthy old8 plus forward/identity/time refusals remain passing. No input drift. Green pending unchanged expanded test bytes.

WR-173-01 Green: shared-checkpoint-green-194-01 exit0,29/29, raw62b15093a2b2fd00102c2178dbcbf99f48d811d2820054c935aa4a44d0762de9; unchanged test source, no input drift. Combined focused race shared-corrections-race-194-01 passed111/111, raw7f928f0e2a81d9faeb383924d5a1f9d8792218cdc094d61e7e97f375334c85b8. Full commandrunner309/309 and flag-grammar12/12 passed.

Physical witness shared-checkpoint-physical-194-01: terminal exit1 after owned interruption,483.36s,1run/0pass, rawdad4e4cba5923e74b8c474fcae152eaa75beb8484a2450eed8e75f65e4534ff0, no input drift. Docker pull waited on docker-credential-desktop before any dbtest database container. Original test SIGINT and remaining owned helper SIGTERM cleaned the attempt. No new physical proof; Firstmate must resolve runtime prerequisite without unauthorized credential/security changes. No product/checkpoint failure inferred.

Firstmate197/198 recovery: demand check194, serial source-lanes check197 and source-visibility197 all pass without source drift. The two pre-command snapshot collisions are preserved infrastructure failures, not behavioral Red. Physical oracle proof remains pending; see VERIFICATION.md and shared-db-diagnosis-197.json.

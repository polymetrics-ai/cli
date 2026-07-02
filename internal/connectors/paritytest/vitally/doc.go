// Package vitallyparity_test is the engine-vs-legacy parity suite for the
// vitally pilot migration (PLAN.md P-2, SPEC.md §5.2). It lives in its own
// package (not internal/connectors/engine) per SPEC.md §6's parity-test
// location decision: per-connector directories give clean 10-way DW-1
// parallelism with no shared Go package namespace collisions across pilot
// agents.
package vitallyparity_test

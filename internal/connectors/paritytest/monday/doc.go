// Package mondayparity_test is the engine-vs-legacy parity suite for the
// monday bundle (wave1-pilot P-8), living in its own directory per
// SPEC §6's parity-test-location decision (disjoint per-agent path guard,
// no shared engine_test package namespace collision across the 10 parallel
// pilot agents). This file exists purely so `go build ./...` sees a real Go
// package even before parity_test.go is added (mirrors xkcd/paritytest's doc.go
// convention referenced by SPEC §6).
package mondayparity_test

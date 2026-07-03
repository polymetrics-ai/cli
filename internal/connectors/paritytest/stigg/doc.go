// Package stiggparity_test is the engine-vs-legacy parity suite for the
// stigg bundle, living in its own directory per the disjoint per-agent
// path-guard convention (mirrors paritytest/monday's doc.go). This file
// exists purely so `go build ./...` sees a real Go package even before
// parity_test.go is added.
package stiggparity_test

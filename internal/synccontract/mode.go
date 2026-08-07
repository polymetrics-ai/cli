// Package synccontract defines the product-wide contract for durable database
// synchronization. It deliberately has no connector, transport, or app
// dependency so source and destination lanes consume the same semantics.
package synccontract

import "fmt"

// Mode is the closed vocabulary for a database synchronization operation.
// Parsing or declaring a mode does not make it executable; native executor and
// conformance admission are checked separately by NativeCommandContract.
type Mode string

const (
	ModeFullOverwrite            Mode = "full_overwrite"
	ModeFullAppend               Mode = "full_append"
	ModeIncrementalAppend        Mode = "incremental_append"
	ModeIncrementalUpsert        Mode = "incremental_upsert"
	ModeIncrementalDedupe        Mode = "incremental_dedupe"
	ModeIncrementalDedupeHistory Mode = "incremental_dedupe_history"
	ModeChangeCapture            Mode = "change_capture"
)

var allModes = []Mode{
	ModeFullOverwrite,
	ModeFullAppend,
	ModeIncrementalAppend,
	ModeIncrementalUpsert,
	ModeIncrementalDedupe,
	ModeIncrementalDedupeHistory,
	ModeChangeCapture,
}

// AllModes returns the closed mode vocabulary in stable display order.
func AllModes() []Mode {
	return append([]Mode(nil), allModes...)
}

// Validate rejects any mode outside the product-wide vocabulary.
func (m Mode) Validate() error {
	for _, candidate := range allModes {
		if m == candidate {
			return nil
		}
	}
	return fmt.Errorf("unsupported sync mode %q", m)
}

// ModeNotExecutableError distinguishes a recognized mode from one whose
// native executor and conformance evidence have not been admitted yet.
type ModeNotExecutableError struct {
	Mode   Mode
	Reason string
}

func (e *ModeNotExecutableError) Error() string {
	if e == nil {
		return "sync mode is not executable"
	}
	if e.Reason == "" {
		return fmt.Sprintf("sync mode %q is not executable: native executor and conformance evidence are required", e.Mode)
	}
	return fmt.Sprintf("sync mode %q is not executable: %s", e.Mode, e.Reason)
}

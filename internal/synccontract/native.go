package synccontract

import (
	"fmt"
	"strings"
)

// NativeCommandContractVersion is the schema version for named database-wire
// commands. It is intentionally independent from checkpoint StateVersion.
const NativeCommandContractVersion uint = 1

// ExecutorReference names fixed product code. It is never populated from a
// caller, so a native contract cannot become a generic SQL, HTTP, or shell
// escape hatch.
type ExecutorReference struct {
	Kind string `json:"kind"`
	ID   string `json:"id"`
}

func (r ExecutorReference) validate() error {
	if r.Kind != "native" || !isNamedIdentifier(r.ID) || isGenericIdentifier(r.ID) {
		return fmt.Errorf("native executor requires kind %q and a concrete executor ID", "native")
	}
	return nil
}

// NativeSyncExecutorDescriptor is the runtime half of a named native command.
// It intentionally has no request string, REST endpoint, or command payload.
type NativeSyncExecutorDescriptor struct {
	Protocol string            `json:"protocol"`
	Command  string            `json:"command"`
	Executor ExecutorReference `json:"executor"`
	Modes    []Mode            `json:"modes"`
}

// NativeSyncExecutor is the explicit admission interface. Implementing a
// generic connector/read method cannot accidentally advertise sync support.
type NativeSyncExecutor interface {
	NativeSyncExecutorDescriptor() NativeSyncExecutorDescriptor
	NativeSyncConformanceEvidence() ConformanceEvidence
}

// NativeCommandContract admits a fixed native database-wire operation. It is
// the alternative to fabricated REST api_surface rows for SQL and other wire
// protocols.
type NativeCommandContract struct {
	ContractVersion uint                `json:"contract_version"`
	Protocol        string              `json:"protocol"`
	Command         string              `json:"command"`
	Executor        ExecutorReference   `json:"executor"`
	Modes           []Mode              `json:"modes"`
	Conformance     ConformanceEvidence `json:"conformance"`
}

// Validate verifies that a native command contains fixed execution identity
// and the complete shared fixture evidence before any mode can be advertised.
func (c NativeCommandContract) Validate() error {
	if c.ContractVersion != NativeCommandContractVersion {
		return fmt.Errorf("unsupported native command contract version %d", c.ContractVersion)
	}
	if !isConcreteNativeIdentifier(c.Protocol) || !isConcreteNativeIdentifier(c.Command) {
		return fmt.Errorf("native command protocol and command must name a concrete database wire operation")
	}
	if err := c.Executor.validate(); err != nil {
		return err
	}
	if len(c.Modes) == 0 {
		return fmt.Errorf("native command requires at least one sync mode")
	}
	seen := make(map[Mode]struct{}, len(c.Modes))
	for _, mode := range c.Modes {
		if err := mode.Validate(); err != nil {
			return err
		}
		if _, exists := seen[mode]; exists {
			return fmt.Errorf("native command declares duplicate sync mode %q", mode)
		}
		seen[mode] = struct{}{}
	}
	if !c.Conformance.matchesRequired() {
		return fmt.Errorf("native command requires complete shared conformance evidence")
	}
	return nil
}

// IsExecutable admits a mode claim only if this valid declaration has a
// matching runtime executor and both sides attest to the same fixture corpus.
func (c NativeCommandContract) IsExecutable(candidate any) bool {
	if c.Validate() != nil {
		return false
	}
	executor, ok := candidate.(NativeSyncExecutor)
	if !ok {
		return false
	}
	descriptor := executor.NativeSyncExecutorDescriptor()
	if descriptor.Protocol != c.Protocol || descriptor.Command != c.Command || descriptor.Executor != c.Executor || !sameModeSet(descriptor.Modes, c.Modes) {
		return false
	}
	evidence := executor.NativeSyncConformanceEvidence()
	return evidence.matchesRequired() && evidence.equal(c.Conformance)
}

func sameModeSet(left, right []Mode) bool {
	if len(left) != len(right) {
		return false
	}
	seen := make(map[Mode]struct{}, len(left))
	for _, mode := range left {
		seen[mode] = struct{}{}
	}
	if len(seen) != len(left) {
		return false
	}
	for _, mode := range right {
		if _, exists := seen[mode]; !exists {
			return false
		}
	}
	return true
}

func isNamedIdentifier(value string) bool {
	if value == "" || strings.TrimSpace(value) != value {
		return false
	}
	for i, r := range value {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_' || r == '-' || r == '.' {
			if i == 0 && (r == '-' || r == '.') {
				return false
			}
			continue
		}
		return false
	}
	return true
}

func isConcreteNativeIdentifier(value string) bool {
	return isNamedIdentifier(value) && !isGenericIdentifier(value)
}

func isGenericIdentifier(value string) bool {
	switch strings.ToLower(value) {
	case "http", "https", "rest", "sql", "query", "execute", "shell", "command":
		return true
	default:
		return false
	}
}

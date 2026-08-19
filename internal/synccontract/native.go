package synccontract

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"sync"
)

// NativeCommandContractVersion is the supported serialization version for a
// native command admission contract.
const NativeCommandContractVersion uint = 1

// ExecutorReference names a concrete native executor without carrying a raw
// transport, query, or shell payload.
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

// NativeSyncExecutorDescriptor declares the fixed operation an executor can
// perform and the closed modes it admits.
type NativeSyncExecutorDescriptor struct {
	Protocol string            `json:"protocol"`
	Command  string            `json:"command"`
	Executor ExecutorReference `json:"executor"`
	Modes    []Mode            `json:"modes"`
}

func (d NativeSyncExecutorDescriptor) validate() error {
	return validateNativeOperation(d.Protocol, d.Command, d.Executor, d.Modes)
}

// NativeSyncRequest supplies the requested mode and the source identity that
// must be checked before an existing checkpoint is resumed.
type NativeSyncRequest struct {
	Mode       Mode                `json:"mode"`
	Resume     ResumeExpectation   `json:"resume"`
	Checkpoint *CheckpointEnvelope `json:"checkpoint,omitempty"`
}

func (r NativeSyncRequest) clone() NativeSyncRequest {
	clone := r
	clone.Resume.SourceGeneration = cloneToken(r.Resume.SourceGeneration)
	if r.Checkpoint != nil {
		checkpoint := r.Checkpoint.Clone()
		clone.Checkpoint = &checkpoint
	}
	return clone
}

// NativeSyncResult returns an observed checkpoint candidate; callers must
// still commit it through downstream durability acknowledgement.
type NativeSyncResult struct {
	CandidateCheckpoint *CheckpointEnvelope `json:"candidate_checkpoint,omitempty"`
}

// NativeExecutorAdmission is the descriptor/evidence portion shared by every
// native source or database-target executor. It intentionally does not imply a
// source RunNativeSync implementation: target database work has a different
// typed lifecycle and must not be forced through NativeSyncRequest.
type NativeExecutorAdmission interface {
	NativeSyncExecutorDescriptor() NativeSyncExecutorDescriptor
	NativeSyncConformanceEvidence() ConformanceEvidence
}

// NativeSyncExecutor is a source-side native implementation. The registry can
// store the smaller admission interface, but Execute requires this runner at
// dispatch time.
type NativeSyncExecutor interface {
	NativeExecutorAdmission
	RunNativeSync(context.Context, NativeSyncRequest) (NativeSyncResult, error)
}

// ErrNativeSyncExecutorRequired identifies an admitted native descriptor that
// is not a runnable source executor. Database target executors may still use
// the shared admission evidence through their own typed consumer boundary.
var ErrNativeSyncExecutorRequired = errors.New("native source execution requires a runnable native sync executor")

// NativeCommandContract is the declarative admission record for a fixed
// native database operation. It deliberately excludes generic query fields.
type NativeCommandContract struct {
	ContractVersion uint                `json:"contract_version"`
	Protocol        string              `json:"protocol"`
	Command         string              `json:"command"`
	Executor        ExecutorReference   `json:"executor"`
	Modes           []Mode              `json:"modes"`
	Conformance     ConformanceEvidence `json:"conformance"`
}

// Validate confirms that the contract names a concrete operation and carries
// the complete shared conformance evidence required for execution.
func (c NativeCommandContract) Validate() error {
	if c.ContractVersion != NativeCommandContractVersion {
		return fmt.Errorf("unsupported native command contract version %d", c.ContractVersion)
	}
	if err := validateNativeOperation(c.Protocol, c.Command, c.Executor, c.Modes); err != nil {
		return err
	}
	if !c.Conformance.matchesRequired() {
		return fmt.Errorf("native command requires complete shared conformance evidence")
	}
	return nil
}

func validateNativeOperation(protocol, command string, executor ExecutorReference, modes []Mode) error {
	if !isConcreteNativeIdentifier(protocol) || !isConcreteNativeIdentifier(command) {
		return fmt.Errorf("native command protocol and command must name a concrete database wire operation")
	}
	if err := executor.validate(); err != nil {
		return err
	}
	if len(modes) == 0 {
		return fmt.Errorf("native command requires at least one sync mode")
	}
	seen := make(map[Mode]struct{}, len(modes))
	for _, mode := range modes {
		if err := mode.Validate(); err != nil {
			return err
		}
		if _, exists := seen[mode]; exists {
			return fmt.Errorf("native command declares duplicate sync mode %q", mode)
		}
		seen[mode] = struct{}{}
	}
	return nil
}

// NativeExecutorRegistry matches admitted contracts to registered native
// executors while protecting registry access from concurrent callers.
type NativeExecutorRegistry struct {
	mu        sync.RWMutex
	executors map[ExecutorReference]NativeExecutorAdmission
}

// NewNativeExecutorRegistry creates a registry and validates every supplied
// executor before making it available.
func NewNativeExecutorRegistry(executors ...NativeExecutorAdmission) (*NativeExecutorRegistry, error) {
	registry := &NativeExecutorRegistry{executors: make(map[ExecutorReference]NativeExecutorAdmission)}
	for _, executor := range executors {
		if err := registry.Register(executor); err != nil {
			return nil, err
		}
	}
	return registry, nil
}

// Register admits one uniquely named executor only when its descriptor and
// conformance evidence are complete.
func (r *NativeExecutorRegistry) Register(executor NativeExecutorAdmission) error {
	if r == nil {
		return fmt.Errorf("native executor registry is required")
	}
	if isNilNativeExecutorAdmission(executor) {
		return fmt.Errorf("native executor admission is required")
	}
	descriptor := executor.NativeSyncExecutorDescriptor()
	if err := descriptor.validate(); err != nil {
		return err
	}
	if !executor.NativeSyncConformanceEvidence().matchesRequired() {
		return fmt.Errorf("native executor requires complete shared conformance evidence")
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.executors == nil {
		r.executors = make(map[ExecutorReference]NativeExecutorAdmission)
	}
	if _, exists := r.executors[descriptor.Executor]; exists {
		return fmt.Errorf("native executor %q is already registered", descriptor.Executor.ID)
	}
	r.executors[descriptor.Executor] = executor
	return nil
}

// Admits reports whether a registered descriptor/evidence admission exactly
// matches contract. It does not imply that the object is a runnable source
// executor; Execute makes that stronger check at dispatch time.
func (r *NativeExecutorRegistry) Admits(contract NativeCommandContract) bool {
	_, err := r.admissionFor(contract)
	return err == nil
}

// Execute validates admission and resume state before dispatching a defensive
// copy of request to its matching native executor.
func (r *NativeExecutorRegistry) Execute(ctx context.Context, contract NativeCommandContract, request NativeSyncRequest) (NativeSyncResult, error) {
	if ctx == nil {
		return NativeSyncResult{}, fmt.Errorf("native sync context is required")
	}
	if err := ctx.Err(); err != nil {
		return NativeSyncResult{}, err
	}
	admission, err := r.admissionFor(contract)
	if err != nil {
		return NativeSyncResult{}, err
	}
	executor, ok := admission.(NativeSyncExecutor)
	if !ok || isNilNativeSyncExecutor(executor) {
		return NativeSyncResult{}, ErrNativeSyncExecutorRequired
	}
	if err := request.Mode.Validate(); err != nil {
		return NativeSyncResult{}, err
	}
	if !containsMode(contract.Modes, request.Mode) {
		return NativeSyncResult{}, fmt.Errorf("native command does not admit sync mode %q", request.Mode)
	}
	if request.Checkpoint != nil {
		if err := request.Checkpoint.ValidateResume(request.Resume); err != nil {
			return NativeSyncResult{}, err
		}
	}
	return executor.RunNativeSync(ctx, request.clone())
}

func (r *NativeExecutorRegistry) admissionFor(contract NativeCommandContract) (NativeExecutorAdmission, error) {
	if r == nil {
		return nil, fmt.Errorf("native executor registry is required")
	}
	if err := contract.Validate(); err != nil {
		return nil, err
	}
	r.mu.RLock()
	executor, ok := r.executors[contract.Executor]
	r.mu.RUnlock()
	if !ok || isNilNativeExecutorAdmission(executor) {
		return nil, fmt.Errorf("native executor %q is not registered", contract.Executor.ID)
	}
	if err := ValidateNativeAdmission(executor, contract); err != nil {
		return nil, err
	}
	return executor, nil
}

// ValidateNativeAdmission confirms that a registered native object exactly
// matches a command's descriptor and required conformance evidence. It is
// exported so non-source consumers (such as the database driver registry) can
// use the one #3810 admission rule without reimplementing it.
func ValidateNativeAdmission(admission NativeExecutorAdmission, contract NativeCommandContract) error {
	if isNilNativeExecutorAdmission(admission) {
		return fmt.Errorf("native executor admission is required")
	}
	if err := contract.Validate(); err != nil {
		return err
	}
	descriptor := admission.NativeSyncExecutorDescriptor()
	if descriptor.Protocol != contract.Protocol || descriptor.Command != contract.Command || descriptor.Executor != contract.Executor || !sameModeSet(descriptor.Modes, contract.Modes) {
		return fmt.Errorf("registered native executor does not match the command contract")
	}
	evidence := admission.NativeSyncConformanceEvidence()
	if !evidence.matchesRequired() || !evidence.equal(contract.Conformance) {
		return fmt.Errorf("registered native executor lacks matching shared conformance evidence")
	}
	return nil
}

func isNilNativeSyncExecutor(executor NativeSyncExecutor) bool {
	return isNilNativeExecutorAdmission(executor)
}

func isNilNativeExecutorAdmission(executor NativeExecutorAdmission) bool {
	if executor == nil {
		return true
	}
	value := reflect.ValueOf(executor)
	switch value.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
		return value.IsNil()
	default:
		return false
	}
}

func containsMode(modes []Mode, expected Mode) bool {
	for _, mode := range modes {
		if mode == expected {
			return true
		}
	}
	return false
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
	for _, segment := range strings.FieldsFunc(strings.ToLower(value), func(r rune) bool {
		return r == '-' || r == '_' || r == '.'
	}) {
		for _, generic := range []string{"http", "https", "rest", "sql", "query", "execute", "shell", "command"} {
			if segment == generic || strings.HasPrefix(segment, generic) {
				return true
			}
		}
	}
	return false
}

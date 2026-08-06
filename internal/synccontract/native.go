package synccontract

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"sync"
)

const NativeCommandContractVersion uint = 1

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

type NativeSyncExecutorDescriptor struct {
	Protocol string            `json:"protocol"`
	Command  string            `json:"command"`
	Executor ExecutorReference `json:"executor"`
	Modes    []Mode            `json:"modes"`
}

func (d NativeSyncExecutorDescriptor) validate() error {
	return validateNativeOperation(d.Protocol, d.Command, d.Executor, d.Modes)
}

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

type NativeSyncResult struct {
	CandidateCheckpoint *CheckpointEnvelope `json:"candidate_checkpoint,omitempty"`
}

type NativeSyncExecutor interface {
	NativeSyncExecutorDescriptor() NativeSyncExecutorDescriptor
	NativeSyncConformanceEvidence() ConformanceEvidence
	RunNativeSync(context.Context, NativeSyncRequest) (NativeSyncResult, error)
}

type NativeCommandContract struct {
	ContractVersion uint                `json:"contract_version"`
	Protocol        string              `json:"protocol"`
	Command         string              `json:"command"`
	Executor        ExecutorReference   `json:"executor"`
	Modes           []Mode              `json:"modes"`
	Conformance     ConformanceEvidence `json:"conformance"`
}

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

type NativeExecutorRegistry struct {
	mu        sync.RWMutex
	executors map[ExecutorReference]NativeSyncExecutor
}

func NewNativeExecutorRegistry(executors ...NativeSyncExecutor) (*NativeExecutorRegistry, error) {
	registry := &NativeExecutorRegistry{executors: make(map[ExecutorReference]NativeSyncExecutor)}
	for _, executor := range executors {
		if err := registry.Register(executor); err != nil {
			return nil, err
		}
	}
	return registry, nil
}

func (r *NativeExecutorRegistry) Register(executor NativeSyncExecutor) error {
	if r == nil {
		return fmt.Errorf("native executor registry is required")
	}
	if isNilNativeSyncExecutor(executor) {
		return fmt.Errorf("native sync executor is required")
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
		r.executors = make(map[ExecutorReference]NativeSyncExecutor)
	}
	if _, exists := r.executors[descriptor.Executor]; exists {
		return fmt.Errorf("native executor %q is already registered", descriptor.Executor.ID)
	}
	r.executors[descriptor.Executor] = executor
	return nil
}

func (r *NativeExecutorRegistry) Admits(contract NativeCommandContract) bool {
	_, err := r.executorFor(contract)
	return err == nil
}

func (r *NativeExecutorRegistry) Execute(ctx context.Context, contract NativeCommandContract, request NativeSyncRequest) (NativeSyncResult, error) {
	executor, err := r.executorFor(contract)
	if err != nil {
		return NativeSyncResult{}, err
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

func (r *NativeExecutorRegistry) executorFor(contract NativeCommandContract) (NativeSyncExecutor, error) {
	if r == nil {
		return nil, fmt.Errorf("native executor registry is required")
	}
	if err := contract.Validate(); err != nil {
		return nil, err
	}
	r.mu.RLock()
	executor, ok := r.executors[contract.Executor]
	r.mu.RUnlock()
	if !ok || isNilNativeSyncExecutor(executor) {
		return nil, fmt.Errorf("native executor %q is not registered", contract.Executor.ID)
	}
	descriptor := executor.NativeSyncExecutorDescriptor()
	if descriptor.Protocol != contract.Protocol || descriptor.Command != contract.Command || descriptor.Executor != contract.Executor || !sameModeSet(descriptor.Modes, contract.Modes) {
		return nil, fmt.Errorf("registered native executor does not match the command contract")
	}
	evidence := executor.NativeSyncConformanceEvidence()
	if !evidence.matchesRequired() || !evidence.equal(contract.Conformance) {
		return nil, fmt.Errorf("registered native executor lacks matching shared conformance evidence")
	}
	return executor, nil
}

func isNilNativeSyncExecutor(executor NativeSyncExecutor) bool {
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

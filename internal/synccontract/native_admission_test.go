package synccontract

import (
	"context"
	"errors"
	"testing"
)

func TestNativeAdmissionIsNotItselfASourceRunner(t *testing.T) {
	contract := NativeCommandContract{
		ContractVersion: NativeCommandContractVersion,
		Protocol:        "postgres-wire",
		Command:         "managed-table-apply",
		Executor:        ExecutorReference{Kind: "native", ID: "postgres-managed-table-v1"},
		Modes:           []Mode{ModeFullAppend},
		Conformance:     RequiredConformanceEvidence(),
	}
	admission := admissionOnly{descriptor: NativeSyncExecutorDescriptor{
		Protocol: contract.Protocol,
		Command:  contract.Command,
		Executor: contract.Executor,
		Modes:    contract.Modes,
	}, evidence: RequiredConformanceEvidence()}

	registry, err := NewNativeExecutorRegistry(admission)
	if err != nil {
		t.Fatal(err)
	}
	if !registry.Admits(contract) {
		t.Fatal("matching declaration/evidence was not admitted")
	}
	if _, err := registry.Execute(context.Background(), contract, NativeSyncRequest{Mode: ModeFullAppend}); !errors.Is(err, ErrNativeSyncExecutorRequired) {
		t.Fatalf("Execute(admission only) error = %v, want ErrNativeSyncExecutorRequired", err)
	}
}

type admissionOnly struct {
	descriptor NativeSyncExecutorDescriptor
	evidence   ConformanceEvidence
}

func (a admissionOnly) NativeSyncExecutorDescriptor() NativeSyncExecutorDescriptor {
	return a.descriptor
}

func (a admissionOnly) NativeSyncConformanceEvidence() ConformanceEvidence { return a.evidence }

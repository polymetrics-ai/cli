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
	}
	admission := admissionOnly{descriptor: NativeSyncExecutorDescriptor{
		Protocol: contract.Protocol,
		Command:  contract.Command,
		Executor: contract.Executor,
		Modes:    contract.Modes,
	}}

	registry, err := NewNativeExecutorRegistry(admission)
	if err != nil {
		t.Fatal(err)
	}
	if !registry.Admits(contract) {
		t.Fatal("matching declaration was not admitted")
	}
	if _, err := registry.Execute(context.Background(), contract, NativeSyncRequest{Mode: ModeFullAppend}); !errors.Is(err, ErrNativeSyncExecutorRequired) {
		t.Fatalf("Execute(admission only) error = %v, want ErrNativeSyncExecutorRequired", err)
	}
}

type admissionOnly struct {
	descriptor NativeSyncExecutorDescriptor
}

func (a admissionOnly) NativeSyncExecutorDescriptor() NativeSyncExecutorDescriptor {
	return a.descriptor
}

package engine

import (
	"os"
	"strings"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/synccontract"
)

func TestEnabledContractActivationRejectsInventedTransportFacts(t *testing.T) {
	bundle, err := Load(os.DirFS("../defs"), "gitlab")
	if err != nil {
		t.Fatalf("load GitLab enabled contract: %v", err)
	}
	if bundle.EnabledContract == nil || bundle.SyncTransport == nil || bundle.SyncTransport.Source == nil {
		t.Fatal("GitLab enabled contract and source transport are required")
	}

	for _, test := range []struct {
		name   string
		mutate func(*connectors.EnabledConnectorContract, *connectors.SyncTransportDescriptor)
		want   string
	}{
		{
			name: "keyed idempotency",
			mutate: func(contract *connectors.EnabledConnectorContract, transport *connectors.SyncTransportDescriptor) {
				contractLane(contract, "sync_transport").Transport.RuntimeDelivery.Idempotency = connectors.DeliveryIdempotencyKeyed
				transport.Source.Delivery.Idempotency = connectors.DeliveryIdempotencyKeyed
			},
			want: "at_least_once/unordered/not_available",
		},
		{
			name: "source ordering",
			mutate: func(contract *connectors.EnabledConnectorContract, transport *connectors.SyncTransportDescriptor) {
				contractLane(contract, "sync_transport").Transport.RuntimeDelivery.Ordering = connectors.DeliveryOrderingSource
				transport.Source.Delivery.Ordering = connectors.DeliveryOrderingSource
			},
			want: "at_least_once/unordered/not_available",
		},
		{
			name: "tombstones",
			mutate: func(contract *connectors.EnabledConnectorContract, transport *connectors.SyncTransportDescriptor) {
				contractLane(contract, "sync_transport").Transport.RuntimeDelivery.Deletes = connectors.DeliveryDeletesTombstone
				transport.Source.Delivery.Deletes = connectors.DeliveryDeletesTombstone
			},
			want: "at_least_once/unordered/not_available",
		},
		{
			name: "incremental cursor mode",
			mutate: func(contract *connectors.EnabledConnectorContract, transport *connectors.SyncTransportDescriptor) {
				contractLane(contract, "sync_transport").Transport.Modes = append(contractLane(contract, "sync_transport").Transport.Modes, string(synccontract.ModeIncrementalDedupe))
				transport.Source.Modes = append(transport.Source.Modes, synccontract.ModeIncrementalDedupe)
			},
			want: "requires source-cited cursor evidence",
		},
		{
			name: "partial source coverage claimed complete",
			mutate: func(contract *connectors.EnabledConnectorContract, transport *connectors.SyncTransportDescriptor) {
				contractLane(contract, "reverse_etl").Source.Coverage = connectors.EnabledCoverageComplete
			},
			want: "source coverage",
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			contract := bundle.EnabledContract.Clone()
			transport := bundle.SyncTransport.Clone()
			test.mutate(contract, transport)
			err := validateEnabledConnectorContractActivation(contract, transport, bundle.Streams, bundle.Writes)
			if err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("activation error = %v, want %q", err, test.want)
			}
		})
	}
}

func TestGitLabGuideRendersPartialSourceCoverage(t *testing.T) {
	bundle, err := Load(os.DirFS("../defs"), "gitlab")
	if err != nil {
		t.Fatalf("load GitLab bundle: %v", err)
	}
	manual := connectors.RenderConnectorManual(New(bundle, nil))
	for _, want := range []string{
		"ENABLED CONNECTOR CONTRACT",
		"direct_read: implemented (source coverage: partial 582/749; deferred=167; unsupported=0)",
		"reverse_etl: implemented (source coverage: partial 147/1003; deferred=856; unsupported=0)",
		"etl: implemented (source coverage: complete 4/4; deferred=0; unsupported=0)",
	} {
		if !strings.Contains(manual, want) {
			t.Fatalf("GitLab manual does not disclose source coverage %q:\n%s", want, manual)
		}
	}
}

func contractLane(contract *connectors.EnabledConnectorContract, name string) *connectors.EnabledConnectorLane {
	for index := range contract.Lanes {
		if contract.Lanes[index].Name == name {
			return &contract.Lanes[index]
		}
	}
	panic("missing lane " + name)
}

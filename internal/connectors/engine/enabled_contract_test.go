package engine

import (
	"os"
	"strings"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/synccontract"
)

func TestEnabledContractActivationRejectsMismatchedTransportFacts(t *testing.T) {
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
			name: "idempotency",
			mutate: func(contract *connectors.EnabledConnectorContract, transport *connectors.SyncTransportDescriptor) {
				contractLane(contract, "sync_transport").Transport.RuntimeDelivery.Idempotency = connectors.DeliveryIdempotencyKeyed
			},
			want: "delivery policy does not match",
		},
		{
			name: "ordering",
			mutate: func(contract *connectors.EnabledConnectorContract, transport *connectors.SyncTransportDescriptor) {
				contractLane(contract, "sync_transport").Transport.RuntimeDelivery.Ordering = connectors.DeliveryOrderingSource
			},
			want: "delivery policy does not match",
		},
		{
			name: "delete handling",
			mutate: func(contract *connectors.EnabledConnectorContract, transport *connectors.SyncTransportDescriptor) {
				contractLane(contract, "sync_transport").Transport.RuntimeDelivery.Deletes = connectors.DeliveryDeletesTombstone
			},
			want: "delivery policy does not match",
		},
		{
			name: "mode",
			mutate: func(contract *connectors.EnabledConnectorContract, transport *connectors.SyncTransportDescriptor) {
				contractLane(contract, "sync_transport").Transport.Modes = append(contractLane(contract, "sync_transport").Transport.Modes, string(synccontract.ModeIncrementalDedupe))
			},
			want: "modes do not match",
		},
		{
			name: "eligible stream",
			mutate: func(contract *connectors.EnabledConnectorContract, transport *connectors.SyncTransportDescriptor) {
				contractLane(contract, "sync_transport").Transport.Streams[0].Stream = "not-a-loaded-stream"
			},
			want: "eligible streams do not match",
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
		"direct_read: implemented (source coverage: partial 582/749; mapped_unproven=0; unmapped=0; deferred=167; unsupported=0)",
		"reverse_etl: implemented (source coverage: partial 381/1003; mapped_unproven=0; unmapped=0; deferred=622; unsupported=0)",
		"etl: implemented (source coverage: complete 4/4; mapped_unproven=0; unmapped=0; deferred=0; unsupported=0)",
	} {
		if !strings.Contains(manual, want) {
			t.Fatalf("GitLab manual does not disclose source coverage %q:\n%s", want, manual)
		}
	}
}

func TestEnabledContractSchemaAdmitsMappedUnprovenStateAndCoverage(t *testing.T) {
	raw, err := os.ReadFile("../defs/gitlab/enabled_connector_contract.json")
	if err != nil {
		t.Fatalf("read enabled connector contract fixture: %v", err)
	}
	mapped := strings.Replace(string(raw), `"state": "implemented"`, `"state": "mapped_unproven"`, 1)
	mapped = strings.Replace(mapped, `"unmapped_mapping": 0`, `"mapped_unproven": 0, "unmapped_mapping": 0`, 1)
	if err := metaSchemas.enabledConnectorContract.Validate(mustDecodeAny([]byte(mapped))); err != nil {
		t.Fatalf("schema rejects mapped-unproven lane state and coverage: %v", err)
	}

	invalid := strings.Replace(mapped, `"state": "mapped_unproven"`, `"state": "invented_state"`, 1)
	if err := metaSchemas.enabledConnectorContract.Validate(mustDecodeAny([]byte(invalid))); err == nil {
		t.Fatal("schema accepts an unknown enabled-contract lane state")
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

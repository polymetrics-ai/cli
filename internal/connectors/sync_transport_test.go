package connectors

import (
	"context"
	"strings"
	"testing"

	"polymetrics.ai/internal/synccontract"
)

func TestSyncTransportDescriptorResolvesDeclaredApplyStrategy(t *testing.T) {
	descriptor := SyncTransportDescriptor{
		Source: &SourceTransportDescriptor{
			Executor:        TransportExecutorReference{Family: TransportExecutorFamilyNativeAPI, ID: "fake_api_source"},
			EligibleStreams: []string{"records"},
			Modes:           []synccontract.Mode{synccontract.ModeFullAppend},
			Delivery:        closedTestDeliveryGuarantees(),
			Conformance:     closedTestConformanceReference(),
		},
		Destination: &DestinationTransportDescriptor{
			Executor:        TransportExecutorReference{Family: TransportExecutorFamilyNativeDatabase, ID: "fake_database_destination"},
			EligibleActions: []string{"stage_append"},
			Modes:           []synccontract.Mode{synccontract.ModeFullAppend},
			Delivery:        closedTestDeliveryGuarantees(),
			Conformance:     closedTestConformanceReference(),
			Acknowledgement: TransportAcknowledgementDurableWarehouse,
			ApplyStrategies: []DestinationApplyStrategy{{
				Mode:     synccontract.ModeFullAppend,
				Strategy: ApplyStrategyAppend,
				Action:   "stage_append",
			}},
		},
	}

	if err := descriptor.Validate(); err != nil {
		t.Fatalf("Validate() = %v", err)
	}
	strategy, err := descriptor.Destination.ApplyStrategyFor(synccontract.ModeFullAppend)
	if err != nil {
		t.Fatalf("ApplyStrategyFor() = %v", err)
	}
	if strategy.Action != "stage_append" || strategy.Strategy != ApplyStrategyAppend {
		t.Fatalf("declared strategy = %+v, want stage_append/append", strategy)
	}
	if strategy.Action == "upsert" {
		t.Fatal("descriptor resolution silently selected legacy upsert")
	}
}

func TestSyncTransportDescriptorRejectsGenericExecutorReference(t *testing.T) {
	descriptor := SyncTransportDescriptor{Source: &SourceTransportDescriptor{
		Executor:        TransportExecutorReference{Family: TransportExecutorFamilyNativeAPI, ID: "generic_http"},
		EligibleStreams: []string{"records"},
		Modes:           []synccontract.Mode{synccontract.ModeFullAppend},
		Delivery:        closedTestDeliveryGuarantees(),
		Conformance:     closedTestConformanceReference(),
	}}

	err := descriptor.Validate()
	if err == nil || !strings.Contains(err.Error(), "concrete") {
		t.Fatalf("Validate() = %v, want generic executor rejection", err)
	}
}

func TestSyncTransportDescriptorRejectsHyphenatedGenericExecutorReference(t *testing.T) {
	descriptor := SyncTransportDescriptor{Source: &SourceTransportDescriptor{
		Executor:        TransportExecutorReference{Family: TransportExecutorFamilyNativeAPI, ID: "generic-http"},
		EligibleStreams: []string{"records"},
		Modes:           []synccontract.Mode{synccontract.ModeFullAppend},
		Delivery:        closedTestDeliveryGuarantees(),
		Conformance:     closedTestConformanceReference(),
	}}

	err := descriptor.Validate()
	if err == nil || !strings.Contains(err.Error(), "concrete") {
		t.Fatalf("Validate() = %v, want generic executor rejection", err)
	}
}

func TestDestinationTransportDescriptorRejectsStrategyOutsideDeclaredModes(t *testing.T) {
	descriptor := DestinationTransportDescriptor{
		Executor:        TransportExecutorReference{Family: TransportExecutorFamilyNativeDatabase, ID: "fake_database_destination"},
		EligibleActions: []string{"stage_append"},
		Modes:           []synccontract.Mode{synccontract.ModeFullAppend},
		Delivery:        closedTestDeliveryGuarantees(),
		Conformance:     closedTestConformanceReference(),
		Acknowledgement: TransportAcknowledgementDurableWarehouse,
		ApplyStrategies: []DestinationApplyStrategy{
			{Mode: synccontract.ModeFullAppend, Strategy: ApplyStrategyAppend, Action: "stage_append"},
			{Mode: synccontract.ModeFullOverwrite, Strategy: ApplyStrategyReplace, Action: "stage_append"},
		},
	}

	err := descriptor.Validate()
	if err == nil || !strings.Contains(err.Error(), "not a declared destination mode") {
		t.Fatalf("Validate() = %v, want strategy outside declared modes rejection", err)
	}
}

func TestDestinationTransportDescriptorRefusesChangeCapture(t *testing.T) {
	descriptor := DestinationTransportDescriptor{
		Executor:        TransportExecutorReference{Family: TransportExecutorFamilyNativeDatabase, ID: "fake_database_destination"},
		EligibleActions: []string{"stage_change_capture"},
		Modes:           []synccontract.Mode{synccontract.ModeChangeCapture},
		Delivery:        closedTestDeliveryGuarantees(),
		Conformance:     closedTestConformanceReference(),
		Acknowledgement: TransportAcknowledgementDurableWarehouse,
		ApplyStrategies: []DestinationApplyStrategy{{
			Mode:     synccontract.ModeChangeCapture,
			Strategy: ApplyStrategyChangeApply,
			Action:   "stage_change_capture",
		}},
	}

	if err := descriptor.Validate(); err == nil {
		t.Fatalf("Validate() = nil, want destination change_capture refusal before executor registration")
	}
}

func TestSyncTransportGuideProjectsDeclaredRolesWithoutCertificationClaim(t *testing.T) {
	descriptor := &SyncTransportDescriptor{
		Source: &SourceTransportDescriptor{
			Executor:        TransportExecutorReference{Family: TransportExecutorFamilyNativeAPI, ID: "fake_api_source"},
			EligibleStreams: []string{"records"},
			Modes:           []synccontract.Mode{synccontract.ModeFullAppend},
			Delivery:        closedTestDeliveryGuarantees(),
			Conformance:     closedTestConformanceReference(),
		},
		Destination: &DestinationTransportDescriptor{
			Executor:        TransportExecutorReference{Family: TransportExecutorFamilyNativeDatabase, ID: "fake_database_destination"},
			EligibleActions: []string{"stage_append"},
			Modes:           []synccontract.Mode{synccontract.ModeFullAppend},
			Delivery:        closedTestDeliveryGuarantees(),
			Conformance:     closedTestConformanceReference(),
			Acknowledgement: TransportAcknowledgementDurableWarehouse,
			ApplyStrategies: []DestinationApplyStrategy{{Mode: synccontract.ModeFullAppend, Strategy: ApplyStrategyAppend, Action: "stage_append"}},
		},
	}
	connector := &syncTransportGuideConnector{descriptor: descriptor}
	manual := RenderConnectorManual(connector)
	for _, want := range []string{
		"SYNC TRANSPORT",
		"Source transport: declared",
		"Destination transport: declared",
		"not a certification claim",
	} {
		if !strings.Contains(manual, want) {
			t.Fatalf("manual missing %q:\n%s", want, manual)
		}
	}
}

func TestSyncTransportEligibilityProjectsDeclaredNoneAcknowledgement(t *testing.T) {
	for _, acknowledgement := range []TransportAcknowledgement{
		TransportAcknowledgementDurableWarehouse,
		TransportAcknowledgementNone,
	} {
		t.Run(string(acknowledgement), func(t *testing.T) {
			destination := &DestinationTransportDescriptor{
				Executor:        TransportExecutorReference{Family: TransportExecutorFamilyNativeDatabase, ID: "fake_database_destination"},
				EligibleActions: []string{"stage_append"},
				Modes:           []synccontract.Mode{synccontract.ModeFullAppend},
				Delivery:        closedTestDeliveryGuarantees(),
				Conformance:     closedTestConformanceReference(),
				Acknowledgement: acknowledgement,
				ApplyStrategies: []DestinationApplyStrategy{{Mode: synccontract.ModeFullAppend, Strategy: ApplyStrategyAppend, Action: "stage_append"}},
			}
			connector := &syncTransportGuideConnector{descriptor: &SyncTransportDescriptor{Destination: destination}}

			eligibility := SyncTransportEligibilityOf(connector)
			if eligibility.Destination.Status != "declared" {
				t.Fatalf("destination status = %q, want declared", eligibility.Destination.Status)
			}
			if eligibility.Destination.Acknowledgement != acknowledgement {
				t.Fatalf("destination acknowledgement = %q, want %q", eligibility.Destination.Acknowledgement, acknowledgement)
			}
			if eligibility.Destination.Executor == nil || *eligibility.Destination.Executor != destination.Executor {
				t.Fatalf("destination executor = %#v, want %#v", eligibility.Destination.Executor, destination.Executor)
			}
			if len(eligibility.Destination.Actions) != 1 || eligibility.Destination.Actions[0] != "stage_append" {
				t.Fatalf("destination actions = %#v, want [stage_append]", eligibility.Destination.Actions)
			}
			if len(eligibility.Destination.ApplyStrategies) != 1 || eligibility.Destination.ApplyStrategies[0] != destination.ApplyStrategies[0] {
				t.Fatalf("destination apply strategies = %#v, want %#v", eligibility.Destination.ApplyStrategies, destination.ApplyStrategies)
			}
		})
	}
}

func TestSyncTransportEligibilityProjectsValidRolesIndependently(t *testing.T) {
	validSource := func() *SourceTransportDescriptor {
		return &SourceTransportDescriptor{
			Executor:        TransportExecutorReference{Family: TransportExecutorFamilyNativeAPI, ID: "fake_api_source"},
			EligibleStreams: []string{"records"},
			Modes:           []synccontract.Mode{synccontract.ModeFullAppend},
			Delivery:        closedTestDeliveryGuarantees(),
			Conformance:     closedTestConformanceReference(),
		}
	}
	validDestination := func() *DestinationTransportDescriptor {
		return &DestinationTransportDescriptor{
			Executor:        TransportExecutorReference{Family: TransportExecutorFamilyNativeDatabase, ID: "fake_database_destination"},
			EligibleActions: []string{"stage_append"},
			Modes:           []synccontract.Mode{synccontract.ModeFullAppend},
			Delivery:        closedTestDeliveryGuarantees(),
			Conformance:     closedTestConformanceReference(),
			Acknowledgement: TransportAcknowledgementNone,
			ApplyStrategies: []DestinationApplyStrategy{{Mode: synccontract.ModeFullAppend, Strategy: ApplyStrategyAppend, Action: "stage_append"}},
		}
	}

	t.Run("valid destination survives invalid source", func(t *testing.T) {
		source := validSource()
		source.Modes = nil
		connector := &syncTransportGuideConnector{descriptor: &SyncTransportDescriptor{
			Source:      source,
			Destination: validDestination(),
		}}

		eligibility := SyncTransportEligibilityOf(connector)
		if eligibility.Source.Status != "unsupported" {
			t.Fatalf("source status = %q, want unsupported", eligibility.Source.Status)
		}
		if eligibility.Destination.Status != "declared" || eligibility.Destination.Acknowledgement != TransportAcknowledgementNone {
			t.Fatalf("destination eligibility = %#v, want declared acknowledgement none", eligibility.Destination)
		}
	})

	t.Run("valid source survives invalid destination", func(t *testing.T) {
		destination := validDestination()
		destination.ApplyStrategies = nil
		connector := &syncTransportGuideConnector{descriptor: &SyncTransportDescriptor{
			Source:      validSource(),
			Destination: destination,
		}}

		eligibility := SyncTransportEligibilityOf(connector)
		if eligibility.Destination.Status != "unsupported" {
			t.Fatalf("destination status = %q, want unsupported", eligibility.Destination.Status)
		}
		if eligibility.Source.Status != "declared" || eligibility.Source.Executor == nil {
			t.Fatalf("source eligibility = %#v, want declared source", eligibility.Source)
		}
	})
}

type syncTransportGuideConnector struct {
	descriptor *SyncTransportDescriptor
}

func (*syncTransportGuideConnector) Name() string { return "guide_transport" }

func (*syncTransportGuideConnector) Metadata() Metadata {
	return Metadata{Name: "guide_transport", DisplayName: "Guide Transport", Description: "fake closed transport guide", IntegrationType: "api"}
}

func (c *syncTransportGuideConnector) Definition() Definition {
	return Definition{Name: c.Name(), DisplayName: "Guide Transport", Description: "fake closed transport guide", IntegrationType: "api", SyncTransport: c.descriptor}
}

func (*syncTransportGuideConnector) Check(context.Context, RuntimeConfig) error { return nil }

func (*syncTransportGuideConnector) Catalog(context.Context, RuntimeConfig) (Catalog, error) {
	return Catalog{Connector: "guide_transport"}, nil
}

func (*syncTransportGuideConnector) Read(context.Context, ReadRequest, func(Record) error) error {
	return nil
}

func (*syncTransportGuideConnector) Write(context.Context, WriteRequest, []Record) (WriteResult, error) {
	return WriteResult{}, nil
}

func closedTestDeliveryGuarantees() DeliveryGuarantees {
	return DeliveryGuarantees{
		Idempotency: DeliveryIdempotencyKeyed,
		Ordering:    DeliveryOrderingSource,
		Deletes:     DeliveryDeletesTombstone,
	}
}

func closedTestConformanceReference() ConformanceEvidenceReference {
	return ConformanceEvidenceReference{Suite: "external_transport_test", RunID: "verified_run_1"}
}

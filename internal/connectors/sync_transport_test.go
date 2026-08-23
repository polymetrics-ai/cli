package connectors

import (
	"context"
	"encoding/json"
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

func TestDestinationTransportDescriptorSelectsPersistedActionWithinMode(t *testing.T) {
	descriptor := DestinationTransportDescriptor{
		Executor:        TransportExecutorReference{Family: TransportExecutorFamilyDeclarativeAPI, ID: "declarative_typed_destination"},
		EligibleActions: []string{"append_widget", "replace_widget"},
		Modes:           []synccontract.Mode{synccontract.ModeFullAppend},
		Delivery:        closedTestDeliveryGuarantees(),
		Conformance:     closedTestConformanceReference(),
		Acknowledgement: TransportAcknowledgementDurableWarehouse,
		ApplyStrategies: []DestinationApplyStrategy{
			{Mode: synccontract.ModeFullAppend, Strategy: ApplyStrategyAppend, Action: "append_widget"},
			{Mode: synccontract.ModeFullAppend, Strategy: ApplyStrategyReplace, Action: "replace_widget"},
		},
	}
	if err := descriptor.Validate(); err != nil {
		t.Fatalf("Validate() multi-action descriptor = %v", err)
	}
	for action, want := range map[string]ApplyStrategy{
		"append_widget":  ApplyStrategyAppend,
		"replace_widget": ApplyStrategyReplace,
	} {
		strategy, err := descriptor.ApplyStrategyForAction(synccontract.ModeFullAppend, action)
		if err != nil || strategy.Action != action || strategy.Strategy != want {
			t.Fatalf("ApplyStrategyForAction(%q) = (%+v, %v), want exact persisted action and %q", action, strategy, err, want)
		}
	}
	if _, err := descriptor.ApplyStrategyForAction(synccontract.ModeFullAppend, ""); err == nil || !strings.Contains(err.Error(), "persisted action selection") {
		t.Fatalf("empty multi-action selection error = %v, want closed persisted-selection refusal", err)
	}
	if _, err := descriptor.ApplyStrategyForAction(synccontract.ModeFullAppend, "foreign_action"); err == nil || !strings.Contains(err.Error(), "does not declare action") {
		t.Fatalf("foreign action selection error = %v, want closed action refusal", err)
	}
	descriptor.ApplyStrategies = append(descriptor.ApplyStrategies, DestinationApplyStrategy{
		Mode: synccontract.ModeFullAppend, Strategy: ApplyStrategyAppend, Action: "append_widget",
	})
	if err := descriptor.Validate(); err == nil || !strings.Contains(err.Error(), "duplicate apply strategy action") {
		t.Fatalf("duplicate mode/action descriptor error = %v, want refusal", err)
	}
}

// TestDestinationSourceBindingJSONOmitsAbsentBatch protects legacy bindings:
// an omitted optional batch must stay absent on generated/catalog surfaces,
// rather than becoming an invalid zero-value declaration.
func TestDestinationSourceBindingJSONOmitsAbsentBatch(t *testing.T) {
	binding := DestinationSourceBinding{
		Action:          "apply_record",
		Executor:        TransportExecutorReference{Family: TransportExecutorFamilyDeclarativeAPI, ID: "declared_source"},
		EligibleStreams: []string{"records"},
		RecordMapping: SourceRecordMapping{
			Kind:   SourceRecordMappingKindInputFields,
			Inputs: []SourceRecordInputBinding{{Input: "target_id", Field: "id"}},
		},
	}
	encoded, err := json.Marshal(binding)
	if err != nil {
		t.Fatalf("marshal binding: %v", err)
	}
	var projected map[string]json.RawMessage
	if err := json.Unmarshal(encoded, &projected); err != nil {
		t.Fatalf("decode binding: %v", err)
	}
	if _, found := projected["batch"]; found {
		t.Fatalf("absent batch serialized as %s, want omitted optional property", encoded)
	}
}

// TestDestinationTransportDescriptorRequiresExactActionClosure ensures the
// declaration does not advertise an action that cannot be selected and that a
// source binding cannot survive without a reachable strategy/write action.
func TestDestinationTransportDescriptorRequiresExactActionClosure(t *testing.T) {
	base := DestinationTransportDescriptor{
		Executor:        TransportExecutorReference{Family: TransportExecutorFamilyDeclarativeAPI, ID: "declarative_typed_destination"},
		EligibleActions: []string{"apply_record", "ghost_action"},
		Modes:           []synccontract.Mode{synccontract.ModeFullAppend},
		Delivery:        closedTestDeliveryGuarantees(),
		Conformance:     closedTestConformanceReference(),
		Acknowledgement: TransportAcknowledgementDurableWarehouse,
		ApplyStrategies: []DestinationApplyStrategy{{
			Mode:     synccontract.ModeFullAppend,
			Strategy: ApplyStrategyAppend,
			Action:   "apply_record",
		}},
		SourceBindings: []DestinationSourceBinding{{
			Action:          "ghost_action",
			Executor:        TransportExecutorReference{Family: TransportExecutorFamilyDeclarativeAPI, ID: "declared_source"},
			EligibleStreams: []string{"records"},
			RecordMapping: SourceRecordMapping{
				Kind:   SourceRecordMappingKindInputFields,
				Inputs: []SourceRecordInputBinding{{Input: "target_id", Field: "id"}},
			},
			Batch: &DestinationBatch{Disposition: DestinationBatchPerRecord, MaxRecords: 1},
		}},
	}
	if err := base.Validate(); err == nil || !strings.Contains(err.Error(), "eligible action") {
		t.Fatalf("ghost eligible action error = %v, want exact strategy closure refusal", err)
	}

	base.EligibleActions = []string{"apply_record"}
	if err := base.Validate(); err == nil || !strings.Contains(err.Error(), "source binding action") {
		t.Fatalf("orphan source binding error = %v, want reachable action refusal", err)
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

func TestDestinationTransportDescriptorRefusesChangeCaptureDestinationMode(t *testing.T) {
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

	if err := descriptor.Validate(); err == nil || !strings.Contains(err.Error(), "change_capture") {
		t.Fatalf("Validate() = %v, want destination change_capture role refusal", err)
	}
}

func TestDestinationTransportDescriptorRefusesChangeCaptureDestinationModeRegardlessOfStrategy(t *testing.T) {
	descriptor := DestinationTransportDescriptor{
		Executor:        TransportExecutorReference{Family: TransportExecutorFamilyNativeDatabase, ID: "fake_database_destination"},
		EligibleActions: []string{"stage_change_capture"},
		Modes:           []synccontract.Mode{synccontract.ModeChangeCapture},
		Delivery:        closedTestDeliveryGuarantees(),
		Conformance:     closedTestConformanceReference(),
		Acknowledgement: TransportAcknowledgementDurableWarehouse,
		ApplyStrategies: []DestinationApplyStrategy{{
			Mode:     synccontract.ModeChangeCapture,
			Strategy: ApplyStrategyAppend,
			Action:   "stage_change_capture",
		}},
	}

	if err := descriptor.Validate(); err == nil || !strings.Contains(err.Error(), "change_capture") {
		t.Fatalf("Validate() = %v, want destination change_capture role refusal", err)
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

package app

import (
	"errors"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/synccontract"
)

func TestResolveTargetCopyWorkersUsesDeclaredPoolBoundForImmutableOverwrite(t *testing.T) {
	destination := &appTransportConnector{
		meta: connectors.Metadata{Name: "copy_target", IntegrationType: "database"},
		descriptor: &connectors.SyncTransportDescriptor{Destination: &connectors.DestinationTransportDescriptor{
			Executor:          connectors.TransportExecutorReference{Family: connectors.TransportExecutorFamilyNativeDatabase, ID: "copy_target"},
			CopyWorkerMaximum: 3,
		}},
	}
	streams := map[string]StreamConfig{"records": {
		SyncMode: string(synccontract.ModeFullOverwrite), TransformPlan: `{"version":1}`, TransformPlanHash: "hash",
	}}

	got, maximum, err := resolveTargetCopyWorkers(destination, streams, 0)
	if err != nil || got != 2 || maximum != 3 {
		t.Fatalf("default target workers = (%d, %d, %v), want (2, 3, nil)", got, maximum, err)
	}
	got, maximum, err = resolveTargetCopyWorkers(destination, streams, 3)
	if err != nil || got != 3 || maximum != 3 {
		t.Fatalf("declared-bound target workers = (%d, %d, %v), want (3, 3, nil)", got, maximum, err)
	}
	_, _, err = resolveTargetCopyWorkers(destination, streams, 4)
	var rangeErr *TargetCopyWorkersRangeError
	if !errors.As(err, &rangeErr) || rangeErr.Maximum != 3 {
		t.Fatalf("over-declared target workers error = %T %v, want max-3 TargetCopyWorkersRangeError", err, err)
	}
}

func TestResolveTargetCopyWorkersRefusesNonImmutableOrUndeclaredDestination(t *testing.T) {
	noCopyDestination := &appTransportConnector{meta: connectors.Metadata{Name: "no_copy_target", IntegrationType: "database"}}
	overwrite := map[string]StreamConfig{"records": {
		SyncMode: string(synccontract.ModeFullOverwrite), TransformPlan: `{"version":1}`, TransformPlanHash: "hash",
	}}
	if got, maximum, err := resolveTargetCopyWorkers(noCopyDestination, overwrite, 0); err != nil || got != 0 || maximum != 0 {
		t.Fatalf("absent target declaration default = (%d, %d, %v), want (0, 0, nil)", got, maximum, err)
	}
	_, _, err := resolveTargetCopyWorkers(noCopyDestination, overwrite, 1)
	var unsupported *TargetCopyWorkersUnsupportedError
	if !errors.As(err, &unsupported) {
		t.Fatalf("undeclared target workers error = %T %v, want TargetCopyWorkersUnsupportedError", err, err)
	}

	copyDestination := &appTransportConnector{
		meta:       connectors.Metadata{Name: "copy_target", IntegrationType: "database"},
		descriptor: &connectors.SyncTransportDescriptor{Destination: &connectors.DestinationTransportDescriptor{CopyWorkerMaximum: 2}},
	}
	nonOverwrite := map[string]StreamConfig{"records": {SyncMode: string(synccontract.ModeFullAppend)}}
	if got, maximum, err := resolveTargetCopyWorkers(copyDestination, nonOverwrite, 0); err != nil || got != 0 || maximum != 2 {
		t.Fatalf("non-overwrite target workers default = (%d, %d, %v), want (0, 2, nil)", got, maximum, err)
	}
	_, _, err = resolveTargetCopyWorkers(copyDestination, nonOverwrite, 1)
	if !errors.As(err, &unsupported) {
		t.Fatalf("non-overwrite target workers error = %T %v, want TargetCopyWorkersUnsupportedError", err, err)
	}
}

package connectors

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestEnabledContractReconcilesExactSemanticSourcePartitions(t *testing.T) {
	contract := semanticPartitionContract()
	operations := []EnabledContractSourceOperation{
		// Both documented reads use POST, proving that HTTP method alone cannot
		// decide the semantic lane. The Vercel create operation is also POST,
		// but belongs to the mutation partition.
		{ID: "jira.rest.getCustomFieldsConfigurations", Method: "POST"},
		{ID: "vercel.rest.artifactQuery", Method: "POST"},
		{ID: "vercel.rest.createAccessGroup", Method: "POST"},
	}

	if err := contract.ReconcileSourceOperations(operations); err != nil {
		t.Fatalf("reconcile semantic POST read and POST mutation partitions: %v", err)
	}

	for _, want := range []struct {
		name  string
		state string
	}{
		{name: "direct_read", state: EnabledLaneMappedUnproven},
		{name: "direct_write", state: EnabledLaneMappedUnproven},
		{name: "binary_download", state: EnabledLaneUnsupported},
		{name: "binary_upload", state: EnabledLaneUnsupported},
		{name: "etl", state: EnabledLaneMappedUnproven},
		{name: "reverse_etl", state: EnabledLaneMappedUnproven},
		{name: "sync_transport", state: EnabledLaneUnsupported},
	} {
		if lane := enabledContractTestLane(&contract, want.name); lane.State != want.state {
			t.Fatalf("%s state = %q, want %q", want.name, lane.State, want.state)
		}
	}

	if lane := enabledContractTestLane(&contract, "direct_read"); lane.State != EnabledLaneMappedUnproven || lane.Source.MappedUnproven != 2 || lane.Source.DeferredFoundation != 0 {
		t.Fatalf("direct-read semantic partition = %+v, want distinct mapped-unproven coverage", lane)
	}
	if lane := enabledContractTestLane(&contract, "reverse_etl"); lane.State != EnabledLaneMappedUnproven || lane.Source.MappedUnproven != 1 || lane.Source.DeferredFoundation != 0 {
		t.Fatalf("reverse-ETL semantic partition = %+v, want distinct mapped-unproven coverage", lane)
	}
}

func TestEnabledContractKeepsLegacyMethodPartitions(t *testing.T) {
	contract := legacyMethodPartitionContract()
	operations := []EnabledContractSourceOperation{
		{ID: "legacy.rest.getThing", Method: "GET"},
		{ID: "legacy.rest.createThing", Method: "POST"},
	}

	if err := contract.ReconcileSourceOperations(operations); err != nil {
		t.Fatalf("reconcile legacy HTTP-method partitions: %v", err)
	}
}

func TestEnabledContractMappedUnprovenStateRoundTrips(t *testing.T) {
	encoded, err := json.Marshal(semanticPartitionContract())
	if err != nil {
		t.Fatalf("marshal semantic contract: %v", err)
	}
	if !strings.Contains(string(encoded), `"state":"mapped_unproven"`) || !strings.Contains(string(encoded), `"mapped_unproven":2`) {
		t.Fatalf("serialized contract does not preserve mapped-unproven state and coverage: %s", encoded)
	}

	var decoded EnabledConnectorContract
	if err := json.Unmarshal(encoded, &decoded); err != nil {
		t.Fatalf("unmarshal semantic contract: %v", err)
	}
	if err := decoded.Validate(); err != nil {
		t.Fatalf("validate mapped-unproven contract: %v", err)
	}
}

func TestEnabledContractRejectsInvalidSemanticPartitions(t *testing.T) {
	operations := []EnabledContractSourceOperation{
		{ID: "jira.rest.getCustomFieldsConfigurations", Method: "POST"},
		{ID: "vercel.rest.artifactQuery", Method: "POST"},
		{ID: "vercel.rest.createAccessGroup", Method: "POST"},
	}

	for _, test := range []struct {
		name   string
		mutate func(*EnabledConnectorContract)
		want   string
	}{
		{
			name: "duplicate source ID across exact partitions",
			mutate: func(contract *EnabledConnectorContract) {
				reverseETL := enabledContractTestLane(contract, "reverse_etl")
				reverseETL.Source.OperationIDs = append(reverseETL.Source.OperationIDs, "vercel.rest.artifactQuery")
				reverseETL.Source.Expected++
				reverseETL.Source.MappedUnproven++
			},
			want: "claimed by both",
		},
		{
			name: "locked source ID omitted from exact partitions",
			mutate: func(contract *EnabledConnectorContract) {
				lane := enabledContractTestLane(contract, "direct_read")
				lane.Source.OperationIDs = lane.Source.OperationIDs[:1]
				lane.Source.Expected--
				lane.Source.MappedUnproven--
			},
			want: "has no partition lane",
		},
		{
			name: "partition cannot combine method and source-ID selectors",
			mutate: func(contract *EnabledConnectorContract) {
				enabledContractTestLane(contract, "direct_read").Source.Methods = []string{"GET"}
			},
			want: "cannot combine methods and operation IDs",
		},
		{
			name: "partitions cannot mix selector styles",
			mutate: func(contract *EnabledConnectorContract) {
				directRead := enabledContractTestLane(contract, "direct_read")
				directRead.Source.OperationIDs = nil
				directRead.Source.Methods = []string{"POST"}
			},
			want: "cannot mix method and operation-ID selectors",
		},
		{
			name: "invalid lane state",
			mutate: func(contract *EnabledConnectorContract) {
				enabledContractTestLane(contract, "direct_read").State = "invented_state"
			},
			want: "state \"invented_state\" is invalid",
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			contract := semanticPartitionContract()
			test.mutate(&contract)
			err := contract.ReconcileSourceOperations(operations)
			if err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("semantic partition error = %v, want %q", err, test.want)
			}
		})
	}
}

func semanticPartitionContract() EnabledConnectorContract {
	return enabledContractTestContract(
		EnabledConnectorLane{
			Name:      "direct_read",
			State:     EnabledLaneMappedUnproven,
			Reason:    "Documented Jira and Vercel POST read semantics identify provider reads; neither is an executable declaration.",
			Citations: enabledContractTestCitations(),
			Artifacts: []string{"operations.json"},
			Source: EnabledContractSourceCoverage{
				Partition:      true,
				OperationIDs:   []string{"jira.rest.getCustomFieldsConfigurations", "vercel.rest.artifactQuery"},
				Coverage:       EnabledCoveragePartial,
				Expected:       2,
				MappedUnproven: 2,
			},
		},
		EnabledConnectorLane{
			Name:      "reverse_etl",
			State:     EnabledLaneMappedUnproven,
			Reason:    "The documented Vercel POST mutation remains mapped-unproven until a typed write binding exists.",
			Citations: enabledContractTestCitations(),
			Artifacts: []string{"writes.json"},
			Source: EnabledContractSourceCoverage{
				Partition:      true,
				OperationIDs:   []string{"vercel.rest.createAccessGroup"},
				Coverage:       EnabledCoveragePartial,
				Expected:       1,
				MappedUnproven: 1,
			},
		},
		EnabledConnectorLane{
			Name:      "direct_write",
			State:     EnabledLaneMappedUnproven,
			Reason:    "The same semantic mutation is a one-record direct-write candidate, not an executable declaration.",
			Citations: enabledContractTestCitations(),
			Artifacts: []string{"writes.json"},
			Source: EnabledContractSourceCoverage{
				OperationIDs:   []string{"vercel.rest.createAccessGroup"},
				Coverage:       EnabledCoveragePartial,
				Expected:       1,
				MappedUnproven: 1,
			},
		},
		EnabledConnectorLane{
			Name:      "etl",
			State:     EnabledLaneMappedUnproven,
			Reason:    "The source-cited continuation is a mapped ETL candidate pending a typed stream declaration.",
			Citations: enabledContractTestCitations(),
			Artifacts: []string{"streams.json"},
			Source: EnabledContractSourceCoverage{
				OperationIDs:   []string{"jira.rest.getCustomFieldsConfigurations"},
				Coverage:       EnabledCoveragePartial,
				Expected:       1,
				MappedUnproven: 1,
			},
		},
		enabledContractTestNotApplicableLane("binary_download"),
		enabledContractTestNotApplicableLane("binary_upload"),
		enabledContractTestNotApplicableLane("sync_transport"),
	)
}

func legacyMethodPartitionContract() EnabledConnectorContract {
	contract := semanticPartitionContract()
	directRead := enabledContractTestLane(&contract, "direct_read")
	directRead.Source.OperationIDs = nil
	directRead.Source.Methods = []string{"GET"}
	directRead.Source.Expected = 1
	directRead.Source.MappedUnproven = 1

	reverseETL := enabledContractTestLane(&contract, "reverse_etl")
	reverseETL.Source.OperationIDs = nil
	reverseETL.Source.Methods = []string{"POST"}

	enabledContractTestLane(&contract, "direct_write").Source.OperationIDs = []string{"legacy.rest.createThing"}
	enabledContractTestLane(&contract, "etl").Source.OperationIDs = []string{"legacy.rest.getThing"}

	return contract
}

func enabledContractTestContract(lanes ...EnabledConnectorLane) EnabledConnectorContract {
	return EnabledConnectorContract{
		SchemaVersion: 1,
		Connector:     "semantic",
		SourceLock: EnabledContractSourceLock{
			Path:              "sources/semantic-operation-source-lock.json",
			SHA256:            strings.Repeat("a", 64),
			Bytes:             1,
			CanonicalEvidence: true,
		},
		Lanes: lanes,
	}
}

func enabledContractTestNotApplicableLane(name string) EnabledConnectorLane {
	return EnabledConnectorLane{
		Name:      name,
		State:     EnabledLaneUnsupported,
		Reason:    "The retained provider source does not document this lane.",
		Citations: enabledContractTestCitations(),
		Artifacts: []string{"sources/semantic-operation-source-lock.json"},
		Source:    EnabledContractSourceCoverage{Coverage: EnabledCoverageNotApplicable},
	}
}

func enabledContractTestCitations() []EnabledContractCitation {
	return []EnabledContractCitation{{URL: "https://example.invalid/source", Location: "paths"}}
}

func enabledContractTestLane(contract *EnabledConnectorContract, name string) *EnabledConnectorLane {
	for index := range contract.Lanes {
		if contract.Lanes[index].Name == name {
			return &contract.Lanes[index]
		}
	}
	panic("missing lane " + name)
}

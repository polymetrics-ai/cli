package certificationcatalog

import "testing"

func TestFlowKindsFailsClosedWithoutValidGeneratedData(t *testing.T) {
	originalData := append([]FlowKind(nil), generatedFlowKindData...)
	originalSet := generatedFlowKindDataSet
	t.Cleanup(func() {
		generatedFlowKindData = originalData
		generatedFlowKindDataSet = originalSet
	})

	generatedFlowKindData = nil
	generatedFlowKindDataSet = false
	if got := FlowKinds(); len(got) != 0 {
		t.Fatalf("FlowKinds without generated data = %#v, want no catalog", got)
	}

	generatedFlowKindData = []FlowKind{}
	generatedFlowKindDataSet = true
	if got := FlowKinds(); len(got) != 0 {
		t.Fatalf("FlowKinds with empty generated data = %#v, want no catalog", got)
	}

	generatedFlowKindData = []FlowKind{{ID: "api to api", SourceRole: "api_source", DestinationRole: "api_destination"}}
	generatedFlowKindDataSet = true
	if got := FlowKinds(); len(got) != 0 {
		t.Fatalf("FlowKinds with invalid generated data = %#v, want no catalog", got)
	}

	generatedFlowKindData = []FlowKind{
		{ID: "api_to_api", SourceRole: "api_source", DestinationRole: "api_destination"},
		{ID: "api_to_api", SourceRole: "database_source", DestinationRole: "database_destination"},
	}
	if got := FlowKinds(); len(got) != 0 {
		t.Fatalf("FlowKinds with duplicate generated data = %#v, want no catalog", got)
	}

	generatedFlowKindData = []FlowKind{{ID: "api_to_api", SourceRole: "api_source", DestinationRole: "api_destination"}}
	got := FlowKinds()
	if len(got) != 1 {
		t.Fatalf("FlowKinds valid data = %#v, want one kind", got)
	}
	got[0].ID = "mutated"
	if later := FlowKinds(); len(later) != 1 || later[0].ID != "api_to_api" {
		t.Fatalf("FlowKinds exposed generated data for mutation: %#v", later)
	}
}

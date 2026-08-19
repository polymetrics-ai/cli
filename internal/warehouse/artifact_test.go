package warehouse

import "testing"

func TestArtifactRefIsConnectorAgnosticAndStructurallyBound(t *testing.T) {
	identity := ArtifactIdentity{
		WorkspaceID:  "workspace_1",
		ConnectorID:  "api-source",
		ConnectionID: "connection_1",
	}
	artifact, err := NewArtifactRef(identity, "records")
	if err != nil {
		t.Fatalf("NewArtifactRef() error = %v", err)
	}
	if got := artifact.Identity(); got != identity {
		t.Fatalf("ArtifactRef.Identity() = %#v, want %#v", got, identity)
	}
	if got := artifact.Table(); got != "records" {
		t.Fatalf("ArtifactRef.Table() = %q, want records", got)
	}

	for _, tc := range []struct {
		name     string
		identity ArtifactIdentity
		table    string
	}{
		{name: "unsafe workspace", identity: ArtifactIdentity{WorkspaceID: "../workspace", ConnectorID: "api-source", ConnectionID: "connection_1"}, table: "records"},
		{name: "unsafe connector", identity: ArtifactIdentity{WorkspaceID: "workspace_1", ConnectorID: "api/source", ConnectionID: "connection_1"}, table: "records"},
		{name: "missing connection", identity: ArtifactIdentity{WorkspaceID: "workspace_1", ConnectorID: "api-source"}, table: "records"},
		{name: "unsafe table", identity: identity, table: "../records"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := NewArtifactRef(tc.identity, tc.table); err == nil {
				t.Fatal("NewArtifactRef() error = nil, want structural refusal")
			}
		})
	}
}

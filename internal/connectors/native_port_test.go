package connectors

import (
	"strings"
	"testing"
)

func TestNativePortPlansCoverCatalog(t *testing.T) {
	plans := NativePortPlans(ConnectorCatalog())
	if got, want := len(plans), 646; got != want {
		t.Fatalf("NativePortPlans len = %d, want %d", got, want)
	}
	for _, plan := range plans {
		if plan.Slug == "" || plan.Name == "" || plan.Family == "" || plan.RuntimeKind == "" || plan.ImplementationStatus == "" {
			t.Fatalf("incomplete native port plan: %+v", plan)
		}
		if len(plan.Conformance) == 0 {
			t.Fatalf("native port plan missing conformance requirements: %+v", plan)
		}
		if plan.ImplementationStatus != ImplementationEnabled && len(plan.ImplementationNotes) == 0 {
			t.Fatalf("planned native port missing implementation notes: %+v", plan)
		}
	}
}

func TestDatabaseCDCNativePortPlans(t *testing.T) {
	tests := []struct {
		slug        string
		wantMode    string
		wantState   string
		wantRequire string
	}{
		{slug: "source-postgres", wantMode: "postgres_logical_replication", wantState: "lsn", wantRequire: "wal_level=logical"},
		{slug: "source-mysql", wantMode: "mysql_binlog", wantState: "gtid_or_binlog_position", wantRequire: "binlog_format=ROW"},
		{slug: "source-mongodb-v2", wantMode: "mongodb_change_streams", wantState: "resume_token", wantRequire: "replica set or sharded cluster"},
	}
	for _, tt := range tests {
		t.Run(tt.slug, func(t *testing.T) {
			def, ok := ConnectorDefinitionBySlug(tt.slug)
			if !ok {
				t.Fatalf("%s not found", tt.slug)
			}
			plan := NativePortPlanForDefinition(def)
			if plan.Family != PortFamilyDatabaseCDCSource {
				t.Fatalf("%s family = %s, want %s", tt.slug, plan.Family, PortFamilyDatabaseCDCSource)
			}
			if !plan.CDC.Supported {
				t.Fatalf("%s CDC not supported in plan: %+v", tt.slug, plan.CDC)
			}
			if !contains(plan.CDC.Modes, tt.wantMode) {
				t.Fatalf("%s CDC modes = %v, want %s", tt.slug, plan.CDC.Modes, tt.wantMode)
			}
			if !contains(plan.CDC.StateFields, tt.wantState) {
				t.Fatalf("%s CDC state fields = %v, want %s", tt.slug, plan.CDC.StateFields, tt.wantState)
			}
			requirements := strings.Join(plan.CDC.Requirements, "\n")
			if !strings.Contains(requirements, tt.wantRequire) {
				t.Fatalf("%s CDC requirements = %v, want %q", tt.slug, plan.CDC.Requirements, tt.wantRequire)
			}
		})
	}
}

func TestNativePortPlanEnabledGithub(t *testing.T) {
	def, ok := ConnectorDefinitionBySlug("source-github")
	if !ok {
		t.Fatal("source-github not found")
	}
	plan := NativePortPlanForDefinition(def)
	if plan.Family != PortFamilyNativeSaaS || plan.PriorityWave != 0 {
		t.Fatalf("github plan = %+v", plan)
	}
	if !contains(plan.ETLOperations, "read") || !contains(plan.ReverseETLOperations, "write") {
		t.Fatalf("github plan missing ETL/reverse ETL operations: %+v", plan)
	}
	if plan.CDC.Supported {
		t.Fatalf("github CDC supported = true")
	}
}

func contains(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}

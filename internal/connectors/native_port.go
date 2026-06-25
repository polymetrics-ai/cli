package connectors

import (
	"fmt"
	"sort"
	"strings"
)

type PortFamily string

const (
	PortFamilyNativeSaaS            PortFamily = "native_saas"
	PortFamilyDeclarativeHTTPSource PortFamily = "declarative_http_source"
	PortFamilyDatabaseCDCSource     PortFamily = "database_cdc_source"
	PortFamilyDatabaseSource        PortFamily = "database_source"
	PortFamilyFileObjectSource      PortFamily = "file_object_source"
	PortFamilyDestinationWriter     PortFamily = "destination_writer"
	PortFamilyCustomGoPort          PortFamily = "custom_go_port"
)

type CDCPlan struct {
	Supported       bool     `json:"supported"`
	Modes           []string `json:"modes,omitempty"`
	Requirements    []string `json:"requirements,omitempty"`
	StateFields     []string `json:"state_fields,omitempty"`
	DeleteSemantics string   `json:"delete_semantics,omitempty"`
	Ordering        string   `json:"ordering,omitempty"`
}

type NativePortPlan struct {
	Slug                 string               `json:"slug"`
	Name                 string               `json:"name"`
	Type                 ConnectorType        `json:"type"`
	DocumentationURL     string               `json:"documentation_url"`
	Family               PortFamily           `json:"family"`
	RuntimeKind          RuntimeKind          `json:"runtime_kind"`
	ImplementationStatus ImplementationStatus `json:"implementation_status"`
	PriorityWave         int                  `json:"priority_wave"`
	ETLOperations        []string             `json:"etl_operations,omitempty"`
	ReverseETLOperations []string             `json:"reverse_etl_operations,omitempty"`
	CDC                  CDCPlan              `json:"cdc"`
	Conformance          []string             `json:"conformance"`
	ImplementationNotes  []string             `json:"implementation_notes,omitempty"`
}

type NativePortPlanSummary struct {
	Total               int            `json:"total"`
	Families            map[string]int `json:"families"`
	Waves               map[string]int `json:"waves"`
	CDCEnabledPlans     int            `json:"cdc_enabled_plans"`
	EnabledRuntimePlans int            `json:"enabled_runtime_plans"`
}

func NativePortPlans(defs []ConnectorDefinition) []NativePortPlan {
	plans := make([]NativePortPlan, 0, len(defs))
	for _, def := range defs {
		plans = append(plans, NativePortPlanForDefinition(def))
	}
	sort.SliceStable(plans, func(i, j int) bool { return plans[i].Slug < plans[j].Slug })
	return plans
}

func NativePortPlanBySlug(slug string) (NativePortPlan, bool) {
	def, ok := ConnectorDefinitionBySlug(slug)
	if !ok {
		return NativePortPlan{}, false
	}
	return NativePortPlanForDefinition(def), true
}

func NativePortPlanForDefinition(def ConnectorDefinition) NativePortPlan {
	cdc := CDCPlanForDefinition(def)
	family := portFamilyForDefinition(def, cdc)
	plan := NativePortPlan{
		Slug:                 def.Slug,
		Name:                 def.Name,
		Type:                 def.Type,
		DocumentationURL:     def.DocumentationURL,
		Family:               family,
		RuntimeKind:          def.RuntimeKind,
		ImplementationStatus: def.ImplementationStatus,
		PriorityWave:         priorityWave(def),
		ETLOperations:        etlOperationsForDefinition(def, cdc),
		ReverseETLOperations: reverseETLOperationsForDefinition(def),
		CDC:                  cdc,
		Conformance:          conformanceForDefinition(def, family, cdc),
		ImplementationNotes:  implementationNotesForDefinition(def, family, cdc),
	}
	sort.Strings(plan.ETLOperations)
	sort.Strings(plan.ReverseETLOperations)
	sort.Strings(plan.Conformance)
	return plan
}

func NativePortPlanCounts(plans []NativePortPlan) NativePortPlanSummary {
	summary := NativePortPlanSummary{
		Total:    len(plans),
		Families: map[string]int{},
		Waves:    map[string]int{},
	}
	for _, plan := range plans {
		summary.Families[string(plan.Family)]++
		summary.Waves[fmt.Sprintf("wave_%d", plan.PriorityWave)]++
		if plan.CDC.Supported {
			summary.CDCEnabledPlans++
		}
		if plan.ImplementationStatus == ImplementationEnabled {
			summary.EnabledRuntimePlans++
		}
	}
	return summary
}

func CDCPlanForDefinition(def ConnectorDefinition) CDCPlan {
	if def.Type != ConnectorTypeSource {
		return CDCPlan{Supported: false}
	}
	slug := strings.ToLower(def.Slug)
	switch {
	case strings.Contains(slug, "postgres"):
		return CDCPlan{
			Supported:       true,
			Modes:           []string{"snapshot", "postgres_logical_replication"},
			Requirements:    []string{"wal_level=logical", "replication role or equivalent permissions", "replication slot", "publication for selected tables", "primary key or REPLICA IDENTITY FULL for update/delete fidelity"},
			StateFields:     []string{"lsn", "slot_name", "publication_name", "snapshot_completed"},
			DeleteSemantics: "logical replication can emit deletes when replica identity is configured for the table",
			Ordering:        "WAL LSN order per replication stream",
		}
	case strings.Contains(slug, "mysql") || strings.Contains(slug, "tidb"):
		return CDCPlan{
			Supported:       true,
			Modes:           []string{"snapshot", "mysql_binlog"},
			Requirements:    []string{"binlog_format=ROW", "binlog_row_image=FULL for complete update/delete records", "GTID enabled or durable binlog file/position state", "replication user with REPLICATION SLAVE/REPLICATION CLIENT permissions"},
			StateFields:     []string{"gtid_or_binlog_position", "server_id", "snapshot_completed"},
			DeleteSemantics: "row-based binlog can emit deletes and before/after update images when configured with full row image",
			Ordering:        "binlog event order by GTID or file/position",
		}
	case strings.Contains(slug, "mongodb"):
		return CDCPlan{
			Supported:       true,
			Modes:           []string{"snapshot", "mongodb_change_streams"},
			Requirements:    []string{"replica set or sharded cluster", "change stream privileges", "durable resume token storage", "same pipeline and options when resuming a stream"},
			StateFields:     []string{"resume_token", "cluster_time", "snapshot_completed"},
			DeleteSemantics: "change streams emit delete operations but deleted document contents depend on pre-image configuration",
			Ordering:        "change stream order within the watched deployment and resume token sequence",
		}
	case strings.Contains(slug, "mssql") || strings.Contains(slug, "sql-server") || strings.Contains(slug, "sqlserver"):
		return CDCPlan{
			Supported:       true,
			Modes:           []string{"snapshot", "sql_server_cdc"},
			Requirements:    []string{"SQL Server CDC enabled for database", "CDC enabled for selected tables", "permissions to read CDC change tables", "retention window longer than sync outage budget"},
			StateFields:     []string{"lsn", "capture_instance", "snapshot_completed"},
			DeleteSemantics: "SQL Server CDC records insert, update, and delete operations in change tables",
			Ordering:        "LSN order per capture instance",
		}
	case strings.Contains(slug, "oracle"):
		return CDCPlan{
			Supported:       true,
			Modes:           []string{"snapshot", "oracle_logminer_or_xstream"},
			Requirements:    []string{"archive logging enabled", "supplemental logging for selected tables", "LogMiner or XStream permissions", "SCN checkpoint retention compatible with sync interval"},
			StateFields:     []string{"scn", "snapshot_completed"},
			DeleteSemantics: "redo mining can capture deletes when supplemental logging is sufficient",
			Ordering:        "SCN order",
		}
	default:
		if def.SourceType == "database" || def.RuntimeKind == RuntimeDatabaseGo {
			return CDCPlan{
				Supported:       false,
				Modes:           []string{"snapshot", "cursor_incremental"},
				Requirements:    []string{"stable cursor field or database-specific CDC implementation before delete propagation"},
				StateFields:     []string{"cursor", "snapshot_completed"},
				DeleteSemantics: "delete detection is unavailable without connector-specific CDC support",
				Ordering:        "cursor order when configured",
			}
		}
		return CDCPlan{Supported: false}
	}
}

func RenderNativePortPlanManual(plan NativePortPlan) string {
	var b strings.Builder
	b.WriteString("NAME\n")
	b.WriteString("  pm connectors port-plan " + plan.Slug + " - " + plan.Name + " native Go implementation plan\n\n")
	b.WriteString("SYNOPSIS\n")
	b.WriteString("  pm connectors port-plan " + plan.Slug + "\n")
	b.WriteString("  pm connectors port-plan " + plan.Slug + " --json\n\n")
	b.WriteString("DESCRIPTION\n")
	b.WriteString("  Shows the native Go port plan for this catalog connector. This is planning and conformance metadata; it does not enable runtime operations by itself.\n\n")
	b.WriteString("NATIVE PORT PLAN\n")
	b.WriteString("  family: " + string(plan.Family) + "\n")
	b.WriteString("  runtime_kind: " + string(plan.RuntimeKind) + "\n")
	b.WriteString("  implementation_status: " + string(plan.ImplementationStatus) + "\n")
	b.WriteString(fmt.Sprintf("  priority_wave: %d\n", plan.PriorityWave))
	for _, note := range plan.ImplementationNotes {
		b.WriteString("  note: " + note + "\n")
	}
	b.WriteString("\nETL\n")
	writeManualList(&b, plan.ETLOperations, "planned ETL operations")
	b.WriteString("\nREVERSE ETL\n")
	writeManualList(&b, plan.ReverseETLOperations, "planned reverse ETL operations")
	b.WriteString("\nCDC\n")
	if plan.CDC.Supported {
		b.WriteString("  supported=true\n")
		writeManualList(&b, plan.CDC.Modes, "modes")
		writeManualList(&b, plan.CDC.Requirements, "requirements")
		writeManualList(&b, plan.CDC.StateFields, "state fields")
		if plan.CDC.DeleteSemantics != "" {
			b.WriteString("  delete_semantics: " + plan.CDC.DeleteSemantics + "\n")
		}
		if plan.CDC.Ordering != "" {
			b.WriteString("  ordering: " + plan.CDC.Ordering + "\n")
		}
	} else {
		b.WriteString("  supported=false\n")
		if len(plan.CDC.Modes) > 0 {
			writeManualList(&b, plan.CDC.Modes, "modes")
		}
		if plan.CDC.DeleteSemantics != "" {
			b.WriteString("  delete_semantics: " + plan.CDC.DeleteSemantics + "\n")
		}
	}
	b.WriteString("\nCONFORMANCE\n")
	writeManualList(&b, plan.Conformance, "required tests")
	b.WriteString("\nSECURITY\n")
	b.WriteString("  Do not request or print secret values while porting connectors.\n")
	b.WriteString("  Do not enable read or write runtime capabilities until conformance tests pass.\n")
	b.WriteString("  Reverse ETL writes must remain plan, preview, approve, then execute.\n\n")
	if plan.DocumentationURL != "" {
		b.WriteString("SEE ALSO\n")
		b.WriteString("  " + plan.DocumentationURL + "\n\n")
	}
	b.WriteString("EXIT STATUS\n")
	b.WriteString("  0 success\n")
	b.WriteString("  1 runtime error\n")
	b.WriteString("  2 usage error\n")
	return b.String()
}

func nativePortPlanSection(def ConnectorDefinition) GuideSection {
	plan := NativePortPlanForDefinition(def)
	lines := []string{
		"family: " + string(plan.Family),
		"priority_wave: " + fmt.Sprint(plan.PriorityWave),
		"etl_operations: " + strings.Join(plan.ETLOperations, ", "),
	}
	if len(plan.ReverseETLOperations) > 0 {
		lines = append(lines, "reverse_etl_operations: "+strings.Join(plan.ReverseETLOperations, ", "))
	} else {
		lines = append(lines, "reverse_etl_operations: none until native write conformance passes")
	}
	if plan.CDC.Supported {
		lines = append(lines, "cdc_modes: "+strings.Join(plan.CDC.Modes, ", "))
		lines = append(lines, "cdc_state_fields: "+strings.Join(plan.CDC.StateFields, ", "))
	}
	lines = append(lines, "conformance: "+strings.Join(plan.Conformance, ", "))
	return GuideSection{Title: "Native Port Plan", Lines: lines}
}

func portFamilyForDefinition(def ConnectorDefinition, cdc CDCPlan) PortFamily {
	if def.ImplementationStatus == ImplementationEnabled && def.PMConnectorName != "" {
		return PortFamilyNativeSaaS
	}
	if cdc.Supported {
		return PortFamilyDatabaseCDCSource
	}
	if def.Type == ConnectorTypeDestination {
		return PortFamilyDestinationWriter
	}
	switch def.RuntimeKind {
	case RuntimeDeclarativeHTTPGo:
		return PortFamilyDeclarativeHTTPSource
	case RuntimeDatabaseGo:
		return PortFamilyDatabaseSource
	case RuntimeFileGo:
		return PortFamilyFileObjectSource
	default:
		return PortFamilyCustomGoPort
	}
}

func priorityWave(def ConnectorDefinition) int {
	if def.ImplementationStatus == ImplementationEnabled {
		return 0
	}
	switch def.ReleaseStage {
	case "generally_available":
		return 1
	case "beta":
		return 2
	default:
		return 3
	}
}

func etlOperationsForDefinition(def ConnectorDefinition, cdc CDCPlan) []string {
	if def.ImplementationStatus == ImplementationEnabled {
		caps := def.RuntimeCapabilities
		ops := []string{}
		if caps.Check {
			ops = append(ops, "check")
		}
		if caps.Catalog {
			ops = append(ops, "catalog")
		}
		if caps.Read {
			ops = append(ops, "read")
		}
		if caps.Write {
			ops = append(ops, "write")
		}
		return ops
	}
	if def.Type == ConnectorTypeDestination {
		return []string{"check", "catalog", "write_append", "write_overwrite", "write_dedup"}
	}
	ops := []string{"check", "catalog", "read_snapshot"}
	if def.SupportsIncremental {
		ops = append(ops, "read_incremental")
	}
	if cdc.Supported {
		ops = append(ops, "read_cdc")
	}
	return ops
}

func reverseETLOperationsForDefinition(def ConnectorDefinition) []string {
	if def.ImplementationStatus == ImplementationEnabled && def.RuntimeCapabilities.ReverseETL {
		return []string{"validate_write", "preview", "write"}
	}
	return nil
}

func conformanceForDefinition(def ConnectorDefinition, family PortFamily, cdc CDCPlan) []string {
	base := []string{"spec", "check", "catalog", "docs_skill", "secret_redaction"}
	if def.Type == ConnectorTypeSource {
		base = append(base, "read_fixture", "state_checkpoint")
	}
	if def.Type == ConnectorTypeDestination {
		base = append(base, "write_fixture", "idempotency", "approval_policy")
	}
	switch family {
	case PortFamilyDeclarativeHTTPSource:
		base = append(base, "pagination", "authenticator", "rate_limit_retry", "schema_mapping")
	case PortFamilyDatabaseCDCSource:
		base = append(base, "snapshot_consistency", "cdc_checkpoint", "delete_semantics", "ordering")
	case PortFamilyDatabaseSource:
		base = append(base, "type_mapping", "cursor_incremental", "query_safety")
	case PortFamilyFileObjectSource:
		base = append(base, "format_detection", "bounded_streaming", "path_safety")
	case PortFamilyDestinationWriter:
		base = append(base, "batch_write", "dedup_write", "overwrite_write")
	case PortFamilyNativeSaaS:
		base = append(base, "pagination", "rate_limit_retry", "write_validation")
	}
	if cdc.Supported {
		base = append(base, "cdc_setup_validation")
	}
	return dedupeStrings(base)
}

func implementationNotesForDefinition(def ConnectorDefinition, family PortFamily, cdc CDCPlan) []string {
	if def.ImplementationStatus == ImplementationEnabled {
		return []string{"Native Go implementation is enabled in the current binary."}
	}
	notes := []string{"Catalog metadata is available; runtime remains disabled until the native Go port passes conformance tests."}
	switch family {
	case PortFamilyDeclarativeHTTPSource:
		notes = append(notes, "Use the declarative HTTP runtime: requester, authenticator, paginator, retriever, stream slicer, cursor/state, rate limiter, and schema mapper.")
	case PortFamilyDatabaseCDCSource:
		notes = append(notes, "Implement snapshot plus CDC with durable checkpoints before enabling incremental delete propagation.")
	case PortFamilyDatabaseSource:
		notes = append(notes, "Implement database snapshot and cursor incremental extraction before adding connector-specific CDC.")
	case PortFamilyFileObjectSource:
		notes = append(notes, "Implement bounded streaming readers for local files, object stores, compression, and CSV/JSON/Parquet formats.")
	case PortFamilyDestinationWriter:
		notes = append(notes, "Implement append, overwrite, dedup/upsert, error receipts, and idempotent retry behavior before enabling ETL writes.")
	case PortFamilyCustomGoPort:
		notes = append(notes, "Implement a custom native Go connector because this catalog entry does not fit a shared runtime family yet.")
	}
	if cdc.Supported {
		notes = append(notes, "CDC support requires setup validation and cleanup guidance for replication resources.")
	}
	return notes
}

func writeManualList(b *strings.Builder, values []string, label string) {
	if len(values) == 0 {
		b.WriteString("  " + label + ": none\n")
		return
	}
	b.WriteString("  " + label + ":\n")
	for _, value := range values {
		b.WriteString("    - " + value + "\n")
	}
}

func dedupeStrings(values []string) []string {
	seen := map[string]bool{}
	out := make([]string, 0, len(values))
	for _, value := range values {
		if value == "" || seen[value] {
			continue
		}
		seen[value] = true
		out = append(out, value)
	}
	sort.Strings(out)
	return out
}

package asana

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"slices"
	"strings"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/engine"
)

// asanaLaneArtifactMapping is the connector-local bridge from one immutable
// source-lane cell to existing definition artifacts. It is deliberately a
// proof shape rather than a second runtime schema: the real artifacts remain
// streams.json, writes.json, api_surface.json, cli_surface.json, and the
// closed event transport definition.
type asanaLaneArtifactMapping struct {
	APISurface *struct {
		Method string `json:"method"`
		Path   string `json:"path"`
	} `json:"api_surface"`
	Artifacts            []string        `json:"artifacts"`
	Artifact             string          `json:"artifact"`
	DescriptorPagination json.RawMessage `json:"descriptor_pagination"`
	DirectRead           string          `json:"direct_read"`
	Stream               string          `json:"stream"`
	WriteAction          string          `json:"write_action"`
	WriteActions         []string        `json:"write_actions"`
	RequestMediaType     string          `json:"request_media_type"`
	BinaryFields         []string        `json:"binary_fields"`
	ProviderLimitBytes   int64           `json:"provider_limit_bytes"`
	Role                 string          `json:"role"`
	EligibleStream       string          `json:"eligible_stream"`
	Executor             struct {
		Family string `json:"family"`
		ID     string `json:"id"`
	} `json:"executor"`
}

type asanaArtifactProjectionGapLedger struct {
	SchemaVersion int    `json:"schema_version"`
	Connector     string `json:"connector"`
	Purpose       string `json:"purpose"`
	SourceLock    struct {
		Path      string `json:"path"`
		SourceURL string `json:"source_url"`
		SHA256    string `json:"sha256"`
		Bytes     int    `json:"bytes"`
	} `json:"source_lock"`
	SourceMatrix struct {
		Path                 string   `json:"path"`
		SourceOperationCount int      `json:"source_operation_count"`
		Lanes                []string `json:"lanes"`
	} `json:"source_matrix"`
	AtlasLookups []struct {
		ID             string `json:"id"`
		Classification string `json:"classification"`
		Owner          string `json:"owner"`
	} `json:"atlas_lookups"`
	NonBindingFiles        json.RawMessage   `json:"non_binding_files"`
	RuntimeDefinitionFiles json.RawMessage   `json:"runtime_definition_files"`
	LaneMapping            json.RawMessage   `json:"lane_mapping"`
	Foundations            []json.RawMessage `json:"foundations"`
	Gaps                   []struct {
		ID          string                        `json:"id"`
		Lane        string                        `json:"lane"`
		Disposition string                        `json:"disposition"`
		TypedReason string                        `json:"typed_reason"`
		AtlasLookup string                        `json:"atlas_lookup"`
		Reason      string                        `json:"reason"`
		SourceIDs   []string                      `json:"source_ids"`
		LaneCells   []asanaMatrixLaneCellEvidence `json:"lane_cells"`
	} `json:"gaps"`
}

type asanaMatrixLaneCellEvidence struct {
	SourceID    string `json:"source_id"`
	Lane        string `json:"lane"`
	Disposition string `json:"disposition"`
}

type asanaArtifactProjectionFoundation struct {
	ID                     string          `json:"id"`
	State                  string          `json:"state"`
	AffectedLane           string          `json:"affected_lane"`
	SourceEvidenceContract string          `json:"source_evidence_contract"`
	Reason                 string          `json:"reason"`
	Source                 json.RawMessage `json:"source"`
	DoesNotBlock           []string        `json:"does_not_block"`
	MatrixProjection       struct {
		SourceIDs   []string                      `json:"source_ids"`
		LaneCells   []asanaMatrixLaneCellEvidence `json:"lane_cells"`
		TypedReason string                        `json:"typed_reason"`
		AtlasLookup string                        `json:"atlas_lookup"`
	} `json:"matrix_projection"`
}

type asanaLaneArtifactHiddenFS struct {
	fs.FS
	hidden string
}

func (f asanaLaneArtifactHiddenFS) Open(name string) (fs.File, error) {
	if name == f.hidden {
		return nil, fs.ErrNotExist
	}
	return f.FS.Open(name)
}

func TestAsanaSourceLaneArtifactsProjectTheTrackAMatrix(t *testing.T) {
	matrix := loadAsanaSourceLaneMatrix(t)
	lock := loadAsanaSourceLaneLock(t)
	descriptor := loadAsanaSourceLaneDescriptor(t)
	bundle := loadBundle(t)
	gaps := loadAsanaArtifactProjectionGapLedger(t)
	definitions := os.DirFS(".")

	if err := validateAsanaSourceLaneArtifactProjection(definitions, matrix, lock, descriptor, bundle, gaps); err != nil {
		t.Fatalf("validate Asana source-lane artifact projection: %v", err)
	}

	// Deliberate red counterpart: an enabled-contract artifact is not merely
	// documentation. The checked-in file is never moved; a hidden filesystem
	// view must still reject the missing sync artifact before green evidence is
	// accepted.
	if err := validateAsanaSourceLaneArtifactProjection(asanaLaneArtifactHiddenFS{FS: definitions, hidden: "sync_transport.json"}, matrix, lock, descriptor, bundle, gaps); err == nil || !strings.Contains(err.Error(), "sync_transport artifact \"sync_transport.json\" is unavailable") {
		t.Fatalf("missing sync artifact validation error = %v, want unavailable sync_transport artifact", err)
	}

	// Deliberate red counterpart: a matrix row cannot survive on aggregate
	// counts alone when its exact definition backlink disappears.
	broken := asanaMatrixWithoutDirectReadBacklink(t, matrix)
	if err := validateAsanaSourceLaneArtifactProjection(definitions, broken, lock, descriptor, bundle, gaps); err == nil || !strings.Contains(err.Error(), "direct_read backlink") {
		t.Fatalf("missing direct-read backlink validation error = %v, want direct_read backlink", err)
	}

	// Deliberate red counterpart: descriptor pagination by itself must never
	// turn a mapped-unproven source cell into an ETL stream.
	promoted := asanaMatrixWithPromotedMappedUnprovenETL(t, matrix)
	if err := validateAsanaSourceLaneArtifactProjection(definitions, promoted, lock, descriptor, bundle, gaps); err == nil || !strings.Contains(err.Error(), "mapped-unproven ETL source") {
		t.Fatalf("promoted mapped-unproven ETL validation error = %v, want descriptor-only ETL rejection", err)
	}
}

func loadAsanaArtifactProjectionGapLedger(t *testing.T) asanaArtifactProjectionGapLedger {
	t.Helper()
	raw, err := os.ReadFile("missing-foundation.json")
	if err != nil {
		t.Fatalf("read Asana gap ledger: %v", err)
	}
	var ledger asanaArtifactProjectionGapLedger
	decoder := json.NewDecoder(strings.NewReader(string(raw)))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&ledger); err != nil {
		t.Fatalf("decode Asana gap ledger: %v", err)
	}
	return ledger
}

func asanaMatrixWithoutDirectReadBacklink(t *testing.T, matrix asanaSourceLaneMatrix) asanaSourceLaneMatrix {
	t.Helper()
	broken := matrix
	broken.SourceOperations = append([]asanaSourceLaneMatrixRow(nil), matrix.SourceOperations...)
	for index, row := range broken.SourceOperations {
		cell, ok := row.Lanes["direct_read"]
		if !ok || cell.Applicability != "applicable" {
			continue
		}
		row.Lanes = mapsClone(row.Lanes)
		cell.Mapping = nil
		row.Lanes["direct_read"] = cell
		broken.SourceOperations[index] = row
		return broken
	}
	t.Fatal("matrix has no applicable direct-read cell to mutate")
	return asanaSourceLaneMatrix{}
}

func asanaMatrixWithPromotedMappedUnprovenETL(t *testing.T, matrix asanaSourceLaneMatrix) asanaSourceLaneMatrix {
	t.Helper()
	broken := matrix
	broken.SourceOperations = append([]asanaSourceLaneMatrixRow(nil), matrix.SourceOperations...)
	for index, row := range broken.SourceOperations {
		cell, ok := row.Lanes["etl"]
		if !ok || cell.Applicability != "applicable" || cell.Disposition != "mapped_unproven" {
			continue
		}
		var mapping map[string]any
		if err := json.Unmarshal(cell.Mapping, &mapping); err != nil {
			t.Fatalf("decode mapped-unproven ETL mapping: %v", err)
		}
		mapping["stream"] = "tasks"
		raw, err := json.Marshal(mapping)
		if err != nil {
			t.Fatalf("encode promoted mapped-unproven ETL mapping: %v", err)
		}
		row.Lanes = mapsClone(row.Lanes)
		cell.Mapping = raw
		row.Lanes["etl"] = cell
		broken.SourceOperations[index] = row
		return broken
	}
	t.Fatal("matrix has no mapped-unproven ETL cell to mutate")
	return asanaSourceLaneMatrix{}
}

func validateAsanaSourceLaneArtifactProjection(definitions fs.FS, matrix asanaSourceLaneMatrix, lock asanaSourceLaneLock, descriptor asanaSourceLaneDescriptor, bundle engine.Bundle, gaps asanaArtifactProjectionGapLedger) error {
	if bundle.EnabledContract == nil {
		return fmt.Errorf("Asana bundle has no enabled connector contract")
	}
	if bundle.Surface == nil {
		return fmt.Errorf("Asana bundle has no API surface")
	}
	if err := validateAsanaEnabledLaneArtifactInventory(definitions, *bundle.EnabledContract); err != nil {
		return err
	}

	streams := make(map[string]engine.StreamSpec, len(bundle.Streams))
	for _, stream := range bundle.Streams {
		streams[stream.Name] = stream
	}
	actions := make(map[string]engine.WriteAction, len(bundle.Writes))
	for _, action := range bundle.Writes {
		actions[action.Name] = action
	}
	endpoints := make(map[string]engine.SurfaceEndpoint, len(bundle.Surface.Endpoints))
	for _, endpoint := range bundle.Surface.Endpoints {
		key := asanaLaneArtifactEndpointKey(endpoint.Method, endpoint.Path)
		if _, duplicate := endpoints[key]; duplicate {
			return fmt.Errorf("api_surface repeats endpoint %s", key)
		}
		endpoints[key] = endpoint
	}
	commands := make(map[string]engine.CLICommand, len(bundle.CLISurface.Commands))
	for _, command := range bundle.CLISurface.Commands {
		if _, duplicate := commands[command.Path]; duplicate {
			return fmt.Errorf("cli_surface repeats command %q", command.Path)
		}
		commands[command.Path] = command
	}
	operations := make(map[string]struct{}, len(bundle.Operations))
	for _, operation := range bundle.Operations {
		if operation.SourceOperation != nil && operation.SourceOperation.ID != "" {
			operations[operation.SourceOperation.ID] = struct{}{}
		}
	}

	counts := map[string]int{}
	deleteCells := 0
	etlMappedUnproven := make([]string, 0, 52)
	syncNotApplicableForStreams := make([]string, 0, 11)
	for _, row := range matrix.SourceOperations {
		if err := validateAsanaDirectReadArtifactBacklink(row, endpoints, commands, operations); err != nil {
			return err
		}
		if err := validateAsanaDirectWriteArtifactBacklink(row, endpoints, commands, actions); err != nil {
			return err
		}
		if err := validateAsanaBinaryArtifactBacklink(row, commands, actions); err != nil {
			return err
		}
		if err := validateAsanaETLArtifactBacklink(definitions, row, streams, endpoints, &etlMappedUnproven); err != nil {
			return err
		}
		if err := validateAsanaReverseETLArtifactBacklink(row, endpoints, actions); err != nil {
			return err
		}
		if err := validateAsanaSyncArtifactBacklink(row, bundle); err != nil {
			return err
		}

		for lane, cell := range row.Lanes {
			if cell.Applicability == "applicable" {
				counts[lane]++
			}
		}
		if row.SourceFacts.Method == "DELETE" {
			deleteCells++
			if row.Lanes["direct_write"].Disposition != "implemented" || row.Lanes["reverse_etl"].Disposition != "implemented" {
				return fmt.Errorf("DELETE source %q lost a direct-write or reverse-ETL disposition", row.SourceID)
			}
		}
		if row.Lanes["etl"].Disposition == "implemented" && row.SourceID != "asana.rest.getTasks" && row.Lanes["sync_transport"].Disposition == "not_applicable" {
			syncNotApplicableForStreams = append(syncNotApplicableForStreams, row.SourceID)
		}
	}
	if !slices.Equal([]int{counts["direct_read"], counts["direct_write"], counts["binary_download"], counts["binary_upload"], counts["etl"], counts["reverse_etl"], counts["sync_transport"]}, []int{119, 130, 0, 1, 64, 130, 3}) {
		return fmt.Errorf("applicable lane counts = direct_read=%d direct_write=%d binary_download=%d binary_upload=%d etl=%d reverse_etl=%d sync_transport=%d, want 119/130/0/1/64/130/3", counts["direct_read"], counts["direct_write"], counts["binary_download"], counts["binary_upload"], counts["etl"], counts["reverse_etl"], counts["sync_transport"])
	}
	if deleteCells != 23 {
		return fmt.Errorf("DELETE source cells = %d, want 23", deleteCells)
	}
	if err := validateAsanaSourceLaneMatrix(matrix, lock, descriptor); err != nil {
		return err
	}
	if err := validateAsanaGapLedger(gaps, matrix, etlMappedUnproven, syncNotApplicableForStreams); err != nil {
		return err
	}
	return nil
}

func validateAsanaEnabledLaneArtifactInventory(definitions fs.FS, contract connectors.EnabledConnectorContract) error {
	expected := map[string][]string{
		"direct_read": {
			"streams.json", "operations.json", "api_surface.json", "cli_surface.json",
			"sources/asana-operation-source-lock.json", asanaSourceLaneMatrixPath,
		},
		"direct_write": {
			"writes.json", "cli_surface.json", "api_surface.json",
			"sources/asana-operation-source-lock.json", asanaSourceLaneMatrixPath,
		},
		"binary_download": {
			"cli_surface.json", "sources/asana-operation-source-lock.json", asanaSourceLaneMatrixPath,
		},
		"binary_upload": {
			"writes.json", "cli_surface.json", "sources/asana-operation-source-lock.json", asanaSourceLaneMatrixPath,
		},
		"etl": {
			"streams.json",
			"schemas/custom_fields.json", "schemas/project_statuses.json", "schemas/projects.json", "schemas/sections.json",
			"schemas/stories.json", "schemas/tags.json", "schemas/tasks.json", "schemas/team_memberships.json",
			"schemas/teams.json", "schemas/users.json", "schemas/workspace_memberships.json", "schemas/workspaces.json",
			"api_surface.json", "cli_surface.json", "sources/asana-operation-descriptor.json",
			"sources/asana-operation-source-lock.json", asanaSourceLaneMatrixPath,
		},
		"reverse_etl": {
			"writes.json", "cli_surface.json", "api_surface.json",
			"sources/asana-operation-source-lock.json", asanaSourceLaneMatrixPath,
		},
		"sync_transport": {
			"sync_transport.json", "event_source_contract.json", "sources/asana-operation-source-lock.json", asanaSourceLaneMatrixPath,
		},
	}
	wantState := map[string]string{
		"direct_read": "implemented", "direct_write": "implemented", "binary_download": "unsupported_with_provider_evidence",
		"binary_upload": "implemented", "etl": "implemented", "reverse_etl": "implemented", "sync_transport": "implemented",
	}
	wantCoverage := map[string]string{
		"direct_read": "complete", "direct_write": "complete", "binary_download": "not_applicable", "binary_upload": "complete",
		"etl": "partial", "reverse_etl": "complete", "sync_transport": "complete",
	}
	wantCounts := map[string][3]int{
		"direct_read": {119, 119, 0}, "direct_write": {130, 130, 0}, "binary_download": {0, 0, 0},
		"binary_upload": {1, 1, 0}, "etl": {64, 12, 52}, "reverse_etl": {130, 130, 0}, "sync_transport": {3, 3, 0},
	}
	if len(contract.Lanes) != len(expected) {
		return fmt.Errorf("enabled contract lane count = %d, want %d", len(contract.Lanes), len(expected))
	}
	seen := map[string]bool{}
	for _, lane := range contract.Lanes {
		artifacts, ok := expected[lane.Name]
		if !ok {
			return fmt.Errorf("enabled contract has unexpected lane %q", lane.Name)
		}
		seen[lane.Name] = true
		if lane.State != wantState[lane.Name] || lane.Source.Coverage != wantCoverage[lane.Name] {
			return fmt.Errorf("%s contract state/coverage = %q/%q, want %q/%q", lane.Name, lane.State, lane.Source.Coverage, wantState[lane.Name], wantCoverage[lane.Name])
		}
		want := wantCounts[lane.Name]
		if got := [3]int{lane.Source.Expected, lane.Source.Implemented, lane.Source.UnmappedMapping}; got != want {
			return fmt.Errorf("%s contract source counts = %v, want %v", lane.Name, got, want)
		}
		if !slices.Equal(lane.Artifacts, artifacts) {
			return fmt.Errorf("%s contract artifacts = %v, want %v", lane.Name, lane.Artifacts, artifacts)
		}
		for _, artifact := range lane.Artifacts {
			if _, err := fs.Stat(definitions, artifact); err != nil {
				return fmt.Errorf("%s artifact %q is unavailable: %w", lane.Name, artifact, err)
			}
		}
	}
	for lane := range expected {
		if !seen[lane] {
			return fmt.Errorf("enabled contract omits %s artifact inventory", lane)
		}
	}
	return nil
}

func validateAsanaDirectReadArtifactBacklink(row asanaSourceLaneMatrixRow, endpoints map[string]engine.SurfaceEndpoint, commands map[string]engine.CLICommand, operations map[string]struct{}) error {
	cell := row.Lanes["direct_read"]
	if cell.Applicability != "applicable" {
		return nil
	}
	mapping, err := asanaLaneArtifactMappingFor(row, "direct_read", cell)
	if err != nil {
		return err
	}
	if mapping.APISurface == nil || mapping.DirectRead == "" || !asanaHasArtifacts(mapping.Artifacts, "api_surface.json", "operations.json", "cli_surface.json") {
		return fmt.Errorf("direct_read backlink for %q is incomplete", row.SourceID)
	}
	endpoint, ok := endpoints[asanaLaneArtifactEndpointKey(mapping.APISurface.Method, mapping.APISurface.Path)]
	if !ok || endpoint.CoveredBy == nil || endpoint.CoveredBy.DirectRead != mapping.DirectRead || mapping.APISurface.Method != row.SourceFacts.Method || mapping.APISurface.Path != row.SourceFacts.Path {
		return fmt.Errorf("direct_read backlink for %q does not resolve exact API surface", row.SourceID)
	}
	command, ok := commands[mapping.DirectRead]
	if !ok || command.Intent != "direct_read" || command.Availability != "implemented" || command.SourceOperation != row.SourceID {
		return fmt.Errorf("direct_read backlink for %q does not resolve an implemented source-bound command", row.SourceID)
	}
	if _, ok := operations[row.SourceID]; !ok {
		return fmt.Errorf("direct_read backlink for %q has no operation declaration", row.SourceID)
	}
	return nil
}

func validateAsanaDirectWriteArtifactBacklink(row asanaSourceLaneMatrixRow, endpoints map[string]engine.SurfaceEndpoint, commands map[string]engine.CLICommand, actions map[string]engine.WriteAction) error {
	cell := row.Lanes["direct_write"]
	if cell.Applicability != "applicable" {
		return nil
	}
	mapping, err := asanaLaneArtifactMappingFor(row, "direct_write", cell)
	if err != nil {
		return err
	}
	if mapping.APISurface == nil || len(mapping.WriteActions) == 0 || !asanaHasArtifacts(mapping.Artifacts, "api_surface.json", "writes.json", "cli_surface.json") {
		return fmt.Errorf("direct_write backlink for %q is incomplete", row.SourceID)
	}
	endpoint, ok := endpoints[asanaLaneArtifactEndpointKey(mapping.APISurface.Method, mapping.APISurface.Path)]
	if !ok || endpoint.CoveredBy == nil || mapping.APISurface.Method != row.SourceFacts.Method || mapping.APISurface.Path != row.SourceFacts.Path || !asanaSameStringSet(asanaSurfaceWriteActions(endpoint), mapping.WriteActions) {
		return fmt.Errorf("direct_write backlink for %q does not resolve exact API/write actions", row.SourceID)
	}
	for _, actionName := range mapping.WriteActions {
		if _, ok := actions[actionName]; !ok {
			return fmt.Errorf("direct_write backlink for %q names absent action %q", row.SourceID, actionName)
		}
		if !asanaHasImplementedDirectWriteCommand(commands, actionName, mapping.APISurface.Method, mapping.APISurface.Path) {
			return fmt.Errorf("direct_write backlink for %q action %q has no implemented CLI command", row.SourceID, actionName)
		}
	}
	return nil
}

func validateAsanaBinaryArtifactBacklink(row asanaSourceLaneMatrixRow, commands map[string]engine.CLICommand, actions map[string]engine.WriteAction) error {
	download := row.Lanes["binary_download"]
	if download.Applicability != "not_applicable" || download.Disposition != "not_applicable" {
		return fmt.Errorf("binary_download source %q is promoted beyond provider evidence", row.SourceID)
	}
	upload := row.Lanes["binary_upload"]
	if upload.Applicability != "applicable" {
		return nil
	}
	mapping, err := asanaLaneArtifactMappingFor(row, "binary_upload", upload)
	if err != nil {
		return err
	}
	if row.SourceID != "asana.rest.createAttachmentForObject" || mapping.WriteAction != "upload_attachment_file" || mapping.Artifact != "writes.json" || mapping.RequestMediaType != "multipart/form-data" || !slices.Equal(mapping.BinaryFields, []string{"file"}) || mapping.ProviderLimitBytes != 104857600 {
		return fmt.Errorf("binary_upload backlink for %q is not the one source-cited attachment action", row.SourceID)
	}
	action, ok := actions[mapping.WriteAction]
	if !ok {
		return fmt.Errorf("binary_upload action %q is absent", mapping.WriteAction)
	}
	if action.Method != row.SourceFacts.Method || action.Path != row.SourceFacts.Path || action.BodyType != "multipart" || action.Multipart == nil || action.Multipart.MaxBytes != mapping.ProviderLimitBytes || !asanaHasRequiredMultipartFilePart(action.Multipart, "file", mapping.ProviderLimitBytes) {
		return fmt.Errorf("binary_upload action %q does not retain the source-cited multipart file contract", mapping.WriteAction)
	}
	for _, command := range commands {
		if command.Intent == "binary_upload" && command.Availability == "implemented" && command.Write == mapping.WriteAction {
			return nil
		}
	}
	return fmt.Errorf("binary_upload backlink for %q has no implemented CLI alias", row.SourceID)
}

func validateAsanaETLArtifactBacklink(definitions fs.FS, row asanaSourceLaneMatrixRow, streams map[string]engine.StreamSpec, endpoints map[string]engine.SurfaceEndpoint, mappedUnproven *[]string) error {
	cell := row.Lanes["etl"]
	if cell.Applicability != "applicable" {
		return nil
	}
	mapping, err := asanaLaneArtifactMappingFor(row, "etl", cell)
	if err != nil {
		return err
	}
	switch cell.Disposition {
	case "implemented":
		stream, ok := streams[mapping.Stream]
		if !ok || mapping.APISurface == nil || stream.SchemaRef == "" || !asanaHasArtifacts(mapping.Artifacts, "streams.json", stream.SchemaRef, "api_surface.json") {
			return fmt.Errorf("ETL backlink for %q is incomplete", row.SourceID)
		}
		if _, err := fs.Stat(definitions, stream.SchemaRef); err != nil {
			return fmt.Errorf("ETL schema backlink for %q %q is unavailable: %w", row.SourceID, stream.SchemaRef, err)
		}
		endpoint, ok := endpoints[asanaLaneArtifactEndpointKey(mapping.APISurface.Method, mapping.APISurface.Path)]
		if !ok || endpoint.CoveredBy == nil || endpoint.CoveredBy.Stream != mapping.Stream || mapping.APISurface.Method != row.SourceFacts.Method || mapping.APISurface.Path != row.SourceFacts.Path {
			return fmt.Errorf("ETL backlink for %q does not resolve stream/API surface", row.SourceID)
		}
	case "mapped_unproven":
		if mapping.Stream != "" || mapping.APISurface != nil || len(mapping.DescriptorPagination) == 0 || !slices.Equal(mapping.Artifacts, []string{"sources/asana-operation-descriptor.json"}) {
			return fmt.Errorf("mapped-unproven ETL source %q is promoted beyond descriptor-only evidence", row.SourceID)
		}
		*mappedUnproven = append(*mappedUnproven, row.SourceID)
	default:
		return fmt.Errorf("applicable ETL source %q has unexpected disposition %q", row.SourceID, cell.Disposition)
	}
	return nil
}

func validateAsanaReverseETLArtifactBacklink(row asanaSourceLaneMatrixRow, endpoints map[string]engine.SurfaceEndpoint, actions map[string]engine.WriteAction) error {
	cell := row.Lanes["reverse_etl"]
	if cell.Applicability != "applicable" {
		return nil
	}
	mapping, err := asanaLaneArtifactMappingFor(row, "reverse_etl", cell)
	if err != nil {
		return err
	}
	if len(mapping.WriteActions) == 0 || !asanaHasArtifacts(mapping.Artifacts, "writes.json", "api_surface.json") {
		return fmt.Errorf("reverse_etl backlink for %q is incomplete", row.SourceID)
	}
	direct, err := asanaLaneArtifactMappingFor(row, "direct_write", row.Lanes["direct_write"])
	if err != nil {
		return err
	}
	endpoint, ok := endpoints[asanaLaneArtifactEndpointKey(direct.APISurface.Method, direct.APISurface.Path)]
	if !ok || !asanaSameStringSet(mapping.WriteActions, direct.WriteActions) || !asanaSameStringSet(asanaSurfaceWriteActions(endpoint), mapping.WriteActions) {
		return fmt.Errorf("reverse_etl backlink for %q does not retain exact direct-write actions", row.SourceID)
	}
	for _, actionName := range mapping.WriteActions {
		if _, ok := actions[actionName]; !ok {
			return fmt.Errorf("reverse_etl backlink for %q names absent action %q", row.SourceID, actionName)
		}
	}
	return nil
}

func validateAsanaSyncArtifactBacklink(row asanaSourceLaneMatrixRow, bundle engine.Bundle) error {
	cell := row.Lanes["sync_transport"]
	if cell.Applicability != "applicable" {
		return nil
	}
	if bundle.SyncTransport == nil || bundle.SyncTransport.Source == nil {
		return fmt.Errorf("sync_transport source %q has no source transport artifact", row.SourceID)
	}
	mapping, err := asanaLaneArtifactMappingFor(row, "sync_transport", cell)
	if err != nil {
		return err
	}
	wantRoles := map[string]string{
		"asana.rest.getEvents": "event_window",
		"asana.rest.getTask":   "hydration",
		"asana.rest.getTasks":  "snapshot",
	}
	if !asanaHasArtifacts(mapping.Artifacts, "event_source_contract.json", "sync_transport.json") || mapping.Role != wantRoles[row.SourceID] || mapping.EligibleStream != "tasks" || mapping.Executor.Family != "declarative_api" || mapping.Executor.ID != "asana_event_token_source" {
		return fmt.Errorf("sync_transport backlink for %q is incomplete", row.SourceID)
	}
	if string(bundle.SyncTransport.Source.Executor.Family) != mapping.Executor.Family || bundle.SyncTransport.Source.Executor.ID != mapping.Executor.ID || !slices.Equal(bundle.SyncTransport.Source.EligibleStreams, []string{"tasks"}) {
		return fmt.Errorf("sync_transport backlink for %q does not resolve the closed task executor", row.SourceID)
	}
	return nil
}

func validateAsanaGapLedger(gaps asanaArtifactProjectionGapLedger, matrix asanaSourceLaneMatrix, etlMappedUnproven, syncNotApplicableForStreams []string) error {
	if gaps.SchemaVersion != 2 || gaps.Connector != "asana" || strings.TrimSpace(gaps.Purpose) == "" || gaps.SourceLock.Path != "sources/asana-operation-source-lock.json" || strings.TrimSpace(gaps.SourceLock.SourceURL) == "" || gaps.SourceLock.SHA256 != asanaSourceOpenAPISHA256 || gaps.SourceLock.Bytes != 3066750 || gaps.SourceMatrix.Path != asanaSourceLaneMatrixPath || gaps.SourceMatrix.SourceOperationCount != len(matrix.SourceOperations) || !slices.Equal(gaps.SourceMatrix.Lanes, matrix.Lanes) {
		return fmt.Errorf("Asana gap ledger is not bound to the Track A source matrix")
	}
	if len(gaps.NonBindingFiles) == 0 || len(gaps.RuntimeDefinitionFiles) == 0 || len(gaps.LaneMapping) == 0 {
		return fmt.Errorf("Asana gap ledger did not carry forward its existing artifact-role inventory")
	}
	lookups := map[string]struct {
		classification string
		owner          string
	}{}
	for _, lookup := range gaps.AtlasLookups {
		if _, duplicate := lookups[lookup.ID]; duplicate {
			return fmt.Errorf("Asana gap ledger has duplicate Atlas lookup %q", lookup.ID)
		}
		lookups[lookup.ID] = struct {
			classification string
			owner          string
		}{classification: lookup.Classification, owner: lookup.Owner}
	}
	for _, lookupID := range []string{
		"source.retention-import.v1",
		"source.projection-admission.v1",
		"runtime.direct-execution.v1",
		"warehouse.stage-etl.v1",
		"warehouse.reverse-etl.v1",
		"transport.sync-contract.v1",
		"asana.event-token-source.v1",
	} {
		lookup, ok := lookups[lookupID]
		if !ok || lookup.classification != "reuse" || strings.TrimSpace(lookup.owner) == "" {
			return fmt.Errorf("Asana gap ledger does not retain reusable Atlas lookup %q", lookupID)
		}
	}
	if err := validateAsanaCarriedFoundationProjections(gaps.Foundations, matrix, syncNotApplicableForStreams, lookups); err != nil {
		return err
	}
	if len(gaps.Gaps) != 1 {
		return fmt.Errorf("Asana newly added gap entries = %d, want exactly one matrix-bound ETL limitation", len(gaps.Gaps))
	}
	want := map[string]struct {
		lane        string
		disposition string
		typedReason string
		atlas       string
		sourceIDs   []string
	}{
		"asana_etl_mapped_unproven_scope_fanout_projection": {
			lane: "etl", disposition: "mapped_unproven", typedReason: "missing_source_backed_stream_scope_or_fanout", atlas: "source.projection-admission.v1", sourceIDs: etlMappedUnproven,
		},
	}
	for _, gap := range gaps.Gaps {
		expectation, ok := want[gap.ID]
		if !ok {
			return fmt.Errorf("Asana gap ledger has unexpected entry %q", gap.ID)
		}
		if gap.Lane != expectation.lane || gap.Disposition != expectation.disposition || gap.TypedReason != expectation.typedReason || gap.AtlasLookup != expectation.atlas || strings.TrimSpace(gap.Reason) == "" || len(gap.SourceIDs) == 0 || len(gap.LaneCells) == 0 {
			return fmt.Errorf("Asana gap ledger entry %q lacks typed exact source/lane evidence", gap.ID)
		}
		if lookup, ok := lookups[gap.AtlasLookup]; !ok || lookup.classification != "reuse" || strings.TrimSpace(lookup.owner) == "" {
			return fmt.Errorf("Asana gap ledger entry %q has no reusable Atlas lookup", gap.ID)
		}
		if !asanaSameStringSet(gap.SourceIDs, expectation.sourceIDs) {
			return fmt.Errorf("Asana gap ledger entry %q source IDs = %v, want exact matrix cells %v", gap.ID, gap.SourceIDs, expectation.sourceIDs)
		}
		cells := make([]string, 0, len(gap.LaneCells))
		for _, cell := range gap.LaneCells {
			if cell.Lane != gap.Lane || cell.Disposition != gap.Disposition {
				return fmt.Errorf("Asana gap ledger entry %q has mismatched lane cell %+v", gap.ID, cell)
			}
			cells = append(cells, cell.SourceID)
		}
		if !asanaSameStringSet(cells, gap.SourceIDs) {
			return fmt.Errorf("Asana gap ledger entry %q lane cells do not exactly match source IDs", gap.ID)
		}
		delete(want, gap.ID)
	}
	if len(want) != 0 {
		return fmt.Errorf("Asana gap ledger is missing %d exact matrix limitations", len(want))
	}
	return nil
}

func validateAsanaCarriedFoundationProjections(rawFoundations []json.RawMessage, matrix asanaSourceLaneMatrix, syncNotApplicableForStreams []string, lookups map[string]struct {
	classification string
	owner          string
}) error {
	if len(rawFoundations) != 7 {
		return fmt.Errorf("Asana carried-forward foundation entries = %d, want 7", len(rawFoundations))
	}
	etlImplemented := asanaMatrixSourceIDs(matrix, "etl", "implemented")
	mutations := asanaMatrixSourceIDs(matrix, "direct_write", "implemented")
	allSourceIDs := make([]string, 0, len(matrix.SourceOperations))
	for _, row := range matrix.SourceOperations {
		allSourceIDs = append(allSourceIDs, row.SourceID)
	}
	expected := map[string]struct {
		state        string
		affectedLane string
		typedReason  string
		atlas        string
		sourceIDs    []string
		laneCells    []asanaMatrixLaneCellEvidence
	}{
		"asana_stream_direct_read_request_bound": {
			state: "resolved_runtime_mapping", affectedLane: "direct_read", typedReason: "interactive_stream_request_budget_bound", atlas: "runtime.direct-execution.v1", sourceIDs: etlImplemented, laneCells: asanaMatrixEvidenceCells(matrix, etlImplemented, []string{"direct_read"}, false),
		},
		"asana_incremental_event_scope_coverage": {
			state: "existing_mapping_backlog", affectedLane: "etl", typedReason: "source_event_scope_or_order_not_cited", atlas: "asana.event-token-source.v1", sourceIDs: syncNotApplicableForStreams, laneCells: append(asanaMatrixEvidenceCells(matrix, syncNotApplicableForStreams, []string{"etl"}, false), asanaMatrixEvidenceCells(matrix, syncNotApplicableForStreams, []string{"sync_transport"}, false)...),
		},
		"asana_attachment_direct_write_multipart": {
			state: "missing_runtime_contract", affectedLane: "direct_write", typedReason: "legacy_operation_multipart_variants_not_declared", atlas: "runtime.direct-execution.v1", sourceIDs: []string{"asana.rest.createAttachmentForObject"}, laneCells: asanaMatrixEvidenceCells(matrix, []string{"asana.rest.createAttachmentForObject"}, []string{"direct_write"}, false),
		},
		"asana_attachment_binary_upload_media_policy": {
			state: "resolved_runtime_mapping", affectedLane: "binary_upload", typedReason: "source_cited_provider_unrestricted_binary_media", atlas: "runtime.direct-execution.v1", sourceIDs: []string{"asana.rest.createAttachmentForObject"}, laneCells: asanaMatrixEvidenceCells(matrix, []string{"asana.rest.createAttachmentForObject"}, []string{"binary_upload"}, false),
		},
		"asana_action_scoped_request_schema_projection": {
			state: "resolved_runtime_mapping", affectedLane: "reverse_etl_and_direct_write", typedReason: "closed_action_request_schema_projection", atlas: "runtime.direct-execution.v1", sourceIDs: mutations, laneCells: asanaMatrixEvidenceCells(matrix, mutations, []string{"direct_write", "reverse_etl"}, false),
		},
		"legacy_asana_openapi_reference_sibling_rows": {
			state: "resolved_runtime_mapping_with_authoring_gap", affectedLane: "reverse_etl_and_direct_write", typedReason: "openapi_reference_sibling_authoring_diagnostic", atlas: "source.retention-import.v1", sourceIDs: []string{"asana.rest.createMembership"}, laneCells: asanaMatrixEvidenceCells(matrix, []string{"asana.rest.createMembership"}, []string{"direct_write", "reverse_etl"}, false),
		},
		"asana_enriched_source_lock_projection_parser": {
			state: "resolved_authoring_mapping", affectedLane: "all_source_mapped_lanes", typedReason: "strict_source_lock_projection_parser", atlas: "source.retention-import.v1", sourceIDs: allSourceIDs, laneCells: asanaMatrixEvidenceCells(matrix, allSourceIDs, matrix.Lanes, true),
		},
	}
	for _, rawFoundation := range rawFoundations {
		var foundation asanaArtifactProjectionFoundation
		decoder := json.NewDecoder(strings.NewReader(string(rawFoundation)))
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&foundation); err != nil {
			return fmt.Errorf("decode carried-forward Asana foundation: %w", err)
		}
		want, ok := expected[foundation.ID]
		if !ok {
			return fmt.Errorf("Asana ledger has unexpected carried-forward foundation %q", foundation.ID)
		}
		projection := foundation.MatrixProjection
		if foundation.State != want.state || foundation.AffectedLane != want.affectedLane || strings.TrimSpace(foundation.Reason) == "" || projection.TypedReason != want.typedReason || projection.AtlasLookup != want.atlas || len(projection.SourceIDs) == 0 || len(projection.LaneCells) == 0 {
			return fmt.Errorf("Asana carried-forward foundation %q lost its state or exact matrix projection", foundation.ID)
		}
		if lookup, ok := lookups[projection.AtlasLookup]; !ok || lookup.classification != "reuse" || strings.TrimSpace(lookup.owner) == "" {
			return fmt.Errorf("Asana carried-forward foundation %q has no reusable Atlas lookup", foundation.ID)
		}
		if !asanaSameStringSet(projection.SourceIDs, want.sourceIDs) || !asanaSameMatrixEvidenceCellSet(projection.LaneCells, want.laneCells) {
			return fmt.Errorf("Asana carried-forward foundation %q does not retain its exact source/lane projection", foundation.ID)
		}
		for _, cell := range projection.LaneCells {
			if !slices.Contains(projection.SourceIDs, cell.SourceID) || !asanaMatrixEvidenceCellMatches(matrix, cell) {
				return fmt.Errorf("Asana carried-forward foundation %q has invalid matrix cell %+v", foundation.ID, cell)
			}
		}
		delete(expected, foundation.ID)
	}
	if len(expected) != 0 {
		return fmt.Errorf("Asana ledger is missing %d carried-forward foundation projections", len(expected))
	}
	return nil
}

func asanaMatrixSourceIDs(matrix asanaSourceLaneMatrix, lane, disposition string) []string {
	ids := make([]string, 0)
	for _, row := range matrix.SourceOperations {
		if row.Lanes[lane].Disposition == disposition {
			ids = append(ids, row.SourceID)
		}
	}
	return ids
}

func asanaMatrixEvidenceCells(matrix asanaSourceLaneMatrix, sourceIDs, lanes []string, applicableOnly bool) []asanaMatrixLaneCellEvidence {
	wantedIDs := make(map[string]bool, len(sourceIDs))
	for _, sourceID := range sourceIDs {
		wantedIDs[sourceID] = true
	}
	wantedLanes := make(map[string]bool, len(lanes))
	for _, lane := range lanes {
		wantedLanes[lane] = true
	}
	cells := make([]asanaMatrixLaneCellEvidence, 0)
	for _, row := range matrix.SourceOperations {
		if !wantedIDs[row.SourceID] {
			continue
		}
		for _, lane := range matrix.Lanes {
			if !wantedLanes[lane] {
				continue
			}
			cell := row.Lanes[lane]
			if applicableOnly && cell.Applicability != "applicable" {
				continue
			}
			cells = append(cells, asanaMatrixLaneCellEvidence{SourceID: row.SourceID, Lane: lane, Disposition: cell.Disposition})
		}
	}
	return cells
}

func asanaMatrixEvidenceCellMatches(matrix asanaSourceLaneMatrix, evidence asanaMatrixLaneCellEvidence) bool {
	for _, row := range matrix.SourceOperations {
		if row.SourceID != evidence.SourceID {
			continue
		}
		cell, ok := row.Lanes[evidence.Lane]
		return ok && cell.Disposition == evidence.Disposition
	}
	return false
}

func asanaSameMatrixEvidenceCellSet(left, right []asanaMatrixLaneCellEvidence) bool {
	keys := func(cells []asanaMatrixLaneCellEvidence) []string {
		out := make([]string, 0, len(cells))
		for _, cell := range cells {
			out = append(out, cell.SourceID+"\x00"+cell.Lane+"\x00"+cell.Disposition)
		}
		slices.Sort(out)
		return out
	}
	return slices.Equal(keys(left), keys(right))
}

func asanaLaneArtifactMappingFor(row asanaSourceLaneMatrixRow, lane string, cell asanaSourceLaneCell) (asanaLaneArtifactMapping, error) {
	if len(cell.Mapping) == 0 {
		return asanaLaneArtifactMapping{}, fmt.Errorf("%s backlink for %q is absent", lane, row.SourceID)
	}
	var mapping asanaLaneArtifactMapping
	if err := json.Unmarshal(cell.Mapping, &mapping); err != nil {
		return asanaLaneArtifactMapping{}, fmt.Errorf("decode %s backlink for %q: %w", lane, row.SourceID, err)
	}
	return mapping, nil
}

func asanaLaneArtifactEndpointKey(method, path string) string {
	return strings.ToUpper(method) + " " + path
}

func asanaHasArtifacts(actual []string, required ...string) bool {
	for _, requirement := range required {
		if !slices.Contains(actual, requirement) {
			return false
		}
	}
	return true
}

func asanaHasRequiredMultipartFilePart(multipart *engine.MultipartSpec, name string, maxBytes int64) bool {
	for _, part := range multipart.Parts {
		if part.Name == name && part.Type == "file" && part.Required && part.MaxBytes == maxBytes {
			return true
		}
	}
	return false
}

func asanaSameStringSet(left, right []string) bool {
	left = append([]string(nil), left...)
	right = append([]string(nil), right...)
	slices.Sort(left)
	slices.Sort(right)
	return slices.Equal(left, right)
}

func asanaSurfaceWriteActions(endpoint engine.SurfaceEndpoint) []string {
	if endpoint.CoveredBy == nil {
		return nil
	}
	actions := append([]string(nil), endpoint.CoveredBy.Writes...)
	if endpoint.CoveredBy.Write != "" {
		actions = append(actions, endpoint.CoveredBy.Write)
	}
	return actions
}

func asanaHasImplementedDirectWriteCommand(commands map[string]engine.CLICommand, action, method, path string) bool {
	for _, command := range commands {
		if command.Intent != "direct_write" || command.Availability != "implemented" || command.Write != action {
			continue
		}
		for _, endpoint := range command.APISurface {
			if strings.EqualFold(endpoint.Method, method) && endpoint.Path == path {
				return true
			}
		}
	}
	return false
}

package app

import (
	"bytes"
	"context"
	"errors"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/synccontract"
)

type scriptedSyncSource struct {
	name      string
	records   []connectors.Record
	failAfter int
	requests  []connectors.ReadRequest
	onRead    func(context.Context, connectors.ReadRequest) error
}

type orderedOpaqueCursorSource struct {
	*scriptedSyncSource
	emitted []string
}

func newOrderedOpaqueCursorSource(name string, records []connectors.Record) *orderedOpaqueCursorSource {
	return &orderedOpaqueCursorSource{scriptedSyncSource: newScriptedSyncSource(name, records)}
}

func (s *orderedOpaqueCursorSource) Read(ctx context.Context, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	s.requests = append(s.requests, req)
	var lower []byte
	if req.CursorState.Present {
		lower = append([]byte(nil), req.CursorState.Token...)
	}
	for _, record := range s.records {
		if err := ctx.Err(); err != nil {
			return err
		}
		cursor, ok := record["updated_at"].([]byte)
		if !ok {
			return errors.New("ordered opaque cursor source requires binary cursors")
		}
		if req.CursorState.Present && bytes.Compare(cursor, lower) <= 0 {
			continue
		}
		if err := emit(cloneRecord(record)); err != nil {
			return err
		}
		id, ok := record["id"].(string)
		if !ok {
			return errors.New("ordered opaque cursor source requires string identifiers")
		}
		s.emitted = append(s.emitted, id)
	}
	return nil
}

func (s *orderedOpaqueCursorSource) CursorStateFromRecord(record connectors.Record, field string) (connectors.OpaqueCursorState, error) {
	value, ok := record[field].([]byte)
	if !ok {
		return connectors.OpaqueCursorState{}, errors.New("ordered opaque cursor source requires binary cursor state")
	}
	return connectors.OpaqueCursorState{Token: append([]byte(nil), value...), Present: true}, nil
}

func (s *orderedOpaqueCursorSource) ValidateCursorField(_ connectors.RuntimeConfig, field string) error {
	if strings.TrimSpace(field) == "" {
		return errors.New("ordered opaque cursor source requires a cursor field")
	}
	return nil
}

func (s *orderedOpaqueCursorSource) CompareCursorStates(left, right connectors.OpaqueCursorState) (int, error) {
	if !left.Present || !right.Present {
		return 0, errors.New("ordered opaque cursor source requires cursor states")
	}
	return bytes.Compare(left.Token, right.Token), nil
}

type boundOrderedOpaqueCursorSource struct {
	*orderedOpaqueCursorSource
}

func newBoundOrderedOpaqueCursorSource(name string, records []connectors.Record) *boundOrderedOpaqueCursorSource {
	return &boundOrderedOpaqueCursorSource{orderedOpaqueCursorSource: newOrderedOpaqueCursorSource(name, records)}
}

func (s *boundOrderedOpaqueCursorSource) ValidateCursorField(config connectors.RuntimeConfig, field string) error {
	if configured := strings.TrimSpace(config.Config["cursor_field"]); configured == "" || configured != strings.TrimSpace(field) {
		return errors.New("bound ordered cursor source field does not match config")
	}
	return nil
}

func newScriptedSyncSource(name string, records []connectors.Record) *scriptedSyncSource {
	return &scriptedSyncSource{name: name, records: records, failAfter: -1}
}

func (s *scriptedSyncSource) Name() string { return s.name }

func (s *scriptedSyncSource) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:         s.Name(),
		DisplayName:  "Scripted Sync Source",
		Description:  "Scripted source for sync-mode tests.",
		Capabilities: connectors.Capabilities{Check: true, Catalog: true, Read: true},
	}
}

func (s *scriptedSyncSource) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	return ctx.Err()
}

func (s *scriptedSyncSource) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	return connectors.Catalog{Connector: s.Name(), Streams: []connectors.Stream{{
		Name:         "records",
		Description:  "Scripted records.",
		PrimaryKey:   []string{"id"},
		CursorFields: []string{"updated_at"},
	}}}, ctx.Err()
}

func (s *scriptedSyncSource) Read(ctx context.Context, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	s.requests = append(s.requests, req)
	if s.onRead != nil {
		if err := s.onRead(ctx, req); err != nil {
			return err
		}
	}
	for i, record := range s.records {
		if err := ctx.Err(); err != nil {
			return err
		}
		if err := emit(cloneRecord(record)); err != nil {
			return err
		}
		if s.failAfter >= 0 && i+1 >= s.failAfter {
			return errors.New("scripted source failure")
		}
	}
	return nil
}

func (s *scriptedSyncSource) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

func setupSyncModeApp(t *testing.T, source connectors.Connector, mode string) (*App, string) {
	return setupSyncModeAppWithCompatibility(t, source, mode, false)
}

func setupSyncModeAppWithCompatibility(t *testing.T, source connectors.Connector, mode string, legacyCompatibility bool) (*App, string) {
	t.Helper()
	ctx := context.Background()
	root := t.TempDir()
	if err := InitProject(root); err != nil {
		t.Fatal(err)
	}
	a, err := Open(root)
	if err != nil {
		t.Fatal(err)
	}
	registry := connectors.NewRegistry()
	registry.Register(source)
	a.registry = registry
	if _, err := a.AddCredential(ctx, AddCredentialRequest{Name: "source", Connector: source.Name()}); err != nil {
		t.Fatal(err)
	}
	if _, err := a.AddCredential(ctx, AddCredentialRequest{
		Name:      "warehouse",
		Connector: "warehouse",
		Config:    map[string]string{"path": filepath.Join(root, ".polymetrics", "warehouse")},
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := a.CreateConnection(ctx, CreateConnectionRequest{
		Name:        "records_to_warehouse",
		Source:      EndpointConfig{Connector: source.Name(), Credential: "source"},
		Destination: EndpointConfig{Connector: "warehouse", Credential: "warehouse"},
		Streams: map[string]StreamConfig{
			"records": {
				SyncMode:            mode,
				LegacyCompatibility: legacyCompatibility,
				CursorField:         "updated_at",
				PrimaryKey:          []string{"id"},
				DestinationTable:    "records",
			},
		},
	}); err != nil {
		t.Fatal(err)
	}
	return a, "records_to_warehouse"
}

func TestStreamIDIsPersistedAndSurvivesStreamRename(t *testing.T) {
	source := newScriptedSyncSource("stream_identity", nil)
	a, connectionName := setupSyncModeApp(t, source, "full_refresh_overwrite")
	connection, ok := a.findConnection(connectionName)
	if !ok {
		t.Fatalf("connection %q was not created", connectionName)
	}
	stream := connection.Streams["records"]
	if stream.StreamID == "" {
		t.Fatal("created stream has no immutable stream ID")
	}
	streamID := stream.StreamID

	for index := range a.state.Connections {
		if a.state.Connections[index].Name != connectionName {
			continue
		}
		stream = a.state.Connections[index].Streams["records"]
		stream.StreamID = "" // simulate a pre-#3981 persisted stream.
		a.state.Connections[index].Streams["records"] = stream
		break
	}
	if err := a.save(); err != nil {
		t.Fatal(err)
	}

	migrated, err := Open(a.root)
	if err != nil {
		t.Fatal(err)
	}
	connection, ok = migrated.findConnection(connectionName)
	if !ok {
		t.Fatalf("migrated connection %q was not found", connectionName)
	}
	stream = connection.Streams["records"]
	if stream.StreamID == "" || stream.StreamID == streamID {
		t.Fatalf("legacy stream ID migration = %q, want a newly persisted immutable ID", stream.StreamID)
	}
	migratedID := stream.StreamID

	for index := range migrated.state.Connections {
		if migrated.state.Connections[index].Name != connectionName {
			continue
		}
		stream = migrated.state.Connections[index].Streams["records"]
		delete(migrated.state.Connections[index].Streams, "records")
		migrated.state.Connections[index].Streams["orders-renamed"] = stream
		break
	}
	if err := migrated.save(); err != nil {
		t.Fatal(err)
	}

	reopened, err := Open(migrated.root)
	if err != nil {
		t.Fatal(err)
	}
	connection, ok = reopened.findConnection(connectionName)
	if !ok {
		t.Fatalf("renamed connection %q was not found", connectionName)
	}
	if got := connection.Streams["orders-renamed"].StreamID; got != migratedID {
		t.Fatalf("stream ID after rename = %q, want %q", got, migratedID)
	}
}

func TestConnectionStreamIdentityDoesNotEscapeAppState(t *testing.T) {
	ctx := context.Background()
	source := newScriptedSyncSource("stream_identity_boundary", nil)
	a, _ := setupSyncModeApp(t, source, "full_refresh_overwrite")
	request := CreateConnectionRequest{
		Name:        "stream_identity_boundary",
		Source:      EndpointConfig{Connector: source.Name(), Credential: "source"},
		Destination: EndpointConfig{Connector: "warehouse", Credential: "warehouse"},
		Streams: map[string]StreamConfig{
			"records": {
				SyncMode:         "full_refresh_overwrite",
				CursorField:      "updated_at",
				PrimaryKey:       []string{"id"},
				DestinationTable: "stream_identity_boundary",
			},
		},
	}

	created, err := a.CreateConnection(ctx, request)
	if err != nil {
		t.Fatal(err)
	}
	if got := request.Streams["records"].StreamID; got != "" {
		t.Fatalf("CreateConnection() assigned caller stream ID %q", got)
	}
	streamID := created.Streams["records"].StreamID
	if streamID == "" {
		t.Fatal("CreateConnection() returned an empty stream ID")
	}

	createdStream := created.Streams["records"]
	createdStream.StreamID = "stream_mutated_return"
	createdStream.PrimaryKey[0] = "mutated_return"
	created.Streams["records"] = createdStream

	listed := a.ListConnections()
	for _, connection := range listed {
		if connection.Name != request.Name {
			continue
		}
		listedStream := connection.Streams["records"]
		listedStream.StreamID = "stream_mutated_list"
		listedStream.PrimaryKey[0] = "mutated_list"
		connection.Streams["records"] = listedStream
		break
	}

	if err := a.save(); err != nil {
		t.Fatal(err)
	}
	reopened, err := Open(a.root)
	if err != nil {
		t.Fatal(err)
	}
	stored, ok := reopened.findConnection(request.Name)
	if !ok {
		t.Fatalf("connection %q was not persisted", request.Name)
	}
	if got := stored.Streams["records"].StreamID; got != streamID {
		t.Fatalf("persisted stream ID = %q, want %q", got, streamID)
	}
	if got := stored.Streams["records"].PrimaryKey; len(got) != 1 || got[0] != "id" {
		t.Fatalf("persisted stream primary key = %v, want [id]", got)
	}
}

func TestCreateConnectionRejectsCallerAssignedStreamID(t *testing.T) {
	ctx := context.Background()
	source := newScriptedSyncSource("stream_identity_assigned_input", nil)
	a, _ := setupSyncModeApp(t, source, "full_refresh_overwrite")
	request := CreateConnectionRequest{
		Name:        "stream_identity_assigned_input",
		Source:      EndpointConfig{Connector: source.Name(), Credential: "source"},
		Destination: EndpointConfig{Connector: "warehouse", Credential: "warehouse"},
		Streams: map[string]StreamConfig{
			"records": {
				StreamID:         "stream_caller_assigned",
				SyncMode:         "full_refresh_overwrite",
				CursorField:      "updated_at",
				PrimaryKey:       []string{"id"},
				DestinationTable: "stream_identity_assigned_input",
			},
		},
	}

	if _, err := a.CreateConnection(ctx, request); err == nil || !strings.Contains(err.Error(), "stream identity is assigned by the application") {
		t.Fatalf("CreateConnection() error = %v, want caller-assigned stream identity refusal", err)
	}
	if got := request.Streams["records"].StreamID; got != "stream_caller_assigned" {
		t.Fatalf("CreateConnection() changed caller stream ID = %q", got)
	}
	if _, found := a.findConnection(request.Name); found {
		t.Fatal("CreateConnection() persisted a connection after rejecting its caller-assigned stream ID")
	}
}

func TestCreateConnectionSaveFailureLeavesRequestStreamIdentityUnchanged(t *testing.T) {
	ctx := context.Background()
	source := newScriptedSyncSource("stream_identity_retry", nil)
	a, _ := setupSyncModeApp(t, source, "full_refresh_overwrite")
	request := CreateConnectionRequest{
		Name:        "stream_identity_retry",
		Source:      EndpointConfig{Connector: source.Name(), Credential: "source"},
		Destination: EndpointConfig{Connector: "warehouse", Credential: "warehouse"},
		Streams: map[string]StreamConfig{
			"records": {
				SyncMode:         "full_refresh_overwrite",
				CursorField:      "updated_at",
				PrimaryKey:       []string{"id"},
				DestinationTable: "stream_identity_retry",
			},
		},
	}

	a.store.Path = ""
	if _, err := a.CreateConnection(ctx, request); err == nil {
		t.Fatal("CreateConnection() error = nil, want persistence failure")
	}
	if got := request.Streams["records"].StreamID; got != "" {
		t.Fatalf("failed CreateConnection() assigned caller stream ID %q", got)
	}

	retry, err := Open(a.root)
	if err != nil {
		t.Fatal(err)
	}
	retry.registry = a.registry
	created, err := retry.CreateConnection(ctx, request)
	if err != nil {
		t.Fatalf("CreateConnection() retry error = %v", err)
	}
	if created.Streams["records"].StreamID == "" {
		t.Fatal("CreateConnection() retry returned an empty stream ID")
	}
}

func TestAllocateUniqueIdentityRetriesCollisions(t *testing.T) {
	used := map[string]struct{}{"stream_duplicate": {}}
	generated := []string{"stream_duplicate", "stream_unique"}
	index := 0
	identity, err := allocateUniqueIdentity("stream", used, func() (string, error) {
		identity := generated[index]
		index++
		return identity, nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if identity != "stream_unique" || index != 2 {
		t.Fatalf("allocateUniqueIdentity() = %q after %d attempts, want unique retry", identity, index)
	}
	if _, ok := used[identity]; !ok {
		t.Fatalf("allocated identity %q was not reserved", identity)
	}
}

func rowsByID(rows []connectors.Record) map[string]connectors.Record {
	out := map[string]connectors.Record{}
	for _, row := range rows {
		out[toComparableString(row["id"])] = row
	}
	return out
}

func TestParseSyncModeMatrix(t *testing.T) {
	tests := []struct {
		raw            string
		source         SourceSyncMode
		destination    DestinationSyncMode
		requiresCursor bool
		requiresPK     bool
	}{
		{"full_refresh_append", SourceSyncFullRefresh, DestinationSyncAppend, false, false},
		{"full_refresh_overwrite", SourceSyncFullRefresh, DestinationSyncOverwrite, false, false},
		{"full_refresh_overwrite_deduped", SourceSyncFullRefresh, DestinationSyncOverwriteDeduped, true, true},
		{"incremental_append", SourceSyncIncremental, DestinationSyncAppend, true, false},
		{"incremental_append_deduped", SourceSyncIncremental, DestinationSyncAppendDeduped, true, true},
	}
	for _, tt := range tests {
		t.Run(tt.raw, func(t *testing.T) {
			mode, err := ParseSyncMode(tt.raw)
			if err != nil {
				t.Fatalf("ParseSyncMode() error = %v", err)
			}
			if mode.Source != tt.source || mode.Destination != tt.destination {
				t.Fatalf("mode = %+v, want source=%s destination=%s", mode, tt.source, tt.destination)
			}
			if mode.RequiresCursor() != tt.requiresCursor || mode.RequiresPrimaryKey() != tt.requiresPK {
				t.Fatalf("requirements cursor=%v pk=%v, want cursor=%v pk=%v", mode.RequiresCursor(), mode.RequiresPrimaryKey(), tt.requiresCursor, tt.requiresPK)
			}
		})
	}
	if _, err := ParseSyncMode("full_refresh_replace"); err == nil {
		t.Fatal("ParseSyncMode(invalid) error = nil")
	} else {
		var unsupported *UnsupportedSyncModeError
		if !errors.As(err, &unsupported) || unsupported.Mode != "full_refresh_replace" {
			t.Fatalf("ParseSyncMode(invalid) error = %T %[1]v, want UnsupportedSyncModeError for the supplied mode", err)
		}
	}
}

func TestValidateSyncModeRequirements(t *testing.T) {
	if err := ValidateStreamSyncConfig(StreamConfig{SyncMode: "incremental_append"}); err == nil {
		t.Fatal("ValidateStreamSyncConfig(incremental without cursor) error = nil")
	}
	if err := ValidateStreamSyncConfig(StreamConfig{SyncMode: "incremental_append_deduped", CursorField: "updated_at"}); err == nil {
		t.Fatal("ValidateStreamSyncConfig(deduped without primary key) error = nil")
	}
	if err := ValidateStreamSyncConfig(StreamConfig{SyncMode: "incremental_append_deduped", CursorField: "updated_at", PrimaryKey: []string{"id"}}); err != nil {
		t.Fatalf("ValidateStreamSyncConfig(valid) error = %v", err)
	}
}

func TestParseSyncModeSeparatesLegacyCompatibilityFromNewNativeAdmission(t *testing.T) {
	legacy, err := ParseStreamSyncMode(StreamConfig{SyncMode: "incremental_append", LegacyCompatibility: true})
	if err != nil {
		t.Fatal(err)
	}
	if legacy.ContractMode != "incremental_append" || !legacy.LegacyCompatibility || legacy.IsContractMode() {
		t.Fatalf("legacy incremental mode = %+v, want compatibility adapter", legacy)
	}

	contract, err := ParseStreamSyncMode(StreamConfig{SyncMode: "incremental_append"})
	if err != nil {
		t.Fatal(err)
	}
	if contract.ContractMode != "incremental_append" || contract.LegacyCompatibility || !contract.IsContractMode() {
		t.Fatalf("incremental_append mode = %+v, want native contract admission", contract)
	}
}

func TestDedupedLegacyAliasesUseTypedContractsBeforeSourceIO(t *testing.T) {
	tests := []struct {
		name     string
		mode     string
		contract synccontract.Mode
	}{
		{name: "full overwrite", mode: "full_refresh_overwrite_deduped", contract: synccontract.ModeFullOverwrite},
		{name: "incremental dedupe", mode: "incremental_append_deduped", contract: synccontract.ModeIncrementalDedupe},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for _, legacyCompatibility := range []bool{false, true} {
				t.Run("legacy_compatibility_"+strconv.FormatBool(legacyCompatibility), func(t *testing.T) {
					parsed, err := ParseStreamSyncMode(StreamConfig{SyncMode: tt.mode, LegacyCompatibility: legacyCompatibility})
					if err != nil {
						t.Fatal(err)
					}
					if parsed.ContractMode != tt.contract || !parsed.IsContractMode() {
						t.Fatalf("parsed %q = %+v, want typed %q contract", tt.mode, parsed, tt.contract)
					}

					source := newScriptedSyncSource("typed_legacy_"+tt.mode+"_"+strconv.FormatBool(legacyCompatibility), nil)
					a, connection := setupSyncModeAppWithCompatibility(t, source, tt.mode, legacyCompatibility)
					_, err = a.RunETL(context.Background(), RunETLRequest{Connection: connection, Stream: "records", BatchSize: 1})
					var modeErr *synccontract.ModeNotExecutableError
					if !errors.As(err, &modeErr) || modeErr.Mode != tt.contract {
						t.Fatalf("RunETL(%q) error = %v, want typed pre-I/O refusal for %q", tt.mode, err, tt.contract)
					}
					if len(source.requests) != 0 {
						t.Fatalf("RunETL(%q) source reads = %d, want typed refusal before I/O", tt.mode, len(source.requests))
					}
				})
			}
		})
	}
}

func TestCanonicalSyncModesRetainParsedContracts(t *testing.T) {
	tests := []struct {
		mode                synccontract.Mode
		source              SourceSyncMode
		destination         DestinationSyncMode
		legacyCompatibility bool
		contractAdmission   bool
		runtimeTypedRefusal bool
	}{
		{mode: synccontract.ModeFullOverwrite, source: SourceSyncFullRefresh, destination: DestinationSyncOverwrite, contractAdmission: true, runtimeTypedRefusal: true},
		{mode: synccontract.ModeFullAppend, source: SourceSyncFullRefresh, destination: DestinationSyncAppend, contractAdmission: true, runtimeTypedRefusal: true},
		{mode: synccontract.ModeIncrementalAppend, source: SourceSyncIncremental, destination: DestinationSyncAppend, contractAdmission: true},
		{mode: synccontract.ModeIncrementalUpsert, source: SourceSyncIncremental, destination: DestinationSyncUpsert, contractAdmission: true, runtimeTypedRefusal: true},
		{mode: synccontract.ModeIncrementalDedupe, source: SourceSyncIncremental, destination: DestinationSyncAppendDeduped, contractAdmission: true, runtimeTypedRefusal: true},
		{mode: synccontract.ModeIncrementalDedupeHistory, source: SourceSyncIncremental, destination: DestinationSyncDedupeHistory, contractAdmission: true, runtimeTypedRefusal: true},
		{mode: synccontract.ModeChangeCapture, source: SourceSyncChangeCapture, destination: DestinationSyncUpsert, contractAdmission: true, runtimeTypedRefusal: true},
	}
	for _, tt := range tests {
		t.Run(string(tt.mode), func(t *testing.T) {
			parsed, err := ParseSyncMode(string(tt.mode))
			if err != nil {
				t.Fatal(err)
			}
			if parsed.ContractMode != tt.mode || parsed.Source != tt.source || parsed.Destination != tt.destination || parsed.LegacyCompatibility != tt.legacyCompatibility || parsed.IsContractMode() != tt.contractAdmission {
				t.Fatalf("ParseSyncMode(%q) = %+v, want unchanged typed canonical mode", tt.mode, parsed)
			}

			source := newScriptedSyncSource("canonical_"+string(tt.mode), nil)
			a, connection := setupSyncModeApp(t, source, string(tt.mode))
			_, err = a.RunETL(context.Background(), RunETLRequest{Connection: connection, Stream: "records", BatchSize: 1})
			if !tt.runtimeTypedRefusal {
				if err != nil {
					t.Fatalf("RunETL(%q) error = %v, want unchanged legacy-compatible execution", tt.mode, err)
				}
				if len(source.requests) == 0 {
					t.Fatalf("RunETL(%q) made no source read, want unchanged legacy-compatible execution", tt.mode)
				}
				return
			}
			var modeErr *synccontract.ModeNotExecutableError
			if !errors.As(err, &modeErr) || modeErr.Mode != tt.mode {
				t.Fatalf("RunETL(%q) error = %v, want unchanged typed pre-I/O refusal", tt.mode, err)
			}
			if len(source.requests) != 0 {
				t.Fatalf("RunETL(%q) source reads = %d, want no read before typed refusal", tt.mode, len(source.requests))
			}
		})
	}
}

func TestCreateConnectionRetainsPublicLegacyConfigurationAndAdmitsTypedAliases(t *testing.T) {
	typedAliases := map[string]bool{
		"full_refresh_overwrite_deduped": true,
		"incremental_append_deduped":     true,
	}
	for _, modeName := range MustSyncModeNames() {
		t.Run(modeName, func(t *testing.T) {
			source := newScriptedSyncSource("legacy_creation_"+modeName, nil)
			a, connection := setupSyncModeAppWithCompatibility(t, source, modeName, false)
			conn, ok := a.findConnection(connection)
			if !ok {
				t.Fatal("connection missing")
			}
			stream := conn.Streams["records"]
			if !stream.LegacyCompatibility {
				t.Fatalf("fresh legacy stream %q was not marked as a compatibility adapter", modeName)
			}
			parsed, err := ParseStreamSyncMode(stream)
			if err != nil {
				t.Fatal(err)
			}
			if parsed.IsContractMode() != typedAliases[modeName] {
				t.Fatalf("fresh legacy stream %q typed admission = %v, want %v", modeName, parsed.IsContractMode(), typedAliases[modeName])
			}
		})
	}
}

func TestLegacyStateMigrationMarksExistingIncrementalAppendAdapter(t *testing.T) {
	source := newScriptedSyncSource("legacy_incremental_migration", nil)
	a, connection := setupSyncModeApp(t, source, "incremental_append")
	a.state.SyncModeCompatibilityVersion = 0
	for connectionIndex := range a.state.Connections {
		if a.state.Connections[connectionIndex].Name != connection {
			continue
		}
		stream := a.state.Connections[connectionIndex].Streams["records"]
		stream.LegacyCompatibility = false
		a.state.Connections[connectionIndex].Streams["records"] = stream
	}

	a.migrateLegacySyncModeCompatibility()
	conn, ok := a.findConnection(connection)
	if !ok {
		t.Fatal("connection missing")
	}
	if !conn.Streams["records"].LegacyCompatibility {
		t.Fatal("legacy incremental stream was not marked as an explicit adapter")
	}
	if a.state.SyncModeCompatibilityVersion != syncModeCompatibilityVersion {
		t.Fatalf("compatibility version = %d, want %d", a.state.SyncModeCompatibilityVersion, syncModeCompatibilityVersion)
	}
}

func TestFullRefreshAppendDuplicatesAcrossRuns(t *testing.T) {
	ctx := context.Background()
	source := newScriptedSyncSource("scripted_append", []connectors.Record{
		{"id": "a", "name": "Ada", "updated_at": "2026-01-01T00:00:00Z"},
		{"id": "g", "name": "Grace", "updated_at": "2026-01-02T00:00:00Z"},
	})
	a, connection := setupSyncModeApp(t, source, "full_refresh_append")

	for i := 0; i < 2; i++ {
		run, err := a.RunETL(ctx, RunETLRequest{Connection: connection, Stream: "records", BatchSize: 1})
		if err != nil {
			t.Fatalf("RunETL(%d) error = %v", i, err)
		}
		if run.Checkpoint["sync_mode"] != "full_refresh_append" {
			t.Fatalf("sync_mode checkpoint = %q", run.Checkpoint["sync_mode"])
		}
	}
	rows, err := a.QueryTable(ctx, QueryTableRequest{Table: "records", Limit: 10})
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 4 {
		t.Fatalf("rows len = %d, want 4 append duplicates", len(rows))
	}
}

func TestFullRefreshOverwriteUsesFailureSafeFinalSwap(t *testing.T) {
	ctx := context.Background()
	source := newScriptedSyncSource("scripted_overwrite", []connectors.Record{
		{"id": "a", "name": "Ada", "updated_at": "2026-01-01T00:00:00Z"},
		{"id": "g", "name": "Grace", "updated_at": "2026-01-02T00:00:00Z"},
	})
	a, connection := setupSyncModeApp(t, source, "full_refresh_overwrite")

	if _, err := a.RunETL(ctx, RunETLRequest{Connection: connection, Stream: "records", BatchSize: 1}); err != nil {
		t.Fatal(err)
	}
	source.records = []connectors.Record{{"id": "k", "name": "Katherine", "updated_at": "2026-01-03T00:00:00Z"}}
	source.failAfter = 1
	if _, err := a.RunETL(ctx, RunETLRequest{Connection: connection, Stream: "records", BatchSize: 1}); err == nil {
		t.Fatal("RunETL(failing overwrite) error = nil")
	}
	rows, err := a.QueryTable(ctx, QueryTableRequest{Table: "records", Limit: 10})
	if err != nil {
		t.Fatal(err)
	}
	if got := sortedNames(rows); got != "Ada,Grace" {
		t.Fatalf("rows after failed overwrite = %s, want Ada,Grace", got)
	}

	source.failAfter = -1
	if _, err := a.RunETL(ctx, RunETLRequest{Connection: connection, Stream: "records", BatchSize: 1}); err != nil {
		t.Fatal(err)
	}
	rows, err = a.QueryTable(ctx, QueryTableRequest{Table: "records", Limit: 10})
	if err != nil {
		t.Fatal(err)
	}
	if got := sortedNames(rows); got != "Katherine" {
		t.Fatalf("rows after successful overwrite = %s, want Katherine", got)
	}
}

func TestIncrementalAppendCommitsCursorOnlyAfterSuccess(t *testing.T) {
	ctx := context.Background()
	source := newScriptedSyncSource("scripted_incremental", []connectors.Record{
		{"id": "a", "name": "Ada", "updated_at": "2026-01-01T00:00:00Z"},
		{"id": "g", "name": "Grace", "updated_at": "2026-01-02T00:00:00Z"},
	})
	a, connection := setupSyncModeApp(t, source, "incremental_append")

	run, err := a.RunETL(ctx, RunETLRequest{Connection: connection, Stream: "records", BatchSize: 2})
	if err != nil {
		t.Fatal(err)
	}
	if run.Checkpoint["cursor"] != "2026-01-02T00:00:00Z" {
		t.Fatalf("cursor = %q", run.Checkpoint["cursor"])
	}

	source.records = []connectors.Record{
		{"id": "g", "name": "Grace resent", "updated_at": "2026-01-02T00:00:00Z"},
		{"id": "k", "name": "Katherine", "updated_at": "2026-01-03T00:00:00Z"},
	}
	run, err = a.RunETL(ctx, RunETLRequest{Connection: connection, Stream: "records", BatchSize: 2})
	if err != nil {
		t.Fatal(err)
	}
	if run.Checkpoint["cursor"] != "2026-01-03T00:00:00Z" {
		t.Fatalf("cursor after second run = %q", run.Checkpoint["cursor"])
	}
	rows, err := a.QueryTable(ctx, QueryTableRequest{Table: "records", Limit: 10})
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 4 {
		t.Fatalf("rows len = %d, want 4 including inclusive resent cursor row", len(rows))
	}

	source.records = []connectors.Record{{"id": "m", "name": "Margaret", "updated_at": "2026-01-04T00:00:00Z"}}
	source.failAfter = 1
	if _, err := a.RunETL(ctx, RunETLRequest{Connection: connection, Stream: "records", BatchSize: 1}); err == nil {
		t.Fatal("RunETL(failing incremental) error = nil")
	}
	state := a.state.StreamStates[streamStateKey(connection, "records")]
	if cursor, present := streamStateCursor(state); !present || cursor != "2026-01-03T00:00:00Z" {
		t.Fatalf("cursor advanced after failed run = %q", cursor)
	}
}

func TestRecordCursorPreservesOpaqueEmptyAndWhitespaceValues(t *testing.T) {
	for _, cursor := range []string{"", "  "} {
		t.Run("cursor_"+strings.ReplaceAll(cursor, " ", "space"), func(t *testing.T) {
			got, err := recordCursor(connectors.Record{"updated_at": cursor}, "updated_at")
			if err != nil || got != cursor {
				t.Fatalf("recordCursor() = %q, %v, want %q, nil", got, err, cursor)
			}
		})
	}
	if compareCursor("", "  ") == 0 {
		t.Fatal("compareCursor() collapsed distinct opaque cursor values")
	}
}

func sortedNames(rows []connectors.Record) string {
	names := make([]string, 0, len(rows))
	for _, row := range rows {
		if value, ok := row["name"].(string); ok {
			names = append(names, value)
		}
	}
	sort.Strings(names)
	return strings.Join(names, ",")
}

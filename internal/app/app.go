package app

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/bundleregistry"
	"polymetrics.ai/internal/connectors/commandrunner"
	"polymetrics.ai/internal/coordination"
	"polymetrics.ai/internal/credential"
	"polymetrics.ai/internal/safety"
	statestore "polymetrics.ai/internal/state"
	"polymetrics.ai/internal/synccontract"
	"polymetrics.ai/internal/synctransport"
	"polymetrics.ai/internal/vault"
	"polymetrics.ai/internal/warehouse"
)

const (
	reversePlanModeConnectorCommand               = "connector_command"
	reversePlanStatusApprovalConsumptionUncertain = "approval_consumption_uncertain"
)

var errStateRevisionConflict = errors.New("project state changed in another process")

type App struct {
	root                    string
	projectDir              string
	statePath               string
	store                   statestore.JSONStore[state]
	state                   state
	deferStateNormalization bool
	deferredState           *state
	deferredStateRevision   uint64
	vault                   *vault.Vault
	ephemeralCredentials    *CertificationEphemeralSession
	approval                *projectWriteApprovalAuthority
	registry                *connectors.Registry
	transports              *synctransport.Registry
	transportStage          synctransport.WarehouseStage
	sqlEngine               sqlQueryEngine
	catalogs                catalogStorage
	authCohorts             *coordination.AuthCohortCoordinator
	rateParking             *coordination.RateParkingCoordinator
}

// sqlQueryEngine is the backend for App.QuerySQL. DuckDB is the only
// implementation; the seam remains so a query path can be substituted in tests
// without the engine choice becoming an install-time option.
type sqlQueryEngine interface {
	QuerySQL(ctx context.Context, req QuerySQLRequest) ([]connectors.Record, error)
	Name() string
}

type state struct {
	Revision                     uint64 `json:"revision"`
	SyncModeCompatibilityVersion uint   `json:"sync_mode_compatibility_version,omitempty"`
	// WorkspaceID is the opaque generated identifier that forms the first
	// component of every warehouse path. It is never a user-supplied name.
	WorkspaceID        string                            `json:"workspace_id,omitempty"`
	Credentials        []CredentialMeta                  `json:"credentials"`
	CredentialBindings map[string]credentialBindingState `json:"credential_bindings"`
	CoordinationSalt   string                            `json:"coordination_salt,omitempty"`
	Connections        []Connection                      `json:"connections"`
	Catalogs           []catalogReference                `json:"catalogs"`
	Runs               []Run                             `json:"runs"`
	ReversePlans       []ReversePlan                     `json:"reverse_plans"`
	ReverseRuns        []ReverseRun                      `json:"reverse_runs"`
	Authorizations     []AuthorizationRecord             `json:"authorizations,omitempty"`
	FlowActionReceipts []FlowActionReceipt               `json:"flow_action_receipts,omitempty"`
	Checkpoints        map[string]map[string]string      `json:"checkpoints,omitempty"`
	StreamStates       map[string]StreamState            `json:"stream_states,omitempty"`
}

// credentialBindingState is protected project-state metadata. The raw binding
// is deliberately absent from CredentialMeta so ordinary list, inspect, JSON
// event, and runtime surfaces cannot reveal it.
type credentialBindingState struct {
	BindingID                     string `json:"binding_id"`
	ProviderFamilyDeclared        bool   `json:"provider_family_declared"`
	AuthProfileDeclared           bool   `json:"auth_profile_declared"`
	DeclarationProvenanceRecorded bool   `json:"declaration_provenance_recorded"`
}

func (binding credentialBindingState) hasExplicitDeclarations() bool {
	return binding.DeclarationProvenanceRecorded && binding.ProviderFamilyDeclared && binding.AuthProfileDeclared
}

func InitProject(root string) error {
	if root == "" {
		root = "."
	}
	projectDir := filepath.Join(root, ".polymetrics")
	for _, dir := range []string{
		projectDir,
		filepath.Join(projectDir, "state"),
		filepath.Join(projectDir, "warehouse"),
		filepath.Join(projectDir, "outbox"),
		filepath.Join(projectDir, "logs"),
	} {
		if err := os.MkdirAll(dir, 0o700); err != nil {
			return fmt.Errorf("create %s: %w", dir, err)
		}
	}
	if activeCertificationEphemeralSession(root) == nil {
		if _, err := vault.Init(projectDir); err != nil {
			return err
		}
	}
	configPath := filepath.Join(projectDir, "config.yaml")
	if _, err := os.Stat(configPath); errors.Is(err, os.ErrNotExist) {
		config := "version: 1\nproject: polymetrics-local\nwarehouse:\n  connector: warehouse\n  path: .polymetrics/warehouse\n"
		if err := os.WriteFile(configPath, []byte(config), 0o600); err != nil {
			return fmt.Errorf("write config: %w", err)
		}
	}
	statePath := filepath.Join(projectDir, "state", "state.json")
	if _, err := os.Stat(statePath); errors.Is(err, os.ErrNotExist) {
		coordinationSalt, err := newCoordinationSalt()
		if err != nil {
			return err
		}
		workspaceID, err := prefixedID("ws")
		if err != nil {
			return err
		}
		initial := state{
			SyncModeCompatibilityVersion: syncModeCompatibilityVersion,
			WorkspaceID:                  workspaceID,
			CredentialBindings:           map[string]credentialBindingState{},
			CoordinationSalt:             coordinationSalt,
			Checkpoints:                  map[string]map[string]string{},
			StreamStates:                 map[string]StreamState{},
		}
		if err := writeJSONAtomic(statePath, initial); err != nil {
			return err
		}
	}
	return nil
}

func Open(root string) (*App, error) {
	return open(root, false)
}

func OpenForReverseExecution(root string) (*App, error) {
	return open(root, true)
}

func open(root string, deferNormalization bool) (*App, error) {
	if root == "" {
		root = "."
	}
	projectDir := filepath.Join(root, ".polymetrics")
	info, err := os.Stat(projectDir)
	if err != nil {
		return nil, fmt.Errorf("open project at %s: %w", projectDir, err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("%s is not a directory", projectDir)
	}
	ephemeralCredentials := activeCertificationEphemeralSession(root)
	var v *vault.Vault
	if ephemeralCredentials == nil {
		if deferNormalization {
			v, err = vault.OpenReadOnly(projectDir)
		} else {
			v, err = vault.Open(projectDir)
		}
		if err != nil {
			return nil, err
		}
	}
	var approval *projectWriteApprovalAuthority
	if ephemeralCredentials != nil {
		approval = ephemeralCredentials.writeApproval()
		if approval == nil {
			return nil, errors.New("certification ephemeral credential session is closed")
		}
	} else {
		approval, err = newProjectWriteApprovalAuthority(projectDir)
		if err != nil {
			return nil, err
		}
	}
	statePath := filepath.Join(projectDir, "state", "state.json")
	a := &App{
		root:                    root,
		projectDir:              projectDir,
		statePath:               statePath,
		store:                   newStateStore(statePath),
		deferStateNormalization: deferNormalization,
		vault:                   v,
		ephemeralCredentials:    ephemeralCredentials,
		approval:                approval,
		registry:                connectors.NewRegistry(),
		transports:              synctransport.NewRegistry(nil),
		catalogs:                newCatalogStorage(projectDir),
	}
	a.sqlEngine = newSQLEngine(a)
	// state.json is atomically replaced by writers, so opening a current project
	// can take a coherent read-only snapshot without contending on their legacy
	// O_EXCL writer marker. This lets a second CLI construct its durable parking
	// coordinator while the first process is resuming a claimed parked run.
	if err := a.load(false); err != nil {
		return nil, err
	}
	authStore, err := coordination.OpenFileAuthCohortHealthStore(filepath.Join(projectDir, "state", "auth-cohorts.json"))
	if err != nil {
		return nil, err
	}
	a.authCohorts = coordination.NewAuthCohortCoordinator(authStore)
	parkingStore, err := coordination.OpenFileRateParkingStore(filepath.Join(projectDir, "state", "rate-parking.json"))
	if err != nil {
		return nil, err
	}
	a.rateParking = coordination.NewRateParkingCoordinator(coordination.RateParkingCoordinatorOptions{
		Store:  parkingStore,
		Resume: a.resumeParkedRateLimitRun,
	})
	if err := a.composeTransportRegistry(); err != nil {
		return nil, err
	}
	if err := a.rateParking.Start(context.Background()); err != nil {
		return nil, fmt.Errorf("start durable rate parking: %w", err)
	}
	return a, nil
}

// IssueLabelTransportIdentity exposes the bundle-owned carrier name needed to
// render the closed transport manual before an App can be opened. It contains
// presentation metadata only; endpoint, record, credential, and action values
// remain connection- and definition-owned.
type IssueLabelTransportIdentity struct {
	ConnectorName string
	DisplayName   string
}

// DefaultIssueLabelTransportIdentity selects the same unique declarative
// definition as the production composition root. Keeping this lookup here
// prevents shared CLI code from carrying a provider name or endpoint policy.
func DefaultIssueLabelTransportIdentity() (IssueLabelTransportIdentity, error) {
	connector, _, err := issueLabelTransportEngine(bundleregistry.New())
	if err != nil {
		return IssueLabelTransportIdentity{}, err
	}
	definition := connector.Definition()
	if definition.Name != connector.Name() || strings.TrimSpace(definition.DisplayName) == "" {
		return IssueLabelTransportIdentity{}, fmt.Errorf("closed issue-label transport definition has no stable presentation identity")
	}
	return IssueLabelTransportIdentity{ConnectorName: definition.Name, DisplayName: definition.DisplayName}, nil
}

func (a *App) ProjectDir() string { return a.projectDir }

func (a *App) projectRoot() string { return filepath.Dir(a.projectDir) }

func (a *App) Registry() *connectors.Registry { return a.registry }

func (a *App) Connectors() []connectors.Metadata {
	return a.registry.List()
}

func (a *App) Connector(name string) (connectors.Metadata, error) {
	if err := connectors.RejectLegacyConnectorName(name); err != nil {
		return connectors.Metadata{}, err
	}
	c, ok := a.registry.Get(name)
	if !ok {
		return connectors.Metadata{}, fmt.Errorf("connector %q not found", name)
	}
	return c.Metadata(), nil
}

func (a *App) load(persist bool) error {
	var (
		loaded state
		err    error
	)
	if persist {
		loaded, err = a.store.Load()
	} else {
		loaded, err = a.store.LoadReadOnly()
	}
	if err != nil {
		return err
	}
	if err := a.normalizeLoadedState(loaded, persist); err != nil {
		return err
	}
	a.rememberDeferredState(loaded.Revision)
	return nil
}

// normalizeLoadedState applies every compatibility invariant after a state
// reload. Callers must not assign a store result to a.state directly.
func (a *App) normalizeLoadedState(loaded state, persist bool) error {
	a.state = loaded
	changed := false
	// Empty maps are omitted from state.json. Initialize their in-memory
	// defaults for callers, but do not turn an otherwise read-only open into a
	// revision-changing persistence operation.
	if a.state.Checkpoints == nil {
		a.state.Checkpoints = map[string]map[string]string{}
	}
	if a.state.StreamStates == nil {
		a.state.StreamStates = map[string]StreamState{}
	}
	catalogRefsChanged := a.dropInvalidCatalogReferences()
	compatibilityChanged := a.migrateLegacySyncModeCompatibility()
	coordinationChanged, err := a.migrateCredentialCoordination()
	if err != nil {
		return err
	}
	identityChanged, err := a.migrateWarehouseIdentity()
	if err != nil {
		return err
	}
	changed = changed || catalogRefsChanged || compatibilityChanged || coordinationChanged || identityChanged
	if persist && changed {
		if err := a.save(); err != nil {
			return fmt.Errorf("persist project identity: %w", err)
		}
	}
	return nil
}

func (a *App) rememberDeferredState(revision uint64) {
	if !a.deferStateNormalization {
		return
	}
	deferred := a.state
	a.deferredState = &deferred
	a.deferredStateRevision = revision
}

func (a *App) stateForApprovalConsumption(current state) (state, error) {
	if !a.deferStateNormalization {
		return current, nil
	}
	if a.deferredState != nil && current.Revision == a.deferredStateRevision {
		return cloneState(*a.deferredState)
	}
	if err := a.normalizeLoadedState(current, false); err != nil {
		return state{}, err
	}
	a.rememberDeferredState(current.Revision)
	return cloneState(a.state)
}

func cloneState(value state) (state, error) {
	data, err := json.Marshal(value)
	if err != nil {
		return state{}, fmt.Errorf("copy state: %w", err)
	}
	var copied state
	if err := json.Unmarshal(data, &copied); err != nil {
		return state{}, fmt.Errorf("copy state: %w", err)
	}
	return copied, nil
}

// migrateWarehouseIdentity gives a project, its connections, and their streams
// opaque persisted identifiers. Workspace and connection IDs form stable
// warehouse path components; stream IDs preserve managed-target identity across
// map-key, display-name, and destination-table changes.
func (a *App) migrateWarehouseIdentity() (bool, error) {
	changed := false
	if strings.TrimSpace(a.state.WorkspaceID) == "" {
		workspaceID, err := prefixedID("ws")
		if err != nil {
			return false, err
		}
		a.state.WorkspaceID = workspaceID
		changed = true
	}
	connectionIDs := make(map[string]struct{}, len(a.state.Connections))
	for _, connection := range a.state.Connections {
		if strings.TrimSpace(connection.ID) == "" {
			continue
		}
		if _, exists := connectionIDs[connection.ID]; exists {
			return false, errors.New("duplicate persisted connection identity")
		}
		connectionIDs[connection.ID] = struct{}{}
	}
	for index := range a.state.Connections {
		connection := &a.state.Connections[index]
		if strings.TrimSpace(connection.ID) == "" {
			connectionID, err := allocateUniquePrefixedID("conn", connectionIDs)
			if err != nil {
				return false, err
			}
			connection.ID = connectionID
			changed = true
		}
		streamIDs := make(map[string]struct{}, len(connection.Streams))
		for _, stream := range connection.Streams {
			if strings.TrimSpace(stream.StreamID) == "" {
				continue
			}
			if _, exists := streamIDs[stream.StreamID]; exists {
				return false, errors.New("duplicate persisted stream identity")
			}
			streamIDs[stream.StreamID] = struct{}{}
		}
		for name, stream := range connection.Streams {
			if strings.TrimSpace(stream.StreamID) != "" {
				continue
			}
			streamID, err := allocateUniquePrefixedID("stream", streamIDs)
			if err != nil {
				return false, err
			}
			stream.StreamID = streamID
			connection.Streams[name] = stream
			changed = true
		}
	}
	return changed, nil
}

func (a *App) dropInvalidCatalogReferences() bool {
	kept := a.state.Catalogs[:0]
	changed := false
	for _, reference := range a.state.Catalogs {
		if !validCatalogReference(reference) {
			changed = true
			continue
		}
		kept = append(kept, reference)
	}
	a.state.Catalogs = kept
	return changed
}

// migrateCredentialCoordination gives pre-#3863 credentials an isolated
// non-secret binding and declaration defaults. It intentionally never opens
// the vault: coordination is an explicit relationship, not secret equality.
func (a *App) migrateCredentialCoordination() (bool, error) {
	changed := false
	if a.state.CredentialBindings == nil {
		a.state.CredentialBindings = map[string]credentialBindingState{}
		changed = true
	}
	if strings.TrimSpace(a.state.CoordinationSalt) == "" {
		salt, err := newCoordinationSalt()
		if err != nil {
			return false, err
		}
		a.state.CoordinationSalt = salt
		changed = true
	}
	for index := range a.state.Credentials {
		credential := &a.state.Credentials[index]
		providerFamily, authProfile, err := credentialCoordinationDeclarations(
			credential.Connector,
			credential.ProviderFamily,
			credential.AuthProfile,
		)
		if err != nil {
			return false, fmt.Errorf("migrate credential coordination metadata: %w", err)
		}
		if credential.ProviderFamily != providerFamily {
			credential.ProviderFamily = providerFamily
			changed = true
		}
		if credential.AuthProfile != authProfile {
			credential.AuthProfile = authProfile
			changed = true
		}
		binding, ok := a.state.CredentialBindings[credential.ID]
		if !ok || strings.TrimSpace(binding.BindingID) == "" {
			bindingID, err := prefixedID("cbind")
			if err != nil {
				return false, err
			}
			binding = credentialBindingState{BindingID: bindingID}
			changed = true
		}
		if !binding.DeclarationProvenanceRecorded {
			binding.ProviderFamilyDeclared = false
			binding.AuthProfileDeclared = false
			binding.DeclarationProvenanceRecorded = true
			changed = true
		}
		a.state.CredentialBindings[credential.ID] = binding
		if _, err := a.newCoordinationIdentity(providerFamily, authProfile, binding.BindingID); err != nil {
			return false, fmt.Errorf("migrate credential coordination metadata: %w", err)
		}
	}
	isolationChanged, err := a.isolateUnverifiedCrossConnectorBindings()
	if err != nil {
		return false, fmt.Errorf("migrate credential coordination metadata: %w", err)
	}
	changed = changed || isolationChanged
	return changed, nil
}

func (a *App) isolateUnverifiedCrossConnectorBindings() (bool, error) {
	credentialsByBinding := make(map[string][]int, len(a.state.Credentials))
	for index, credential := range a.state.Credentials {
		binding, ok := a.state.CredentialBindings[credential.ID]
		if !ok || strings.TrimSpace(binding.BindingID) == "" {
			return false, errors.New("credential coordination metadata is unavailable")
		}
		credentialsByBinding[binding.BindingID] = append(credentialsByBinding[binding.BindingID], index)
	}

	changed := false
	for _, indexes := range credentialsByBinding {
		if len(indexes) < 2 {
			continue
		}
		connector := a.state.Credentials[indexes[0]].Connector
		crossConnector := false
		for _, index := range indexes[1:] {
			if a.state.Credentials[index].Connector != connector {
				crossConnector = true
				break
			}
		}
		if !crossConnector {
			continue
		}
		for _, index := range indexes {
			credential := a.state.Credentials[index]
			binding := a.state.CredentialBindings[credential.ID]
			if binding.hasExplicitDeclarations() {
				continue
			}
			bindingID, err := prefixedID("cbind")
			if err != nil {
				return false, err
			}
			binding.BindingID = bindingID
			a.state.CredentialBindings[credential.ID] = binding
			changed = true
		}
	}
	return changed, nil
}

func newCoordinationSalt() (string, error) {
	salt, err := randomToken(32)
	if err != nil {
		return "", fmt.Errorf("generate coordination salt: %w", err)
	}
	return salt, nil
}

func (a *App) newCoordinationIdentity(providerFamily, authProfile, bindingID string) (connectors.CoordinationIdentity, error) {
	if strings.TrimSpace(a.state.CoordinationSalt) == "" {
		return connectors.CoordinationIdentity{}, errors.New("credential coordination identity is unavailable")
	}
	return connectors.NewCoordinationIdentity([]byte(a.state.CoordinationSalt), connectors.CredentialBinding{
		BindingID:      bindingID,
		ProviderFamily: providerFamily,
		AuthProfile:    authProfile,
	})
}

func (a *App) coordinationIdentityForCredential(credential CredentialMeta) (connectors.CoordinationIdentity, error) {
	binding, err := a.credentialBindingForCredential(credential)
	if err != nil {
		return connectors.CoordinationIdentity{}, err
	}
	return a.newCoordinationIdentity(credential.ProviderFamily, credential.AuthProfile, binding.BindingID)
}

func (a *App) credentialBindingForCredential(credential CredentialMeta) (credentialBindingState, error) {
	binding, ok := a.state.CredentialBindings[credential.ID]
	if !ok || strings.TrimSpace(binding.BindingID) == "" {
		return credentialBindingState{}, errors.New("credential coordination metadata is unavailable")
	}
	return binding, nil
}

func (a *App) migrateLegacySyncModeCompatibility() bool {
	if a.state.SyncModeCompatibilityVersion >= syncModeCompatibilityVersion {
		return false
	}
	for connectionIndex := range a.state.Connections {
		for streamName, stream := range a.state.Connections[connectionIndex].Streams {
			if isLegacySyncModeName(stream.SyncMode) {
				stream.LegacyCompatibility = true
				a.state.Connections[connectionIndex].Streams[streamName] = stream
			}
		}
	}
	a.state.SyncModeCompatibilityVersion = syncModeCompatibilityVersion
	return true
}

func (a *App) save() error {
	expectedRevision := a.state.Revision
	next := a.state
	updated, err := a.store.Update(func(current state) (state, error) {
		if current.Revision != expectedRevision {
			return current, errStateRevisionConflict
		}
		next.Revision = current.Revision + 1
		return next, nil
	})
	if err == nil || errors.Is(err, errStateRevisionConflict) || stateStoreCommitMayHaveSucceeded(err) {
		a.state = updated
	}
	return err
}

func (a *App) updateState(update func(state) (state, error)) (state, error) {
	updated, err := a.store.Update(func(current state) (state, error) {
		next, updateErr := update(current)
		if updateErr != nil {
			return current, updateErr
		}
		next.Revision = current.Revision + 1
		return next, nil
	})
	if err == nil || stateStoreCommitMayHaveSucceeded(err) {
		a.state = updated
	}
	return updated, err
}

func (a *App) updateStateAfterPreflight(preflight func(state) error, update func(state) (state, error)) (state, error) {
	updated, err := a.store.UpdateAfterPreflight(preflight, func(current state) (state, error) {
		next, updateErr := update(current)
		if updateErr != nil {
			return current, updateErr
		}
		next.Revision = current.Revision + 1
		return next, nil
	})
	if err == nil || stateStoreCommitMayHaveSucceeded(err) {
		a.state = updated
	}
	return updated, err
}

func stateStoreCommitMayHaveSucceeded(err error) bool {
	return statestore.CommitOutcomeForError(err).MayHaveCommitted()
}

// terminalStateReload is the only way a terminal presentation path may use a
// state update that reported a post-rename or unlock failure. Its State is a
// fresh read of the durable file, never the update callback's speculative
// value.
type terminalStateReload struct {
	State            state
	MayHaveCommitted bool
}

// reloadExactTerminalState classifies an update result once for ordinary ETL
// and reverse finalization. A definite no-commit has no terminal result to
// present. A may-have-committed result must be reloaded before any caller can
// return it to the CLI or another durable boundary.
func (a *App) reloadExactTerminalState(persistErr error) (terminalStateReload, error) {
	if persistErr == nil || !stateStoreCommitMayHaveSucceeded(persistErr) {
		return terminalStateReload{}, nil
	}
	if a == nil || strings.TrimSpace(a.store.Path) == "" {
		return terminalStateReload{}, errors.New("durable terminal state reload is unavailable")
	}
	loaded, err := a.store.LoadReadOnly()
	if err != nil {
		return terminalStateReload{}, fmt.Errorf("reload durable terminal state: %w", err)
	}
	if err := a.normalizeLoadedState(loaded, false); err != nil {
		return terminalStateReload{}, fmt.Errorf("normalize reloaded terminal state: %w", err)
	}
	a.rememberDeferredState(loaded.Revision)
	return terminalStateReload{State: a.state, MayHaveCommitted: true}, nil
}

func terminalETLRunFromState(loaded state, runID string) (Run, error) {
	for _, run := range loaded.Runs {
		if run.ID != runID {
			continue
		}
		if !IsTerminalETLRunStatus(run.Status) || run.CompletedAt.IsZero() {
			return Run{}, fmt.Errorf("durable ETL run %q is not terminal", runID)
		}
		return run, nil
	}
	return Run{}, fmt.Errorf("durable ETL run %q was not found", runID)
}

// IsTerminalETLRunStatus reports statuses that are durably presentable to
// callers. A reconciliation-required run is terminal delivery evidence, not a
// runnable work item: it can only be repaired from its stored receipt.
func IsTerminalETLRunStatus(status string) bool {
	return status == "completed" || status == "failed" || status == ETLRunStatusDeliveredReconciliationRequired
}

func terminalReverseRunFromState(loaded state, planID, runID string) (ReverseRun, error) {
	plan, err := reversePlanFromState(loaded, planID)
	if err != nil {
		return ReverseRun{}, fmt.Errorf("reload durable reverse plan: %w", err)
	}
	if plan.Status != "executed" && plan.Status != "failed" {
		return ReverseRun{}, fmt.Errorf("durable reverse plan %q is not terminal", planID)
	}
	for _, run := range loaded.ReverseRuns {
		if run.ID != runID {
			continue
		}
		if run.PlanID != planID || (run.Status != "completed" && run.Status != "failed") || run.CompletedAt.IsZero() {
			return ReverseRun{}, fmt.Errorf("durable reverse run %q is not terminal", runID)
		}
		return run, nil
	}
	return ReverseRun{}, fmt.Errorf("durable reverse run %q was not found", runID)
}

func newStateStore(path string) statestore.JSONStore[state] {
	return statestore.JSONStore[state]{
		Path: path,
		Initial: func() state {
			return state{
				CredentialBindings: map[string]credentialBindingState{},
				Checkpoints:        map[string]map[string]string{},
				StreamStates:       map[string]StreamState{},
			}
		},
		Locker: statestore.FileLock{Path: path + ".lock"},
	}
}

func (a *App) AddCredential(ctx context.Context, req AddCredentialRequest) (CredentialMeta, error) {
	if a.ephemeralCredentials != nil {
		return CredentialMeta{}, errors.New("certification ephemeral sessions do not persist credentials")
	}
	if strings.TrimSpace(req.Name) == "" {
		return CredentialMeta{}, errors.New("credential name is required")
	}
	if err := connectors.RejectLegacyConnectorName(req.Connector); err != nil {
		return CredentialMeta{}, err
	}
	connector, ok := a.registry.Get(req.Connector)
	if !ok {
		return CredentialMeta{}, fmt.Errorf("connector %q not found", req.Connector)
	}
	if _, ok := a.findCredential(req.Name); ok {
		return CredentialMeta{}, fmt.Errorf("credential %q already exists", req.Name)
	}
	providerFamily, authProfile, err := credentialCoordinationDeclarations(req.Connector, req.ProviderFamily, req.AuthProfile)
	if err != nil {
		return CredentialMeta{}, err
	}
	bindingID, err := prefixedID("cbind")
	if err != nil {
		return CredentialMeta{}, err
	}
	binding := credentialBindingState{
		BindingID:                     bindingID,
		ProviderFamilyDeclared:        strings.TrimSpace(req.ProviderFamily) != "",
		AuthProfileDeclared:           strings.TrimSpace(req.AuthProfile) != "",
		DeclarationProvenanceRecorded: true,
	}
	if strings.TrimSpace(req.LinkCredential) != "" {
		linked, ok := a.findCredential(req.LinkCredential)
		if !ok {
			return CredentialMeta{}, errors.New("link credential not found")
		}
		linkedBinding, err := a.credentialBindingForCredential(linked)
		if err != nil {
			return CredentialMeta{}, errors.New("linked credential coordination metadata is unavailable")
		}
		if err := a.validateCredentialLinkCohort(
			CredentialMeta{Connector: req.Connector, ProviderFamily: providerFamily, AuthProfile: authProfile},
			binding,
			linkedBinding,
		); err != nil {
			return CredentialMeta{}, err
		}
		binding.BindingID = linkedBinding.BindingID
	}
	if _, err := a.newCoordinationIdentity(providerFamily, authProfile, binding.BindingID); err != nil {
		return CredentialMeta{}, err
	}
	id, err := prefixedID("cred")
	if err != nil {
		return CredentialMeta{}, err
	}
	if req.Config == nil {
		req.Config = map[string]string{}
	}
	if req.Secrets == nil {
		req.Secrets = map[string]string{}
	}
	if err := a.validateCredentialConfig(req.Connector, req.Config); err != nil {
		return CredentialMeta{}, err
	}
	if err := connectors.ValidateConfiguration(connector, req.Config); err != nil {
		return CredentialMeta{}, fmt.Errorf("credential configuration: %w", err)
	}
	if err := credential.RequirePersistentValues(req.Secrets); err != nil {
		return CredentialMeta{}, err
	}
	if err := a.vault.Put(ctx, id, req.Secrets); err != nil {
		return CredentialMeta{}, err
	}
	now := time.Now().UTC()
	fields := make([]string, 0, len(req.Secrets))
	for k := range req.Secrets {
		fields = append(fields, k)
	}
	sort.Strings(fields)
	meta := CredentialMeta{
		ID:             id,
		Name:           req.Name,
		Connector:      req.Connector,
		ProviderFamily: providerFamily,
		AuthProfile:    authProfile,
		Config:         cloneStringMap(req.Config),
		SecretFields:   fields,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
	if a.state.CredentialBindings == nil {
		a.state.CredentialBindings = map[string]credentialBindingState{}
	}
	a.state.Credentials = append(a.state.Credentials, meta)
	a.state.CredentialBindings[id] = binding
	if err := a.save(); err != nil {
		return CredentialMeta{}, err
	}
	return meta, nil
}

// LinkCredential explicitly joins one credential to another credential's
// protected non-secret binding. It never reads the vault and refuses to bridge
// incompatible declared provider families or auth profiles.
func (a *App) LinkCredential(name, target string) (CredentialMeta, error) {
	sourceIndex := a.credentialIndex(name)
	targetIndex := a.credentialIndex(target)
	if sourceIndex < 0 || targetIndex < 0 {
		return CredentialMeta{}, errors.New("credential link target not found")
	}
	if sourceIndex == targetIndex {
		return CredentialMeta{}, errors.New("credential cannot link to itself")
	}
	source := a.state.Credentials[sourceIndex]
	targetCredential := a.state.Credentials[targetIndex]
	sourceBinding, err := a.credentialBindingForCredential(source)
	if err != nil {
		return CredentialMeta{}, err
	}
	targetBinding, err := a.credentialBindingForCredential(targetCredential)
	if err != nil {
		return CredentialMeta{}, errors.New("credential link target coordination metadata is unavailable")
	}
	if err := a.validateCredentialLinkCohort(source, sourceBinding, targetBinding); err != nil {
		return CredentialMeta{}, err
	}
	if _, err := a.newCoordinationIdentity(source.ProviderFamily, source.AuthProfile, targetBinding.BindingID); err != nil {
		return CredentialMeta{}, err
	}
	sourceBinding.BindingID = targetBinding.BindingID
	a.state.CredentialBindings[source.ID] = sourceBinding
	a.state.Credentials[sourceIndex].UpdatedAt = time.Now().UTC()
	if err := a.save(); err != nil {
		return CredentialMeta{}, err
	}
	linked, ok := a.findCredential(name)
	if !ok {
		return CredentialMeta{}, errors.New("linked credential not found")
	}
	return linked, nil
}

func credentialCoordinationDeclarations(connector, providerFamily, authProfile string) (string, string, error) {
	if strings.TrimSpace(providerFamily) == "" {
		providerFamily = connector
	}
	if strings.TrimSpace(authProfile) == "" {
		authProfile = "default"
	}
	if _, err := connectors.NewCoordinationIdentity([]byte("credential-coordination-validation"), connectors.CredentialBinding{
		BindingID:      "binding-validation",
		ProviderFamily: providerFamily,
		AuthProfile:    authProfile,
	}); err != nil {
		return "", "", &CredentialCoordinationDeclarationError{err: err}
	}
	return providerFamily, authProfile, nil
}

type credentialCohortMember struct {
	credential CredentialMeta
	binding    credentialBindingState
}

func (a *App) validateCredentialLinkCohort(source CredentialMeta, sourceBinding, targetBinding credentialBindingState) error {
	members := make([]credentialCohortMember, 0, len(a.state.Credentials)+1)
	sourceIncluded := false
	for _, credential := range a.state.Credentials {
		binding, err := a.credentialBindingForCredential(credential)
		if err != nil {
			return err
		}
		if binding.BindingID != targetBinding.BindingID {
			continue
		}
		members = append(members, credentialCohortMember{credential: credential, binding: binding})
		if credential.ID == source.ID {
			sourceIncluded = true
		}
	}
	if !sourceIncluded {
		members = append(members, credentialCohortMember{credential: source, binding: sourceBinding})
	}
	for left := 0; left < len(members); left++ {
		for right := left + 1; right < len(members); right++ {
			if err := validateCredentialLinkPair(members[left], members[right]); err != nil {
				return err
			}
		}
	}
	return nil
}

func validateCredentialLinkPair(source, target credentialCohortMember) error {
	if source.credential.Connector != target.credential.Connector && (!source.binding.hasExplicitDeclarations() || !target.binding.hasExplicitDeclarations()) {
		return &CredentialLinkValidationError{err: errors.New("cross-connector credential link requires explicitly declared provider family and auth profile")}
	}
	if source.credential.ProviderFamily != target.credential.ProviderFamily {
		return &CredentialLinkValidationError{err: errors.New("credential link requires matching provider family")}
	}
	if source.credential.AuthProfile != target.credential.AuthProfile {
		return &CredentialLinkValidationError{err: errors.New("credential link requires matching auth profile")}
	}
	return nil
}

func (a *App) credentialIndex(name string) int {
	for index, credential := range a.state.Credentials {
		if credential.Name == name || credential.ID == name {
			return index
		}
	}
	return -1
}

func (a *App) ListCredentials() []CredentialMeta {
	out := append([]CredentialMeta(nil), a.state.Credentials...)
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out
}

func (a *App) validateCredentialConfig(connector string, config map[string]string) error {
	path := config["path"]
	if path == "" {
		return nil
	}
	switch connector {
	case "warehouse", "outbox":
		allowExternal := strings.EqualFold(config["allow_external_path"], "true")
		return safety.ValidateLocalWritePath(a.projectRoot(), path, connector+" path", allowExternal)
	default:
		return safety.RejectDangerousChars(path, connector+" path")
	}
}

func (a *App) InspectCredential(name string) (CredentialMeta, error) {
	cred, ok := a.findCredential(name)
	if !ok {
		return CredentialMeta{}, fmt.Errorf("credential %q not found", name)
	}
	return cred, nil
}

func (a *App) TestCredential(ctx context.Context, name string) (CredentialMeta, error) {
	repairCtx := context.WithValue(ctx, authRepairContextKey{}, true)
	cred, runtime, err := a.resolveCredential(repairCtx, name, nil)
	if err != nil {
		return CredentialMeta{}, err
	}
	connector, ok := a.registry.Get(cred.Connector)
	if !ok {
		return CredentialMeta{}, fmt.Errorf("connector %q not found", cred.Connector)
	}
	cohort := runtime.CoordinationIdentity.AuthCohortKey()
	// A dedicated test must reach the provider even while ordinary work is
	// fenced; it is the only production repair path.
	runtime.AuthenticationAdmission = nil
	if err := connector.Check(ctx, runtime); err != nil {
		if connectors.IsVerifiedAuthenticationFailure(err) {
			if fenceErr := a.authCohorts.Fence(cohort, coordination.AuthenticationOutcomeVerifiedInvalid); fenceErr != nil {
				return CredentialMeta{}, fenceErr
			}
		}
		return CredentialMeta{}, err
	}
	if _, err := a.authCohorts.Repair(cohort, coordination.AuthenticationOutcomeVerifiedHealthy); err != nil {
		return CredentialMeta{}, err
	}
	for i := range a.state.Credentials {
		if a.state.Credentials[i].Name == name {
			a.state.Credentials[i].LastValidatedAt = time.Now().UTC()
			cred = a.state.Credentials[i]
			break
		}
	}
	if err := a.save(); err != nil {
		return CredentialMeta{}, err
	}
	return cred, nil
}

func (a *App) RemoveCredential(ctx context.Context, name string) error {
	for i, cred := range a.state.Credentials {
		if cred.Name == name {
			if err := a.vault.Delete(ctx, cred.ID); err != nil {
				return err
			}
			a.state.Credentials = append(a.state.Credentials[:i], a.state.Credentials[i+1:]...)
			delete(a.state.CredentialBindings, cred.ID)
			return a.save()
		}
	}
	return fmt.Errorf("credential %q not found", name)
}

// ValidateConnectionName rejects connection names that two distinct
// connections could not be told apart by once folded into a path. See
// warehouse.ValidateConnectionName for the rule and why it rejects rather than
// rewrites.
func ValidateConnectionName(name string) error {
	return warehouse.ValidateConnectionName(name)
}

// connectionMaterializesLocalWarehouse is deliberately credential-free: this
// admission check must inspect the whole configured inventory before a run
// begins, without opening a vault entry or contacting a provider. A persisted
// endpoint normally carries its connector name; the credential metadata is a
// legacy fallback only and never exposes its value.
func (a *App) connectionMaterializesLocalWarehouse(connection Connection) bool {
	connectorName := connection.Destination.Connector
	if connectorName == "" {
		if credential, ok := a.findCredential(connection.Destination.Credential); ok {
			connectorName = credential.Connector
		}
	}
	destination, ok := a.registry.Get(connectorName)
	if !ok {
		return false
	}
	materializer, ok := destination.(connectors.LocalWarehouseMaterializer)
	return ok && materializer.MaterializesLocalWarehouse()
}

func validateSameOwnerCaseEquivalentDestinationTables(connections []Connection) error {
	collisions := sameOwnerCaseEquivalentDestinationCollisions(connections)
	if len(collisions) == 0 {
		return nil
	}
	return collisions[0].err()
}

// validateConfiguredLocalWarehouseDestinationTables checks every configured
// local-warehouse connection as one immutable inventory. It is called before
// beginRun, so a persisted legacy collision cannot create a run, checkpoint,
// owner directory, WAL, temporary file, or Parquet output.
func (a *App) configuredLocalWarehouseDestinationCollisions() []warehouseDestinationCollision {
	connections := make([]Connection, 0, len(a.state.Connections))
	for _, connection := range a.state.Connections {
		if a.connectionMaterializesLocalWarehouse(connection) {
			connections = append(connections, connection)
		}
	}
	return sameOwnerCaseEquivalentDestinationCollisions(connections)
}

func (a *App) validateConfiguredLocalWarehouseDestinationTables() error {
	collisions := a.configuredLocalWarehouseDestinationCollisions()
	if len(collisions) == 0 {
		return nil
	}
	return collisions[0].err()
}

func (a *App) CreateConnection(ctx context.Context, req CreateConnectionRequest) (Connection, error) {
	req = cloneCreateConnectionRequest(req)
	if err := ValidateConnectionName(req.Name); err != nil {
		return Connection{}, err
	}
	if _, ok := a.findConnection(req.Name); ok {
		return Connection{}, fmt.Errorf("connection %q already exists", req.Name)
	}
	if existing, ok := a.findConnectionFold(req.Name); ok {
		return Connection{}, fmt.Errorf("connection %q is ambiguous with existing connection %q: connection names must differ by more than letter case", req.Name, existing.Name)
	}
	if len(req.Streams) == 0 {
		return Connection{}, errors.New("at least one stream is required")
	}
	source, sourceRuntime, err := a.resolveEndpoint(ctx, req.Source)
	if err != nil {
		return Connection{}, fmt.Errorf("resolve source: %w", err)
	}
	destination, _, err := a.resolveEndpoint(ctx, req.Destination)
	if err != nil {
		return Connection{}, fmt.Errorf("resolve destination: %w", err)
	}
	for name, stream := range req.Streams {
		modeName := stream.SyncMode
		if modeName == "" {
			modeName = DefaultUserFacingSyncMode
		}
		mode, modeErr := ParseStreamSyncMode(StreamConfig{SyncMode: modeName, LegacyCompatibility: stream.LegacyCompatibility})
		if modeErr != nil {
			return Connection{}, modeErr
		}
		destinationDescriptor, declared := connectors.DestinationTransportDescriptorOf(destination)
		if stream.DestinationAction != "" && (!declared || destinationDescriptor.Executor != declarativeTypedDestinationReference) {
			return Connection{}, fmt.Errorf("stream %q selects destination_action but destination connector %q is not a declarative typed destination", name, destination.Name())
		}
		if declared && destinationDescriptor.Executor == declarativeTypedDestinationReference {
			if a.transports == nil {
				return Connection{}, fmt.Errorf("declarative typed destination transport registry is unavailable")
			}
			resolved, preflightErr := a.transports.Preflight(synctransport.PreflightRequest{Source: source, Destination: destination, Stream: name, Mode: mode.ContractMode, DestinationAction: stream.DestinationAction})
			if preflightErr != nil {
				return Connection{}, fmt.Errorf("validate persisted destination action for stream %q: %w", name, preflightErr)
			}
			if selectionErr := validateDeclarativeTypedDestinationSelection(source, destination, name, mode.ContractMode, resolved.ApplyStrategy); selectionErr != nil {
				return Connection{}, fmt.Errorf("validate persisted destination mapping for stream %q: %w", name, selectionErr)
			}
		}
	}
	materializesWarehouse := false
	if materializer, ok := destination.(connectors.LocalWarehouseMaterializer); ok {
		materializesWarehouse = materializer.MaterializesLocalWarehouse()
	}
	catalog, catalogErr := a.catalogForEndpoint(ctx, source, sourceRuntime, false)
	if errors.Is(catalogErr, errCatalogStale) {
		return Connection{}, catalogErr
	}
	streamIDs := make(map[string]struct{}, len(req.Streams))
	for name, stream := range req.Streams {
		if strings.TrimSpace(stream.StreamID) != "" {
			return Connection{}, errors.New("stream identity is assigned by the application")
		}
		if stream.SyncMode == "" {
			stream.SyncMode = DefaultUserFacingSyncMode
		}
		if isLegacySyncModeName(stream.SyncMode) && !postgresManagedTargetContractMode(source, destination, stream.SyncMode) {
			stream.LegacyCompatibility = true
		}
		mode, err := ParseStreamSyncMode(stream)
		if err != nil {
			return Connection{}, err
		}
		stream.SyncMode = mode.Name
		if mode.LegacyCompatibility {
			stream.LegacyCompatibility = true
		}
		if catalogErr == nil {
			if sourceStream, ok := findCatalogStream(catalog, name); ok {
				if stream.CursorField == "" && len(sourceStream.CursorFields) > 0 {
					stream.CursorField = sourceStream.CursorFields[0]
				}
				if len(stream.PrimaryKey) == 0 && len(sourceStream.PrimaryKey) > 0 {
					stream.PrimaryKey = append([]string(nil), sourceStream.PrimaryKey...)
				}
			}
		}
		if stream.DestinationTable == "" {
			stream.DestinationTable = name
		}
		if err := validateConnectionTransformPlan(ctx, source, sourceRuntime, name, &stream); err != nil {
			return Connection{}, fmt.Errorf("validate transform for stream %q: %w", name, err)
		}
		// Against the local warehouse the stream and table names are path
		// components, so they are held to the same rule as the connection
		// name: rejected at creation rather than coerced into something that
		// resolves. Creation is the only moment this can honestly be caught —
		// every sync of a connection carrying an unusable name fails at the
		// same place, and a connection can be neither edited nor deleted.
		if materializesWarehouse {
			if _, err := warehouse.PathComponent("stream", name); err != nil {
				return Connection{}, err
			}
			if _, err := warehouse.PathComponent("table", stream.DestinationTable); err != nil {
				return Connection{}, err
			}
		}
		if err := ValidateStreamSyncConfig(stream); err != nil {
			return Connection{}, fmt.Errorf("validate stream %q: %w", name, err)
		}
		streamID, err := allocateUniquePrefixedID("stream", streamIDs)
		if err != nil {
			return Connection{}, err
		}
		stream.StreamID = streamID
		req.Streams[name] = stream
	}
	targetCopyWorkers, _, err := resolveTargetCopyWorkers(destination, req.Streams, req.TargetCopyWorkers)
	if err != nil {
		return Connection{}, err
	}
	if materializesWarehouse {
		if err := validateSameOwnerCaseEquivalentDestinationTables([]Connection{{
			Name:    req.Name,
			Streams: req.Streams,
		}}); err != nil {
			return Connection{}, err
		}
	}
	connectionIDs := make(map[string]struct{}, len(a.state.Connections))
	for _, connection := range a.state.Connections {
		if strings.TrimSpace(connection.ID) != "" {
			connectionIDs[connection.ID] = struct{}{}
		}
	}
	connectionID, err := allocateUniquePrefixedID("conn", connectionIDs)
	if err != nil {
		return Connection{}, err
	}
	now := time.Now().UTC()
	conn := Connection{
		ID:                connectionID,
		Name:              req.Name,
		Source:            req.Source,
		Destination:       req.Destination,
		Streams:           req.Streams,
		TargetCopyWorkers: targetCopyWorkers,
		CreatedAt:         now,
		UpdatedAt:         now,
	}
	stored := cloneConnection(conn)
	a.state.Connections = append(a.state.Connections, stored)
	if err := a.save(); err != nil {
		return Connection{}, err
	}
	return cloneConnection(stored), nil
}

func (a *App) ListConnections() []Connection {
	out := make([]Connection, len(a.state.Connections))
	for index, connection := range a.state.Connections {
		out[index] = cloneConnection(connection)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out
}

func (a *App) RefreshCatalog(ctx context.Context, connectionName string) (CatalogSnapshot, error) {
	conn, ok := a.findConnection(connectionName)
	if !ok {
		return CatalogSnapshot{}, fmt.Errorf("connection %q not found", connectionName)
	}
	source, runtime, err := a.resolveEndpoint(ctx, conn.Source)
	if err != nil {
		return CatalogSnapshot{}, err
	}
	catalog, err := a.catalogForEndpoint(ctx, source, runtime, true)
	if err != nil {
		return CatalogSnapshot{}, err
	}
	return CatalogSnapshot{Connection: conn.Name, Catalog: catalog, UpdatedAt: time.Now().UTC()}, nil
}

var errCatalogStale = errors.New("cached catalog is stale; run pm catalog refresh --connection <name> before using this schema")

// catalogForEndpoint is the only application-level catalog cache boundary.
// It is intentionally endpoint-shaped rather than source-shaped: source
// refresh is implemented today, and a future dynamic destination can use the
// same connector/account store without a second persistence model. Dynamic
// catalogs are persisted after the file reaches durability; static catalogs
// retain their existing cheap in-memory behavior unless explicitly refreshed.
func (a *App) catalogForEndpoint(ctx context.Context, connector connectors.Connector, runtime connectors.RuntimeConfig, forceRefresh bool) (connectors.Catalog, error) {
	reference, err := a.catalogReference(connector.Name(), runtime)
	if err != nil {
		return connectors.Catalog{}, err
	}
	if !forceRefresh {
		if reference, ok := a.accountCatalogReference(reference); ok {
			stored, err := a.catalogs.read(reference)
			if err != nil {
				return connectors.Catalog{}, fmt.Errorf("read cached catalog: %w", err)
			}
			snapshot := catalogSnapshotWithStaleness(CatalogSnapshot{Catalog: stored.Catalog, UpdatedAt: stored.UpdatedAt}, time.Now().UTC())
			if snapshot.Catalog.Discovery != nil && snapshot.Catalog.Discovery.Stale {
				return connectors.Catalog{}, errCatalogStale
			}
			return snapshot.Catalog, nil
		}
	}
	runtime.ForceCatalogRefresh = forceRefresh
	catalog, err := connector.Catalog(ctx, runtime)
	if err != nil {
		return connectors.Catalog{}, err
	}
	if catalog.Discovery == nil && !forceRefresh {
		return catalog, nil
	}
	if err := a.persistAccountCatalog(reference, catalog, time.Now().UTC()); err != nil {
		return connectors.Catalog{}, err
	}
	return catalog, nil
}

// persistAccountCatalog is the catalog durability boundary. The file is fully
// synced before the state pointer can be written. A failed state commit can
// leave only an unreferenced durable file, never a state reference to a
// missing or unsynced catalog.
func (a *App) persistAccountCatalog(reference catalogReference, catalog connectors.Catalog, updatedAt time.Time) error {
	if err := a.catalogs.write(reference, catalog, updatedAt); err != nil {
		return err
	}
	return a.putCatalogReference(reference)
}

// putCatalogReference is the sole state.json write boundary for account
// catalogs. Connection callers derive this reference from their connector and
// opaque coordination identity, so multiple connections to one account share
// a single catalog file.
func (a *App) putCatalogReference(reference catalogReference) error {
	for i := range a.state.Catalogs {
		if a.state.Catalogs[i].Connector == reference.Connector && a.state.Catalogs[i].AccountKey == reference.AccountKey {
			if a.state.Catalogs[i] == reference {
				return nil
			}
			a.state.Catalogs[i] = reference
			return a.save()
		}
	}
	a.state.Catalogs = append(a.state.Catalogs, reference)
	return a.save()
}

func (a *App) ShowCatalog(ctx context.Context, connectionName string) (CatalogSnapshot, error) {
	conn, ok := a.findConnection(connectionName)
	if !ok {
		return CatalogSnapshot{}, fmt.Errorf("connection %q not found", connectionName)
	}
	reference, err := a.catalogReferenceForEndpoint(conn.Source)
	if err != nil {
		return CatalogSnapshot{}, err
	}
	if reference, ok := a.accountCatalogReference(reference); ok {
		stored, err := a.catalogs.read(reference)
		if err != nil {
			return CatalogSnapshot{}, fmt.Errorf("read cached catalog: %w", err)
		}
		return catalogSnapshotWithStaleness(CatalogSnapshot{Connection: conn.Name, Catalog: stored.Catalog, UpdatedAt: stored.UpdatedAt}, time.Now().UTC()), nil
	}
	return a.RefreshCatalog(ctx, connectionName)
}

// catalogReferenceForEndpoint derives the account partition without opening
// the vault. A schema-only cache can therefore be inspected even when a
// caller is not about to perform an authenticated provider request.
func (a *App) catalogReferenceForEndpoint(endpoint EndpointConfig) (catalogReference, error) {
	if err := connectors.RejectLegacyConnectorName(endpoint.Connector); err != nil {
		return catalogReference{}, err
	}
	credential, ok := a.findCredential(endpoint.Credential)
	if !ok {
		return catalogReference{}, errors.New("configured credential not found")
	}
	if endpoint.Connector != "" && endpoint.Connector != credential.Connector {
		return catalogReference{}, fmt.Errorf("connector configuration does not match its credential")
	}
	identity, err := a.coordinationIdentityForCredential(credential)
	if err != nil {
		return catalogReference{}, err
	}
	return a.catalogReference(credential.Connector, connectors.RuntimeConfig{CoordinationIdentity: identity})
}

func (a *App) accountCatalogReference(want catalogReference) (catalogReference, bool) {
	for _, reference := range a.state.Catalogs {
		if reference.Connector == want.Connector && reference.AccountKey == want.AccountKey {
			return reference, true
		}
	}
	return catalogReference{}, false
}

func catalogSnapshotWithStaleness(snapshot CatalogSnapshot, now time.Time) CatalogSnapshot {
	if snapshot.Catalog.Discovery == nil {
		return snapshot
	}
	status := *snapshot.Catalog.Discovery
	status.Failures = append([]connectors.DiscoveryFailure(nil), status.Failures...)
	status.Cached = true
	status.Stale = !status.ExpiresAt.IsZero() && !now.Before(status.ExpiresAt)
	snapshot.Catalog.Discovery = &status
	return snapshot
}

func (a *App) RunETL(ctx context.Context, req RunETLRequest) (Run, error) {
	if req.MaxInFlightBatches < 0 || req.MaxInFlightBatches > 8 {
		return Run{}, fmt.Errorf("max in-flight batches must be between 1 and 8")
	}
	conn, ok := a.findConnection(req.Connection)
	if !ok {
		return Run{}, fmt.Errorf("connection %q not found", req.Connection)
	}
	stream, ok := conn.Streams[req.Stream]
	if !ok {
		return Run{}, fmt.Errorf("stream %q not configured on connection %q", req.Stream, req.Connection)
	}
	if delivered, pending := a.deliveredReconciliationFor(req.Connection, req.Stream); pending {
		return a.reconcileDeliveredTransportRun(ctx, delivered)
	}
	if a.connectionMaterializesLocalWarehouse(conn) {
		if err := a.validateConfiguredLocalWarehouseDestinationTables(); err != nil {
			return Run{}, err
		}
	}
	batchSize := req.BatchSize
	if batchSize <= 0 {
		batchSize = 1000
	}
	var rateParkingResumeCheckpoint *synccontract.CheckpointEnvelope
	if req.rateParkingResumeCheckpoint != nil {
		checkpoint := req.rateParkingResumeCheckpoint.Clone()
		rateParkingResumeCheckpoint = &checkpoint
	}
	runID := req.rateParkingRearmAttemptRunID
	if runID == "" {
		var err error
		runID, err = prefixedID("run")
		if err != nil {
			return Run{}, err
		}
	} else {
		parkedRunID, resuming := rateParkingResumeRunID(ctx)
		if !resuming || parkedRunID == runID {
			return Run{}, errors.New("rate-limit rearm attempt is not linked to a parked run")
		}
		parkedRun, found := a.runByID(parkedRunID)
		if !found || parkedRun.RateParkingRearmAttemptRunID != runID {
			return Run{}, errors.New("rate-limit rearm attempt link is unavailable")
		}
		if _, found := a.runByID(runID); found {
			return Run{}, errors.New("rate-limit rearm attempt run already exists")
		}
	}
	run := Run{ID: runID, Type: "etl", Connection: req.Connection, Stream: req.Stream, Status: "running", BatchSize: batchSize, StartedAt: time.Now().UTC()}
	if _, err := a.beginRun(run); err != nil {
		return Run{}, fmt.Errorf("start ETL run: %w", err)
	}

	mode, err := ParseStreamSyncMode(stream)
	if err != nil {
		return a.failRun(runID, err)
	}
	stream.SyncMode = mode.Name
	if err := ValidateStreamSyncConfig(stream); err != nil {
		return a.failRun(runID, err)
	}
	source, sourceCredential, sourceRuntime, err := a.resolveEndpointWithCredential(ctx, conn.Source)
	if err != nil {
		return a.failRun(runID, err)
	}
	destination, destRuntime, err := a.resolveEndpoint(ctx, conn.Destination)
	if err != nil {
		return a.failRun(runID, err)
	}
	destinationDescriptor, declared := connectors.DestinationTransportDescriptorOf(destination)
	if stream.DestinationAction != "" && (!declared || destinationDescriptor.Executor != declarativeTypedDestinationReference) {
		return a.failRun(runID, fmt.Errorf("stream %q selects destination_action but destination connector %q is not a declarative typed destination", req.Stream, destination.Name()))
	}
	if declared && destinationDescriptor.Executor == declarativeTypedDestinationReference {
		if a.transports == nil {
			return a.failRun(runID, fmt.Errorf("declarative typed destination transport registry is unavailable"))
		}
		resolved, err := a.transports.Preflight(synctransport.PreflightRequest{
			Source: source, Destination: destination, Stream: req.Stream, Mode: mode.ContractMode, DestinationAction: stream.DestinationAction,
		})
		if err != nil {
			return a.failRun(runID, fmt.Errorf("validate persisted destination action for stream %q: %w", req.Stream, err))
		}
		if err := validateDeclarativeTypedDestinationSelection(source, destination, req.Stream, mode.ContractMode, resolved.ApplyStrategy); err != nil {
			return a.failRun(runID, fmt.Errorf("validate persisted destination mapping for stream %q: %w", req.Stream, err))
		}
	}
	sourceExpectation := streamResumeExpectation(source, sourceCredential, sourceRuntime, req.Stream)
	dispatchRequest := etlModeDispatchRequest{
		runID:                       runID,
		connection:                  conn,
		source:                      source,
		sourceRuntime:               sourceRuntime,
		destination:                 destination,
		destinationRuntime:          destRuntime,
		sourceExpectation:           sourceExpectation,
		streamName:                  req.Stream,
		stream:                      stream,
		mode:                        mode,
		batchSize:                   batchSize,
		maxInFlightBatches:          req.MaxInFlightBatches,
		destinationApproval:         req.DestinationApproval,
		rateParkingResumeCheckpoint: rateParkingResumeCheckpoint,
	}
	return a.dispatchETLMode(ctx, dispatchRequest)
}

func hasDestinationApproval(approval synctransport.DestinationApproval) bool {
	return approval.PlanID != "" || approval.ApprovalToken != "" || approval.Confirmation.Kind != "" || approval.Evidence != nil
}

func (a *App) runConnectorETL(ctx context.Context, runID string, conn Connection, source connectors.Connector, sourceRuntime connectors.RuntimeConfig, destination connectors.Connector, destRuntime connectors.RuntimeConfig, sourceExpectation synccontract.ResumeExpectation, streamName string, stream StreamConfig, mode SyncMode, batchSize int) (etlExecutionResult, error) {
	if mode.IsDeduped() {
		return etlExecutionResult{}, fmt.Errorf("sync mode %s requires the local warehouse destination in this dependency-free implementation", mode.Name)
	}
	durableDestination, ok := destination.(synccontract.DurableETLDestination)
	if !ok {
		return etlExecutionResult{}, &synccontract.DestinationDurabilityAdmissionError{Destination: destination.Name()}
	}
	stateKey := streamStateKey(conn.Name, streamName)
	prior := a.state.StreamStates[stateKey]
	if prior.Checkpoint != nil {
		if err := validateStreamStateResume(prior, sourceExpectation); err != nil {
			return etlExecutionResult{}, err
		}
	}
	generationID := prior.GenerationID
	if generationID == 0 || mode.IsOverwrite() {
		generationID++
	}
	now := time.Now().UTC().Format(time.RFC3339Nano)
	result := etlExecutionResult{}
	batch := make([]connectors.Record, 0, batchSize)
	firstWrite := true
	cursorTracker, err := newStreamCursorTracker(prior, source, sourceRuntime, stream.CursorField, mode.Source)
	if err != nil {
		return etlExecutionResult{}, err
	}
	observedAt := time.Time{}

	flush := func(force bool) error {
		if len(batch) == 0 {
			if force && mode.IsOverwrite() && firstWrite {
				writeResult, err := destination.Write(ctx, connectors.WriteRequest{
					Stream:     streamName,
					Table:      stream.DestinationTable,
					Action:     "upsert",
					Overwrite:  true,
					Config:     destRuntime,
					PrimaryKey: stream.PrimaryKey,
				}, nil)
				firstWrite = false
				if err != nil {
					return err
				}
				return validateCompleteETLBatchWrite(writeResult, 0)
			}
			return nil
		}
		writeResult, err := destination.Write(ctx, connectors.WriteRequest{
			Stream:     streamName,
			Table:      stream.DestinationTable,
			Action:     "upsert",
			Overwrite:  mode.IsOverwrite() && firstWrite,
			Config:     destRuntime,
			PrimaryKey: stream.PrimaryKey,
		}, batch)
		firstWrite = false
		if err != nil {
			return err
		}
		if err := validateCompleteETLBatchWrite(writeResult, len(batch)); err != nil {
			return err
		}
		result.RecordsLoaded += writeResult.RecordsWritten
		result.RecordsFailed += writeResult.RecordsFailed
		result.BatchCount++
		batch = batch[:0]
		return nil
	}

	err = source.Read(ctx, cursorTracker.readRequest(streamName, sourceRuntime, prior, generationID, mode.Source), func(record connectors.Record) error {
		result.RecordsRead++
		cursor := ""
		if stream.CursorField != "" {
			var include bool
			var err error
			cursor, _, include, err = cursorTracker.observe(record, stream.CursorField, mode.Source)
			if err != nil {
				return err
			}
			if !include {
				return nil
			}
		}
		r := cloneRecord(record)
		r["_polymetrics_run_id"] = runID
		r["_polymetrics_synced_at"] = now
		r["_polymetrics_deleted"] = isDeletedRecord(record)
		if stream.CursorField != "" {
			r["_polymetrics_cursor"] = cursor
		}
		result.RecordsTransformed++
		observedAt = time.Now().UTC()
		batch = append(batch, r)
		if len(batch) >= batchSize {
			return flush(false)
		}
		return nil
	})
	if err != nil {
		return result, err
	}
	if err := flush(true); err != nil {
		return result, err
	}
	if observedAt.IsZero() {
		observedAt = time.Now().UTC()
	}
	acknowledgement, err := durableDestination.AcknowledgeETLDurability(ctx, runID)
	if err != nil {
		return result, err
	}
	if acknowledgement.Sink != destination.Name() {
		return result, fmt.Errorf("durable downstream acknowledgement sink %q does not match destination %q", acknowledgement.Sink, destination.Name())
	}
	nextCursor, nextCursorObserved := cursorTracker.checkpoint()
	updated, err := committedLegacyStreamState(conn, sourceExpectation, streamName, stream, runID, nextCursor, nextCursorObserved, generationID, result.RecordsLoaded, observedAt, acknowledgement)
	if err != nil {
		return result, err
	}
	reportedCursor, reportedCursorObserved := cursorTracker.reportedCursor()
	result.Checkpoint = checkpointForResult(result, mode, stateKey, updated, reportedCursor, reportedCursorObserved)
	result.PendingStreamState = &pendingStreamState{Key: stateKey, State: updated}
	return result, nil
}

func validateCompleteETLBatchWrite(result connectors.WriteResult, batchSize int) error {
	if result.RecordsWritten != batchSize || result.RecordsFailed != 0 {
		return fmt.Errorf("destination write reported %d records written and %d failed for batch of %d", result.RecordsWritten, result.RecordsFailed, batchSize)
	}
	return nil
}

func (a *App) beginRun(run Run) (Run, error) {
	previousRuns := a.state.Runs
	a.state.Runs = append(append([]Run(nil), a.state.Runs...), run)
	if err := a.save(); err != nil {
		if !errors.Is(err, errStateRevisionConflict) && !stateStoreCommitMayHaveSucceeded(err) {
			a.state.Runs = previousRuns
		}
		return Run{}, err
	}
	return run, nil
}

type acknowledgedTransportCompletion struct {
	key   string
	state StreamState
}

// completeAcknowledgedTransportRun carries the state this App persisted for
// the closed transport checkpoint. It is an eligibility witness, not a refresh:
// a stale finalization may rebase only while the latest stream state still
// matches it and the target run remains running.
func (a *App) completeAcknowledgedTransportRun(runID string, result etlExecutionResult) (Run, error) {
	if result.PendingStreamState == nil {
		return a.completeRun(runID, result)
	}
	acknowledged, present := a.state.StreamStates[result.PendingStreamState.Key]
	if !present {
		return a.completeRun(runID, result)
	}
	return a.completeRunWithAcknowledgedTransportState(runID, result, &acknowledgedTransportCompletion{
		key:   result.PendingStreamState.Key,
		state: cloneStreamState(acknowledged),
	})
}

// failAcknowledgedTransportRun uses a completed checkpoint only as an
// eligibility witness for terminal failure. It does not refresh state, retry a
// checkpoint, or replay destination work.
func (a *App) failAcknowledgedTransportRun(runID string, result etlExecutionResult, runErr error) (Run, error) {
	if errors.Is(runErr, errTransportStreamStateConflict) {
		return a.failRunWithResult(runID, result, runErr)
	}
	if result.PendingStreamState == nil {
		return a.failRunWithResult(runID, result, runErr)
	}
	acknowledged, present := a.state.StreamStates[result.PendingStreamState.Key]
	if !present {
		return a.failRunWithResult(runID, result, runErr)
	}
	if result.DeliveryReconciliation != nil {
		return a.persistDeliveredReconciliationRun(runID, result, runErr, &acknowledgedTransportCompletion{
			key:   result.PendingStreamState.Key,
			state: cloneStreamState(acknowledged),
		})
	}
	return a.failRunWithAcknowledgedTransportState(runID, result, runErr, &acknowledgedTransportCompletion{
		key:   result.PendingStreamState.Key,
		state: cloneStreamState(acknowledged),
	})
}

func (a *App) completeRun(runID string, result etlExecutionResult) (Run, error) {
	return a.completeRunWithAcknowledgedTransportState(runID, result, nil)
}

func (a *App) completeRunWithAcknowledgedTransportState(runID string, result etlExecutionResult, acknowledged *acknowledgedTransportCompletion) (Run, error) {
	if result.PendingStreamState == nil {
		return Run{}, errors.New("completed ETL run is missing pending stream state")
	}
	expectedRevision := a.state.Revision
	completedAt := time.Now().UTC()
	transitionedInCallback := false
	updated, persistErr := a.updateState(func(current state) (state, error) {
		rebased := current.Revision != expectedRevision
		if rebased && acknowledged == nil {
			return current, errStateRevisionConflict
		}
		if rebased {
			currentStreamState, present := current.StreamStates[acknowledged.key]
			if !present || !transportStreamStateEqual(currentStreamState, acknowledged.state) {
				return current, fmt.Errorf("acknowledged transport stream state changed before completion: %w", errStateRevisionConflict)
			}
		}
		found := false
		for i := range current.Runs {
			if current.Runs[i].ID != runID {
				continue
			}
			if rebased && current.Runs[i].Status != "running" {
				return current, fmt.Errorf("acknowledged transport run %q has status %q, want running before completion: %w", runID, current.Runs[i].Status, errStateRevisionConflict)
			}
			current.Runs[i].Status = "completed"
			current.Runs[i].RecordsRead = result.RecordsRead
			current.Runs[i].RecordsTransformed = result.RecordsTransformed
			current.Runs[i].RecordsLoaded = result.RecordsLoaded
			current.Runs[i].RecordsFailed = result.RecordsFailed
			current.Runs[i].BatchCount = result.BatchCount
			current.Runs[i].Checkpoint = cloneStringMap(result.Checkpoint)
			current.Runs[i].TransportPhaseMeasurement = cloneTransportPhaseMeasurement(result.TransportPhaseMeasurement)
			current.Runs[i].DestinationResults = cloneDestinationResults(result.DestinationResults)
			current.Runs[i].CompletedAt = completedAt
			found = true
			transitionedInCallback = true
			break
		}
		if !found {
			if rebased && acknowledged != nil {
				return current, fmt.Errorf("acknowledged transport run %q not found before completion: %w", runID, errStateRevisionConflict)
			}
			return current, fmt.Errorf("run %q not found", runID)
		}
		if current.Checkpoints == nil {
			current.Checkpoints = map[string]map[string]string{}
		}
		current.Checkpoints[runID] = cloneStringMap(result.Checkpoint)
		if current.StreamStates == nil {
			current.StreamStates = map[string]StreamState{}
		}
		current.StreamStates[result.PendingStreamState.Key] = cloneStreamState(result.PendingStreamState.State)
		return current, nil
	})
	if persistErr != nil {
		if acknowledged != nil && transitionedInCallback && stateStoreCommitMayHaveSucceeded(persistErr) {
			reloaded, reloadErr := a.reloadExactTerminalState(persistErr)
			if reloadErr == nil {
				run, runErr := terminalETLRunFromState(reloaded.State, runID)
				if runErr == nil {
					return run, persistErr
				}
				return Run{}, errors.Join(persistErr, runErr)
			}
			return Run{}, errors.Join(persistErr, reloadErr)
		}
		return Run{}, persistErr
	}
	for _, run := range updated.Runs {
		if run.ID == runID {
			return run, nil
		}
	}
	return Run{}, fmt.Errorf("completed run %q was not stored", runID)
}

func (a *App) failRunWithAcknowledgedTransportState(runID string, result etlExecutionResult, runErr error, acknowledged *acknowledgedTransportCompletion) (Run, error) {
	expectedRevision := a.state.Revision
	completedAt := time.Now().UTC()
	transitionedInCallback := false
	updated, persistErr := a.updateState(func(current state) (state, error) {
		rebased := current.Revision != expectedRevision
		if rebased {
			currentStreamState, present := current.StreamStates[acknowledged.key]
			if !present || !transportStreamStateEqual(currentStreamState, acknowledged.state) {
				return current, fmt.Errorf("acknowledged transport stream state changed before failure: %w", errStateRevisionConflict)
			}
		}
		for i := range current.Runs {
			if current.Runs[i].ID != runID {
				continue
			}
			if current.Runs[i].Status != "running" {
				return current, fmt.Errorf("acknowledged transport run %q has status %q, want running before failure: %w", runID, current.Runs[i].Status, errStateRevisionConflict)
			}
			current.Runs[i].Status = "failed"
			current.Runs[i].RecordsRead = result.RecordsRead
			current.Runs[i].RecordsTransformed = result.RecordsTransformed
			current.Runs[i].RecordsLoaded = result.RecordsLoaded
			current.Runs[i].RecordsFailed = result.RecordsFailed
			current.Runs[i].BatchCount = result.BatchCount
			current.Runs[i].Checkpoint = cloneStringMap(result.Checkpoint)
			current.Runs[i].TransportPhaseMeasurement = cloneTransportPhaseMeasurement(result.TransportPhaseMeasurement)
			current.Runs[i].DestinationResults = cloneDestinationResults(result.DestinationResults)
			current.Runs[i].Error = safety.RedactErrorText(runErr.Error())
			current.Runs[i].CompletedAt = completedAt
			transitionedInCallback = true
			return current, nil
		}
		return current, fmt.Errorf("acknowledged transport run %q not found before failure: %w", runID, errStateRevisionConflict)
	})
	if persistErr != nil {
		if transitionedInCallback && stateStoreCommitMayHaveSucceeded(persistErr) {
			reloaded, reloadErr := a.reloadExactTerminalState(persistErr)
			if reloadErr == nil {
				durableRun, terminalErr := terminalETLRunFromState(reloaded.State, runID)
				if terminalErr == nil {
					return durableRun, errors.Join(runErr, fmt.Errorf("persist failed ETL run: %w", persistErr))
				}
				return Run{}, errors.Join(runErr, fmt.Errorf("persist failed ETL run: %w", terminalErr))
			}
			return Run{}, errors.Join(runErr, fmt.Errorf("persist failed ETL run: %w", errors.Join(persistErr, reloadErr)))
		}
		return Run{}, errors.Join(runErr, fmt.Errorf("persist failed ETL run: %w", persistErr))
	}
	for _, run := range updated.Runs {
		if run.ID == runID {
			return run, runErr
		}
	}
	return Run{}, errors.Join(runErr, fmt.Errorf("failed run %q was not stored", runID))
}

func (a *App) GetRun(id string) (Run, error) {
	for _, run := range a.state.Runs {
		if run.ID == id {
			return run, nil
		}
	}
	return Run{}, fmt.Errorf("run %q not found", id)
}

func (a *App) QueryTable(ctx context.Context, req QueryTableRequest) ([]connectors.Record, error) {
	if req.Limit <= 0 {
		req.Limit = 100
	}
	return a.readWarehouseTable(ctx, req.Table, req.Connection, req.Limit)
}

func (a *App) ReadActionSource(ctx context.Context, req ActionSourceReadRequest) ([]connectors.Record, error) {
	return a.readWarehouseTable(ctx, req.Table, req.Connection, 0)
}

func (a *App) readWarehouseTable(ctx context.Context, table, connection string, limit int) ([]connectors.Record, error) {
	if table == "" {
		return nil, errors.New("table is required")
	}
	cfg := connectors.RuntimeConfig{
		ProjectDir: a.projectDir,
		Config: map[string]string{
			"path": a.warehouseRoot(),
		},
	}
	if connection != "" {
		// The unattributed selector names the root-level tables no connection
		// owns, so it deliberately does not resolve through findConnection.
		if connection != warehouse.UnattributedConnection {
			if _, ok := a.findConnection(connection); !ok {
				return nil, fmt.Errorf("connection %q not found", connection)
			}
		}
		cfg.Config["connection"] = connection
	}
	warehouseConnector, ok := a.registry.Get("warehouse")
	if !ok {
		return nil, errors.New("warehouse connector not registered")
	}
	rows := make([]connectors.Record, 0)
	err := warehouseConnector.Read(ctx, connectors.ReadRequest{Stream: table, Config: cfg, Limit: limit}, connectors.LimitEmitter(limit, func(record connectors.Record) error {
		rows = append(rows, record)
		return nil
	}))
	if err := connectors.IgnoreReadLimit(err); err != nil {
		return nil, err
	}
	return rows, nil
}

func (a *App) QuerySQL(ctx context.Context, req QuerySQLRequest) ([]connectors.Record, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if req.Connection != "" && req.Connection != warehouse.UnattributedConnection {
		if _, ok := a.findConnection(req.Connection); !ok {
			return nil, fmt.Errorf("connection %q not found", req.Connection)
		}
	}
	req.sameOwnerCaseEquivalentDestinationCollisions = a.configuredLocalWarehouseDestinationCollisions()
	return a.sqlEngine.QuerySQL(ctx, req)
}

// warehouseRoot is this project's local warehouse root.
func (a *App) warehouseRoot() string {
	return filepath.Join(a.projectDir, "warehouse")
}

// pinReverseSourceConnection resolves which connection's table a reverse plan
// is built from, and returns a selector that keeps resolving to that same table
// afterwards. Pinning it at plan time is what stops an approved plan becoming
// unexecutable the day a second connection materializes the same table name:
// preview and run take no connection selector of their own, so without a pin
// they would fall back to an unscoped read and be refused as ambiguous.
func (a *App) pinReverseSourceConnection(table, connection string) (string, error) {
	found, err := warehouse.FindTable(a.warehouseRoot(), table, connection)
	if err != nil {
		return "", err
	}
	if found.Connection == "" {
		return warehouse.UnattributedConnection, nil
	}
	return found.Connection, nil
}

// reverseSourceRemedy states the recovery available to a stored plan whose
// source table has become ambiguous. Only plans created before the owning
// connection was pinned can reach this, and preview and run accept no
// connection selector, so re-planning is the honest answer rather than naming a
// flag those commands do not have.
func reverseSourceRemedy(plan ReversePlan) string {
	return fmt.Sprintf(
		"reverse plan %q records no source connection, so re-create it with `pm reverse plan %s --source-table %s --connection <name>` and approve the new plan",
		plan.ID, plan.Name, plan.SourceTable,
	)
}

// QueryEngineName reports which SQL engine backs QuerySQL. There is one, in
// every build: "duckdb".
func (a *App) QueryEngineName() string {
	return a.sqlEngine.Name()
}

func (a *App) PlanReverseETL(ctx context.Context, req PlanReverseETLRequest) (ReversePlan, error) {
	if req.Name == "" {
		return ReversePlan{}, errors.New("reverse plan name is required")
	}
	if req.Action == "" {
		req.Action = "upsert"
	}
	if len(req.Mappings) == 0 {
		return ReversePlan{}, errors.New("at least one field mapping is required")
	}
	if req.Limit <= 0 {
		req.Limit = 100000
	}
	// Refuse before any rows are read, any plan row is stored, or any approval
	// token is minted: a non-batchable action must leave nothing approvable
	// behind.
	if err := a.guardBatchableAction(req.DestinationConnector, req.Action, req.SourceTable); err != nil {
		return ReversePlan{}, err
	}
	sourceConnection, err := a.pinReverseSourceConnection(req.SourceTable, req.SourceConnection)
	if err != nil {
		return ReversePlan{}, warehouse.WithAmbiguityRemedy(err, "pass --connection to choose one")
	}
	records, err := a.QueryTable(ctx, QueryTableRequest{Table: req.SourceTable, Connection: sourceConnection, Limit: req.Limit})
	if err != nil {
		return ReversePlan{}, err
	}
	mapped := mapReverseRecords(records, req.Mappings)
	dest := EndpointConfig{Connector: req.DestinationConnector, Credential: req.DestinationCredential, Config: req.DestinationConfig}
	destination, runtime, err := a.resolveEndpoint(ctx, dest)
	if err != nil {
		return ReversePlan{}, fmt.Errorf("resolve reverse destination: %w", err)
	}
	if !destination.Metadata().Capabilities.Write {
		return ReversePlan{}, fmt.Errorf("connector %q does not support reverse ETL writes", destination.Name())
	}
	if validator, ok := destination.(connectors.WriteValidator); ok {
		if err := validator.ValidateWrite(ctx, connectors.WriteRequest{
			Stream: "records",
			Table:  req.Name,
			Action: req.Action,
			Config: runtime,
		}, mapped); err != nil {
			return ReversePlan{}, fmt.Errorf("validate reverse destination: %w", err)
		}
	}
	id, err := prefixedID("rplan")
	if err != nil {
		return ReversePlan{}, err
	}
	payloadIdentity, err := payloadIdentitiesForRecords(runtime.ProjectDir, mapped)
	if err != nil {
		return ReversePlan{}, err
	}
	planHash, err := reversePlanHash(req.Name, req.SourceTable, req.DestinationConnector, req.DestinationCredential, req.Action, req.DestinationConfig, req.Mappings, mapped, payloadIdentity)
	if err != nil {
		return ReversePlan{}, err
	}
	sampleCount := min(3, len(mapped))
	redactFields := reversePlanRedactFields(destination, req.Action)
	confirmation := a.confirmationPolicyForAction(req.DestinationConnector, req.Action)
	challenge := string(confirmation.Kind)
	created := time.Now().UTC()
	expires := created.Add(24 * time.Hour)
	var planSeal *connectors.WritePlanSeal
	if confirmation.Kind != "" {
		seal, err := a.approval.IssueWritePlanSeal(connectors.WritePlanSealRequest{
			PlanID: id, PlanHash: planHash, Connector: req.DestinationConnector, Operation: req.Action,
			CredentialRevision: runtime.CredentialRevision, ConfigurationDigest: runtime.ConfigurationDigest,
			Batchable: a.actionIsBatchable(req.DestinationConnector, req.Action), Scope: runtime.WriteApprovalScope,
			Confirmation: confirmation,
		})
		if err != nil {
			return ReversePlan{}, err
		}
		planSeal = &seal
		created = seal.IssuedAt
		expires = seal.ExpiresAt
	}
	plan := ReversePlan{
		ID:                    id,
		Name:                  req.Name,
		Status:                "planned",
		SourceTable:           req.SourceTable,
		SourceConnection:      sourceConnection,
		DestinationConnector:  req.DestinationConnector,
		DestinationCredential: req.DestinationCredential,
		DestinationConfig:     cloneStringMap(req.DestinationConfig),
		Action:                req.Action,
		Mappings:              cloneStringMap(req.Mappings),
		PayloadIdentity:       payloadIdentity,
		ConfirmationChallenge: challenge,
		ConfirmationPolicy:    confirmationFromChallenge(challenge),
		RedactFields:          redactFields,
		RecordCount:           len(records),
		Sample:                RedactReversePlanRecords(mapped[:sampleCount], redactFields),
		PlanHash:              planHash,
		PlanSeal:              planSeal,
		CreatedAt:             created,
		ExpiresAt:             expires,
	}
	if challenge == "" {
		token, err := randomToken(18)
		if err != nil {
			return ReversePlan{}, err
		}
		plan.ApprovalTokenHash = hashString(token)
		plan.ApprovalToken = token
	}
	stored := plan
	stored.ApprovalToken = ""
	a.state.ReversePlans = append(a.state.ReversePlans, stored)
	if err := a.save(); err != nil {
		return ReversePlan{}, err
	}
	return plan, nil
}

func (a *App) PlanConnectorCommand(ctx context.Context, req PlanConnectorCommandRequest) (ReversePlan, *connectors.WritePreview, error) {
	if err := connectors.RejectLegacyConnectorName(req.Connector); err != nil {
		return ReversePlan{}, nil, err
	}
	connector, runtime, err := a.ResolveConnectorCredential(ctx, req.Connector, req.Credential, req.Config)
	if err != nil {
		return ReversePlan{}, nil, err
	}
	writeCommand, err := commandrunner.BuildWriteCommand(ctx, connector, commandrunner.Request{
		Path:    req.Path,
		Flags:   req.Flags,
		Config:  runtime,
		Preview: false,
	})
	if err != nil {
		return ReversePlan{}, nil, err
	}
	id, err := prefixedID("rplan")
	if err != nil {
		return ReversePlan{}, nil, err
	}
	name := strings.TrimSpace(req.Name)
	if name == "" {
		name = strings.ReplaceAll(writeCommand.Command, " ", "_")
	}
	payloadIdentity, err := payloadIdentitiesForConnectorCommand(runtime.ProjectDir, connector, writeCommand.Operation, writeCommand.Record)
	if err != nil {
		return ReversePlan{}, nil, err
	}
	if req.Preview {
		runtime.ApprovedPayloadSHA256 = approvedPayloadSHA256(payloadIdentity)
		if writeCommand.Operation != "" {
			directWriter, ok := connector.(connectors.OperationDirectWriter)
			if !ok {
				return ReversePlan{}, nil, fmt.Errorf("connector %q does not support direct-write previews", connector.Name())
			}
			preview, err := directWriter.PreviewOperationDirectWrite(ctx, connectors.OperationDirectWriteRequest{
				Operation:    writeCommand.Operation,
				Config:       runtime,
				PathParams:   writeCommand.PathParams,
				Query:        writeCommand.Query,
				Headers:      writeCommand.Headers,
				HeaderValues: writeCommand.HeaderValues,
				Body:         map[string]any(writeCommand.Record),
			})
			if err != nil {
				return ReversePlan{}, nil, err
			}
			writeCommand.Preview = &preview
		} else {
			dryRunner, ok := connector.(connectors.DryRunWriter)
			if !ok {
				return ReversePlan{}, nil, fmt.Errorf("connector %q does not support reverse ETL previews", connector.Name())
			}
			preview, err := dryRunner.DryRunWrite(ctx, connectors.WriteRequest{Action: writeCommand.Write, Config: runtime}, []connectors.Record{writeCommand.Record})
			if err != nil {
				return ReversePlan{}, nil, err
			}
			writeCommand.Preview = &preview
		}
	}
	planHash, err := connectorCommandPlanHash(name, req.Connector, req.Credential, req.Config, writeCommand.Command, req.Path, writeCommand.Write, writeCommand.Record, payloadIdentity)
	if writeCommand.Operation != "" {
		planHash, err = operationConnectorCommandPlanHash(name, req.Connector, req.Credential, req.Config, writeCommand.Command, req.Path, writeCommand.Operation, writeCommand.PathParams, writeCommand.Query, writeCommand.Headers, writeCommand.HeaderValues, writeCommand.Record, payloadIdentity)
	}
	if err != nil {
		return ReversePlan{}, nil, err
	}
	confirmation := confirmationFromChallenge(writeCommand.ConfirmationChallenge)
	redactFields, structuredBody, err := connectorCommandRedactFields(connector, writeCommand.Operation, writeCommand.Write)
	if err != nil {
		return ReversePlan{}, nil, err
	}
	withheldRecord, withheldFields, sample, err := connectorCommandPlanRecords(connector, writeCommand.Operation, structuredBody, writeCommand.Record, writeCommand.RedactedRecord, redactFields)
	if err != nil {
		return ReversePlan{}, nil, err
	}
	created := time.Now().UTC()
	expires := created.Add(24 * time.Hour)
	var planSeal *connectors.WritePlanSeal
	if confirmation.Kind != "" {
		seal, err := a.approval.IssueWritePlanSeal(connectors.WritePlanSealRequest{
			PlanID: id, PlanHash: planHash, Mode: reversePlanModeConnectorCommand,
			Connector: req.Connector, Operation: writeCommand.Write,
			CredentialRevision: runtime.CredentialRevision, ConfigurationDigest: runtime.ConfigurationDigest,
			Batchable: writeCommand.Batchable, Scope: runtime.WriteApprovalScope,
			Confirmation: confirmation,
		})
		if err != nil {
			return ReversePlan{}, nil, err
		}
		planSeal = &seal
		created = seal.IssuedAt
		expires = seal.ExpiresAt
	}
	plan := ReversePlan{
		ID:                           id,
		Name:                         name,
		Status:                       "planned",
		Mode:                         reversePlanModeConnectorCommand,
		DestinationConnector:         req.Connector,
		DestinationCredential:        req.Credential,
		DestinationConfig:            cloneStringMap(req.Config),
		Action:                       writeCommand.Write,
		Mappings:                     map[string]string{},
		ConnectorCommand:             writeCommand.Command,
		ConnectorCommandPath:         append([]string(nil), req.Path...),
		ConnectorCommandOperation:    writeCommand.Operation,
		ConnectorCommandPathParams:   cloneStringMap(writeCommand.PathParams),
		ConnectorCommandQuery:        cloneStringMap(writeCommand.Query),
		ConnectorCommandHeaders:      cloneStringMap(writeCommand.Headers),
		ConnectorCommandHeaderValues: cloneStringSliceMap(writeCommand.HeaderValues),
		ConnectorCommandRecord:       withheldRecord,
		PayloadIdentity:              payloadIdentity,
		ConfirmationChallenge:        writeCommand.ConfirmationChallenge,
		ConfirmationPolicy:           confirmationFromChallenge(writeCommand.ConfirmationChallenge),
		RecordCount:                  1,
		RedactFields:                 redactFields,
		WithheldFields:               withheldFields,
		Sample:                       sample,
		PlanHash:                     planHash,
		PlanSeal:                     planSeal,
		CreatedAt:                    created,
		ExpiresAt:                    expires,
	}
	if strings.TrimSpace(writeCommand.ConfirmationChallenge) == "" && writeCommand.Operation == "" {
		token, err := randomToken(18)
		if err != nil {
			return ReversePlan{}, nil, err
		}
		plan.ApprovalTokenHash = hashString(token)
		plan.ApprovalToken = token
	}
	stored := plan
	stored.ApprovalToken = ""
	a.state.ReversePlans = append(a.state.ReversePlans, stored)
	if err := a.save(); err != nil {
		return ReversePlan{}, nil, err
	}
	if writeCommand.Preview != nil && writeCommand.Operation != "" {
		plan, err = a.persistOperationDirectWritePreview(plan, *writeCommand.Preview)
		if err != nil {
			return ReversePlan{}, nil, err
		}
	} else if writeCommand.Preview != nil && a.confirmationChallengeForPlan(plan) != "" {
		plan, err = a.persistDestructivePreview(plan, *writeCommand.Preview)
		if err != nil {
			return ReversePlan{}, nil, err
		}
	}
	return plan, writeCommand.Preview, nil
}

// reconstituteConnectorCommandRecord restores the fields the plan withheld from
// disk using the operator's re-supplied command flags. The resulting record is
// what every hash, preview and dispatch downstream uses, so a missing field, a
// wrong value and a tampered plan all fail on the existing plan-hash check.
//
// It iterates WithheldFields, the fields the plan actually removed, not the
// declared redact list: a declared field the operator never supplied is not
// owed back, and demanding it would strand the plan behind a precondition its
// own hash cannot satisfy.
func (a *App) reconstituteConnectorCommandRecord(plan ReversePlan, writer connectors.Connector, withheldFlags map[string][]string) (connectors.Record, error) {
	structuredBody := false
	var resolver connectors.OperationDirectWriteBodyValueResolver
	var transformer connectors.OperationDirectWriteBodyPlanTransformer
	if strings.TrimSpace(plan.ConnectorCommandOperation) != "" {
		metadata, err := connectorCommandDirectWriteMetadata(writer, plan.ConnectorCommandOperation)
		if err != nil {
			return nil, err
		}
		structuredBody = metadata.StructuredBody
		if structuredBody {
			var ok bool
			resolver, ok = writer.(connectors.OperationDirectWriteBodyValueResolver)
			if !ok {
				return nil, fmt.Errorf("connector %q does not expose direct-write body resolution", writer.Name())
			}
			transformer, ok = writer.(connectors.OperationDirectWriteBodyPlanTransformer)
			if !ok {
				return nil, fmt.Errorf("connector %q does not expose direct-write body plan transformation", writer.Name())
			}
		}
	}
	pending := make([]string, 0, len(plan.WithheldFields))
	for _, field := range plan.WithheldFields {
		field = strings.TrimSpace(field)
		present := recordHasField(plan.ConnectorCommandRecord, field)
		if structuredBody {
			_, resolved, err := resolver.ResolveOperationDirectWriteBodyValue(plan.ConnectorCommandOperation, map[string]any(plan.ConnectorCommandRecord), field)
			if err != nil {
				return nil, err
			}
			present = resolved
		}
		if !present {
			pending = append(pending, strings.TrimSpace(field))
		}
	}
	if len(pending) == 0 {
		return cloneRecord(plan.ConnectorCommandRecord), nil
	}
	supplied, missing, err := commandrunner.ReconstituteWithheldFields(writer, plan.ConnectorCommandPath, pending, withheldFlags)
	if err != nil {
		return nil, err
	}
	if len(missing) > 0 {
		return nil, fmt.Errorf(
			"reverse plan %q withheld %s from disk; re-supply %s on this command to preview or approve it",
			plan.ID, strings.Join(pending, ", "), strings.Join(missing, " "),
		)
	}
	if structuredBody {
		merged, err := transformer.MergeOperationDirectWriteBodyFragments(plan.ConnectorCommandOperation, map[string]any(plan.ConnectorCommandRecord), map[string]any(supplied))
		if err != nil {
			return nil, err
		}
		return connectors.Record(merged), nil
	}
	return mergeRecordFields(plan.ConnectorCommandRecord, supplied), nil
}

func (a *App) PreviewConnectorCommandPlan(ctx context.Context, id string, withheldFlags map[string][]string) (ReversePlan, connectors.WritePreview, error) {
	plan, err := a.GetReversePlan(id)
	if err != nil {
		return ReversePlan{}, connectors.WritePreview{}, err
	}
	if plan.Mode != reversePlanModeConnectorCommand {
		return ReversePlan{}, connectors.WritePreview{}, fmt.Errorf("reverse plan %q is not a connector command plan", id)
	}
	if err := approvalConsumptionUncertainError(plan, nil); err != nil {
		return ReversePlan{}, connectors.WritePreview{}, err
	}
	if err := a.previewabilityError(plan, time.Now().UTC()); err != nil {
		return ReversePlan{}, connectors.WritePreview{}, err
	}
	writer, runtime, err := a.resolveEndpoint(ctx, EndpointConfig{
		Connector:  plan.DestinationConnector,
		Credential: plan.DestinationCredential,
		Config:     plan.DestinationConfig,
	})
	if err != nil {
		return ReversePlan{}, connectors.WritePreview{}, err
	}
	if err := a.verifyPlanSealForRuntime(plan, runtime); err != nil {
		return ReversePlan{}, connectors.WritePreview{}, err
	}
	record, err := a.reconstituteConnectorCommandRecord(plan, writer, withheldFlags)
	if err != nil {
		return ReversePlan{}, connectors.WritePreview{}, err
	}
	payloadIdentity, err := payloadIdentitiesForConnectorCommand(runtime.ProjectDir, writer, plan.ConnectorCommandOperation, record)
	if err != nil {
		return ReversePlan{}, connectors.WritePreview{}, err
	}
	currentHash, err := connectorCommandHashForPlan(plan, record, payloadIdentity)
	if err != nil {
		return ReversePlan{}, connectors.WritePreview{}, err
	}
	if currentHash != plan.PlanHash {
		return ReversePlan{}, connectors.WritePreview{}, errors.New("reverse plan command payload changed before preview")
	}
	runtime.ApprovedPayloadSHA256 = approvedPayloadSHA256(payloadIdentity)
	if plan.ConnectorCommandOperation != "" {
		directWriter, ok := writer.(connectors.OperationDirectWriter)
		if !ok {
			return ReversePlan{}, connectors.WritePreview{}, fmt.Errorf("connector %q no longer supports direct-write previews", writer.Name())
		}
		preview, err := directWriter.PreviewOperationDirectWrite(ctx, connectors.OperationDirectWriteRequest{
			Operation:    plan.ConnectorCommandOperation,
			Config:       runtime,
			PathParams:   plan.ConnectorCommandPathParams,
			Query:        plan.ConnectorCommandQuery,
			Headers:      plan.ConnectorCommandHeaders,
			HeaderValues: plan.ConnectorCommandHeaderValues,
			Body:         map[string]any(record),
		})
		if err != nil {
			return ReversePlan{}, connectors.WritePreview{}, err
		}
		plan, err = a.persistOperationDirectWritePreview(plan, preview)
		if err != nil {
			return ReversePlan{}, connectors.WritePreview{}, err
		}
		return plan, preview, nil
	}
	if validator, ok := writer.(connectors.WriteValidator); ok {
		if err := validator.ValidateWrite(ctx, connectors.WriteRequest{Action: plan.Action, Config: runtime}, []connectors.Record{record}); err != nil {
			return ReversePlan{}, connectors.WritePreview{}, err
		}
	}
	dryRunner, ok := writer.(connectors.DryRunWriter)
	if !ok {
		return ReversePlan{}, connectors.WritePreview{}, fmt.Errorf("connector %q does not support reverse ETL previews", writer.Name())
	}
	preview, err := dryRunner.DryRunWrite(ctx, connectors.WriteRequest{Action: plan.Action, Config: runtime}, []connectors.Record{record})
	if err != nil {
		return ReversePlan{}, connectors.WritePreview{}, err
	}
	if a.confirmationChallengeForPlan(plan) != "" {
		plan, err = a.persistDestructivePreview(plan, preview)
		if err != nil {
			return ReversePlan{}, connectors.WritePreview{}, err
		}
	}
	return plan, preview, nil
}

func connectorCommandHashForPlan(plan ReversePlan, record connectors.Record, payloadIdentity []PayloadIdentity) (string, error) {
	if plan.ConnectorCommandOperation != "" {
		return operationConnectorCommandPlanHash(
			plan.Name,
			plan.DestinationConnector,
			plan.DestinationCredential,
			plan.DestinationConfig,
			plan.ConnectorCommand,
			plan.ConnectorCommandPath,
			plan.ConnectorCommandOperation,
			plan.ConnectorCommandPathParams,
			plan.ConnectorCommandQuery,
			plan.ConnectorCommandHeaders,
			plan.ConnectorCommandHeaderValues,
			record,
			payloadIdentity,
		)
	}
	return connectorCommandPlanHash(
		plan.Name,
		plan.DestinationConnector,
		plan.DestinationCredential,
		plan.DestinationConfig,
		plan.ConnectorCommand,
		plan.ConnectorCommandPath,
		plan.Action,
		record,
		payloadIdentity,
	)
}

// PreviewReversePlan materializes the exact mapped write request without
// dispatching it. Destructive plans become approvable only after this preview
// identity is persisted; source-row or payload drift fails before a token is
// minted.
func (a *App) PreviewReversePlan(ctx context.Context, id string, withheldFlags map[string][]string) (ReversePlan, connectors.WritePreview, error) {
	plan, err := a.GetReversePlan(id)
	if err != nil {
		return ReversePlan{}, connectors.WritePreview{}, err
	}
	if err := approvalConsumptionUncertainError(plan, nil); err != nil {
		return ReversePlan{}, connectors.WritePreview{}, err
	}
	if plan.Mode == reversePlanModeConnectorCommand {
		return a.PreviewConnectorCommandPlan(ctx, id, withheldFlags)
	}
	if err := a.previewabilityError(plan, time.Now().UTC()); err != nil {
		return ReversePlan{}, connectors.WritePreview{}, err
	}
	if err := a.guardBatchableAction(plan.DestinationConnector, plan.Action, plan.SourceTable); err != nil {
		return ReversePlan{}, connectors.WritePreview{}, err
	}
	writer, runtime, err := a.resolveEndpoint(ctx, EndpointConfig{
		Connector:  plan.DestinationConnector,
		Credential: plan.DestinationCredential,
		Config:     plan.DestinationConfig,
	})
	if err != nil {
		return ReversePlan{}, connectors.WritePreview{}, err
	}
	if err := a.verifyPlanSealForRuntime(plan, runtime); err != nil {
		return ReversePlan{}, connectors.WritePreview{}, err
	}
	// Hash and preview the precise slice that the plan approved. A row outside
	// that slice is not part of this plan's payload, so treating it as drift
	// would make every limited plan reject its own unchanged source.
	records, err := a.QueryTable(ctx, QueryTableRequest{Table: plan.SourceTable, Connection: plan.SourceConnection, Limit: max(1, plan.RecordCount)})
	if err != nil {
		return ReversePlan{}, connectors.WritePreview{}, warehouse.WithAmbiguityRemedy(err, reverseSourceRemedy(plan))
	}
	mapped := mapReverseRecords(records, plan.Mappings)
	payloadIdentity, err := payloadIdentitiesForRecords(runtime.ProjectDir, mapped)
	if err != nil {
		return ReversePlan{}, connectors.WritePreview{}, err
	}
	runtime.ApprovedPayloadSHA256 = approvedPayloadSHA256(payloadIdentity)
	currentHash, err := reversePlanHash(plan.Name, plan.SourceTable, plan.DestinationConnector, plan.DestinationCredential, plan.Action, plan.DestinationConfig, plan.Mappings, mapped, payloadIdentity)
	if err != nil {
		return ReversePlan{}, connectors.WritePreview{}, err
	}
	if currentHash != plan.PlanHash {
		return ReversePlan{}, connectors.WritePreview{}, errors.New("reverse plan source rows or payload files changed before preview")
	}
	request := connectors.WriteRequest{Stream: "records", Table: plan.Name, Action: plan.Action, Config: runtime}
	if validator, ok := writer.(connectors.WriteValidator); ok {
		if err := validator.ValidateWrite(ctx, request, mapped); err != nil {
			return ReversePlan{}, connectors.WritePreview{}, err
		}
	}
	dryRunner, ok := writer.(connectors.DryRunWriter)
	if !ok {
		return ReversePlan{}, connectors.WritePreview{}, fmt.Errorf("connector %q does not support reverse ETL previews", writer.Name())
	}
	preview, err := dryRunner.DryRunWrite(ctx, request, mapped)
	if err != nil {
		return ReversePlan{}, connectors.WritePreview{}, err
	}
	if a.confirmationChallengeForPlan(plan) != "" {
		plan, err = a.persistDestructivePreview(plan, preview)
		if err != nil {
			return ReversePlan{}, connectors.WritePreview{}, err
		}
	}
	return plan, preview, nil
}

func (a *App) persistDestructivePreview(plan ReversePlan, preview connectors.WritePreview) (ReversePlan, error) {
	if strings.TrimSpace(preview.Digest) == "" {
		return ReversePlan{}, fmt.Errorf("connector preview for destructive plan %q has no digest", plan.ID)
	}
	if strings.TrimSpace(preview.ApprovalTarget.Connector) == "" || preview.ApprovalTarget.Connector != plan.DestinationConnector || preview.ApprovalTarget.Operation != plan.Action {
		return ReversePlan{}, fmt.Errorf("connector preview for destructive plan %q has no matching approval target", plan.ID)
	}
	token, err := randomToken(18)
	if err != nil {
		return ReversePlan{}, err
	}
	now := time.Now().UTC()
	var issued ReversePlan
	updated, err := a.updateState(func(current state) (state, error) {
		for i := range current.ReversePlans {
			stored := current.ReversePlans[i]
			if stored.ID != plan.ID {
				continue
			}
			if err := approvalConsumptionUncertainError(stored, nil); err != nil {
				return current, err
			}
			if err := a.previewabilityError(stored, now); err != nil {
				return current, err
			}
			if stored.PlanHash != plan.PlanHash || stored.DestinationConnector != plan.DestinationConnector || stored.DestinationCredential != plan.DestinationCredential || stored.Action != plan.Action {
				return current, fmt.Errorf("reverse plan %q changed while its preview was prepared", plan.ID)
			}
			if err := a.verifyPlanSealForTarget(stored, preview.ApprovalTarget); err != nil {
				return current, err
			}
			grant, err := a.approval.IssueWriteGrant(connectors.WriteApprovalGrantRequest{
				PlanID:        stored.ID,
				PlanHash:      stored.PlanHash,
				Mode:          stored.Mode,
				PlanSeal:      stored.PlanSeal,
				PreviewDigest: preview.Digest,
				ApprovalToken: token,
				Target:        preview.ApprovalTarget,
				Confirmation:  a.confirmationPolicyForPlan(stored),
			})
			if err != nil {
				return current, err
			}
			stored.Status = "previewed"
			stored.PreviewDigest = preview.Digest
			stored.PreviewedAt = now
			stored.ApprovalTokenHash = hashString(token)
			stored.ApprovalGrant = &grant
			stored.ApprovalConsumedAt = time.Time{}
			current.ReversePlans[i] = stored
			issued = stored
			issued.ApprovalToken = token
			return current, nil
		}
		return current, fmt.Errorf("reverse plan %q not found", plan.ID)
	})
	if err != nil {
		return ReversePlan{}, err
	}
	a.state = updated
	return issued, nil
}

// persistOperationDirectWritePreview records a no-network direct-write preview
// before it exposes an approval token. Destructive operations delegate to the
// project-wide authenticated grant flow; non-destructive writes still require
// a persisted preview and a single-use token, but do not mint approval
// evidence that the engine gate does not require.
func (a *App) persistOperationDirectWritePreview(plan ReversePlan, preview connectors.WritePreview) (ReversePlan, error) {
	if plan.ConnectorCommandOperation == "" {
		return ReversePlan{}, fmt.Errorf("reverse plan %q is not a direct-write command", plan.ID)
	}
	if strings.TrimSpace(preview.Digest) == "" {
		return ReversePlan{}, fmt.Errorf("connector preview for direct-write plan %q has no digest", plan.ID)
	}
	if strings.TrimSpace(preview.ApprovalTarget.Connector) == "" || preview.ApprovalTarget.Connector != plan.DestinationConnector || preview.ApprovalTarget.Operation != plan.ConnectorCommandOperation {
		return ReversePlan{}, fmt.Errorf("connector preview for direct-write plan %q has no matching approval target", plan.ID)
	}
	if a.confirmationChallengeForPlan(plan) != "" {
		return a.persistDestructivePreview(plan, preview)
	}

	token, err := randomToken(18)
	if err != nil {
		return ReversePlan{}, err
	}
	now := time.Now().UTC()
	var issued ReversePlan
	updated, err := a.updateState(func(current state) (state, error) {
		for i := range current.ReversePlans {
			stored := current.ReversePlans[i]
			if stored.ID != plan.ID {
				continue
			}
			if err := approvalConsumptionUncertainError(stored, nil); err != nil {
				return current, err
			}
			if err := a.previewabilityError(stored, now); err != nil {
				return current, err
			}
			if stored.PlanHash != plan.PlanHash || stored.DestinationConnector != plan.DestinationConnector || stored.DestinationCredential != plan.DestinationCredential || stored.ConnectorCommandOperation != plan.ConnectorCommandOperation {
				return current, fmt.Errorf("reverse plan %q changed while its direct-write preview was prepared", plan.ID)
			}
			stored.Status = "previewed"
			stored.PreviewDigest = preview.Digest
			stored.PreviewedAt = now
			stored.ApprovalTokenHash = hashString(token)
			stored.ApprovalGrant = nil
			stored.ApprovalConsumedAt = time.Time{}
			current.ReversePlans[i] = stored
			issued = stored
			issued.ApprovalToken = token
			return current, nil
		}
		return current, fmt.Errorf("reverse plan %q not found", plan.ID)
	})
	if err != nil {
		return ReversePlan{}, err
	}
	a.state = updated
	return issued, nil
}

func (a *App) previewabilityError(plan ReversePlan, now time.Time) error {
	if plan.Status != "planned" && plan.Status != "previewed" {
		return fmt.Errorf("reverse plan %q was already %s", plan.ID, plan.Status)
	}
	if a.confirmationPolicyForPlan(plan).Kind == "" && (plan.CreatedAt.IsZero() || plan.ExpiresAt.IsZero() || !plan.ExpiresAt.After(plan.CreatedAt) || now.Before(plan.CreatedAt) || !now.Before(plan.ExpiresAt)) {
		return fmt.Errorf("reverse plan %q approval has expired or is not active", plan.ID)
	}
	return nil
}

func confirmationFromChallenge(challenge string) connectors.WriteConfirmation {
	confirmation, err := connectors.ParseWriteConfirmation(challenge)
	if err != nil {
		return connectors.WriteConfirmation{}
	}
	return confirmation
}

func (a *App) confirmationPolicyForAction(connectorName, actionName string) connectors.WriteConfirmation {
	connector, ok := a.registry.Get(connectorName)
	if !ok {
		return connectors.WriteConfirmation{}
	}
	for _, action := range connectors.ManifestOf(connector).WriteActions {
		if action.Name == actionName {
			return connectors.ConfirmationForWriteAction(action)
		}
	}
	return connectors.WriteConfirmation{}
}

func (a *App) confirmationPolicyForPlan(plan ReversePlan) connectors.WriteConfirmation {
	// Prefer the current connector manifest so a local state edit cannot remove
	// a destructive-action confirmation gate from an already-created plan. The
	// stored plan challenge remains a compatibility fallback for older plans or
	// connectors that are temporarily unavailable.
	if plan.ConnectorCommandOperation != "" {
		if connector, ok := a.registry.Get(plan.DestinationConnector); ok {
			if provider, ok := connector.(connectors.OperationDirectWriteMetadataProvider); ok {
				if metadata, err := provider.OperationDirectWriteMetadata(plan.ConnectorCommandOperation); err == nil && metadata.Operation == plan.ConnectorCommandOperation {
					if confirmation := confirmationFromChallenge(metadata.ConfirmationChallenge); confirmation.Kind != "" {
						return confirmation
					}
				}
			}
		}
	}
	if confirmation := a.confirmationPolicyForAction(plan.DestinationConnector, plan.Action); confirmation.Kind != "" {
		return confirmation
	}
	if plan.ConfirmationPolicy.Kind != "" {
		return plan.ConfirmationPolicy
	}
	return confirmationFromChallenge(plan.ConfirmationChallenge)
}

func (a *App) actionIsBatchable(connectorName, actionName string) bool {
	connector, ok := a.registry.Get(connectorName)
	if !ok {
		return true
	}
	for _, action := range connectors.ManifestOf(connector).WriteActions {
		if action.Name == actionName {
			return action.IsBatchable()
		}
	}
	return true
}

func (a *App) planIsBatchable(plan ReversePlan) (bool, error) {
	if plan.ConnectorCommandOperation == "" {
		return a.actionIsBatchable(plan.DestinationConnector, plan.Action), nil
	}
	connector, ok := a.registry.Get(plan.DestinationConnector)
	if !ok {
		return false, fmt.Errorf("connector %q is unavailable for direct-write batchability validation", plan.DestinationConnector)
	}
	provider, ok := connector.(connectors.OperationDirectWriteMetadataProvider)
	if !ok {
		return false, fmt.Errorf("connector %q no longer exposes direct-write metadata", plan.DestinationConnector)
	}
	metadata, err := provider.OperationDirectWriteMetadata(plan.ConnectorCommandOperation)
	if err != nil {
		return false, err
	}
	if metadata.Operation != plan.ConnectorCommandOperation {
		return false, fmt.Errorf("connector %q direct-write metadata no longer matches operation %q", plan.DestinationConnector, plan.ConnectorCommandOperation)
	}
	return metadata.Batchable, nil
}

func (a *App) verifyPlanSealForRuntime(plan ReversePlan, runtime connectors.RuntimeConfig) error {
	confirmation := a.confirmationPolicyForPlan(plan)
	if confirmation.Kind == "" {
		return nil
	}
	if plan.PlanSeal == nil {
		return fmt.Errorf("reverse plan %q has no authenticated plan seal", plan.ID)
	}
	batchable, err := a.planIsBatchable(plan)
	if err != nil {
		return err
	}
	return a.approval.VerifyWritePlanSeal(*plan.PlanSeal, connectors.WritePlanSealExpectation{
		PlanID: plan.ID, PlanHash: plan.PlanHash, Mode: plan.Mode,
		Connector: plan.DestinationConnector, Operation: plan.Action,
		CredentialRevision: runtime.CredentialRevision, ConfigurationDigest: runtime.ConfigurationDigest,
		Batchable: batchable, Scope: runtime.WriteApprovalScope,
		Confirmation: confirmation,
	})
}

func (a *App) verifyPlanSealForTarget(plan ReversePlan, target connectors.WriteApprovalTarget) error {
	if plan.PlanSeal == nil {
		return fmt.Errorf("reverse plan %q has no authenticated plan seal", plan.ID)
	}
	return a.approval.VerifyWritePlanSeal(*plan.PlanSeal, connectors.WritePlanSealExpectation{
		PlanID: plan.ID, PlanHash: plan.PlanHash, Mode: plan.Mode,
		Connector: target.Connector, Operation: target.Operation,
		CredentialRevision: target.CredentialRevision, ConfigurationDigest: target.ConfigurationDigest,
		Batchable: target.Batchable, Scope: target.Scope, Confirmation: target.Confirmation,
	})
}

func (a *App) confirmationChallengeForPlan(plan ReversePlan) string {
	return string(a.confirmationPolicyForPlan(plan).Kind)
}

func (a *App) planRequiresPersistedPreview(plan ReversePlan) bool {
	return plan.ConnectorCommandOperation != "" || isIssueLabelTransportMode(plan.Mode) || a.confirmationChallengeForPlan(plan) != ""
}

func (a *App) validatePlanConfirmation(plan ReversePlan, got connectors.WriteConfirmation) error {
	want := a.confirmationPolicyForPlan(plan)
	if want.Kind == "" {
		return nil
	}
	if got.Kind != want.Kind {
		return fmt.Errorf("reverse plan %q requires typed confirmation: pass --confirm %s", plan.ID, want.Kind)
	}
	return nil
}

func (a *App) GetReversePlan(id string) (ReversePlan, error) {
	if a == nil {
		return ReversePlan{}, errors.New("app is required")
	}
	// Production Apps always have a configured store and must reload here so
	// unattended flows revalidate the latest durable approval. A small set of
	// isolated policy tests intentionally construct an in-memory App; preserve
	// their pre-existing lookup semantics without creating a fake state path.
	if strings.TrimSpace(a.store.Path) == "" {
		return reversePlanFromState(a.state, id)
	}
	return a.loadReversePlan(id)
}

func (a *App) ListReversePlans() []ReversePlan {
	out := append([]ReversePlan(nil), a.state.ReversePlans...)
	sort.Slice(out, func(i, j int) bool { return out[i].CreatedAt.Before(out[j].CreatedAt) })
	return out
}

func (a *App) GetReverseRun(id string) (ReverseRun, error) {
	for _, run := range a.state.ReverseRuns {
		if run.ID == id {
			return run, nil
		}
	}
	return ReverseRun{}, fmt.Errorf("reverse run %q not found", id)
}

func (a *App) ListReverseRuns() []ReverseRun {
	out := append([]ReverseRun(nil), a.state.ReverseRuns...)
	sort.Slice(out, func(i, j int) bool { return out[i].StartedAt.Before(out[j].StartedAt) })
	return out
}

func (a *App) RunReverseETL(ctx context.Context, req RunReverseETLRequest) (ReverseRun, error) {
	plan, err := a.loadReversePlan(req.PlanID)
	if err != nil {
		return ReverseRun{}, err
	}
	if err := approvalConsumptionUncertainError(plan, nil); err != nil {
		return ReverseRun{}, err
	}
	if plan.AuthorizationReference != "" {
		if req.ApprovalToken != "" {
			return ReverseRun{}, &AuthorizationTokenReplayError{Reference: plan.AuthorizationReference}
		}
		if plan.Mode == reversePlanModeConnectorCommand {
			return ReverseRun{}, &AuthorizationScopeChangedError{Reference: plan.AuthorizationReference, Property: "plan_mode"}
		}
		return a.runAuthorizedBulkReversePlan(ctx, plan)
	}
	if plan.Status != "planned" && plan.ApprovalTokenHash == "" {
		return ReverseRun{}, errors.New("reverse plan approval has already been consumed")
	}
	if err := a.previewabilityError(plan, time.Now().UTC()); err != nil {
		return ReverseRun{}, err
	}
	if a.planRequiresPersistedPreview(plan) && (plan.Status != "previewed" || plan.PreviewDigest == "" || plan.PreviewedAt.IsZero()) {
		return ReverseRun{}, fmt.Errorf("reverse plan %q must be previewed before approval", plan.ID)
	}
	if plan.ApprovalTokenHash == "" {
		return ReverseRun{}, errors.New("reverse plan approval has already been consumed")
	}
	if !constantTimeStringEqual(hashString(req.ApprovalToken), plan.ApprovalTokenHash) {
		return ReverseRun{}, errors.New("approval token is invalid")
	}
	if err := a.validatePlanConfirmation(plan, req.Confirmation); err != nil {
		return ReverseRun{}, err
	}
	if plan.Mode == reversePlanModeConnectorCommand {
		return a.runConnectorCommandPlan(ctx, plan, req)
	}
	return a.runBulkReversePlan(ctx, plan, req)
}

func (a *App) runBulkReversePlan(ctx context.Context, plan ReversePlan, req RunReverseETLRequest) (ReverseRun, error) {
	if err := a.guardBatchableAction(plan.DestinationConnector, plan.Action, plan.SourceTable); err != nil {
		return ReverseRun{}, err
	}
	dest := EndpointConfig{Connector: plan.DestinationConnector, Credential: plan.DestinationCredential, Config: plan.DestinationConfig}
	writer, credential, runtime, err := a.resolveEndpointWithCredential(ctx, dest)
	if err != nil {
		return ReverseRun{}, err
	}
	scope := a.authorizationScopeForReversePlan(plan, credential, runtime)
	authorization, err := newAuthorizationRecord(scope, time.Now().UTC())
	if err != nil {
		return ReverseRun{}, err
	}
	// Re-read exactly the records whose mapped payload produced PlanHash. This
	// keeps the hash check and the dispatched records on the same approved slice.
	records, err := a.QueryTable(ctx, QueryTableRequest{Table: plan.SourceTable, Connection: plan.SourceConnection, Limit: max(1, plan.RecordCount)})
	if err != nil {
		return ReverseRun{}, warehouse.WithAmbiguityRemedy(err, reverseSourceRemedy(plan))
	}
	mappedForHash := mapReverseRecords(records, plan.Mappings)
	payloadIdentity, err := payloadIdentitiesForRecords(runtime.ProjectDir, mappedForHash)
	if err != nil {
		return ReverseRun{}, err
	}
	planHash, err := reversePlanHash(plan.Name, plan.SourceTable, plan.DestinationConnector, plan.DestinationCredential, plan.Action, plan.DestinationConfig, plan.Mappings, mappedForHash, payloadIdentity)
	if err != nil {
		return ReverseRun{}, err
	}
	if planHash != plan.PlanHash {
		if err := a.invalidateReversePlan(plan); err != nil {
			return ReverseRun{}, err
		}
		return ReverseRun{}, errors.New("reverse plan source rows or payload files changed since approval")
	}
	runtime.ApprovedPayloadSHA256 = approvedPayloadSHA256(plan.PayloadIdentity)
	mapped := mapReverseRecords(records, plan.Mappings)
	writeRequest := connectors.WriteRequest{Stream: "records", Table: plan.Name, Action: plan.Action, Config: runtime}
	preview, err := a.validateDestructivePreview(ctx, writer, plan, writeRequest, mapped)
	if err != nil {
		return ReverseRun{}, err
	}
	runID, err := prefixedID("rrun")
	if err != nil {
		return ReverseRun{}, err
	}
	run := ReverseRun{ID: runID, PlanID: plan.ID, Status: "running", RecordsStaged: len(mapped), StartedAt: time.Now().UTC()}
	evidence, _, err := a.consumePlanApproval(plan, req, preview, &authorization)
	if err != nil {
		return ReverseRun{}, err
	}
	writeRequest.Approval = evidence
	result, err := writer.Write(ctx, writeRequest, mapped)
	return a.finishReverseWrite(plan.ID, run, result, runtime, len(mapped), err)
}

// runAuthorizedBulkReversePlan is the standing-authorization path. It
// re-derives and verifies the content-free scope before it reads the table or
// reaches a destination connector, then intentionally dispatches current
// content rather than reusing the old payload-bound PlanHash.
func (a *App) runAuthorizedBulkReversePlan(ctx context.Context, plan ReversePlan) (ReverseRun, error) {
	if err := a.guardBatchableAction(plan.DestinationConnector, plan.Action, plan.SourceTable); err != nil {
		return ReverseRun{}, err
	}
	writer, credential, runtime, err := a.resolveEndpointWithCredential(ctx, EndpointConfig{
		Connector: plan.DestinationConnector, Credential: plan.DestinationCredential, Config: plan.DestinationConfig,
	})
	if err != nil {
		return ReverseRun{}, err
	}
	scope := a.authorizationScopeForReversePlan(plan, credential, runtime)
	if _, err := a.requireAuthorization(plan.AuthorizationReference, scope, time.Now().UTC()); err != nil {
		return ReverseRun{}, err
	}
	records, err := a.QueryTable(ctx, QueryTableRequest{Table: plan.SourceTable, Connection: plan.SourceConnection, Limit: 100000})
	if err != nil {
		return ReverseRun{}, warehouse.WithAmbiguityRemedy(err, reverseSourceRemedy(plan))
	}
	mapped := mapReverseRecords(records, plan.Mappings)
	writeRequest := connectors.WriteRequest{Stream: "records", Table: plan.Name, Action: plan.Action, Config: runtime}
	if scope.ConfirmationPolicy.Kind != "" {
		if _, err := validateAuthorizedDestructivePreview(ctx, writer, writeRequest, mapped); err != nil {
			return ReverseRun{}, err
		}
		evidence, err := durableAuthorizationEvidence(scope)
		if err != nil {
			return ReverseRun{}, err
		}
		writeRequest.Approval = evidence
	}
	runID, err := prefixedID("rrun")
	if err != nil {
		return ReverseRun{}, err
	}
	run := ReverseRun{ID: runID, PlanID: plan.ID, Status: "running", RecordsStaged: len(mapped), StartedAt: time.Now().UTC()}
	result, err := writer.Write(ctx, writeRequest, mapped)
	return a.finishReverseWrite(plan.ID, run, result, runtime, len(mapped), err)
}

func (a *App) validateDestructivePreview(ctx context.Context, writer connectors.Connector, plan ReversePlan, request connectors.WriteRequest, records []connectors.Record) (connectors.WritePreview, error) {
	if a.confirmationPolicyForPlan(plan).Kind == "" {
		return connectors.WritePreview{}, nil
	}
	dryRunner, ok := writer.(connectors.DryRunWriter)
	if !ok {
		return connectors.WritePreview{}, fmt.Errorf("connector %q no longer supports the required destructive preview", writer.Name())
	}
	preview, err := dryRunner.DryRunWrite(ctx, request, records)
	if err != nil {
		return connectors.WritePreview{}, fmt.Errorf("revalidate destructive preview: %w", err)
	}
	if strings.TrimSpace(plan.PreviewDigest) == "" || strings.TrimSpace(preview.Digest) == "" || subtle.ConstantTimeCompare([]byte(plan.PreviewDigest), []byte(preview.Digest)) != 1 {
		return connectors.WritePreview{}, fmt.Errorf("reverse plan %q no longer matches its approved preview", plan.ID)
	}
	return preview, nil
}

func validateAuthorizedDestructivePreview(ctx context.Context, writer connectors.Connector, request connectors.WriteRequest, records []connectors.Record) (connectors.WritePreview, error) {
	dryRunner, ok := writer.(connectors.DryRunWriter)
	if !ok {
		return connectors.WritePreview{}, fmt.Errorf("connector %q no longer supports the required destructive preview", writer.Name())
	}
	preview, err := dryRunner.DryRunWrite(ctx, request, records)
	if err != nil {
		return connectors.WritePreview{}, fmt.Errorf("revalidate durable authorization preview: %w", err)
	}
	return preview, nil
}

func (a *App) runConnectorCommandPlan(ctx context.Context, plan ReversePlan, req RunReverseETLRequest) (ReverseRun, error) {
	writer, runtime, err := a.resolveEndpoint(ctx, EndpointConfig{
		Connector:  plan.DestinationConnector,
		Credential: plan.DestinationCredential,
		Config:     plan.DestinationConfig,
	})
	if err != nil {
		return ReverseRun{}, err
	}
	record, err := a.reconstituteConnectorCommandRecord(plan, writer, req.WithheldFlags)
	if err != nil {
		return ReverseRun{}, err
	}
	payloadIdentity, err := payloadIdentitiesForConnectorCommand(runtime.ProjectDir, writer, plan.ConnectorCommandOperation, record)
	if err != nil {
		return ReverseRun{}, err
	}
	planHash, err := connectorCommandHashForPlan(plan, record, payloadIdentity)
	if err != nil {
		return ReverseRun{}, err
	}
	if planHash != plan.PlanHash {
		if err := a.invalidateReversePlan(plan); err != nil {
			return ReverseRun{}, err
		}
		return ReverseRun{}, errors.New("reverse plan command payload changed since approval")
	}
	runtime.ApprovedPayloadSHA256 = approvedPayloadSHA256(plan.PayloadIdentity)
	if plan.ConnectorCommandOperation != "" {
		return a.runOperationDirectWritePlan(ctx, writer, runtime, plan, record, req)
	}
	runID, err := prefixedID("rrun")
	if err != nil {
		return ReverseRun{}, err
	}
	records := []connectors.Record{record}
	writeRequest := connectors.WriteRequest{Stream: "records", Table: plan.Name, Action: plan.Action, Config: runtime}
	preview, err := a.validateDestructivePreview(ctx, writer, plan, writeRequest, records)
	if err != nil {
		return ReverseRun{}, err
	}
	run := ReverseRun{ID: runID, PlanID: plan.ID, Status: "running", RecordsStaged: len(records), StartedAt: time.Now().UTC()}
	evidence, _, err := a.consumePlanApproval(plan, req, preview, nil)
	if err != nil {
		return ReverseRun{}, err
	}
	writeRequest.Approval = evidence
	result, err := writer.Write(ctx, writeRequest, records)
	if err == nil && (result.RecordsWritten != len(records) || result.RecordsFailed != 0 || result.RecordsUnchanged != 0) {
		err = fmt.Errorf(
			"connector command acknowledgement is incomplete: wrote %d, unchanged %d, failed %d of %d records",
			result.RecordsWritten,
			result.RecordsUnchanged,
			result.RecordsFailed,
			len(records),
		)
	}
	return a.finishReverseWrite(plan.ID, run, result, runtime, len(records), err)
}

func (a *App) runOperationDirectWritePlan(ctx context.Context, writer connectors.Connector, runtime connectors.RuntimeConfig, plan ReversePlan, record connectors.Record, req RunReverseETLRequest) (ReverseRun, error) {
	directWriter, ok := writer.(connectors.OperationDirectWriter)
	if !ok {
		return ReverseRun{}, fmt.Errorf("connector %q no longer supports direct writes", writer.Name())
	}
	operationRequest := connectors.OperationDirectWriteRequest{
		Operation:    plan.ConnectorCommandOperation,
		Config:       runtime,
		PathParams:   plan.ConnectorCommandPathParams,
		Query:        plan.ConnectorCommandQuery,
		Headers:      plan.ConnectorCommandHeaders,
		HeaderValues: plan.ConnectorCommandHeaderValues,
		Body:         map[string]any(record),
	}
	preview, err := validateOperationDirectWritePreview(ctx, directWriter, plan, operationRequest)
	if err != nil {
		return ReverseRun{}, err
	}
	runID, err := prefixedID("rrun")
	if err != nil {
		return ReverseRun{}, err
	}
	run := ReverseRun{ID: runID, PlanID: plan.ID, Status: "running", RecordsStaged: 1, StartedAt: time.Now().UTC()}
	evidence, _, err := a.consumePlanApproval(plan, req, preview, nil)
	if err != nil {
		return ReverseRun{}, err
	}
	operationRequest.Approval = evidence
	operationRequest.PreviewDigest = preview.Digest
	operationResult, writeErr := directWriter.OperationDirectWrite(ctx, operationRequest)
	writeResult := connectors.WriteResult{RecordsWritten: 1}
	if operationResult.Operation != "" {
		safeOperationResult := connectors.SanitizeOperationDirectWriteResultForOutput(operationResult, runtime.Secrets)
		run.OperationDirectWrite = &safeOperationResult
	}
	if writeErr != nil {
		writeResult = connectors.WriteResult{RecordsFailed: 1}
	}
	return a.finishOperationDirectWrite(plan.ID, run, writeResult, runtime, 1, writeErr)
}

func validateOperationDirectWritePreview(ctx context.Context, writer connectors.OperationDirectWriter, plan ReversePlan, request connectors.OperationDirectWriteRequest) (connectors.WritePreview, error) {
	preview, err := writer.PreviewOperationDirectWrite(ctx, request)
	if err != nil {
		return connectors.WritePreview{}, fmt.Errorf("revalidate direct-write preview: %w", err)
	}
	if strings.TrimSpace(plan.PreviewDigest) == "" || strings.TrimSpace(preview.Digest) == "" || subtle.ConstantTimeCompare([]byte(plan.PreviewDigest), []byte(preview.Digest)) != 1 {
		return connectors.WritePreview{}, fmt.Errorf("reverse plan %q no longer matches its approved preview", plan.ID)
	}
	return preview, nil
}

func (a *App) loadReversePlan(id string) (ReversePlan, error) {
	loaded, err := a.store.LoadReadOnly()
	if err != nil {
		return ReversePlan{}, err
	}
	if a.deferredState != nil && loaded.Revision == a.deferredStateRevision {
		a.state = *a.deferredState
	} else {
		if err := a.normalizeLoadedState(loaded, false); err != nil {
			return ReversePlan{}, err
		}
		a.rememberDeferredState(loaded.Revision)
	}
	return reversePlanFromState(a.state, id)
}

func reversePlanFromState(loaded state, id string) (ReversePlan, error) {
	for _, plan := range loaded.ReversePlans {
		if plan.ID == id {
			return plan, nil
		}
	}
	return ReversePlan{}, fmt.Errorf("reverse plan %q not found", id)
}

func reversePlanMatchesExpected(stored, expected ReversePlan) bool {
	return stored.PlanHash == expected.PlanHash &&
		stored.DestinationConnector == expected.DestinationConnector &&
		stored.DestinationCredential == expected.DestinationCredential &&
		stored.Action == expected.Action &&
		stored.Mode == expected.Mode &&
		stored.ConnectorCommandOperation == expected.ConnectorCommandOperation
}

func (a *App) consumePlanApproval(expected ReversePlan, req RunReverseETLRequest, preview connectors.WritePreview, authorization *AuthorizationRecord) (*connectors.WriteApprovalEvidence, ReversePlan, error) {
	now := time.Now().UTC()
	var consumed ReversePlan
	var evidence *connectors.WriteApprovalEvidence
	updated, err := a.updateStateAfterPreflight(func(current state) error {
		return a.validateApprovalConsumptionState(current, expected, req, preview, now)
	}, func(current state) (state, error) {
		next, candidate, candidateEvidence, consumeErr := a.approvalConsumptionState(current, expected, req, preview, authorization, now)
		if consumeErr != nil {
			return current, consumeErr
		}
		consumed = candidate
		evidence = candidateEvidence
		return next, nil
	})
	if err != nil {
		if stateStoreCommitMayHaveSucceeded(err) {
			a.deferredState = nil
			a.deferredStateRevision = 0
			return nil, ReversePlan{}, approvalConsumptionUncertainError(consumed, err)
		}
		return nil, ReversePlan{}, err
	}
	a.state = updated
	a.deferredState = nil
	a.deferredStateRevision = 0
	return evidence, consumed, nil
}

func (a *App) validateApprovalConsumptionState(current state, expected ReversePlan, req RunReverseETLRequest, preview connectors.WritePreview, now time.Time) error {
	_, _, err := a.approvalConsumptionCandidate(current, expected, req, preview, now)
	return err
}

func (a *App) approvalConsumptionCandidate(current state, expected ReversePlan, req RunReverseETLRequest, preview connectors.WritePreview, now time.Time) (int, ReversePlan, error) {
	for i := range current.ReversePlans {
		stored := current.ReversePlans[i]
		if stored.ID != expected.ID {
			continue
		}
		if err := approvalConsumptionUncertainError(stored, nil); err != nil {
			return 0, ReversePlan{}, err
		}
		if stored.ApprovalTokenHash == "" {
			return 0, ReversePlan{}, errors.New("reverse plan approval has already been consumed")
		}
		if err := a.previewabilityError(stored, now); err != nil {
			return 0, ReversePlan{}, err
		}
		if !reversePlanMatchesExpected(stored, expected) {
			return 0, ReversePlan{}, fmt.Errorf("reverse plan %q changed before approval consumption", stored.ID)
		}
		if err := a.validatePlanConfirmation(stored, req.Confirmation); err != nil {
			return 0, ReversePlan{}, err
		}
		if !constantTimeStringEqual(stored.ApprovalTokenHash, hashString(req.ApprovalToken)) {
			return 0, ReversePlan{}, errors.New("approval token is invalid")
		}
		if a.planRequiresPersistedPreview(stored) {
			if stored.Status != "previewed" || stored.PreviewDigest == "" || stored.PreviewedAt.IsZero() {
				return 0, ReversePlan{}, fmt.Errorf("reverse plan %q must be previewed before approval", stored.ID)
			}
			if !constantTimeStringEqual(stored.PreviewDigest, preview.Digest) {
				return 0, ReversePlan{}, fmt.Errorf("reverse plan %q no longer matches its approved preview", stored.ID)
			}
		}
		if a.confirmationPolicyForPlan(stored).Kind != "" {
			if stored.ApprovalGrant == nil {
				return 0, ReversePlan{}, fmt.Errorf("reverse plan %q has no destructive approval grant", stored.ID)
			}
			if err := a.approval.ValidateWriteGrant(*stored.ApprovalGrant, connectors.WriteApprovalExpectation{
				PlanID:        expected.ID,
				PlanHash:      expected.PlanHash,
				Mode:          expected.Mode,
				PreviewDigest: preview.Digest,
				ApprovalToken: req.ApprovalToken,
				Target:        preview.ApprovalTarget,
				Confirmation:  req.Confirmation,
			}, stored.PlanSeal); err != nil {
				return 0, ReversePlan{}, err
			}
		}
		return i, stored, nil
	}
	return 0, ReversePlan{}, fmt.Errorf("reverse plan %q not found", expected.ID)
}

func (a *App) approvalConsumptionState(current state, expected ReversePlan, req RunReverseETLRequest, preview connectors.WritePreview, authorization *AuthorizationRecord, now time.Time) (state, ReversePlan, *connectors.WriteApprovalEvidence, error) {
	i, stored, err := a.approvalConsumptionCandidate(current, expected, req, preview, now)
	if err != nil {
		return current, ReversePlan{}, nil, err
	}
	var evidence *connectors.WriteApprovalEvidence
	if a.confirmationPolicyForPlan(stored).Kind != "" {
		verified, err := a.approval.VerifyWriteGrant(*stored.ApprovalGrant, connectors.WriteApprovalExpectation{
			PlanID:        expected.ID,
			PlanHash:      expected.PlanHash,
			Mode:          expected.Mode,
			PreviewDigest: preview.Digest,
			ApprovalToken: req.ApprovalToken,
			Target:        preview.ApprovalTarget,
			Confirmation:  req.Confirmation,
		}, stored.PlanSeal)
		if err != nil {
			return current, ReversePlan{}, nil, err
		}
		evidence = verified
	}
	normalized, err := a.stateForApprovalConsumption(current)
	if err != nil {
		return current, ReversePlan{}, nil, err
	}
	current = normalized
	if i >= len(current.ReversePlans) || current.ReversePlans[i].ID != expected.ID {
		return current, ReversePlan{}, nil, fmt.Errorf("reverse plan %q changed before approval consumption", expected.ID)
	}
	stored = current.ReversePlans[i]
	stored.Status = reversePlanStatusApprovalConsumptionUncertain
	stored.ApprovalTokenHash = ""
	stored.ApprovalGrant = nil
	stored.ApprovalConsumedAt = now
	stored.ApprovalUncertainAt = now
	if authorization != nil {
		if stored.AuthorizationReference != "" {
			return current, ReversePlan{}, nil, fmt.Errorf("reverse plan %q already has durable authorization %q", stored.ID, stored.AuthorizationReference)
		}
		record := cloneAuthorizationRecord(*authorization)
		if strings.TrimSpace(record.Reference) == "" || strings.TrimSpace(record.ScopeIdentity) == "" {
			return current, ReversePlan{}, nil, errors.New("durable authorization record is incomplete")
		}
		for _, existing := range current.Authorizations {
			if existing.Reference == record.Reference {
				return current, ReversePlan{}, nil, fmt.Errorf("durable authorization %q already exists", record.Reference)
			}
		}
		stored.AuthorizationReference = record.Reference
		current.Authorizations = append(current.Authorizations, record)
	}
	current.ReversePlans[i] = stored
	return current, stored, evidence, nil
}

func approvalConsumptionUncertainError(plan ReversePlan, cause error) error {
	if plan.Status != reversePlanStatusApprovalConsumptionUncertain {
		return nil
	}
	return &ApprovalConsumptionUncertainError{
		PlanID:     plan.ID,
		ConsumedAt: plan.ApprovalConsumedAt,
		err:        cause,
	}
}

func (a *App) finishReverseWrite(planID string, run ReverseRun, result connectors.WriteResult, runtime connectors.RuntimeConfig, staged int, writeErr error) (ReverseRun, error) {
	return a.finishReverseWriteWithErrorText(planID, run, result, runtime, staged, writeErr, func(err error) string {
		return connectors.SanitizeWriteErrorForOutput(err, runtime.Secrets)
	})
}

// finishOperationDirectWrite preserves the direct-write error text in its
// persisted report. Both direct and ordinary typed reverse-ETL results retain
// complete provider output through their closed result contracts.
func (a *App) finishOperationDirectWrite(planID string, run ReverseRun, result connectors.WriteResult, runtime connectors.RuntimeConfig, staged int, writeErr error) (ReverseRun, error) {
	return a.finishReverseWriteWithErrorText(planID, run, result, runtime, staged, writeErr, func(err error) string {
		return connectors.SanitizeWriteErrorForOutput(err, runtime.Secrets)
	})
}

func (a *App) finishReverseWriteWithErrorText(planID string, run ReverseRun, result connectors.WriteResult, runtime connectors.RuntimeConfig, staged int, writeErr error, errorText func(error) string) (ReverseRun, error) {
	output, outputErr := json.Marshal(connectors.SanitizeWriteResultForOutput(result, runtime.Secrets))
	if outputErr != nil && writeErr == nil {
		writeErr = fmt.Errorf("encode complete reverse destination result: %w", outputErr)
	}
	if outputErr == nil {
		run.DestinationResult = append(json.RawMessage(nil), output...)
	}
	run.RecordsSucceeded = result.RecordsWritten
	run.RecordsFailed = result.RecordsFailed
	run.CompletedAt = time.Now().UTC()
	planStatus := "executed"
	if writeErr != nil {
		run.Status = "failed"
		planStatus = "failed"
		if run.RecordsFailed == 0 {
			run.RecordsFailed = staged - result.RecordsWritten
		}
		run.Error = errorText(writeErr)
	} else {
		run.Status = "completed"
	}
	_, persistErr := a.updateState(func(current state) (state, error) {
		current.ReverseRuns = append(current.ReverseRuns, run)
		for i := range current.ReversePlans {
			if current.ReversePlans[i].ID == planID && (current.ReversePlans[i].Status == "executing" || current.ReversePlans[i].Status == reversePlanStatusApprovalConsumptionUncertain) {
				current.ReversePlans[i].Status = planStatus
				current.ReversePlans[i].ApprovalUncertainAt = time.Time{}
				break
			}
		}
		return current, nil
	})
	if persistErr != nil {
		if !stateStoreCommitMayHaveSucceeded(persistErr) {
			return ReverseRun{}, errors.Join(writeErr, persistErr)
		}
		reloaded, reloadErr := a.reloadExactTerminalState(persistErr)
		if reloadErr != nil {
			return ReverseRun{}, errors.Join(writeErr, persistErr, reloadErr)
		}
		restored, restoreErr := terminalReverseRunFromState(reloaded.State, planID, run.ID)
		if restoreErr != nil {
			return ReverseRun{}, errors.Join(writeErr, persistErr, restoreErr)
		}
		return restored, errors.Join(writeErr, persistErr)
	}
	return run, writeErr
}

type runtimeSecretSanitizedError struct {
	cause   error
	message string
}

func (e *runtimeSecretSanitizedError) Error() string {
	if e == nil {
		return ""
	}
	return e.message
}

func (e *runtimeSecretSanitizedError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.cause
}

func sanitizeRuntimeError(err error, runtimes ...connectors.RuntimeConfig) error {
	if err == nil {
		return nil
	}
	secrets := make(map[string]string)
	for _, runtime := range runtimes {
		for name, value := range runtime.Secrets {
			secrets[name] = value
		}
	}
	return &runtimeSecretSanitizedError{cause: err, message: connectors.SanitizeWriteErrorForOutput(err, secrets)}
}

func (a *App) invalidateReversePlan(expected ReversePlan) error {
	now := time.Now().UTC()
	updated, err := a.updateStateAfterPreflight(func(current state) error {
		_, preflightErr := reversePlanInvalidationCandidate(current, expected)
		return preflightErr
	}, func(current state) (state, error) {
		return invalidateReversePlanState(current, expected, now)
	})
	if err != nil {
		return err
	}
	a.state = updated
	return nil
}

func reversePlanInvalidationCandidate(current state, expected ReversePlan) (int, error) {
	for i := range current.ReversePlans {
		stored := current.ReversePlans[i]
		if stored.ID != expected.ID {
			continue
		}
		if err := approvalConsumptionUncertainError(stored, nil); err != nil {
			return 0, err
		}
		if stored.ApprovalTokenHash == "" {
			return 0, errors.New("reverse plan approval has already been consumed")
		}
		if !reversePlanMatchesExpected(stored, expected) || !constantTimeStringEqual(stored.ApprovalTokenHash, expected.ApprovalTokenHash) {
			return 0, fmt.Errorf("reverse plan %q changed before approval consumption", stored.ID)
		}
		return i, nil
	}
	return 0, fmt.Errorf("reverse plan %q not found", expected.ID)
}

func invalidateReversePlanState(current state, expected ReversePlan, now time.Time) (state, error) {
	i, err := reversePlanInvalidationCandidate(current, expected)
	if err != nil {
		return current, err
	}
	stored := current.ReversePlans[i]
	stored.Status = "invalidated"
	stored.ApprovalTokenHash = ""
	stored.ApprovalGrant = nil
	stored.ApprovalConsumedAt = now
	current.ReversePlans[i] = stored
	return current, nil
}

func constantTimeStringEqual(left, right string) bool {
	return subtle.ConstantTimeCompare([]byte(left), []byte(right)) == 1
}

func (a *App) resolveEndpoint(ctx context.Context, endpoint EndpointConfig) (connectors.Connector, connectors.RuntimeConfig, error) {
	connector, _, runtime, err := a.resolveEndpointWithCredential(ctx, endpoint)
	return connector, runtime, err
}

func (a *App) resolveEndpointWithCredential(ctx context.Context, endpoint EndpointConfig) (connectors.Connector, CredentialMeta, connectors.RuntimeConfig, error) {
	if err := connectors.RejectLegacyConnectorName(endpoint.Connector); err != nil {
		return nil, CredentialMeta{}, connectors.RuntimeConfig{}, err
	}
	cred, runtime, err := a.resolveCredential(ctx, endpoint.Credential, endpoint.Config)
	if err != nil {
		return nil, CredentialMeta{}, connectors.RuntimeConfig{}, err
	}
	if err := connectors.RejectLegacyConnectorName(cred.Connector); err != nil {
		return nil, CredentialMeta{}, connectors.RuntimeConfig{}, err
	}
	if endpoint.Connector != "" && endpoint.Connector != cred.Connector {
		return nil, CredentialMeta{}, connectors.RuntimeConfig{}, fmt.Errorf("credential %q is for connector %q, not %q", endpoint.Credential, cred.Connector, endpoint.Connector)
	}
	connector, ok := a.registry.Get(cred.Connector)
	if !ok {
		return nil, CredentialMeta{}, connectors.RuntimeConfig{}, fmt.Errorf("connector %q not found", cred.Connector)
	}
	return connector, cred, runtime, nil
}

func (a *App) ResolveConnectorCredential(ctx context.Context, connectorName, credentialName string, overlay map[string]string) (connectors.Connector, connectors.RuntimeConfig, error) {
	if strings.TrimSpace(credentialName) == "" {
		return nil, connectors.RuntimeConfig{}, errors.New("missing --credential")
	}
	return a.resolveEndpoint(ctx, EndpointConfig{
		Connector:  connectorName,
		Credential: credentialName,
		Config:     overlay,
	})
}

func (a *App) resolveCredential(ctx context.Context, name string, overlay map[string]string) (CredentialMeta, connectors.RuntimeConfig, error) {
	var parkingAdmission connectors.RateParkingAdmission = a.rateParking
	if isRateParkingResume(ctx) {
		parkingAdmission = nil
	}
	resolveAuthAdmission := func(identity connectors.CoordinationIdentity) (connectors.AuthenticationAdmission, error) {
		if isAuthRepair(ctx) {
			return nil, nil
		}
		return coordination.NewAuthCohortRuntime(ctx, a.authCohorts, identity.AuthCohortKey())
	}
	if a.ephemeralCredentials != nil {
		credential, secrets, ok := a.ephemeralCredentials.credential(name)
		if ok {
			providerFamily, authProfile, err := credentialCoordinationDeclarations(credential.Connector, "", "")
			if err != nil {
				return CredentialMeta{}, connectors.RuntimeConfig{}, err
			}
			credential.ProviderFamily = providerFamily
			credential.AuthProfile = authProfile
			coordinationIdentity, err := a.newCoordinationIdentity(providerFamily, authProfile, credential.ID)
			if err != nil {
				return CredentialMeta{}, connectors.RuntimeConfig{}, err
			}
			config := cloneStringMap(credential.Config)
			for key, value := range overlay {
				config[key] = value
			}
			credentialRevision, err := a.approval.CredentialRevision(credential.ID, secrets)
			if err != nil {
				return CredentialMeta{}, connectors.RuntimeConfig{}, err
			}
			configurationDigest, err := a.approval.ConfigurationDigest(credential.ID, config)
			if err != nil {
				return CredentialMeta{}, connectors.RuntimeConfig{}, err
			}
			authAdmission, err := resolveAuthAdmission(coordinationIdentity)
			if err != nil {
				return CredentialMeta{}, connectors.RuntimeConfig{}, err
			}
			return credential, connectors.RuntimeConfig{
				ProjectDir:              a.projectDir,
				Config:                  config,
				Secrets:                 secrets,
				CoordinationIdentity:    coordinationIdentity,
				CredentialRevision:      credentialRevision,
				ConfigurationDigest:     configurationDigest,
				WriteApprovalScope:      connectors.WriteApprovalScopeProject,
				SecretStore:             ephemeralCertificationSecretStore{},
				AuthenticationAdmission: authAdmission,
				RateParkingAdmission:    parkingAdmission,
			}, nil
		}
	}
	cred, ok := a.findCredential(name)
	if !ok {
		return CredentialMeta{}, connectors.RuntimeConfig{}, fmt.Errorf("credential %q not found", name)
	}
	coordinationIdentity, err := a.coordinationIdentityForCredential(cred)
	if err != nil {
		return CredentialMeta{}, connectors.RuntimeConfig{}, err
	}
	secrets, err := a.vault.Get(ctx, cred.ID)
	if err != nil {
		return CredentialMeta{}, connectors.RuntimeConfig{}, err
	}
	config := cloneStringMap(cred.Config)
	for k, v := range overlay {
		config[k] = v
	}
	credentialRevision, err := a.approval.CredentialRevision(cred.ID, secrets)
	if err != nil {
		return CredentialMeta{}, connectors.RuntimeConfig{}, err
	}
	configurationDigest, err := a.approval.ConfigurationDigest(cred.ID, config)
	if err != nil {
		return CredentialMeta{}, connectors.RuntimeConfig{}, err
	}
	authAdmission, err := resolveAuthAdmission(coordinationIdentity)
	if err != nil {
		return CredentialMeta{}, connectors.RuntimeConfig{}, err
	}
	return cred, connectors.RuntimeConfig{
		ProjectDir:           a.projectDir,
		Config:               config,
		Secrets:              secrets,
		CoordinationIdentity: coordinationIdentity,
		CredentialRevision:   credentialRevision,
		ConfigurationDigest:  configurationDigest,
		WriteApprovalScope:   connectors.WriteApprovalScopeProject,
		// Scoped to this credential, so a provider-rotated secret (an OAuth2
		// refresh token) is written back to the same encrypted vault entry it
		// was read from, and to no other.
		SecretStore:             a.credentialSecretStore(cred.ID),
		AuthenticationAdmission: authAdmission,
		RateParkingAdmission:    parkingAdmission,
	}, nil
}

func (a *App) findCredential(name string) (CredentialMeta, bool) {
	if a.ephemeralCredentials != nil {
		if credential, _, ok := a.ephemeralCredentials.credential(name); ok {
			return credential, true
		}
	}
	for _, cred := range a.state.Credentials {
		if cred.Name == name || cred.ID == name {
			return cred, true
		}
	}
	return CredentialMeta{}, false
}

func (a *App) findConnection(name string) (Connection, bool) {
	for _, conn := range a.state.Connections {
		if conn.Name == name {
			return cloneConnection(conn), true
		}
	}
	return Connection{}, false
}

// GetConnection refreshes project state and positively resolves one persisted
// ETL connection by name. It never opens the credential vault.
func (a *App) GetConnection(name string) (Connection, error) {
	if a == nil {
		return Connection{}, errors.New("app is required")
	}
	loaded, err := a.store.LoadReadOnly()
	if err != nil {
		return Connection{}, err
	}
	if err := a.normalizeLoadedState(loaded, false); err != nil {
		return Connection{}, err
	}
	connection, ok := a.findConnection(name)
	if !ok {
		return Connection{}, fmt.Errorf("connection %q not found", name)
	}
	return connection, nil
}

func (a *App) findConnectionByID(id string) (Connection, bool) {
	for _, conn := range a.state.Connections {
		if conn.ID == id {
			return cloneConnection(conn), true
		}
	}
	return Connection{}, false
}

// findConnectionFold finds a connection whose name differs from name only by
// letter case. Case alone is too weak a distinction to hang two tenants'
// warehouse data on, so creation refuses it.
func (a *App) findConnectionFold(name string) (Connection, bool) {
	for _, conn := range a.state.Connections {
		if strings.EqualFold(conn.Name, name) {
			return cloneConnection(conn), true
		}
	}
	return Connection{}, false
}

// failRun keeps persisted run state truthful: the typed transport-conflict path
// must not return the JSON store's speculative callback state after a definite
// pre-rename failure.
func (a *App) failRun(runID string, runErr error) (Run, error) {
	return a.failRunWithResult(runID, etlExecutionResult{}, runErr)
}

func (a *App) failRunWithResult(runID string, result etlExecutionResult, runErr error) (Run, error) {
	expectedRevision := a.state.Revision
	completedAt := time.Now().UTC()
	transportStateConflict := errors.Is(runErr, errTransportStreamStateConflict)
	transitionedInCallback := false
	targetAlreadyTerminal := false
	updated, persistErr := a.updateState(func(current state) (state, error) {
		if !transportStateConflict && current.Revision != expectedRevision {
			return current, errStateRevisionConflict
		}
		for i := range current.Runs {
			if current.Runs[i].ID != runID {
				continue
			}
			if transportStateConflict && current.Runs[i].Status != "running" {
				targetAlreadyTerminal = true
				return current, fmt.Errorf("transport conflict run %q has status %q, want running before finalization", runID, current.Runs[i].Status)
			}
			current.Runs[i].Status = "failed"
			current.Runs[i].RecordsRead = result.RecordsRead
			current.Runs[i].RecordsTransformed = result.RecordsTransformed
			current.Runs[i].RecordsLoaded = result.RecordsLoaded
			current.Runs[i].RecordsFailed = result.RecordsFailed
			current.Runs[i].BatchCount = result.BatchCount
			current.Runs[i].Checkpoint = cloneStringMap(result.Checkpoint)
			current.Runs[i].TransportPhaseMeasurement = cloneTransportPhaseMeasurement(result.TransportPhaseMeasurement)
			current.Runs[i].DestinationResults = cloneDestinationResults(result.DestinationResults)
			current.Runs[i].Error = safety.RedactErrorText(runErr.Error())
			current.Runs[i].CompletedAt = completedAt
			transitionedInCallback = true
			return current, nil
		}
		return current, fmt.Errorf("run %q not found", runID)
	})
	if persistErr != nil {
		if transportStateConflict && targetAlreadyTerminal {
			if run, storedErr := terminalETLRunFromState(updated, runID); storedErr == nil {
				return run, errors.Join(runErr, fmt.Errorf("persist failed ETL run: %w", persistErr))
			} else {
				return Run{}, errors.Join(runErr, fmt.Errorf("persist failed ETL run: %w", errors.Join(persistErr, storedErr)))
			}
		}
		if transitionedInCallback && stateStoreCommitMayHaveSucceeded(persistErr) {
			reloaded, reloadErr := a.reloadExactTerminalState(persistErr)
			if reloadErr != nil {
				return Run{}, errors.Join(runErr, fmt.Errorf("persist failed ETL run: %w", errors.Join(persistErr, reloadErr)))
			}
			run, storedErr := terminalETLRunFromState(reloaded.State, runID)
			if storedErr != nil {
				return Run{}, errors.Join(runErr, fmt.Errorf("persist failed ETL run: %w", errors.Join(persistErr, storedErr)))
			}
			return run, errors.Join(runErr, fmt.Errorf("persist failed ETL run: %w", persistErr))
		}
		return Run{}, errors.Join(runErr, fmt.Errorf("persist failed ETL run: %w", persistErr))
	}
	for _, run := range updated.Runs {
		if run.ID == runID {
			return run, runErr
		}
	}
	return Run{}, errors.Join(runErr, fmt.Errorf("failed run %q was not stored", runID))
}

func writeJSONAtomic(path string, v any) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	tmp := path + ".tmp"
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal json: %w", err)
	}
	b = append(b, '\n')
	if err := os.WriteFile(tmp, b, 0o600); err != nil {
		return fmt.Errorf("write temp json: %w", err)
	}
	if err := os.Rename(tmp, path); err != nil {
		return fmt.Errorf("rename temp json: %w", err)
	}
	return nil
}

func prefixedID(prefix string) (string, error) {
	token, err := randomToken(8)
	if err != nil {
		return "", err
	}
	return prefix + "_" + token, nil
}

func allocateUniquePrefixedID(prefix string, used map[string]struct{}) (string, error) {
	return allocateUniqueIdentity(prefix, used, func() (string, error) {
		return prefixedID(prefix)
	})
}

func allocateUniqueIdentity(kind string, used map[string]struct{}, generate func() (string, error)) (string, error) {
	const maxAttempts = 32
	for attempt := 0; attempt < maxAttempts; attempt++ {
		identity, err := generate()
		if err != nil {
			return "", err
		}
		if _, exists := used[identity]; exists {
			continue
		}
		used[identity] = struct{}{}
		return identity, nil
	}
	return "", fmt.Errorf("allocate unique %s identity: too many collisions", kind)
}

func randomToken(bytes int) (string, error) {
	b := make([]byte, bytes)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generate random token: %w", err)
	}
	return hex.EncodeToString(b), nil
}

// Package pmbroker contains CLI-side PM Broker context, profile, and
// contract-version seams. It is deliberately metadata-only: live broker clients,
// provider SDKs, credential storage, and provider operations are out of scope.
package pmbroker

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"unicode"
)

const (
	// CurrentStateVersion is the version for the safe PM Broker user-context state file.
	CurrentStateVersion = 1
	// LegacyLocalContextName is synthesized for unmigrated local projects.
	LegacyLocalContextName = "legacy-local"
)

// Immutable ID aliases mirror the PM Broker /v1 contract identity boundaries.
type (
	OrganizationID  string
	WorkspaceID     string
	EnvironmentID   string
	BrokerProfileID string
)

var (
	organizationIDPattern  = regexp.MustCompile(`^org_[a-z0-9]{16,32}$`)
	workspaceIDPattern     = regexp.MustCompile(`^wks_[a-z0-9]{16,32}$`)
	environmentIDPattern   = regexp.MustCompile(`^env_[a-z0-9]{16,32}$`)
	brokerProfileIDPattern = regexp.MustCompile(`^bpf_[a-z0-9]{16,32}$`)
	contextNamePattern     = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._-]{0,63}$`)
	policyBindingPattern   = regexp.MustCompile(`^policy_[A-Za-z0-9_-]{3,128}$`)
)

var (
	// ErrInvalidIdentityBoundary means an Organization, Workspace, Environment, or BrokerProfile tuple is malformed.
	ErrInvalidIdentityBoundary = errors.New("pmbroker: invalid identity boundary")
	// ErrInvalidContextName means a named PM Broker context is absent or malformed.
	ErrInvalidContextName = errors.New("pmbroker: invalid context name")
	// ErrContextNotFound means context resolution named a context that is not in safe user state.
	ErrContextNotFound = errors.New("pmbroker: context not found")
	// ErrContextMismatch means two context requirements disagree and execution must stop safely.
	ErrContextMismatch = errors.New("pmbroker: context mismatch")
	// ErrDuplicateContext means a user state file contains duplicate context names.
	ErrDuplicateContext = errors.New("pmbroker: duplicate context")
	// ErrInvalidRuntimeMode means a runtime mode is not one of remote, local, or hybrid.
	ErrInvalidRuntimeMode = errors.New("pmbroker: invalid runtime mode")
	// ErrInvalidRuntimeOperation means a runtime-mode policy operation is unknown.
	ErrInvalidRuntimeOperation = errors.New("pmbroker: invalid runtime operation")
	// ErrHybridPolicyRequired means hybrid mode was selected without a policy binding.
	ErrHybridPolicyRequired = errors.New("pmbroker: hybrid runtime mode requires policy binding")
	// ErrInvalidPolicyBindingID means a runtime policy binding identifier is malformed.
	ErrInvalidPolicyBindingID = errors.New("pmbroker: invalid hybrid policy binding")
	// ErrProductionLocalFallbackForbidden means a production write or schedule tried to allow a local fallback.
	ErrProductionLocalFallbackForbidden = errors.New("pmbroker: production writes and scheduled jobs cannot use local fallback")
	// ErrUnsafeState means a persisted PM Broker state file contains an unsafe or unknown field.
	ErrUnsafeState = errors.New("pmbroker: unsafe user state")
)

// IsValid reports whether id matches the PM Broker Organization ID shape.
func (id OrganizationID) IsValid() bool { return organizationIDPattern.MatchString(string(id)) }

// IsValid reports whether id matches the PM Broker Workspace ID shape.
func (id WorkspaceID) IsValid() bool { return workspaceIDPattern.MatchString(string(id)) }

// IsValid reports whether id matches the PM Broker Environment ID shape.
func (id EnvironmentID) IsValid() bool { return environmentIDPattern.MatchString(string(id)) }

// IsValid reports whether id matches the PM Broker BrokerProfile ID shape.
func (id BrokerProfileID) IsValid() bool { return brokerProfileIDPattern.MatchString(string(id)) }

// Organization is safe, server-authorized organization metadata cached by the CLI.
type Organization struct {
	ID          OrganizationID `json:"organization_id"`
	DisplayName string         `json:"display_name"`
}

// Workspace is safe, server-authorized workspace metadata cached by the CLI.
type Workspace struct {
	ID             WorkspaceID    `json:"workspace_id"`
	OrganizationID OrganizationID `json:"organization_id"`
	DisplayName    string         `json:"display_name"`
}

// EnvironmentType classifies a PM Broker environment without granting authority by name.
type EnvironmentType string

const (
	EnvironmentTypeDevelopment EnvironmentType = "development"
	EnvironmentTypeStaging     EnvironmentType = "staging"
	EnvironmentTypeProduction  EnvironmentType = "production"
	EnvironmentTypeEphemeral   EnvironmentType = "ephemeral"
)

// Environment is safe, server-authorized environment metadata cached by the CLI.
type Environment struct {
	ID             EnvironmentID   `json:"environment_id"`
	WorkspaceID    WorkspaceID     `json:"workspace_id"`
	OrganizationID OrganizationID  `json:"organization_id"`
	DisplayName    string          `json:"display_name"`
	Type           EnvironmentType `json:"environment_type"`
}

// BrokerProfile is safe profile metadata for a PM Broker identity boundary.
type BrokerProfile struct {
	ID             BrokerProfileID `json:"broker_profile_id"`
	OrganizationID OrganizationID  `json:"organization_id"`
	WorkspaceID    WorkspaceID     `json:"workspace_id"`
	EnvironmentID  EnvironmentID   `json:"environment_id"`
	DisplayName    string          `json:"display_name"`
}

// RuntimeMode chooses where PM Broker-backed work is evaluated.
type RuntimeMode string

const (
	RuntimeModeRemote RuntimeMode = "remote"
	RuntimeModeLocal  RuntimeMode = "local"
	RuntimeModeHybrid RuntimeMode = "hybrid"
)

// RuntimeOperation describes the policy-sensitive operation being selected.
type RuntimeOperation string

const (
	RuntimeOperationRead         RuntimeOperation = "read"
	RuntimeOperationWrite        RuntimeOperation = "write"
	RuntimeOperationScheduledJob RuntimeOperation = "scheduled_job"
)

// RuntimeModeSelection is stored on a context. Hybrid requires PolicyBindingID.
type RuntimeModeSelection struct {
	Mode            RuntimeMode `json:"mode"`
	PolicyBindingID string      `json:"policy_binding_id,omitempty"`
}

// RuntimeSelection validates a mode for one concrete operation.
type RuntimeSelection struct {
	Mode            RuntimeMode      `json:"mode"`
	Environment     EnvironmentType  `json:"environment_type"`
	Operation       RuntimeOperation `json:"operation"`
	PolicyBindingID string           `json:"policy_binding_id,omitempty"`
}

// Context names a tuple of Organization, Workspace, Environment, and BrokerProfile.
type Context struct {
	Name          string               `json:"name"`
	Organization  Organization         `json:"organization"`
	Workspace     Workspace            `json:"workspace"`
	Environment   Environment          `json:"environment"`
	BrokerProfile BrokerProfile        `json:"broker_profile"`
	Runtime       RuntimeModeSelection `json:"runtime"`
}

// UserState is the versioned safe user-level PM Broker context state.
type UserState struct {
	Version       int       `json:"version"`
	ActiveContext string    `json:"active_context,omitempty"`
	Contexts      []Context `json:"contexts,omitempty"`
}

// DefaultRuntimeMode returns the default runtime mode for an environment.
func DefaultRuntimeMode(_ EnvironmentType) RuntimeMode {
	return RuntimeModeRemote
}

// ParseRuntimeMode validates a user-facing runtime mode string.
func ParseRuntimeMode(raw string) (RuntimeMode, error) {
	mode := RuntimeMode(raw)
	switch mode {
	case RuntimeModeRemote, RuntimeModeLocal, RuntimeModeHybrid:
		return mode, nil
	default:
		return "", ErrInvalidRuntimeMode
	}
}

// ParseEnvironmentType validates a user-facing environment type string.
func ParseEnvironmentType(raw string) (EnvironmentType, error) {
	envType := EnvironmentType(strings.ToLower(strings.TrimSpace(raw)))
	if envType.IsValid() {
		return envType, nil
	}
	return "", ErrInvalidIdentityBoundary
}

// IsValid reports whether t is a supported environment type.
func (t EnvironmentType) IsValid() bool {
	switch t {
	case EnvironmentTypeDevelopment, EnvironmentTypeStaging, EnvironmentTypeProduction, EnvironmentTypeEphemeral:
		return true
	default:
		return false
	}
}

// IsValid reports whether op is a supported runtime policy operation.
func (op RuntimeOperation) IsValid() bool {
	switch op {
	case RuntimeOperationRead, RuntimeOperationWrite, RuntimeOperationScheduledJob:
		return true
	default:
		return false
	}
}

// Validate applies PM Broker runtime-mode safety policy.
func (s RuntimeSelection) Validate() error {
	mode, err := ParseRuntimeMode(string(s.Mode))
	if err != nil {
		return err
	}
	if !s.Environment.IsValid() {
		return ErrInvalidIdentityBoundary
	}
	if !s.Operation.IsValid() {
		return ErrInvalidRuntimeOperation
	}
	if s.PolicyBindingID != "" && !validPolicyBindingID(s.PolicyBindingID) {
		return ErrInvalidPolicyBindingID
	}
	if mode == RuntimeModeHybrid && s.PolicyBindingID == "" {
		return ErrHybridPolicyRequired
	}
	if s.Environment == EnvironmentTypeProduction && (s.Operation == RuntimeOperationWrite || s.Operation == RuntimeOperationScheduledJob) && mode != RuntimeModeRemote {
		return ErrProductionLocalFallbackForbidden
	}
	return nil
}

// Validate checks that the context runtime mode can be stored safely.
func (s RuntimeModeSelection) Validate(envType EnvironmentType) error {
	return RuntimeSelection{Mode: s.Mode, Environment: envType, Operation: RuntimeOperationRead, PolicyBindingID: s.PolicyBindingID}.Validate()
}

// Validate checks safe Organization metadata.
func (o Organization) Validate() error {
	if !o.ID.IsValid() || !safeDisplayName(o.DisplayName) {
		return ErrInvalidIdentityBoundary
	}
	return nil
}

// Validate checks safe Workspace metadata and its Organization boundary.
func (w Workspace) Validate() error {
	if !w.ID.IsValid() || !w.OrganizationID.IsValid() || !safeDisplayName(w.DisplayName) {
		return ErrInvalidIdentityBoundary
	}
	return nil
}

// Validate checks safe Environment metadata and its parent boundaries.
func (e Environment) Validate() error {
	if !e.ID.IsValid() || !e.WorkspaceID.IsValid() || !e.OrganizationID.IsValid() || !safeDisplayName(e.DisplayName) || !e.Type.IsValid() {
		return ErrInvalidIdentityBoundary
	}
	return nil
}

// Validate checks safe BrokerProfile metadata and its identity boundary.
func (p BrokerProfile) Validate() error {
	if !p.ID.IsValid() || !p.OrganizationID.IsValid() || !p.WorkspaceID.IsValid() || !p.EnvironmentID.IsValid() || !safeDisplayName(p.DisplayName) {
		return ErrInvalidIdentityBoundary
	}
	return nil
}

// ValidateContextName checks the user-facing PM Broker context name shape.
func ValidateContextName(name string) error {
	if name != strings.TrimSpace(name) || !contextNamePattern.MatchString(name) {
		return ErrInvalidContextName
	}
	return nil
}

// Validate checks tuple consistency for a named context.
func (c Context) Validate() error {
	if err := ValidateContextName(c.Name); err != nil {
		return err
	}
	if err := c.Organization.Validate(); err != nil {
		return err
	}
	if err := c.Workspace.Validate(); err != nil {
		return err
	}
	if err := c.Environment.Validate(); err != nil {
		return err
	}
	if err := c.BrokerProfile.Validate(); err != nil {
		return err
	}
	if c.Workspace.OrganizationID != c.Organization.ID || c.Environment.OrganizationID != c.Organization.ID || c.Environment.WorkspaceID != c.Workspace.ID || c.BrokerProfile.OrganizationID != c.Organization.ID || c.BrokerProfile.WorkspaceID != c.Workspace.ID || c.BrokerProfile.EnvironmentID != c.Environment.ID {
		return ErrInvalidIdentityBoundary
	}
	return c.Runtime.Validate(c.Environment.Type)
}

// Validate checks state version, unique names, active context, and nested contexts.
func (s UserState) Validate() error {
	if s.Version != CurrentStateVersion {
		return fmt.Errorf("%w: unsupported or missing version %d; initialize current schema with pm context create or migrate the state file explicitly", ErrUnsafeState, s.Version)
	}
	seen := map[string]bool{}
	activeFound := s.ActiveContext == ""
	for _, ctx := range s.Contexts {
		if err := ctx.Validate(); err != nil {
			return err
		}
		if seen[ctx.Name] {
			return ErrDuplicateContext
		}
		seen[ctx.Name] = true
		if ctx.Name == s.ActiveContext {
			activeFound = true
		}
	}
	if !activeFound {
		return ErrContextNotFound
	}
	return nil
}

func (s UserState) hasData() bool {
	return s.Version != 0 || s.ActiveContext != "" || len(s.Contexts) > 0
}

// ContextByName returns a context by name.
func (s UserState) ContextByName(name string) (Context, bool) {
	for _, ctx := range s.Contexts {
		if ctx.Name == name {
			return ctx, true
		}
	}
	return Context{}, false
}

// Organizations returns de-duplicated cached organization metadata.
func (s UserState) Organizations() []Organization {
	seen := map[OrganizationID]Organization{}
	for _, ctx := range s.Contexts {
		seen[ctx.Organization.ID] = ctx.Organization
	}
	out := make([]Organization, 0, len(seen))
	for _, org := range seen {
		out = append(out, org)
	}
	sort.Slice(out, func(i, j int) bool { return string(out[i].ID) < string(out[j].ID) })
	return out
}

// Workspaces returns de-duplicated cached workspace metadata.
func (s UserState) Workspaces(organizationID OrganizationID) []Workspace {
	seen := map[WorkspaceID]Workspace{}
	for _, ctx := range s.Contexts {
		if organizationID != "" && ctx.Workspace.OrganizationID != organizationID {
			continue
		}
		seen[ctx.Workspace.ID] = ctx.Workspace
	}
	out := make([]Workspace, 0, len(seen))
	for _, workspace := range seen {
		out = append(out, workspace)
	}
	sort.Slice(out, func(i, j int) bool { return string(out[i].ID) < string(out[j].ID) })
	return out
}

// Environments returns de-duplicated cached environment metadata.
func (s UserState) Environments(workspaceID WorkspaceID) []Environment {
	seen := map[EnvironmentID]Environment{}
	for _, ctx := range s.Contexts {
		if workspaceID != "" && ctx.Environment.WorkspaceID != workspaceID {
			continue
		}
		seen[ctx.Environment.ID] = ctx.Environment
	}
	out := make([]Environment, 0, len(seen))
	for _, env := range seen {
		out = append(out, env)
	}
	sort.Slice(out, func(i, j int) bool { return string(out[i].ID) < string(out[j].ID) })
	return out
}

// Store persists safe PM Broker user context state.
type Store struct {
	Path string
}

// DefaultUserStatePath returns an OS-standard Polymetrics PM Broker user config path.
func DefaultUserStatePath() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("resolve user config directory: %w", err)
	}
	return filepath.Join(dir, "polymetrics", "pm-broker", "contexts.json"), nil
}

// Load returns safe user state, or an empty versioned state when no file exists.
func (s Store) Load() (UserState, error) {
	if s.Path == "" {
		return UserState{}, errors.New("pmbroker: state path is required")
	}
	info, err := os.Lstat(s.Path)
	if errors.Is(err, os.ErrNotExist) {
		return UserState{Version: CurrentStateVersion}, nil
	}
	if err != nil {
		return UserState{}, fmt.Errorf("stat pm broker context state: %w", err)
	}
	if !info.Mode().IsRegular() || info.Mode().Perm()&0o022 != 0 {
		return UserState{}, fmt.Errorf("%w: state file must be regular and not group/world-writable", ErrUnsafeState)
	}
	file, err := os.Open(s.Path)
	if err != nil {
		return UserState{}, fmt.Errorf("open pm broker context state: %w", err)
	}
	state, err := decodeUserState(file)
	if err != nil {
		_ = file.Close()
		return UserState{}, err
	}
	if err := file.Close(); err != nil {
		return UserState{}, fmt.Errorf("close pm broker context state: %w", err)
	}
	if err := state.Validate(); err != nil {
		return UserState{}, err
	}
	return state, nil
}

// Save writes safe user state with owner-readable permissions.
func (s Store) Save(state UserState) error {
	if s.Path == "" {
		return errors.New("pmbroker: state path is required")
	}
	if err := state.Validate(); err != nil {
		return err
	}
	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return fmt.Errorf("encode pm broker context state: %w", err)
	}
	data = append(data, '\n')
	dir := filepath.Dir(s.Path)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return fmt.Errorf("create pm broker context state directory: %w", err)
	}
	tmp, err := os.CreateTemp(dir, ".contexts.json.tmp-*")
	if err != nil {
		return fmt.Errorf("create pm broker context temp file: %w", err)
	}
	tmpPath := tmp.Name()
	defer func() { _ = os.Remove(tmpPath) }()
	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()
		return fmt.Errorf("write pm broker context temp file: %w", err)
	}
	if err := tmp.Chmod(0o600); err != nil {
		_ = tmp.Close()
		return fmt.Errorf("chmod pm broker context temp file: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("close pm broker context temp file: %w", err)
	}
	if err := os.Rename(tmpPath, s.Path); err != nil {
		return fmt.Errorf("replace pm broker context state: %w", err)
	}
	return nil
}

// UpsertContext inserts or replaces one context by name.
func (s Store) UpsertContext(ctx Context) (UserState, error) {
	if err := ctx.Validate(); err != nil {
		return UserState{}, err
	}
	state, err := s.Load()
	if err != nil {
		return UserState{}, err
	}
	replaced := false
	for i := range state.Contexts {
		if state.Contexts[i].Name == ctx.Name {
			state.Contexts[i] = ctx
			replaced = true
			break
		}
	}
	if !replaced {
		state.Contexts = append(state.Contexts, ctx)
	}
	if state.ActiveContext == "" {
		state.ActiveContext = ctx.Name
	}
	if err := s.Save(state); err != nil {
		return UserState{}, err
	}
	return state, nil
}

// UseContext marks name as the active safe user context.
func (s Store) UseContext(name string) (UserState, error) {
	state, err := s.Load()
	if err != nil {
		return UserState{}, err
	}
	if _, ok := state.ContextByName(name); !ok {
		return UserState{}, ErrContextNotFound
	}
	state.ActiveContext = name
	if err := s.Save(state); err != nil {
		return UserState{}, err
	}
	return state, nil
}

func decodeUserState(r io.Reader) (UserState, error) {
	decoder := json.NewDecoder(r)
	decoder.DisallowUnknownFields()
	var state UserState
	if err := decoder.Decode(&state); err != nil {
		return UserState{}, fmt.Errorf("%w: %v", ErrUnsafeState, err)
	}
	var extra struct{}
	if err := decoder.Decode(&extra); err == nil {
		return UserState{}, fmt.Errorf("%w: trailing JSON", ErrUnsafeState)
	} else if !errors.Is(err, io.EOF) {
		return UserState{}, fmt.Errorf("%w: %v", ErrUnsafeState, err)
	}
	return state, nil
}

// ResolveSource names the winning context-resolution source.
type ResolveSource string

const (
	ResolveSourceExplicit        ResolveSource = "explicit"
	ResolveSourceApprovalBound   ResolveSource = "approval_bound"
	ResolveSourceProjectRequired ResolveSource = "project_required"
	ResolveSourceActiveUser      ResolveSource = "active_user"
	ResolveSourceProjectDefault  ResolveSource = "project_default"
	ResolveSourceLegacyLocal     ResolveSource = "legacy_local"
)

// ResolveRequest contains all context candidates in precedence order.
type ResolveRequest struct {
	State                  UserState
	ExplicitContext        string
	ApprovalBoundContext   string
	ProjectRequiredContext string
	ProjectDefaultContext  string
	AllowLegacyLocal       bool
}

// ResolvedContext is the selected context plus its source.
type ResolvedContext struct {
	Context Context       `json:"context"`
	Source  ResolveSource `json:"source"`
}

// ResolveContext selects a context using the safe, documented precedence order.
func ResolveContext(req ResolveRequest) (ResolvedContext, error) {
	if req.State.hasData() {
		if err := req.State.Validate(); err != nil {
			return ResolvedContext{}, err
		}
	}
	if req.ApprovalBoundContext != "" {
		bound, err := contextByRequiredName(req.State, req.ApprovalBoundContext)
		if err != nil {
			return ResolvedContext{}, err
		}
		for _, candidate := range []string{req.ExplicitContext, req.ProjectRequiredContext, req.State.ActiveContext, req.ProjectDefaultContext} {
			if candidate == "" {
				continue
			}
			ctx, err := contextByRequiredName(req.State, candidate)
			if err != nil {
				return ResolvedContext{}, err
			}
			if ctx.identityBoundary() != bound.identityBoundary() {
				return ResolvedContext{}, ErrContextMismatch
			}
		}
		return ResolvedContext{Context: bound, Source: ResolveSourceApprovalBound}, nil
	}
	if req.ProjectRequiredContext != "" {
		for _, candidate := range []string{req.ExplicitContext} {
			if candidate != "" && candidate != req.ProjectRequiredContext {
				return ResolvedContext{}, ErrContextMismatch
			}
		}
	}
	candidates := []struct {
		name   string
		source ResolveSource
	}{
		{name: req.ExplicitContext, source: ResolveSourceExplicit},
		{name: req.ApprovalBoundContext, source: ResolveSourceApprovalBound},
		{name: req.ProjectRequiredContext, source: ResolveSourceProjectRequired},
		{name: req.State.ActiveContext, source: ResolveSourceActiveUser},
		{name: req.ProjectDefaultContext, source: ResolveSourceProjectDefault},
	}
	for _, candidate := range candidates {
		if candidate.name == "" {
			continue
		}
		ctx, err := contextByRequiredName(req.State, candidate.name)
		if err != nil {
			return ResolvedContext{}, err
		}
		return ResolvedContext{Context: ctx, Source: candidate.source}, nil
	}
	if req.AllowLegacyLocal {
		return ResolvedContext{Context: LegacyLocalContext(), Source: ResolveSourceLegacyLocal}, nil
	}
	return ResolvedContext{}, ErrContextNotFound
}

type contextIdentityBoundary struct {
	OrganizationID  OrganizationID
	WorkspaceID     WorkspaceID
	EnvironmentID   EnvironmentID
	BrokerProfileID BrokerProfileID
}

func (c Context) identityBoundary() contextIdentityBoundary {
	return contextIdentityBoundary{
		OrganizationID:  c.Organization.ID,
		WorkspaceID:     c.Workspace.ID,
		EnvironmentID:   c.Environment.ID,
		BrokerProfileID: c.BrokerProfile.ID,
	}
}

func contextByRequiredName(state UserState, name string) (Context, error) {
	ctx, ok := state.ContextByName(name)
	if !ok {
		return Context{}, ErrContextNotFound
	}
	if err := ctx.Validate(); err != nil {
		return Context{}, err
	}
	return ctx, nil
}

// LegacyLocalContext returns the synthesized local context for unmigrated projects.
func LegacyLocalContext() Context {
	return Context{
		Name:          LegacyLocalContextName,
		Organization:  Organization{ID: "org_legacylocal0000000000", DisplayName: "Legacy Local Organization"},
		Workspace:     Workspace{ID: "wks_legacylocal0000000000", OrganizationID: "org_legacylocal0000000000", DisplayName: "Legacy Local Workspace"},
		Environment:   Environment{ID: "env_legacylocal0000000000", WorkspaceID: "wks_legacylocal0000000000", OrganizationID: "org_legacylocal0000000000", DisplayName: "Legacy Local Environment", Type: EnvironmentTypeDevelopment},
		BrokerProfile: BrokerProfile{ID: "bpf_legacylocal0000000000", OrganizationID: "org_legacylocal0000000000", WorkspaceID: "wks_legacylocal0000000000", EnvironmentID: "env_legacylocal0000000000", DisplayName: "Legacy Local Broker Profile"},
		Runtime:       RuntimeModeSelection{Mode: RuntimeModeLocal},
	}
}

func safeDisplayName(value string) bool {
	value = strings.TrimSpace(value)
	if value == "" || len(value) > 128 {
		return false
	}
	for _, r := range value {
		if unicode.IsControl(r) {
			return false
		}
	}
	return true
}

func validPolicyBindingID(value string) bool {
	return policyBindingPattern.MatchString(value)
}

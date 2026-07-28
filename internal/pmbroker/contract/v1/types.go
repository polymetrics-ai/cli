package contractv1

import (
	"errors"
	"regexp"
	"strings"
	"time"
)

// ContractVersion is the negotiated PM Broker /v1 compatibility version.
type ContractVersion string

const (
	// ContractVersion1 is the only accepted version in the initial /v1 contract.
	ContractVersion1 ContractVersion = "1.0"
)

const (
	// HeaderAPIVersion is required on typed /v1 operations.
	HeaderAPIVersion = "PM-Broker-API-Version"
	// IncompatibleContractVersionMessage is the exact client-safe refusal message.
	IncompatibleContractVersionMessage = "requested contract version is not supported"
)

// ErrorCode is a stable PM Broker error identifier.
type ErrorCode string

const (
	// ErrorCodeIncompatibleContractVersion is returned with HTTP 426 on version refusal.
	ErrorCodeIncompatibleContractVersion ErrorCode = "incompatible_contract_version"
)

// OrganizationID identifies the top-level tenant and authorization boundary.
type OrganizationID string

// WorkspaceID identifies a workspace inside an organization.
type WorkspaceID string

// EnvironmentID identifies an environment inside a workspace.
type EnvironmentID string

// BrokerProfileID identifies a broker profile scoped to an identity boundary.
type BrokerProfileID string

// ConnectorConnectionID identifies a connector connection without exposing credentials.
type ConnectorConnectionID string

// ExecutionPlanID identifies an immutable execution plan.
type ExecutionPlanID string

// OpaqueSecretReferenceID identifies a broker-held secret reference, never raw secret bytes.
type OpaqueSecretReferenceID string

// ExecutionPlanDigest identifies the canonical digest for an execution plan.
type ExecutionPlanDigest string

// IdempotencyKey lets callers safely retry execution-plan creation.
type IdempotencyKey string

// CorrelationID is safe to return in errors and correlation logs.
type CorrelationID string

var (
	contractVersionPattern       = regexp.MustCompile(`^[0-9]+\.[0-9]+$`)
	organizationIDPattern        = regexp.MustCompile(`^org_[a-z0-9]{16,32}$`)
	workspaceIDPattern           = regexp.MustCompile(`^wks_[a-z0-9]{16,32}$`)
	environmentIDPattern         = regexp.MustCompile(`^env_[a-z0-9]{16,32}$`)
	brokerProfileIDPattern       = regexp.MustCompile(`^bpf_[a-z0-9]{16,32}$`)
	connectorConnectionIDPattern = regexp.MustCompile(`^ccn_[a-z0-9]{16,32}$`)
	executionPlanIDPattern       = regexp.MustCompile(`^epl_[a-z0-9]{16,32}$`)
	opaqueSecretRefIDPattern     = regexp.MustCompile(`^secretref_[a-z0-9]{16,32}$`)
	executionPlanDigestPattern   = regexp.MustCompile(`^sha256:[a-f0-9]{64}$`)
	idempotencyKeyPattern        = regexp.MustCompile(`^idem_[A-Za-z0-9_-]{20,128}$`)
	correlationIDPattern         = regexp.MustCompile(`^corr_[A-Za-z0-9_-]{16,64}$`)
	displayHintPattern           = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9 ._-]{0,79}$`)
)

var unsafeDisplayHintMarkers = []string{
	"access_token",
	"client_secret",
	"credential",
	"password",
	"private_key",
	"raw_secret",
	"refresh_token",
	"secret_value",
	"service_account",
	"token",
}

var (
	// ErrInvalidIdentityBoundary means an organization, workspace, or environment ID is malformed.
	ErrInvalidIdentityBoundary = errors.New("contractv1: invalid identity boundary")
	// ErrInvalidOpaqueSecretReference means a secret reference is malformed or not opaque.
	ErrInvalidOpaqueSecretReference = errors.New("contractv1: invalid opaque secret reference")
	// ErrUnsafeDisplayHint means a display hint contains control characters or secret-like markers.
	ErrUnsafeDisplayHint = errors.New("contractv1: unsafe display hint")
	// ErrRawSecretExportForbidden means a message attempted to mark a secret reference exportable.
	ErrRawSecretExportForbidden = errors.New("contractv1: raw secret export forbidden")
	// ErrInvalidExecutionIntent means an execution intent is absent, ambiguous, or malformed.
	ErrInvalidExecutionIntent = errors.New("contractv1: invalid execution intent")
	// ErrInvalidExecutionPlan means execution-plan metadata is malformed.
	ErrInvalidExecutionPlan = errors.New("contractv1: invalid execution plan")
	// ErrInvalidErrorResponse means a safe error response is malformed.
	ErrInvalidErrorResponse = errors.New("contractv1: invalid error response")
	// ErrUnexpectedResponse means the fake broker returned an unexpected status or payload.
	ErrUnexpectedResponse = errors.New("contractv1: unexpected broker response")
)

func (version ContractVersion) headerValue() (string, bool) {
	if version == "" {
		return "", false
	}
	value := string(version)
	if !contractVersionPattern.MatchString(value) {
		return "invalid", true
	}
	return value, true
}

// IsValid reports whether the organization ID matches the /v1 contract pattern.
func (id OrganizationID) IsValid() bool { return organizationIDPattern.MatchString(string(id)) }

// IsValid reports whether the workspace ID matches the /v1 contract pattern.
func (id WorkspaceID) IsValid() bool { return workspaceIDPattern.MatchString(string(id)) }

// IsValid reports whether the environment ID matches the /v1 contract pattern.
func (id EnvironmentID) IsValid() bool { return environmentIDPattern.MatchString(string(id)) }

// IsValid reports whether the broker profile ID matches the /v1 contract pattern.
func (id BrokerProfileID) IsValid() bool { return brokerProfileIDPattern.MatchString(string(id)) }

// IsValid reports whether the connector connection ID matches the /v1 contract pattern.
func (id ConnectorConnectionID) IsValid() bool {
	return connectorConnectionIDPattern.MatchString(string(id))
}

// IsValid reports whether the execution plan ID matches the /v1 contract pattern.
func (id ExecutionPlanID) IsValid() bool { return executionPlanIDPattern.MatchString(string(id)) }

// IsValid reports whether the opaque secret reference ID matches the /v1 contract pattern.
func (id OpaqueSecretReferenceID) IsValid() bool {
	return opaqueSecretRefIDPattern.MatchString(string(id))
}

// IsValid reports whether the execution plan digest is sha256-shaped.
func (digest ExecutionPlanDigest) IsValid() bool {
	return executionPlanDigestPattern.MatchString(string(digest))
}

// IsValid reports whether the idempotency key matches the /v1 contract pattern.
func (key IdempotencyKey) IsValid() bool { return idempotencyKeyPattern.MatchString(string(key)) }

// IsSafe reports whether the correlation ID is safe for client-visible errors.
func (id CorrelationID) IsSafe() bool { return correlationIDPattern.MatchString(string(id)) }

// Compatibility is the /v1 version-negotiation response.
type Compatibility struct {
	CurrentVersion             ContractVersion   `json:"current_version"`
	MinimumClientVersion       ContractVersion   `json:"minimum_client_version"`
	SupportedVersions          []ContractVersion `json:"supported_versions"`
	IncompatibleVersionRefusal SafeError         `json:"incompatible_version_refusal"`
}

// Organization is a PM Broker organization fixture.
type Organization struct {
	OrganizationID OrganizationID `json:"organization_id"`
	DisplayName    string         `json:"display_name"`
}

// Workspace is a PM Broker workspace fixture.
type Workspace struct {
	WorkspaceID    WorkspaceID    `json:"workspace_id"`
	OrganizationID OrganizationID `json:"organization_id"`
	DisplayName    string         `json:"display_name"`
}

// Environment is a PM Broker environment fixture.
type Environment struct {
	EnvironmentID   EnvironmentID  `json:"environment_id"`
	WorkspaceID     WorkspaceID    `json:"workspace_id"`
	OrganizationID  OrganizationID `json:"organization_id"`
	DisplayName     string         `json:"display_name"`
	EnvironmentType string         `json:"environment_type"`
}

// BrokerProfile is a PM Broker profile fixture. AllowedConnectorKinds is pinned
// contract metadata only; the synthetic fake broker serves only the synthetic
// connector connection fixture in this package.
type BrokerProfile struct {
	BrokerProfileID       BrokerProfileID `json:"broker_profile_id"`
	OrganizationID        OrganizationID  `json:"organization_id"`
	WorkspaceID           WorkspaceID     `json:"workspace_id"`
	EnvironmentID         EnvironmentID   `json:"environment_id"`
	DisplayName           string          `json:"display_name"`
	AllowedConnectorKinds []string        `json:"allowed_connector_kinds"`
	AuthRegistryMode      string          `json:"auth_registry_mode"`
}

// Validate reports whether the broker profile matches the pinned fixture metadata.
func (profile BrokerProfile) Validate() error {
	if !profile.BrokerProfileID.IsValid() || !profile.OrganizationID.IsValid() ||
		!profile.WorkspaceID.IsValid() || !profile.EnvironmentID.IsValid() {
		return ErrInvalidIdentityBoundary
	}
	if profile.AuthRegistryMode != "internal_experimental" {
		return ErrInvalidIdentityBoundary
	}
	if len(profile.AllowedConnectorKinds) != 2 || profile.AllowedConnectorKinds[0] != "synthetic" ||
		profile.AllowedConnectorKinds[1] != "gcp" {
		return ErrInvalidIdentityBoundary
	}
	return nil
}

// OpaqueSecretReference carries only a broker-held reference, never raw secret bytes.
type OpaqueSecretReference struct {
	Kind        string                  `json:"kind"`
	Ref         OpaqueSecretReferenceID `json:"ref"`
	DisplayHint string                  `json:"display_hint"`
	Exportable  bool                    `json:"exportable"`
}

// NewOpaqueSecretReference returns a non-exportable opaque reference fixture value.
func NewOpaqueSecretReference(ref OpaqueSecretReferenceID, displayHint string) OpaqueSecretReference {
	return OpaqueSecretReference{
		Kind:        "opaque_secret_reference",
		Ref:         ref,
		DisplayHint: displayHint,
		Exportable:  false,
	}
}

// Validate reports whether the reference is opaque and non-exportable.
func (ref OpaqueSecretReference) Validate() error {
	if ref.Kind != "opaque_secret_reference" || !ref.Ref.IsValid() {
		return ErrInvalidOpaqueSecretReference
	}
	if !isSafeDisplayHint(ref.DisplayHint) {
		return ErrUnsafeDisplayHint
	}
	if ref.Exportable {
		return ErrRawSecretExportForbidden
	}
	return nil
}

func isSafeDisplayHint(displayHint string) bool {
	if !displayHintPattern.MatchString(displayHint) {
		return false
	}
	lowerHint := strings.ToLower(displayHint)
	for _, marker := range unsafeDisplayHintMarkers {
		if strings.Contains(lowerHint, marker) {
			return false
		}
	}
	return true
}

// ConnectorConnection is a PM Broker connector connection fixture.
type ConnectorConnection struct {
	ConnectorConnectionID ConnectorConnectionID `json:"connector_connection_id"`
	OrganizationID        OrganizationID        `json:"organization_id"`
	WorkspaceID           WorkspaceID           `json:"workspace_id"`
	EnvironmentID         EnvironmentID         `json:"environment_id"`
	BrokerProfileID       BrokerProfileID       `json:"broker_profile_id"`
	ConnectorKind         string                `json:"connector_kind"`
	AuthRef               OpaqueSecretReference `json:"auth_ref"`
	Status                string                `json:"status"`
	WriteMode             string                `json:"write_mode"`
}

// Validate reports whether the connector connection preserves the safe /v1 boundaries.
func (connection ConnectorConnection) Validate() error {
	if !connection.ConnectorConnectionID.IsValid() || !connection.OrganizationID.IsValid() ||
		!connection.WorkspaceID.IsValid() || !connection.EnvironmentID.IsValid() ||
		!connection.BrokerProfileID.IsValid() {
		return ErrInvalidIdentityBoundary
	}
	if connection.ConnectorKind != "synthetic" || connection.Status != "ready" || connection.WriteMode != "deny" {
		return ErrInvalidExecutionPlan
	}
	return connection.AuthRef.Validate()
}

// ExecutionIntent is a closed typed envelope for execution-plan intents.
type ExecutionIntent struct {
	ValidateConnectorConnection *ValidateConnectorConnectionIntent `json:"validate_connector_connection,omitempty"`
}

// ValidateConnectorConnectionIntent requests connector-connection validation.
type ValidateConnectorConnectionIntent struct {
	Kind                  string                `json:"kind"`
	ConnectorConnectionID ConnectorConnectionID `json:"connector_connection_id"`
}

// Validate reports whether the execution intent is present, typed, and unambiguous.
func (intent ExecutionIntent) Validate() error {
	if intent.ValidateConnectorConnection == nil {
		return ErrInvalidExecutionIntent
	}
	if intent.ValidateConnectorConnection.Kind != "validate_connector_connection" ||
		!intent.ValidateConnectorConnection.ConnectorConnectionID.IsValid() {
		return ErrInvalidExecutionIntent
	}
	return nil
}

// ExecutionPlanRequest is the closed typed request for execution-plan creation.
type ExecutionPlanRequest struct {
	OrganizationID        OrganizationID        `json:"organization_id"`
	WorkspaceID           WorkspaceID           `json:"workspace_id"`
	EnvironmentID         EnvironmentID         `json:"environment_id"`
	BrokerProfileID       BrokerProfileID       `json:"broker_profile_id"`
	ConnectorConnectionID ConnectorConnectionID `json:"connector_connection_id"`
	IdempotencyKey        IdempotencyKey        `json:"idempotency_key"`
	Intent                ExecutionIntent       `json:"intent"`
}

// Validate reports whether the request matches the pinned /v1 execution-plan contract.
func (request ExecutionPlanRequest) Validate() error {
	if !request.OrganizationID.IsValid() || !request.WorkspaceID.IsValid() ||
		!request.EnvironmentID.IsValid() || !request.BrokerProfileID.IsValid() ||
		!request.ConnectorConnectionID.IsValid() {
		return ErrInvalidIdentityBoundary
	}
	if !request.IdempotencyKey.IsValid() {
		return ErrInvalidExecutionPlan
	}
	return request.Intent.Validate()
}

// ExecutionPlan is the immutable planned response from the fake broker.
type ExecutionPlan struct {
	ExecutionPlanID       ExecutionPlanID       `json:"execution_plan_id"`
	Digest                ExecutionPlanDigest   `json:"digest"`
	OrganizationID        OrganizationID        `json:"organization_id"`
	WorkspaceID           WorkspaceID           `json:"workspace_id"`
	EnvironmentID         EnvironmentID         `json:"environment_id"`
	BrokerProfileID       BrokerProfileID       `json:"broker_profile_id"`
	ConnectorConnectionID ConnectorConnectionID `json:"connector_connection_id"`
	IdempotencyKey        IdempotencyKey        `json:"idempotency_key"`
	Intent                ExecutionIntent       `json:"intent"`
	Status                string                `json:"status"`
	CreatedAt             time.Time             `json:"created_at"`
	ExpiresAt             time.Time             `json:"expires_at"`
}

// Validate reports whether the plan matches the pinned /v1 execution-plan contract.
func (plan ExecutionPlan) Validate() error {
	request := ExecutionPlanRequest{
		OrganizationID:        plan.OrganizationID,
		WorkspaceID:           plan.WorkspaceID,
		EnvironmentID:         plan.EnvironmentID,
		BrokerProfileID:       plan.BrokerProfileID,
		ConnectorConnectionID: plan.ConnectorConnectionID,
		IdempotencyKey:        plan.IdempotencyKey,
		Intent:                plan.Intent,
	}
	if !plan.ExecutionPlanID.IsValid() || !plan.Digest.IsValid() || plan.Status != "planned" {
		return ErrInvalidExecutionPlan
	}
	if !plan.ExpiresAt.After(plan.CreatedAt) {
		return ErrInvalidExecutionPlan
	}
	if err := request.Validate(); err != nil {
		return err
	}
	return nil
}

// SafeError is the client-safe PM Broker error shape.
type SafeError struct {
	Code              ErrorCode         `json:"code"`
	Message           string            `json:"message"`
	CorrelationID     CorrelationID     `json:"correlation_id"`
	SupportedVersions []ContractVersion `json:"supported_versions,omitempty"`
}

// ValidateIncompatibleVersion reports whether the error is the exact /v1 version refusal.
func (safeError SafeError) ValidateIncompatibleVersion() error {
	if safeError.Code != ErrorCodeIncompatibleContractVersion ||
		safeError.Message != IncompatibleContractVersionMessage || !safeError.CorrelationID.IsSafe() ||
		len(safeError.SupportedVersions) != 1 || safeError.SupportedVersions[0] != ContractVersion1 {
		return ErrInvalidErrorResponse
	}
	return nil
}

// IncompatibleContractVersionErrorResponse is the HTTP 426 refusal envelope.
type IncompatibleContractVersionErrorResponse struct {
	Error SafeError `json:"error"`
}

// Validate reports whether the response is the exact incompatible-version refusal fixture.
func (response IncompatibleContractVersionErrorResponse) Validate() error {
	return response.Error.ValidateIncompatibleVersion()
}

// SyntheticFixtures groups the deterministic fixture values accepted by PM Broker PR #35.
type SyntheticFixtures struct {
	Compatibility            Compatibility                            `json:"compatibility"`
	Organization             Organization                             `json:"organization"`
	Workspace                Workspace                                `json:"workspace"`
	Environment              Environment                              `json:"environment"`
	BrokerProfile            BrokerProfile                            `json:"broker_profile"`
	ConnectorConnection      ConnectorConnection                      `json:"connector_connection"`
	ExecutionPlanRequest     ExecutionPlanRequest                     `json:"execution_plan_request"`
	ExecutionPlan            ExecutionPlan                            `json:"execution_plan"`
	IncompatibleVersionError IncompatibleContractVersionErrorResponse `json:"incompatible_version_error"`
}

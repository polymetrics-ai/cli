package app

import (
	"context"
	"errors"
	"sort"
	"strings"
	"time"

	"polymetrics.ai/internal/connectors"
)

// AuthorizationStreamTable names one source stream/table and its destination
// table in a durable authorization scope. It intentionally describes shape,
// never record content.
type AuthorizationStreamTable struct {
	Stream           string `json:"stream"`
	SourceTable      string `json:"source_table"`
	DestinationTable string `json:"destination_table"`
}

// AuthorizationScope is the content-free shape approved for repeated
// execution. DestinationConnection is an opaque non-secret connection
// reference; credential/configuration values are represented only by their
// derived revisions and digests.
type AuthorizationScope struct {
	SourceConnection               string                       `json:"source_connection"`
	DestinationConnection          string                       `json:"destination_connection"`
	DestinationCredentialRevision  string                       `json:"destination_credential_revision"`
	StreamTables                   []AuthorizationStreamTable   `json:"stream_tables"`
	FieldMappings                  map[string]string            `json:"field_mappings"`
	WriteAction                    string                       `json:"write_action"`
	DestinationConfigurationDigest string                       `json:"destination_configuration_digest"`
	EnabledOperations              []string                     `json:"enabled_operations,omitempty"`
	ConfirmationPolicy             connectors.WriteConfirmation `json:"confirmation_policy"`
	ExpiresAt                      time.Time                    `json:"expires_at"`
}

// AuthorizationRecord is the durable, revocable standing authorization. Its
// reference is safe to persist in a schedule; it has no approval token, raw
// credential, raw destination configuration, payload, cursor, or run data.
type AuthorizationRecord struct {
	Reference     string             `json:"reference"`
	ScopeIdentity string             `json:"scope_identity"`
	Scope         AuthorizationScope `json:"scope"`
	CreatedAt     time.Time          `json:"created_at"`
	RevokedAt     time.Time          `json:"revoked_at,omitempty"`
}

// AuthorizationScopeIdentity returns the stable content-free identity for an
// approved execution shape. Canonicalization makes semantic stream/operation
// sets order independent while preserving every bound property.
func AuthorizationScopeIdentity(scope AuthorizationScope) (string, error) {
	canonical := canonicalAuthorizationScope(scope)
	return hashJSON(canonical)
}

func canonicalAuthorizationScope(scope AuthorizationScope) AuthorizationScope {
	canonical := cloneAuthorizationScope(scope)
	canonical.ExpiresAt = canonical.ExpiresAt.UTC()
	sort.Slice(canonical.StreamTables, func(i, j int) bool {
		left, right := canonical.StreamTables[i], canonical.StreamTables[j]
		if left.Stream != right.Stream {
			return left.Stream < right.Stream
		}
		if left.SourceTable != right.SourceTable {
			return left.SourceTable < right.SourceTable
		}
		return left.DestinationTable < right.DestinationTable
	})
	sort.Strings(canonical.EnabledOperations)
	return canonical
}

func cloneAuthorizationScope(scope AuthorizationScope) AuthorizationScope {
	clone := scope
	clone.StreamTables = append([]AuthorizationStreamTable(nil), scope.StreamTables...)
	clone.FieldMappings = cloneStringMap(scope.FieldMappings)
	clone.EnabledOperations = append([]string(nil), scope.EnabledOperations...)
	return clone
}

func cloneAuthorizationRecord(record AuthorizationRecord) AuthorizationRecord {
	clone := record
	clone.Scope = cloneAuthorizationScope(record.Scope)
	return clone
}

// AuthorizationScopeForReversePlan derives a safe scope from a stored bulk
// reverse plan. It deliberately does not read source rows, payload identities,
// record counts, timestamps, cursor state, or a run ID.
func (a *App) AuthorizationScopeForReversePlan(ctx context.Context, planID string) (AuthorizationScope, error) {
	plan, err := a.loadReversePlan(planID)
	if err != nil {
		return AuthorizationScope{}, err
	}
	if plan.Mode == reversePlanModeConnectorCommand {
		return AuthorizationScope{}, errors.New("connector command plans do not have a source-table authorization scope")
	}
	_, credential, runtime, err := a.resolveEndpointWithCredential(ctx, EndpointConfig{
		Connector: plan.DestinationConnector, Credential: plan.DestinationCredential, Config: plan.DestinationConfig,
	})
	if err != nil {
		return AuthorizationScope{}, err
	}
	return a.authorizationScopeForReversePlan(plan, credential, runtime), nil
}

func (a *App) authorizationScopeForReversePlan(plan ReversePlan, credential CredentialMeta, runtime connectors.RuntimeConfig) AuthorizationScope {
	confirmation := a.confirmationPolicyForPlan(plan)
	// A bulk reverse plan enables exactly its declared action. Keeping the
	// closed one-item operation set in the scope means a later destructive or
	// keyed action cannot borrow this authorization merely because its other
	// destination metadata matches.
	enabledOperations := []string{plan.Action}
	return canonicalAuthorizationScope(AuthorizationScope{
		SourceConnection:              plan.SourceConnection,
		DestinationConnection:         credential.ID,
		DestinationCredentialRevision: runtime.CredentialRevision,
		StreamTables: []AuthorizationStreamTable{{
			Stream: "records", SourceTable: plan.SourceTable, DestinationTable: plan.Name,
		}},
		FieldMappings:                  cloneStringMap(plan.Mappings),
		WriteAction:                    plan.Action,
		DestinationConfigurationDigest: runtime.ConfigurationDigest,
		EnabledOperations:              enabledOperations,
		ConfirmationPolicy:             confirmation,
		ExpiresAt:                      plan.ExpiresAt,
	})
}

func newAuthorizationRecord(scope AuthorizationScope, now time.Time) (AuthorizationRecord, error) {
	identity, err := AuthorizationScopeIdentity(scope)
	if err != nil {
		return AuthorizationRecord{}, err
	}
	reference, err := prefixedID("auth")
	if err != nil {
		return AuthorizationRecord{}, err
	}
	return AuthorizationRecord{
		Reference: reference, ScopeIdentity: identity, Scope: canonicalAuthorizationScope(scope), CreatedAt: now.UTC(),
	}, nil
}

// ListAuthorizations returns defensive copies of safe durable authorization
// records. It never surfaces approval tokens or raw credential material.
func (a *App) ListAuthorizations() []AuthorizationRecord {
	out := make([]AuthorizationRecord, len(a.state.Authorizations))
	for i, record := range a.state.Authorizations {
		out[i] = cloneAuthorizationRecord(record)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].CreatedAt.Before(out[j].CreatedAt) })
	return out
}

// RevokeAuthorization is the explicit per-connection off switch. Revocation
// is idempotent so a retry cannot accidentally restore authorization.
func (a *App) RevokeAuthorization(reference string) error {
	if strings.TrimSpace(reference) == "" {
		return errors.New("authorization reference is required")
	}
	now := time.Now().UTC()
	updated, err := a.updateState(func(current state) (state, error) {
		for i := range current.Authorizations {
			if current.Authorizations[i].Reference != reference {
				continue
			}
			if current.Authorizations[i].RevokedAt.IsZero() {
				current.Authorizations[i].RevokedAt = now
			}
			return current, nil
		}
		return current, fmtAuthorizationNotFound(reference)
	})
	if err != nil {
		return err
	}
	a.state = updated
	return nil
}

func fmtAuthorizationNotFound(reference string) error {
	return errors.New("authorization " + reference + " not found")
}

func (a *App) requireAuthorization(reference string, actual AuthorizationScope, now time.Time) (AuthorizationRecord, error) {
	loaded, err := a.store.LoadReadOnly()
	if err != nil {
		return AuthorizationRecord{}, err
	}
	if err := a.normalizeLoadedState(loaded, false); err != nil {
		return AuthorizationRecord{}, err
	}
	for _, record := range a.state.Authorizations {
		if record.Reference != reference {
			continue
		}
		if !record.RevokedAt.IsZero() {
			return AuthorizationRecord{}, &AuthorizationRevokedError{Reference: reference}
		}
		if !now.UTC().Before(record.Scope.ExpiresAt.UTC()) {
			return AuthorizationRecord{}, &AuthorizationExpiredError{Reference: reference}
		}
		storedIdentity, err := AuthorizationScopeIdentity(record.Scope)
		if err != nil {
			return AuthorizationRecord{}, err
		}
		if !constantTimeStringEqual(storedIdentity, record.ScopeIdentity) {
			return AuthorizationRecord{}, &AuthorizationScopeChangedError{Reference: reference, Property: "stored_scope_identity"}
		}
		actualIdentity, err := AuthorizationScopeIdentity(actual)
		if err != nil {
			return AuthorizationRecord{}, err
		}
		if !constantTimeStringEqual(actualIdentity, record.ScopeIdentity) {
			return AuthorizationRecord{}, &AuthorizationScopeChangedError{Reference: reference, Property: authorizationScopeDifference(record.Scope, actual)}
		}
		return cloneAuthorizationRecord(record), nil
	}
	return AuthorizationRecord{}, fmtAuthorizationNotFound(reference)
}

func authorizationScopeDifference(expected, actual AuthorizationScope) string {
	expected = canonicalAuthorizationScope(expected)
	actual = canonicalAuthorizationScope(actual)
	switch {
	case expected.SourceConnection != actual.SourceConnection:
		return "source_connection"
	case expected.DestinationConnection != actual.DestinationConnection:
		return "destination_connection"
	case !constantTimeStringEqual(expected.DestinationCredentialRevision, actual.DestinationCredentialRevision):
		return "destination_credential_revision"
	case !sameAuthorizationStreamTables(expected.StreamTables, actual.StreamTables):
		return "stream_table_set"
	case !sameStringMap(expected.FieldMappings, actual.FieldMappings):
		return "field_mappings"
	case expected.WriteAction != actual.WriteAction:
		return "write_action"
	case !constantTimeStringEqual(expected.DestinationConfigurationDigest, actual.DestinationConfigurationDigest):
		return "destination_configuration"
	case !sameStringSlice(expected.EnabledOperations, actual.EnabledOperations):
		return "enabled_operations"
	case expected.ConfirmationPolicy.Kind != actual.ConfirmationPolicy.Kind:
		return "confirmation_policy"
	case !expected.ExpiresAt.Equal(actual.ExpiresAt):
		return "expiry"
	default:
		return "scope_identity"
	}
}

func sameAuthorizationStreamTables(left, right []AuthorizationStreamTable) bool {
	if len(left) != len(right) {
		return false
	}
	for i := range left {
		if left[i] != right[i] {
			return false
		}
	}
	return true
}

func sameStringMap(left, right map[string]string) bool {
	if len(left) != len(right) {
		return false
	}
	for key, leftValue := range left {
		rightValue, ok := right[key]
		if !ok || leftValue != rightValue {
			return false
		}
	}
	return true
}

func sameStringSlice(left, right []string) bool {
	if len(left) != len(right) {
		return false
	}
	for i := range left {
		if left[i] != right[i] {
			return false
		}
	}
	return true
}

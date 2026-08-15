package warehouse

import "errors"

// ArtifactIdentity is the connector-agnostic owner triple for one warehouse
// materialization. It contains durable identifiers only—never a credential,
// DSN, display name, or connector implementation—and is shared by the
// warehouse layout and typed connector-mediation values.
type ArtifactIdentity struct {
	WorkspaceID  string
	ConnectorID  string
	ConnectionID string
}

// Validate confirms that every identity component can safely address the
// structural warehouse layout without rewriting or collapsing names.
func (i ArtifactIdentity) Validate() error {
	if !SafePathPart(i.WorkspaceID) || !SafePathPart(i.ConnectorID) || !SafePathPart(i.ConnectionID) {
		return errors.New("warehouse artifact identity is incomplete or invalid")
	}
	return nil
}

// SameIdentity compares exactly the structural ownership triple. It excludes
// display and credential concepts because neither is part of an artifact.
func (i ArtifactIdentity) SameIdentity(other ArtifactIdentity) bool {
	return i.WorkspaceID == other.WorkspaceID &&
		i.ConnectorID == other.ConnectorID &&
		i.ConnectionID == other.ConnectionID
}

// ArtifactRef is a connector-agnostic typed address of one durable warehouse
// materialization. It gives a connector-side leg a warehouse address without
// naming another connector or a direct source-to-destination route.
type ArtifactRef struct {
	identity ArtifactIdentity
	table    string
}

// NewArtifactRef creates a structurally safe warehouse materialization
// address. It does not create, open, read, or write the artifact.
func NewArtifactRef(identity ArtifactIdentity, table string) (ArtifactRef, error) {
	ref := ArtifactRef{identity: identity, table: table}
	if err := ref.Validate(); err != nil {
		return ArtifactRef{}, err
	}
	return ref, nil
}

// Validate checks the artifact address before a shared warehouse mediator
// resolves it to a durable Parquet materialization.
func (r ArtifactRef) Validate() error {
	if err := r.identity.Validate(); err != nil || !SafePathPart(r.table) {
		return errors.New("warehouse artifact reference is invalid")
	}
	return nil
}

// Identity returns the artifact's durable owner triple.
func (r ArtifactRef) Identity() ArtifactIdentity { return r.identity }

// Table returns the opaque logical table component selected within the owner's
// warehouse region.
func (r ArtifactRef) Table() string { return r.table }

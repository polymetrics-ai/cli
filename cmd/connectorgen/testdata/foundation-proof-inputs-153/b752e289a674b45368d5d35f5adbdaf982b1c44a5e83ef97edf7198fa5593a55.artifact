package database

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"errors"

	"polymetrics.ai/internal/warehouse"
)

const managedTargetNameDomain = "polymetrics-managed-target-v1"

// TargetOwner is the complete, source-derived owner of a managed database
// target. It deliberately reuses the warehouse artifact identity: workspace,
// source connector, and source connection. It has no target credential, DSN,
// or display-name field.
type TargetOwner struct {
	identity ConnectionIdentity
}

// NewTargetOwner creates a managed-target owner from the source-owned
// warehouse identity.
func NewTargetOwner(identity ConnectionIdentity) (TargetOwner, error) {
	owner := TargetOwner{identity: identity}
	if err := owner.validate(); err != nil {
		return TargetOwner{}, errors.New("database managed target owner is invalid")
	}
	return owner, nil
}

func (o TargetOwner) validate() error {
	return validateDatabaseConnectionIdentity(o.identity)
}

// Identity returns the complete opaque source owner triple. It is structural
// identity only; a display rename cannot change managed-target ownership.
func (o TargetOwner) Identity() ConnectionIdentity { return o.identity }

func (o TargetOwner) sameIdentity(other TargetOwner) bool {
	return o.identity.SameIdentity(other.identity)
}

// TargetDatabaseIdentity is the opaque native identity of the destination
// database instance. It is asserted by plans and control records, but is never
// an input to a physical managed-target name: credential rotation within the
// same database must not move a relation.
type TargetDatabaseIdentity struct {
	kind  string
	value string
}

// NewTargetDatabaseIdentity creates an opaque native database identity observed
// by a driver. It cannot be a DSN, credential, or rendered SQL fragment.
func NewTargetDatabaseIdentity(kind, value string) (TargetDatabaseIdentity, error) {
	identity := TargetDatabaseIdentity{kind: kind, value: value}
	if err := identity.validate(); err != nil {
		return TargetDatabaseIdentity{}, errors.New("database managed target database identity is invalid")
	}
	return identity, nil
}

func (i TargetDatabaseIdentity) validate() error {
	if !validOpaqueName(i.kind) || !validOpaqueIdentityComponent(i.value) {
		return errors.New("database managed target database identity is invalid")
	}
	return nil
}

func (i TargetDatabaseIdentity) sameIdentity(other TargetDatabaseIdentity) bool {
	return i.kind == other.kind && i.value == other.value
}

// Kind returns the driver's closed native-identity kind.
func (i TargetDatabaseIdentity) Kind() string { return i.kind }

// Value returns the opaque native identity component.
func (i TargetDatabaseIdentity) Value() string { return i.value }

// NativeNamespaceIdentity is the opaque identity of an observed namespace. It
// detects a same-named namespace that has been replaced independently from a
// managed relation.
type NativeNamespaceIdentity struct {
	kind  string
	value string
}

// NewNativeNamespaceIdentity creates an opaque native namespace identity.
func NewNativeNamespaceIdentity(kind, value string) (NativeNamespaceIdentity, error) {
	identity := NativeNamespaceIdentity{kind: kind, value: value}
	if err := identity.validate(); err != nil {
		return NativeNamespaceIdentity{}, errors.New("database managed target namespace identity is invalid")
	}
	return identity, nil
}

func (i NativeNamespaceIdentity) validate() error {
	if !validOpaqueName(i.kind) || !validOpaqueIdentityComponent(i.value) {
		return errors.New("database managed target namespace identity is invalid")
	}
	return nil
}

func (i NativeNamespaceIdentity) sameIdentity(other NativeNamespaceIdentity) bool {
	return i.kind == other.kind && i.value == other.value
}

// Kind returns the driver's closed native-identity kind.
func (i NativeNamespaceIdentity) Kind() string { return i.kind }

// Value returns the opaque native identity component.
func (i NativeNamespaceIdentity) Value() string { return i.value }

// ManagedTargetRef is the deterministically derived physical namespace and
// relation for one source-owned warehouse artifact. Its names are opaque hashes
// rather than user display text or credentials. A future native driver may
// render only these structured components through its own closed primitives.
type ManagedTargetRef struct {
	owner     TargetOwner
	artifact  warehouse.ArtifactRef
	streamID  string
	namespace string
	relation  string
}

// NewManagedTargetRef derives a target address from an asserted source owner
// and its warehouse artifact plus the stream's persisted immutable ID. The
// artifact table remains provenance only; display/table changes cannot relocate
// a managed target. It neither opens a database nor creates a namespace,
// relation, or control record.
func NewManagedTargetRef(owner TargetOwner, artifact warehouse.ArtifactRef, streamID string) (ManagedTargetRef, error) {
	if err := owner.validate(); err != nil || artifact.Validate() != nil || !owner.identity.SameIdentity(artifact.Identity()) || !validOpaqueIdentityComponent(streamID) {
		return ManagedTargetRef{}, errors.New("database managed target reference is invalid")
	}
	ref := ManagedTargetRef{
		owner:     owner,
		artifact:  artifact,
		streamID:  streamID,
		namespace: deriveManagedTargetName("namespace", owner.identity.WorkspaceID, owner.identity.ConnectorID, owner.identity.ConnectionID),
		relation: deriveManagedTargetName(
			"relation",
			owner.identity.WorkspaceID,
			owner.identity.ConnectorID,
			owner.identity.ConnectionID,
			streamID,
		),
	}
	if err := ref.validate(); err != nil {
		return ManagedTargetRef{}, errors.New("database managed target reference is invalid")
	}
	return ref, nil
}

func (r ManagedTargetRef) validate() error {
	if err := r.owner.validate(); err != nil || r.artifact.Validate() != nil || !r.owner.identity.SameIdentity(r.artifact.Identity()) || !validOpaqueIdentityComponent(r.streamID) {
		return errors.New("database managed target reference is invalid")
	}
	wantNamespace := deriveManagedTargetName(
		"namespace",
		r.owner.identity.WorkspaceID,
		r.owner.identity.ConnectorID,
		r.owner.identity.ConnectionID,
	)
	wantRelation := deriveManagedTargetName(
		"relation",
		r.owner.identity.WorkspaceID,
		r.owner.identity.ConnectorID,
		r.owner.identity.ConnectionID,
		r.streamID,
	)
	if r.namespace != wantNamespace || r.relation != wantRelation ||
		validateIdentifierComponent(r.namespace) != nil || validateIdentifierComponent(r.relation) != nil {
		return errors.New("database managed target reference is invalid")
	}
	return nil
}

func (r ManagedTargetRef) sameTarget(other ManagedTargetRef) bool {
	return r.owner.sameIdentity(other.owner) &&
		r.artifact.Identity().SameIdentity(other.artifact.Identity()) &&
		r.streamID == other.streamID &&
		r.namespace == other.namespace &&
		r.relation == other.relation
}

func (r ManagedTargetRef) lockKey() string { return r.namespace }

// Owner returns the asserted source owner for this managed target.
func (r ManagedTargetRef) Owner() TargetOwner { return r.owner }

// SourceArtifact returns the source-owned warehouse artifact from which this
// target address receives provenance. Its table name is not target identity.
func (r ManagedTargetRef) SourceArtifact() warehouse.ArtifactRef { return r.artifact }

// StreamID returns the persisted immutable stream identity used for this
// relation. It is deliberately separate from the stream map key and table name.
func (r ManagedTargetRef) StreamID() string { return r.streamID }

// Namespace returns the opaque, deterministic physical namespace component.
func (r ManagedTargetRef) Namespace() string { return r.namespace }

// Relation returns the opaque, deterministic physical relation component.
func (r ManagedTargetRef) Relation() string { return r.relation }

// ManagedTargetSchema pins the managed-target schema contract. A future
// driver reports its observed schema through the same value; this package never
// synthesizes an ALTER or attempts schema evolution.
type ManagedTargetSchema struct {
	version     uint
	fingerprint SchemaFingerprint
}

// NewManagedTargetSchema creates a complete schema version/fingerprint pair.
func NewManagedTargetSchema(version uint, fingerprint SchemaFingerprint) (ManagedTargetSchema, error) {
	schema := ManagedTargetSchema{version: version, fingerprint: fingerprint}
	if err := schema.validate(); err != nil {
		return ManagedTargetSchema{}, errors.New("database managed target schema is invalid")
	}
	return schema, nil
}

func (s ManagedTargetSchema) validate() error {
	if s.version == 0 || s.fingerprint.IsZero() {
		return errors.New("database managed target schema is invalid")
	}
	return nil
}

func (s ManagedTargetSchema) sameSchema(other ManagedTargetSchema) bool {
	return s.version == other.version && s.fingerprint == other.fingerprint
}

// Version returns the managed schema contract version.
func (s ManagedTargetSchema) Version() uint { return s.version }

// Fingerprint returns the immutable schema fingerprint.
func (s ManagedTargetSchema) Fingerprint() SchemaFingerprint { return s.fingerprint }

// ManagedTargetNamespaceOwnerRecord is the durable assertion that a namespace
// belongs to one source connection in one observed target database. One record
// may own many per-stream relations; it is not a relation control row.
type ManagedTargetNamespaceOwnerRecord struct {
	owner     TargetOwner
	targetDB  TargetDatabaseIdentity
	namespace string
	native    NativeNamespaceIdentity
}

// NewManagedTargetNamespaceOwnerRecord constructs a namespace-only ownership
// assertion from an exact source owner and one target in that namespace.
func NewManagedTargetNamespaceOwnerRecord(owner TargetOwner, target ManagedTargetRef, targetDB TargetDatabaseIdentity, native NativeNamespaceIdentity) (ManagedTargetNamespaceOwnerRecord, error) {
	record := ManagedTargetNamespaceOwnerRecord{owner: owner, targetDB: targetDB, namespace: target.namespace, native: native}
	if err := record.validate(); err != nil || !owner.sameIdentity(target.owner) {
		return ManagedTargetNamespaceOwnerRecord{}, errors.New("database managed target namespace owner record is invalid")
	}
	return record, nil
}

func (r ManagedTargetNamespaceOwnerRecord) validate() error {
	if err := r.owner.validate(); err != nil || r.targetDB.validate() != nil || r.native.validate() != nil {
		return errors.New("database managed target namespace owner record is invalid")
	}
	wantNamespace := deriveManagedTargetName("namespace", r.owner.identity.WorkspaceID, r.owner.identity.ConnectorID, r.owner.identity.ConnectionID)
	if r.namespace != wantNamespace || validateIdentifierComponent(r.namespace) != nil {
		return errors.New("database managed target namespace owner record is invalid")
	}
	return nil
}

// Owner returns the source connection that owns the namespace.
func (r ManagedTargetNamespaceOwnerRecord) Owner() TargetOwner { return r.owner }

// TargetDatabase returns the opaque database instance identity asserted here.
func (r ManagedTargetNamespaceOwnerRecord) TargetDatabase() TargetDatabaseIdentity { return r.targetDB }

// Namespace returns the physical namespace asserted by this owner record.
func (r ManagedTargetNamespaceOwnerRecord) Namespace() string { return r.namespace }

// NativeIdentity returns the opaque native namespace identity.
func (r ManagedTargetNamespaceOwnerRecord) NativeIdentity() NativeNamespaceIdentity { return r.native }

// ManagedTargetControlRecord is the durable assertion that a physical target
// is ours. It binds source-derived ownership, target address, destination
// database identity, native relation identity, and schema contract together.
// Its complete equality—not a name match—is required for repeat admission.
type ManagedTargetControlRecord struct {
	owner    TargetOwner
	target   ManagedTargetRef
	targetDB TargetDatabaseIdentity
	native   NativeRelationIdentity
	schema   ManagedTargetSchema
}

// NewManagedTargetControlRecord validates one complete control assertion. It
// is constructed by a future native driver only after it can observe the
// relation it created or inspected.
func NewManagedTargetControlRecord(owner TargetOwner, target ManagedTargetRef, targetDB TargetDatabaseIdentity, native NativeRelationIdentity, schema ManagedTargetSchema) (ManagedTargetControlRecord, error) {
	record := ManagedTargetControlRecord{owner: owner, target: target, targetDB: targetDB, native: native, schema: schema}
	if err := record.validate(); err != nil {
		return ManagedTargetControlRecord{}, errors.New("database managed target control record is invalid")
	}
	return record, nil
}

func (r ManagedTargetControlRecord) validate() error {
	if err := r.owner.validate(); err != nil || r.target.validate() != nil || !r.owner.sameIdentity(r.target.owner) {
		return errors.New("database managed target control record is invalid")
	}
	if r.targetDB.validate() != nil || r.native.Kind == "" || r.native.Value == "" || r.native.validate() != nil {
		return errors.New("database managed target control record is invalid")
	}
	if err := r.schema.validate(); err != nil {
		return errors.New("database managed target control record is invalid")
	}
	return nil
}

// Owner returns the source-derived owner asserted by this record.
func (r ManagedTargetControlRecord) Owner() TargetOwner { return r.owner }

// Target returns the physical target address asserted by this record.
func (r ManagedTargetControlRecord) Target() ManagedTargetRef { return r.target }

// TargetDatabase returns the opaque destination database identity asserted by
// this per-relation record.
func (r ManagedTargetControlRecord) TargetDatabase() TargetDatabaseIdentity { return r.targetDB }

// NativeIdentity returns the opaque native identity that detects replacement
// of a same-named relation.
func (r ManagedTargetControlRecord) NativeIdentity() NativeRelationIdentity { return r.native }

// Schema returns the version/fingerprint assertion for the target schema.
func (r ManagedTargetControlRecord) Schema() ManagedTargetSchema { return r.schema }

func deriveManagedTargetName(kind string, components ...string) string {
	hash := sha256.New()
	hash.Write([]byte(managedTargetNameDomain))
	hash.Write([]byte{0})
	writeManagedTargetHashComponent(hash, kind)
	for _, component := range components {
		writeManagedTargetHashComponent(hash, component)
	}
	// A 128-bit digest keeps both names under the conservative 63-byte database
	// identifier budget. The complete control record detects the astronomically
	// unlikely derived-name collision rather than treating a hash match as proof
	// of ownership.
	sum := hash.Sum(nil)
	return "pmmt_" + kind + "_" + hex.EncodeToString(sum[:16])
}

func writeManagedTargetHashComponent(hash interface{ Write([]byte) (int, error) }, component string) {
	var length [8]byte
	binary.BigEndian.PutUint64(length[:], uint64(len(component)))
	_, _ = hash.Write(length[:])
	_, _ = hash.Write([]byte(component))
}

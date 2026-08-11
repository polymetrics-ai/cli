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

// ManagedTargetRef is the deterministically derived physical namespace and
// relation for one source-owned warehouse artifact. Its names are opaque hashes
// rather than user display text or credentials. A future native driver may
// render only these structured components through its own closed primitives.
type ManagedTargetRef struct {
	owner     TargetOwner
	artifact  warehouse.ArtifactRef
	namespace string
	relation  string
}

// NewManagedTargetRef derives a target address from an asserted source owner
// and its warehouse artifact. It neither opens a database nor creates a
// namespace, relation, or control record.
func NewManagedTargetRef(owner TargetOwner, artifact warehouse.ArtifactRef) (ManagedTargetRef, error) {
	if err := owner.validate(); err != nil || artifact.Validate() != nil || !owner.identity.SameIdentity(artifact.Identity()) {
		return ManagedTargetRef{}, errors.New("database managed target reference is invalid")
	}
	ref := ManagedTargetRef{
		owner:     owner,
		artifact:  artifact,
		namespace: deriveManagedTargetName("namespace", owner.identity.WorkspaceID, owner.identity.ConnectorID, owner.identity.ConnectionID),
		relation: deriveManagedTargetName(
			"relation",
			owner.identity.WorkspaceID,
			owner.identity.ConnectorID,
			owner.identity.ConnectionID,
			artifact.Table(),
		),
	}
	if err := ref.validate(); err != nil {
		return ManagedTargetRef{}, errors.New("database managed target reference is invalid")
	}
	return ref, nil
}

func (r ManagedTargetRef) validate() error {
	if err := r.owner.validate(); err != nil || r.artifact.Validate() != nil || !r.owner.identity.SameIdentity(r.artifact.Identity()) {
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
		r.artifact.Table(),
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
		r.artifact.Table() == other.artifact.Table() &&
		r.namespace == other.namespace &&
		r.relation == other.relation
}

func (r ManagedTargetRef) lockKey() string {
	return r.namespace + "\x00" + r.relation
}

// Owner returns the asserted source owner for this managed target.
func (r ManagedTargetRef) Owner() TargetOwner { return r.owner }

// SourceArtifact returns the source-owned warehouse artifact from which this
// target address was derived.
func (r ManagedTargetRef) SourceArtifact() warehouse.ArtifactRef { return r.artifact }

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

// ManagedTargetControlRecord is the durable assertion that a physical target
// is ours. It binds source-derived ownership, target address, native relation
// identity, and schema contract together. Its complete equality—not a name
// match—is required for repeat admission.
type ManagedTargetControlRecord struct {
	owner  TargetOwner
	target ManagedTargetRef
	native NativeRelationIdentity
	schema ManagedTargetSchema
}

// NewManagedTargetControlRecord validates one complete control assertion. It
// is constructed by a future native driver only after it can observe the
// relation it created or inspected.
func NewManagedTargetControlRecord(owner TargetOwner, target ManagedTargetRef, native NativeRelationIdentity, schema ManagedTargetSchema) (ManagedTargetControlRecord, error) {
	record := ManagedTargetControlRecord{owner: owner, target: target, native: native, schema: schema}
	if err := record.validate(); err != nil {
		return ManagedTargetControlRecord{}, errors.New("database managed target control record is invalid")
	}
	return record, nil
}

func (r ManagedTargetControlRecord) validate() error {
	if err := r.owner.validate(); err != nil || r.target.validate() != nil || !r.owner.sameIdentity(r.target.owner) {
		return errors.New("database managed target control record is invalid")
	}
	if r.native.Kind == "" || r.native.Value == "" || r.native.validate() != nil {
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

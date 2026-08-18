package database

import (
	"context"
	"errors"
)

var (
	// ErrManagedTargetDeliveryLedgerInvalid prevents an incomplete or unasserted
	// target identity from reaching durable delivery storage.
	ErrManagedTargetDeliveryLedgerInvalid = errors.New("database managed target delivery ledger is invalid")
	// ErrManagedTargetDeliveryLedgerStore hides arbitrary storage/driver detail
	// from callers. A later native target driver must not expose credentials or
	// transport detail through this shared contract.
	ErrManagedTargetDeliveryLedgerStore = errors.New("database managed target delivery ledger storage failed")
)

// ManagedTargetDeliveryLedgerKey is the complete immutable address of one
// managed target's delivery history. It is derived only from an asserted
// control record: source owner, destination database, and the managed relation
// identity. It deliberately retains no ArtifactRef or source table/display
// name, so a mutable source-artifact rename cannot relocate ledger state.
type ManagedTargetDeliveryLedgerKey struct {
	owner     TargetOwner
	targetDB  TargetDatabaseIdentity
	streamID  string
	namespace string
	relation  string
}

// NewManagedTargetDeliveryLedgerKey derives a durable delivery address from a
// complete per-relation control assertion. An unasserted target cannot be used
// as ledger authority.
func NewManagedTargetDeliveryLedgerKey(control ManagedTargetControlRecord) (ManagedTargetDeliveryLedgerKey, error) {
	if err := control.validate(); err != nil {
		return ManagedTargetDeliveryLedgerKey{}, ErrManagedTargetDeliveryLedgerInvalid
	}
	target := control.Target()
	key := ManagedTargetDeliveryLedgerKey{
		owner:     control.Owner(),
		targetDB:  control.TargetDatabase(),
		streamID:  target.StreamID(),
		namespace: target.Namespace(),
		relation:  target.Relation(),
	}
	if err := key.validate(); err != nil {
		return ManagedTargetDeliveryLedgerKey{}, ErrManagedTargetDeliveryLedgerInvalid
	}
	return key, nil
}

func (k ManagedTargetDeliveryLedgerKey) validate() error {
	if err := k.owner.validate(); err != nil || k.targetDB.validate() != nil || !validOpaqueIdentityComponent(k.streamID) {
		return ErrManagedTargetDeliveryLedgerInvalid
	}
	wantNamespace := deriveManagedTargetName("namespace", k.owner.identity.WorkspaceID, k.owner.identity.ConnectorID, k.owner.identity.ConnectionID)
	wantRelation := deriveManagedTargetName("relation", k.owner.identity.WorkspaceID, k.owner.identity.ConnectorID, k.owner.identity.ConnectionID, k.streamID)
	if k.namespace != wantNamespace || k.relation != wantRelation ||
		validateIdentifierComponent(k.namespace) != nil || validateIdentifierComponent(k.relation) != nil {
		return ErrManagedTargetDeliveryLedgerInvalid
	}
	return nil
}

// Owner returns the asserted source owner of this ledger entry.
func (k ManagedTargetDeliveryLedgerKey) Owner() TargetOwner { return k.owner }

// TargetDatabase returns the asserted destination database identity of this
// ledger entry.
func (k ManagedTargetDeliveryLedgerKey) TargetDatabase() TargetDatabaseIdentity { return k.targetDB }

// StreamID returns the persisted immutable stream identity of this ledger
// entry. It is never a display name or source table name.
func (k ManagedTargetDeliveryLedgerKey) StreamID() string { return k.streamID }

// Namespace returns the deterministic managed namespace identity.
func (k ManagedTargetDeliveryLedgerKey) Namespace() string { return k.namespace }

// Relation returns the deterministic immutable managed relation identity.
func (k ManagedTargetDeliveryLedgerKey) Relation() string { return k.relation }

// ManagedTargetDeliveryRecord is opaque delivery evidence stored for one
// managed target ledger key. Later write-session work may define richer
// receipts and baselines; this shared foundation deliberately does not accept a
// source checkpoint, transaction stage, SQL result, or destination-DML handle.
type ManagedTargetDeliveryRecord struct {
	deliveryID string
}

// NewManagedTargetDeliveryRecord validates one opaque durable-delivery record
// identifier. Its value is caller-visible evidence, not a credential or a
// display/table name.
func NewManagedTargetDeliveryRecord(deliveryID string) (ManagedTargetDeliveryRecord, error) {
	record := ManagedTargetDeliveryRecord{deliveryID: deliveryID}
	if err := record.validate(); err != nil {
		return ManagedTargetDeliveryRecord{}, ErrManagedTargetDeliveryLedgerInvalid
	}
	return record, nil
}

func (r ManagedTargetDeliveryRecord) validate() error {
	if !validOpaqueIdentityComponent(r.deliveryID) {
		return ErrManagedTargetDeliveryLedgerInvalid
	}
	return nil
}

// DeliveryID returns the opaque delivery evidence identifier.
func (r ManagedTargetDeliveryRecord) DeliveryID() string { return r.deliveryID }

// ManagedTargetDeliveryLedgerStore is the driver-owned durable storage port
// for target delivery history. StoreManagedTargetDelivery must return only
// after it has durably persisted the supplied record at exactly the supplied
// key. It must not reinterpret the key as a mutable display or table name.
//
// Native driver storage is intentionally deferred. The shared ledger owns the
// typed identity boundary so every future driver uses the same address.
type ManagedTargetDeliveryLedgerStore interface {
	LoadManagedTargetDelivery(context.Context, ManagedTargetDeliveryLedgerKey) (ManagedTargetDeliveryRecord, bool, error)
	StoreManagedTargetDelivery(context.Context, ManagedTargetDeliveryLedgerKey, ManagedTargetDeliveryRecord) error
}

// ManagedTargetDeliveryLedger validates immutable target identity before it
// reaches the driver-owned durable store. It does not perform DDL, database
// writes, source checkpoint advancement, or transaction staging.
type ManagedTargetDeliveryLedger struct {
	store ManagedTargetDeliveryLedgerStore
}

// NewManagedTargetDeliveryLedger creates the shared driver-neutral ledger
// facade around a durable target-store port.
func NewManagedTargetDeliveryLedger(store ManagedTargetDeliveryLedgerStore) (*ManagedTargetDeliveryLedger, error) {
	if isNilInterface(store) {
		return nil, ErrManagedTargetDeliveryLedgerInvalid
	}
	return &ManagedTargetDeliveryLedger{store: store}, nil
}

// Record replaces the ledger evidence for one immutable target key only after
// the configured store reports durable persistence. Invalid identities are
// rejected before any storage mutation.
func (l *ManagedTargetDeliveryLedger) Record(ctx context.Context, key ManagedTargetDeliveryLedgerKey, record ManagedTargetDeliveryRecord) error {
	if l == nil || isNilInterface(l.store) || ctx == nil || key.validate() != nil || record.validate() != nil {
		return ErrManagedTargetDeliveryLedgerInvalid
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	if err := l.store.StoreManagedTargetDelivery(ctx, key, record); err != nil {
		return ErrManagedTargetDeliveryLedgerStore
	}
	return nil
}

// Lookup reads the delivery evidence for one immutable target key. A malformed
// stored record fails closed rather than being returned as target authority.
func (l *ManagedTargetDeliveryLedger) Lookup(ctx context.Context, key ManagedTargetDeliveryLedgerKey) (ManagedTargetDeliveryRecord, bool, error) {
	if l == nil || isNilInterface(l.store) || ctx == nil || key.validate() != nil {
		return ManagedTargetDeliveryRecord{}, false, ErrManagedTargetDeliveryLedgerInvalid
	}
	if err := ctx.Err(); err != nil {
		return ManagedTargetDeliveryRecord{}, false, err
	}
	record, found, err := l.store.LoadManagedTargetDelivery(ctx, key)
	if err != nil {
		return ManagedTargetDeliveryRecord{}, false, ErrManagedTargetDeliveryLedgerStore
	}
	if !found {
		return ManagedTargetDeliveryRecord{}, false, nil
	}
	if err := record.validate(); err != nil {
		return ManagedTargetDeliveryRecord{}, false, ErrManagedTargetDeliveryLedgerStore
	}
	return record, true, nil
}

package database

import (
	"context"
	"errors"
	"sync"
)

var (
	// ErrManagedTargetPlanInvalid prevents an untyped or mismatched request from
	// becoming a mutation. A provisioning plan always carries its asserted owner.
	ErrManagedTargetPlanInvalid = errors.New("database managed target provisioning plan is invalid")
	// ErrManagedTargetOwnerMissing refuses a relation that has no readable typed
	// ownership assertion; it is never adopted as a managed target.
	ErrManagedTargetOwnerMissing = errors.New("database managed target ownership record is missing")
	// ErrManagedTargetOwnerUnreadable refuses an owner record that cannot be
	// proven. Its cause is deliberately not rendered because it may be driver
	// supplied transport detail.
	ErrManagedTargetOwnerUnreadable = errors.New("database managed target ownership record is unreadable")
	// ErrManagedTargetOwnerForeign refuses a target asserted for another source
	// owner, even if it has a familiar physical name.
	ErrManagedTargetOwnerForeign = errors.New("database managed target is owned by another source")
	// ErrManagedTargetNameCollision refuses a same-owner control assertion for a
	// different target rather than reusing or overwriting it.
	ErrManagedTargetNameCollision = errors.New("database managed target name collides with another target")
	// ErrManagedTargetReplaced detects a same-named relation whose native
	// identity changed after the control record was made.
	ErrManagedTargetReplaced = errors.New("database managed target native identity changed")
	// ErrManagedTargetSchemaDrift refuses schema hash/version mismatches. This
	// layer deliberately has no migration or ALTER path.
	ErrManagedTargetSchemaDrift = errors.New("database managed target schema drifted")
	// ErrManagedTargetOrphaned refuses a durable control record whose physical
	// relation is gone. Recreating it would adopt an unprovable replacement.
	ErrManagedTargetOrphaned = errors.New("database managed target control record is orphaned")
	// ErrManagedTargetUnverifiable covers malformed or impossible observations.
	ErrManagedTargetUnverifiable = errors.New("database managed target state cannot be proven")
	// ErrManagedTargetProvisioning hides arbitrary driver errors from callers;
	// drivers must not cause credentials or transport detail to be rendered.
	ErrManagedTargetProvisioning = errors.New("database managed target provisioning failed")
)

// ManagedTargetControlState reports whether a native driver could read the
// durable control record. Unknown is intentionally not an admissible state.
type ManagedTargetControlState uint8

const (
	ManagedTargetControlUnknown ManagedTargetControlState = iota
	ManagedTargetControlAbsent
	ManagedTargetControlPresent
	ManagedTargetControlUnreadable
)

// ManagedTargetObservation is a driver's non-mutating view of one derived
// target. A present relation must carry observed native and schema identities;
// a control record is supplied only when ControlState is Present.
type ManagedTargetObservation struct {
	NamespacePresent bool
	RelationPresent  bool
	ControlState     ManagedTargetControlState
	ControlRecord    ManagedTargetControlRecord
	NativeIdentity   NativeRelationIdentity
	Schema           ManagedTargetSchema
}

func (o ManagedTargetObservation) validate() error {
	if o.RelationPresent && !o.NamespacePresent {
		return ErrManagedTargetUnverifiable
	}
	if !o.RelationPresent && (o.NativeIdentity != (NativeRelationIdentity{}) || o.Schema != (ManagedTargetSchema{})) {
		return ErrManagedTargetUnverifiable
	}
	switch o.ControlState {
	case ManagedTargetControlAbsent:
		if o.ControlRecord != (ManagedTargetControlRecord{}) {
			return ErrManagedTargetUnverifiable
		}
	case ManagedTargetControlPresent:
		if err := o.ControlRecord.validate(); err != nil {
			return ErrManagedTargetUnverifiable
		}
	case ManagedTargetControlUnreadable:
		if o.ControlRecord != (ManagedTargetControlRecord{}) {
			return ErrManagedTargetUnverifiable
		}
	default:
		return ErrManagedTargetUnverifiable
	}
	if o.RelationPresent && o.ControlState != ManagedTargetControlUnreadable {
		if o.NativeIdentity.Kind == "" || o.NativeIdentity.Value == "" || o.NativeIdentity.validate() != nil || o.Schema.validate() != nil {
			return ErrManagedTargetUnverifiable
		}
	}
	return nil
}

// ManagedTargetProvisioningPlan is the immutable, typed authority for a
// create-or-assert operation. It cannot carry SQL, credentials, a driver,
// arbitrary target names, or an unasserted owner.
type ManagedTargetProvisioningPlan struct {
	owner  TargetOwner
	target ManagedTargetRef
	schema ManagedTargetSchema
}

// NewManagedTargetProvisioningPlan validates a mutation authority before any
// driver is called. The supplied owner must be exactly the owner from which the
// target was derived.
func NewManagedTargetProvisioningPlan(owner TargetOwner, target ManagedTargetRef, schema ManagedTargetSchema) (ManagedTargetProvisioningPlan, error) {
	plan := ManagedTargetProvisioningPlan{owner: owner, target: target, schema: schema}
	if err := plan.validate(); err != nil {
		return ManagedTargetProvisioningPlan{}, ErrManagedTargetPlanInvalid
	}
	return plan, nil
}

func (p ManagedTargetProvisioningPlan) validate() error {
	if err := p.owner.validate(); err != nil || p.target.validate() != nil || !p.owner.sameIdentity(p.target.owner) || p.schema.validate() != nil {
		return ErrManagedTargetPlanInvalid
	}
	return nil
}

// Owner returns the owner that every mutation must assert to the driver.
func (p ManagedTargetProvisioningPlan) Owner() TargetOwner { return p.owner }

// Target returns the derived target address this plan alone may create/assert.
func (p ManagedTargetProvisioningPlan) Target() ManagedTargetRef { return p.target }

// Schema returns the schema contract this plan asserts. A mismatch is refused,
// not repaired or evolved.
func (p ManagedTargetProvisioningPlan) Schema() ManagedTargetSchema { return p.schema }

// ManagedTargetLock is a target-scoped driver lock acquired before every
// observation and held through the final assertion. A native implementation
// uses the database's own cross-process coordination primitive; this contract
// never substitutes an unscoped process mutex for that proof.
type ManagedTargetLock interface {
	ReleaseManagedTargetLock()
}

// ManagedTargetProvisioningDriver is the small driver-neutral port used by the
// shared contract. A native driver owns target-scoped cross-process locking,
// database-specific DDL, and durable control-record storage later; it receives
// no generic SQL from here. CreateManagedTarget receives both the typed plan
// and its asserted owner, and callers must re-observe before a create is
// considered successful.
type ManagedTargetProvisioningDriver interface {
	AcquireManagedTargetLock(context.Context, ManagedTargetRef) (ManagedTargetLock, error)
	ObserveManagedTarget(context.Context, ManagedTargetRef) (ManagedTargetObservation, error)
	CreateManagedTarget(context.Context, ManagedTargetProvisioningPlan, TargetOwner) error
}

type managedTargetGate struct {
	token      chan struct{}
	references int
}

// ManagedTargetProvisioner serializes create-or-assert per derived physical
// target and enforces the fail-closed truth table. It stores no credential,
// display name, or driver observation after a call returns.
type ManagedTargetProvisioner struct {
	driver ManagedTargetProvisioningDriver

	mu    sync.Mutex
	gates map[string]*managedTargetGate
}

// NewManagedTargetProvisioner creates the shared state-machine executor around
// one driver port. It is intentionally not a database driver registration or a
// capability promotion.
func NewManagedTargetProvisioner(driver ManagedTargetProvisioningDriver) (*ManagedTargetProvisioner, error) {
	if isNilInterface(driver) {
		return nil, errors.New("database managed target provisioning driver is required")
	}
	return &ManagedTargetProvisioner{driver: driver, gates: make(map[string]*managedTargetGate)}, nil
}

// CreateOrAssert creates a target only from the absent-target/absent-control
// state. Every other state is asserted exactly or refused. It has no adoption,
// reconciliation, replacement, or schema-evolution behavior.
func (p *ManagedTargetProvisioner) CreateOrAssert(ctx context.Context, plan ManagedTargetProvisioningPlan) (ManagedTargetControlRecord, error) {
	if ctx == nil {
		return ManagedTargetControlRecord{}, ErrManagedTargetPlanInvalid
	}
	if err := ctx.Err(); err != nil {
		return ManagedTargetControlRecord{}, err
	}
	if p == nil || isNilInterface(p.driver) || plan.validate() != nil {
		return ManagedTargetControlRecord{}, ErrManagedTargetPlanInvalid
	}
	release, err := p.acquire(ctx, plan.target.lockKey())
	if err != nil {
		return ManagedTargetControlRecord{}, err
	}
	defer release()
	if err := ctx.Err(); err != nil {
		return ManagedTargetControlRecord{}, err
	}
	targetLock, err := p.acquireDriverLock(ctx, plan.target)
	if err != nil {
		return ManagedTargetControlRecord{}, err
	}
	defer targetLock.ReleaseManagedTargetLock()

	observation, err := p.observe(ctx, plan.target)
	if err != nil {
		return ManagedTargetControlRecord{}, err
	}
	create, record, err := assessManagedTargetObservation(plan, observation)
	if err != nil {
		return ManagedTargetControlRecord{}, err
	}
	if !create {
		return record, nil
	}
	if err := ctx.Err(); err != nil {
		return ManagedTargetControlRecord{}, err
	}
	createErr := p.driver.CreateManagedTarget(ctx, plan, plan.owner)
	record, reassertErr := p.reassertAfterCreate(ctx, plan)
	if reassertErr != nil {
		return ManagedTargetControlRecord{}, reassertErr
	}
	if createErr != nil {
		if contextErr := ctx.Err(); contextErr != nil {
			return ManagedTargetControlRecord{}, contextErr
		}
		return ManagedTargetControlRecord{}, ErrManagedTargetProvisioning
	}
	if contextErr := ctx.Err(); contextErr != nil {
		return ManagedTargetControlRecord{}, contextErr
	}
	return record, nil
}

func (p *ManagedTargetProvisioner) reassertAfterCreate(ctx context.Context, plan ManagedTargetProvisioningPlan) (ManagedTargetControlRecord, error) {
	observation, err := p.observe(context.WithoutCancel(ctx), plan.target)
	if err != nil {
		return ManagedTargetControlRecord{}, err
	}
	create, record, err := assessManagedTargetObservation(plan, observation)
	if err != nil {
		return ManagedTargetControlRecord{}, err
	}
	if create {
		return ManagedTargetControlRecord{}, ErrManagedTargetUnverifiable
	}
	return record, nil
}

func (p *ManagedTargetProvisioner) observe(ctx context.Context, target ManagedTargetRef) (ManagedTargetObservation, error) {
	if err := ctx.Err(); err != nil {
		return ManagedTargetObservation{}, err
	}
	observation, err := p.driver.ObserveManagedTarget(ctx, target)
	if err != nil {
		if contextErr := ctx.Err(); contextErr != nil {
			return ManagedTargetObservation{}, contextErr
		}
		return ManagedTargetObservation{}, ErrManagedTargetUnverifiable
	}
	if err := ctx.Err(); err != nil {
		return ManagedTargetObservation{}, err
	}
	return observation, nil
}

func (p *ManagedTargetProvisioner) acquireDriverLock(ctx context.Context, target ManagedTargetRef) (ManagedTargetLock, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	targetLock, err := p.driver.AcquireManagedTargetLock(ctx, target)
	if err != nil {
		if contextErr := ctx.Err(); contextErr != nil {
			return nil, contextErr
		}
		return nil, ErrManagedTargetUnverifiable
	}
	if isNilInterface(targetLock) {
		return nil, ErrManagedTargetUnverifiable
	}
	if err := ctx.Err(); err != nil {
		targetLock.ReleaseManagedTargetLock()
		return nil, err
	}
	return targetLock, nil
}

func assessManagedTargetObservation(plan ManagedTargetProvisioningPlan, observation ManagedTargetObservation) (bool, ManagedTargetControlRecord, error) {
	if observation.ControlState == ManagedTargetControlUnreadable {
		return false, ManagedTargetControlRecord{}, ErrManagedTargetOwnerUnreadable
	}
	if err := observation.validate(); err != nil {
		return false, ManagedTargetControlRecord{}, ErrManagedTargetUnverifiable
	}
	if !observation.RelationPresent {
		switch observation.ControlState {
		case ManagedTargetControlPresent:
			return false, ManagedTargetControlRecord{}, ErrManagedTargetOrphaned
		case ManagedTargetControlAbsent:
			if observation.NamespacePresent {
				return false, ManagedTargetControlRecord{}, ErrManagedTargetNameCollision
			}
			return true, ManagedTargetControlRecord{}, nil
		}
		return false, ManagedTargetControlRecord{}, ErrManagedTargetUnverifiable
	}
	if observation.ControlState == ManagedTargetControlAbsent {
		return false, ManagedTargetControlRecord{}, ErrManagedTargetOwnerMissing
	}
	record := observation.ControlRecord
	if !record.owner.sameIdentity(plan.owner) {
		return false, ManagedTargetControlRecord{}, ErrManagedTargetOwnerForeign
	}
	if !record.target.sameTarget(plan.target) {
		return false, ManagedTargetControlRecord{}, ErrManagedTargetNameCollision
	}
	if !record.schema.sameSchema(plan.schema) || !observation.Schema.sameSchema(record.schema) {
		return false, ManagedTargetControlRecord{}, ErrManagedTargetSchemaDrift
	}
	if record.native != observation.NativeIdentity {
		return false, ManagedTargetControlRecord{}, ErrManagedTargetReplaced
	}
	return false, record, nil
}

func (p *ManagedTargetProvisioner) acquire(ctx context.Context, key string) (func(), error) {
	p.mu.Lock()
	gate := p.gates[key]
	if gate == nil {
		gate = &managedTargetGate{token: make(chan struct{}, 1)}
		gate.token <- struct{}{}
		p.gates[key] = gate
	}
	gate.references++
	p.mu.Unlock()

	select {
	case <-ctx.Done():
		p.releaseReference(key, gate)
		return nil, ctx.Err()
	case <-gate.token:
	}
	if err := ctx.Err(); err != nil {
		gate.token <- struct{}{}
		p.releaseReference(key, gate)
		return nil, err
	}
	var once sync.Once
	return func() {
		once.Do(func() {
			gate.token <- struct{}{}
			p.releaseReference(key, gate)
		})
	}, nil
}

func (p *ManagedTargetProvisioner) releaseReference(key string, gate *managedTargetGate) {
	p.mu.Lock()
	defer p.mu.Unlock()
	gate.references--
	if gate.references == 0 && p.gates[key] == gate {
		delete(p.gates, key)
	}
}

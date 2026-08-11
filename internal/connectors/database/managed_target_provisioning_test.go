package database_test

import (
	"context"
	"errors"
	"strings"
	"sync"
	"testing"

	"polymetrics.ai/internal/connectors/database"
	"polymetrics.ai/internal/warehouse"
)

// TestManagedTargetProvisioningTruthTable is the shared, driver-neutral F2
// admission contract. The fake below is deliberately the only driver: this
// test proves state transitions rather than command exits or PostgreSQL DDL.
func TestManagedTargetProvisioningTruthTable(t *testing.T) {
	identity := database.ConnectionIdentity{
		WorkspaceID:  "workspace-1",
		ConnectorID:  "source-postgres",
		ConnectionID: "source-connection-1",
	}
	artifact, err := warehouse.NewArtifactRef(identity, "orders")
	if err != nil {
		t.Fatal(err)
	}
	owner, err := database.NewTargetOwner(artifact.Identity())
	if err != nil {
		t.Fatal(err)
	}
	ref, err := database.NewManagedTargetRef(owner, artifact)
	if err != nil {
		t.Fatal(err)
	}
	schema := testManagedTargetSchema(t, 1, 1)
	native := database.NativeRelationIdentity{Kind: "fixture-native-id", Value: "relation-1"}
	control := testManagedTargetControl(t, owner, ref, native, schema)
	created := database.ManagedTargetObservation{
		NamespacePresent: true,
		RelationPresent:  true,
		ControlState:     database.ManagedTargetControlPresent,
		ControlRecord:    control,
		NativeIdentity:   native,
		Schema:           schema,
	}
	plan, err := database.NewManagedTargetProvisioningPlan(owner, ref, schema)
	if err != nil {
		t.Fatal(err)
	}

	// Both physical names are opaque hashes. The structural identity remains
	// inspectable through its typed owner, but display and credential material
	// never becomes part of a physical namespace/relation name.
	for _, forbidden := range []string{
		identity.WorkspaceID,
		identity.ConnectorID,
		identity.ConnectionID,
		"display-name-that-must-not-leak",
		"credential-that-must-not-leak",
	} {
		if strings.Contains(ref.Namespace(), forbidden) || strings.Contains(ref.Relation(), forbidden) {
			t.Fatalf("managed target physical name leaked %q: namespace=%q relation=%q", forbidden, ref.Namespace(), ref.Relation())
		}
	}
	if len(ref.Namespace()) > 63 || len(ref.Relation()) > 63 {
		t.Fatalf("managed target physical names exceed the conservative 63-byte identifier budget: namespace=%q relation=%q", ref.Namespace(), ref.Relation())
	}
	if got := owner.Identity(); !got.SameIdentity(identity) {
		t.Fatalf("target owner identity = %#v, want source artifact identity %#v", got, identity)
	}

	foreignIdentity := identity
	foreignIdentity.ConnectionID = "other-source-connection"
	foreignOwner, err := database.NewTargetOwner(foreignIdentity)
	if err != nil {
		t.Fatal(err)
	}
	foreignArtifact, err := warehouse.NewArtifactRef(foreignIdentity, "orders")
	if err != nil {
		t.Fatal(err)
	}
	foreignRef, err := database.NewManagedTargetRef(foreignOwner, foreignArtifact)
	if err != nil {
		t.Fatal(err)
	}
	foreignControl := testManagedTargetControl(t, foreignOwner, foreignRef, native, schema)
	otherArtifact, err := warehouse.NewArtifactRef(identity, "other-orders")
	if err != nil {
		t.Fatal(err)
	}
	collidingRef, err := database.NewManagedTargetRef(owner, otherArtifact)
	if err != nil {
		t.Fatal(err)
	}
	collidingControl := testManagedTargetControl(t, owner, collidingRef, native, schema)
	replacementNative := database.NativeRelationIdentity{Kind: "fixture-native-id", Value: "relation-2"}
	driftedSchema := testManagedTargetSchema(t, 2, 2)

	tests := []struct {
		name            string
		observation     database.ManagedTargetObservation
		createResult    database.ManagedTargetObservation
		plan            database.ManagedTargetProvisioningPlan
		cancelled       bool
		concurrentCalls int
		wantError       error
		wantCreates     int
		wantCreated     bool
	}{
		{
			name: "absent namespace creates a new managed target",
			observation: database.ManagedTargetObservation{
				ControlState: database.ManagedTargetControlAbsent,
			},
			createResult: created,
			plan:         plan,
			wantCreates:  1,
			wantCreated:  true,
		},
		{
			name: "occupied namespace without a control record is a collision",
			observation: database.ManagedTargetObservation{
				NamespacePresent: true,
				ControlState:     database.ManagedTargetControlAbsent,
			},
			plan:        plan,
			wantError:   database.ErrManagedTargetNameCollision,
			wantCreates: 0,
		},
		{
			name:         "first create reasserts exact control state",
			observation:  database.ManagedTargetObservation{ControlState: database.ManagedTargetControlAbsent},
			createResult: created,
			plan:         plan,
			wantCreates:  1,
			wantCreated:  true,
		},
		{
			name:        "post-create foreign control record is refused",
			observation: database.ManagedTargetObservation{ControlState: database.ManagedTargetControlAbsent},
			createResult: database.ManagedTargetObservation{
				NamespacePresent: true,
				RelationPresent:  true,
				ControlState:     database.ManagedTargetControlPresent,
				ControlRecord:    foreignControl,
				NativeIdentity:   native,
				Schema:           schema,
			},
			plan:        plan,
			wantError:   database.ErrManagedTargetOwnerForeign,
			wantCreates: 1,
			wantCreated: true,
		},
		{
			name:        "repeat with correct owner is idempotently asserted",
			observation: created,
			plan:        plan,
			wantCreates: 0,
		},
		{
			name: "relation with missing owner record is refused",
			observation: database.ManagedTargetObservation{
				NamespacePresent: true,
				RelationPresent:  true,
				ControlState:     database.ManagedTargetControlAbsent,
				NativeIdentity:   native,
				Schema:           schema,
			},
			plan:        plan,
			wantError:   database.ErrManagedTargetOwnerMissing,
			wantCreates: 0,
		},
		{
			name: "foreign owner record is refused",
			observation: database.ManagedTargetObservation{
				NamespacePresent: true,
				RelationPresent:  true,
				ControlState:     database.ManagedTargetControlPresent,
				ControlRecord:    foreignControl,
				NativeIdentity:   native,
				Schema:           schema,
			},
			plan:        plan,
			wantError:   database.ErrManagedTargetOwnerForeign,
			wantCreates: 0,
		},
		{
			name: "unreadable owner record is refused",
			observation: database.ManagedTargetObservation{
				NamespacePresent: true,
				RelationPresent:  true,
				ControlState:     database.ManagedTargetControlUnreadable,
			},
			plan:        plan,
			wantError:   database.ErrManagedTargetOwnerUnreadable,
			wantCreates: 0,
		},
		{
			name: "name collision is refused rather than adopted",
			observation: database.ManagedTargetObservation{
				NamespacePresent: true,
				RelationPresent:  true,
				ControlState:     database.ManagedTargetControlPresent,
				ControlRecord:    collidingControl,
				NativeIdentity:   native,
				Schema:           schema,
			},
			plan:        plan,
			wantError:   database.ErrManagedTargetNameCollision,
			wantCreates: 0,
		},
		{
			name: "native identity replacement is refused",
			observation: database.ManagedTargetObservation{
				NamespacePresent: true,
				RelationPresent:  true,
				ControlState:     database.ManagedTargetControlPresent,
				ControlRecord:    control,
				NativeIdentity:   replacementNative,
				Schema:           schema,
			},
			plan:        plan,
			wantError:   database.ErrManagedTargetReplaced,
			wantCreates: 0,
		},
		{
			name: "schema hash or version drift is refused rather than evolved",
			observation: database.ManagedTargetObservation{
				NamespacePresent: true,
				RelationPresent:  true,
				ControlState:     database.ManagedTargetControlPresent,
				ControlRecord:    control,
				NativeIdentity:   native,
				Schema:           driftedSchema,
			},
			plan:        plan,
			wantError:   database.ErrManagedTargetSchemaDrift,
			wantCreates: 0,
		},
		{
			name: "orphaned control record never recreates a target",
			observation: database.ManagedTargetObservation{
				ControlState:  database.ManagedTargetControlPresent,
				ControlRecord: control,
			},
			plan:        plan,
			wantError:   database.ErrManagedTargetOrphaned,
			wantCreates: 0,
		},
		{
			name:        "untyped plan cannot mutate a target",
			observation: database.ManagedTargetObservation{ControlState: database.ManagedTargetControlAbsent},
			plan:        database.ManagedTargetProvisioningPlan{},
			wantError:   database.ErrManagedTargetPlanInvalid,
			wantCreates: 0,
		},
		{
			name:         "cancelled provisioning fails before mutation",
			observation:  database.ManagedTargetObservation{ControlState: database.ManagedTargetControlAbsent},
			createResult: created,
			plan:         plan,
			cancelled:    true,
			wantError:    context.Canceled,
			wantCreates:  0,
			wantCreated:  false,
		},
		{
			name:            "concurrent provisioning creates once then asserts",
			observation:     database.ManagedTargetObservation{ControlState: database.ManagedTargetControlAbsent},
			createResult:    created,
			plan:            plan,
			concurrentCalls: 12,
			wantCreates:     1,
			wantCreated:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			driver := newManagedTargetDriverFake(tt.observation, tt.createResult)
			provisioner, err := database.NewManagedTargetProvisioner(driver)
			if err != nil {
				t.Fatal(err)
			}
			ctx := context.Background()
			if tt.cancelled {
				cancelled, cancel := context.WithCancel(ctx)
				cancel()
				ctx = cancelled
			}

			if tt.concurrentCalls > 0 {
				driver.blockCreate()
				peerProvisioner, err := database.NewManagedTargetProvisioner(driver)
				if err != nil {
					t.Fatal(err)
				}
				type result struct {
					control database.ManagedTargetControlRecord
					err     error
				}
				results := make(chan result, tt.concurrentCalls)
				for index := 0; index < tt.concurrentCalls; index++ {
					activeProvisioner := provisioner
					if index%2 != 0 {
						activeProvisioner = peerProvisioner
					}
					go func() {
						got, provisionErr := activeProvisioner.CreateOrAssert(context.Background(), tt.plan)
						results <- result{control: got, err: provisionErr}
					}()
				}
				<-driver.createStarted
				driver.releaseCreate()
				for index := 0; index < tt.concurrentCalls; index++ {
					got := <-results
					if got.err != nil {
						t.Fatalf("concurrent CreateOrAssert() error = %v", got.err)
					}
					assertManagedTargetControl(t, got.control, control)
				}
			} else {
				got, provisionErr := provisioner.CreateOrAssert(ctx, tt.plan)
				if tt.wantError != nil {
					if !errors.Is(provisionErr, tt.wantError) {
						t.Fatalf("CreateOrAssert() error = %v, want %v", provisionErr, tt.wantError)
					}
				} else if provisionErr != nil {
					t.Fatalf("CreateOrAssert() error = %v", provisionErr)
				} else {
					assertManagedTargetControl(t, got, control)
				}
			}

			if got := driver.createCallCount(); got != tt.wantCreates {
				t.Fatalf("managed target create calls = %d, want %d", got, tt.wantCreates)
			}
			if tt.wantCreated && !driver.createdWithAssertedOwner(owner, ref) {
				t.Fatal("mutation did not receive the validated typed plan and its asserted owner")
			}
		})
	}

	// An asserted owner must match the target reference before the fake is ever
	// allowed to mutate. This is a construction-time refusal, not a runtime
	// preference that a driver can ignore.
	if _, err := database.NewManagedTargetProvisioningPlan(foreignOwner, ref, schema); !errors.Is(err, database.ErrManagedTargetPlanInvalid) {
		t.Fatalf("NewManagedTargetProvisioningPlan(mismatched owner) error = %v, want ErrManagedTargetPlanInvalid", err)
	}

	t.Run("cancellation while another provision owns the target lock fails closed", func(t *testing.T) {
		driver := newManagedTargetDriverFake(
			database.ManagedTargetObservation{ControlState: database.ManagedTargetControlAbsent},
			created,
		)
		driver.blockCreate()
		provisioner, err := database.NewManagedTargetProvisioner(driver)
		if err != nil {
			t.Fatal(err)
		}
		firstDone := make(chan error, 1)
		go func() {
			_, err := provisioner.CreateOrAssert(context.Background(), plan)
			firstDone <- err
		}()
		<-driver.createStarted
		cancelled, cancel := context.WithCancel(context.Background())
		cancel()
		if _, err := provisioner.CreateOrAssert(cancelled, plan); !errors.Is(err, context.Canceled) {
			t.Fatalf("CreateOrAssert(cancelled while locked) error = %v, want context.Canceled", err)
		}
		driver.releaseCreate()
		if err := <-firstDone; err != nil {
			t.Fatalf("first CreateOrAssert() error = %v", err)
		}
		if got := driver.createCallCount(); got != 1 {
			t.Fatalf("create calls after locked cancellation = %d, want 1", got)
		}
	})

	t.Run("post-create outcomes are reasserted under the target lock", func(t *testing.T) {
		foreignCreated := database.ManagedTargetObservation{
			NamespacePresent: true,
			RelationPresent:  true,
			ControlState:     database.ManagedTargetControlPresent,
			ControlRecord:    foreignControl,
			NativeIdentity:   native,
			Schema:           schema,
		}
		tests := []struct {
			name                string
			createResult        database.ManagedTargetObservation
			createErr           error
			cancelAfterMutation bool
			wantError           error
		}{
			{
				name:                "cancellation after a committed create is reasserted before returning",
				createResult:        created,
				cancelAfterMutation: true,
				wantError:           context.Canceled,
			},
			{
				name:         "driver error after a committed create is reasserted before returning",
				createResult: created,
				createErr:    errors.New("fake post-mutation create failure"),
				wantError:    database.ErrManagedTargetProvisioning,
			},
			{
				name:                "cancellation after a foreign mutation returns the ownership classification",
				createResult:        foreignCreated,
				cancelAfterMutation: true,
				wantError:           database.ErrManagedTargetOwnerForeign,
			},
			{
				name:         "driver error after a foreign mutation returns the ownership classification",
				createResult: foreignCreated,
				createErr:    errors.New("fake post-mutation create failure"),
				wantError:    database.ErrManagedTargetOwnerForeign,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				driver := newManagedTargetDriverFake(
					database.ManagedTargetObservation{ControlState: database.ManagedTargetControlAbsent},
					tt.createResult,
				)
				provisioner, err := database.NewManagedTargetProvisioner(driver)
				if err != nil {
					t.Fatal(err)
				}
				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()
				driver.configureCreateOutcome(tt.createErr, func() {
					if tt.cancelAfterMutation {
						cancel()
					}
				})

				if _, err := provisioner.CreateOrAssert(ctx, plan); !errors.Is(err, tt.wantError) {
					t.Fatalf("CreateOrAssert() error = %v, want %v", err, tt.wantError)
				}
				if got := driver.createCallCount(); got != 1 {
					t.Fatalf("create calls = %d, want 1", got)
				}
				if got := driver.observeCallCount(); got != 2 {
					t.Fatalf("observations after a create invocation = %d, want 2", got)
				}
				if !driver.observationsHeldTargetLock() {
					t.Fatal("a managed target observation ran without the target lock")
				}
			})
		}
	})
}

func testManagedTargetSchema(t *testing.T, version uint, marker byte) database.ManagedTargetSchema {
	t.Helper()
	var fingerprint database.SchemaFingerprint
	fingerprint[0] = marker
	schema, err := database.NewManagedTargetSchema(version, fingerprint)
	if err != nil {
		t.Fatal(err)
	}
	return schema
}

func testManagedTargetControl(t *testing.T, owner database.TargetOwner, ref database.ManagedTargetRef, native database.NativeRelationIdentity, schema database.ManagedTargetSchema) database.ManagedTargetControlRecord {
	t.Helper()
	control, err := database.NewManagedTargetControlRecord(owner, ref, native, schema)
	if err != nil {
		t.Fatal(err)
	}
	return control
}

func assertManagedTargetControl(t *testing.T, got, want database.ManagedTargetControlRecord) {
	t.Helper()
	if !got.Owner().Identity().SameIdentity(want.Owner().Identity()) ||
		got.Target().Namespace() != want.Target().Namespace() ||
		got.Target().Relation() != want.Target().Relation() ||
		got.NativeIdentity() != want.NativeIdentity() ||
		got.Schema() != want.Schema() {
		t.Fatalf("managed target control = %#v, want %#v", got, want)
	}
}

type managedTargetDriverFake struct {
	mu                  sync.Mutex
	observation         database.ManagedTargetObservation
	createResult        database.ManagedTargetObservation
	createCalls         int
	createPlans         []database.ManagedTargetProvisioningPlan
	createOwners        []database.TargetOwner
	targetLock          chan struct{}
	activeTargetLocks   int
	mutationWithoutLock bool
	observeCalls        int
	observationNoLock   bool
	createStarted       chan struct{}
	createRelease       chan struct{}
	createStartedOnce   sync.Once
	createAfterMutation func()
	createAfterError    error
}

func newManagedTargetDriverFake(observation, createResult database.ManagedTargetObservation) *managedTargetDriverFake {
	driver := &managedTargetDriverFake{
		observation:   observation,
		createResult:  createResult,
		createStarted: make(chan struct{}),
		targetLock:    make(chan struct{}, 1),
	}
	driver.targetLock <- struct{}{}
	return driver
}

func (f *managedTargetDriverFake) AcquireManagedTargetLock(ctx context.Context, _ database.ManagedTargetRef) (database.ManagedTargetLock, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-f.targetLock:
	}
	f.mu.Lock()
	f.activeTargetLocks++
	f.mu.Unlock()
	return &managedTargetDriverFakeLock{driver: f}, nil
}

func (f *managedTargetDriverFake) ObserveManagedTarget(ctx context.Context, _ database.ManagedTargetRef) (database.ManagedTargetObservation, error) {
	if err := ctx.Err(); err != nil {
		return database.ManagedTargetObservation{}, err
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	f.observeCalls++
	if f.activeTargetLocks != 1 {
		f.observationNoLock = true
	}
	return f.observation, nil
}

func (f *managedTargetDriverFake) CreateManagedTarget(ctx context.Context, plan database.ManagedTargetProvisioningPlan, owner database.TargetOwner) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	f.mu.Lock()
	f.createCalls++
	if f.activeTargetLocks != 1 {
		f.mutationWithoutLock = true
	}
	f.createPlans = append(f.createPlans, plan)
	f.createOwners = append(f.createOwners, owner)
	release := f.createRelease
	f.mu.Unlock()
	f.createStartedOnce.Do(func() { close(f.createStarted) })
	if release != nil {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-release:
		}
	}
	f.mu.Lock()
	f.observation = f.createResult
	afterMutation := f.createAfterMutation
	afterError := f.createAfterError
	f.mu.Unlock()
	if afterMutation != nil {
		afterMutation()
	}
	return afterError
}

func (f *managedTargetDriverFake) blockCreate() {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.createRelease = make(chan struct{})
}

func (f *managedTargetDriverFake) releaseCreate() {
	f.mu.Lock()
	release := f.createRelease
	f.mu.Unlock()
	if release != nil {
		close(release)
	}
}

func (f *managedTargetDriverFake) configureCreateOutcome(afterError error, afterMutation func()) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.createAfterMutation = afterMutation
	f.createAfterError = afterError
}

func (f *managedTargetDriverFake) createCallCount() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.createCalls
}

func (f *managedTargetDriverFake) observeCallCount() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.observeCalls
}

func (f *managedTargetDriverFake) observationsHeldTargetLock() bool {
	f.mu.Lock()
	defer f.mu.Unlock()
	return !f.observationNoLock
}

func (f *managedTargetDriverFake) createdWithAssertedOwner(owner database.TargetOwner, ref database.ManagedTargetRef) bool {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.mutationWithoutLock || len(f.createPlans) == 0 || len(f.createPlans) != len(f.createOwners) {
		return false
	}
	for index, plan := range f.createPlans {
		if !f.createOwners[index].Identity().SameIdentity(owner.Identity()) ||
			!plan.Owner().Identity().SameIdentity(owner.Identity()) ||
			plan.Target().Namespace() != ref.Namespace() ||
			plan.Target().Relation() != ref.Relation() {
			return false
		}
	}
	return true
}

type managedTargetDriverFakeLock struct {
	driver *managedTargetDriverFake
	once   sync.Once
}

func (l *managedTargetDriverFakeLock) ReleaseManagedTargetLock() {
	l.once.Do(func() {
		l.driver.mu.Lock()
		l.driver.activeTargetLocks--
		l.driver.mu.Unlock()
		l.driver.targetLock <- struct{}{}
	})
}

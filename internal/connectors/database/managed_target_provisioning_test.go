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
// test proves state transitions rather than command exits or driver-specific DDL.
func TestManagedTargetProvisioningTruthTable(t *testing.T) {
	identity := database.ConnectionIdentity{
		WorkspaceID:  "workspace-1",
		ConnectorID:  "source-database",
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
	ref, err := database.NewManagedTargetRef(owner, artifact, "stream-orders")
	if err != nil {
		t.Fatal(err)
	}
	targetDatabase := testTargetDatabase(t, "database-1")
	namespaceNative := testNativeNamespace(t, "namespace-1")
	schema := testManagedTargetSchema(t, 1, 1)
	native := database.NativeRelationIdentity{Kind: "fixture-native-id", Value: "relation-1"}
	namespaceOwner := testManagedTargetNamespaceOwner(t, owner, ref, targetDatabase, namespaceNative)
	control := testManagedTargetControl(t, owner, ref, targetDatabase, native, schema)
	created := database.ManagedTargetObservation{
		NamespacePresent: true,
		RelationPresent:  true,
		ControlState:     database.ManagedTargetControlPresent,
		ControlRecord:    control,
		NativeIdentity:   native,
		Schema:           schema,
	}
	plan, err := database.NewManagedTargetProvisioningPlan(owner, ref, targetDatabase, schema)
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
		"stream-orders",
		"database-1",
		"full_append",
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
	foreignRef, err := database.NewManagedTargetRef(foreignOwner, foreignArtifact, "stream-foreign")
	if err != nil {
		t.Fatal(err)
	}
	foreignControl := testManagedTargetControl(t, foreignOwner, foreignRef, targetDatabase, native, schema)
	foreignNamespaceOwner := testManagedTargetNamespaceOwner(t, foreignOwner, foreignRef, targetDatabase, namespaceNative)
	otherArtifact, err := warehouse.NewArtifactRef(identity, "other-orders")
	if err != nil {
		t.Fatal(err)
	}
	collidingRef, err := database.NewManagedTargetRef(owner, otherArtifact, "stream-orders-secondary")
	if err != nil {
		t.Fatal(err)
	}
	collidingControl := testManagedTargetControl(t, owner, collidingRef, targetDatabase, native, schema)
	secondPlan, err := database.NewManagedTargetProvisioningPlan(owner, collidingRef, targetDatabase, schema)
	if err != nil {
		t.Fatal(err)
	}
	secondCreated := database.ManagedTargetObservation{
		NamespacePresent: true,
		RelationPresent:  true,
		ControlState:     database.ManagedTargetControlPresent,
		ControlRecord:    collidingControl,
		NativeIdentity:   native,
		Schema:           schema,
	}
	replacementNative := database.NativeRelationIdentity{Kind: "fixture-native-id", Value: "relation-2"}
	replacementNamespaceNative := testNativeNamespace(t, "namespace-2")
	movedTargetDatabase := testTargetDatabase(t, "database-2")
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
			// A single control record made this state look ambiguous. It is the
			// certified defect: an owned namespace may legitimately contain a
			// second stream relation whose per-relation control is absent.
			name: "owned namespace allows second stream relation",
			observation: database.ManagedTargetObservation{
				NamespacePresent: true,
				ControlState:     database.ManagedTargetControlAbsent,
			},
			createResult: secondCreated,
			plan:         secondPlan,
			wantCreates:  1,
			wantCreated:  true,
		},
		{
			name: "namespace without an owner record is refused",
			observation: database.ManagedTargetObservation{
				NamespacePresent:    true,
				NamespaceOwnerState: database.ManagedTargetNamespaceOwnerAbsent,
				ControlState:        database.ManagedTargetControlAbsent,
			},
			plan:        plan,
			wantError:   database.ErrManagedTargetNamespaceOwnerMissing,
			wantCreates: 0,
		},
		{
			name: "foreign namespace owner record is refused",
			observation: database.ManagedTargetObservation{
				NamespacePresent:     true,
				NamespaceOwnerState:  database.ManagedTargetNamespaceOwnerPresent,
				NamespaceOwnerRecord: foreignNamespaceOwner,
				ControlState:         database.ManagedTargetControlAbsent,
			},
			plan:        plan,
			wantError:   database.ErrManagedTargetOwnerForeign,
			wantCreates: 0,
		},
		{
			name: "unreadable namespace owner record is refused",
			observation: database.ManagedTargetObservation{
				NamespacePresent:    true,
				NamespaceOwnerState: database.ManagedTargetNamespaceOwnerUnreadable,
				ControlState:        database.ManagedTargetControlAbsent,
			},
			plan:        plan,
			wantError:   database.ErrManagedTargetOwnerUnreadable,
			wantCreates: 0,
		},
		{
			name: "replaced namespace is refused",
			observation: database.ManagedTargetObservation{
				NamespacePresent: true,
				NamespaceNative:  replacementNamespaceNative,
				ControlState:     database.ManagedTargetControlAbsent,
			},
			plan:        plan,
			wantError:   database.ErrManagedTargetNamespaceReplaced,
			wantCreates: 0,
		},
		{
			name: "moved target database is refused",
			observation: database.ManagedTargetObservation{
				TargetDatabase:   movedTargetDatabase,
				NamespacePresent: true,
				ControlState:     database.ManagedTargetControlAbsent,
			},
			plan:        plan,
			wantError:   database.ErrManagedTargetMoved,
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

	if ref.Namespace() != collidingRef.Namespace() || ref.Relation() == collidingRef.Relation() {
		t.Fatalf("same owner stream addresses = (%q, %q), (%q, %q), want same namespace and distinct relations", ref.Namespace(), ref.Relation(), collidingRef.Namespace(), collidingRef.Relation())
	}
	renamedArtifact, err := warehouse.NewArtifactRef(identity, "orders-renamed")
	if err != nil {
		t.Fatal(err)
	}
	renamedRef, err := database.NewManagedTargetRef(owner, renamedArtifact, ref.StreamID())
	if err != nil {
		t.Fatal(err)
	}
	if renamedRef.Namespace() != ref.Namespace() || renamedRef.Relation() != ref.Relation() {
		t.Fatalf("renamed stream moved managed target: got (%q, %q), want (%q, %q)", renamedRef.Namespace(), renamedRef.Relation(), ref.Namespace(), ref.Relation())
	}
	newConnectionIdentity := identity
	newConnectionIdentity.ConnectionID = "source-connection-2"
	newConnectionOwner, err := database.NewTargetOwner(newConnectionIdentity)
	if err != nil {
		t.Fatal(err)
	}
	newConnectionArtifact, err := warehouse.NewArtifactRef(newConnectionIdentity, "orders")
	if err != nil {
		t.Fatal(err)
	}
	newConnectionRef, err := database.NewManagedTargetRef(newConnectionOwner, newConnectionArtifact, ref.StreamID())
	if err != nil {
		t.Fatal(err)
	}
	if newConnectionRef.Namespace() == ref.Namespace() {
		t.Fatalf("new connection reused namespace %q", ref.Namespace())
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			driver := newManagedTargetDriverFake(
				withManagedTargetNamespaceOwner(tt.observation, targetDatabase, namespaceNative, namespaceOwner),
				withManagedTargetNamespaceOwner(tt.createResult, targetDatabase, namespaceNative, namespaceOwner),
			)
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
					want := control
					if tt.plan.Target().Relation() == collidingRef.Relation() {
						want = collidingControl
					}
					assertManagedTargetControl(t, got, want)
				}
			}

			if got := driver.createCallCount(); got != tt.wantCreates {
				t.Fatalf("managed target create calls = %d, want %d", got, tt.wantCreates)
			}
			if tt.wantCreated && !driver.createdWithAssertedOwner(tt.plan.Owner(), tt.plan.Target()) {
				t.Fatal("mutation did not receive the validated typed plan and its asserted owner")
			}
		})
	}

	// An asserted owner must match the target reference before the fake is ever
	// allowed to mutate. This is a construction-time refusal, not a runtime
	// preference that a driver can ignore.
	if _, err := database.NewManagedTargetProvisioningPlan(foreignOwner, ref, targetDatabase, schema); !errors.Is(err, database.ErrManagedTargetPlanInvalid) {
		t.Fatalf("NewManagedTargetProvisioningPlan(mismatched owner) error = %v, want ErrManagedTargetPlanInvalid", err)
	}

	t.Run("cancellation while another provision owns the target lock fails closed", func(t *testing.T) {
		driver := newManagedTargetDriverFake(
			withManagedTargetNamespaceOwner(database.ManagedTargetObservation{ControlState: database.ManagedTargetControlAbsent}, targetDatabase, namespaceNative, namespaceOwner),
			withManagedTargetNamespaceOwner(created, targetDatabase, namespaceNative, namespaceOwner),
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
					withManagedTargetNamespaceOwner(database.ManagedTargetObservation{ControlState: database.ManagedTargetControlAbsent}, targetDatabase, namespaceNative, namespaceOwner),
					withManagedTargetNamespaceOwner(tt.createResult, targetDatabase, namespaceNative, namespaceOwner),
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

func TestManagedTargetProvisioningConcurrentStreamsShareNamespaceOwner(t *testing.T) {
	identity := database.ConnectionIdentity{
		WorkspaceID:  "workspace-1",
		ConnectorID:  "source-database",
		ConnectionID: "source-connection-1",
	}
	owner, err := database.NewTargetOwner(identity)
	if err != nil {
		t.Fatal(err)
	}
	targetDatabase := testTargetDatabase(t, "database-1")
	namespaceNative := testNativeNamespace(t, "namespace-1")
	schema := testManagedTargetSchema(t, 1, 1)
	plans := make([]database.ManagedTargetProvisioningPlan, 0, 2)
	for _, stream := range []struct {
		table string
		id    string
	}{
		{table: "orders", id: "stream-orders"},
		{table: "customers", id: "stream-customers"},
	} {
		artifact, err := warehouse.NewArtifactRef(identity, stream.table)
		if err != nil {
			t.Fatal(err)
		}
		ref, err := database.NewManagedTargetRef(owner, artifact, stream.id)
		if err != nil {
			t.Fatal(err)
		}
		plan, err := database.NewManagedTargetProvisioningPlan(owner, ref, targetDatabase, schema)
		if err != nil {
			t.Fatal(err)
		}
		plans = append(plans, plan)
	}
	if plans[0].Target().Namespace() != plans[1].Target().Namespace() || plans[0].Target().Relation() == plans[1].Target().Relation() {
		t.Fatalf("concurrent stream targets = (%q, %q), (%q, %q), want shared namespace and distinct relations", plans[0].Target().Namespace(), plans[0].Target().Relation(), plans[1].Target().Namespace(), plans[1].Target().Relation())
	}

	driver := newNamespaceProvisioningDriverFake(targetDatabase, namespaceNative)
	first, err := database.NewManagedTargetProvisioner(driver)
	if err != nil {
		t.Fatal(err)
	}
	second, err := database.NewManagedTargetProvisioner(driver)
	if err != nil {
		t.Fatal(err)
	}
	type result struct {
		control database.ManagedTargetControlRecord
		err     error
	}
	results := make(chan result, len(plans))
	start := make(chan struct{})
	for index, plan := range plans {
		provisioner := first
		if index%2 != 0 {
			provisioner = second
		}
		go func(plan database.ManagedTargetProvisioningPlan) {
			<-start
			control, err := provisioner.CreateOrAssert(context.Background(), plan)
			results <- result{control: control, err: err}
		}(plan)
	}
	close(start)
	for range plans {
		got := <-results
		if got.err != nil {
			t.Fatalf("concurrent CreateOrAssert() error = %v", got.err)
		}
		if got.control.Target().Namespace() != plans[0].Target().Namespace() {
			t.Fatalf("concurrent control namespace = %q, want %q", got.control.Target().Namespace(), plans[0].Target().Namespace())
		}
	}
	if got := driver.createCallCount(); got != len(plans) {
		t.Fatalf("concurrent stream creates = %d, want %d", got, len(plans))
	}
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

func TestManagedTargetProvisioningPlanCarriesSharedMapping(t *testing.T) {
	identity := database.ConnectionIdentity{
		WorkspaceID:  "workspace-1",
		ConnectorID:  "postgres",
		ConnectionID: "source-connection-1",
	}
	artifact, err := warehouse.NewArtifactRef(identity, "orders")
	if err != nil {
		t.Fatal(err)
	}
	owner, err := database.NewTargetOwner(identity)
	if err != nil {
		t.Fatal(err)
	}
	target, err := database.NewManagedTargetRef(owner, artifact, "stream-orders")
	if err != nil {
		t.Fatal(err)
	}
	mapping := testDatabaseWriteMapping(t, "source_id", "target_id")
	plan, err := database.NewManagedTargetProvisioningPlan(
		owner,
		target,
		testTargetDatabase(t, "database-1"),
		testManagedTargetSchema(t, 1, 1),
		mapping,
	)
	if err != nil {
		t.Fatalf("NewManagedTargetProvisioningPlan() error = %v", err)
	}
	got, ok := plan.Mapping()
	if !ok {
		t.Fatal("Mapping() did not preserve the shared DDL mapping")
	}
	columns := got.Columns()
	if len(columns) != 1 || columns[0].Source != "source_id" || columns[0].Target != "target_id" {
		t.Fatalf("Mapping() columns = %#v, want the sealed shared mapping", columns)
	}
}

func testTargetDatabase(t *testing.T, value string) database.TargetDatabaseIdentity {
	t.Helper()
	identity, err := database.NewTargetDatabaseIdentity("fixture-target-database", value)
	if err != nil {
		t.Fatal(err)
	}
	return identity
}

func testNativeNamespace(t *testing.T, value string) database.NativeNamespaceIdentity {
	t.Helper()
	identity, err := database.NewNativeNamespaceIdentity("fixture-native-namespace", value)
	if err != nil {
		t.Fatal(err)
	}
	return identity
}

func testManagedTargetNamespaceOwner(t *testing.T, owner database.TargetOwner, ref database.ManagedTargetRef, targetDatabase database.TargetDatabaseIdentity, native database.NativeNamespaceIdentity) database.ManagedTargetNamespaceOwnerRecord {
	t.Helper()
	record, err := database.NewManagedTargetNamespaceOwnerRecord(owner, ref, targetDatabase, native)
	if err != nil {
		t.Fatal(err)
	}
	return record
}

func withManagedTargetNamespaceOwner(observation database.ManagedTargetObservation, targetDatabase database.TargetDatabaseIdentity, native database.NativeNamespaceIdentity, owner database.ManagedTargetNamespaceOwnerRecord) database.ManagedTargetObservation {
	if observation.TargetDatabase == (database.TargetDatabaseIdentity{}) {
		observation.TargetDatabase = targetDatabase
	}
	if observation.NamespacePresent {
		if observation.NamespaceNative == (database.NativeNamespaceIdentity{}) {
			observation.NamespaceNative = native
		}
		if observation.NamespaceOwnerState == database.ManagedTargetNamespaceOwnerUnknown {
			observation.NamespaceOwnerState = database.ManagedTargetNamespaceOwnerPresent
			observation.NamespaceOwnerRecord = owner
		}
	}
	return observation
}

func testManagedTargetControl(t *testing.T, owner database.TargetOwner, ref database.ManagedTargetRef, targetDatabase database.TargetDatabaseIdentity, native database.NativeRelationIdentity, schema database.ManagedTargetSchema) database.ManagedTargetControlRecord {
	t.Helper()
	control, err := database.NewManagedTargetControlRecord(owner, ref, targetDatabase, native, schema)
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
		got.TargetDatabase().Kind() != want.TargetDatabase().Kind() ||
		got.TargetDatabase().Value() != want.TargetDatabase().Value() ||
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

type namespaceProvisioningDriverFake struct {
	mu              sync.Mutex
	targetDatabase  database.TargetDatabaseIdentity
	namespaceNative database.NativeNamespaceIdentity
	namespaceOwner  database.ManagedTargetNamespaceOwnerRecord
	controls        map[string]database.ManagedTargetControlRecord
	targetLock      chan struct{}
	createCalls     int
}

func newNamespaceProvisioningDriverFake(targetDatabase database.TargetDatabaseIdentity, namespaceNative database.NativeNamespaceIdentity) *namespaceProvisioningDriverFake {
	driver := &namespaceProvisioningDriverFake{
		targetDatabase:  targetDatabase,
		namespaceNative: namespaceNative,
		controls:        make(map[string]database.ManagedTargetControlRecord),
		targetLock:      make(chan struct{}, 1),
	}
	driver.targetLock <- struct{}{}
	return driver
}

func (f *namespaceProvisioningDriverFake) AcquireManagedTargetLock(ctx context.Context, _ database.ManagedTargetRef) (database.ManagedTargetLock, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-f.targetLock:
		return &namespaceProvisioningDriverFakeLock{driver: f}, nil
	}
}

func (f *namespaceProvisioningDriverFake) ObserveManagedTarget(ctx context.Context, target database.ManagedTargetRef) (database.ManagedTargetObservation, error) {
	if err := ctx.Err(); err != nil {
		return database.ManagedTargetObservation{}, err
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	observation := database.ManagedTargetObservation{
		TargetDatabase: f.targetDatabase,
		ControlState:   database.ManagedTargetControlAbsent,
	}
	if f.namespaceOwner == (database.ManagedTargetNamespaceOwnerRecord{}) {
		return observation, nil
	}
	observation.NamespacePresent = true
	observation.NamespaceNative = f.namespaceNative
	observation.NamespaceOwnerState = database.ManagedTargetNamespaceOwnerPresent
	observation.NamespaceOwnerRecord = f.namespaceOwner
	if control, ok := f.controls[target.Relation()]; ok {
		observation.RelationPresent = true
		observation.ControlState = database.ManagedTargetControlPresent
		observation.ControlRecord = control
		observation.NativeIdentity = control.NativeIdentity()
		observation.Schema = control.Schema()
	}
	return observation, nil
}

func (f *namespaceProvisioningDriverFake) CreateManagedTarget(ctx context.Context, plan database.ManagedTargetProvisioningPlan, owner database.TargetOwner) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	if !owner.Identity().SameIdentity(plan.Owner().Identity()) {
		return errors.New("fake received an unasserted owner")
	}
	if f.namespaceOwner == (database.ManagedTargetNamespaceOwnerRecord{}) {
		namespaceOwner, err := database.NewManagedTargetNamespaceOwnerRecord(plan.Owner(), plan.Target(), plan.TargetDatabase(), f.namespaceNative)
		if err != nil {
			return err
		}
		f.namespaceOwner = namespaceOwner
	}
	native := database.NativeRelationIdentity{Kind: "fixture-native-id", Value: "relation-" + plan.Target().StreamID()}
	control, err := database.NewManagedTargetControlRecord(plan.Owner(), plan.Target(), plan.TargetDatabase(), native, plan.Schema())
	if err != nil {
		return err
	}
	f.controls[plan.Target().Relation()] = control
	f.createCalls++
	return nil
}

func (f *namespaceProvisioningDriverFake) createCallCount() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.createCalls
}

type namespaceProvisioningDriverFakeLock struct {
	driver *namespaceProvisioningDriverFake
	once   sync.Once
}

func (l *namespaceProvisioningDriverFakeLock) ReleaseManagedTargetLock() {
	l.once.Do(func() { l.driver.targetLock <- struct{}{} })
}

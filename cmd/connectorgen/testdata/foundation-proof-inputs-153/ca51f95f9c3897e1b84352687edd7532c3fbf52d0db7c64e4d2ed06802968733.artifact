package database

import (
	"context"
	"errors"
	"reflect"
	"sync"

	"polymetrics.ai/internal/synccontract"
)

// Driver describes only a registered database adapter identity. It does not
// promise an open connection, read implementation, target DDL, write session,
// receipt, CDC, or public capability.
type Driver interface {
	DatabaseDriverDescriptor() DriverDescriptor
}

// DriverDescriptor is the registered counterpart of DriverDeclaration.
type DriverDescriptor struct {
	ID         string
	Protocol   string
	APIVersion uint
}

func (d DriverDescriptor) declaration() DriverDeclaration {
	return DriverDeclaration(d)
}

func (d DriverDescriptor) validate() error { return d.declaration().validate() }

// NativeAdmittedDriver joins a registered database driver to one or more
// shared #3810 descriptor/evidence admissions. Each returned admission is one
// concrete native leg, so a source-to-warehouse descriptor cannot also stand
// in for a warehouse-to-target descriptor. It deliberately does not require a
// source RunNativeSync method; source dispatch remains consumer-owned in
// synccontract.NativeSyncExecutor.
type NativeAdmittedDriver interface {
	Driver
	DatabaseNativeAdmissions() []DatabaseNativeAdmission
}

// ResolveWriteDriver resolves the same exact registered definition/driver
// identity as Resolve, then requires the explicit write-session port. A
// descriptor alone still grants no write session or receipt authority.
func (r *DriverRegistry) ResolveWriteDriver(ctx context.Context, definition Definition) (DatabaseWriteDriver, error) {
	driver, err := r.Resolve(ctx, definition)
	if err != nil {
		return nil, err
	}
	writeDriver, ok := driver.(DatabaseWriteDriver)
	if !ok || isNilInterface(writeDriver) {
		return nil, ErrDatabaseWriteSessionUnavailable
	}
	return writeDriver, nil
}

// DatabaseNativeAdmission binds one native descriptor to a sealed database
// warehouse leg.
type DatabaseNativeAdmission struct {
	descriptor synccontract.NativeSyncExecutorDescriptor
	leg        databaseWarehouseLeg
}

var _ synccontract.NativeExecutorAdmission = DatabaseNativeAdmission{}

// NewDatabaseInboundAdmission binds shared native evidence to a
// source-to-warehouse database leg.
func NewDatabaseInboundAdmission(admission synccontract.NativeExecutorAdmission) (DatabaseNativeAdmission, error) {
	return newDatabaseNativeAdmission(databaseWarehouseLegInbound, admission)
}

// NewDatabaseOutboundAdmission binds shared native evidence to a
// warehouse-to-target database leg.
func NewDatabaseOutboundAdmission(admission synccontract.NativeExecutorAdmission) (DatabaseNativeAdmission, error) {
	return newDatabaseNativeAdmission(databaseWarehouseLegOutbound, admission)
}

func newDatabaseNativeAdmission(leg databaseWarehouseLeg, admission synccontract.NativeExecutorAdmission) (DatabaseNativeAdmission, error) {
	if isNilNativeAdmission(admission) {
		return DatabaseNativeAdmission{}, errors.New("database native admission is required")
	}
	descriptor := admission.NativeSyncExecutorDescriptor()
	contract := databaseNativeCommandContract(descriptor)
	if err := contract.Validate(); err != nil {
		return DatabaseNativeAdmission{}, err
	}
	return DatabaseNativeAdmission{
		descriptor: cloneNativeExecutorDescriptor(descriptor),
		leg:        leg,
	}, nil
}

func (a DatabaseNativeAdmission) NativeSyncExecutorDescriptor() synccontract.NativeSyncExecutorDescriptor {
	return cloneNativeExecutorDescriptor(a.descriptor)
}

func (a DatabaseNativeAdmission) nativeCommandContract() synccontract.NativeCommandContract {
	return databaseNativeCommandContract(a.descriptor)
}

func (a DatabaseNativeAdmission) clone() DatabaseNativeAdmission {
	return DatabaseNativeAdmission{
		descriptor: cloneNativeExecutorDescriptor(a.descriptor),
		leg:        a.leg,
	}
}

func (a DatabaseNativeAdmission) matches(leg databaseWarehouseLeg, contract synccontract.NativeCommandContract) error {
	if a.leg != leg {
		return ErrNativeDriverAdmissionMismatch
	}
	return synccontract.ValidateNativeAdmission(a, contract)
}

func cloneNativeExecutorDescriptor(descriptor synccontract.NativeSyncExecutorDescriptor) synccontract.NativeSyncExecutorDescriptor {
	clone := descriptor
	clone.Modes = append([]synccontract.Mode(nil), descriptor.Modes...)
	return clone
}

func cloneDatabaseNativeAdmissions(admissions []DatabaseNativeAdmission) []DatabaseNativeAdmission {
	clone := make([]DatabaseNativeAdmission, len(admissions))
	for index := range admissions {
		clone[index] = admissions[index].clone()
	}
	return clone
}

func databaseNativeCommandContract(descriptor synccontract.NativeSyncExecutorDescriptor) synccontract.NativeCommandContract {
	return synccontract.NativeCommandContract{
		ContractVersion: synccontract.NativeCommandContractVersion,
		Protocol:        descriptor.Protocol,
		Command:         descriptor.Command,
		Executor:        descriptor.Executor,
		Modes:           append([]synccontract.Mode(nil), descriptor.Modes...),
	}
}

var (
	// ErrDriverNotRegistered prevents a definition declaration from becoming an
	// operation merely because its JSON is syntactically valid.
	ErrDriverNotRegistered = errors.New("database driver is not registered")
	// ErrNativeDriverAdmissionRequired prevents a descriptor-only reference
	// driver from being used as an admitted database operation.
	ErrNativeDriverAdmissionRequired = errors.New("database driver requires matching native executor admission")
	// ErrNativeDriverAdmissionMismatch prevents one native admission descriptor
	// from being treated as a different database warehouse leg.
	ErrNativeDriverAdmissionMismatch    = errors.New("database driver lacks a matching native executor admission")
	ErrNativeDriverAdmissionLegConflict = errors.New("database driver reuses one native executor admission across warehouse legs")
	// ErrDriverModeNotDeclared prevents a driver from using a closed sync mode
	// that its own database.json has not explicitly declared. An empty mode list
	// admits no operation, even if an object has matching native evidence.
	ErrDriverModeNotDeclared = errors.New("database driver does not declare the requested sync mode")
)

// DriverRegistry maps closed driver identities to actual registered objects.
// Its lock protects registration and concurrent resolution; it stores neither
// credentials nor mutable definition projections.
type DriverRegistry struct {
	mu               sync.RWMutex
	drivers          map[string]Driver
	nativeAdmissions map[string][]DatabaseNativeAdmission
}

// NewDriverRegistry creates a registry and validates every supplied driver.
func NewDriverRegistry(drivers ...Driver) (*DriverRegistry, error) {
	registry := &DriverRegistry{
		drivers:          make(map[string]Driver),
		nativeAdmissions: make(map[string][]DatabaseNativeAdmission),
	}
	for _, driver := range drivers {
		if err := registry.Register(driver); err != nil {
			return nil, err
		}
	}
	return registry, nil
}

// Register adds exactly one descriptor-valid driver for its stable ID.
func (r *DriverRegistry) Register(driver Driver) error {
	if r == nil {
		return errors.New("database driver registry is required")
	}
	if isNilDriver(driver) {
		return errors.New("database driver is required")
	}
	descriptor := driver.DatabaseDriverDescriptor()
	if err := descriptor.validate(); err != nil {
		return err
	}
	var admissions []DatabaseNativeAdmission
	if admitted, ok := driver.(NativeAdmittedDriver); ok && !isNilNativeAdmittedDriver(admitted) {
		admissions = cloneDatabaseNativeAdmissions(admitted.DatabaseNativeAdmissions())
		if err := validateDatabaseNativeAdmissionLegs(admissions); err != nil {
			return err
		}
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.drivers == nil {
		r.drivers = make(map[string]Driver)
	}
	if r.nativeAdmissions == nil {
		r.nativeAdmissions = make(map[string][]DatabaseNativeAdmission)
	}
	if _, exists := r.drivers[descriptor.ID]; exists {
		return errors.New("database driver is already registered")
	}
	if err := validateRegisteredDatabaseNativeAdmissionLegs(admissions, r.nativeAdmissions); err != nil {
		return err
	}
	r.drivers[descriptor.ID] = driver
	if len(admissions) > 0 {
		r.nativeAdmissions[descriptor.ID] = admissions
	}
	return nil
}

// Resolve requires an exact registered driver/protocol/API version match for a
// loaded definition. It does not itself admit an operation.
func (r *DriverRegistry) Resolve(ctx context.Context, definition Definition) (Driver, error) {
	if ctx == nil {
		return nil, errors.New("database driver resolution context is required")
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if r == nil {
		return nil, errors.New("database driver registry is required")
	}
	if err := definition.Validate(); err != nil {
		return nil, err
	}
	declaration := definition.Driver()
	r.mu.RLock()
	driver, exists := r.drivers[declaration.ID]
	r.mu.RUnlock()
	if !exists || isNilDriver(driver) {
		return nil, ErrDriverNotRegistered
	}
	if driver.DatabaseDriverDescriptor().declaration() != declaration {
		return nil, errors.New("registered database driver does not match the definition declaration")
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return driver, nil
}

// Admit validates both the definition's exact registered driver and a sealed,
// warehouse-bound layer-two command. A declaration-only Driver always fails;
// no capability, direct source-to-target path, or execution can be inferred
// from a manifest alone.
func (r *DriverRegistry) Admit(ctx context.Context, definition Definition, command DatabaseWarehouseCommand) (NativeAdmittedDriver, error) {
	if isNilInterface(command) {
		return nil, errors.New("database warehouse command is required")
	}
	contract := command.nativeCommandContract()
	leg := command.databaseWarehouseLeg()
	driver, err := r.Resolve(ctx, definition)
	if err != nil {
		return nil, err
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	admitted, ok := driver.(NativeAdmittedDriver)
	if !ok || isNilNativeAdmittedDriver(admitted) {
		return nil, ErrNativeDriverAdmissionRequired
	}
	if contract.Protocol != definition.Driver().Protocol {
		return nil, errors.New("native command protocol does not match the database definition")
	}
	if !definitionAdmitsModes(definition.AdmittedModes(), contract.Modes) {
		return nil, ErrDriverModeNotDeclared
	}
	if err := validateDriverNativeAdmission(r.registeredNativeAdmissions(definition.Driver().ID), leg, contract); err != nil {
		return nil, err
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return admitted, nil
}

func definitionAdmitsModes(declared, requested []synccontract.Mode) bool {
	if len(declared) == 0 || len(requested) == 0 {
		return false
	}
	allowed := make(map[synccontract.Mode]struct{}, len(declared))
	for _, mode := range declared {
		allowed[mode] = struct{}{}
	}
	for _, mode := range requested {
		if _, exists := allowed[mode]; !exists {
			return false
		}
	}
	return true
}

func (r *DriverRegistry) registeredNativeAdmissions(driverID string) []DatabaseNativeAdmission {
	r.mu.RLock()
	admissions := cloneDatabaseNativeAdmissions(r.nativeAdmissions[driverID])
	r.mu.RUnlock()
	return admissions
}

func validateDriverNativeAdmission(admissions []DatabaseNativeAdmission, leg databaseWarehouseLeg, contract synccontract.NativeCommandContract) error {
	if err := validateDatabaseNativeAdmissionLegs(admissions); err != nil {
		return err
	}
	for _, admission := range admissions {
		if err := admission.matches(leg, contract); err == nil {
			return nil
		}
	}
	return ErrNativeDriverAdmissionMismatch
}

func validateDatabaseNativeAdmissionLegs(admissions []DatabaseNativeAdmission) error {
	for index, admission := range admissions {
		for _, previous := range admissions[:index] {
			if !databaseNativeAdmissionLegsConflict(admission, previous) {
				continue
			}
			return ErrNativeDriverAdmissionLegConflict
		}
	}
	return nil
}

func validateRegisteredDatabaseNativeAdmissionLegs(admissions []DatabaseNativeAdmission, registered map[string][]DatabaseNativeAdmission) error {
	for _, existingAdmissions := range registered {
		for _, admission := range admissions {
			for _, existing := range existingAdmissions {
				if databaseNativeAdmissionLegsConflict(admission, existing) {
					return ErrNativeDriverAdmissionLegConflict
				}
			}
		}
	}
	return nil
}

func databaseNativeAdmissionLegsConflict(left, right DatabaseNativeAdmission) bool {
	return left.leg != right.leg && sameDatabaseNativeAdmissionContract(left, right)
}

func sameDatabaseNativeAdmissionContract(left, right DatabaseNativeAdmission) bool {
	return synccontract.ValidateNativeAdmission(left, right.nativeCommandContract()) == nil
}

func isNilDriver(driver Driver) bool {
	return isNilInterface(driver)
}

func isNilNativeAdmittedDriver(driver NativeAdmittedDriver) bool {
	return isNilInterface(driver)
}

func isNilNativeAdmission(admission synccontract.NativeExecutorAdmission) bool {
	return isNilInterface(admission)
}

func isNilInterface(value any) bool {
	if value == nil {
		return true
	}
	reflected := reflect.ValueOf(value)
	switch reflected.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
		return reflected.IsNil()
	default:
		return false
	}
}

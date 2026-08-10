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

// NativeAdmittedDriver joins a registered database driver to the shared #3810
// descriptor/evidence admission contract. It deliberately does not require a
// source RunNativeSync method; source dispatch remains consumer-owned in
// synccontract.NativeSyncExecutor.
type NativeAdmittedDriver interface {
	Driver
	synccontract.NativeExecutorAdmission
}

var (
	// ErrDriverNotRegistered prevents a definition declaration from becoming an
	// operation merely because its JSON is syntactically valid.
	ErrDriverNotRegistered = errors.New("database driver is not registered")
	// ErrNativeDriverAdmissionRequired prevents a descriptor-only reference
	// driver from being used as an admitted database operation.
	ErrNativeDriverAdmissionRequired = errors.New("database driver requires matching native executor admission")
)

// DriverRegistry maps closed driver identities to actual registered objects.
// Its lock protects registration and concurrent resolution; it stores neither
// credentials nor mutable definition projections.
type DriverRegistry struct {
	mu      sync.RWMutex
	drivers map[string]Driver
}

// NewDriverRegistry creates a registry and validates every supplied driver.
func NewDriverRegistry(drivers ...Driver) (*DriverRegistry, error) {
	registry := &DriverRegistry{drivers: make(map[string]Driver)}
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
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.drivers == nil {
		r.drivers = make(map[string]Driver)
	}
	if _, exists := r.drivers[descriptor.ID]; exists {
		return errors.New("database driver is already registered")
	}
	r.drivers[descriptor.ID] = driver
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

// Admit validates both the definition's exact registered driver and the shared
// native command descriptor/evidence. A declaration-only Driver always fails;
// no capability or execution path can be inferred from a manifest alone.
func (r *DriverRegistry) Admit(ctx context.Context, definition Definition, contract synccontract.NativeCommandContract) (NativeAdmittedDriver, error) {
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
	if err := synccontract.ValidateNativeAdmission(admitted, contract); err != nil {
		return nil, err
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return admitted, nil
}

func isNilDriver(driver Driver) bool {
	return isNilInterface(driver)
}

func isNilNativeAdmittedDriver(driver NativeAdmittedDriver) bool {
	return isNilInterface(driver)
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

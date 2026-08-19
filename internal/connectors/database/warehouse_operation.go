package database

import (
	"errors"

	"polymetrics.ai/internal/synccontract"
)

// DatabaseWarehouseCommand is sealed to this package. A database admission
// can be constructed only as one layer-two leg against the connector-agnostic
// warehouse mediator, never from a raw contract plus another connector.
type DatabaseWarehouseCommand interface {
	databaseWarehouseCommand()
	nativeCommandContract() synccontract.NativeCommandContract
	databaseWarehouseLeg() databaseWarehouseLeg
}

type databaseWarehouseLeg uint8

const (
	databaseWarehouseLegInbound databaseWarehouseLeg = iota + 1
	databaseWarehouseLegOutbound
)

// DatabaseInboundCommand admits a database extraction operation whose only
// far-side address is a shared warehouse artifact.
type DatabaseInboundCommand struct {
	inbound  WarehouseInboundRef
	contract synccontract.NativeCommandContract
}

// NewDatabaseInboundCommand constructs a non-executing, warehouse-bound
// database source admission record.
func NewDatabaseInboundCommand(inbound WarehouseInboundRef, contract synccontract.NativeCommandContract) (DatabaseInboundCommand, error) {
	if err := inbound.validate(); err != nil {
		return DatabaseInboundCommand{}, errors.New("database warehouse inbound command is invalid")
	}
	if err := contract.Validate(); err != nil {
		return DatabaseInboundCommand{}, err
	}
	return DatabaseInboundCommand{inbound: inbound, contract: cloneNativeCommandContract(contract)}, nil
}

func (DatabaseInboundCommand) databaseWarehouseCommand() {}

func (DatabaseInboundCommand) databaseWarehouseLeg() databaseWarehouseLeg {
	return databaseWarehouseLegInbound
}

func (c DatabaseInboundCommand) nativeCommandContract() synccontract.NativeCommandContract {
	return cloneNativeCommandContract(c.contract)
}

// Inbound returns the source-to-warehouse leg. No target is available from
// this command.
func (c DatabaseInboundCommand) Inbound() WarehouseInboundRef { return c.inbound }

// Contract returns a defensive native admission projection.
func (c DatabaseInboundCommand) Contract() synccontract.NativeCommandContract {
	return c.nativeCommandContract()
}

// DatabaseOutboundCommand admits a database apply operation whose only input
// is a shared warehouse artifact; it carries no source connector.
type DatabaseOutboundCommand struct {
	outbound WarehouseOutboundRef
	contract synccontract.NativeCommandContract
}

// NewDatabaseOutboundCommand constructs a non-executing, warehouse-bound
// database destination admission record.
func NewDatabaseOutboundCommand(outbound WarehouseOutboundRef, contract synccontract.NativeCommandContract) (DatabaseOutboundCommand, error) {
	if err := outbound.validate(); err != nil {
		return DatabaseOutboundCommand{}, errors.New("database warehouse outbound command is invalid")
	}
	if err := contract.Validate(); err != nil {
		return DatabaseOutboundCommand{}, err
	}
	return DatabaseOutboundCommand{outbound: outbound, contract: cloneNativeCommandContract(contract)}, nil
}

func (DatabaseOutboundCommand) databaseWarehouseCommand() {}

func (DatabaseOutboundCommand) databaseWarehouseLeg() databaseWarehouseLeg {
	return databaseWarehouseLegOutbound
}

func (c DatabaseOutboundCommand) nativeCommandContract() synccontract.NativeCommandContract {
	return cloneNativeCommandContract(c.contract)
}

// Outbound returns the warehouse-to-target leg. No source is available from
// this command.
func (c DatabaseOutboundCommand) Outbound() WarehouseOutboundRef { return c.outbound }

// Contract returns a defensive native admission projection.
func (c DatabaseOutboundCommand) Contract() synccontract.NativeCommandContract {
	return c.nativeCommandContract()
}

func cloneNativeCommandContract(contract synccontract.NativeCommandContract) synccontract.NativeCommandContract {
	clone := contract
	clone.Modes = append([]synccontract.Mode(nil), contract.Modes...)
	clone.Conformance.FixtureIDs = append([]string(nil), contract.Conformance.FixtureIDs...)
	return clone
}

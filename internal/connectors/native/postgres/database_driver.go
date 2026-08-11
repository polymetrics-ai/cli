package postgres

import "polymetrics.ai/internal/connectors/database"

// DatabaseDriver is PostgreSQL's compile-time reference seam for the shared
// typed database foundation. It advertises only the stable driver identity;
// it is not registered and implements no SQL execution, target provisioning,
// write session, receipt, or CDC capability. Those arrive in their separately
// gated F2/F3/F4/P-unit slices.
type DatabaseDriver struct{}

// DatabaseDriverDescriptor returns the closed PostgreSQL wire-driver identity
// expected by defs/postgres/database.json.
func (DatabaseDriver) DatabaseDriverDescriptor() database.DriverDescriptor {
	return database.DriverDescriptor{
		ID:         "postgres",
		Protocol:   "postgres-wire",
		APIVersion: 1,
	}
}

var _ database.Driver = DatabaseDriver{}

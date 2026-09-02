package nativeset

import (
	"polymetrics.ai/internal/connectors"
	nativedynamodb "polymetrics.ai/internal/connectors/native/dynamodb"
	nativemysql "polymetrics.ai/internal/connectors/native/mysql"
	nativepostgres "polymetrics.ai/internal/connectors/native/postgres"
)

type databaseFactory struct {
	name string
	new  func() connectors.Connector
}

var protectedDatabaseFactories = []databaseFactory{
	{
		name: "dynamodb",
		new:  func() connectors.Connector { return nativedynamodb.New() },
	},
	{
		name: "mysql",
		new:  func() connectors.Connector { return nativemysql.New() },
	},
	{
		name: "postgres",
		new:  func() connectors.Connector { return nativepostgres.New() },
	},
}

// DatabaseConnectorFor preserves the three existing native database
// registrations. All other connectors execute through their rendered engine
// bundle, so API registrations cannot overwrite a declarative executor.
func DatabaseConnectorFor(name string) (connectors.Connector, bool) {
	for _, factory := range protectedDatabaseFactories {
		if factory.name == name {
			return factory.new(), true
		}
	}
	return nil, false
}

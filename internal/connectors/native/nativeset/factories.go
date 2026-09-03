package nativeset

import (
	"polymetrics.ai/internal/connectors"
	bingads "polymetrics.ai/internal/connectors/native/bing-ads"
	nativedynamodb "polymetrics.ai/internal/connectors/native/dynamodb"
	nativefaker "polymetrics.ai/internal/connectors/native/faker"
	nativehubspot "polymetrics.ai/internal/connectors/native/hubspot"
	nativemysql "polymetrics.ai/internal/connectors/native/mysql"
	nativepostgres "polymetrics.ai/internal/connectors/native/postgres"
	tallyprime "polymetrics.ai/internal/connectors/native/tally-prime"
)

type nativeFactory struct {
	name string
	new  func() connectors.Connector
}

var protectedDatabaseFactories = []nativeFactory{
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

var protectedCompatibilityFactories = []nativeFactory{
	{name: "bing-ads", new: func() connectors.Connector { return bingads.New() }},
	{name: "faker", new: func() connectors.Connector { return nativefaker.New() }},
	{name: "hubspot", new: func() connectors.Connector { return nativehubspot.New() }},
	{name: "tally-prime", new: func() connectors.Connector { return tallyprime.New() }},
}

// DatabaseConnectorFor selects protected native implementations. The
// compatibility API set is deliberately explicit and outside the 28 S1A
// generic-migration cohort until each connector receives its own source lock.
func DatabaseConnectorFor(name string) (connectors.Connector, bool) {
	for _, factory := range append(protectedDatabaseFactories, protectedCompatibilityFactories...) {
		if factory.name == name {
			return factory.new(), true
		}
	}
	return nil, false
}

package nativeset

import (
	"polymetrics.ai/internal/connectors"
	amazonsqs "polymetrics.ai/internal/connectors/native/amazon-sqs"
	bingads "polymetrics.ai/internal/connectors/native/bing-ads"
	"polymetrics.ai/internal/connectors/native/dynamodb"
	"polymetrics.ai/internal/connectors/native/faker"
	"polymetrics.ai/internal/connectors/native/hubspot"
	"polymetrics.ai/internal/connectors/native/mysql"
	"polymetrics.ai/internal/connectors/native/postgres"
	tallyprime "polymetrics.ai/internal/connectors/native/tally-prime"
)

type Factory struct {
	Name string
	New  func() connectors.Connector
}

func Factories() []Factory {
	factories := []Factory{
		{Name: "amazon-sqs", New: func() connectors.Connector { return amazonsqs.New() }},
		{Name: "bing-ads", New: func() connectors.Connector { return bingads.New() }},
		{Name: "dynamodb", New: func() connectors.Connector { return dynamodb.New() }},
		{Name: "faker", New: func() connectors.Connector { return faker.New() }},
		{Name: "hubspot", New: func() connectors.Connector { return hubspot.New() }},
		{Name: "mysql", New: func() connectors.Connector { return mysql.New() }},
		{Name: "postgres", New: func() connectors.Connector { return postgres.New() }},
		{Name: "tally-prime", New: func() connectors.Connector { return tallyprime.New() }},
	}
	return append(factories, promotedFactories()...)
}

func RegisterInto(registry *connectors.Registry) {
	for _, factory := range Factories() {
		registry.Register(factory.New())
	}
}

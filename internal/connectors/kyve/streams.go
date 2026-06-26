package kyve

import "polymetrics.ai/internal/connectors"

type streamEndpoint struct {
	resource    string
	recordsPath string
	mapRecord   func(map[string]any) connectors.Record
}

var streamEndpoints = map[string]streamEndpoint{
	"pools":      {resource: "kyve/query/v1beta1/pools", recordsPath: "pools", mapRecord: poolRecord},
	"stakers":    {resource: "kyve/query/v1beta1/stakers", recordsPath: "stakers", mapRecord: accountRecord},
	"funders":    {resource: "kyve/query/v1beta1/funders", recordsPath: "funders", mapRecord: accountRecord},
	"validators": {resource: "cosmos/staking/v1beta1/validators", recordsPath: "validators", mapRecord: validatorRecord},
}

func streams() []connectors.Stream {
	return []connectors.Stream{
		{Name: "pools", Description: "KYVE storage pools.", PrimaryKey: []string{"id"}, Fields: []connectors.Field{{Name: "id", Type: "string"}, {Name: "name", Type: "string"}, {Name: "runtime", Type: "string"}}},
		{Name: "stakers", Description: "KYVE stakers.", PrimaryKey: []string{"address"}, Fields: []connectors.Field{{Name: "address", Type: "string"}, {Name: "amount", Type: "string"}}},
		{Name: "funders", Description: "KYVE funders.", PrimaryKey: []string{"address"}, Fields: []connectors.Field{{Name: "address", Type: "string"}, {Name: "amount", Type: "string"}}},
		{Name: "validators", Description: "Cosmos validators for the KYVE network endpoint.", PrimaryKey: []string{"operator_address"}, Fields: []connectors.Field{{Name: "operator_address", Type: "string"}, {Name: "moniker", Type: "string"}, {Name: "status", Type: "string"}}},
	}
}

func poolRecord(item map[string]any) connectors.Record {
	return connectors.Record{"id": item["id"], "name": item["name"], "runtime": item["runtime"]}
}
func accountRecord(item map[string]any) connectors.Record {
	return connectors.Record{"address": first(item, "address", "account"), "amount": first(item, "amount", "balance")}
}
func validatorRecord(item map[string]any) connectors.Record {
	return connectors.Record{"operator_address": item["operator_address"], "moniker": nested(item, "description", "moniker"), "status": item["status"]}
}
func first(item map[string]any, keys ...string) any {
	for _, key := range keys {
		if item[key] != nil {
			return item[key]
		}
	}
	return nil
}
func nested(item map[string]any, objectKey, field string) any {
	obj, ok := item[objectKey].(map[string]any)
	if !ok {
		return nil
	}
	return obj[field]
}

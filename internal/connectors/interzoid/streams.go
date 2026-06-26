package interzoid

import "polymetrics.ai/internal/connectors"

// inputParam describes one config-driven request parameter a stream sends to its
// Interzoid lookup endpoint: paramName is the upstream query key, configKey is
// the RuntimeConfig.Config key it is read from, and required marks it mandatory.
type inputParam struct {
	paramName string
	configKey string
	required  bool
}

// streamEndpoint maps a stream to its fixed Interzoid lookup endpoint, the
// config-driven inputs it forwards, and the mapper that flattens the single
// JSON object the API returns into a connectors.Record. Interzoid lookups return
// one object at the response root (no array, no pagination), so each Read yields
// exactly one record.
type streamEndpoint struct {
	resource  string
	inputs    []inputParam
	mapRecord func(item map[string]any, echo map[string]string) connectors.Record
}

// interzoidStreamEndpoints is the per-stream routing table. Endpoints and
// parameter names mirror the upstream Airbyte manifest
// (airbyte/source-interzoid): getcompanymatchadvanced, getfullnamematch,
// getaddressmatchadvanced, getorgstandard.
var interzoidStreamEndpoints = map[string]streamEndpoint{
	"company_name_matching": {
		resource: "getcompanymatchadvanced",
		inputs: []inputParam{
			{paramName: "company", configKey: "company", required: true},
			{paramName: "algorithm", configKey: "company_match_algorithm", required: false},
		},
		mapRecord: simKeyRecord("company", "company"),
	},
	"individual_name_matching": {
		resource: "getfullnamematch",
		inputs: []inputParam{
			{paramName: "fullname", configKey: "fullname", required: true},
		},
		mapRecord: simKeyRecord("fullname", "fullname"),
	},
	"street_address_matching": {
		resource: "getaddressmatchadvanced",
		inputs: []inputParam{
			{paramName: "address", configKey: "address", required: true},
			{paramName: "algorithm", configKey: "address_match_algorithm", required: false},
		},
		mapRecord: simKeyRecord("address", "address"),
	},
	"standardize_company_names": {
		resource: "getorgstandard",
		inputs: []inputParam{
			{paramName: "org", configKey: "org", required: true},
		},
		mapRecord: standardRecord,
	},
}

// interzoidStreams returns the connector's published stream catalog. Each lookup
// is keyed by the similarity key (SimKey) it returns, except the standardize
// stream which is keyed by the standardized output (Standard). There is no
// incremental cursor: every stream is full-refresh only.
func interzoidStreams() []connectors.Stream {
	return []connectors.Stream{
		{
			Name:        "company_name_matching",
			Description: "Interzoid company-name similarity key lookup (getcompanymatchadvanced).",
			PrimaryKey:  []string{"SimKey"},
			Fields:      simKeyFields("query_company"),
		},
		{
			Name:        "individual_name_matching",
			Description: "Interzoid individual full-name similarity key lookup (getfullnamematch).",
			PrimaryKey:  []string{"SimKey"},
			Fields:      simKeyFields("query_fullname"),
		},
		{
			Name:        "street_address_matching",
			Description: "Interzoid street-address similarity key lookup (getaddressmatchadvanced).",
			PrimaryKey:  []string{"SimKey"},
			Fields:      simKeyFields("query_address"),
		},
		{
			Name:        "standardize_company_names",
			Description: "Interzoid organization-name standardization (getorgstandard).",
			PrimaryKey:  []string{"Standard"},
			Fields: []connectors.Field{
				{Name: "Code", Type: "string"},
				{Name: "Standard", Type: "string"},
				{Name: "Credits", Type: "string"},
				{Name: "query_org", Type: "string"},
			},
		},
	}
}

func simKeyFields(echoField string) []connectors.Field {
	return []connectors.Field{
		{Name: "Code", Type: "string"},
		{Name: "SimKey", Type: "string"},
		{Name: "Credits", Type: "string"},
		{Name: echoField, Type: "string"},
	}
}

// simKeyRecord builds a mapper for the three similarity-key streams. echoConfig
// is the config key whose input value is echoed back onto the record under
// query_<echoName> so downstream consumers can join the SimKey to its input.
func simKeyRecord(echoName, echoConfig string) func(map[string]any, map[string]string) connectors.Record {
	return func(item map[string]any, echo map[string]string) connectors.Record {
		return connectors.Record{
			"Code":              item["Code"],
			"SimKey":            item["SimKey"],
			"Credits":           item["Credits"],
			"query_" + echoName: echo[echoConfig],
		}
	}
}

func standardRecord(item map[string]any, echo map[string]string) connectors.Record {
	return connectors.Record{
		"Code":      item["Code"],
		"Standard":  item["Standard"],
		"Credits":   item["Credits"],
		"query_org": echo["org"],
	}
}

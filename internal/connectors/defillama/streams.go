package defillama

import "polymetrics.ai/internal/connectors"

// hostKind selects which DefiLlama host base a stream reads from. DefiLlama
// splits its public API across a few hostnames (the main api.llama.fi, plus
// dedicated stablecoins host). Each stream declares the host it belongs to so
// the read path can resolve the right base URL (and per-host override config).
type hostKind int

const (
	hostMain        hostKind = iota // api.llama.fi
	hostStablecoins                 // stablecoins.llama.fi
)

// streamEndpoint maps a stream name to the DefiLlama resource path, the host it
// lives on, the dotted JSON path where its record array lives, and the record
// mapper that flattens its objects into a connectors.Record.
type streamEndpoint struct {
	// resource is the path segment relative to the resolved host base.
	resource string
	// host selects which DefiLlama host base to use.
	host hostKind
	// recordsPath is the dotted JSON path to the records array. "" means the
	// response body is itself a top-level array.
	recordsPath string
	// query holds fixed query params appended to every request for the stream.
	query map[string]string
	// mapRecord flattens a raw DefiLlama object into a connectors.Record.
	mapRecord func(map[string]any) connectors.Record
	// paginated indicates the connector should page the (full) array returned by
	// DefiLlama client-side with limit/offset to keep payloads bounded.
	paginated bool
}

// defillamaStreamEndpoints is the per-stream routing table. Adding a stream is a
// matter of adding one entry here plus a Stream definition in defillamaStreams.
var defillamaStreamEndpoints = map[string]streamEndpoint{
	"protocols": {
		resource:    "protocols",
		host:        hostMain,
		recordsPath: "",
		mapRecord:   protocolRecord,
		paginated:   true,
	},
	"chains": {
		resource:    "v2/chains",
		host:        hostMain,
		recordsPath: "",
		mapRecord:   chainRecord,
		paginated:   true,
	},
	"stablecoins": {
		resource:    "stablecoins",
		host:        hostStablecoins,
		recordsPath: "peggedAssets",
		query:       map[string]string{"includePrices": "true"},
		mapRecord:   stablecoinRecord,
	},
	"dexs": {
		resource:    "overview/dexs",
		host:        hostMain,
		recordsPath: "protocols",
		query: map[string]string{
			"excludeTotalDataChart":          "true",
			"excludeTotalDataChartBreakdown": "true",
		},
		mapRecord: overviewRecord,
	},
	"fees": {
		resource:    "overview/fees",
		host:        hostMain,
		recordsPath: "protocols",
		query: map[string]string{
			"excludeTotalDataChart":          "true",
			"excludeTotalDataChartBreakdown": "true",
		},
		mapRecord: overviewRecord,
	},
}

// defillamaStreams returns the connector's published stream catalog. DefiLlama is
// a full-refresh-only public API, so streams have no incremental cursor field.
func defillamaStreams() []connectors.Stream {
	return []connectors.Stream{
		{
			Name:        "protocols",
			Description: "All DeFi protocols tracked by DefiLlama with current TVL.",
			PrimaryKey:  []string{"id"},
			Fields:      protocolFields(),
		},
		{
			Name:        "chains",
			Description: "Current TVL of every chain tracked by DefiLlama.",
			PrimaryKey:  []string{"name"},
			Fields:      chainFields(),
		},
		{
			Name:        "stablecoins",
			Description: "All stablecoins with circulating supply and peg metadata.",
			PrimaryKey:  []string{"id"},
			Fields:      stablecoinFields(),
		},
		{
			Name:        "dexs",
			Description: "DEX protocols with trading volume overview metrics.",
			PrimaryKey:  []string{"defillamaId"},
			Fields:      overviewFields(),
		},
		{
			Name:        "fees",
			Description: "Protocols with fees and revenue overview metrics.",
			PrimaryKey:  []string{"defillamaId"},
			Fields:      overviewFields(),
		},
	}
}

func protocolFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "slug", Type: "string"},
		{Name: "symbol", Type: "string"},
		{Name: "category", Type: "string"},
		{Name: "chain", Type: "string"},
		{Name: "chains", Type: "array"},
		{Name: "tvl", Type: "number"},
		{Name: "mcap", Type: "number"},
		{Name: "change_1d", Type: "number"},
		{Name: "change_7d", Type: "number"},
		{Name: "url", Type: "string"},
	}
}

func chainFields() []connectors.Field {
	return []connectors.Field{
		{Name: "name", Type: "string"},
		{Name: "gecko_id", Type: "string"},
		{Name: "tokenSymbol", Type: "string"},
		{Name: "cmcId", Type: "string"},
		{Name: "chainId", Type: "number"},
		{Name: "tvl", Type: "number"},
	}
}

func stablecoinFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "symbol", Type: "string"},
		{Name: "gecko_id", Type: "string"},
		{Name: "pegType", Type: "string"},
		{Name: "pegMechanism", Type: "string"},
		{Name: "price", Type: "number"},
		{Name: "circulating", Type: "object"},
	}
}

func overviewFields() []connectors.Field {
	return []connectors.Field{
		{Name: "defillamaId", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "displayName", Type: "string"},
		{Name: "category", Type: "string"},
		{Name: "chains", Type: "array"},
		{Name: "total24h", Type: "number"},
		{Name: "total7d", Type: "number"},
		{Name: "total30d", Type: "number"},
		{Name: "totalAllTime", Type: "number"},
		{Name: "change_1d", Type: "number"},
	}
}

func protocolRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":        item["id"],
		"name":      item["name"],
		"slug":      item["slug"],
		"symbol":    item["symbol"],
		"category":  item["category"],
		"chain":     item["chain"],
		"chains":    item["chains"],
		"tvl":       item["tvl"],
		"mcap":      item["mcap"],
		"change_1d": item["change_1d"],
		"change_7d": item["change_7d"],
		"url":       item["url"],
	}
}

func chainRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"name":        item["name"],
		"gecko_id":    item["gecko_id"],
		"tokenSymbol": item["tokenSymbol"],
		"cmcId":       item["cmcId"],
		"chainId":     item["chainId"],
		"tvl":         item["tvl"],
	}
}

func stablecoinRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":           item["id"],
		"name":         item["name"],
		"symbol":       item["symbol"],
		"gecko_id":     item["gecko_id"],
		"pegType":      item["pegType"],
		"pegMechanism": item["pegMechanism"],
		"price":        item["price"],
		"circulating":  item["circulating"],
	}
}

func overviewRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"defillamaId":  item["defillamaId"],
		"name":         item["name"],
		"displayName":  item["displayName"],
		"category":     item["category"],
		"chains":       item["chains"],
		"total24h":     item["total24h"],
		"total7d":      item["total7d"],
		"total30d":     item["total30d"],
		"totalAllTime": item["totalAllTime"],
		"change_1d":    item["change_1d"],
	}
}

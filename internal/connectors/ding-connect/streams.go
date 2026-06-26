package dingconnect

import "polymetrics.ai/internal/connectors"

// streamEndpoint maps a stream name to the DingConnect API resource path
// (relative to base_url) it reads from, plus the record mapper that flattens its
// items. The DingConnect API exposes read-only catalog/reference endpoints under
// /api/V1 that each return a JSON envelope {"Items":[...]}.
type streamEndpoint struct {
	// resource is the DingConnect endpoint path segment (e.g. "GetCountries").
	resource string
	// mapRecord flattens a raw DingConnect item into a connectors.Record.
	mapRecord func(map[string]any) connectors.Record
}

// dingStreamEndpoints is the per-stream routing table. Adding a stream means
// adding one entry here plus a Stream definition in dingStreams; the read path is
// fully data-driven from this table. The resource is the path under /api/V1.
var dingStreamEndpoints = map[string]streamEndpoint{
	"countries":  {resource: "GetCountries", mapRecord: dingCountryRecord},
	"currencies": {resource: "GetCurrencies", mapRecord: dingCurrencyRecord},
	"regions":    {resource: "GetRegions", mapRecord: dingRegionRecord},
	"providers":  {resource: "GetProviders", mapRecord: dingProviderRecord},
	"products":   {resource: "GetProducts", mapRecord: dingProductRecord},
}

// dingStreams returns the connector's published stream catalog. DingConnect
// reference resources do not carry a natural id; the Airbyte source assigns a
// synthetic primary key "uuid", which this connector mirrors. These resources are
// full-refresh (no incremental cursor).
func dingStreams() []connectors.Stream {
	return []connectors.Stream{
		{
			Name:        "countries",
			Description: "Countries supported by DingConnect, with ISO codes and dialing info.",
			PrimaryKey:  []string{"uuid"},
			Fields:      dingCountryFields(),
		},
		{
			Name:        "currencies",
			Description: "Currencies supported by DingConnect.",
			PrimaryKey:  []string{"uuid"},
			Fields:      dingCurrencyFields(),
		},
		{
			Name:        "regions",
			Description: "Regions (administrative areas) within supported countries.",
			PrimaryKey:  []string{"uuid"},
			Fields:      dingRegionFields(),
		},
		{
			Name:        "providers",
			Description: "Top-up / service providers available through DingConnect.",
			PrimaryKey:  []string{"uuid"},
			Fields:      dingProviderFields(),
		},
		{
			Name:        "products",
			Description: "Products (SKUs) offered by DingConnect providers.",
			PrimaryKey:  []string{"uuid"},
			Fields:      dingProductFields(),
		},
	}
}

func dingCountryFields() []connectors.Field {
	return []connectors.Field{
		{Name: "uuid", Type: "string"},
		{Name: "CountryIso", Type: "string"},
		{Name: "CountryName", Type: "string"},
		{Name: "InternationalDialingInformation", Type: "object"},
		{Name: "RegionCodes", Type: "array"},
	}
}

func dingCurrencyFields() []connectors.Field {
	return []connectors.Field{
		{Name: "uuid", Type: "string"},
		{Name: "CurrencyIso", Type: "string"},
		{Name: "CurrencyName", Type: "string"},
	}
}

func dingRegionFields() []connectors.Field {
	return []connectors.Field{
		{Name: "uuid", Type: "string"},
		{Name: "CountryIso", Type: "string"},
		{Name: "RegionCode", Type: "string"},
		{Name: "RegionName", Type: "string"},
	}
}

func dingProviderFields() []connectors.Field {
	return []connectors.Field{
		{Name: "uuid", Type: "string"},
		{Name: "ProviderCode", Type: "string"},
		{Name: "Name", Type: "string"},
		{Name: "CountryIso", Type: "string"},
		{Name: "RegionCodes", Type: "array"},
		{Name: "PaymentTypes", Type: "array"},
		{Name: "CustomerCareNumber", Type: "string"},
		{Name: "LogoUrl", Type: "string"},
		{Name: "ValidationRegex", Type: "string"},
	}
}

func dingProductFields() []connectors.Field {
	return []connectors.Field{
		{Name: "uuid", Type: "string"},
		{Name: "SkuCode", Type: "string"},
		{Name: "ProviderCode", Type: "string"},
		{Name: "RegionCode", Type: "string"},
		{Name: "DefaultDisplayText", Type: "string"},
		{Name: "LocalizationKey", Type: "string"},
		{Name: "ProcessingMode", Type: "string"},
		{Name: "RedemptionMechanism", Type: "string"},
		{Name: "CommissionRate", Type: "number"},
		{Name: "Minimum", Type: "object"},
		{Name: "Maximum", Type: "object"},
		{Name: "Benefits", Type: "array"},
		{Name: "PaymentTypes", Type: "array"},
		{Name: "ValidityPeriodIso", Type: "string"},
	}
}

func dingCountryRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"uuid":                            dingUUID(item, "CountryIso"),
		"CountryIso":                      item["CountryIso"],
		"CountryName":                     item["CountryName"],
		"InternationalDialingInformation": item["InternationalDialingInformation"],
		"RegionCodes":                     item["RegionCodes"],
	}
}

func dingCurrencyRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"uuid":         dingUUID(item, "CurrencyIso"),
		"CurrencyIso":  item["CurrencyIso"],
		"CurrencyName": item["CurrencyName"],
	}
}

func dingRegionRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"uuid":       dingUUID(item, "CountryIso", "RegionCode"),
		"CountryIso": item["CountryIso"],
		"RegionCode": item["RegionCode"],
		"RegionName": item["RegionName"],
	}
}

func dingProviderRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"uuid":               dingUUID(item, "ProviderCode"),
		"ProviderCode":       item["ProviderCode"],
		"Name":               item["Name"],
		"CountryIso":         item["CountryIso"],
		"RegionCodes":        item["RegionCodes"],
		"PaymentTypes":       item["PaymentTypes"],
		"CustomerCareNumber": item["CustomerCareNumber"],
		"LogoUrl":            item["LogoUrl"],
		"ValidationRegex":    item["ValidationRegex"],
	}
}

func dingProductRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"uuid":                dingUUID(item, "SkuCode"),
		"SkuCode":             item["SkuCode"],
		"ProviderCode":        item["ProviderCode"],
		"RegionCode":          item["RegionCode"],
		"DefaultDisplayText":  item["DefaultDisplayText"],
		"LocalizationKey":     item["LocalizationKey"],
		"ProcessingMode":      item["ProcessingMode"],
		"RedemptionMechanism": item["RedemptionMechanism"],
		"CommissionRate":      item["CommissionRate"],
		"Minimum":             item["Minimum"],
		"Maximum":             item["Maximum"],
		"Benefits":            item["Benefits"],
		"PaymentTypes":        item["PaymentTypes"],
		"ValidityPeriodIso":   item["ValidityPeriodIso"],
	}
}

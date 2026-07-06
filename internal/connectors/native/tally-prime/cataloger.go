package tallyprime

import (
	"context"

	"polymetrics.ai/internal/connectors"
)

// collectionDef pairs a stream name with the TDL Collection definition that
// exports it: the TallyPrime object TYPE, the FETCH field list (TallyPrime
// wire field names, in the order requested), and the Collection NAME/ID sent
// as the envelope's HEADER.ID and COLLECTION NAME attribute.
type collectionDef struct {
	stream    string
	id        string
	tallyType string
	fetch     []string
}

// collectionDefs is the fixed set of core-object collections this connector
// supports, keyed by stream name (design's "core objects: companies,
// ledgers, groups, stock_items, vouchers"). This is TallyPrime's equivalent
// of a fixed single-stream catalog (cf. native/amazon-sqs's messages
// stream) — a hand-written, non-runtime-discovered set of collections, even
// though capabilities.dynamic_schema is true for the structural reason
// docs.md explains (no streams.json, since the engine's declarative dialect
// cannot express TDL Collection requests).
var collectionDefs = map[string]collectionDef{
	"companies": {
		stream:    "companies",
		id:        "ID_Companies",
		tallyType: "Company",
		fetch:     []string{"NAME", "STARTINGFROM", "BOOKSFROM", "STATENAME", "PINCODE"},
	},
	"ledgers": {
		stream:    "ledgers",
		id:        "ID_Ledgers",
		tallyType: "Ledger",
		fetch:     []string{"NAME", "PARENT", "OPENINGBALANCE", "CLOSINGBALANCE", "ISBILLWISEON"},
	},
	"groups": {
		stream:    "groups",
		id:        "ID_Groups",
		tallyType: "Group",
		fetch:     []string{"NAME", "PARENT", "ISREVENUE", "ISDEEMEDPOSITIVE", "AFFECTSGROSSPROFIT"},
	},
	"stock_items": {
		stream:    "stock_items",
		id:        "ID_StockItems",
		tallyType: "StockItem",
		fetch:     []string{"NAME", "PARENT", "BASEUNITS", "CLOSINGBALANCE", "CLOSINGVALUE"},
	},
	"vouchers": {
		stream:    "vouchers",
		id:        "ID_Vouchers",
		tallyType: "Voucher",
		fetch:     []string{"GUID", "VOUCHERNUMBER", "VOUCHERTYPENAME", "DATE", "PARTYNAME", "AMOUNT", "NARRATION"},
	},
}

// streamOrder is collectionDefs' deterministic iteration order (design's
// listed core-object order), used by Catalog so repeated calls return
// streams in a stable order.
var streamOrder = []string{"companies", "ledgers", "groups", "stock_items", "vouchers"}

// Catalog returns the five core-object streams. TallyPrime has no schema
// registry to introspect at runtime — the field lists are fixed by this
// connector's collectionDefs, matching native/amazon-sqs's identical
// fixed-catalog-under-dynamic_schema precedent (see docs.md's Overview for
// why dynamic_schema is still true structurally).
func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	if _, err := resolveConfig(cfg); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: streamDefinitions()}, nil
}

// streamDefinitions builds the connectors.Stream list for every
// collectionDef, in streamOrder.
func streamDefinitions() []connectors.Stream {
	streams := make([]connectors.Stream, 0, len(streamOrder))
	for _, name := range streamOrder {
		streams = append(streams, streamDefinition(name))
	}
	return streams
}

func streamDefinition(name string) connectors.Stream {
	switch name {
	case "companies":
		return connectors.Stream{
			Name:        "companies",
			Description: "TallyPrime companies (TDL Collection TYPE=Company). Always a full export; no incremental cursor.",
			PrimaryKey:  []string{"name"},
			Fields: []connectors.Field{
				{Name: "name", Type: "string"},
				{Name: "starting_from", Type: "string"},
				{Name: "books_from", Type: "string"},
				{Name: "state_name", Type: "string"},
				{Name: "pincode", Type: "string"},
			},
		}
	case "ledgers":
		return connectors.Stream{
			Name:        "ledgers",
			Description: "TallyPrime ledgers (TDL Collection TYPE=Ledger). Always a full export; no incremental cursor.",
			PrimaryKey:  []string{"name"},
			Fields: []connectors.Field{
				{Name: "name", Type: "string"},
				{Name: "parent", Type: "string"},
				{Name: "opening_balance", Type: "number"},
				{Name: "closing_balance", Type: "number"},
				{Name: "is_bill_wise_on", Type: "boolean"},
			},
		}
	case "groups":
		return connectors.Stream{
			Name:        "groups",
			Description: "TallyPrime account groups (TDL Collection TYPE=Group). Always a full export; no incremental cursor.",
			PrimaryKey:  []string{"name"},
			Fields: []connectors.Field{
				{Name: "name", Type: "string"},
				{Name: "parent", Type: "string"},
				{Name: "is_revenue", Type: "boolean"},
				{Name: "is_deemedpositive", Type: "boolean"},
				{Name: "affects_gross_profit", Type: "boolean"},
			},
		}
	case "stock_items":
		return connectors.Stream{
			Name:        "stock_items",
			Description: "TallyPrime stock items (TDL Collection TYPE=StockItem). Always a full export; no incremental cursor.",
			PrimaryKey:  []string{"name"},
			Fields: []connectors.Field{
				{Name: "name", Type: "string"},
				{Name: "parent", Type: "string"},
				{Name: "base_units", Type: "string"},
				{Name: "closing_balance", Type: "number"},
				{Name: "closing_value", Type: "number"},
			},
		}
	case "vouchers":
		return connectors.Stream{
			Name:         "vouchers",
			Description:  "TallyPrime transaction vouchers (TDL Collection TYPE=Voucher). Supports incremental reads on date, plus optional from_date/to_date config bounds sent as SVFROMDATE/SVTODATE.",
			PrimaryKey:   []string{"guid"},
			CursorFields: []string{"date"},
			Fields: []connectors.Field{
				{Name: "guid", Type: "string"},
				{Name: "voucher_number", Type: "string"},
				{Name: "voucher_type", Type: "string"},
				{Name: "date", Type: "string"},
				{Name: "party_name", Type: "string"},
				{Name: "amount", Type: "number"},
				{Name: "narration", Type: "string"},
			},
		}
	default:
		return connectors.Stream{Name: name}
	}
}

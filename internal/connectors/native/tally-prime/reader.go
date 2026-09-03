package tallyprime

import (
	"context"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"strconv"
	"strings"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/connsdk"
)

// InitialState satisfies connectors.StatefulReader. Every stream starts with
// an empty incremental cursor (full export); only vouchers ever advances it
// past "" (master streams have no server-side "modified since" concept —
// see docs.md's Streams notes), mirroring native/postgres's InitialState
// convention.
func (c Connector) InitialState(ctx context.Context, stream string, cfg connectors.RuntimeConfig) (map[string]string, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return connsdk.WithCursor(map[string]string{"stream": stream}, ""), nil
}

// Read exports one collection (req.Stream) and emits its records. Master
// streams (companies/ledgers/groups/stock_items) always emit a full
// snapshot; vouchers additionally filters out records at or before the
// incremental cursor lower bound carried in req.State, and advances a
// returned cursor is not tracked here (the caller persists State from the
// last emitted record the same way native/postgres's cursor convention
// works — this connector does not mutate req.State itself, matching
// connectors.Reader's emit-only contract).
func (c Connector) Read(ctx context.Context, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if req.Stream == "" {
		return fmt.Errorf("tally-prime read requires a stream (one of companies, ledgers, groups, stock_items, vouchers)")
	}
	def, ok := collectionDefs[req.Stream]
	if !ok {
		return fmt.Errorf("tally-prime stream %q not found", req.Stream)
	}

	if fixtureMode(req.Config) {
		return c.readFixture(ctx, req, def, emit)
	}

	conn, err := resolveConfig(req.Config)
	if err != nil {
		return err
	}

	body, err := c.postCollection(ctx, req.Config, conn, def)
	if err != nil {
		return fmt.Errorf("read tally-prime %s: %w", req.Stream, err)
	}

	records, err := decodeCollection(body, conn.envelopeFormat, def)
	if err != nil {
		return fmt.Errorf("read tally-prime %s: %w", req.Stream, err)
	}

	lowerBound := connsdk.Cursor(req.State)
	for _, rec := range records {
		if err := ctx.Err(); err != nil {
			return err
		}
		if req.Stream == "vouchers" && lowerBound != "" {
			if date, _ := rec["date"].(string); date != "" && date <= lowerBound {
				continue
			}
		}
		if err := emit(rec); err != nil {
			return err
		}
	}
	return nil
}

// --- response decoding ---------------------------------------------------

// decodeCollection dispatches to the JSON or XML response decoder for one
// collectionDef's exported field list, based on conn.envelopeFormat.
func decodeCollection(body []byte, format string, def collectionDef) ([]connectors.Record, error) {
	if format == "xml" {
		return decodeCollectionXML(body, def)
	}
	return decodeCollectionJSON(body, def)
}

// jsonEnvelopeResponse mirrors TallyPrime's native JSON export response
// shape: an ENVELOPE.BODY.DATA.COLLECTION object whose keys are the
// COLLECTION NAME and whose values are the exported rows, each row a map of
// FETCH field name -> value (docs.md's Streams notes).
type jsonEnvelopeResponse struct {
	Envelope struct {
		Body struct {
			Data struct {
				Collection map[string][]map[string]json.RawMessage `json:"COLLECTION"`
			} `json:"DATA"`
		} `json:"BODY"`
	} `json:"ENVELOPE"`
}

func decodeCollectionJSON(body []byte, def collectionDef) ([]connectors.Record, error) {
	var resp jsonEnvelopeResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("decode tally-prime json envelope: %w", err)
	}
	rows := resp.Envelope.Body.Data.Collection[def.id]
	records := make([]connectors.Record, 0, len(rows))
	for _, row := range rows {
		records = append(records, mapRow(def, func(field string) (string, bool) {
			raw, ok := row[field]
			if !ok {
				return "", false
			}
			return rawJSONValue(raw), true
		}))
	}
	return records, nil
}

// rawJSONValue unwraps a json.RawMessage into a plain string, whether the
// source value was a JSON string or a bare number/bool.
func rawJSONValue(raw json.RawMessage) string {
	var s string
	if err := json.Unmarshal(raw, &s); err == nil {
		return s
	}
	return strings.Trim(string(raw), `"`)
}

// xmlEnvelopeResponse mirrors the XML fallback response shape: an
// ENVELOPE containing one <COLLECTION.NAME> element per exported row's
// field set, where each child element name is a FETCH field.
type xmlEnvelopeResponse struct {
	XMLName xml.Name        `xml:"ENVELOPE"`
	Rows    []xmlCollection `xml:"COLLECTION"`
}

// xmlCollection captures one exported row as a flat set of named child
// elements (TallyPrime's XML export emits one element per FETCH field,
// named identically to the FETCH request).
type xmlCollection struct {
	Fields []xmlField `xml:",any"`
}

type xmlField struct {
	XMLName xml.Name
	Value   string `xml:",chardata"`
}

func decodeCollectionXML(body []byte, def collectionDef) ([]connectors.Record, error) {
	var resp xmlEnvelopeResponse
	if err := xml.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("decode tally-prime xml envelope: %w", err)
	}
	records := make([]connectors.Record, 0, len(resp.Rows))
	for _, row := range resp.Rows {
		values := map[string]string{}
		for _, f := range row.Fields {
			values[strings.ToUpper(f.XMLName.Local)] = f.Value
		}
		records = append(records, mapRow(def, func(field string) (string, bool) {
			v, ok := values[field]
			return v, ok
		}))
	}
	return records, nil
}

// mapRow projects one collectionDef's FETCH field values (looked up via get)
// onto a connectors.Record, using the stream's coarse field-type vocabulary
// (numbers/booleans parsed best-effort; everything else stays a string).
func mapRow(def collectionDef, get func(field string) (string, bool)) connectors.Record {
	rec := connectors.Record{}
	stream := streamDefinition(def.stream)
	fieldTypes := map[string]string{}
	for _, f := range stream.Fields {
		fieldTypes[strings.ToUpper(wireFieldName(def, f.Name))] = f.Type
	}
	for _, field := range def.fetch {
		raw, ok := get(field)
		outName := outputFieldName(field)
		if !ok {
			rec[outName] = nil
			continue
		}
		rec[outName] = coerce(raw, fieldTypes[field])
	}
	return rec
}

// wireFieldName maps an output field name (e.g. "opening_balance") back to
// its wire FETCH name (e.g. "OPENINGBALANCE") for the given collectionDef,
// by scanning def.fetch for the matching name (case-insensitive match on
// the flattened form).
func wireFieldName(def collectionDef, outputName string) string {
	target := strings.ReplaceAll(outputName, "_", "")
	for _, wire := range def.fetch {
		if strings.EqualFold(strings.ReplaceAll(wire, "_", ""), target) {
			return wire
		}
	}
	return strings.ToUpper(outputName)
}

// outputFieldName maps a TallyPrime wire FETCH field name (e.g.
// "OPENINGBALANCE") to this connector's snake_case output name (e.g.
// "opening_balance"), via a fixed table (TallyPrime field names have no
// general-purpose casing convention to derive this from mechanically).
var wireToOutputField = map[string]string{
	"NAME":               "name",
	"STARTINGFROM":       "starting_from",
	"BOOKSFROM":          "books_from",
	"STATENAME":          "state_name",
	"PINCODE":            "pincode",
	"PARENT":             "parent",
	"OPENINGBALANCE":     "opening_balance",
	"CLOSINGBALANCE":     "closing_balance",
	"ISBILLWISEON":       "is_bill_wise_on",
	"ISREVENUE":          "is_revenue",
	"ISDEEMEDPOSITIVE":   "is_deemedpositive",
	"AFFECTSGROSSPROFIT": "affects_gross_profit",
	"BASEUNITS":          "base_units",
	"CLOSINGVALUE":       "closing_value",
	"GUID":               "guid",
	"VOUCHERNUMBER":      "voucher_number",
	"VOUCHERTYPENAME":    "voucher_type",
	"DATE":               "date",
	"PARTYNAME":          "party_name",
	"AMOUNT":             "amount",
	"NARRATION":          "narration",
}

func outputFieldName(wire string) string {
	if name, ok := wireToOutputField[wire]; ok {
		return name
	}
	return strings.ToLower(wire)
}

// coerce converts a raw wire string into the connectors.Field type declared
// for this output field ("number"/"boolean" parsed best-effort; unparsable
// or "string" values pass through unchanged).
func coerce(raw, fieldType string) any {
	switch fieldType {
	case "number":
		if f, err := strconv.ParseFloat(strings.TrimSpace(raw), 64); err == nil {
			return f
		}
		return raw
	case "boolean":
		trimmed := strings.ToLower(strings.TrimSpace(raw))
		switch trimmed {
		case "yes", "true", "1":
			return true
		case "no", "false", "0", "":
			return false
		default:
			return raw
		}
	default:
		return raw
	}
}

// --- fixture mode ---------------------------------------------------------

// readFixture emits canned rows for a fixture stream without any network
// access, applying the same vouchers-only cursor-filter behavior as the live
// path so cursor semantics are exercised credential-free (mirrors
// native/postgres.readFixture's convention).
func (c Connector) readFixture(ctx context.Context, req connectors.ReadRequest, def collectionDef, emit func(connectors.Record) error) error {
	rows := fixtureRows(def.stream)
	lowerBound := connsdk.Cursor(req.State)
	for _, rec := range rows {
		if err := ctx.Err(); err != nil {
			return err
		}
		if def.stream == "vouchers" && lowerBound != "" {
			if date, _ := rec["date"].(string); date != "" && date <= lowerBound {
				continue
			}
		}
		if err := emit(copyRecord(rec)); err != nil {
			return err
		}
	}
	return nil
}

// fixtureRows returns deterministic canned rows for a fixture stream.
func fixtureRows(stream string) []connectors.Record {
	switch stream {
	case "companies":
		return []connectors.Record{
			{"name": "Acme Retail Pvt Ltd", "starting_from": "20250401", "books_from": "20250401", "state_name": "Karnataka", "pincode": "560001"},
		}
	case "ledgers":
		return []connectors.Record{
			{"name": "Sales Account", "parent": "Sales Accounts", "opening_balance": 0.0, "closing_balance": 125000.50, "is_bill_wise_on": false},
			{"name": "ABC Distributors", "parent": "Sundry Debtors", "opening_balance": 5000.0, "closing_balance": 18250.75, "is_bill_wise_on": true},
		}
	case "groups":
		return []connectors.Record{
			{"name": "Sundry Debtors", "parent": "Current Assets", "is_revenue": false, "is_deemedpositive": true, "affects_gross_profit": false},
			{"name": "Sales Accounts", "parent": "Primary", "is_revenue": true, "is_deemedpositive": false, "affects_gross_profit": true},
		}
	case "stock_items":
		return []connectors.Record{
			{"name": "Widget A", "parent": "Finished Goods", "base_units": "Nos", "closing_balance": 240.0, "closing_value": 48000.0},
		}
	case "vouchers":
		return []connectors.Record{
			{"guid": "fixture-guid-1", "voucher_number": "1001", "voucher_type": "Sales", "date": "20250601", "party_name": "ABC Distributors", "amount": 12500.0, "narration": "Fixture sales voucher 1"},
			{"guid": "fixture-guid-2", "voucher_number": "1002", "voucher_type": "Sales", "date": "20250615", "party_name": "ABC Distributors", "amount": 5750.75, "narration": "Fixture sales voucher 2"},
		}
	default:
		return nil
	}
}

// copyRecord returns a shallow copy of a record so emitted fixture rows are
// not aliased between calls (mirrors native/postgres.copyRecord).
func copyRecord(in connectors.Record) connectors.Record {
	out := make(connectors.Record, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
}

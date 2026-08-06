package postgres

import (
	"encoding/binary"
	"encoding/json"
	"errors"
	"reflect"
	"testing"

	"polymetrics.ai/internal/connectors"
)

const testRelationID uint32 = 42

type testColumn struct {
	name   string
	typeID uint32
}

type tupleField struct {
	kind  byte
	value string
}

func TestPGOutputDecoderDML(t *testing.T) {
	columns := []testColumn{
		{name: "id", typeID: 23},
		{name: "email", typeID: 25},
		{name: "active", typeID: 16},
		{name: "score", typeID: 1700},
	}

	cases := []struct {
		name    string
		message []byte
		lsn     string
		want    connectors.CDCEvent
	}{
		{
			name: "insert",
			message: insertMessage(testRelationID,
				textField("7"),
				textField("ada@example.invalid"),
				textField("t"),
				textField("98.5"),
			),
			lsn: "0/16B6C50",
			want: connectors.CDCEvent{
				Operation: "insert",
				Record: connectors.Record{
					"id":     7,
					"email":  "ada@example.invalid",
					"active": true,
					"score":  "98.5",
				},
				State: connectors.Record{"lsn": "0/16B6C50"},
			},
		},
		{
			name:    "update",
			message: updateMessage(testRelationID, nil, []tupleField{textField("7"), textField("grace@example.invalid"), textField("f"), textField("99.25")}),
			want: connectors.CDCEvent{
				Operation: "update",
				Record: connectors.Record{
					"id":     7,
					"email":  "grace@example.invalid",
					"active": false,
					"score":  "99.25",
				},
			},
		},
		{
			name:    "delete with key tuple",
			message: deleteMessage(testRelationID, 'K', textField("7")),
			want: connectors.CDCEvent{
				Operation: "delete",
				Record:    connectors.Record{"id": 7},
			},
		},
		{
			name: "update null and unchanged toast",
			message: updateMessage(testRelationID, nil, []tupleField{
				textField("8"),
				nullField(),
				unchangedField(),
				textField("12.75"),
			}),
			want: connectors.CDCEvent{
				Operation: "update",
				Record: connectors.Record{
					"id":    8,
					"email": nil,
					"score": "12.75",
				},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			dec := newPGOutputDecoder("public", "users")
			if events, err := dec.decode(relationMessage(testRelationID, "public", "users", columns...), ""); err != nil {
				t.Fatalf("decode relation: %v", err)
			} else if len(events) != 0 {
				t.Fatalf("relation emitted %d event(s), want 0", len(events))
			}

			events, err := dec.decode(tc.message, tc.lsn)
			if err != nil {
				t.Fatalf("decode DML: %v", err)
			}
			if len(events) != 1 {
				t.Fatalf("decode emitted %d event(s), want 1", len(events))
			}
			if !reflect.DeepEqual(events[0], tc.want) {
				t.Fatalf("event mismatch\n got: %#v\nwant: %#v", events[0], tc.want)
			}
		})
	}
}

func TestDecodeTextValuePreservesNumericPrecision(t *testing.T) {
	if got := decodeTextValue(1700, "9007199254740993"); got != "9007199254740993" {
		t.Fatalf("decodeTextValue(numeric) = %#v, want exact decimal text", got)
	}
}

func TestPGOutputDecoderEncodesNonFiniteFloatsAsJSONStrings(t *testing.T) {
	cases := []struct {
		name   string
		typeID uint32
		raw    string
	}{
		{name: "float4 NaN", typeID: 700, raw: "NaN"},
		{name: "float4 infinity", typeID: 700, raw: "Infinity"},
		{name: "float4 negative infinity", typeID: 700, raw: "-Infinity"},
		{name: "float8 NaN", typeID: 701, raw: "NaN"},
		{name: "float8 infinity", typeID: 701, raw: "Infinity"},
		{name: "float8 negative infinity", typeID: 701, raw: "-Infinity"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			dec := newPGOutputDecoder("public", "measurements")
			if _, err := dec.decode(relationMessage(testRelationID, "public", "measurements", testColumn{name: "value", typeID: tc.typeID}), ""); err != nil {
				t.Fatalf("decode relation: %v", err)
			}
			events, err := dec.decode(insertMessage(testRelationID, textField(tc.raw)), "")
			if err != nil {
				t.Fatalf("decode insert: %v", err)
			}
			if len(events) != 1 {
				t.Fatalf("decode emitted %d event(s), want 1", len(events))
			}
			value, ok := events[0].Record["value"].(string)
			if !ok || value != tc.raw {
				t.Fatalf("decoded value = %#v, want string %q", events[0].Record["value"], tc.raw)
			}
			payload, err := json.Marshal(events[0])
			if err != nil {
				t.Fatalf("marshal CDC event: %v", err)
			}
			var roundTrip connectors.CDCEvent
			if err := json.Unmarshal(payload, &roundTrip); err != nil {
				t.Fatalf("unmarshal CDC event: %v", err)
			}
			value, ok = roundTrip.Record["value"].(string)
			if !ok || value != tc.raw {
				t.Fatalf("round-trip value = %#v, want string %q", roundTrip.Record["value"], tc.raw)
			}
		})
	}
}

func TestPGOutputDecoderConsumesTypeMetadata(t *testing.T) {
	dec := newPGOutputDecoder("public", "users")
	events, err := dec.decode(typeMessage(23, "public", "custom_status"), "")
	if err != nil {
		t.Fatalf("decode type metadata: %v", err)
	}
	if len(events) != 0 {
		t.Fatalf("type metadata emitted %d event(s), want 0", len(events))
	}
}

func TestPGOutputDecoderConsumesOriginMetadata(t *testing.T) {
	dec := newPGOutputDecoder("public", "users")
	events, err := dec.decode(originMessage(42, "upstream"), "")
	if err != nil {
		t.Fatalf("decode origin metadata: %v", err)
	}
	if len(events) != 0 {
		t.Fatalf("origin metadata emitted %d event(s), want 0", len(events))
	}
}

func TestPGOutputDecoderRejectsSelectedReplicaIdentityUpdates(t *testing.T) {
	dec := newPGOutputDecoder("public", "users")
	columns := []testColumn{{name: "id", typeID: 23}, {name: "email", typeID: 25}}
	if _, err := dec.decode(relationMessage(testRelationID, "public", "users", columns...), ""); err != nil {
		t.Fatalf("decode relation: %v", err)
	}

	for _, tc := range []struct {
		name    string
		message []byte
		want    error
	}{
		{
			name:    "changed replica identity key",
			message: updateMessage(testRelationID, []tupleField{textField("7")}, []tupleField{textField("8"), textField("grace@example.invalid")}),
			want:    errCDCReplicaIdentityUpdate,
		},
		{
			name:    "replica identity full",
			message: updateMessageWithOldTuple(testRelationID, 'O', []tupleField{textField("7"), textField("ada@example.invalid")}, []tupleField{textField("7"), textField("grace@example.invalid")}),
			want:    errCDCReplicaIdentityFullUpdate,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := dec.decode(tc.message, ""); !errors.Is(err, tc.want) {
				t.Fatalf("decode update = %v, want %v", err, tc.want)
			}
		})
	}
}

func TestPGOutputDecoderSkipsUnrequestedReplicaIdentityUpdates(t *testing.T) {
	dec := newPGOutputDecoder("public", "users")
	columns := []testColumn{{name: "id", typeID: 23}, {name: "email", typeID: 25}}
	if _, err := dec.decode(relationMessage(testRelationID+1, "public", "orders", columns...), ""); err != nil {
		t.Fatalf("decode relation: %v", err)
	}
	events, err := dec.decode(updateMessage(testRelationID+1, []tupleField{textField("7")}, []tupleField{textField("8"), textField("grace@example.invalid")}), "")
	if err != nil {
		t.Fatalf("decode unrelated update: %v", err)
	}
	if len(events) != 0 {
		t.Fatalf("unrelated update emitted %d event(s), want 0", len(events))
	}
}

func TestPGOutputDecoderErrors(t *testing.T) {
	dec := newPGOutputDecoder("public", "users")
	cases := []struct {
		name    string
		message []byte
	}{
		{name: "unknown relation", message: insertMessage(99, textField("1"))},
		{name: "unsupported message", message: []byte{'X'}},
		{name: "truncated relation", message: []byte{'R', 0, 0}},
		{name: "unsupported tuple kind", message: append(insertPrefix(testRelationID), tupleData(tupleField{kind: 'b', value: "raw"})...)},
	}

	if _, err := dec.decode(relationMessage(testRelationID, "public", "users", testColumn{name: "id", typeID: 23}), ""); err != nil {
		t.Fatalf("decode relation: %v", err)
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := dec.decode(tc.message, ""); err == nil {
				t.Fatal("decode error = nil, want error")
			}
		})
	}
}

func TestPGOutputDecoderFiltersUnrequestedRelations(t *testing.T) {
	dec := newPGOutputDecoder("public", "users")
	columns := []testColumn{{name: "id", typeID: 23}}
	if _, err := dec.decode(relationMessage(testRelationID, "public", "users", columns...), ""); err != nil {
		t.Fatalf("decode target relation: %v", err)
	}
	if _, err := dec.decode(relationMessage(testRelationID+1, "public", "orders", columns...), ""); err != nil {
		t.Fatalf("decode unrelated relation: %v", err)
	}

	events, err := dec.decode(insertMessage(testRelationID+1, textField("7")), "")
	if err != nil {
		t.Fatalf("decode unrelated insert: %v", err)
	}
	if len(events) != 0 {
		t.Fatalf("unrelated insert emitted %d event(s), want 0", len(events))
	}

	events, err = dec.decode(insertMessage(testRelationID, textField("8")), "")
	if err != nil {
		t.Fatalf("decode target insert: %v", err)
	}
	if len(events) != 1 || events[0].Record["id"] != 8 {
		t.Fatalf("target insert events = %#v, want selected relation event", events)
	}
}

func relationMessage(id uint32, schema, table string, columns ...testColumn) []byte {
	var b []byte
	b = append(b, 'R')
	b = appendUint32(b, id)
	b = appendCString(b, schema)
	b = appendCString(b, table)
	b = append(b, 'd') // replica identity
	b = appendUint16(b, uint16(len(columns)))
	for i, col := range columns {
		flags := byte(0)
		if i == 0 {
			flags = 1
		}
		b = append(b, flags)
		b = appendCString(b, col.name)
		b = appendUint32(b, col.typeID)
		b = appendUint32(b, 0xffffffff) // typmod -1
	}
	return b
}

func typeMessage(id uint32, schema, name string) []byte {
	var b []byte
	b = append(b, 'Y')
	b = appendUint32(b, id)
	b = appendCString(b, schema)
	b = appendCString(b, name)
	return b
}

func originMessage(lsn uint64, name string) []byte {
	var b []byte
	b = append(b, 'O')
	b = appendUint64(b, lsn)
	b = appendCString(b, name)
	return b
}

func insertMessage(relID uint32, fields ...tupleField) []byte {
	b := insertPrefix(relID)
	return append(b, tupleData(fields...)...)
}

func insertPrefix(relID uint32) []byte {
	var b []byte
	b = append(b, 'I')
	b = appendUint32(b, relID)
	b = append(b, 'N')
	return b
}

func updateMessage(relID uint32, oldKey, newTuple []tupleField) []byte {
	if oldKey == nil {
		return updateMessageWithOldTuple(relID, 0, nil, newTuple)
	}
	return updateMessageWithOldTuple(relID, 'K', oldKey, newTuple)
}

func updateMessageWithOldTuple(relID uint32, tupleKind byte, oldTuple, newTuple []tupleField) []byte {
	var b []byte
	b = append(b, 'U')
	b = appendUint32(b, relID)
	if tupleKind != 0 {
		b = append(b, tupleKind)
		b = append(b, tupleData(oldTuple...)...)
	}
	b = append(b, 'N')
	b = append(b, tupleData(newTuple...)...)
	return b
}

func deleteMessage(relID uint32, tupleKind byte, fields ...tupleField) []byte {
	var b []byte
	b = append(b, 'D')
	b = appendUint32(b, relID)
	b = append(b, tupleKind)
	b = append(b, tupleData(fields...)...)
	return b
}

func tupleData(fields ...tupleField) []byte {
	var b []byte
	b = appendUint16(b, uint16(len(fields)))
	for _, field := range fields {
		b = append(b, field.kind)
		if field.kind == 't' || field.kind == 'b' {
			b = appendUint32(b, uint32(len(field.value)))
			b = append(b, field.value...)
		}
	}
	return b
}

func textField(value string) tupleField {
	return tupleField{kind: 't', value: value}
}

func nullField() tupleField {
	return tupleField{kind: 'n'}
}

func unchangedField() tupleField {
	return tupleField{kind: 'u'}
}

func appendCString(b []byte, s string) []byte {
	b = append(b, s...)
	return append(b, 0)
}

func appendUint16(b []byte, v uint16) []byte {
	var tmp [2]byte
	binary.BigEndian.PutUint16(tmp[:], v)
	return append(b, tmp[:]...)
}

func appendUint32(b []byte, v uint32) []byte {
	var tmp [4]byte
	binary.BigEndian.PutUint32(tmp[:], v)
	return append(b, tmp[:]...)
}

func appendUint64(b []byte, v uint64) []byte {
	var tmp [8]byte
	binary.BigEndian.PutUint64(tmp[:], v)
	return append(b, tmp[:]...)
}

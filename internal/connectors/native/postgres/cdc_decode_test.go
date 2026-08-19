package postgres

import (
	"bytes"
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
					"score":  98.5,
				},
				State: connectors.Record{"lsn": "0/16B6C50"},
			},
		},
		{
			name:    "update with old key",
			message: updateMessage(testRelationID, []tupleField{textField("7")}, []tupleField{textField("7"), textField("grace@example.invalid"), textField("f"), textField("99.25")}),
			want: connectors.CDCEvent{
				Operation: "update",
				Record: connectors.Record{
					"id":     7,
					"email":  "grace@example.invalid",
					"active": false,
					"score":  99.25,
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
					"score": 12.75,
				},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			dec := newPGOutputDecoder()
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

func TestPGOutputDecoderErrors(t *testing.T) {
	dec := newPGOutputDecoder()
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

func TestPGOutputDecoderFiltersUnselectedRelation(t *testing.T) {
	const otherRelationID uint32 = 43
	decoder := newPGOutputDecoderForRelation("public.users")
	if _, err := decoder.decode(relationMessage(testRelationID, "public", "users", testColumn{name: "id", typeID: 23}), ""); err != nil {
		t.Fatalf("decode selected relation: %v", err)
	}
	if _, err := decoder.decode(relationMessage(otherRelationID, "public", "noise", testColumn{name: "id", typeID: 23}), ""); err != nil {
		t.Fatalf("decode unselected relation: %v", err)
	}

	events, err := decoder.decode(insertMessage(otherRelationID, textField("99")), "0/20")
	if err != nil {
		t.Fatalf("decode unselected DML: %v", err)
	}
	if len(events) != 0 {
		t.Fatalf("unselected relation emitted %d event(s), want 0", len(events))
	}

	events, err = decoder.decode(insertMessage(testRelationID, textField("7")), "0/21")
	if err != nil {
		t.Fatalf("decode selected DML: %v", err)
	}
	if len(events) != 1 || events[0].Record["id"] != 7 {
		t.Fatalf("selected relation events = %#v, want only the selected record", events)
	}
}

func TestPGOutputDecoderTruncateMapsOnlyTheSelectedRelation(t *testing.T) {
	const otherRelationID uint32 = 43
	decoder := newPGOutputDecoderForRelation("public.users")
	if _, err := decoder.decode(relationMessage(testRelationID, "public", "users", testColumn{name: "id", typeID: 23}), ""); err != nil {
		t.Fatalf("decode selected relation: %v", err)
	}
	if _, err := decoder.decode(relationMessage(otherRelationID, "public", "noise", testColumn{name: "id", typeID: 23}), ""); err != nil {
		t.Fatalf("decode unselected relation: %v", err)
	}

	events, err := decoder.truncate([]uint32{otherRelationID, testRelationID}, "0/30")
	if err != nil {
		t.Fatalf("decode truncate: %v", err)
	}
	if len(events) != 1 || events[0].Operation != "truncate" || len(events[0].Record) != 0 || events[0].State["lsn"] != "0/30" {
		t.Fatalf("truncate events = %#v, want one selected empty-record truncate", events)
	}
}

func TestPGOutputDecoderRoundTripsNonASCIITextExactly(t *testing.T) {
	const want = "Málaga 東京"
	decoder := newPGOutputDecoderForRelation("public.users")
	if _, err := decoder.decode(relationMessage(testRelationID, "public", "users", testColumn{name: "value", typeID: 25}), ""); err != nil {
		t.Fatalf("decode relation: %v", err)
	}
	events, err := decoder.decode(insertMessage(testRelationID, textField(want)), "0/20")
	if err != nil {
		t.Fatalf("decode insert: %v", err)
	}
	if len(events) != 1 {
		t.Fatalf("decode emitted %d event(s), want 1", len(events))
	}
	got, ok := events[0].Record["value"].(string)
	if !ok || !bytes.Equal([]byte(got), []byte(want)) {
		t.Fatalf("decoded value = %q, want byte-exact %q", got, want)
	}
	payload, err := json.Marshal(events[0])
	if err != nil {
		t.Fatalf("marshal CDC event: %v", err)
	}
	var roundTrip connectors.CDCEvent
	if err := json.Unmarshal(payload, &roundTrip); err != nil {
		t.Fatalf("unmarshal CDC event: %v", err)
	}
	got, ok = roundTrip.Record["value"].(string)
	if !ok || !bytes.Equal([]byte(got), []byte(want)) {
		t.Fatalf("round-trip value = %q, want byte-exact %q", got, want)
	}
}

func TestPGOutputDecoderRejectsInvalidUTF8BeforeCheckpointBoundary(t *testing.T) {
	t.Run("tuple text", func(t *testing.T) {
		decoder := newPGOutputDecoderForRelation("public.users")
		if _, err := decoder.decode(relationMessage(testRelationID, "public", "users", testColumn{name: "value", typeID: 25}), ""); err != nil {
			t.Fatalf("decode relation: %v", err)
		}
		_, err := decoder.decode(insertMessage(testRelationID, textField(string([]byte{0xff}))), "0/20")
		if !errors.Is(err, errCDCInvalidUTF8) {
			t.Fatalf("decode invalid tuple = %v, want invalid UTF-8 rejection", err)
		}
	})

	for _, tc := range []struct {
		name    string
		message []byte
	}{
		{name: "origin metadata", message: originMessage(42, string([]byte{0xff}))},
		{name: "type metadata", message: typeMessage(23, string([]byte{0xff}), "custom_status")},
	} {
		t.Run(tc.name, func(t *testing.T) {
			decoder := newPGOutputDecoder()
			if _, err := decoder.decode(tc.message, ""); !errors.Is(err, errCDCInvalidUTF8) {
				t.Fatalf("decode invalid metadata = %v, want invalid UTF-8 rejection", err)
			}
		})
	}
}

func TestDecodeTextValueMakesNonFiniteFloatsJSONSafe(t *testing.T) {
	for _, tc := range []struct {
		typeID uint32
		raw    string
	}{
		{typeID: 700, raw: "NaN"},
		{typeID: 701, raw: "Infinity"},
		{typeID: 1700, raw: "-Infinity"},
	} {
		got := decodeTextValue(tc.typeID, tc.raw)
		if got != tc.raw {
			t.Fatalf("decodeTextValue(%d, %q) = %#v, want JSON-safe string", tc.typeID, tc.raw, got)
		}
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
	var b []byte
	b = append(b, 'U')
	b = appendUint32(b, relID)
	if oldKey != nil {
		b = append(b, 'K')
		b = append(b, tupleData(oldKey...)...)
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

func originMessage(originLSN uint64, name string) []byte {
	b := []byte{'O'}
	b = appendUint64(b, originLSN)
	return appendCString(b, name)
}

func typeMessage(typeID uint32, namespace, name string) []byte {
	b := []byte{'Y'}
	b = appendUint32(b, typeID)
	b = appendCString(b, namespace)
	return appendCString(b, name)
}

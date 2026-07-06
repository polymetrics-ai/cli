package postgres

import (
	"encoding/binary"
	"fmt"
	"io"
	"math"
	"strconv"

	"polymetrics.ai/internal/connectors"
)

type pgoutputDecoder struct {
	relations map[uint32]pgoutputRelation
}

type pgoutputRelation struct {
	id      uint32
	schema  string
	table   string
	columns []pgoutputColumn
}

type pgoutputColumn struct {
	name   string
	typeID uint32
}

func newPGOutputDecoder() *pgoutputDecoder {
	return &pgoutputDecoder{relations: map[uint32]pgoutputRelation{}}
}

func (d *pgoutputDecoder) decode(message []byte, lsn string) ([]connectors.CDCEvent, error) {
	if len(message) == 0 {
		return nil, io.ErrUnexpectedEOF
	}
	r := pgoutputReader{buf: message[1:]}
	switch message[0] {
	case 'B', 'C':
		return nil, nil
	case 'R':
		rel, err := r.relation()
		if err != nil {
			return nil, err
		}
		d.relations[rel.id] = rel
		return nil, r.done()
	case 'I':
		ev, err := d.decodeInsert(&r, lsn)
		if err != nil {
			return nil, err
		}
		return []connectors.CDCEvent{ev}, r.done()
	case 'U':
		ev, err := d.decodeUpdate(&r, lsn)
		if err != nil {
			return nil, err
		}
		return []connectors.CDCEvent{ev}, r.done()
	case 'D':
		ev, err := d.decodeDelete(&r, lsn)
		if err != nil {
			return nil, err
		}
		return []connectors.CDCEvent{ev}, r.done()
	default:
		return nil, fmt.Errorf("pgoutput: unsupported message type %q", message[0])
	}
}

func (d *pgoutputDecoder) decodeInsert(r *pgoutputReader, lsn string) (connectors.CDCEvent, error) {
	rel, err := d.relation(r.uint32())
	if err != nil {
		return connectors.CDCEvent{}, err
	}
	tag := r.byte()
	if tag != 'N' {
		return connectors.CDCEvent{}, fmt.Errorf("pgoutput insert: expected new tuple tag 'N', got %q", tag)
	}
	rec, err := r.tuple(rel)
	if err != nil {
		return connectors.CDCEvent{}, fmt.Errorf("pgoutput insert: %w", err)
	}
	return cdcEvent("insert", rec, lsn), nil
}

func (d *pgoutputDecoder) decodeUpdate(r *pgoutputReader, lsn string) (connectors.CDCEvent, error) {
	rel, err := d.relation(r.uint32())
	if err != nil {
		return connectors.CDCEvent{}, err
	}
	tag := r.byte()
	if tag == 'K' || tag == 'O' {
		if _, err := r.tuple(rel); err != nil {
			return connectors.CDCEvent{}, fmt.Errorf("pgoutput update: old tuple: %w", err)
		}
		tag = r.byte()
	}
	if tag != 'N' {
		return connectors.CDCEvent{}, fmt.Errorf("pgoutput update: expected new tuple tag 'N', got %q", tag)
	}
	rec, err := r.tuple(rel)
	if err != nil {
		return connectors.CDCEvent{}, fmt.Errorf("pgoutput update: %w", err)
	}
	return cdcEvent("update", rec, lsn), nil
}

func (d *pgoutputDecoder) decodeDelete(r *pgoutputReader, lsn string) (connectors.CDCEvent, error) {
	rel, err := d.relation(r.uint32())
	if err != nil {
		return connectors.CDCEvent{}, err
	}
	tag := r.byte()
	if tag != 'K' && tag != 'O' {
		return connectors.CDCEvent{}, fmt.Errorf("pgoutput delete: expected old tuple tag 'K' or 'O', got %q", tag)
	}
	rec, err := r.tuple(rel)
	if err != nil {
		return connectors.CDCEvent{}, fmt.Errorf("pgoutput delete: %w", err)
	}
	return cdcEvent("delete", rec, lsn), nil
}

func (d *pgoutputDecoder) relation(id uint32) (pgoutputRelation, error) {
	rel, ok := d.relations[id]
	if !ok {
		return pgoutputRelation{}, fmt.Errorf("pgoutput: relation %d not known; relation message must arrive before DML", id)
	}
	return rel, nil
}

func cdcEvent(op string, rec connectors.Record, lsn string) connectors.CDCEvent {
	ev := connectors.CDCEvent{Operation: op, Record: rec}
	if lsn != "" {
		ev.State = connectors.Record{"lsn": lsn}
	}
	return ev
}

type pgoutputReader struct {
	buf []byte
	err error
}

func (r *pgoutputReader) done() error {
	if r.err != nil {
		return r.err
	}
	if len(r.buf) != 0 {
		return fmt.Errorf("pgoutput: %d trailing byte(s)", len(r.buf))
	}
	return nil
}

func (r *pgoutputReader) byte() byte {
	if r.err != nil {
		return 0
	}
	if len(r.buf) < 1 {
		r.err = io.ErrUnexpectedEOF
		return 0
	}
	b := r.buf[0]
	r.buf = r.buf[1:]
	return b
}

func (r *pgoutputReader) uint16() uint16 {
	if r.err != nil {
		return 0
	}
	if len(r.buf) < 2 {
		r.err = io.ErrUnexpectedEOF
		return 0
	}
	v := binary.BigEndian.Uint16(r.buf[:2])
	r.buf = r.buf[2:]
	return v
}

func (r *pgoutputReader) uint32() uint32 {
	if r.err != nil {
		return 0
	}
	if len(r.buf) < 4 {
		r.err = io.ErrUnexpectedEOF
		return 0
	}
	v := binary.BigEndian.Uint32(r.buf[:4])
	r.buf = r.buf[4:]
	return v
}

func (r *pgoutputReader) int32() int32 {
	return int32(r.uint32())
}

func (r *pgoutputReader) cstring() string {
	if r.err != nil {
		return ""
	}
	for i, b := range r.buf {
		if b == 0 {
			s := string(r.buf[:i])
			r.buf = r.buf[i+1:]
			return s
		}
	}
	r.err = io.ErrUnexpectedEOF
	return ""
}

func (r *pgoutputReader) bytes(n int) []byte {
	if r.err != nil {
		return nil
	}
	if n < 0 || len(r.buf) < n {
		r.err = io.ErrUnexpectedEOF
		return nil
	}
	out := r.buf[:n]
	r.buf = r.buf[n:]
	return out
}

func (r *pgoutputReader) relation() (pgoutputRelation, error) {
	rel := pgoutputRelation{
		id:     r.uint32(),
		schema: r.cstring(),
		table:  r.cstring(),
	}
	_ = r.byte() // replica identity
	columnCount := int(r.uint16())
	rel.columns = make([]pgoutputColumn, 0, columnCount)
	for i := 0; i < columnCount; i++ {
		_ = r.byte() // column flags
		name := r.cstring()
		typeID := r.uint32()
		_ = r.int32() // type modifier
		rel.columns = append(rel.columns, pgoutputColumn{name: name, typeID: typeID})
	}
	if err := r.done(); err != nil {
		return pgoutputRelation{}, fmt.Errorf("pgoutput relation: %w", err)
	}
	return rel, nil
}

func (r *pgoutputReader) tuple(rel pgoutputRelation) (connectors.Record, error) {
	columnCount := int(r.uint16())
	if r.err != nil {
		return nil, r.err
	}
	if columnCount > len(rel.columns) {
		return nil, fmt.Errorf("tuple has %d column(s), relation %s.%s has %d", columnCount, rel.schema, rel.table, len(rel.columns))
	}
	rec := connectors.Record{}
	for i := 0; i < columnCount; i++ {
		col := rel.columns[i]
		kind := r.byte()
		switch kind {
		case 'n':
			rec[col.name] = nil
		case 'u':
			// Unchanged TOAST value: the logical message does not carry the
			// value, so omit it rather than inventing nil or an empty string.
		case 't':
			raw := r.bytes(int(r.int32()))
			if r.err != nil {
				return nil, r.err
			}
			rec[col.name] = decodeTextValue(col.typeID, string(raw))
		default:
			return nil, fmt.Errorf("column %q has unsupported tuple kind %q", col.name, kind)
		}
	}
	if r.err != nil {
		return nil, r.err
	}
	return rec, nil
}

func decodeTextValue(typeID uint32, raw string) any {
	switch typeID {
	case 16: // bool
		return raw == "t" || raw == "true"
	case 20, 21, 23: // int8, int2, int4
		if v, err := strconv.ParseInt(raw, 10, 64); err == nil {
			if v >= math.MinInt && v <= math.MaxInt {
				return int(v)
			}
			return v
		}
	case 700, 701, 1700: // float4, float8, numeric
		if v, err := strconv.ParseFloat(raw, 64); err == nil {
			return v
		}
	}
	return raw
}

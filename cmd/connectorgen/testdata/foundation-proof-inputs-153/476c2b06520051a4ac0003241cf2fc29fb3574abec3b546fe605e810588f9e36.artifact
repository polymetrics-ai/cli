package database

import (
	"errors"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/synccontract"
)

// OrderedRecord pairs a source record with its source-owned ordering tuple.
// The tuple stays opaque: native targets may compare its bytes but must not
// parse, normalize, or replace it with target-observation time.
type OrderedRecord struct {
	Record   connectors.Record
	Position synccontract.CheckpointPosition
}

var (
	// ErrTombstoneEnvelopeInvalid refuses malformed or implicit delete input.
	// A caller cannot represent deletion by omitting an ordinary record.
	ErrTombstoneEnvelopeInvalid = errors.New("database tombstone envelope is invalid")
	// ErrDatabaseWriteInputInvalid refuses an unbounded or plan-mismatched
	// record/tombstone input before a native session is opened.
	ErrDatabaseWriteInputInvalid = errors.New("database write input is invalid")
)

// TombstoneEnvelope is the only write-input path for target deletes. It wraps
// validated synccontract tombstones so physical absence from Records is not a
// deletion signal and cannot be mistaken for one by a session implementation.
type TombstoneEnvelope struct {
	tombstones []synccontract.Tombstone
}

// NewTombstoneEnvelope validates and defensively copies explicit delete
// events. An empty envelope is valid and represents no requested deletes.
func NewTombstoneEnvelope(tombstones []synccontract.Tombstone) (TombstoneEnvelope, error) {
	envelope := TombstoneEnvelope{tombstones: cloneDatabaseWriteTombstones(tombstones)}
	if err := envelope.validate(); err != nil {
		return TombstoneEnvelope{}, ErrTombstoneEnvelopeInvalid
	}
	return envelope, nil
}

func (e TombstoneEnvelope) validate() error {
	seen := make(map[string]struct{}, len(e.tombstones))
	for _, tombstone := range e.tombstones {
		if tombstone.Operation != synccontract.OperationDelete || tombstone.Validate() != nil {
			return ErrTombstoneEnvelopeInvalid
		}
		key := string(tombstone.EventID)
		if _, exists := seen[key]; exists {
			return ErrTombstoneEnvelopeInvalid
		}
		seen[key] = struct{}{}
	}
	return nil
}

// Count returns the exact number of explicit tombstones.
func (e TombstoneEnvelope) Count() int { return len(e.tombstones) }

// Tombstones returns independent copies of the explicit delete events.
func (e TombstoneEnvelope) Tombstones() []synccontract.Tombstone {
	return cloneDatabaseWriteTombstones(e.tombstones)
}

// DatabaseWriteInput combines ordinary records and explicit tombstones for a
// single sealed write plan. It contains no implied delete set.
type DatabaseWriteInput struct {
	records   []OrderedRecord
	tombstone TombstoneEnvelope
}

// NewDatabaseWriteInput constructs one bounded-input candidate. Exact record
// and tombstone counts are checked against a plan immediately before session
// opening, where approval and driver identity are also checked.
func NewDatabaseWriteInput(records []connectors.Record, tombstones TombstoneEnvelope) (DatabaseWriteInput, error) {
	ordered := make([]OrderedRecord, len(records))
	for index, record := range records {
		ordered[index].Record = record
	}
	return NewOrderedDatabaseWriteInput(ordered, tombstones)
}

// NewOrderedDatabaseWriteInput constructs a candidate that retains the
// source ordering tuple for a plan whose conditional order fence is enabled.
// The legacy constructor remains available for plans that do not require a
// fence, so a sealed workset cannot silently acquire synthetic ordering.
func NewOrderedDatabaseWriteInput(records []OrderedRecord, tombstones TombstoneEnvelope) (DatabaseWriteInput, error) {
	input := DatabaseWriteInput{
		records:   cloneDatabaseWriteOrderedRecords(records),
		tombstone: TombstoneEnvelope{tombstones: tombstones.Tombstones()},
	}
	if err := input.validate(); err != nil {
		return DatabaseWriteInput{}, ErrDatabaseWriteInputInvalid
	}
	return input, nil
}

func (i DatabaseWriteInput) validate() error {
	if (len(i.records) == 0 && i.tombstone.Count() == 0) || i.tombstone.validate() != nil {
		return ErrDatabaseWriteInputInvalid
	}
	for _, record := range i.records {
		if record.Record == nil {
			return ErrDatabaseWriteInputInvalid
		}
	}
	return nil
}

func (i DatabaseWriteInput) validateFor(plan DatabaseWritePlan) error {
	if i.validate() != nil || len(i.records) != plan.RecordCount() || i.tombstone.Count() != plan.TombstoneCount() {
		return ErrDatabaseWriteInputInvalid
	}
	if plan.ConditionalOrderFence() {
		for _, record := range i.records {
			if len(record.Position.Primary) == 0 || len(record.Position.TieBreaker) == 0 {
				return ErrDatabaseWriteInputInvalid
			}
		}
	}
	return nil
}

// Records returns independent top-level record projections.
func (i DatabaseWriteInput) Records() []connectors.Record {
	records := make([]connectors.Record, len(i.records))
	for index := range i.records {
		records[index] = cloneDatabaseWriteRecord(i.records[index].Record)
	}
	return records
}

// OrderedRecords returns independent record and source-position projections.
func (i DatabaseWriteInput) OrderedRecords() []OrderedRecord {
	return cloneDatabaseWriteOrderedRecords(i.records)
}

// Tombstones returns the typed explicit delete envelope.
func (i DatabaseWriteInput) Tombstones() TombstoneEnvelope {
	return TombstoneEnvelope{tombstones: i.tombstone.Tombstones()}
}

func (i DatabaseWriteInput) batches(limit int) ([]WriteBatch, error) {
	if i.validate() != nil || limit <= 0 {
		return nil, ErrDatabaseWriteInputInvalid
	}
	batches := make([]WriteBatch, 0, (len(i.records)+i.tombstone.Count()+limit-1)/limit)
	sequence := uint64(1)
	for start := 0; start < len(i.records); start += limit {
		end := start + limit
		if end > len(i.records) {
			end = len(i.records)
		}
		batch, err := newWriteBatch(sequence, i.records[start:end], nil)
		if err != nil {
			return nil, err
		}
		batches = append(batches, batch)
		sequence++
	}
	tombstones := i.tombstone.Tombstones()
	for start := 0; start < len(tombstones); start += limit {
		end := start + limit
		if end > len(tombstones) {
			end = len(tombstones)
		}
		batch, err := newWriteBatch(sequence, nil, tombstones[start:end])
		if err != nil {
			return nil, err
		}
		batches = append(batches, batch)
		sequence++
	}
	if len(batches) == 0 {
		return nil, ErrDatabaseWriteInputInvalid
	}
	return batches, nil
}

func cloneDatabaseWriteTombstones(tombstones []synccontract.Tombstone) []synccontract.Tombstone {
	clone := make([]synccontract.Tombstone, len(tombstones))
	for index := range tombstones {
		clone[index] = tombstones[index].Clone()
	}
	return clone
}

func cloneDatabaseWriteOrderedRecords(records []OrderedRecord) []OrderedRecord {
	clone := make([]OrderedRecord, len(records))
	for index, record := range records {
		clone[index] = OrderedRecord{
			Record:   cloneDatabaseWriteRecord(record.Record),
			Position: record.Position.Clone(),
		}
	}
	return clone
}

func cloneDatabaseWriteRecord(record connectors.Record) connectors.Record {
	clone := make(connectors.Record, len(record))
	for key, value := range record {
		clone[key] = value
	}
	return clone
}

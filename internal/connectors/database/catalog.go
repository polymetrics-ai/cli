package database

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"sort"
	"strconv"
	"strings"
)

// KeyKind is the closed catalog key vocabulary. Both kinds are unique and may
// make a read order stable when their full ordered column list is a suffix.
type KeyKind string

const (
	KeyPrimary KeyKind = "primary"
	KeyUnique  KeyKind = "unique"
)

// Column captures a relation-bound column's logical type and nullability.
// Generated/default/identity semantics are deliberately left to a later
// managed-target contract rather than inferred here.
type Column struct {
	Ref      ColumnRef
	Type     LogicalType
	Nullable bool
	Ordinal  int
	Native   *NativeType
}

func (c Column) clone() Column {
	clone := c
	clone.Type = c.Type.clone()
	if c.Native != nil {
		native := c.Native.clone()
		clone.Native = &native
	}
	return clone
}

func (c Column) validate(relation RelationRef) error {
	if !c.Ref.Relation.equal(relation) || c.Ordinal <= 0 {
		return errors.New("database catalog column is not bound to its relation")
	}
	if err := c.Ref.validate(); err != nil {
		return errors.New("database catalog column reference is invalid")
	}
	if err := c.Type.validate(0); err != nil {
		return errors.New("database catalog column logical type is invalid")
	}
	if c.Native != nil {
		if err := c.Native.validate(); err != nil {
			return errors.New("database catalog column native type is invalid")
		}
	}
	return nil
}

// Key is a relation-bound primary or unique key. It holds structured columns,
// never a comma-delimited identifier list.
type Key struct {
	Name    string
	Kind    KeyKind
	Columns []ColumnRef
}

func (k Key) clone() Key {
	clone := k
	clone.Columns = append([]ColumnRef(nil), k.Columns...)
	return clone
}

func (k Key) validate(relation RelationRef, columns []Column) error {
	if !validOpaqueName(k.Name) || (k.Kind != KeyPrimary && k.Kind != KeyUnique) || len(k.Columns) == 0 {
		return errors.New("database catalog key is invalid")
	}
	seen := make(map[string]struct{}, len(k.Columns))
	for _, ref := range k.Columns {
		if !ref.Relation.equal(relation) || !containsColumn(columns, ref) {
			return errors.New("database catalog key references an unknown column")
		}
		if _, exists := seen[ref.Name]; exists {
			return errors.New("database catalog key contains a duplicate column")
		}
		seen[ref.Name] = struct{}{}
	}
	return nil
}

// Relation is a normalized catalog relation and its columns/keys. It is an
// input/output projection; Catalog copies it rather than retaining callers'
// mutable slices.
type Relation struct {
	Ref            RelationRef
	NativeIdentity NativeRelationIdentity
	Columns        []Column
	Keys           []Key
}

func (r Relation) clone() Relation {
	clone := r
	clone.Columns = make([]Column, len(r.Columns))
	for i := range r.Columns {
		clone.Columns[i] = r.Columns[i].clone()
	}
	clone.Keys = make([]Key, len(r.Keys))
	for i := range r.Keys {
		clone.Keys[i] = r.Keys[i].clone()
	}
	return clone
}

func (r Relation) validate(catalog CatalogRef) error {
	if err := r.Ref.validate(); err != nil || r.Ref.Schema.Catalog.Name != catalog.Name {
		return errors.New("database catalog relation reference is invalid")
	}
	if err := r.NativeIdentity.validate(); err != nil {
		return err
	}
	if len(r.Columns) == 0 {
		return errors.New("database catalog relation requires columns")
	}
	ordinals := make(map[int]struct{}, len(r.Columns))
	columnNames := make(map[string]struct{}, len(r.Columns))
	for _, column := range r.Columns {
		if err := column.validate(r.Ref); err != nil {
			return err
		}
		if _, exists := ordinals[column.Ordinal]; exists {
			return errors.New("database catalog relation has duplicate column ordinals")
		}
		if _, exists := columnNames[column.Ref.Name]; exists {
			return errors.New("database catalog relation has duplicate column names")
		}
		ordinals[column.Ordinal] = struct{}{}
		columnNames[column.Ref.Name] = struct{}{}
	}
	keyNames := make(map[string]struct{}, len(r.Keys))
	for _, key := range r.Keys {
		if err := key.validate(r.Ref, r.Columns); err != nil {
			return err
		}
		if _, exists := keyNames[key.Name]; exists {
			return errors.New("database catalog relation has duplicate key names")
		}
		keyNames[key.Name] = struct{}{}
	}
	return nil
}

// SchemaFingerprint is a stable SHA-256 projection of the normalized catalog.
// It is a value rather than an exposed mutable byte slice.
type SchemaFingerprint [sha256.Size]byte

// IsZero reports whether no fingerprint was established.
func (f SchemaFingerprint) IsZero() bool { return f == SchemaFingerprint{} }

// Bytes returns a new byte slice every time.
func (f SchemaFingerprint) Bytes() []byte { return append([]byte(nil), f[:]...) }

// String returns the canonical hexadecimal fingerprint encoding.
func (f SchemaFingerprint) String() string { return hex.EncodeToString(f[:]) }

// Catalog owns immutable copies of its normalized relation projection.
type Catalog struct {
	ref         CatalogRef
	relations   []Relation
	fingerprint SchemaFingerprint
}

// NewCatalog validates and canonicalizes a discovered catalog. Unknown native
// types can remain explicitly opaque, but a malformed catalog never reaches a
// planner.
func NewCatalog(ref CatalogRef, relations []Relation) (Catalog, error) {
	if err := ref.validate(); err != nil {
		return Catalog{}, errors.New("database catalog reference is invalid")
	}
	if len(relations) == 0 {
		return Catalog{}, errors.New("database catalog requires at least one relation")
	}
	cloned := make([]Relation, len(relations))
	for i := range relations {
		cloned[i] = relations[i].clone()
		if err := cloned[i].validate(ref); err != nil {
			return Catalog{}, err
		}
	}
	sort.Slice(cloned, func(i, j int) bool { return relationKey(cloned[i].Ref) < relationKey(cloned[j].Ref) })
	for i := 1; i < len(cloned); i++ {
		if cloned[i-1].Ref.equal(cloned[i].Ref) {
			return Catalog{}, errors.New("database catalog has duplicate relations")
		}
	}
	fingerprint := fingerprintCatalog(ref, cloned)
	return Catalog{ref: ref, relations: cloned, fingerprint: fingerprint}, nil
}

// Ref returns the catalog identifier.
func (c Catalog) Ref() CatalogRef { return c.ref }

// Relations returns a defensive copy of the complete relation projection.
func (c Catalog) Relations() []Relation {
	relations := make([]Relation, len(c.relations))
	for i := range c.relations {
		relations[i] = c.relations[i].clone()
	}
	return relations
}

// Fingerprint returns the catalog's stable schema fingerprint.
func (c Catalog) Fingerprint() SchemaFingerprint { return c.fingerprint }

func (c Catalog) relation(ref RelationRef) (Relation, bool) {
	for _, relation := range c.relations {
		if relation.Ref.equal(ref) {
			return relation.clone(), true
		}
	}
	return Relation{}, false
}

func (c Catalog) validate() error {
	if err := c.ref.validate(); err != nil || len(c.relations) == 0 || c.fingerprint.IsZero() {
		return errors.New("database catalog is invalid")
	}
	for _, relation := range c.relations {
		if err := relation.validate(c.ref); err != nil {
			return err
		}
	}
	return nil
}

func containsColumn(columns []Column, ref ColumnRef) bool {
	for _, column := range columns {
		if column.Ref.equal(ref) {
			return true
		}
	}
	return false
}

func relationKey(ref RelationRef) string {
	return ref.Schema.Catalog.Name + "\x00" + ref.Schema.Name + "\x00" + ref.Name
}

func fingerprintCatalog(ref CatalogRef, relations []Relation) SchemaFingerprint {
	var builder strings.Builder
	builder.WriteString("catalog\x00")
	builder.WriteString(ref.Name)
	for _, relation := range relations {
		builder.WriteString("\x00relation\x00")
		builder.WriteString(relationKey(relation.Ref))
		builder.WriteString("\x00native\x00")
		builder.WriteString(relation.NativeIdentity.Kind)
		builder.WriteString("\x00")
		builder.WriteString(relation.NativeIdentity.Value)
		columns := append([]Column(nil), relation.Columns...)
		sort.Slice(columns, func(i, j int) bool { return columns[i].Ordinal < columns[j].Ordinal })
		for _, column := range columns {
			builder.WriteString("\x00column\x00")
			builder.WriteString(column.Ref.Name)
			builder.WriteString("\x00")
			builder.WriteString(strconv.Itoa(column.Ordinal))
			builder.WriteString("\x00")
			builder.WriteString(strconv.FormatBool(column.Nullable))
			builder.WriteString("\x00")
			builder.WriteString(canonicalLogicalType(column.Type))
			if column.Native != nil {
				builder.WriteString("\x00native_type\x00")
				builder.WriteString(column.Native.Name)
				for _, modifier := range column.Native.Modifiers {
					builder.WriteString("\x00")
					builder.WriteString(modifier)
				}
			}
		}
		keys := append([]Key(nil), relation.Keys...)
		sort.Slice(keys, func(i, j int) bool { return keys[i].Name < keys[j].Name })
		for _, key := range keys {
			builder.WriteString("\x00key\x00")
			builder.WriteString(key.Name)
			builder.WriteString("\x00")
			builder.WriteString(string(key.Kind))
			for _, column := range key.Columns {
				builder.WriteString("\x00")
				builder.WriteString(column.Name)
			}
		}
	}
	return sha256.Sum256([]byte(builder.String()))
}

func canonicalLogicalType(t LogicalType) string {
	var builder strings.Builder
	builder.WriteString(string(t.kind))
	builder.WriteString("/")
	builder.WriteString(strconv.Itoa(int(t.bits)))
	builder.WriteString("/")
	builder.WriteString(strconv.Itoa(int(t.precision)))
	builder.WriteString("/")
	builder.WriteString(strconv.Itoa(int(t.scale)))
	builder.WriteString("/")
	builder.WriteString(strconv.FormatUint(uint64(t.maxLength), 10))
	builder.WriteString("/")
	builder.WriteString(t.collation)
	builder.WriteString("/")
	builder.WriteString(strconv.FormatBool(t.withTimezone))
	builder.WriteString("/")
	builder.WriteString(t.opaqueEngine)
	builder.WriteString("/")
	builder.WriteString(t.opaqueName)
	for _, option := range t.opaqueOptions {
		builder.WriteString("/")
		builder.WriteString(option)
	}
	if t.element != nil {
		builder.WriteString("/[")
		builder.WriteString(canonicalLogicalType(*t.element))
		builder.WriteString("]")
	}
	return builder.String()
}

package connectors

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"path"
	"reflect"
	"sort"
	"strings"
	"unicode"
	"unicode/utf8"
)

const SourceVisibilityConnectorLimit = 32 << 20
const SourceVisibilityTotalLimit = 64 << 20

type SourceLane string

func SourceLanes() []SourceLane {
	return []SourceLane{"direct_read", "direct_write", "binary_download", "binary_upload", "etl", "reverse_etl", "sync_transport"}
}

type SourceOperationKey struct {
	Connector string `json:"connector"`
	Inventory string `json:"inventory"`
	ID        string `json:"id"`
}
type SourceCellSelection struct {
	Source SourceOperationKey `json:"source"`
	Lane   SourceLane         `json:"lane"`
}

// SourceVisibilityArtifact is immutable metadata, separate from execution identity.
// Selection decodes only this connector; ordinary commands never consult it.
type SourceVisibilityArtifact struct {
	SchemaVersion                 int
	Connector, Coverage, CohortID string
	Bytes                         int
	SHA256, KeySHA256, Payload    string
}
type SourceDocumentRef struct {
	DocumentID             string `json:"document_id"`
	Path                   string `json:"path"`
	RetainedFileSHA256     string `json:"retained_file_sha256"`
	Bytes                  int64  `json:"bytes"`
	UpstreamDeclaredSHA256 string `json:"upstream_declared_sha256,omitempty"`
	UpstreamDeclaredBytes  int64  `json:"upstream_declared_bytes,omitempty"`
	UpstreamBytesVerified  bool   `json:"upstream_bytes_verified,omitempty"`
}
type SourceFactCitation struct {
	DocumentID  string `json:"document_id"`
	Pointer     string `json:"pointer"`
	ValueSHA256 string `json:"value_sha256"`
	Section     string `json:"section,omitempty"`
	Part        string `json:"part,omitempty"`
}
type SourceReason struct {
	Code string `json:"code"`
	Text string `json:"text"`
}
type SourceCapabilityObservation struct {
	Kind         string   `json:"kind"`
	Name         string   `json:"name"`
	AtlasID      *string  `json:"atlas_id"`
	CandidateIDs []string `json:"candidate_ids"`
	GapID        string   `json:"gap_id,omitempty"`
	Provenance   string   `json:"provenance"`
}
type SourceFieldMapping struct {
	SourceCitation int     `json:"source_citation"`
	TargetKind     string  `json:"target_kind"`
	TargetPointer  *string `json:"target_pointer"`
}

type SourceArtifactRef struct {
	Role             string               `json:"role"`
	Kind             string               `json:"kind"`
	Connector        string               `json:"connector"`
	ID               string               `json:"id"`
	Lane             SourceLane           `json:"lane"`
	Artifact         string               `json:"artifact"`
	Pointer          string               `json:"pointer"`
	ArtifactSHA256   string               `json:"artifact_sha256"`
	CanonicalID      string               `json:"canonical_id"`
	CanonicalPointer string               `json:"canonical_pointer"`
	Generation       string               `json:"generation"`
	SchemaRole       string               `json:"schema_role"`
	SourceSchema     *int                 `json:"source_schema,omitempty"`
	FieldMappings    []SourceFieldMapping `json:"field_mappings,omitempty"`
}
type SourceCellObservation struct {
	Lane          SourceLane                  `json:"lane"`
	Applicability string                      `json:"applicability"`
	SourceState   string                      `json:"source_state"`
	RuleID        string                      `json:"rule_id"`
	SourceReason  SourceReason                `json:"source_reason"`
	LaneFacts     []int                       `json:"lane_facts"`
	Capability    SourceCapabilityObservation `json:"capability"`
	Intended      []SourceArtifactRef         `json:"intended"`
	Admitted      []SourceArtifactRef         `json:"admitted"`
	OwnerRefs     []string                    `json:"owner_refs"`
	GapRefs       []string                    `json:"gap_refs"`
}
type SourceOperationObservation struct {
	Source           SourceOperationKey      `json:"source"`
	Class            string                  `json:"class"`
	DocumentID       string                  `json:"document_id"`
	Pointer          string                  `json:"pointer"`
	SourceLocation   string                  `json:"source_location"`
	Observed         bool                    `json:"observed"`
	Protocol         string                  `json:"protocol"`
	Method           string                  `json:"method"`
	Path             string                  `json:"path"`
	IdentityCitation int                     `json:"identity_citation"`
	DisplayCitations []int                   `json:"display_citations"`
	Cells            []SourceCellObservation `json:"cells"`
}
type SourceVisibility struct {
	SchemaVersion  int                          `json:"schema_version"`
	Connector      string                       `json:"connector"`
	Coverage       string                       `json:"coverage"`
	CohortID       string                       `json:"cohort_id"`
	CohortSHA256   string                       `json:"cohort_sha256"`
	AtlasSHA256    string                       `json:"atlas_sha256"`
	KeySHA256      string                       `json:"key_sha256"`
	OperationCount int                          `json:"operation_count"`
	CellCount      int                          `json:"cell_count"`
	Documents      []SourceDocumentRef          `json:"documents"`
	Citations      []SourceFactCitation         `json:"citations"`
	Operations     []SourceOperationObservation `json:"operations"`
}
type SourceCellView struct {
	Selection         SourceCellSelection        `json:"selection"`
	Operation         SourceOperationObservation `json:"operation"`
	Cell              SourceCellObservation      `json:"cell"`
	Documents         []SourceDocumentRef        `json:"documents"`
	Citations         []SourceFactCitation       `json:"citations"`
	LaneEvidenceScope string                     `json:"lane_evidence_scope"`
	ExecutionChecked  bool                       `json:"execution_checked"`
}
type SourceCellPreflight struct {
	Kind             string         `json:"kind"`
	Cell             SourceCellView `json:"cell"`
	ExecutionChecked bool           `json:"execution_checked"`
}
type SourceIncompatibility struct {
	Scope string `json:"scope"`
	Axis  string `json:"axis"`
	Code  string `json:"code"`
}
type SourceSelectionError struct {
	Selection       SourceCellSelection    `json:"selection"`
	View            SourceCellView         `json:"view"`
	Kind            string                 `json:"kind"`
	Code            string                 `json:"code"`
	Incompatibility *SourceIncompatibility `json:"incompatibility,omitempty"`
	Cause           error                  `json:"-"`
}

func (e *SourceSelectionError) Error() string {
	message := fmt.Sprintf("source %s/%s/%s lane %s: %s; %s %s; capability=%s; reason=%s", e.Selection.Source.Connector, e.Selection.Source.Inventory, e.Selection.Source.ID, e.Selection.Lane, e.Code, e.View.Operation.Method, e.View.Operation.Path, e.View.Cell.Capability.Name, e.View.Cell.SourceReason.Code)
	if id := e.View.Cell.Capability.AtlasID; id != nil {
		message += "; atlas=" + *id
	}
	if gap := e.View.Cell.Capability.GapID; gap != "" {
		message += "; gap=" + gap
	}
	if i := e.View.Operation.IdentityCitation; i >= 0 && i < len(e.View.Citations) {
		ref := e.View.Citations[i]
		message += "; operation citation=" + ref.DocumentID + "#" + ref.Pointer
	}
	return message
}
func (e *SourceSelectionError) Unwrap() error { return e.Cause }

type SourceVisibilityDataError struct {
	Connector string
	Cause     error
}

func (e *SourceVisibilityDataError) Error() string {
	return fmt.Sprintf("source visibility for %q is invalid: %v", e.Connector, e.Cause)
}
func (e *SourceVisibilityDataError) Unwrap() error { return e.Cause }

type SourceSelectionInputError struct {
	Selection SourceCellSelection
	Code      string
}

func (e *SourceSelectionInputError) Error() string {
	return e.Code + ": exact connector, inventory, source ID and lane required"
}

func sourceHash(b []byte) string { h := sha256.Sum256(b); return hex.EncodeToString(h[:]) }
func sourceHashValid(s string) bool {
	b, e := hex.DecodeString(s)
	return e == nil && len(b) == 32 && strings.ToLower(s) == s
}
func sourceText(s string) bool {
	if s == "" || !utf8.ValidString(s) {
		return false
	}
	for _, r := range s {
		if unicode.IsControl(r) {
			return false
		}
	}
	return true
}
func sourceOptionalText(s string) bool { return s == "" || sourceText(s) }

func sourcePath(s string) bool {
	return sourceText(s) && !strings.Contains(s, "\\") && !strings.HasPrefix(s, "/") && path.Clean(s) == s && s != "." && s != ".." && !strings.HasPrefix(s, "../")
}
func sourcePointer(s string) bool {
	if s == "" {
		return true
	}
	if !strings.HasPrefix(s, "/") || !sourceText(s) {
		return false
	}
	for i := 0; i < len(s); i++ {
		if s[i] == '~' {
			i++
			if i == len(s) || (s[i] != '0' && s[i] != '1') {
				return false
			}
		}
	}
	return true
}
func sourceKeyLess(a, b SourceOperationKey) bool {
	if a.Connector != b.Connector {
		return a.Connector < b.Connector
	}
	if a.Inventory != b.Inventory {
		return a.Inventory < b.Inventory
	}
	return a.ID < b.ID
}
func SourceKeyDigest(keys []SourceOperationKey) string {
	keys = append([]SourceOperationKey(nil), keys...)
	sort.Slice(keys, func(i, j int) bool { return sourceKeyLess(keys[i], keys[j]) })
	b, _ := json.Marshal(keys)
	return sourceHash(b)
}

// validateSourceJSON rejects duplicate fields before typed decoding, with bounded
// depth/node work. DisallowUnknownFields then rejects executable additions.
func validateSourceJSON(d *json.Decoder, depth int, nodes *int) error {
	*nodes++
	if depth > 64 || *nodes > 2000000 {
		return errors.New("source metadata structure exceeds bounds")
	}
	token, err := d.Token()
	if err != nil {
		return err
	}
	delim, ok := token.(json.Delim)
	if !ok {
		return nil
	}
	switch delim {
	case '{':
		seen := map[string]bool{}
		for d.More() {
			t, e := d.Token()
			if e != nil {
				return e
			}
			k, ok := t.(string)
			if !ok || seen[k] {
				return errors.New("duplicate source metadata member")
			}
			seen[k] = true
			if e = validateSourceJSON(d, depth+1, nodes); e != nil {
				return e
			}
		}
	case '[':
		for d.More() {
			if e := validateSourceJSON(d, depth+1, nodes); e != nil {
				return e
			}
		}
	default:
		return errors.New("unexpected source metadata delimiter")
	}
	_, err = d.Token()
	return err
}

func DecodeSourceVisibility(ctx context.Context, artifact SourceVisibilityArtifact) (SourceVisibility, error) {
	var result SourceVisibility
	fail := func(e error) (SourceVisibility, error) {
		return SourceVisibility{}, &SourceVisibilityDataError{Connector: artifact.Connector, Cause: e}
	}
	if ctx == nil {
		return fail(errors.New("source context is required"))
	}
	if e := ctx.Err(); e != nil {
		return result, e
	}
	if artifact.SchemaVersion != 1 || !sourceText(artifact.Connector) {
		return fail(errors.New("source artifact header invalid"))
	}
	if artifact.Coverage == "not_in_cohort" {
		if artifact.Payload != "" || artifact.Bytes != 0 || artifact.SHA256 != "" || artifact.KeySHA256 != "" {
			return fail(errors.New("non-cohort payload is not empty"))
		}
		return SourceVisibility{SchemaVersion: 1, Connector: artifact.Connector, Coverage: "not_in_cohort", CohortID: artifact.CohortID, Documents: []SourceDocumentRef{}, Citations: []SourceFactCitation{}, Operations: []SourceOperationObservation{}}, nil
	}
	if artifact.Coverage != "in_cohort" || artifact.Bytes <= 0 || artifact.Bytes > SourceVisibilityConnectorLimit || len(artifact.Payload) != artifact.Bytes || sourceHash([]byte(artifact.Payload)) != artifact.SHA256 {
		return fail(errors.New("source payload coverage/size/digest invalid"))
	}
	d := json.NewDecoder(strings.NewReader(artifact.Payload))
	nodes := 0
	if e := validateSourceJSON(d, 0, &nodes); e != nil {
		return fail(e)
	}
	if _, e := d.Token(); e != io.EOF {
		return fail(errors.New("source payload trailing value"))
	}
	d = json.NewDecoder(strings.NewReader(artifact.Payload))
	d.DisallowUnknownFields()
	if e := d.Decode(&result); e != nil {
		return fail(e)
	}
	var fields any
	if err := json.Unmarshal([]byte(artifact.Payload), &fields); err != nil {
		return fail(err)
	}
	if err := validateSourceRequired(fields, reflect.TypeFor[SourceVisibility]()); err != nil {
		return fail(err)
	}
	if result.Connector != artifact.Connector || result.CohortID != artifact.CohortID || result.KeySHA256 != artifact.KeySHA256 {
		return fail(errors.New("source payload/header mismatch"))
	}
	if e := ValidateSourceVisibility(result); e != nil {
		return fail(e)
	}
	if e := ctx.Err(); e != nil {
		return SourceVisibility{}, e
	}
	return result, nil
}

func ValidateSourceVisibility(v SourceVisibility) error {
	bad := func(s string) error { return fmt.Errorf("source metadata: %s", s) }
	if v.SchemaVersion != 1 || v.Coverage != "in_cohort" || !sourceText(v.Connector) || !sourceText(v.CohortID) || !sourceHashValid(v.CohortSHA256) || !sourceHashValid(v.AtlasSHA256) || !sourceHashValid(v.KeySHA256) || v.OperationCount != len(v.Operations) || v.OperationCount == 0 || v.CellCount != 7*v.OperationCount || v.Documents == nil || v.Citations == nil {
		return bad("header/count invalid")
	}
	docs := map[string]bool{}
	for _, doc := range v.Documents {
		if !sourceText(doc.DocumentID) || docs[doc.DocumentID] || !sourcePath(doc.Path) || !sourceHashValid(doc.RetainedFileSHA256) || doc.Bytes <= 0 {
			return bad("document invalid")
		}
		docs[doc.DocumentID] = true
		if doc.UpstreamDeclaredSHA256 != "" && !sourceHashValid(doc.UpstreamDeclaredSHA256) {
			return bad("upstream document hash invalid")
		}
		if doc.UpstreamBytesVerified && (doc.UpstreamDeclaredSHA256 != doc.RetainedFileSHA256 || doc.UpstreamDeclaredBytes != doc.Bytes) {
			return bad("upstream document verification invalid")
		}
	}
	for _, c := range v.Citations {
		if !docs[c.DocumentID] || !sourcePointer(c.Pointer) || !sourceHashValid(c.ValueSHA256) || !sourceOptionalText(c.Section) || !sourceOptionalText(c.Part) {
			return bad("citation invalid")
		}
	}
	cite := func(i int) bool { return i >= 0 && i < len(v.Citations) }
	keys := make([]SourceOperationKey, 0, len(v.Operations))
	lanes := SourceLanes()
	for i, o := range v.Operations {
		if o.Source.Connector != v.Connector || !sourceText(o.Source.Inventory) || !sourceText(o.Source.ID) || !o.Observed || !docs[o.DocumentID] || !sourcePointer(o.Pointer) || !cite(o.IdentityCitation) || o.DisplayCitations == nil || len(o.Cells) != 7 || (o.Class != "primary" && o.Class != "supplement") {
			return bad("operation identity/citation invalid")
		}
		if !sourceOptionalText(o.Method) || !sourceOptionalText(o.Path) || !sourceOptionalText(o.Protocol) || !sourceOptionalText(o.SourceLocation) {
			return bad("operation display text invalid")
		}
		if i > 0 && !sourceKeyLess(v.Operations[i-1].Source, o.Source) {
			return bad("duplicate or unsorted source identity")
		}
		keys = append(keys, o.Source)
		for _, c := range o.DisplayCitations {
			if !cite(c) {
				return bad("display citation invalid")
			}
		}
		for j, c := range o.Cells {
			if c.Lane != lanes[j] || !sourceText(c.RuleID) || !sourceText(c.SourceReason.Text) || c.LaneFacts == nil || c.Intended == nil || c.Admitted == nil || c.OwnerRefs == nil || c.GapRefs == nil {
				return bad("lane shape invalid")
			}
			for _, ref := range c.LaneFacts {
				if !cite(ref) {
					return bad("lane citation invalid")
				}
			}
			if c.Applicability != "applicable" && c.Applicability != "not_applicable" && c.Applicability != "undetermined" {
				return bad("applicability invalid")
			}
			cap := c.Capability
			if !sourceText(cap.Name) || !sourceText(cap.Provenance) || cap.CandidateIDs == nil {
				return bad("capability observation invalid")
			}
			for _, id := range cap.CandidateIDs {
				if _, _, ok := sourceCapabilityMetadata(id); !ok {
					return bad("unknown capability candidate")
				}
			}
			if cap.AtlasID != nil {
				if _, _, ok := sourceCapabilityMetadata(*cap.AtlasID); !ok {
					return bad("unknown capability")
				}
			}
			switch c.SourceState {
			case "mapped_unproven":
				if (c.SourceReason.Code != "facts_unresolved" && c.SourceReason.Code != "materialization_unproven") || cap.Kind != "unresolved_mapping" || cap.AtlasID != nil || cap.GapID != "" || len(c.GapRefs) != 0 || c.Applicability == "not_applicable" {
					return bad("mapping observation invalid")
				}
			case "not_applicable":
				if c.SourceReason.Code != "source_exclusion" || c.Applicability != "not_applicable" || cap.Kind != "not_required" || cap.AtlasID != nil || cap.GapID != "" || len(c.GapRefs) != 0 || len(c.Admitted) != 0 || len(c.Intended) != 0 {
					return bad("source exclusion invalid")
				}
			case "missing_foundation":
				if c.SourceReason.Code != "existing_foundation_demand" || c.Applicability != "applicable" || cap.Kind != "known_gap" || cap.AtlasID == nil || len(c.GapRefs) != 1 || c.GapRefs[0] != cap.GapID || len(c.Admitted) != 0 {
					return bad("foundation demand invalid")
				}
				if !sourceGapValid(v, o, c) {
					return bad("source gap identity/provenance invalid")
				}
			case "implemented":
				if c.SourceReason.Code != "behavior_proven" || c.Applicability != "applicable" || len(c.Admitted) == 0 || cap.Kind != "known_fit" || cap.AtlasID == nil || cap.GapID != "" {
					return bad("admitted binding invalid")
				}
			default:
				return bad("unknown source state")
			}
			if c.SourceReason.Code != "facts_unresolved" && len(c.LaneFacts) == 0 {
				return bad("required lane evidence missing")
			}
			for _, group := range []struct {
				role string
				refs []SourceArtifactRef
			}{{"intended", c.Intended}, {"admitted", c.Admitted}} {
				for _, ref := range group.refs {
					switch ref.Kind {
					case "canonical_operation", "schema", "sync_transport", "operation", "write", "stream", "command":
					default:
						return bad("artifact reference kind invalid")
					}
					if ref.SchemaRole != "" && ref.SchemaRole != "request" && ref.SchemaRole != "response" && ref.SchemaRole != "record" {
						return bad("artifact schema role invalid")
					}
					if ref.SourceSchema != nil && !cite(*ref.SourceSchema) {
						return bad("artifact source schema citation invalid")
					}
					for _, mapping := range ref.FieldMappings {
						if !cite(mapping.SourceCitation) || (mapping.TargetKind != "schema" && mapping.TargetKind != "config" && mapping.TargetKind != "parameter") || (mapping.TargetPointer != nil && !sourcePointer(*mapping.TargetPointer)) {
							return bad("artifact field mapping invalid")
						}
					}
					if ref.Role != group.role || ref.Connector != v.Connector || ref.Lane != c.Lane || !sourceText(ref.ID) || !sourcePath(ref.Artifact) || !sourcePointer(ref.Pointer) || !sourceHashValid(ref.ArtifactSHA256) || !sourceText(ref.CanonicalID) || !sourcePointer(ref.CanonicalPointer) || !sourceText(ref.Generation) {
						return bad("artifact reference invalid")
					}
				}
			}
		}
	}
	if SourceKeyDigest(keys) != v.KeySHA256 {
		return bad("source key digest mismatch")
	}
	return nil
}

func InspectSourceCell(v SourceVisibility, s SourceCellSelection) (SourceCellView, error) {
	invalid := func(code string) (SourceCellView, error) {
		return SourceCellView{}, &SourceSelectionInputError{Selection: s, Code: code}
	}
	if !sourceText(s.Source.Connector) || !sourceText(s.Source.Inventory) || !sourceText(s.Source.ID) {
		return invalid("source_selection_invalid")
	}
	lane := -1
	for i, l := range SourceLanes() {
		if l == s.Lane {
			lane = i
		}
	}
	if lane < 0 {
		return invalid("source_selection_invalid")
	}
	for _, o := range v.Operations {
		if o.Source == s.Source {
			if len(o.Cells) != 7 {
				return SourceCellView{}, &SourceVisibilityDataError{Connector: s.Source.Connector, Cause: errors.New("selected source lane set invalid")}
			}
			c := o.Cells[lane]
			scope := "operation_identity_and_lane_basis"
			if len(c.LaneFacts) == 0 {
				scope = "operation_identity_only_lane_unresolved"
			}
			return SourceCellView{Selection: s, Operation: o, Cell: c, Documents: v.Documents, Citations: v.Citations, LaneEvidenceScope: scope}, nil
		}
	}
	return invalid("source_cell_not_found")
}
func preflightSourceCell(view SourceCellView) (SourceCellPreflight, error) {
	c := view.Cell
	e := &SourceSelectionError{Selection: view.Selection, View: view, Kind: c.SourceState}
	switch c.SourceState {
	case "mapped_unproven":
		e.Code = "source_mapping_unproven"
	case "missing_foundation":
		e.Code = "missing_foundation"
	case "not_applicable":
		e.Kind = "incompatible"
		e.Code = "source_lane_not_applicable"
		e.Incompatibility = &SourceIncompatibility{Scope: "source_lane", Axis: "lane", Code: e.Code}
	case "implemented":
		return SourceCellPreflight{Kind: "binding_observed", Cell: view}, nil
	default:
		return SourceCellPreflight{}, &SourceVisibilityDataError{Connector: view.Selection.Source.Connector, Cause: errors.New("unknown selected source state")}
	}
	return SourceCellPreflight{}, e
}

func (r *Registry) SourceVisibility(ctx context.Context, connector string) (SourceVisibility, error) {
	if ctx == nil || r == nil {
		return SourceVisibility{}, errors.New("source registry and context are required")
	}
	if err := ctx.Err(); err != nil {
		return SourceVisibility{}, err
	}
	r.mu.RLock()
	entry, ok := r.metadata[connector]
	r.mu.RUnlock()
	if !ok {
		return SourceVisibility{}, &SourceSelectionInputError{Selection: SourceCellSelection{Source: SourceOperationKey{Connector: connector}}, Code: "source_cell_not_found"}
	}
	if entry.SourceVisibility.Connector != connector {
		return SourceVisibility{}, &SourceVisibilityDataError{Connector: connector, Cause: errors.New("source artifact does not belong to selected metadata entry")}
	}
	return DecodeSourceVisibility(ctx, entry.SourceVisibility)
}

// Every non-optional JSON member must be present. Null is allowed only for a
// nullable pointer, not as a surrogate for a required scalar, object or array.
func validateSourceRequired(value any, typ reflect.Type) error {
	if typ.Kind() == reflect.Pointer {
		if value == nil {
			return nil
		}
		return validateSourceRequired(value, typ.Elem())
	}
	if value == nil {
		return errors.New("required source metadata value is null")
	}
	switch typ.Kind() {
	case reflect.Struct:
		object, ok := value.(map[string]any)
		if !ok {
			return errors.New("source metadata object required")
		}
		for i := 0; i < typ.NumField(); i++ {
			field := typ.Field(i)
			tag := field.Tag.Get("json")
			if tag == "-" || tag == "" {
				continue
			}
			name, optional, _ := strings.Cut(tag, ",")
			child, present := object[name]
			if !present {
				if optional == "omitempty" {
					continue
				}
				return fmt.Errorf("required source metadata member %s absent", name)
			}
			if err := validateSourceRequired(child, field.Type); err != nil {
				return fmt.Errorf("%s: %w", name, err)
			}
		}
	case reflect.Slice:
		items, ok := value.([]any)
		if !ok {
			return errors.New("source metadata array required")
		}
		for _, item := range items {
			if err := validateSourceRequired(item, typ.Elem()); err != nil {
				return err
			}
		}
	}
	return nil
}

// PreflightSource selects only validated immutable metadata. No supplied view
// or artifact reference can skip validation and become an execution claim.
func (r *Registry) PreflightSource(ctx context.Context, selection SourceCellSelection) (SourceCellPreflight, error) {
	v, err := r.SourceVisibility(ctx, selection.Source.Connector)
	if err != nil {
		return SourceCellPreflight{}, err
	}
	view, err := InspectSourceCell(v, selection)
	if err != nil {
		return SourceCellPreflight{}, err
	}
	return preflightSourceCell(view)
}

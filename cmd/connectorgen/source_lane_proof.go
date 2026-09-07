package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"os"
	"reflect"
	"sort"
	"strconv"
	"strings"
)

const sourceLaneProofPath = "data/connector-canon/batch1-source-lane-proofs.json"

var errSourceLaneProofCapacity = errors.New("proof record capacity exceeded")

type sourceLaneProofInput struct {
	Path   string `json:"path"`
	SHA256 string `json:"sha256"`
	Role   string `json:"role"`
}

type sourceLaneProofRecord struct {
	ID                 string                 `json:"id"`
	Key                sourceOperationKey     `json:"key"`
	Lane               string                 `json:"lane"`
	Targets            []sourceLaneTargetRef  `json:"targets"`
	Inputs             []sourceLaneProofInput `json:"inputs,omitempty"`
	InputSetSHA256     string                 `json:"input_set_sha256,omitempty"`
	TestPath           string                 `json:"test_path"`
	TestSymbol         string                 `json:"test_symbol"`
	SelectedTest       string                 `json:"selected_test"`
	Package            string                 `json:"package"`
	ExecutionClass     string                 `json:"execution_class"`
	Scope              string                 `json:"scope"`
	ReceiptPath        string                 `json:"receipt_path"`
	ReceiptSHA256      string                 `json:"receipt_sha256"`
	ObservableContract string                 `json:"observable_contract"`
	Limitations        []string               `json:"limitations"`
	ClaimCurrent       bool                   `json:"claim_current"`
}

// Reviews are supplied by the authoring caller, never loaded from proof JSON.
// The owner reviews assertion meaning, complete input coverage and required
// lane targets. Hashes alone cannot establish those semantic facts.
type sourceLaneProofReview struct {
	Record  sourceLaneProofRecord
	Fixture bool
}

type sourceLaneProofInputs struct {
	Records           []sourceLaneProofRecord
	Diagnostics       []sourceLaneDiagnostic
	accepted          map[string]bool
	byCell            map[sourceLaneProofCell]sourceLaneProofRecord
	diagnosticsByCell map[sourceLaneProofCell][]sourceLaneDiagnostic
	globalDiagnostics []sourceLaneDiagnostic
	Stats             sourceLaneProofReadStats
}

type sourceLaneProofCell struct {
	Key  sourceOperationKey
	Lane string
}
type sourceLaneProofInputSet struct {
	SHA256 string                 `json:"sha256"`
	Inputs []sourceLaneProofInput `json:"inputs"`
}
type sourceLaneProofCatalog struct {
	Reviews   []sourceLaneProofReview
	InputSets []sourceLaneProofInputSet
}
type sourceLaneProofLimits struct {
	DocumentBytes, UniqueBytes, InputBytes, TargetBytes int64
	Tuples, Files, Pins                                 int
}
type sourceLaneProofPolicy struct {
	MaxRecords int
	keys       map[sourceOperationKey]bool
	limits     sourceLaneProofLimits
	// Tests observe the completed real reader call. This is not a substitute
	// reader and cannot fabricate bytes, hashes, or accepted file identities.
	afterRead func(sourceLaneProofReadEvent)
}
type sourceLaneProofReadEvent = sourceProofReadEvent
type sourceLaneProofReadStats = sourceProofReadStats

func newSourceLaneProofPolicy(keys []sourceOperationKey) (sourceLaneProofPolicy, error) {
	p := sourceLaneProofPolicy{keys: map[sourceOperationKey]bool{}, limits: sourceLaneProofLimits{DocumentBytes: 128 << 20, UniqueBytes: 512 << 20, InputBytes: 4 << 20, TargetBytes: 64 << 20, Tuples: 131072, Files: 65536, Pins: 4096}}
	if len(keys) > math.MaxInt/7 {
		return p, fmt.Errorf("proof cohort capacity overflow")
	}
	for _, key := range keys {
		if !validSourceID(key.Connector) || !validSourceID(key.Inventory) || !validSourceID(key.ID) || p.keys[key] {
			return p, fmt.Errorf("proof cohort key invalid")
		}
		p.keys[key] = true
	}
	p.MaxRecords = len(keys) * 7
	return p, nil
}

// The versioned set digest covers complete sorted path/digest/role tuples.
func sourceLaneProofSetHash(inputs []sourceLaneProofInput) string {
	ordered := append([]sourceLaneProofInput(nil), inputs...)
	sort.Slice(ordered, func(i, j int) bool {
		if ordered[i].Path != ordered[j].Path {
			return ordered[i].Path < ordered[j].Path
		}
		if ordered[i].SHA256 != ordered[j].SHA256 {
			return ordered[i].SHA256 < ordered[j].SHA256
		}
		return ordered[i].Role < ordered[j].Role
	})
	raw, _ := json.Marshal(struct {
		Version int                    `json:"version"`
		Inputs  []sourceLaneProofInput `json:"inputs"`
	}{1, ordered})
	return sourceBytesHash(raw)
}

type sourceLaneProofSet struct {
	inputs []sourceLaneProofInput
	tests  map[string]bool
	valid  bool
	hash   string
}

func validateSourceLaneProofSet(inputs []sourceLaneProofInput, limit int) sourceLaneProofSet {
	set := sourceLaneProofSet{inputs: inputs, tests: map[string]bool{}}
	if len(inputs) == 0 || len(inputs) > limit {
		return set
	}
	seen := map[string]bool{}
	code, mod, sum := false, false, false
	for _, in := range inputs {
		if seen[in.Path] || !sourceLaneProofSafePath(in.Path) || !sourceLaneProofDigest(in.SHA256) {
			return set
		}
		seen[in.Path] = true
		switch in.Role {
		case "code", "test":
			if (!strings.HasPrefix(in.Path, "cmd/") && !strings.HasPrefix(in.Path, "internal/")) || !strings.HasSuffix(in.Path, ".go") {
				return set
			}
			if in.Role == "code" {
				code = true
			} else {
				set.tests[in.Path] = true
			}
		case "dependency":
			switch in.Path {
			case "go.mod":
				mod = true
			case "go.sum":
				sum = true
			default:
				return set
			}
		default:
			return set
		}
	}
	set.valid = code && mod && sum && len(set.tests) > 0
	return set
}

// Preserve lane policy in the lane adapter; only file mechanics are shared.
type sourceLaneProofCache struct {
	*sourceProofFileCache
	policy    sourceLaneProofPolicy
	setChecks map[string]string
	setPaths  map[string][]string
}

func newSourceLaneProofCache(ctx context.Context, root *os.Root, policy sourceLaneProofPolicy) sourceLaneProofCache {
	return sourceLaneProofCache{sourceProofFileCache: &sourceProofFileCache{
		ctx: ctx, root: root, limits: sourceProofFileLimits{UniqueBytes: policy.limits.UniqueBytes, Files: policy.limits.Files},
		afterRead: policy.afterRead, files: map[string]*sourceProofFile{},
	}, policy: policy}
}

// Decode incrementally, refusing array/count/depth growth before allocating
// unbounded record/set slices. The byte read itself is independently bounded.
func decodeSourceLaneProofDocument(ctx context.Context, raw []byte, p sourceLaneProofPolicy) ([]sourceLaneProofRecord, []sourceLaneProofInputSet, error) {
	d := json.NewDecoder(bytes.NewReader(raw))
	d.UseNumber()
	tok, err := d.Token()
	if err != nil || tok != json.Delim('{') {
		return nil, nil, fmt.Errorf("invalid proof document")
	}
	seen := map[string]bool{}
	version := 0
	records := []sourceLaneProofRecord(nil)
	sets := []sourceLaneProofInputSet(nil)
	tuples := 0
	for d.More() {
		if ctx.Err() != nil {
			return nil, nil, ctx.Err()
		}
		token, err := d.Token()
		name, ok := token.(string)
		if err != nil || !ok || seen[name] {
			return nil, nil, fmt.Errorf("invalid proof field")
		}
		seen[name] = true
		switch name {
		case "schema_version":
			if d.Decode(&version) != nil {
				return nil, nil, fmt.Errorf("invalid version")
			}
		case "records", "input_sets":
			if token, err := d.Token(); err != nil || token != json.Delim('[') {
				return nil, nil, fmt.Errorf("invalid proof collection")
			}
			if name == "records" {
				records = []sourceLaneProofRecord{}
			} else {
				sets = []sourceLaneProofInputSet{}
			}
			count := 0
			for d.More() {
				if ctx.Err() != nil {
					return nil, nil, ctx.Err()
				}
				if count >= p.MaxRecords {
					return records, sets, errSourceLaneProofCapacity
				}
				count++
				// Tokens are bounded before decoding the per-entry typed object.
				var entry json.RawMessage
				if err := d.Decode(&entry); err != nil {
					return nil, nil, err
				}
				if err := sourceLaneProofEntryBudget(entry, &tuples, p); err != nil {
					return nil, nil, err
				}
				if name == "records" {
					var representation struct {
						Inputs json.RawMessage `json:"inputs"`
						Set    json.RawMessage `json:"input_set_sha256"`
					}
					if json.Unmarshal(entry, &representation) != nil || len(representation.Inputs) > 0 && len(representation.Set) > 0 {
						return nil, nil, fmt.Errorf("ambiguous input representation")
					}
					var r sourceLaneProofRecord
					if decodeSourceJSON(entry, &r) != nil || decodeStrictJSON(entry, &r) != nil {
						return nil, nil, fmt.Errorf("invalid record")
					}
					records = append(records, r)
				} else {
					var s sourceLaneProofInputSet
					if decodeSourceJSON(entry, &s) != nil || decodeStrictJSON(entry, &s) != nil {
						return nil, nil, fmt.Errorf("invalid input set")
					}
					sets = append(sets, s)
				}
			}
			if _, err := d.Token(); err != nil {
				return nil, nil, err
			}
		default:
			return nil, nil, fmt.Errorf("unknown proof field")
		}
	}
	if _, err := d.Token(); err != nil {
		return nil, nil, err
	}
	if d.Decode(new(any)) != io.EOF || version != 1 || !seen["records"] {
		return nil, nil, fmt.Errorf("invalid proof envelope")
	}
	return records, sets, nil
}

func sourceLaneProofEntryBudget(raw []byte, tuples *int, p sourceLaneProofPolicy) error {
	d := json.NewDecoder(bytes.NewReader(raw))
	d.UseNumber()
	var walk func(int, string) error
	walk = func(depth int, field string) error {
		if depth > 16 {
			return fmt.Errorf("proof nesting exceeded")
		}
		tok, err := d.Token()
		if err != nil {
			return err
		}
		delim, ok := tok.(json.Delim)
		if !ok {
			return nil
		}
		switch delim {
		case '{':
			for d.More() {
				key, err := d.Token()
				if err != nil {
					return err
				}
				name, ok := key.(string)
				if !ok {
					return fmt.Errorf("invalid proof key")
				}
				if err := walk(depth+1, name); err != nil {
					return err
				}
			}
			_, err = d.Token()
			return err
		case '[':
			count := 0
			for d.More() {
				count++
				if field == "inputs" {
					*tuples++
					if count > p.limits.Pins || *tuples > p.limits.Tuples {
						return fmt.Errorf("proof tuple limit exceeded")
					}
				}
				if field == "targets" && count > 64 {
					return fmt.Errorf("proof target limit exceeded")
				}
				if count > p.limits.Tuples {
					return fmt.Errorf("proof array limit exceeded")
				}
				if err := walk(depth+1, field); err != nil {
					return err
				}
			}
			_, err = d.Token()
			return err
		default:
			return fmt.Errorf("invalid proof token")
		}
	}
	return walk(0, "")
}

func loadSourceLaneProofs(ctx context.Context, repo string, policy sourceLaneProofPolicy, catalog sourceLaneProofCatalog) sourceLaneProofInputs {
	result := sourceLaneProofInputs{Records: []sourceLaneProofRecord{}, accepted: map[string]bool{}, byCell: map[sourceLaneProofCell]sourceLaneProofRecord{}, diagnosticsByCell: map[sourceLaneProofCell][]sourceLaneDiagnostic{}}
	add := func(r sourceLaneProofRecord, code, severity string) {
		d := sourceLaneDiagnostic{Key: r.Key, Lanes: []string{r.Lane}, Stage: "proof", Code: code, Pointer: sourceLaneProofPath, Owner: r.Key.Connector, Severity: severity}
		result.Diagnostics = append(result.Diagnostics, d)
		if r.Key == (sourceOperationKey{}) {
			result.globalDiagnostics = append(result.globalDiagnostics, d)
		} else {
			key := sourceLaneProofCell{r.Key, r.Lane}
			result.diagnosticsByCell[key] = append(result.diagnosticsByCell[key], d)
		}
	}
	if ctx == nil {
		ctx = context.Background()
	}
	if policy.keys == nil || policy.MaxRecords < 0 || policy.MaxRecords > len(policy.keys)*7 || policy.limits.UniqueBytes <= 0 || policy.limits.UniqueBytes > 512<<20 || policy.limits.DocumentBytes <= 0 || policy.limits.DocumentBytes > 128<<20 || policy.limits.Files <= 0 || policy.limits.Files > 65536 || policy.limits.Tuples <= 0 || policy.limits.Tuples > 131072 || policy.limits.Pins <= 0 || policy.limits.Pins > 4096 || policy.limits.InputBytes <= 0 || policy.limits.InputBytes > 4<<20 || policy.limits.TargetBytes <= 0 || policy.limits.TargetBytes > 64<<20 {
		add(sourceLaneProofRecord{}, "proof_policy_invalid", "error")
		return result
	}
	root, err := os.OpenRoot(repo)
	if err != nil {
		add(sourceLaneProofRecord{}, "proof_root_unavailable", "error")
		return result
	}
	// This confinement handle is read-only: no data is flushed by Close.
	// Read/validation failures already use the caller's existing result channel;
	// teardown is not an additional source-consistency or proof-authority gate.
	defer func() { _ = root.Close() }()
	cache := newSourceLaneProofCache(ctx, root, policy)
	raw, doc := cache.get(sourceLaneProofPath, policy.limits.DocumentBytes, false)
	if doc.code == "missing" {
		return result
	}
	if doc.code != "" {
		code := "proof_document_invalid"
		if doc.code == "proof_cancelled" {
			code = doc.code
		}
		add(sourceLaneProofRecord{}, code, "error")
		result.Stats = cache.stats
		return result
	}
	records, declared, err := decodeSourceLaneProofDocument(ctx, raw, policy)
	if err != nil {
		if errors.Is(err, errSourceLaneProofCapacity) {
			add(sourceLaneProofRecord{}, "proof_record_capacity_exceeded", "error")
		}
		code := "proof_document_invalid"
		if ctx.Err() != nil {
			code = "proof_cancelled"
		}
		add(sourceLaneProofRecord{}, code, "error")
		result.Stats = cache.stats
		return result
	}
	result.Records = records
	trustedSets := map[string]sourceLaneProofInputSet{}
	duplicateTrustedSets := map[string]bool{}
	for _, s := range catalog.InputSets {
		if _, ok := trustedSets[s.SHA256]; ok {
			duplicateTrustedSets[s.SHA256] = true
		}
		trustedSets[s.SHA256] = s
	}
	sets := map[string]sourceLaneProofSet{}
	setCounts := map[string]int{}
	for _, s := range declared {
		setCounts[s.SHA256]++
	}
	for _, s := range declared {
		if ctx.Err() != nil {
			add(sourceLaneProofRecord{}, "proof_cancelled", "error")
			break
		}
		valid := validateSourceLaneProofSet(s.Inputs, policy.limits.Pins)
		trusted, reviewed := trustedSets[s.SHA256]
		if setCounts[s.SHA256] != 1 || duplicateTrustedSets[s.SHA256] || !valid.valid || s.SHA256 != sourceLaneProofSetHash(s.Inputs) || !reviewed || s.SHA256 != sourceLaneProofSetHash(trusted.Inputs) {
			add(sourceLaneProofRecord{}, "proof_input_set_invalid", "error")
			continue
		}
		valid.hash = s.SHA256
		sets[s.SHA256] = valid
	}
	reviews := map[string]sourceLaneProofReview{}
	reviewCounts := map[string]int{}
	ids := map[string]int{}
	cells := map[sourceLaneProofCell]int{}
	for _, review := range catalog.Reviews {
		reviews[review.Record.ID] = review
		reviewCounts[review.Record.ID]++
	}
	for _, r := range records {
		ids[r.ID]++
		cells[sourceLaneProofCell{r.Key, r.Lane}]++
	}
	inlineSets := map[string]sourceLaneProofSet{}
	used := map[string][]string{}
	usedSet := map[string]string{}
	for _, r := range records {
		if ctx.Err() != nil {
			add(r, "proof_cancelled", "error")
			continue
		}
		if ids[r.ID] != 1 || cells[sourceLaneProofCell{r.Key, r.Lane}] != 1 {
			add(r, "proof_duplicate_claim", "error")
			continue
		}
		review, count := reviews[r.ID], reviewCounts[r.ID]
		if count > 1 || count == 1 && !reflect.DeepEqual(review.Record, r) {
			add(r, "proof_review_mismatch", "error")
			continue
		}
		if !sourceLaneProofShape(r) || (len(r.Inputs) > 0) == (r.InputSetSHA256 != "") {
			add(r, "proof_record_invalid", "error")
			continue
		}
		if !policy.keys[r.Key] {
			add(r, "proof_source_unknown", "error")
			continue
		}
		var set sourceLaneProofSet
		if r.InputSetSHA256 != "" {
			set = sets[r.InputSetSHA256]
		} else {
			hash := sourceLaneProofSetHash(r.Inputs)
			var ok bool
			set, ok = inlineSets[hash]
			if !ok {
				set = validateSourceLaneProofSet(r.Inputs, policy.limits.Pins)
				set.hash = hash
				inlineSets[hash] = set
			}
		}
		if !set.valid || !set.tests[r.TestPath] {
			add(r, "proof_record_invalid", "error")
			continue
		}
		if count == 0 {
			severity := "deficit"
			if r.ClaimCurrent {
				severity = "error"
			}
			add(r, "proof_review_unavailable", severity)
			continue
		}
		productionScope := (r.Scope == "provider_live" && r.ExecutionClass == "C1") || (r.Scope == "connector_hermetic" && r.ExecutionClass == "C2")
		if (review.Fixture && (r.Scope != "hermetic_fixture" || r.ExecutionClass != "C2" || r.Key.Inventory != "fixture")) || (!review.Fixture && (!productionScope || r.Key.Inventory == "fixture")) {
			add(r, "proof_scope_unproven", "deficit")
			continue
		}
		code, severity, paths := assessSourceLaneProofFiles(&cache, r, set.inputs)
		usedSet[r.ID] = set.hash
		used[r.ID] = paths
		if code != "" {
			add(r, code, severity)
			continue
		}
		result.accepted[r.ID] = true
	}
	cache.finalize()
	if ctx.Err() != nil {
		doc.code = "proof_cancelled"
	}
	setInvalid := map[string]string{}
	for hash, paths := range cache.setPaths {
		for _, name := range paths {
			if f := cache.files[name]; f != nil && f.code != "" {
				setInvalid[hash] = f.code
				break
			}
		}
	}
	if doc.code != "" {
		add(sourceLaneProofRecord{}, doc.code, "error")
	}
	for _, r := range records {
		if !result.accepted[r.ID] {
			continue
		}
		bad := doc.code
		if code := setInvalid[usedSet[r.ID]]; code != "" {
			bad = code
		}
		for _, name := range used[r.ID] {
			if f := cache.files[name]; f != nil && f.code != "" {
				bad = f.code
				break
			}
		}
		if bad != "" {
			delete(result.accepted, r.ID)
			add(r, bad, "error")
			continue
		}
		result.byCell[sourceLaneProofCell{r.Key, r.Lane}] = r
	}
	result.Stats = cache.stats
	return result
}

func assessSourceLaneProofFiles(cache *sourceLaneProofCache, r sourceLaneProofRecord, inputs []sourceLaneProofInput) (string, string, []string) {
	paths := []string{sourceLaneProofPath}
	stale := func(code string) (string, string, []string) {
		if r.ClaimCurrent {
			return "proof_current_claim_invalid", "error", paths
		}
		return code, "deficit", paths
	}
	setID := r.InputSetSHA256
	if setID == "" {
		setID = sourceLaneProofSetHash(inputs)
	}
	if cache.setChecks == nil {
		cache.setChecks = map[string]string{}
		cache.setPaths = map[string][]string{}
	}
	code, checked := cache.setChecks[setID]
	if !checked {
		for _, in := range inputs {
			cache.setPaths[setID] = append(cache.setPaths[setID], in.Path)
			_, f := cache.get(in.Path, cache.policy.limits.InputBytes, false)
			if f.code == "missing" || f.code == "" && f.hash != in.SHA256 {
				code = "proof_inputs_outdated"
				break
			}
			if f.code != "" {
				code = f.code
				break
			}
		}
		cache.setChecks[setID] = code
	}
	if code == "proof_inputs_outdated" {
		return stale(code)
	}
	if code != "" {
		return code, "error", paths
	}
	for _, ref := range r.Targets {
		paths = append(paths, ref.Artifact)
		_, f := cache.get(ref.Artifact, cache.policy.limits.TargetBytes, false)
		if f.code == "missing" {
			return stale("proof_targets_outdated")
		}
		if f.code != "" {
			return f.code, "error", paths
		}
		if f.hash != ref.ArtifactSHA256 {
			return stale("proof_targets_outdated")
		}
	}
	paths = append(paths, r.ReceiptPath)
	_, f := cache.get(r.ReceiptPath, cache.policy.limits.InputBytes, true)
	if f.code == "missing" {
		return stale("proof_receipt_unavailable")
	}
	if f.code != "" {
		return f.code, "error", paths
	}
	if f.hash != r.ReceiptSHA256 {
		return "proof_receipt_digest_invalid", "error", paths
	}
	if !f.parsed || !f.result.successful(r) {
		return "proof_result_invalid", "error", paths
	}
	return "", "", paths
}

func sourceLaneProofSafePath(p string) bool { return sourceProofSafePath(p) }
func sourceLaneProofDigest(s string) bool   { return sourceProofDigest(s) }

func sourceLaneProofShape(r sourceLaneProofRecord) bool {
	if !validSourceID(r.ID) || !validSourceID(r.Key.Connector) || !validSourceID(r.Key.Inventory) || !validSourceID(r.Key.ID) || !validSourceID(r.ObservableContract) || len(r.Limitations) == 0 || len(r.Targets) == 0 || len(r.Targets) > 64 || len(r.Inputs) > 4096 {
		return false
	}
	lane := false
	for _, name := range sourceLaneNames() {
		lane = lane || name == r.Lane
	}
	if !lane {
		return false
	}
	if !sourceLaneProofSafePath(r.TestPath) || !strings.HasPrefix(r.TestPath, "cmd/") && !strings.HasPrefix(r.TestPath, "internal/") || !strings.HasSuffix(r.TestPath, "_test.go") || !validSourceID(r.Package) || !strings.HasPrefix(r.TestSymbol, "Test") || strings.Contains(r.TestSymbol, "/") || !strings.HasPrefix(r.SelectedTest, r.TestSymbol+"/") || !validSourceID(r.SelectedTest) {
		return false
	}
	if !sourceLaneProofSafePath(r.ReceiptPath) || !strings.HasPrefix(r.ReceiptPath, "data/connector-canon/proof-receipts/") || !strings.HasSuffix(r.ReceiptPath, ".jsonl") || !sourceLaneProofDigest(r.ReceiptSHA256) {
		return false
	}
	for _, s := range r.Limitations {
		if !validSourceID(s) {
			return false
		}
	}
	for i, ref := range r.Targets {
		duplicate := false
		for _, prior := range r.Targets[:i] {
			duplicate = duplicate || sourceLaneTargetRefEqual(prior, ref)
		}
		pointerValid := strings.HasPrefix(ref.Pointer, "/")
		if ref.Kind == "schema" {
			parts := strings.Split(ref.CanonicalPointer, "/")
			canonicalRole := len(parts) == 5 && parts[0] == "" && parts[1] == "operations" && parts[3] == "schema_refs" && parts[4] == string(ref.SchemaRole)
			if canonicalRole {
				index, err := strconv.Atoi(parts[2])
				canonicalRole = err == nil && index >= 0 && strconv.Itoa(index) == parts[2]
			}
			pointerValid = ref.Pointer == "" && ref.Artifact == "internal/connectors/defs/"+ref.Connector+"/"+ref.ID && ref.SchemaRole != "" && canonicalRole
		}
		if duplicate || sourceLaneTargetRefShape(ref) != nil || !pointerValid || ref.Connector != r.Key.Connector || ref.Lane != r.Lane || !validSourceID(ref.ID) || !sourceLaneProofSafePath(ref.Artifact) || !strings.HasPrefix(ref.Artifact, "internal/connectors/defs/"+r.Key.Connector+"/") || !strings.HasSuffix(ref.Artifact, ".json") || !sourceLaneProofDigest(ref.ArtifactSHA256) || !validSourceID(ref.CanonicalID) || !strings.HasPrefix(ref.CanonicalPointer, "/") || !sourceLaneProofDigest(ref.Generation) {
			return false
		}
	}
	return true
}

// The reviewed original receipt supplies execution authenticity; this checks
// actual selected result events, not a PASS string or declaration. No code or
// command from the document is ever executed.
func (result sourceProofResult) successful(r sourceLaneProofRecord) bool {
	return result.successfulSelection(r.Package, r.SelectedTest)
}

// assessSourceLaneProof preserves source membership and independently derived
// exclusions/gaps. Evidence absence is an explicit deficit, never applicability.
func assessSourceLaneProof(key sourceOperationKey, cells []sourceLaneCell, inputs sourceLaneProofInputs) []sourceLaneCell {
	out := append([]sourceLaneCell(nil), cells...)
	for i := range out {
		out[i].Diagnostics = append([]sourceLaneDiagnostic(nil), out[i].Diagnostics...)
		out[i].ProofRefs = []string{}
		for _, d := range inputs.globalDiagnostics {
			d.Key = key
			d.Owner = key.Connector
			d.Lanes = []string{out[i].Lane}
			out[i].Diagnostics = append(out[i].Diagnostics, d)
		}
		out[i].Diagnostics = append(out[i].Diagnostics, inputs.diagnosticsByCell[sourceLaneProofCell{key, out[i].Lane}]...)
		if out[i].State != "not_applicable" && out[i].State != "missing_foundation" {
			out[i].State = "mapped_unproven"
			if r, exists := inputs.byCell[sourceLaneProofCell{key, out[i].Lane}]; exists {
				blocked := out[i].Applicability != "applicable" || len(out[i].References) == 0
				for _, d := range out[i].Diagnostics {
					blocked = blocked || d.Severity == "error" || d.Stage == "reference"
				}
				for _, required := range r.Targets {
					found := false
					for _, ref := range out[i].References {
						found = found || sourceLaneTargetRefEqual(ref, required)
					}
					blocked = blocked || !found
				}
				if blocked {
					out[i].Diagnostics = append(out[i].Diagnostics, sourceLaneDiagnostic{Key: key, Lanes: []string{out[i].Lane}, Stage: "proof", Code: "proof_prerequisites_unproven", Owner: key.Connector, Severity: "deficit"})
					continue
				}
				out[i].State = "implemented"
				out[i].ProofRefs = []string{r.ID}
				out[i].Reason = sourceLaneReason{Code: "behavior_proven", Text: "Reviewed lane-specific behavior is current within the proof's declared scope."}
			}
			if out[i].State != "implemented" {
				out[i].Diagnostics = append(out[i].Diagnostics, sourceLaneDiagnostic{Key: key, Lanes: []string{out[i].Lane}, Stage: "proof", Code: "proof_unavailable", Owner: key.Connector, Severity: "deficit"})
			}
		}
	}
	return out
}

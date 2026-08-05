package engine

import (
	"encoding/json"
	"fmt"
	"reflect"
	"regexp"
	"regexp/syntax"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"polymetrics.ai/internal/safety"
)

// Schema is a compiled instance of the engine's minimal draft-07 subset. It is
// compiled once (CompileSchema) and validated many times (Validate). The
// compiled form is opaque outside the package; callers only see the accessor
// methods below.
type Schema struct {
	node *schemaNode
}

// schemaNode is the compiled representation of one (sub-)schema object.
type schemaNode struct {
	// types holds the accepted JSON types ("string", "number", "integer",
	// "boolean", "object", "array", "null"); empty means "any type".
	types []string

	required             []string
	properties           map[string]*schemaNode
	items                *schemaNode
	allOf                []*schemaNode
	anyOf                []*schemaNode
	oneOf                []*schemaNode
	enum                 []any
	pattern              *regexp.Regexp
	patternDomain        regexpStringDomain
	minProperties        int
	hasMinProperties     bool
	additionalProperties bool // true unless explicitly set to false
	hasAdditionalProps   bool

	// minItems/maxItems are the array-cardinality pair. Each carries its own
	// has* flag rather than relying on the zero value, because an EXPLICIT
	// "minItems": 0 (a documented "may be empty") must stay distinguishable
	// from an absent keyword — the same distinction PaginationSpec.StartPage
	// solves one layer up with a pointer.
	//
	// Bounded provider search depends on the same pair: an "ids[] bounded to
	// 100" contract has to be declarable in the bundle and enforceable at load
	// time, and unknown keywords are a compile error in this dialect.
	minItems    int
	hasMinItems bool
	maxItems    int
	hasMaxItems bool

	// extensions
	secret      bool     // x-secret
	primaryKey  []string // x-primary-key (only meaningful at the root)
	cursorField string   // x-cursor-field (only meaningful at the root)

	// defaultVal/hasDefault capture the "default" annotation keyword's raw
	// decoded value (gap-loop cycle-1 item 6, REVIEW-A.md C3): "default" was
	// previously accepted-but-only-preserved (never read back out anywhere);
	// this is now consumed by Defaults() so the engine can materialize a
	// spec.json property's default into RuntimeConfig.Config when a caller's
	// config omits that key entirely.
	defaultVal any
	hasDefault bool
}

// annotationKeywords are accepted but only preserved, never enforced.
var annotationKeywords = map[string]bool{
	"format":      true,
	"default":     true,
	"title":       true,
	"description": true,
	"$schema":     true,
}

// structuralKeywords are the only keywords this dialect understands
// structurally.
var structuralKeywords = map[string]bool{
	"type":                 true,
	"required":             true,
	"properties":           true,
	"items":                true,
	"allOf":                true,
	"anyOf":                true,
	"oneOf":                true,
	"enum":                 true,
	"pattern":              true,
	"minProperties":        true,
	"minItems":             true,
	"maxItems":             true,
	"additionalProperties": true,
	"x-secret":             true,
	"x-primary-key":        true,
	"x-cursor-field":       true,
}

var validTypes = map[string]bool{
	"string":  true,
	"number":  true,
	"integer": true,
	"boolean": true,
	"object":  true,
	"array":   true,
	"null":    true,
}

// CompileSchema parses and compiles a draft-07 subset schema document. Unknown
// keywords are a compile error, keeping bundles honest.
func CompileSchema(raw json.RawMessage) (*Schema, error) {
	var m map[string]json.RawMessage
	if err := json.Unmarshal(raw, &m); err != nil {
		return nil, fmt.Errorf("compile schema: invalid json: %w", err)
	}
	node, err := compileNode(m)
	if err != nil {
		return nil, err
	}
	return &Schema{node: node}, nil
}

type SchemaPathInfo struct {
	Types    []string
	IsObject bool
	IsArray  bool
}

type RequestBodyInput struct {
	Type       string
	Values     []string
	Format     string
	AllowEmpty *bool
	Required   bool
	MinItems   int
	MaxItems   int
	arrayItem  bool
}

type dynamicRequestBodyValue struct{}

type regexpStringDomain struct {
	program *syntax.Prog
}

func (s *Schema) RequiredMappingPaths() ([]string, error) {
	if s == nil || s.node == nil {
		return nil, fmt.Errorf("schema is nil")
	}
	if err := s.node.rejectAlternativeMappings(""); err != nil {
		return nil, err
	}
	paths := map[string]bool{}
	if err := s.node.collectRequiredMappingPaths("", paths); err != nil {
		return nil, err
	}
	out := make([]string, 0, len(paths))
	for path := range paths {
		out = append(out, path)
	}
	sort.Strings(out)
	return out, nil
}

func (s *Schema) RequiredRequestBodyPointers() ([]string, error) {
	if s == nil || s.node == nil {
		return nil, fmt.Errorf("schema is nil")
	}
	if err := s.node.rejectAlternativeRequestMappings("/body"); err != nil {
		return nil, err
	}
	paths := map[string]bool{}
	if err := s.node.collectRequiredRequestMappings("/body", paths, s.node); err != nil {
		return nil, err
	}
	out := make([]string, 0, len(paths))
	for path := range paths {
		out = append(out, path)
	}
	sort.Strings(out)
	return out, nil
}

func (s *Schema) RequestFieldPointerInfo(pointer string) (SchemaPathInfo, error) {
	if s == nil || s.node == nil {
		return SchemaPathInfo{}, fmt.Errorf("schema is nil")
	}
	namespace, tokens, err := parseRequestFieldPointer(pointer)
	if err != nil {
		return SchemaPathInfo{}, err
	}
	if namespace != "body" {
		return SchemaPathInfo{}, fmt.Errorf("request field pointer %q is not body-scoped", pointer)
	}
	node, _, err := s.requestMappingNode(tokens, pointer)
	if err != nil {
		return SchemaPathInfo{}, err
	}
	return node.mappingPathInfo()
}

func (s *Schema) ValidateRequestBodyInputType(pointer, inputType string) error {
	if inputType == "enum" {
		inputType = "string"
	}
	return s.ValidateRequestBodyInput(pointer, RequestBodyInput{Type: inputType})
}

func (s *Schema) ValidateRequestBodyInput(pointer string, input RequestBodyInput) error {
	info, err := s.RequestFieldPointerInfo(pointer)
	if err != nil {
		return err
	}
	if !inputTypeAccepted(info, input.Type) {
		if input.Type != "" && input.Type != "string" && input.Type != "enum" && input.Type != "integer" && input.Type != "boolean" && input.Type != "string_array" {
			return fmt.Errorf("request mapping %q uses unsupported input type %q", pointer, input.Type)
		}
		return fmt.Errorf("request mapping %q input type %q is incompatible with schema types %v", pointer, input.Type, info.Types)
	}
	namespace, tokens, err := parseRequestFieldPointer(pointer)
	if err != nil {
		return err
	}
	node, _, err := s.requestMappingNode(tokens, pointer)
	if err != nil {
		return err
	}
	var itemNode *schemaNode
	var itemPointer string
	if input.Type == "string_array" && !containsSchemaType(info.Types, "any") {
		itemTokens := append(append([]string(nil), tokens...), "0")
		itemPointer = requestFieldPointer(namespace, itemTokens...)
		itemInfo, err := s.RequestFieldPointerInfo(itemPointer)
		if err != nil {
			return fmt.Errorf("string_array mapping %q has no addressable item schema: %w", pointer, err)
		}
		if !containsSchemaType(itemInfo.Types, "string") && !containsSchemaType(itemInfo.Types, "any") {
			return fmt.Errorf("string_array mapping %q requires string items, schema accepts %v", pointer, itemInfo.Types)
		}
		itemNode, _, err = s.requestMappingNode(itemTokens, itemPointer)
		if err != nil {
			return err
		}
	}
	if input.Type == "string_array" {
		if err := node.validateRequestInputCardinality(pointer, input); err != nil {
			return err
		}
	}
	if err := node.validateRequestInputDomain(pointer, input); err != nil {
		return err
	}
	if itemNode != nil {
		if err := itemNode.validateRequestInputDomain(itemPointer, RequestBodyInput{Type: "string", Required: input.Required, arrayItem: true}); err != nil {
			return err
		}
	}
	return nil
}

func inputTypeAccepted(info SchemaPathInfo, inputType string) bool {
	accepts := func(types ...string) bool {
		for _, actual := range info.Types {
			if actual == "any" {
				return true
			}
			for _, expected := range types {
				if actual == expected {
					return true
				}
			}
		}
		return false
	}
	switch inputType {
	case "", "string", "enum":
		return accepts("string")
	case "integer":
		return accepts("integer", "number")
	case "boolean":
		return accepts("boolean")
	case "string_array":
		return accepts("array") && (info.IsArray || containsSchemaType(info.Types, "any"))
	default:
		return false
	}
}

func (n *schemaNode) validateRequestInputDomain(pointer string, input RequestBodyInput) error {
	if input.Type == "enum" && len(input.Values) == 0 {
		return fmt.Errorf("enum mapping %q has no declared input values", pointer)
	}
	if input.Type == "enum" {
		hasEmittableValue := false
		for _, candidate := range input.Values {
			if !requestInputDomainAcceptsValue(input, candidate) {
				continue
			}
			hasEmittableValue = true
			if err := n.validate(candidate, ""); err != nil {
				return fmt.Errorf("request mapping %q enum value %q is not accepted by schema constraints: %w", pointer, candidate, err)
			}
		}
		if !hasEmittableValue {
			return fmt.Errorf("enum mapping %q has no values the CLI can emit", pointer)
		}
		return nil
	}
	constraints := n.mappingEnumConstraints()
	if len(constraints) > 0 {
		for _, candidate := range constraints[0] {
			if !requestInputDomainAcceptsValue(input, candidate) {
				continue
			}
			if err := n.validate(candidate, ""); err == nil {
				return nil
			}
		}
		return fmt.Errorf("request mapping %q input domain has no values accepted by schema enum constraints", pointer)
	}
	if input.Type != "" && input.Type != "string" {
		return nil
	}
	domains := n.mappingPatternDomains()
	if len(domains) > 0 && !requestPatternDomainsOverlapInput(domains, input) {
		return fmt.Errorf("request mapping %q input domain has no values accepted by schema pattern constraints", pointer)
	}
	return nil
}

func requestInputDomainAcceptsValue(input RequestBodyInput, value any) bool {
	switch input.Type {
	case "", "string":
		candidate, ok := value.(string)
		return ok && requestInputStringAcceptsValue(input, candidate)
	case "enum":
		candidate, ok := value.(string)
		if !ok || !requestInputStringAcceptsValue(input, candidate) {
			return false
		}
		for _, accepted := range input.Values {
			if candidate == accepted {
				return true
			}
		}
		return false
	case "integer":
		return input.Format == "" && typeMatches(value, []string{"integer"})
	case "boolean":
		return input.Format == "" && typeMatches(value, []string{"boolean"})
	case "string_array":
		elements, ok := arrayElements(value)
		if !ok || input.Required && len(elements) == 0 || input.MinItems > 0 && len(elements) < input.MinItems || input.MaxItems > 0 && len(elements) > input.MaxItems {
			return false
		}
		for _, element := range elements {
			candidate, ok := element.(string)
			if !ok || !stringArrayItemEmittable(candidate) {
				return false
			}
		}
		return true
	default:
		return false
	}
}

func requestInputStringAcceptsValue(input RequestBodyInput, value string) bool {
	if !commandInputStringSafe(value) {
		return false
	}
	if input.arrayItem && !stringArrayItemEmittable(value) {
		return false
	}
	empty := CommandInputValueEmpty(value)
	if input.Required && empty {
		return false
	}
	if input.AllowEmpty != nil && empty {
		return *input.AllowEmpty
	}
	return CommandInputFormatAccepts(input.Format, value)
}

func stringArrayItemEmittable(value string) bool {
	if !commandInputStringSafe(value) {
		return false
	}
	items := ParseStringArrayInput([]string{value})
	return len(items) == 1 && items[0] == value
}

func commandInputStringSafe(value string) bool {
	return safety.RejectDangerousChars(value, "flag value") == nil
}

func ParseStringArrayInput(values []string) []string {
	var out []string
	for _, raw := range values {
		for _, item := range strings.Split(raw, ",") {
			item = strings.TrimSpace(item)
			if item != "" {
				out = append(out, item)
			}
		}
	}
	return out
}

func CommandInputValueEmpty(value any) bool {
	switch typed := value.(type) {
	case nil:
		return true
	case string:
		return strings.TrimSpace(typed) == ""
	case []string:
		return len(typed) == 0
	default:
		return false
	}
}

func CommandInputFormatSupported(format string) bool {
	return format == "" || format == "date-time"
}

func CommandInputFormatAccepts(format, value string) bool {
	switch format {
	case "":
		return true
	case "date-time":
		_, err := time.Parse(time.RFC3339, strings.TrimSpace(value))
		return err == nil
	default:
		return false
	}
}

func (n *schemaNode) mappingEnumConstraints() [][]any {
	var out [][]any
	if len(n.enum) > 0 {
		out = append(out, n.enum)
	}
	for _, branch := range n.allOf {
		out = append(out, branch.mappingEnumConstraints()...)
	}
	return out
}

func (n *schemaNode) mappingPatternDomains() []regexpStringDomain {
	var out []regexpStringDomain
	if n.pattern != nil {
		out = append(out, n.patternDomain)
	}
	for _, branch := range n.allOf {
		out = append(out, branch.mappingPatternDomains()...)
	}
	return out
}

func compileRegexpStringDomain(pattern string) (regexpStringDomain, error) {
	program, err := compileRegexpProgram(pattern, true)
	if err != nil {
		return regexpStringDomain{}, err
	}
	return regexpStringDomain{program: program}, nil
}

func compileRegexpProgram(pattern string, search bool) (*syntax.Prog, error) {
	expr, err := syntax.Parse(pattern, syntax.Perl)
	if err != nil {
		return nil, err
	}
	subexpressions := []*syntax.Regexp{{Op: syntax.OpBeginText}}
	if search {
		subexpressions = append(subexpressions, &syntax.Regexp{Op: syntax.OpStar, Sub: []*syntax.Regexp{{Op: syntax.OpAnyChar}}})
	}
	subexpressions = append(subexpressions, expr)
	if search {
		subexpressions = append(subexpressions, &syntax.Regexp{Op: syntax.OpStar, Sub: []*syntax.Regexp{{Op: syntax.OpAnyChar}}})
	}
	subexpressions = append(subexpressions, &syntax.Regexp{Op: syntax.OpEndText})
	return syntax.Compile((&syntax.Regexp{Op: syntax.OpConcat, Sub: subexpressions}).Simplify())
}

type requestPatternSearchState struct {
	programs [][]uint32
	input    requestInputStringState
}

type requestInputStringState struct {
	started       bool
	hasNonSpace   bool
	trailingSpace bool
	previousWord  bool
}

func requestPatternDomainsOverlapInput(domains []regexpStringDomain, input RequestBodyInput) bool {
	programs := make([]*syntax.Prog, 0, len(domains)+1)
	for _, domain := range domains {
		if domain.program == nil {
			return true
		}
		programs = append(programs, domain.program)
	}
	formatProgram, formatKnown := requestInputFormatProgram(input)
	if !formatKnown {
		return false
	}
	if formatProgram != nil {
		programs = append(programs, formatProgram)
	}
	initial := requestPatternSearchState{programs: make([][]uint32, len(programs))}
	for i, program := range programs {
		initial.programs[i] = []uint32{uint32(program.Start)}
	}
	alphabet := regexpProductAlphabet(programs)
	queue := []requestPatternSearchState{initial}
	seen := map[string]bool{requestPatternSearchKey(initial): true}
	const stateLimit = 16384
	for len(queue) > 0 {
		state := queue[0]
		queue = queue[1:]
		if requestPatternSearchAccepts(programs, state, input) {
			return true
		}
		before := regexpBoundaryRune(state.input)
		for _, candidate := range alphabet {
			nextInput, ok := state.input.consume(candidate, input)
			if !ok {
				continue
			}
			next := requestPatternSearchState{programs: make([][]uint32, len(programs)), input: nextInput}
			matched := true
			for i, program := range programs {
				next.programs[i] = regexpProgramStep(program, state.programs[i], before, candidate)
				if len(next.programs[i]) == 0 {
					matched = false
					break
				}
			}
			if !matched {
				continue
			}
			key := requestPatternSearchKey(next)
			if seen[key] {
				continue
			}
			if len(seen) >= stateLimit {
				return true
			}
			seen[key] = true
			queue = append(queue, next)
		}
	}
	return false
}

func requestInputFormatProgram(input RequestBodyInput) (*syntax.Prog, bool) {
	switch input.Format {
	case "":
		return nil, true
	case "date-time":
		const safeSpace = `[ \x{A0}\x{1680}\x{2000}-\x{200A}\x{202F}\x{205F}\x{3000}]`
		core := `[0-9]{4}-[0-9]{2}-[0-9]{2}T[0-9]{1,2}:[0-9]{2}:[0-9]{2}(?:[.,][0-9]+)?(?:Z|[+-][0-9]{2}:[0-9]{2})`
		if !input.Required && input.AllowEmpty != nil && *input.AllowEmpty {
			core = `(?:` + core + `)?`
		}
		program, err := compileRegexpProgram(safeSpace+`*`+core+safeSpace+`*`, false)
		if err != nil {
			return nil, false
		}
		return program, true
	default:
		return nil, false
	}
}

func requestPatternSearchAccepts(programs []*syntax.Prog, state requestPatternSearchState, input RequestBodyInput) bool {
	if !state.input.accepts(input) {
		return false
	}
	before := regexpBoundaryRune(state.input)
	for i, program := range programs {
		if !regexpProgramAccepts(program, state.programs[i], before) {
			return false
		}
	}
	return true
}

func (state requestInputStringState) consume(candidate rune, input RequestBodyInput) (requestInputStringState, bool) {
	if !commandInputStringSafe(string(candidate)) {
		return requestInputStringState{}, false
	}
	space := CommandInputValueEmpty(string(candidate))
	if input.arrayItem && (candidate == ',' || !state.started && space) {
		return requestInputStringState{}, false
	}
	state.started = true
	state.hasNonSpace = state.hasNonSpace || !space
	state.trailingSpace = space
	state.previousWord = syntax.IsWordChar(candidate)
	return state, true
}

func (state requestInputStringState) accepts(input RequestBodyInput) bool {
	if input.arrayItem {
		return state.started && state.hasNonSpace && !state.trailingSpace
	}
	empty := !state.hasNonSpace
	if input.Required && empty {
		return false
	}
	if input.AllowEmpty != nil && empty {
		return *input.AllowEmpty
	}
	return true
}

func regexpBoundaryRune(state requestInputStringState) rune {
	if !state.started {
		return -1
	}
	if state.previousWord {
		return 'a'
	}
	return '-'
}

func regexpProgramStep(program *syntax.Prog, raw []uint32, before, candidate rune) []uint32 {
	closure := regexpProgramClosure(program, raw, before, candidate)
	next := map[uint32]bool{}
	for _, pc := range closure {
		instruction := program.Inst[pc]
		if regexpInstructionMatchesRune(instruction, candidate) {
			next[instruction.Out] = true
		}
	}
	return sortedProgramCounters(next)
}

func regexpProgramAccepts(program *syntax.Prog, raw []uint32, before rune) bool {
	for _, pc := range regexpProgramClosure(program, raw, before, -1) {
		if program.Inst[pc].Op == syntax.InstMatch {
			return true
		}
	}
	return false
}

func regexpProgramClosure(program *syntax.Prog, raw []uint32, before, after rune) []uint32 {
	stack := append([]uint32(nil), raw...)
	seen := map[uint32]bool{}
	leaves := map[uint32]bool{}
	for len(stack) > 0 {
		pc := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		if seen[pc] || int(pc) >= len(program.Inst) {
			continue
		}
		seen[pc] = true
		instruction := program.Inst[pc]
		switch instruction.Op {
		case syntax.InstAlt, syntax.InstAltMatch:
			stack = append(stack, instruction.Out, instruction.Arg)
		case syntax.InstCapture, syntax.InstNop:
			stack = append(stack, instruction.Out)
		case syntax.InstEmptyWidth:
			if instruction.MatchEmptyWidth(before, after) {
				stack = append(stack, instruction.Out)
			}
		case syntax.InstRune, syntax.InstRune1, syntax.InstRuneAny, syntax.InstRuneAnyNotNL, syntax.InstMatch:
			leaves[pc] = true
		}
	}
	return sortedProgramCounters(leaves)
}

func regexpInstructionMatchesRune(instruction syntax.Inst, candidate rune) bool {
	switch instruction.Op {
	case syntax.InstRune, syntax.InstRune1:
		return instruction.MatchRune(candidate)
	case syntax.InstRuneAny:
		return true
	case syntax.InstRuneAnyNotNL:
		return candidate != '\n'
	default:
		return false
	}
}

func sortedProgramCounters(counters map[uint32]bool) []uint32 {
	out := make([]uint32, 0, len(counters))
	for counter := range counters {
		out = append(out, counter)
	}
	sort.Slice(out, func(i, j int) bool { return out[i] < out[j] })
	return out
}

func requestPatternSearchKey(state requestPatternSearchState) string {
	var key strings.Builder
	for _, counters := range state.programs {
		key.WriteByte('|')
		for _, counter := range counters {
			key.WriteString(strconv.FormatUint(uint64(counter), 10))
			key.WriteByte(',')
		}
	}
	key.WriteByte('|')
	for _, value := range []bool{state.input.started, state.input.hasNonSpace, state.input.trailingSpace, state.input.previousWord} {
		if value {
			key.WriteByte('1')
		} else {
			key.WriteByte('0')
		}
	}
	return key.String()
}

func regexpProductAlphabet(programs []*syntax.Prog) []rune {
	boundaries := map[int]bool{0: true, utf8.MaxRune + 1: true, 0xD800: true, 0xE000: true}
	addRange := func(first, last rune) {
		if first < 0 {
			first = 0
		}
		if last > utf8.MaxRune {
			last = utf8.MaxRune
		}
		if first > last {
			return
		}
		boundaries[int(first)] = true
		boundaries[int(last)+1] = true
	}
	for candidate := rune(0); candidate < 128; candidate++ {
		addRange(candidate, candidate)
	}
	for _, table := range []*unicode.RangeTable{unicode.White_Space} {
		for _, item := range table.R16 {
			for candidate := rune(item.Lo); candidate <= rune(item.Hi); candidate += rune(item.Stride) {
				addRange(candidate, candidate)
			}
		}
		for _, item := range table.R32 {
			for candidate := rune(item.Lo); candidate <= rune(item.Hi); candidate += rune(item.Stride) {
				addRange(candidate, candidate)
			}
		}
	}
	for _, span := range [][2]rune{{0, 0x1F}, {0x7F, 0x9F}, {0x200B, 0x200D}, {0x2028, 0x202E}, {0x2066, 0x2069}, {0xFEFF, 0xFEFF}} {
		addRange(span[0], span[1])
	}
	for _, program := range programs {
		for _, instruction := range program.Inst {
			switch instruction.Op {
			case syntax.InstRune, syntax.InstRune1:
				if len(instruction.Rune) == 1 {
					candidate := instruction.Rune[0]
					addRange(candidate, candidate)
					if syntax.Flags(instruction.Arg)&syntax.FoldCase != 0 {
						for folded := unicode.SimpleFold(candidate); folded != candidate; folded = unicode.SimpleFold(folded) {
							addRange(folded, folded)
						}
					}
					continue
				}
				for i := 0; i+1 < len(instruction.Rune); i += 2 {
					addRange(instruction.Rune[i], instruction.Rune[i+1])
				}
			}
		}
	}
	ordered := make([]int, 0, len(boundaries))
	for boundary := range boundaries {
		ordered = append(ordered, boundary)
	}
	sort.Ints(ordered)
	alphabet := make([]rune, 0, len(ordered))
	for i := 0; i+1 < len(ordered); i++ {
		candidate := rune(ordered[i])
		if candidate >= 0xD800 && candidate < 0xE000 {
			candidate = 0xE000
		}
		if int(candidate) >= ordered[i+1] || !utf8.ValidRune(candidate) || !commandInputStringSafe(string(candidate)) {
			continue
		}
		alphabet = append(alphabet, candidate)
	}
	return alphabet
}

func (n *schemaNode) validateRequestInputCardinality(pointer string, input RequestBodyInput) error {
	inputMin := input.MinItems
	inputMax := input.MaxItems
	if inputMin < 0 || inputMax < 0 || inputMax > 0 && inputMin > inputMax {
		return fmt.Errorf("string_array mapping %q has invalid input cardinality %d..%d", pointer, inputMin, inputMax)
	}
	if input.Required && inputMin < 1 {
		inputMin = 1
	}
	lower := inputMin
	upper := inputMax
	hasUpper := inputMax > 0
	schemaMin, hasSchemaMin, schemaMax, hasSchemaMax := n.mappingArrayCardinality()
	if hasSchemaMin && schemaMin > lower {
		lower = schemaMin
	}
	if hasSchemaMax && (!hasUpper || schemaMax < upper) {
		upper = schemaMax
		hasUpper = true
	}
	if hasUpper && lower > upper {
		return fmt.Errorf("string_array mapping %q has no cardinality accepted by both CLI bounds and schema bounds", pointer)
	}
	return nil
}

func (n *schemaNode) mappingArrayCardinality() (int, bool, int, bool) {
	minimum, hasMinimum := n.minItems, n.hasMinItems
	maximum, hasMaximum := n.maxItems, n.hasMaxItems
	for _, branch := range n.allOf {
		branchMin, hasBranchMin, branchMax, hasBranchMax := branch.mappingArrayCardinality()
		if hasBranchMin && (!hasMinimum || branchMin > minimum) {
			minimum, hasMinimum = branchMin, true
		}
		if hasBranchMax && (!hasMaximum || branchMax < maximum) {
			maximum, hasMaximum = branchMax, true
		}
	}
	return minimum, hasMinimum, maximum, hasMaximum
}

func (s *Schema) ValidateRequestBodyPointerAssignments(required, optional []string) error {
	all := append(append([]string(nil), required...), optional...)
	if err := ValidateRequestFieldPointerAssignments(all); err != nil {
		return err
	}
	if _, err := s.assembleDynamicRequestBody(required); err != nil {
		return err
	}
	for _, pointer := range optional {
		pointers := append(append([]string(nil), required...), pointer)
		if _, err := s.assembleDynamicRequestBody(pointers); err != nil {
			return err
		}
	}
	if _, err := s.assembleDynamicRequestBody(all); err != nil {
		return err
	}
	return nil
}

func (s *Schema) ValidateEffectiveRequestBody(static map[string]any, required, optional []string) error {
	if err := s.ValidateRequestBodyPointerAssignments(required, optional); err != nil {
		return err
	}
	sets := make([][]string, 0, len(optional)+2)
	sets = append(sets, required)
	for _, pointer := range optional {
		sets = append(sets, append(append([]string(nil), required...), pointer))
	}
	sets = append(sets, append(append([]string(nil), required...), optional...))
	for _, pointers := range sets {
		dynamic, err := s.assembleDynamicRequestBody(pointers)
		if err != nil {
			return err
		}
		body := MergeRequestBody(static, dynamic)
		if err := s.node.validateWithDynamic(body, "", true); err != nil {
			return err
		}
	}
	return nil
}

func (s *Schema) assembleDynamicRequestBody(pointers []string) (map[string]any, error) {
	ordered := append([]string(nil), pointers...)
	sort.Slice(ordered, func(i, j int) bool {
		return requestFieldPointerLess(ordered[i], ordered[j])
	})
	body := map[string]any{}
	for _, pointer := range ordered {
		namespace, _, err := parseRequestFieldPointer(pointer)
		if err != nil {
			return nil, err
		}
		if namespace != "body" {
			return nil, fmt.Errorf("request mapping %q is not body-scoped", pointer)
		}
		if err := s.SetRequestBodyPointer(body, pointer, dynamicRequestBodyValue{}); err != nil {
			return nil, err
		}
	}
	return body, nil
}

func requestFieldPointerLess(left, right string) bool {
	leftNamespace, leftTokens, leftErr := parseRequestFieldPointer(left)
	rightNamespace, rightTokens, rightErr := parseRequestFieldPointer(right)
	if leftErr != nil || rightErr != nil {
		return left < right
	}
	if leftNamespace != rightNamespace {
		return leftNamespace < rightNamespace
	}
	limit := len(leftTokens)
	if len(rightTokens) < limit {
		limit = len(rightTokens)
	}
	for i := 0; i < limit; i++ {
		leftIndex, leftErr := strconv.Atoi(leftTokens[i])
		rightIndex, rightErr := strconv.Atoi(rightTokens[i])
		if leftErr == nil && rightErr == nil && leftIndex != rightIndex {
			return leftIndex < rightIndex
		}
		if leftTokens[i] != rightTokens[i] {
			return leftTokens[i] < rightTokens[i]
		}
	}
	return len(leftTokens) < len(rightTokens)
}

func (s *Schema) SetRequestBodyPointer(body map[string]any, pointer string, value any) error {
	if s == nil || s.node == nil {
		return fmt.Errorf("schema is nil")
	}
	namespace, tokens, err := parseRequestFieldPointer(pointer)
	if err != nil {
		return err
	}
	if namespace != "body" {
		return fmt.Errorf("request field pointer %q is not body-scoped", pointer)
	}
	updated, err := s.node.setRequestPointerValue(body, tokens, value, pointer)
	if err != nil {
		return err
	}
	if _, ok := updated.(map[string]any); !ok {
		return fmt.Errorf("request body schema root must be an object")
	}
	return nil
}

func (n *schemaNode) setRequestPointerValue(current any, tokens []string, value any, pointer string) (any, error) {
	if len(tokens) == 0 {
		return value, nil
	}
	if err := n.mappingConstraintError(pointer); err != nil {
		return nil, err
	}
	isArray := n.mappingIsArray()
	isObject := n.mappingIsObject()
	if isArray && isObject {
		return nil, fmt.Errorf("request field %q traverses an ambiguous object-or-array schema", pointer)
	}
	if isArray {
		itemsSchema, ok := n.mappingItems()
		if !ok {
			return nil, fmt.Errorf("request field %q traverses an array without an item schema", pointer)
		}
		if err := validateMappingArrayIndex(tokens[0]); err != nil {
			return nil, err
		}
		index, err := strconv.Atoi(tokens[0])
		if err != nil {
			return nil, fmt.Errorf("array segment %q is not a numeric index", tokens[0])
		}
		var items []any
		if current != nil {
			var ok bool
			items, ok = current.([]any)
			if !ok {
				return nil, fmt.Errorf("request field %q conflicts with existing non-array value", pointer)
			}
		}
		if index > len(items) {
			return nil, fmt.Errorf("request field %q uses sparse array index %d", pointer, index)
		}
		if index == len(items) {
			items = append(items, nil)
		}
		updated, err := itemsSchema.setRequestPointerValue(items[index], tokens[1:], value, pointer)
		if err != nil {
			return nil, err
		}
		items[index] = updated
		return items, nil
	}
	if !isObject {
		return nil, fmt.Errorf("request field %q traverses a non-container schema", pointer)
	}
	var object map[string]any
	if current == nil {
		object = map[string]any{}
	} else {
		var ok bool
		object, ok = current.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("request field %q conflicts with existing non-object value", pointer)
		}
	}
	child, ok := n.mappingProperty(tokens[0])
	if !ok {
		return nil, fmt.Errorf("record field %q is not declared", tokens[0])
	}
	updated, err := child.setRequestPointerValue(object[tokens[0]], tokens[1:], value, pointer)
	if err != nil {
		return nil, err
	}
	object[tokens[0]] = updated
	return object, nil
}

func (s *Schema) ValidateRequestFieldPointerValue(pointer string, value any) error {
	if s == nil || s.node == nil {
		return fmt.Errorf("schema is nil")
	}
	namespace, tokens, err := parseRequestFieldPointer(pointer)
	if err != nil {
		return err
	}
	if namespace != "body" {
		return fmt.Errorf("request field pointer %q is not body-scoped", pointer)
	}
	node, _, err := s.requestMappingNode(tokens, pointer)
	if err != nil {
		return err
	}
	return node.validate(value, pointer)
}

func (s *Schema) ValidatePartialRequestBody(body map[string]any) error {
	if s == nil || s.node == nil {
		return fmt.Errorf("schema is nil")
	}
	return s.node.validatePartial(body, "")
}

func (s *Schema) MappingPath(path string) (SchemaPathInfo, error) {
	if s == nil || s.node == nil {
		return SchemaPathInfo{}, fmt.Errorf("schema is nil")
	}
	if strings.TrimSpace(path) == "" {
		return s.node.mappingPathInfo()
	}
	current := s.node
	for _, part := range strings.Split(path, ".") {
		if part == "" {
			return SchemaPathInfo{}, fmt.Errorf("empty path segment")
		}
		if err := current.mappingConstraintError(path); err != nil {
			return SchemaPathInfo{}, err
		}
		if current.mappingIsArray() {
			if err := validateMappingArrayIndex(part); err != nil {
				return SchemaPathInfo{}, err
			}
			items, ok := current.mappingItems()
			if !ok {
				return SchemaPathInfo{}, fmt.Errorf("array segment %q has no item schema", part)
			}
			current = items
			continue
		}
		child, ok := current.mappingProperty(part)
		if !ok {
			return SchemaPathInfo{}, fmt.Errorf("record field %q is not declared", part)
		}
		current = child
	}
	return current.mappingPathInfo()
}

func (n *schemaNode) rejectAlternativeMappings(path string) error {
	if err := n.mappingConstraintError(path); err != nil {
		return err
	}
	if len(n.anyOf) > 0 {
		return fmt.Errorf("schema %s uses anyOf alternatives that required CLI flags cannot express", displayMappingPath(path))
	}
	if len(n.oneOf) > 0 {
		return fmt.Errorf("schema %s uses oneOf alternatives that required CLI flags cannot express", displayMappingPath(path))
	}
	names := make([]string, 0, len(n.properties))
	for name := range n.properties {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		if err := n.properties[name].rejectAlternativeMappings(joinMappingPath(path, name)); err != nil {
			return err
		}
	}
	if n.items != nil {
		if err := n.items.rejectAlternativeMappings(joinMappingPath(path, "0")); err != nil {
			return err
		}
	}
	for _, branch := range n.allOf {
		if err := branch.rejectAlternativeMappings(path); err != nil {
			return err
		}
	}
	return nil
}

func (n *schemaNode) rejectAlternativeRequestMappings(pointer string) error {
	if err := n.mappingConstraintError(pointer); err != nil {
		return err
	}
	if len(n.anyOf) > 0 {
		return fmt.Errorf("schema %s uses anyOf alternatives that required CLI flags cannot express", pointer)
	}
	if len(n.oneOf) > 0 {
		return fmt.Errorf("schema %s uses oneOf alternatives that required CLI flags cannot express", pointer)
	}
	names := make([]string, 0, len(n.properties))
	for name := range n.properties {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		if err := n.properties[name].rejectAlternativeRequestMappings(appendRequestFieldPointer(pointer, name)); err != nil {
			return err
		}
	}
	if n.items != nil {
		if err := n.items.rejectAlternativeRequestMappings(appendRequestFieldPointer(pointer, "0")); err != nil {
			return err
		}
	}
	for _, branch := range n.allOf {
		if err := branch.rejectAlternativeRequestMappings(pointer); err != nil {
			return err
		}
	}
	return nil
}

func (n *schemaNode) collectRequiredMappingPaths(prefix string, paths map[string]bool) error {
	return n.collectRequiredMappingPathsWithScope(prefix, paths, n)
}

func (n *schemaNode) collectRequiredMappingPathsWithScope(prefix string, paths map[string]bool, scope *schemaNode) error {
	for _, name := range n.required {
		path := joinMappingPath(prefix, name)
		child, ok := scope.mappingProperty(name)
		if !ok {
			return fmt.Errorf("schema required property %q is not declared", path)
		}
		nested := map[string]bool{}
		if err := child.collectRequiredNodeMappingPaths(path, nested); err != nil {
			return err
		}
		if len(nested) == 0 {
			paths[path] = true
			continue
		}
		for nestedPath := range nested {
			paths[nestedPath] = true
		}
	}
	for _, branch := range n.allOf {
		if err := branch.collectRequiredMappingPathsWithScope(prefix, paths, scope); err != nil {
			return err
		}
	}
	return nil
}

func (n *schemaNode) collectRequiredRequestMappings(prefix string, paths map[string]bool, scope *schemaNode) error {
	for _, name := range n.required {
		path := appendRequestFieldPointer(prefix, name)
		child, ok := scope.mappingProperty(name)
		if !ok {
			return fmt.Errorf("schema required property %q is not declared", path)
		}
		nested := map[string]bool{}
		if err := child.collectRequiredRequestNodeMappings(path, nested); err != nil {
			return err
		}
		if len(nested) == 0 {
			paths[path] = true
			continue
		}
		for nestedPath := range nested {
			paths[nestedPath] = true
		}
	}
	for _, branch := range n.allOf {
		if err := branch.collectRequiredRequestMappings(prefix, paths, scope); err != nil {
			return err
		}
	}
	return nil
}

func (n *schemaNode) collectRequiredRequestNodeMappings(prefix string, paths map[string]bool) error {
	if n.mappingIsArray() {
		items, ok := n.mappingItems()
		if !ok {
			return nil
		}
		return items.collectRequiredRequestNodeMappings(appendRequestFieldPointer(prefix, "0"), paths)
	}
	if !n.mappingIsObject() {
		return nil
	}
	return n.collectRequiredRequestMappings(prefix, paths, n)
}

func (n *schemaNode) collectRequiredNodeMappingPaths(prefix string, paths map[string]bool) error {
	if n.mappingIsArray() {
		items, ok := n.mappingItems()
		if !ok {
			return nil
		}
		return items.collectRequiredNodeMappingPaths(joinMappingPath(prefix, "0"), paths)
	}
	if !n.mappingIsObject() {
		return nil
	}
	return n.collectRequiredMappingPaths(prefix, paths)
}

func (n *schemaNode) mappingProperty(name string) (*schemaNode, bool) {
	var matches []*schemaNode
	if child := n.properties[name]; child != nil {
		matches = append(matches, child)
	}
	for _, branch := range n.allOf {
		if child, ok := branch.mappingProperty(name); ok {
			matches = append(matches, child)
		}
	}
	if len(matches) == 0 {
		return nil, false
	}
	if len(matches) == 1 {
		return matches[0], true
	}
	return &schemaNode{allOf: matches, additionalProperties: true}, true
}

func (n *schemaNode) mappingItems() (*schemaNode, bool) {
	var matches []*schemaNode
	if n.items != nil {
		matches = append(matches, n.items)
	}
	for _, branch := range n.allOf {
		if items, ok := branch.mappingItems(); ok {
			matches = append(matches, items)
		}
	}
	if len(matches) == 0 {
		return nil, false
	}
	if len(matches) == 1 {
		return matches[0], true
	}
	return &schemaNode{allOf: matches, additionalProperties: true}, true
}

func (n *schemaNode) mappingPathInfo() (SchemaPathInfo, error) {
	typeSet := n.mappingTypeSet()
	if typeSet.unsatisfiable {
		return SchemaPathInfo{}, fmt.Errorf("schema has unsatisfiable allOf type constraints")
	}
	types := typeSet.types
	if !typeSet.constrained {
		switch {
		case n.mappingIsArray():
			types = []string{"array"}
		case n.mappingIsObject():
			types = []string{"object"}
		default:
			types = []string{"any"}
		}
	}
	return SchemaPathInfo{Types: types, IsObject: containsSchemaType(types, "object") || n.mappingHasProperties(), IsArray: containsSchemaType(types, "array") || n.items != nil}, nil
}

type schemaMappingTypeSet struct {
	types         []string
	constrained   bool
	unsatisfiable bool
}

func (n *schemaNode) mappingTypeSet() schemaMappingTypeSet {
	var constrained map[string]bool
	unsatisfiable := false
	apply := func(typeSet schemaMappingTypeSet) {
		if typeSet.unsatisfiable {
			unsatisfiable = true
			return
		}
		if !typeSet.constrained {
			return
		}
		candidate := make(map[string]bool, len(typeSet.types))
		for _, typ := range typeSet.types {
			candidate[typ] = true
		}
		if constrained == nil {
			constrained = candidate
			return
		}
		constrained = intersectSchemaMappingTypes(constrained, candidate)
		if len(constrained) == 0 {
			unsatisfiable = true
		}
	}
	apply(schemaMappingTypeSet{types: n.types, constrained: len(n.types) > 0})
	for _, branch := range n.allOf {
		apply(branch.mappingTypeSet())
	}
	out := make([]string, 0, len(constrained))
	for typ := range constrained {
		out = append(out, typ)
	}
	sort.Strings(out)
	return schemaMappingTypeSet{types: out, constrained: constrained != nil, unsatisfiable: unsatisfiable}
}

func intersectSchemaMappingTypes(left, right map[string]bool) map[string]bool {
	out := make(map[string]bool)
	for typ := range left {
		if right[typ] {
			out[typ] = true
		}
	}
	if left["number"] && right["integer"] || left["integer"] && right["number"] {
		out["integer"] = true
	}
	return out
}

func (n *schemaNode) mappingConstraintError(path string) error {
	if !n.mappingTypeSet().unsatisfiable {
		return nil
	}
	return fmt.Errorf("schema %s has unsatisfiable allOf type constraints", displayMappingPath(path))
}

func (n *schemaNode) mappingHasProperties() bool {
	if len(n.properties) > 0 {
		return true
	}
	for _, branch := range n.allOf {
		if branch.mappingHasProperties() {
			return true
		}
	}
	return false
}

func (n *schemaNode) mappingIsObject() bool {
	return containsSchemaType(n.mappingTypeSet().types, "object") || n.mappingHasProperties()
}

func (n *schemaNode) mappingIsArray() bool {
	_, hasItems := n.mappingItems()
	return containsSchemaType(n.mappingTypeSet().types, "array") || hasItems
}

func (s *Schema) canonicalRequestMappingTokens(tokens []string) ([]string, error) {
	if s == nil || s.node == nil {
		return nil, fmt.Errorf("schema is nil")
	}
	_, canonical, err := s.requestMappingNode(tokens, requestFieldPointer("body", tokens...))
	return canonical, err
}

func (s *Schema) requestMappingNode(tokens []string, pointer string) (*schemaNode, []string, error) {
	current := s.node
	canonical := make([]string, 0, len(tokens))
	for _, part := range tokens {
		if err := current.mappingConstraintError(pointer); err != nil {
			return nil, nil, err
		}
		if current.mappingIsArray() {
			if err := validateMappingArrayIndex(part); err != nil {
				return nil, nil, err
			}
			items, ok := current.mappingItems()
			if !ok {
				return nil, nil, fmt.Errorf("array segment %q has no item schema", part)
			}
			canonical = append(canonical, "0")
			current = items
			continue
		}
		child, ok := current.mappingProperty(part)
		if !ok {
			return nil, nil, fmt.Errorf("record field %q is not declared", part)
		}
		canonical = append(canonical, part)
		current = child
	}
	if err := current.mappingConstraintError(pointer); err != nil {
		return nil, nil, err
	}
	return current, canonical, nil
}

func containsSchemaType(types []string, target string) bool {
	for _, typ := range types {
		if typ == target {
			return true
		}
	}
	return false
}

func validateMappingArrayIndex(part string) error {
	if part == "" {
		return fmt.Errorf("array segment %q is not a numeric index", part)
	}
	for _, r := range part {
		if r < '0' || r > '9' {
			return fmt.Errorf("array segment %q is not a numeric index", part)
		}
	}
	if len(part) > 1 && strings.HasPrefix(part, "0") {
		return fmt.Errorf("array index %q must not have leading zeroes", part)
	}
	if len(part) > 3 || len(part) == 3 && part > "128" {
		return fmt.Errorf("array index %q exceeds max %d", part, 128)
	}
	return nil
}

func joinMappingPath(prefix, part string) string {
	if prefix == "" {
		return part
	}
	return prefix + "." + part
}

func displayMappingPath(path string) string {
	if path == "" {
		return "/"
	}
	return path
}

func compileNode(m map[string]json.RawMessage) (*schemaNode, error) {
	for k := range m {
		if annotationKeywords[k] || structuralKeywords[k] {
			continue
		}
		return nil, fmt.Errorf("compile schema: unknown keyword %q", k)
	}

	n := &schemaNode{additionalProperties: true}

	if raw, ok := m["type"]; ok {
		types, err := compileTypes(raw)
		if err != nil {
			return nil, err
		}
		n.types = types
	}

	if raw, ok := m["required"]; ok {
		var req []string
		if err := json.Unmarshal(raw, &req); err != nil {
			return nil, fmt.Errorf("compile schema: required: %w", err)
		}
		n.required = req
	}

	if raw, ok := m["properties"]; ok {
		var props map[string]map[string]json.RawMessage
		if err := json.Unmarshal(raw, &props); err != nil {
			return nil, fmt.Errorf("compile schema: properties: %w", err)
		}
		n.properties = make(map[string]*schemaNode, len(props))
		for name, sub := range props {
			child, err := compileNode(sub)
			if err != nil {
				return nil, fmt.Errorf("compile schema: properties.%s: %w", name, err)
			}
			n.properties[name] = child
		}
	}

	if raw, ok := m["items"]; ok {
		var sub map[string]json.RawMessage
		if err := json.Unmarshal(raw, &sub); err != nil {
			return nil, fmt.Errorf("compile schema: items: %w", err)
		}
		child, err := compileNode(sub)
		if err != nil {
			return nil, fmt.Errorf("compile schema: items: %w", err)
		}
		n.items = child
	}

	var err error
	if n.allOf, err = compileSchemaBranches(m, "allOf"); err != nil {
		return nil, err
	}
	if n.anyOf, err = compileSchemaBranches(m, "anyOf"); err != nil {
		return nil, err
	}
	if n.oneOf, err = compileSchemaBranches(m, "oneOf"); err != nil {
		return nil, err
	}

	if raw, ok := m["enum"]; ok {
		var vals []any
		dec := json.NewDecoder(strings.NewReader(string(raw)))
		dec.UseNumber()
		if err := dec.Decode(&vals); err != nil {
			return nil, fmt.Errorf("compile schema: enum: %w", err)
		}
		n.enum = vals
	}

	if raw, ok := m["pattern"]; ok {
		var pat string
		if err := json.Unmarshal(raw, &pat); err != nil {
			return nil, fmt.Errorf("compile schema: pattern: %w", err)
		}
		re, err := regexp.Compile(pat)
		if err != nil {
			return nil, fmt.Errorf("compile schema: pattern %q: %w", pat, err)
		}
		domain, err := compileRegexpStringDomain(pat)
		if err != nil {
			return nil, fmt.Errorf("compile schema: pattern %q: %w", pat, err)
		}
		n.pattern = re
		n.patternDomain = domain
	}

	if raw, ok := m["minProperties"]; ok {
		var mp int
		if err := json.Unmarshal(raw, &mp); err != nil {
			return nil, fmt.Errorf("compile schema: minProperties: %w", err)
		}
		n.minProperties = mp
		n.hasMinProperties = true
	}

	if err := compileArrayCardinality(m, n); err != nil {
		return nil, err
	}

	if raw, ok := m["default"]; ok {
		var def any
		dec := json.NewDecoder(strings.NewReader(string(raw)))
		dec.UseNumber()
		if err := dec.Decode(&def); err != nil {
			return nil, fmt.Errorf("compile schema: default: %w", err)
		}
		n.defaultVal = def
		n.hasDefault = true
	}

	if raw, ok := m["additionalProperties"]; ok {
		var ap bool
		if err := json.Unmarshal(raw, &ap); err != nil {
			return nil, fmt.Errorf("compile schema: additionalProperties: only bool form supported: %w", err)
		}
		n.additionalProperties = ap
		n.hasAdditionalProps = true
	}

	if raw, ok := m["x-secret"]; ok {
		var secret bool
		if err := json.Unmarshal(raw, &secret); err != nil {
			return nil, fmt.Errorf("compile schema: x-secret: %w", err)
		}
		n.secret = secret
	}

	if raw, ok := m["x-primary-key"]; ok {
		var pk []string
		if err := json.Unmarshal(raw, &pk); err != nil {
			return nil, fmt.Errorf("compile schema: x-primary-key: %w", err)
		}
		n.primaryKey = pk
	}

	if raw, ok := m["x-cursor-field"]; ok {
		var cf string
		if err := json.Unmarshal(raw, &cf); err != nil {
			return nil, fmt.Errorf("compile schema: x-cursor-field: %w", err)
		}
		n.cursorField = cf
	}
	if err := n.validateClosedAllOfCompatibility(); err != nil {
		return nil, err
	}

	return n, nil
}

func (n *schemaNode) validateClosedAllOfCompatibility() error {
	sets := n.closedAllOfPropertySets(nil)
	if len(sets) < 2 {
		return nil
	}
	first := sets[0]
	for _, candidate := range sets[1:] {
		if equalStringSets(first, candidate) {
			continue
		}
		return fmt.Errorf("compile schema: incompatible closed-object allOf property sets %v and %v", sortedStringSet(first), sortedStringSet(candidate))
	}
	return nil
}

func (n *schemaNode) closedAllOfPropertySets(sets []map[string]bool) []map[string]bool {
	if n.hasAdditionalProps && !n.additionalProperties {
		properties := make(map[string]bool, len(n.properties))
		for name := range n.properties {
			properties[name] = true
		}
		sets = append(sets, properties)
	}
	for _, branch := range n.allOf {
		sets = branch.closedAllOfPropertySets(sets)
	}
	return sets
}

func equalStringSets(left, right map[string]bool) bool {
	if len(left) != len(right) {
		return false
	}
	for value := range left {
		if !right[value] {
			return false
		}
	}
	return true
}

func sortedStringSet(values map[string]bool) []string {
	out := make([]string, 0, len(values))
	for value := range values {
		out = append(out, value)
	}
	sort.Strings(out)
	return out
}

func compileSchemaBranches(m map[string]json.RawMessage, keyword string) ([]*schemaNode, error) {
	raw, ok := m[keyword]
	if !ok {
		return nil, nil
	}
	var branches []map[string]json.RawMessage
	if err := json.Unmarshal(raw, &branches); err != nil {
		return nil, fmt.Errorf("compile schema: %s: %w", keyword, err)
	}
	if len(branches) == 0 {
		return nil, fmt.Errorf("compile schema: %s must be non-empty", keyword)
	}
	compiled := make([]*schemaNode, 0, len(branches))
	for i, branch := range branches {
		node, err := compileNode(branch)
		if err != nil {
			return nil, fmt.Errorf("compile schema: %s[%d]: %w", keyword, i, err)
		}
		compiled = append(compiled, node)
	}
	return compiled, nil
}

// compileArrayCardinality compiles the draft-07 minItems/maxItems pair.
//
// The engine dialect deliberately had no array-cardinality keyword, which made
// "this documented request array must not be empty" inexpressible: a bundle
// could declare an array field required, but not that it carries at least one
// element, so an operation whose provider rejects an empty array could not be
// declared executable without risking a malformed request. Two bundles already
// document the gap in prose (defs/drip/writes.json, defs/zoho-bigin/writes.json)
// and Airtable's operation ledger blocks 25 endpoints on it by name.
//
// Both bounds must be non-negative integers (draft-07), and a declared maxItems
// below a declared minItems is unsatisfiable — always an authoring mistake, so
// it fails at compile time rather than rejecting every instance at runtime.
func compileArrayCardinality(m map[string]json.RawMessage, n *schemaNode) error {
	if raw, ok := m["minItems"]; ok {
		v, err := compileNonNegativeInt(raw, "minItems")
		if err != nil {
			return err
		}
		n.minItems = v
		n.hasMinItems = true
	}
	if raw, ok := m["maxItems"]; ok {
		v, err := compileNonNegativeInt(raw, "maxItems")
		if err != nil {
			return err
		}
		n.maxItems = v
		n.hasMaxItems = true
	}
	if n.hasMinItems && n.hasMaxItems && n.maxItems < n.minItems {
		return fmt.Errorf("compile schema: maxItems %d is below minItems %d", n.maxItems, n.minItems)
	}
	return nil
}

func compileNonNegativeInt(raw json.RawMessage, keyword string) (int, error) {
	var v int
	if err := json.Unmarshal(raw, &v); err != nil {
		return 0, fmt.Errorf("compile schema: %s: %w", keyword, err)
	}
	if v < 0 {
		return 0, fmt.Errorf("compile schema: %s must be non-negative, got %d", keyword, v)
	}
	return v, nil
}

func compileTypes(raw json.RawMessage) ([]string, error) {
	var single string
	if err := json.Unmarshal(raw, &single); err == nil {
		if !validTypes[single] {
			return nil, fmt.Errorf("compile schema: unknown type %q", single)
		}
		return []string{single}, nil
	}
	var multi []string
	if err := json.Unmarshal(raw, &multi); err != nil {
		return nil, fmt.Errorf("compile schema: type: %w", err)
	}
	for _, t := range multi {
		if !validTypes[t] {
			return nil, fmt.Errorf("compile schema: unknown type %q", t)
		}
	}
	return multi, nil
}

// Validate checks v (already decoded via encoding/json, ideally with
// UseNumber for integer fidelity) against the compiled schema. Errors name a
// JSON-pointer-ish path to the offending value.
func (s *Schema) Validate(v any) error {
	return s.node.validate(v, "")
}

func (n *schemaNode) validatePartial(v any, path string) error {
	if len(n.types) > 0 && !typeMatches(v, n.types) {
		return fmt.Errorf("%s: value does not match type %v", displayPath(path), n.types)
	}
	if len(n.enum) > 0 {
		matched := false
		for _, want := range n.enum {
			if enumEquals(v, want) {
				matched = true
				break
			}
		}
		if !matched {
			return fmt.Errorf("%s: value not in enum %v", displayPath(path), n.enum)
		}
	}
	if val, ok := v.(string); ok && n.pattern != nil && !n.pattern.MatchString(val) {
		return fmt.Errorf("%s: value does not match pattern %q", displayPath(path), n.pattern.String())
	}
	if obj, ok := v.(map[string]any); ok {
		if n.hasAdditionalProps && !n.additionalProperties {
			keys := make([]string, 0, len(obj))
			for key := range obj {
				keys = append(keys, key)
			}
			sort.Strings(keys)
			for _, key := range keys {
				if _, declared := n.properties[key]; !declared {
					return fmt.Errorf("%s/%s: additional property not allowed", displayPath(path), key)
				}
			}
		}
		for key, child := range n.properties {
			value, present := obj[key]
			if !present {
				continue
			}
			if err := child.validatePartial(value, path+"/"+key); err != nil {
				return err
			}
		}
	}
	if elems, ok := arrayElements(v); ok {
		if err := n.validateArrayCardinality(len(elems), path); err != nil {
			return err
		}
		if n.items != nil {
			for i, elem := range elems {
				if err := n.items.validate(elem, fmt.Sprintf("%s/%d", path, i)); err != nil {
					return err
				}
			}
		}
	}
	for i, branch := range n.allOf {
		if err := branch.validatePartial(v, path); err != nil {
			return fmt.Errorf("%s: allOf[%d]: %w", displayPath(path), i, err)
		}
	}
	if len(n.anyOf) > 0 || len(n.oneOf) > 0 {
		return fmt.Errorf("%s: partial validation cannot evaluate alternative schema branches", displayPath(path))
	}
	return nil
}

func (n *schemaNode) validate(v any, path string) error {
	return n.validateWithDynamic(v, path, false)
}

func (n *schemaNode) validateWithDynamic(v any, path string, allowDynamic bool) error {
	if allowDynamic {
		if _, ok := v.(dynamicRequestBodyValue); ok {
			return nil
		}
	}
	if len(n.types) > 0 && !typeMatches(v, n.types) {
		return fmt.Errorf("%s: value does not match type %v", displayPath(path), n.types)
	}

	if len(n.enum) > 0 {
		matched := false
		for _, want := range n.enum {
			if enumEquals(v, want) {
				matched = true
				break
			}
		}
		if !matched {
			return fmt.Errorf("%s: value not in enum %v", displayPath(path), n.enum)
		}
	}

	switch val := v.(type) {
	case string:
		if n.pattern != nil && !n.pattern.MatchString(val) {
			return fmt.Errorf("%s: value does not match pattern %q", displayPath(path), n.pattern.String())
		}
	case map[string]any:
		if err := n.validateObjectWithDynamic(val, path, allowDynamic); err != nil {
			return err
		}
	}
	if elems, ok := arrayElements(v); ok {
		// Cardinality applies to array INSTANCES only, per draft-07
		// applicability. "required and non-empty" is therefore the composition
		// required + minItems, exactly as it is in real draft-07 — enforcing on
		// an absent value instead would silently change the meaning of every
		// optional array field already declared in a bundle.
		if err := n.validateArrayCardinality(len(elems), path); err != nil {
			return err
		}
		if n.items != nil {
			for i, elem := range elems {
				if err := n.items.validateWithDynamic(elem, fmt.Sprintf("%s/%d", path, i), allowDynamic); err != nil {
					return err
				}
			}
		}
	}

	for i, branch := range n.allOf {
		if err := branch.validateWithDynamic(v, path, allowDynamic); err != nil {
			return fmt.Errorf("%s: allOf[%d]: %w", displayPath(path), i, err)
		}
	}
	if len(n.anyOf) > 0 {
		matched := false
		for _, branch := range n.anyOf {
			if branch.validateWithDynamic(v, path, allowDynamic) == nil {
				matched = true
				break
			}
		}
		if !matched {
			return fmt.Errorf("%s: value does not match anyOf", displayPath(path))
		}
	}
	if len(n.oneOf) > 0 {
		matches := 0
		for _, branch := range n.oneOf {
			if branch.validateWithDynamic(v, path, allowDynamic) == nil {
				matches++
			}
		}
		if matches != 1 {
			return fmt.Errorf("%s: value matches %d oneOf branches", displayPath(path), matches)
		}
	}

	return nil
}

func (n *schemaNode) validateArrayCardinality(count int, path string) error {
	if n.hasMinItems && count < n.minItems {
		return fmt.Errorf("%s: minItems %d not satisfied (got %d)", displayPath(path), n.minItems, count)
	}
	if n.hasMaxItems && count > n.maxItems {
		return fmt.Errorf("%s: maxItems %d exceeded (got %d)", displayPath(path), n.maxItems, count)
	}
	return nil
}

func (n *schemaNode) validateObject(obj map[string]any, path string) error {
	return n.validateObjectWithDynamic(obj, path, false)
}

func (n *schemaNode) validateObjectWithDynamic(obj map[string]any, path string, allowDynamic bool) error {
	if n.hasMinProperties && len(obj) < n.minProperties {
		return fmt.Errorf("%s: minProperties %d not satisfied (got %d)", displayPath(path), n.minProperties, len(obj))
	}

	for _, req := range n.required {
		if _, ok := obj[req]; !ok {
			return fmt.Errorf("%s/%s: required property missing", displayPath(path), req)
		}
	}

	if n.hasAdditionalProps && !n.additionalProperties {
		keys := make([]string, 0, len(obj))
		for k := range obj {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			if _, declared := n.properties[k]; !declared {
				return fmt.Errorf("%s/%s: additional property not allowed", displayPath(path), k)
			}
		}
	}

	if n.properties != nil {
		keys := make([]string, 0, len(obj))
		for k := range obj {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			child, ok := n.properties[k]
			if !ok {
				continue
			}
			if err := child.validateWithDynamic(obj[k], path+"/"+k, allowDynamic); err != nil {
				return err
			}
		}
	}

	return nil
}

func displayPath(path string) string {
	if path == "" {
		return "/"
	}
	return path
}

// typeMatches reports whether v's JSON-decoded runtime type is one of types.
func typeMatches(v any, types []string) bool {
	for _, t := range types {
		if valueIsType(v, t) {
			return true
		}
	}
	return false
}

func arrayElements(v any) ([]any, bool) {
	if v == nil {
		return nil, false
	}
	if elems, ok := v.([]any); ok {
		return elems, true
	}
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Slice && rv.Kind() != reflect.Array {
		return nil, false
	}
	elems := make([]any, rv.Len())
	for i := 0; i < rv.Len(); i++ {
		elems[i] = rv.Index(i).Interface()
	}
	return elems, true
}

func valueIsType(v any, t string) bool {
	switch t {
	case "null":
		return v == nil
	case "string":
		_, ok := v.(string)
		return ok
	case "boolean":
		_, ok := v.(bool)
		return ok
	case "object":
		_, ok := v.(map[string]any)
		return ok
	case "array":
		_, ok := arrayElements(v)
		return ok
	case "integer":
		return isIntegerNumber(v)
	case "number":
		return isNumber(v)
	default:
		return false
	}
}

func isNumber(v any) bool {
	switch v.(type) {
	case json.Number, float64, float32, int, int64:
		return true
	default:
		return false
	}
}

func isIntegerNumber(v any) bool {
	switch n := v.(type) {
	case json.Number:
		if _, err := n.Int64(); err == nil {
			return true
		}
		f, err := n.Float64()
		return err == nil && f == float64(int64(f))
	case float64:
		return n == float64(int64(n))
	case int, int64:
		return true
	default:
		return false
	}
}

func enumEquals(v, want any) bool {
	vn, vok := normalizeNumber(v)
	wn, wok := normalizeNumber(want)
	if vok && wok {
		return vn == wn
	}
	return fmt.Sprint(v) == fmt.Sprint(want) && sameKind(v, want)
}

func sameKind(a, b any) bool {
	switch a.(type) {
	case string:
		_, ok := b.(string)
		return ok
	case bool:
		_, ok := b.(bool)
		return ok
	default:
		return true
	}
}

func normalizeNumber(v any) (float64, bool) {
	switch n := v.(type) {
	case json.Number:
		f, err := n.Float64()
		if err != nil {
			return 0, false
		}
		return f, true
	case float64:
		return n, true
	case int:
		return float64(n), true
	case int64:
		return float64(n), true
	default:
		return 0, false
	}
}

// SecretKeys returns the top-level property names marked x-secret: true.
func (s *Schema) SecretKeys() []string {
	if s.node.properties == nil {
		return nil
	}
	var out []string
	for name, child := range s.node.properties {
		if child.secret {
			out = append(out, name)
		}
	}
	sort.Strings(out)
	return out
}

// Properties returns the top-level declared property names.
func (s *Schema) Properties() []string {
	if s.node.properties == nil {
		return nil
	}
	out := make([]string, 0, len(s.node.properties))
	for name := range s.node.properties {
		out = append(out, name)
	}
	sort.Strings(out)
	return out
}

// Defaults returns the top-level property-name -> stringified-default map
// for every root property that declares a JSON Schema "default" annotation
// (gap-loop cycle-1 item 6, REVIEW-A.md C3). The stored default's raw
// decoded JSON value is stringified via the same rules
// engine/interpolate.go's stringify() uses for every other config-shaped
// value (a JSON string returns verbatim; a number/bool/other value is
// formatted via fmt.Sprint) so a materialized default slots into
// RuntimeConfig.Config — a map[string]string — exactly like any other
// caller-supplied config value would. A property with no "default" key at
// all is simply absent from the returned map (never a zero-value entry).
func (s *Schema) Defaults() map[string]string {
	if s.node.properties == nil {
		return nil
	}
	var out map[string]string
	for name, child := range s.node.properties {
		if !child.hasDefault {
			continue
		}
		if out == nil {
			out = make(map[string]string, len(s.node.properties))
		}
		out[name] = stringify(child.defaultVal)
	}
	return out
}

// DefaultTypeMismatches returns the sorted list of root property names whose
// declared "default" JSON value does NOT type-check against that same
// property's declared "type" (gap-loop cycle-1 item 6 validate rule,
// REVIEW-A.md C3: "default must type-check"). Used by
// `cmd/connectorgen validate` to hard-fail a spec.json whose default would
// silently materialize a value of the wrong shape into config (e.g. a
// default: 100 on a "type":"string" property, or a default:"yes" on a
// "type":"boolean" property) — connectorgen's own compiled *Schema already
// enforces "properties" structurally, so this reuses the same type-matching
// logic Validate() uses for ordinary instance data, applied to the default
// value itself.
func (s *Schema) DefaultTypeMismatches() []string {
	if s.node.properties == nil {
		return nil
	}
	var out []string
	for name, child := range s.node.properties {
		if !child.hasDefault {
			continue
		}
		if len(child.types) > 0 && !typeMatches(child.defaultVal, child.types) {
			out = append(out, name)
		}
	}
	sort.Strings(out)
	return out
}

// RequiredKeys returns the root-level "required" property name list (F4,
// REVIEW.md: resolveHeaders needs to distinguish a declared-but-OPTIONAL
// config key, whose runtime absence is tolerated by omitting the header that
// references it — e.g. Stripe-Account/account_id — from a declared-and-
// REQUIRED one, whose absence is a hard configuration error, never silently
// swallowed).
func (s *Schema) RequiredKeys() []string {
	return s.node.required
}

// PrimaryKeys returns the root-level x-primary-key list.
func (s *Schema) PrimaryKeys() []string {
	return s.node.primaryKey
}

// CursorFieldName returns the root-level x-cursor-field value ("" when unset).
func (s *Schema) CursorFieldName() string {
	return s.node.cursorField
}

// StreamSchema pairs a compiled record schema with its primary key and cursor
// field extensions, extracted once at bundle-load time for convenient reuse.
type StreamSchema struct {
	*Schema
	PrimaryKey  []string // x-primary-key
	CursorField string   // x-cursor-field
}

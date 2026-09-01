package boundary

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

var tokenPattern = regexp.MustCompile(`[A-Za-z0-9][A-Za-z0-9_-]*`)
var tokenOnlyPattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9_-]*$`)

type lexicon struct {
	connectors []connectorLexeme
	byName     map[string]connectorLexeme
}

type connectorLexeme struct {
	Name                   string
	DisplayName            string
	tokenAliases           []lexemeAlias
	weakTokenAliases       []lexemeAlias
	phraseAliases          []lexemeAlias
	weakPhraseAliases      []lexemeAlias
	commandTokenAliases    []lexemeAlias
	commandPhraseAliases   []lexemeAlias
	literalPrefixes        []string
	identifierPrefixes     []string
	weakIdentifierPrefixes []string
	commandIdentifierRoots []string
	identifierContains     []string
	weakDocs               bool
}

type lexemeAlias struct {
	Value string
	Lower string
	Alias bool
}

type metadataFile struct {
	Name            string   `json:"name"`
	DisplayName     string   `json:"display_name"`
	IntegrationType string   `json:"integration_type"`
	DocsURL         string   `json:"docs_url"`
	Aliases         []string `json:"aliases"`
}

func loadLexicon(root string) (lexicon, error) {
	defsDir := filepath.Join(root, "internal", "connectors", "defs")
	entries, err := os.ReadDir(defsDir)
	if err != nil {
		return lexicon{}, fmt.Errorf("read connector defs: %w", err)
	}

	seen := map[string]connectorLexeme{}
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		dirName := entry.Name()
		meta, err := readMetadata(filepath.Join(defsDir, dirName, "metadata.json"))
		if err != nil {
			return lexicon{}, fmt.Errorf("load connector metadata %s: %w", dirName, err)
		}
		name := strings.TrimSpace(meta.Name)
		if name == "" {
			return lexicon{}, fmt.Errorf("load connector metadata %s: name is required", dirName)
		}
		name = strings.ToLower(name)
		seen[name] = newConnectorLexeme(name, meta)
	}
	if len(seen) == 0 {
		return lexicon{}, fmt.Errorf("no connector metadata loaded from %s", defsDir)
	}

	connectors := make([]connectorLexeme, 0, len(seen))
	for _, c := range seen {
		connectors = append(connectors, c)
	}
	sort.Slice(connectors, func(i, j int) bool { return connectors[i].Name < connectors[j].Name })
	return lexicon{connectors: connectors, byName: seen}, nil
}

func newConnectorLexeme(name string, meta metadataFile) connectorLexeme {
	display := strings.TrimSpace(meta.DisplayName)
	strongName := strongConnectorNameAlias(name, meta)
	c := connectorLexeme{Name: name, DisplayName: display, weakDocs: (meta.IntegrationType == "" || strings.EqualFold(meta.IntegrationType, "api")) && strings.TrimSpace(meta.DocsURL) != ""}
	c.addTokenAlias("source-"+name, true, false)
	c.addTokenAlias("destination-"+name, true, false)
	c.addLiteralAlias(name, false, !strongName)
	if strongName {
		c.addLiteralPrefix(name)
		c.addIdentifierPrefix(strings.ReplaceAll(name, "-", ""))
	} else if weakIdentifierAlias(name, meta) {
		c.addWeakIdentifierPrefix(strings.ReplaceAll(name, "-", ""))
	}
	for _, alias := range append([]string{display}, meta.Aliases...) {
		c.addMetadataAlias(name, alias, meta)
	}
	c.sortAliases()
	return c
}

func readMetadata(path string) (metadataFile, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return metadataFile{}, err
	}
	var meta metadataFile
	if err := json.Unmarshal(b, &meta); err != nil {
		return metadataFile{}, err
	}
	return meta, nil
}

func compactDisplayName(display string) string {
	var b strings.Builder
	for _, r := range display {
		if (r >= 'A' && r <= 'Z') || (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			b.WriteRune(r)
		}
	}
	return b.String()
}

func (c *connectorLexeme) addMetadataAlias(name, alias string, meta metadataFile) {
	alias = strings.TrimSpace(alias)
	if alias == "" {
		return
	}
	strong := strongMetadataAlias(name, alias, meta)
	c.addLiteralAlias(alias, false, !strong)
	compact := compactDisplayName(alias)
	if compact != "" && compact != alias {
		c.addLiteralAlias(compact, false, !strong)
	}
	if !strong {
		return
	}
	if compactLower := strings.ToLower(compact); len(compactLower) >= 5 {
		c.addIdentifierPrefix(compactLower)
		if alias != simpleDisplayName(name) {
			c.addIdentifierContains(compactLower)
		}
	}
}

func (c *connectorLexeme) addLiteralAlias(value string, legacy, weak bool) {
	value = strings.TrimSpace(value)
	if value == "" {
		return
	}
	if tokenOnlyPattern.MatchString(value) {
		c.addTokenAlias(value, legacy, weak)
		return
	}
	c.addPhraseAlias(value, legacy, weak)
}

func (c *connectorLexeme) addTokenAlias(value string, legacy, weak bool) {
	alias := lexemeAlias{Value: value, Lower: strings.ToLower(value), Alias: legacy}
	if weak {
		c.weakTokenAliases = append(c.weakTokenAliases, alias)
		return
	}
	c.tokenAliases = append(c.tokenAliases, alias)
}

func (c *connectorLexeme) addPhraseAlias(value string, legacy, weak bool) {
	alias := lexemeAlias{Value: value, Lower: strings.ToLower(value), Alias: legacy}
	if weak {
		c.weakPhraseAliases = append(c.weakPhraseAliases, alias)
		return
	}
	c.phraseAliases = append(c.phraseAliases, alias)
}

func (c *connectorLexeme) addLiteralPrefix(prefix string) {
	prefix = strings.ToLower(strings.TrimSpace(prefix))
	if prefix != "" {
		c.literalPrefixes = append(c.literalPrefixes, prefix)
	}
}

func (c *connectorLexeme) addIdentifierPrefix(prefix string) {
	prefix = strings.ToLower(strings.TrimSpace(prefix))
	if prefix != "" {
		c.identifierPrefixes = append(c.identifierPrefixes, prefix)
	}
}

func (c *connectorLexeme) addWeakIdentifierPrefix(prefix string) {
	prefix = strings.ToLower(strings.TrimSpace(prefix))
	if prefix != "" {
		c.weakIdentifierPrefixes = append(c.weakIdentifierPrefixes, prefix)
	}
}

func (c *connectorLexeme) addCommandIdentifierRoot(prefix string) {
	prefix = strings.ToLower(strings.TrimSpace(prefix))
	if prefix != "" {
		c.commandIdentifierRoots = append(c.commandIdentifierRoots, prefix)
	}
}

func (c *connectorLexeme) addIdentifierContains(alias string) {
	alias = strings.ToLower(strings.TrimSpace(alias))
	if alias != "" {
		c.identifierContains = append(c.identifierContains, alias)
	}
}

func (c *connectorLexeme) sortAliases() {
	c.tokenAliases = uniqueAliases(c.tokenAliases)
	c.weakTokenAliases = uniqueAliases(c.weakTokenAliases)
	c.phraseAliases = uniqueAliases(c.phraseAliases)
	c.weakPhraseAliases = uniqueAliases(c.weakPhraseAliases)
	c.commandTokenAliases = uniqueAliases(c.commandTokenAliases)
	c.commandPhraseAliases = uniqueAliases(c.commandPhraseAliases)
	c.literalPrefixes = uniqueStrings(c.literalPrefixes)
	c.identifierPrefixes = uniqueStrings(c.identifierPrefixes)
	c.weakIdentifierPrefixes = uniqueStrings(c.weakIdentifierPrefixes)
	c.commandIdentifierRoots = uniqueStrings(c.commandIdentifierRoots)
	c.identifierContains = uniqueStrings(c.identifierContains)
}

func uniqueAliases(in []lexemeAlias) []lexemeAlias {
	seen := map[string]bool{}
	var out []lexemeAlias
	for _, alias := range in {
		key := alias.Lower + "\x00" + fmt.Sprint(alias.Alias)
		if alias.Lower == "" || seen[key] {
			continue
		}
		seen[key] = true
		out = append(out, alias)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Lower < out[j].Lower })
	return out
}

func uniqueStrings(in []string) []string {
	seen := map[string]bool{}
	var out []string
	for _, value := range in {
		if value == "" || seen[value] {
			continue
		}
		seen[value] = true
		out = append(out, value)
	}
	sort.Strings(out)
	return out
}

func strongConnectorNameAlias(name string, meta metadataFile) bool {
	if strings.Contains(name, "-") {
		return true
	}
	if meta.IntegrationType != "" && !strings.EqualFold(meta.IntegrationType, "api") {
		return false
	}
	return strongMetadataAlias(name, meta.DisplayName, meta)
}

func weakIdentifierAlias(name string, meta metadataFile) bool {
	if strings.Contains(name, "-") {
		return false
	}
	if meta.IntegrationType != "" && !strings.EqualFold(meta.IntegrationType, "api") {
		return false
	}
	return len(strings.ReplaceAll(name, "-", "")) >= 3
}

func strongMetadataAlias(name, alias string, meta metadataFile) bool {
	alias = strings.TrimSpace(alias)
	if alias == "" {
		return false
	}
	if meta.IntegrationType != "" && !strings.EqualFold(meta.IntegrationType, "api") {
		return false
	}
	compact := compactDisplayName(alias)
	if compact == "" {
		return false
	}
	nameCompact := strings.ReplaceAll(name, "-", "")
	compactLower := strings.ToLower(compact)
	if strings.Contains(name, "-") {
		return compactLower == nameCompact || strings.Contains(nameCompact, compactLower)
	}
	if compactLower != nameCompact {
		return len(compactLower) >= 5 && (strings.Contains(nameCompact, compactLower) || strings.Contains(compactLower, nameCompact))
	}
	return alias != simpleDisplayName(name)
}

func simpleDisplayName(name string) string {
	parts := strings.Split(name, "-")
	for i, part := range parts {
		if part == "" {
			continue
		}
		parts[i] = strings.ToUpper(part[:1]) + strings.ToLower(part[1:])
	}
	return strings.Join(parts, " ")
}

type literalMatch struct {
	Connector string
	Match     string
	Policy    bool
	Alias     bool
	Exact     bool
}

func (lx lexicon) literalMatches(value string, includeWeakExact, includeWeakDocs bool) []literalMatch {
	if value == "" || len(lx.connectors) == 0 {
		return nil
	}
	matchesByKey := map[string]literalMatch{}
	for _, c := range lx.connectors {
		for _, alias := range c.phraseAliases {
			if containsDelimitedFold(value, alias.Lower) {
				match := literalMatch{Connector: c.Name, Match: alias.Value, Alias: alias.Alias, Exact: !alias.Alias}
				matchesByKey[literalMatchKey(match)] = match
			}
		}
		if includeWeakExact || includeWeakDocs {
			for _, alias := range c.weakPhraseAliases {
				if c.allowWeakAlias(value, alias.Lower, includeWeakExact, includeWeakDocs) && containsDelimitedFold(value, alias.Lower) {
					match := literalMatch{Connector: c.Name, Match: alias.Value, Alias: alias.Alias, Exact: !alias.Alias}
					matchesByKey[literalMatchKey(match)] = match
				}
			}
		}
		for _, alias := range c.commandPhraseAliases {
			if literalHasCommandAliasContext(value, alias.Lower) {
				match := literalMatch{Connector: c.Name, Match: alias.Value, Exact: true}
				matchesByKey[literalMatchKey(match)] = match
			}
		}
		for _, alias := range c.commandTokenAliases {
			if literalHasCommandAliasContext(value, alias.Lower) {
				match := literalMatch{Connector: c.Name, Match: alias.Value, Exact: true}
				matchesByKey[literalMatchKey(match)] = match
			}
		}
	}
	tokens := tokenPattern.FindAllString(value, -1)
	for _, token := range tokens {
		lower := strings.ToLower(token)
		for _, c := range lx.connectors {
			match := matchToken(c, token, lower, c.allowWeakAlias(value, lower, includeWeakExact, includeWeakDocs))
			if match.Connector == "" {
				continue
			}
			matchesByKey[literalMatchKey(match)] = match
		}
	}
	matches := make([]literalMatch, 0, len(matchesByKey))
	for _, m := range matchesByKey {
		matches = append(matches, m)
	}
	sort.Slice(matches, func(i, j int) bool {
		if matches[i].Connector != matches[j].Connector {
			return matches[i].Connector < matches[j].Connector
		}
		if matches[i].Match != matches[j].Match {
			return matches[i].Match < matches[j].Match
		}
		return rulePriority(matches[i]) < rulePriority(matches[j])
	})
	return matches
}

func matchToken(c connectorLexeme, token, lower string, includeWeakExact bool) literalMatch {
	for _, alias := range c.tokenAliases {
		if lower == alias.Lower {
			return literalMatch{Connector: c.Name, Match: token, Alias: alias.Alias, Exact: !alias.Alias}
		}
	}
	if includeWeakExact {
		for _, alias := range c.weakTokenAliases {
			if lower == alias.Lower {
				return literalMatch{Connector: c.Name, Match: token, Alias: alias.Alias, Exact: !alias.Alias}
			}
		}
	}
	for _, prefix := range c.literalPrefixes {
		if tokenHasConnectorPrefix(lower, prefix) || valueHasConnectorCompoundPolicyPrefix(token, lower, prefix) {
			return literalMatch{Connector: c.Name, Match: token, Policy: true}
		}
	}
	for _, prefix := range c.weakIdentifierPrefixes {
		if valueHasConnectorCompoundPolicyPrefix(token, lower, prefix) {
			return literalMatch{Connector: c.Name, Match: token, Policy: true}
		}
	}
	for _, prefix := range c.commandIdentifierRoots {
		if valueHasConnectorCompoundPolicyPrefix(token, lower, prefix) {
			return literalMatch{Connector: c.Name, Match: token, Policy: true}
		}
	}
	return literalMatch{}
}

func (c connectorLexeme) allowWeakAlias(value, lowerAlias string, includeWeakExact, includeWeakDocs bool) bool {
	if includeWeakExact {
		return true
	}
	return includeWeakDocs && c.weakDocs && literalHasConnectorCommandContext(value, lowerAlias)
}

func (lx lexicon) identifierMatches(identifier string) []literalMatch {
	if identifier == "" {
		return nil
	}
	var matches []literalMatch
	lowerIdentifier := strings.ToLower(identifier)
	for _, c := range lx.connectors {
		if c.matchesIdentifier(identifier, lowerIdentifier) {
			matches = append(matches, literalMatch{Connector: c.Name, Match: identifier, Policy: true})
		}
	}
	return matches
}

func (c connectorLexeme) matchesIdentifier(identifier, lowerIdentifier string) bool {
	for _, alias := range c.identifierContains {
		if strings.Contains(lowerIdentifier, alias) {
			return true
		}
	}
	for _, prefix := range c.identifierPrefixes {
		if valueHasConnectorPrefix(identifier, lowerIdentifier, prefix) {
			return true
		}
	}
	for _, prefix := range c.weakIdentifierPrefixes {
		if valueHasConnectorCompoundPolicyPrefix(identifier, lowerIdentifier, prefix) {
			return true
		}
	}
	for _, prefix := range c.commandIdentifierRoots {
		if valueHasConnectorCompoundPolicyPrefix(identifier, lowerIdentifier, prefix) {
			return true
		}
	}
	return false
}

func tokenHasConnectorPrefix(token, prefix string) bool {
	if !strings.HasPrefix(token, prefix) || len(token) == len(prefix) {
		return false
	}
	next := token[len(prefix)]
	return next == '_' || next == '-'
}

func valueHasConnectorPrefix(value, lowerValue, prefix string) bool {
	if !strings.HasPrefix(lowerValue, prefix) {
		return false
	}
	if len(value) == len(prefix) {
		return true
	}
	next := value[len(prefix)]
	return next == '_' || next == '-' || (next >= 'A' && next <= 'Z')
}

func valueHasConnectorCompoundPolicyPrefix(value, lowerValue, prefix string) bool {
	if !valueHasConnectorPrefix(value, lowerValue, prefix) || len(value) == len(prefix) {
		return false
	}
	tail := identifierTailComponents(value[len(prefix):])
	return weakIdentifierTailLooksLikePolicy(tail)
}

func identifierTailComponents(tail string) []string {
	var components []string
	inComponent := false
	startIdx := 0
	for i := 0; i < len(tail); i++ {
		ch := tail[i]
		if ch == '_' || ch == '-' {
			if inComponent {
				components = append(components, strings.ToLower(tail[startIdx:i]))
			}
			inComponent = false
			continue
		}
		if !isASCIIAlphaNumeric(ch) {
			if inComponent {
				components = append(components, strings.ToLower(tail[startIdx:i]))
			}
			inComponent = false
			continue
		}
		start := !inComponent
		if i > 0 && isASCIIUpper(ch) {
			prev := tail[i-1]
			nextLower := i+1 < len(tail) && isASCIILower(tail[i+1])
			if isASCIILower(prev) || isASCIIDigit(prev) || (isASCIIUpper(prev) && nextLower) {
				start = true
			}
		}
		if start {
			if inComponent {
				components = append(components, strings.ToLower(tail[startIdx:i]))
			}
			startIdx = i
		}
		inComponent = true
	}
	if inComponent {
		components = append(components, strings.ToLower(tail[startIdx:]))
	}
	return components
}

func weakIdentifierTailLooksLikePolicy(components []string) bool {
	if len(components) < 2 {
		return false
	}
	for _, component := range components {
		if component == "policy" || component == "fallback" {
			return true
		}
	}
	return false
}

func isASCIIAlphaNumeric(ch byte) bool {
	return isASCIILower(ch) || isASCIIUpper(ch) || isASCIIDigit(ch)
}

func isASCIILower(ch byte) bool {
	return ch >= 'a' && ch <= 'z'
}

func isASCIIUpper(ch byte) bool {
	return ch >= 'A' && ch <= 'Z'
}

func isASCIIDigit(ch byte) bool {
	return ch >= '0' && ch <= '9'
}

func containsDelimitedFold(value, lowerAlias string) bool {
	lowerValue := strings.ToLower(value)
	start := 0
	for {
		idx := strings.Index(lowerValue[start:], lowerAlias)
		if idx < 0 {
			return false
		}
		idx += start
		before := idx - 1
		after := idx + len(lowerAlias)
		if isAliasBoundary(lowerValue, before) && isAliasBoundary(lowerValue, after) {
			return true
		}
		start = idx + 1
	}
}

func literalHasConnectorCommandContext(value, lowerAlias string) bool {
	lowerValue := strings.ToLower(value)
	needles := []string{
		"pm " + lowerAlias,
		"pm connectors inspect " + lowerAlias,
		"--connector " + lowerAlias,
		"source-" + lowerAlias,
		"destination-" + lowerAlias,
	}
	for _, needle := range needles {
		if strings.Contains(lowerValue, needle) {
			return true
		}
	}
	return false
}

func literalHasCommandAliasContext(value, lowerAlias string) bool {
	lowerValue := strings.ToLower(value)
	start := 0
	for {
		idx := strings.Index(lowerValue[start:], lowerAlias)
		if idx < 0 {
			return false
		}
		idx += start
		after := idx + len(lowerAlias)
		if isCommandAliasStartBoundary(lowerValue, idx) && isCommandAliasEndBoundary(lowerValue, after) {
			return true
		}
		start = idx + 1
	}
}

func isCommandAliasStartBoundary(value string, idx int) bool {
	if idx <= 0 {
		return true
	}
	ch := value[idx-1]
	return isASCIIWhitespace(ch) || strings.ContainsRune("`'\"$([{=;|&", rune(ch))
}

func isCommandAliasEndBoundary(value string, idx int) bool {
	if idx >= len(value) {
		return true
	}
	ch := value[idx]
	return isASCIIWhitespace(ch) || strings.ContainsRune("`'\")]};|&", rune(ch))
}

func isASCIIWhitespace(ch byte) bool {
	return ch == ' ' || ch == '\t' || ch == '\n' || ch == '\r'
}

func isAliasBoundary(value string, idx int) bool {
	if idx < 0 || idx >= len(value) {
		return true
	}
	ch := value[idx]
	return (ch < 'a' || ch > 'z') && (ch < '0' || ch > '9')
}

func literalMatchKey(match literalMatch) string {
	return match.Connector + "\x00" + match.Match + "\x00" + fmt.Sprint(match.Policy) + "\x00" + fmt.Sprint(match.Alias) + "\x00" + fmt.Sprint(match.Exact)
}

func rulePriority(m literalMatch) int {
	switch {
	case m.Alias:
		return 0
	case m.Policy:
		return 1
	case m.Exact:
		return 2
	default:
		return 3
	}
}

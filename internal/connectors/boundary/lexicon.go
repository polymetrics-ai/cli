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

type lexicon struct {
	connectors []connectorLexeme
	byName     map[string]connectorLexeme
}

type connectorLexeme struct {
	Name           string
	DisplayName    string
	DisplayCompact string
	Contextual     bool
}

type metadataFile struct {
	Name        string `json:"name"`
	DisplayName string `json:"display_name"`
}

func loadLexicon(root string) (lexicon, error) {
	defsDir := filepath.Join(root, "internal", "connectors", "defs")
	entries, err := os.ReadDir(defsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return lexicon{byName: map[string]connectorLexeme{}}, nil
		}
		return lexicon{}, fmt.Errorf("read connector defs: %w", err)
	}

	seen := map[string]connectorLexeme{}
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		dirName := entry.Name()
		meta := readMetadata(filepath.Join(defsDir, dirName, "metadata.json"))
		name := strings.TrimSpace(meta.Name)
		if name == "" {
			name = dirName
		}
		name = strings.ToLower(name)
		display := strings.TrimSpace(meta.DisplayName)
		seen[name] = connectorLexeme{
			Name:           name,
			DisplayName:    display,
			DisplayCompact: compactDisplayName(display),
			Contextual:     isContextualConnectorName(name),
		}
	}

	connectors := make([]connectorLexeme, 0, len(seen))
	for _, c := range seen {
		connectors = append(connectors, c)
	}
	sort.Slice(connectors, func(i, j int) bool { return connectors[i].Name < connectors[j].Name })
	return lexicon{connectors: connectors, byName: seen}, nil
}

func readMetadata(path string) metadataFile {
	b, err := os.ReadFile(path)
	if err != nil {
		return metadataFile{}
	}
	var meta metadataFile
	if err := json.Unmarshal(b, &meta); err != nil {
		return metadataFile{}
	}
	return meta
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

func isContextualConnectorName(name string) bool {
	// Names in this set are real connector IDs but also ordinary runtime words,
	// package concepts, or infrastructure names. They are intentionally not
	// matched by bare-token scans; only explicit source-/destination aliases are
	// high-confidence enough for the boundary guard.
	switch name {
	case "box", "drift", "file", "float", "harness", "harvest", "height", "local", "merge", "mode", "outbox", "postgres", "rss", "sample", "segment", "tempo", "warehouse":
		return true
	default:
		return false
	}
}

type literalMatch struct {
	Connector string
	Match     string
	Policy    bool
	Alias     bool
	Exact     bool
}

func (lx lexicon) literalMatches(value string) []literalMatch {
	if value == "" || len(lx.connectors) == 0 {
		return nil
	}
	tokens := tokenPattern.FindAllString(value, -1)
	if len(tokens) == 0 {
		return nil
	}
	matchesByKey := map[string]literalMatch{}
	for _, token := range tokens {
		lower := strings.ToLower(token)
		for _, c := range lx.connectors {
			match := matchToken(c, token, lower)
			if match.Connector == "" {
				continue
			}
			key := match.Connector + "\x00" + match.Match + "\x00" + fmt.Sprint(match.Policy) + "\x00" + fmt.Sprint(match.Alias) + "\x00" + fmt.Sprint(match.Exact)
			matchesByKey[key] = match
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

func matchToken(c connectorLexeme, token, lower string) literalMatch {
	if lower == "source-"+c.Name || lower == "destination-"+c.Name {
		return literalMatch{Connector: c.Name, Match: token, Alias: true}
	}
	if lower == c.Name {
		if c.Contextual {
			return literalMatch{}
		}
		return literalMatch{Connector: c.Name, Match: token, Exact: true}
	}
	if c.Contextual {
		return literalMatch{}
	}
	if strings.HasPrefix(lower, c.Name+"_") || strings.HasPrefix(lower, c.Name+"-") {
		return literalMatch{Connector: c.Name, Match: token, Policy: true}
	}
	return literalMatch{}
}

func (lx lexicon) identifierMatches(identifier string) []literalMatch {
	if identifier == "" {
		return nil
	}
	var matches []literalMatch
	lowerIdentifier := strings.ToLower(identifier)
	for _, c := range lx.connectors {
		if c.Contextual {
			continue
		}
		compactName := strings.ReplaceAll(c.Name, "-", "")
		if len(c.DisplayCompact) >= 5 && strings.Contains(identifier, c.DisplayCompact) {
			matches = append(matches, literalMatch{Connector: c.Name, Match: identifier, Policy: true})
			continue
		}
		if len(compactName) >= 5 && identifierHasConnectorPrefix(identifier, lowerIdentifier, compactName) {
			matches = append(matches, literalMatch{Connector: c.Name, Match: identifier, Policy: true})
		}
	}
	return matches
}

func identifierHasConnectorPrefix(identifier, lowerIdentifier, compactName string) bool {
	if !strings.HasPrefix(lowerIdentifier, compactName) {
		return false
	}
	if len(identifier) == len(compactName) {
		return true
	}
	next := identifier[len(compactName)]
	return next == '_' || next == '-' || (next >= 'A' && next <= 'Z')
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

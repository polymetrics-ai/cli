package connectors

import "strings"

// Direction describes whether a connector reads (source), writes (destination),
// or both. Under the unified per-system model a single connector may be
// bidirectional (e.g. postgres reads via CDC and writes as a destination).
type Direction string

const (
	DirectionSource        Direction = "source"
	DirectionDestination   Direction = "destination"
	DirectionBidirectional Direction = "bidirectional"
)

// BareName returns the canonical per-system name for a connector slug by
// stripping the legacy upstream-style "source-"/"destination-" prefix. This is the
// name used for per-system package folders and the unified connector identity:
//
//	BareName("source-github")      == "github"
//	BareName("destination-bigquery") == "bigquery"
//	BareName("github")             == "github"   // already bare
func BareName(slug string) string {
	slug = strings.TrimSpace(strings.ToLower(slug))
	if rest := strings.TrimPrefix(slug, "source-"); rest != slug {
		return rest
	}
	if rest := strings.TrimPrefix(slug, "destination-"); rest != slug {
		return rest
	}
	return slug
}

// DirectionForSlug derives the direction implied by a legacy slug prefix. A bare
// slug with no prefix returns the empty Direction.
func DirectionForSlug(slug string) Direction {
	s := strings.TrimSpace(strings.ToLower(slug))
	switch {
	case strings.HasPrefix(s, "source-"):
		return DirectionSource
	case strings.HasPrefix(s, "destination-"):
		return DirectionDestination
	default:
		return ""
	}
}

// ConnectorDefinitionsByBareName returns every catalog entry whose bare system
// name matches name. For the systems that exist as both a source and a
// destination (the unify pairs) this returns more than one entry; for everything
// else it returns zero or one.
func ConnectorDefinitionsByBareName(name string) []ConnectorDefinition {
	bare := BareName(name)
	var out []ConnectorDefinition
	for _, entry := range ConnectorCatalog() {
		if BareName(entry.Slug) == bare {
			out = append(out, entry)
		}
	}
	return out
}

// CanonicalDirection reports the combined direction for a bare system name based
// on which catalog entries exist for it. A system present as both source and
// destination is bidirectional — the shape the unified connector package targets.
func CanonicalDirection(name string) Direction {
	var hasSource, hasDest bool
	for _, entry := range ConnectorDefinitionsByBareName(name) {
		switch entry.Type {
		case ConnectorTypeSource:
			hasSource = true
		case ConnectorTypeDestination:
			hasDest = true
		}
	}
	switch {
	case hasSource && hasDest:
		return DirectionBidirectional
	case hasDest:
		return DirectionDestination
	case hasSource:
		return DirectionSource
	default:
		return ""
	}
}

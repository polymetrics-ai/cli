// Package moderation implements the repo content-moderation trusted-event
// filter: it decides which GitHub issue/PR/comment events should be scored and
// potentially actioned, and which must be skipped because the author is a real
// contributor. See docs/plans/repo-bot-protection-and-ml-content-moderation-plan.md
// §4.8a.
//
// The filter is pure logic with no I/O so it is shared by the pm flow query
// layer, the pi-mono moderation agent, and the auto-moderation GitHub Action.
// It must never let a maintainer/collaborator event trigger moderation.
package moderation

import "strings"

// Event is the subset of a GitHub event the filter consumes.
//
// AuthorAssociationValue is GitHub's author_association enum: OWNER, MEMBER,
// COLLABORATOR, CONTRIBUTOR, FIRST_TIMER, FIRST_TIME_CONTRIBUTOR, MANNEQUIN,
// NONE. Only the author_association + the collaborators roster + the
// trust-logins allow-list decide trust; the rest of the event is carried
// through unchanged by callers.
type Event interface {
	UserLogin() string
	AuthorAssociationValue() string
}

// DefaultExcludeAssociations is the default exclusion set: the association
// values that mark a real contributor and so must NOT trigger moderation.
//
// FIRST_TIMER, FIRST_TIME_CONTRIBUTOR, MANNEQUIN, and NONE are the moderation
// targets (outsiders + newcomers + imported identities).
var DefaultExcludeAssociations = []string{
	"OWNER",
	"MEMBER",
	"COLLABORATOR",
	"CONTRIBUTOR",
}

// FilterOptions parameterizes the trusted-event filter.
//
//   - ExcludeAssociations: association values to drop. nil (not empty) applies
//     DefaultExcludeAssociations; an empty (non-nil) slice excludes nothing.
//   - TrustLogins: an allow-list of logins to always drop regardless of
//     association (known-good bots, e.g. coderabbitai, dependabot), mirroring
//     the planned `pm flow run --trust-login` flag.
type FilterOptions struct {
	ExcludeAssociations []string
	TrustLogins         []string
}

// Filter returns the subset of events that should be moderated: events whose
// author_association is not in the exclusion set, whose login is not in the
// trust-logins allow-list, and whose login is not in the collaborators roster
// (the anti-join safety net, since author_association can lag or be NONE for a
// real collaborator reached via team membership).
//
// Both association matching and roster lookups are case-insensitive (GitHub
// logins are case-insensitive; author_association on the wire is uppercase).
// A event with an empty login never passes the roster/trust-logins checks but
// is otherwise filtered by association alone.
func Filter[E Event](events []E, opts FilterOptions, collaborators []string) []E {
	exclude := excludeSet(opts)
	trust := toSet(opts.TrustLogins)
	roster := toSet(collaborators)

	out := make([]E, 0, len(events))
	for _, e := range events {
		assoc := strings.ToUpper(strings.TrimSpace(e.AuthorAssociationValue()))
		login := strings.ToLower(strings.TrimSpace(e.UserLogin()))
		if exclude[assoc] {
			continue
		}
		if login != "" {
			if trust[login] || roster[login] {
				continue
			}
		}
		out = append(out, e)
	}
	return out
}

func excludeSet(opts FilterOptions) map[string]bool {
	if opts.ExcludeAssociations == nil {
		set := make(map[string]bool, len(DefaultExcludeAssociations))
		for _, a := range DefaultExcludeAssociations {
			set[strings.ToUpper(a)] = true
		}
		return set
	}
	set := make(map[string]bool, len(opts.ExcludeAssociations))
	for _, a := range opts.ExcludeAssociations {
		set[strings.ToUpper(strings.TrimSpace(a))] = true
	}
	return set
}

func toSet(values []string) map[string]bool {
	if len(values) == 0 {
		return map[string]bool{}
	}
	set := make(map[string]bool, len(values))
	for _, v := range values {
		s := strings.ToLower(strings.TrimSpace(v))
		if s != "" {
			set[s] = true
		}
	}
	return set
}

package moderation

import (
	"testing"
)

// Event is the minimal moderation event shape: the fields the trusted-event
// filter consumes. It mirrors the GitHub issue/PR/comment streams'
// author_association + user.login + body + html_url + created_at.
type testEvent struct {
	Login             string
	AuthorAssociation string
	Body              string
	URL               string
}

func (e testEvent) UserLogin() string              { return e.Login }
func (e testEvent) AuthorAssociationValue() string { return e.AuthorAssociation }

// FilterOptions mirrors the §4.8a CLI surface: --exclude-associations and
// --trust-login. DefaultExcludeAssociations is applied when nil.
func defaultOpts() FilterOptions {
	return FilterOptions{ExcludeAssociations: nil, TrustLogins: nil}
}

func wantIDs(in []testEvent) []testEvent { return in }

func TestFilter_DropsDefaultContributors(t *testing.T) {
	events := []testEvent{
		{"alice", "OWNER", "release notes", "u1"},
		{"bob", "MEMBER", "plans", "u2"},
		{"carol", "COLLABORATOR", "", "u3"},
		{"dave", "CONTRIBUTOR", "pr", "u4"},
		{"eve", "NONE", "buy cheap watches", "u5"},
		{"frank", "FIRST_TIME_CONTRIBUTOR", "first pr", "u6"},
		{"grace", "MANNEQUIN", "", "u7"},
	}
	got := Filter[testEvent](events, defaultOpts(), nil)
	if len(got) != 3 {
		t.Fatalf("got %d events, want 3 (NONE + FIRST_TIME_CONTRIBUTOR + MANNEQUIN): %+v", len(got), logins(got))
	}
	for _, e := range got {
		switch e.Login {
		case "eve", "frank", "grace":
		default:
			t.Fatalf("unexpected %q survived default exclusion", e.Login)
		}
	}
}

func TestFilter_CustomExcludeAssociations(t *testing.T) {
	events := []testEvent{
		{"alice", "OWNER", "x", "u1"},
		{"eve", "NONE", "spam", "u5"},
		{"frank", "FIRST_TIME_CONTRIBUTOR", "first", "u6"},
	}
	// Stricter: also exclude FIRST_TIME_CONTRIBUTOR -> moderate only NONE.
	got := Filter[testEvent](events, FilterOptions{
		ExcludeAssociations: []string{"OWNER", "MEMBER", "COLLABORATOR", "CONTRIBUTOR", "FIRST_TIME_CONTRIBUTOR"},
	}, nil)
	if len(got) != 1 || got[0].Login != "eve" {
		t.Fatalf("got %+v, want [eve]", logins(got))
	}
	// Looser: exclude nothing -> all survive.
	got = Filter[testEvent](events, FilterOptions{ExcludeAssociations: []string{}}, nil)
	if len(got) != 3 {
		t.Fatalf("looser exclusion got %d, want 3: %+v", len(got), logins(got))
	}
}

func TestFilter_TrustLoginsAntiJoin(t *testing.T) {
	events := []testEvent{
		{"eve", "NONE", "spam", "u5"},
		{"knownbot", "NONE", "automated note", "u6"},
	}
	got := Filter[testEvent](events, defaultOpts(), nil) // no trust-logins -> both survive
	if len(got) != 2 {
		t.Fatalf("baseline got %d, want 2", len(got))
	}
	got = Filter[testEvent](events, FilterOptions{TrustLogins: []string{"knownbot"}}, nil)
	if len(got) != 1 || got[0].Login != "eve" {
		t.Fatalf("trust-login got %+v, want [eve]", logins(got))
	}
}

func TestFilter_CollaboratorRosterAntiJoin(t *testing.T) {
	// eve is NONE by association but is in the collaborators roster -> drop.
	events := []testEvent{
		{"eve", "NONE", "spam", "u5"},
		{"mallory", "NONE", "zip", "u8"},
	}
	roster := []string{"eve"}
	got := Filter[testEvent](events, defaultOpts(), roster)
	if len(got) != 1 || got[0].Login != "mallory" {
		t.Fatalf("roster anti-join got %+v, want [mallory]", logins(got))
	}
}

func TestFilter_AssociationIsCaseInsensitive(t *testing.T) {
	events := []testEvent{
		{"alice", "owner", "x", "u1"},
		{"bob", "Member", "x", "u2"},
		{"eve", "none", "spam", "u3"},
	}
	got := Filter[testEvent](events, defaultOpts(), nil)
	if len(got) != 1 || got[0].Login != "eve" {
		t.Fatalf("case-insensitive got %+v, want [eve]", logins(got))
	}
}

func TestFilter_IncidentAcceptance(t *testing.T) {
	// §0 acceptance: the owner's OWN items are dropped; the 3 throwaway NONE
	// accounts are kept. Collab roster is the maintainer.
	events := []testEvent{
		{"karthik-sivadas", "OWNER", "sub-issues created", "c1"},
		{"nadebopo78", "NONE", "[monday_fix.zip]", "c2"},
		{"capakopugo", "NONE", "[monday_cli_fix.zip]", "c3"},
		{"bomokoma91", "NONE", "[gitlab_fix.zip]", "c4"},
	}
	roster := []string{"karthik-sivadas"}
	got := Filter[testEvent](events, defaultOpts(), roster)
	if len(got) != 3 {
		t.Fatalf("incident acceptance got %d, want 3: %+v", len(got), logins(got))
	}
	for _, e := range got {
		if e.Login == "karthik-sivadas" {
			t.Fatalf("owner event survived filter: %+v", logins(got))
		}
	}
}

func TestFilter_NilEventsIsSafe(t *testing.T) {
	got := Filter[testEvent](nil, defaultOpts(), nil)
	if got != nil && len(got) != 0 {
		t.Fatalf("nil events got %+v, want empty", got)
	}
}

func logins(in []testEvent) []string {
	out := make([]string, 0, len(in))
	for _, e := range in {
		out = append(out, e.Login)
	}
	return out
}

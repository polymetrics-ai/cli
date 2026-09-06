package engine

import "testing"

func TestOffsetCountPaginatorUsesCombinedProviderWindow(t *testing.T) {
	paginator, err := newPaginator(PaginationSpec{Type: "offset_count", LimitParam: "limit", PageSize: 100}, 100, "customers.customer")
	if err != nil {
		t.Fatalf("newPaginator: %v", err)
	}
	first := paginator.Start()
	if got := first.Query.Get("limit"); got != "0,100" {
		t.Fatalf("first combined limit = %q, want 0,100", got)
	}
	second := paginator.Next(nil, 100)
	if second == nil || second.Query.Get("limit") != "100,100" {
		t.Fatalf("second combined limit = %#v, want 100,100", second)
	}
	if next := paginator.Next(nil, 1); next != nil {
		t.Fatalf("short provider page produced %#v, want stop", next)
	}
}

func TestOffsetCountPaginatorRejectsMalformedDeclarationBeforeIO(t *testing.T) {
	for _, spec := range []PaginationSpec{
		{Type: "offset_count", PageSize: 100},
		{Type: "offset_count", LimitParam: "limit"},
	} {
		if _, err := newPaginator(spec, 0, "customers.customer"); err == nil {
			t.Fatalf("offset_count declaration %#v was accepted", spec)
		}
	}
}

func TestOffsetCountDeclarationRejectedAtLoadBoundary(t *testing.T) {
	if err := validateOffsetCountPagination(nil, []StreamSpec{{Name: "customers", Pagination: &PaginationSpec{Type: "offset_count", LimitParam: "limit"}}}); err == nil {
		t.Fatal("malformed offset_count stream passed load validation")
	}
}

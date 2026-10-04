package transactionLogRepository

import (
	"Bea-Cukai/model"
	"fmt"
	"strings"
	"testing"
)

func TestBuildQuery_FiltersAreBoundNotInlined(t *testing.T) {
	evil := "x' OR '1'='1"
	q, args := buildQuery(model.TransactionLogRequest{StartDate: "2026-09-01", EndDate: "2026-09-30", UserName: evil})

	if strings.Contains(q, evil) || strings.Contains(q, "2026-09-01") {
		t.Fatalf("filter values must not be inlined:\n%s", q)
	}
	want := []any{"2026-09-01", "2026-09-30", "2026-09-01", "2026-09-30", "%" + evil + "%"}
	if fmt.Sprint(args) != fmt.Sprint(want) {
		t.Errorf("args = %v, want %v", args, want)
	}
	if strings.Count(q, "?") != len(want) {
		t.Errorf("placeholders = %d, args = %d", strings.Count(q, "?"), len(want))
	}
}

func TestBuildQuery_NoFilters(t *testing.T) {
	q, args := buildQuery(model.TransactionLogRequest{StartDate: "2026-09-01"}) // date filter needs both ends
	if len(args) != 0 || strings.Contains(q, "?") {
		t.Errorf("expected no params, got %v in:\n%s", args, q)
	}
}

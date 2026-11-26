package dashboardadmin

import (
	"reflect"
	"testing"

	"casino/internal/content"
)

func TestParseFilterDefaultsAndLimits(t *testing.T) {
	filter, err := ParseFilter(map[string]string{"startTime": "2026-07-01 00:00:00", "endTime": "2026-07-07 23:59:59", "currency": "PKR", "timezone": "+08:00", "granularity": "day", "limit": "20"})
	if err != nil {
		t.Fatalf("ParseFilter() error = %v", err)
	}
	if filter.Currency != "PKR" || filter.Limit != 20 || filter.Granularity != "day" {
		t.Fatalf("unexpected filter: %#v", filter)
	}
}

func TestParseFilterRejectsInvalidRange(t *testing.T) {
	_, err := ParseFilter(map[string]string{"startTime": "2026-07-08 00:00:00", "endTime": "2026-07-07 23:59:59", "currency": "PKR"})
	if err == nil {
		t.Fatal("expected invalid range error")
	}
}

func TestParseFilterRejectsInvalidTimezone(t *testing.T) {
	_, err := ParseFilter(map[string]string{"startTime": "2026-07-01 00:00:00", "endTime": "2026-07-07 23:59:59", "currency": "PKR", "timezone": "Asia/Kuala_Lumpur"})
	if err == nil {
		t.Fatal("expected invalid timezone error")
	}
}

func TestPercentChangeHandlesZero(t *testing.T) {
	if value := percentChange("10.00", "0.00"); value != nil {
		t.Fatalf("expected nil change, got %v", value)
	}
	if value := percentChange("15.00", "10.00"); value != "50.00" {
		t.Fatalf("expected 50.00, got %v", value)
	}
}

func TestDashboardScopeSQLAddsH5AndSite(t *testing.T) {
	where, args := dashboardScopeSQL(content.Scope{SiteID: 3}, "u")
	if where != " AND u.site_id IN (0, ?)" {
		t.Fatalf("unexpected scope where: %s", where)
	}
	if !reflect.DeepEqual(args, []any{uint64(3)}) {
		t.Fatalf("unexpected args: %#v", args)
	}
}

func TestDashboardScopeSQLSkipsPlatformAdmin(t *testing.T) {
	where, args := dashboardScopeSQL(content.Scope{Platform: true}, "u")
	if where != "" || len(args) != 0 {
		t.Fatalf("platform scope should not add h5 filter: %q %#v", where, args)
	}
}

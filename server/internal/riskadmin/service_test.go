package riskadmin

import (
	"errors"
	"testing"
)

func TestCompareUsesExactDecimals(t *testing.T) {
	tests := []struct {
		name     string
		actual   any
		operator string
		expected string
		want     bool
	}{
		{name: "large decimal remains exact", actual: "9007199254740993.01", operator: "GREATER_THAN", expected: "9007199254740993.00", want: true},
		{name: "decimal less than", actual: "0.099", operator: "LESS_THAN", expected: "0.1", want: true},
		{name: "window threshold", actual: "500.00", operator: "SUM_WITHIN", expected: "500", want: true},
		{name: "membership", actual: "PK", operator: "IN", expected: "MY, PK, SG", want: true},
		{name: "invalid numeric fact", actual: "not-a-number", operator: "GREATER_THAN", expected: "1", want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := compare(tt.actual, tt.operator, tt.expected); got != tt.want {
				t.Fatalf("compare() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNormalizeBlacklist(t *testing.T) {
	tests := []struct {
		kind  string
		value string
		want  string
	}{
		{kind: "EMAIL", value: " User@Example.COM ", want: "user@example.com"},
		{kind: "MOBILE", value: "+92 301-234-5678", want: "923012345678"},
		{kind: "IP", value: "2001:0db8::1", want: "2001:db8::1"},
		{kind: "BANK_ACCOUNT", value: " pk 12 34 ", want: "PK1234"},
		{kind: "DEVICE", value: " Device-A ", want: "device-a"},
	}
	for _, tt := range tests {
		t.Run(tt.kind, func(t *testing.T) {
			got, err := normalizeBlacklist(tt.kind, tt.value)
			if err != nil {
				t.Fatalf("normalizeBlacklist() error = %v", err)
			}
			if got != tt.want {
				t.Fatalf("normalizeBlacklist() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestNormalizeBlacklistRejectsInvalidValues(t *testing.T) {
	_, err := normalizeBlacklist("IP", "invalid")
	var domainErr *Error
	if !errors.As(err, &domainErr) || domainErr.Code != "INVALID_BLACKLIST_VALUE" {
		t.Fatalf("expected INVALID_BLACKLIST_VALUE, got %v", err)
	}
}

func TestEventTransition(t *testing.T) {
	allowed := [][2]string{{"PENDING", "UNDER_REVIEW"}, {"UNDER_REVIEW", "CONFIRMED"}, {"CONFIRMED", "RESTRICTED"}}
	for _, transition := range allowed {
		if !eventTransition(transition[0], transition[1]) {
			t.Fatalf("expected transition %s -> %s to be allowed", transition[0], transition[1])
		}
	}
	forbidden := [][2]string{{"CLEARED", "BLOCKED"}, {"CONFIRMED", "UNDER_REVIEW"}, {"BLOCKED", "CLEARED"}}
	for _, transition := range forbidden {
		if eventTransition(transition[0], transition[1]) {
			t.Fatalf("expected transition %s -> %s to be rejected", transition[0], transition[1])
		}
	}
}

func TestNormalizedPageBounds(t *testing.T) {
	page, size, offset := normalizedPage(ListQuery{Page: 0, PageSize: 1000})
	if page != 1 || size != 100 || offset != 0 {
		t.Fatalf("normalizedPage() = (%d,%d,%d)", page, size, offset)
	}
}

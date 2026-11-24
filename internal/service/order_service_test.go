package service

import (
	"testing"
	"time"

	"golang-web/internal/timeutil"
)

func TestOrderLocked(t *testing.T) {
	today, _ := timeutil.ParseDateISO("2026-11-24")
	for _, tc := range []struct {
		date   string
		locked bool
	}{
		{"2026-11-26", false}, {"2026-11-25", true}, {"2026-11-24", true}, {"2026-11-23", true}, {"2026-12-24", false},
	} {
		d, _ := timeutil.ParseDateISO(tc.date)
		if got := OrderLocked(d, today); got != tc.locked {
			t.Errorf("OrderLocked(%s)=%v want %v", tc.date, got, tc.locked)
		}
	}
	now := time.Date(2026, 11, 23, 18, 0, 0, 0, time.UTC)
	event, _ := timeutil.ParseDateISO("2026-11-26")
	if OrderLocked(event, timeutil.Today(now)) {
		t.Fatal("H-2 event should remain editable at 24 Nov 01:00 WIB")
	}
}

func TestParsePortions(t *testing.T) {
	for _, s := range []string{"1", "100", " 7 ", "100000"} {
		if _, ok := parsePortions(s); !ok {
			t.Errorf("%q should be valid", s)
		}
	}
	for _, s := range []string{"", "0", "-5", "1.5", "1e3", "abc", "100001"} {
		if _, ok := parsePortions(s); ok {
			t.Errorf("%q should be invalid", s)
		}
	}
}

func TestParseOrderDates(t *testing.T) {
	event := "2026-12-10"
	_, production, errs := parseOrderDates(event, "")
	if errs != nil || timeutil.FormatDateISO(production) != event {
		t.Fatalf("default production: %v %v", production, errs)
	}
	_, _, errs = parseOrderDates(event, "2026-12-11")
	if errs == nil {
		t.Fatal("production after event should fail")
	}
	for _, date := range []string{"2026-12-10", "2026-12-01"} {
		if _, _, errs = parseOrderDates(event, date); errs != nil {
			t.Errorf("production %s: %v", date, errs)
		}
	}
}

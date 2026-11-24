package timeutil

import (
	"testing"
	"time"
)

func TestParseDateISO(t *testing.T) {
	valid, err := ParseDateISO("2026-12-01")
	if err != nil || FormatDateISO(valid) != "2026-12-01" {
		t.Fatalf("valid date: %v, %v", valid, err)
	}
	for _, input := range []string{"", "2026-13-01", "01-12-2026", "2026-12-01T00:00:00Z"} {
		if _, err := ParseDateISO(input); err == nil {
			t.Errorf("expected %q to fail", input)
		}
	}
}

func TestTodayUsesWIBAndFormat(t *testing.T) {
	now := time.Date(2026, 11, 23, 18, 0, 0, 0, time.UTC)
	today := Today(now)
	if got := FormatDateISO(today); got != "2026-11-24" {
		t.Fatalf("Today = %s", got)
	}
	if got := FormatDateID(time.Date(2026, 12, 1, 0, 0, 0, 0, time.UTC)); got != "1 Des 2026" {
		t.Fatalf("FormatDateID = %s", got)
	}
}

func TestDaysUntil(t *testing.T) {
	day, _ := ParseDateISO("2026-11-24")
	event, _ := ParseDateISO("2026-11-26")
	if got := DaysUntil(event, day); got != 2 {
		t.Fatalf("DaysUntil = %d", got)
	}
}

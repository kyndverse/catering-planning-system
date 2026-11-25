package service

import (
	"golang-web/internal/timeutil"
	"testing"
	"time"
)

func TestStockNumberParsing(t *testing.T) {
	for _, s := range []string{"2", " 3.5 ", "0.000001", "100000"} {
		if _, ok := ParsePositiveQuantity(s); !ok {
			t.Errorf("quantity %q rejected", s)
		}
	}
	for _, s := range []string{"", "0", "0.0", "-1", "abc", "1e3", "1,5", "1.1234567", "1234567890123", ".5", "5."} {
		if _, ok := ParsePositiveQuantity(s); ok {
			t.Errorf("quantity %q accepted", s)
		}
	}
	for _, s := range []string{"0", "140000", "140000.5", "140000.50", " 70000 "} {
		if _, ok := ParseMoney(s); !ok {
			t.Errorf("money %q rejected", s)
		}
	}
	for _, s := range []string{"", "-1", "abc", "1e3", "1.234", "1,5", "12345678901234"} {
		if _, ok := ParseMoney(s); ok {
			t.Errorf("money %q accepted", s)
		}
	}
	if !IsZeroMoney("0") || !IsZeroMoney("0.00") || IsZeroMoney("0.01") {
		t.Error("zero money")
	}
}
func TestStockSigns(t *testing.T) {
	if x, ok := SignedAdjustment("increase", "5"); !ok || x != "5" {
		t.Fatal(x)
	}
	if x, ok := SignedAdjustment("decrease", "5"); !ok || x != "-5" {
		t.Fatal(x)
	}
	for _, d := range []string{"", "up", "INCREASE"} {
		if _, ok := SignedAdjustment(d, "5"); ok {
			t.Fatal(d)
		}
	}
	if NegateForUsage("3000") != "-3000" {
		t.Fatal("usage sign")
	}
}

func TestMovementDateUsesWIBToday(t *testing.T) {
	now := time.Date(2026, 11, 24, 18, 0, 0, 0, time.UTC) // 25 Nov, 01:00 WIB
	if _, errs := parseStockDate("2026-11-25", now); errs != nil {
		t.Fatal(errs)
	}
	if _, errs := parseStockDate("2026-11-26", now); errs == nil {
		t.Fatal("future date accepted")
	}
	if got := timeutil.Today(now); timeutil.FormatDateISO(got) != "2026-11-25" {
		t.Fatal(got)
	}
}

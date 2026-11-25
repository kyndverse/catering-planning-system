package view

import (
	"golang-web/internal/repository"
	"testing"
)

func TestStockFormatters(t *testing.T) {
	cases := map[string]string{"10000": "+10000", "-3000": "-3000", "3500.0": "+3500", "0.5": "+0.5", "-0.25": "-0.25", "1000.000000": "+1000"}
	for in, want := range cases {
		if got := FormatSigned(in); got != want {
			t.Errorf("%s=%s", in, got)
		}
	}
	if Equivalent("10000", "5000", "sack 5 kg") != "2 sack 5 kg" || Equivalent("1", "3", "kg") != "0.33 kg" || Equivalent("1", "0", "kg") != "" {
		t.Fatal("equivalent")
	}
	if StockEquivalent("-200", "2000", "pouch 2 L") != "-0.1 pouch 2 L" || !IsNegative("-200") || IsNegative("-0") {
		t.Fatal("stock equivalent")
	}
}
func TestBuildMovementsURL(t *testing.T) {
	if got := BuildMovementsURL(repository.MovementFilter{}, 1); got != "/stock/movements" {
		t.Fatal(got)
	}
	if got := BuildMovementsURL(repository.MovementFilter{Type: "USAGE"}, 3); got != "/stock/movements?page=3&type=USAGE" {
		t.Fatal(got)
	}
}

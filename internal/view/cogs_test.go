package view

import (
	"math/big"
	"testing"
)

func rat(raw string) *big.Rat { r, _ := new(big.Rat).SetString(raw); return r }

func TestCogsCurrencyFormatting(t *testing.T) {
	for _, tc := range []struct{ in, want string }{{"10263", "Rp10.263"}, {"1234567", "Rp1.234.567"}, {"999.5", "Rp1.000"}, {"-25000", "-Rp25.000"}, {"-0.4", "Rp0"}} {
		if got := FormatRupiah(rat(tc.in), 0, 0); got != tc.want {
			t.Errorf("%s = %s, want %s", tc.in, got, tc.want)
		}
	}
	for _, tc := range []struct{ in, want string }{{"14.4", "Rp14,40"}, {"14", "Rp14,00"}, {"0.0714285714", "Rp0,0714"}} {
		if got := FormatRupiah(rat(tc.in), 2, 4); got != tc.want {
			t.Errorf("%s = %s", tc.in, got)
		}
	}
}
func TestCogsPercentFormatting(t *testing.T) {
	for _, tc := range []struct{ in, want string }{{"6/100", "+6%"}, {"-1/6", "-16.7%"}, {"0", "0%"}, {"1/1000", "+0.1%"}, {"1/10000", "0%"}} {
		if got := FormatPercentSigned(rat(tc.in)); got != tc.want {
			t.Errorf("%s = %s", tc.in, got)
		}
	}
}

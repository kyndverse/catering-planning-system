package view

import (
	"fmt"
	"golang-web/internal/timeutil"
	"math/big"
	"strings"
	"time"
)

// PriceRecord adalah satu baris tabel ingredient_prices apa adanya dari database.
type PriceRecord struct {
	ID                   int
	PricePerPurchaseUnit string    // kolom NUMERIC dibaca sebagai string, tanpa float
	EffectiveFrom        time.Time // kolom DATE
}

// PriceRow adalah baris siap tampil untuk template prices.html.
type PriceRow struct {
	ID                   int
	EffectiveFrom        string // "1 Nov 2026"
	PricePerPurchaseUnit string // "Rp72.000"
	PricePerBaseUnit     string // "Rp14,40"
	Status               string // "current" | "upcoming" | "past"
}

// BuildPriceRows mengubah hasil query menjadi baris tampilan dan menentukan
// harga yang sedang berlaku (current).
//
// Syarat: records sudah diurutkan effective_from MENURUN, yaitu
//
//	SELECT id, price_per_purchase_unit::text, effective_from
//	FROM ingredient_prices WHERE ingredient_id = $1
//	ORDER BY effective_from DESC
//
// baseQty adalah ingredients.base_qty_per_purchase_unit (string NUMERIC).
// current bernilai nil jika belum ada harga yang tanggal berlakunya <= hari ini.
func BuildPriceRows(records []PriceRecord, baseQty string, now time.Time) (rows []PriceRow, current *PriceRow, err error) {
	base, ok := new(big.Rat).SetString(baseQty)
	if !ok || base.Sign() <= 0 {
		return nil, nil, fmt.Errorf("invalid base_qty_per_purchase_unit: %q", baseQty)
	}

	today := timeutil.FormatDateISO(timeutil.Today(now))
	rows = make([]PriceRow, 0, len(records))
	currentIdx := -1

	for _, rec := range records {
		price, ok := new(big.Rat).SetString(rec.PricePerPurchaseUnit)
		if !ok {
			return nil, nil, fmt.Errorf("invalid price %q (id %d)", rec.PricePerPurchaseUnit, rec.ID)
		}

		// Membandingkan "YYYY-MM-DD" sebagai string aman dan bebas masalah zona waktu.
		eff := timeutil.FormatDateISO(rec.EffectiveFrom)

		status := "past"
		switch {
		case eff > today:
			status = "upcoming"
		case currentIdx == -1: // baris pertama (terbaru) yang sudah berlaku
			status = "current"
			currentIdx = len(rows)
		}

		rows = append(rows, PriceRow{
			ID:                   rec.ID,
			EffectiveFrom:        timeutil.FormatDateID(rec.EffectiveFrom),
			PricePerPurchaseUnit: FormatRupiah(price, 0, 0),
			PricePerBaseUnit:     FormatRupiah(new(big.Rat).Quo(price, base), 2, 2),
			Status:               status,
		})
	}

	if currentIdx >= 0 {
		current = &rows[currentIdx] // diambil setelah slice final, jadi pointer aman
	}
	return rows, current, nil
}

// formatRupiah: 72000 -> "Rp72.000"; 14.4 (2 desimal) -> "Rp14,40".
// Pembulatan hanya untuk tampilan; nilai asli tetap eksak di database.
func FormatRupiah(r *big.Rat, minDec, maxDec int) string {
	if r == nil {
		return ""
	}
	if minDec < 0 {
		minDec = 0
	}
	if maxDec < minDec {
		maxDec = minDec
	}
	rounded := r.FloatString(maxDec)
	negative := strings.HasPrefix(rounded, "-")
	rounded = strings.TrimPrefix(rounded, "-")
	intPart, frac, _ := strings.Cut(rounded, ".")
	if maxDec > minDec {
		frac = strings.TrimRight(frac, "0")
		if len(frac) < minDec {
			frac += strings.Repeat("0", minDec-len(frac))
		}
	}

	var b strings.Builder
	for i, c := range intPart {
		if i > 0 && (len(intPart)-i)%3 == 0 {
			b.WriteByte('.')
		}
		b.WriteRune(c)
	}

	out := "Rp" + b.String()
	if len(frac) > 0 {
		out += "," + frac
	}
	if negative && (intPart != "0" || strings.Trim(frac, "0") != "") {
		out = "-" + out
	}
	return out
}

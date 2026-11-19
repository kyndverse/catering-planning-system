package view

import (
	"golang-web/internal/model"
	"math/big"
	"regexp"
	"strings"
)

// ExplosionRow dan Explosion adalah data siap tampil untuk bagian
// "Kebutuhan Bahan Baku" di show.html (.explosion.Qty dan .explosion.Rows).
type ExplosionRow struct {
	IngredientName string
	Quantity       string
	Unit           string
}

type Explosion struct {
	Qty  string
	Rows []ExplosionRow
}

// Hanya desimal biasa dengan titik, dibatasi panjangnya. Menolak "1e3", "1/3", "-5", "abc".
var qtyPattern = regexp.MustCompile(`^\d{1,12}(\.\d{1,6})?$`)

// ResolveExplosionQty menentukan jumlah yang dihitung dari query ?qty=.
//   - kosong            -> default, tanpa pesan error
//   - valid dan > 0     -> dipakai apa adanya (sudah dinormalisasi)
//   - tidak valid / <= 0 -> default, ok = false (handler boleh menampilkan pesan)
//
// Default: 1 untuk resep berhasil "portion" (hitungan per porsi), selain itu
// sebanyak yield_qty resep (satu batch penuh, mis. 1000 g bumbu dasar).
func ResolveExplosionQty(raw, yieldQty, yieldUnit string) (qty string, ok bool) {
	def := defaultExplosionQty(yieldQty, yieldUnit)

	raw = strings.TrimSpace(raw)
	if raw == "" {
		return def, true
	}
	if !qtyPattern.MatchString(raw) {
		return def, false
	}
	r, valid := new(big.Rat).SetString(raw)
	if !valid || r.Sign() <= 0 {
		return def, false
	}
	return FormatQuantity(raw), true
}

func defaultExplosionQty(yieldQty, yieldUnit string) string {
	if yieldUnit == "portion" {
		return "1"
	}
	return FormatQuantity(yieldQty)
}

// NewExplosion mengubah hasil repository menjadi data tampilan.
func NewExplosion(qty string, reqs []model.IngredientRequirement) Explosion {
	rows := make([]ExplosionRow, 0, len(reqs))
	for _, q := range reqs {
		rows = append(rows, ExplosionRow{
			IngredientName: q.IngredientName,
			Quantity:       FormatQuantity(q.Quantity),
			Unit:           q.Unit,
		})
	}
	return Explosion{Qty: qty, Rows: rows}
}

// FormatQuantity membulatkan ke 2 desimal untuk TAMPILAN lalu membuang nol di belakang:
// "15000.0000000000000000" -> "15000", "14.50000" -> "14.5", "0.333333" -> "0.33".
// Nilai di database tetap eksak. Jika bukan angka valid, dikembalikan apa adanya.
func FormatQuantity(s string) string {
	r, ok := new(big.Rat).SetString(strings.TrimSpace(s))
	if !ok {
		return s
	}
	out := r.FloatString(2)
	if strings.Contains(out, ".") {
		out = strings.TrimRight(out, "0")
		out = strings.TrimSuffix(out, ".")
	}
	if out == "" || out == "-" {
		return "0"
	}
	return out
}

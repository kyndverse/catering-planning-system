package view

import (
	"golang-web/internal/model"
	"golang-web/internal/repository"
	"golang-web/internal/timeutil"
	"math/big"
	"net/url"
	"strconv"
	"strings"
)

type StockRow struct {
	IngredientID                                        int64
	Name, Stock, Unit, PurchaseEquivalent, LastMovement string
	Negative                                            bool
}
type MovementRow struct {
	Date, IngredientName, Type, Quantity, Unit, Equivalent, ProductionDate, TotalCost, Note string
	Negative                                                                                bool
}
type IngredientOptionRow struct {
	ID                                                          int64
	Name, BaseUnit, PurchaseUnit, BaseQtyPerPurchaseUnit, Stock string
}
type Pagination struct {
	Page, TotalPages int
	HasPrev, HasNext bool
	PrevURL, NextURL string
}

func FormatSigned(s string) string {
	r, ok := new(big.Rat).SetString(strings.TrimSpace(s))
	if !ok {
		return s
	}
	v := FormatQuantity(s)
	if r.Sign() >= 0 {
		return "+" + v
	}
	return v
}
func IsNegative(s string) bool {
	r, ok := new(big.Rat).SetString(strings.TrimSpace(s))
	return ok && r.Sign() < 0
}
func Equivalent(q, b, u string) string {
	a, ok := new(big.Rat).SetString(q)
	if !ok {
		return ""
	}
	a.Abs(a)
	return equivalent(a, b, u, false)
}
func StockEquivalent(q, b, u string) string {
	a, ok := new(big.Rat).SetString(q)
	if !ok {
		return ""
	}
	negative := a.Sign() < 0
	a.Abs(a)
	return equivalent(a, b, u, negative)
}
func equivalent(a *big.Rat, b, u string, negative bool) string {
	base, ok := new(big.Rat).SetString(b)
	if !ok || base.Sign() <= 0 {
		return ""
	}
	v := FormatQuantity(new(big.Rat).Quo(a, base).RatString())
	if negative {
		v = "-" + v
	}
	return v + " " + u
}
func BuildMovementsURL(f repository.MovementFilter, page int) string {
	q := url.Values{}
	if f.IngredientID != "" {
		q.Set("ingredient_id", f.IngredientID)
	}
	if f.Type != "" {
		q.Set("type", f.Type)
	}
	if f.From != "" {
		q.Set("from", f.From)
	}
	if f.To != "" {
		q.Set("to", f.To)
	}
	if page > 1 {
		q.Set("page", strconv.Itoa(page))
	}
	if len(q) == 0 {
		return "/stock/movements"
	}
	return "/stock/movements?" + q.Encode()
}
func StockRows(x []model.StockRecord) []StockRow {
	out := make([]StockRow, 0, len(x))
	for _, v := range x {
		d := ""
		if v.LastMovement != nil {
			d = timeutil.FormatDateID(*v.LastMovement)
		}
		out = append(out, StockRow{v.IngredientID, v.Name, FormatQuantity(v.StockQty), v.BaseUnit, StockEquivalent(v.StockQty, v.BaseQtyPerPurchaseUnit, v.PurchaseUnit), d, IsNegative(v.StockQty)})
	}
	return out
}
func MovementRows(x []model.MovementRecord) []MovementRow {
	out := make([]MovementRow, 0, len(x))
	for _, v := range x {
		pd := ""
		if v.ProductionDate != nil {
			pd = timeutil.FormatDateID(*v.ProductionDate)
		}
		cost := ""
		if v.TotalCost != nil {
			cost = FormatRupiahString(*v.TotalCost)
		}
		eq := ""
		if v.Type == "PURCHASE" {
			eq = Equivalent(v.Quantity, v.BaseQtyPerPurchaseUnit, v.PurchaseUnit)
		}
		out = append(out, MovementRow{timeutil.FormatDateID(v.Date), v.IngredientName, v.Type, FormatSigned(v.Quantity), v.BaseUnit, eq, pd, cost, v.Note, IsNegative(v.Quantity)})
	}
	return out
}
func IngredientOptionRows(x []model.IngredientOption) []IngredientOptionRow {
	out := make([]IngredientOptionRow, 0, len(x))
	for _, v := range x {
		out = append(out, IngredientOptionRow{v.ID, v.Name, v.BaseUnit, v.PurchaseUnit, v.BaseQtyPerPurchaseUnit, FormatQuantity(v.StockQty)})
	}
	return out
}

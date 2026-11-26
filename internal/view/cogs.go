package view

import (
	"golang-web/internal/model"
	"golang-web/internal/timeutil"
	"math/big"
	"strings"
)

type CogsMenuRow struct {
	ID                      int64
	Name, Cogs, Status      string
	ItemCount, MissingCount int
}
type CogsDetailRow struct {
	IngredientName, Quantity, Unit, PricePerBaseUnit, PriceSince, Cost, Share string
	HasPrice                                                                  bool
}

func FormatRupiahSigned(r *big.Rat) string {
	v := FormatRupiah(r, 0, 0)
	if r != nil && r.Sign() > 0 && v != "Rp0" {
		return "+" + v
	}
	return v
}
func FormatPercentSigned(p *big.Rat) string {
	if p == nil {
		return ""
	}
	v := new(big.Rat).Mul(p, big.NewRat(100, 1))
	s := v.FloatString(1)
	negative := strings.HasPrefix(s, "-")
	s = strings.TrimPrefix(s, "-")
	s = strings.TrimRight(strings.TrimRight(s, "0"), ".")
	if s == "" {
		s = "0"
	}
	if s == "0" {
		return "0%"
	}
	if negative {
		return "-" + s + "%"
	}
	return "+" + s + "%"
}
func FormatSharePercent(part, total *big.Rat) string {
	if part == nil || total == nil || total.Sign() <= 0 {
		return ""
	}
	return new(big.Rat).Mul(new(big.Rat).Quo(part, total), big.NewRat(100, 1)).FloatString(0) + "%"
}
func FormatQuantityMax4(raw string) string {
	r, ok := new(big.Rat).SetString(raw)
	if !ok {
		return raw
	}
	s := r.FloatString(4)
	return strings.TrimRight(strings.TrimRight(s, "0"), ".")
}
func CogsMenus(items []model.MenuCogs) []CogsMenuRow {
	out := make([]CogsMenuRow, 0, len(items))
	for _, x := range items {
		v := CogsMenuRow{ID: x.Menu.ID, Name: x.Menu.Name, Status: x.Status, ItemCount: len(x.Lines), MissingCount: x.Missing}
		if x.Status != "empty" {
			v.Cogs = FormatRupiah(x.Total, 0, 0)
		}
		out = append(out, v)
	}
	return out
}
func CogsDetail(x *model.MenuCogs) (CogsMenuRow, []CogsDetailRow) {
	menu := CogsMenus([]model.MenuCogs{*x})[0]
	rows := make([]CogsDetailRow, 0, len(x.Lines))
	for _, line := range x.Lines {
		v := CogsDetailRow{IngredientName: line.IngredientName, Quantity: FormatQuantityMax4(line.Qty.RatString()), Unit: line.Unit, HasPrice: line.Cost != nil}
		if v.HasPrice {
			v.PricePerBaseUnit = FormatRupiah(line.PricePerBase, 2, 4)
			if line.PriceSince != nil {
				v.PriceSince = timeutil.FormatDateID(*line.PriceSince)
			}
			v.Cost = FormatRupiah(line.Cost, 0, 0)
			v.Share = FormatSharePercent(line.Cost, x.Total)
		}
		rows = append(rows, v)
	}
	return menu, rows
}

type UsageDateRow struct {
	Date, Label               string
	OrderCount, TotalPortions int
	UsageRecorded             bool
}
type UsageProduction struct {
	Date, Label, PrevDate, NextDate string
	OrderCount, TotalPortions       int
}
type UsageRow struct {
	IngredientName, Unit, Planned, Actual, Variance, VariancePct, VarianceValue, Status string
	HasPrice                                                                            bool
}
type UsageSummaryView struct {
	OverCount, UnderCount, UnrecordedCount int
	TotalVarianceValue                     string
	HasValue, ValueOver, PriceIncomplete   bool
}

func UsageDates(rows []model.UsageDateSummary) []UsageDateRow {
	out := make([]UsageDateRow, 0, len(rows))
	for _, x := range rows {
		out = append(out, UsageDateRow{Date: timeutil.FormatDateISO(x.Date), Label: timeutil.FormatDateID(x.Date), OrderCount: x.OrderCount, TotalPortions: x.TotalPortions, UsageRecorded: x.UsageRecorded})
	}
	return out
}
func UsageReport(x *model.UsageReport) (UsageProduction, []UsageRow, UsageSummaryView) {
	p := UsageProduction{Date: timeutil.FormatDateISO(x.Date), Label: timeutil.FormatDateID(x.Date), PrevDate: timeutil.FormatDateISO(x.Date.AddDate(0, 0, -1)), NextDate: timeutil.FormatDateISO(x.Date.AddDate(0, 0, 1)), OrderCount: x.OrderCount, TotalPortions: x.TotalPortions}
	rows := make([]UsageRow, 0, len(x.Lines))
	for _, line := range x.Lines {
		v := UsageRow{IngredientName: line.IngredientName, Unit: line.Unit, Planned: FormatQuantity(line.Planned.RatString()), Actual: FormatQuantity(line.Actual.RatString()), Status: line.Status, HasPrice: line.Value != nil}
		if line.Status != "unrecorded" {
			v.Variance = FormatSigned(line.Variance.RatString())
			v.VariancePct = FormatPercentSigned(line.Pct)
			if v.HasPrice {
				v.VarianceValue = FormatRupiahSigned(line.Value)
			}
		}
		rows = append(rows, v)
	}
	s := UsageSummaryView{OverCount: x.Summary.OverCount, UnderCount: x.Summary.UnderCount, UnrecordedCount: x.Summary.UnrecordedCount, HasValue: x.Summary.HasValue, PriceIncomplete: x.Summary.PriceIncomplete}
	if s.HasValue {
		s.TotalVarianceValue = FormatRupiahSigned(x.Summary.TotalValue)
		s.ValueOver = x.Summary.TotalValue.Sign() > 0
	}
	return p, rows, s
}

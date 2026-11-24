package view

import (
	"math/big"

	"golang-web/internal/model"
	"golang-web/internal/timeutil"
)

type ProductionDateRow struct {
	Date, Label               string
	OrderCount, TotalPortions int
	AllLocked                 bool
}
type ProductionHeader struct {
	Date, Label, PrevDate, NextDate string
	OrderCount, TotalPortions       int
	HasEditableOrders               bool
}
type ProductionOrder struct {
	ID                      int64
	CustomerName, EventDate string
	TotalPortions           int
	Locked                  bool
}
type ProductionRequirementRow struct {
	Name, Required, Stock, Unit string
	Short                       bool
}
type ShoppingRow struct {
	Name, Shortage, Unit                         string
	BuyQty                                       int64
	PurchaseUnit, PricePerPurchaseUnit, Subtotal string
	HasPrice                                     bool
}
type ShoppingView struct {
	Rows            []ShoppingRow
	TotalCost       string
	PriceIncomplete bool
}

func ProductionDates(rows []model.ProductionDateSummary) []ProductionDateRow {
	out := make([]ProductionDateRow, 0, len(rows))
	for _, row := range rows {
		out = append(out, ProductionDateRow{Date: timeutil.FormatDateISO(row.Date), Label: timeutil.FormatDateID(row.Date), OrderCount: row.OrderCount, TotalPortions: row.TotalPortions, AllLocked: row.AllLocked})
	}
	return out
}
func ProductionPlan(plan *model.ProductionPlan) (ProductionHeader, []ProductionOrder, []model.ProductionMenu, []ProductionRequirementRow, ShoppingView) {
	header := ProductionHeader{Date: timeutil.FormatDateISO(plan.Date), Label: timeutil.FormatDateID(plan.Date), PrevDate: timeutil.FormatDateISO(plan.Date.AddDate(0, 0, -1)), NextDate: timeutil.FormatDateISO(plan.Date.AddDate(0, 0, 1)), OrderCount: len(plan.Orders), TotalPortions: plan.TotalPortions, HasEditableOrders: plan.HasEditableOrders}
	orders := make([]ProductionOrder, 0, len(plan.Orders))
	for _, row := range plan.Orders {
		orders = append(orders, ProductionOrder{ID: row.ID, CustomerName: row.CustomerName, EventDate: timeutil.FormatDateID(row.EventDate), TotalPortions: row.TotalPortions, Locked: row.Locked})
	}
	requirements := make([]ProductionRequirementRow, 0, len(plan.Requirements))
	for _, row := range plan.Requirements {
		requirements = append(requirements, ProductionRequirementRow{Name: row.Name, Unit: row.Unit, Required: FormatQuantity(row.Required), Stock: FormatQuantity(row.Stock), Short: row.Short})
	}
	shopping := ShoppingView{Rows: make([]ShoppingRow, 0, len(plan.Shopping.Items)), TotalCost: FormatRupiahString(plan.Shopping.TotalCost), PriceIncomplete: plan.Shopping.PriceIncomplete}
	for _, row := range plan.Shopping.Items {
		v := ShoppingRow{Name: row.Name, Unit: row.Unit, Shortage: FormatQuantity(row.Shortage), BuyQty: row.BuyQty, PurchaseUnit: row.PurchaseUnit, HasPrice: row.HasPrice}
		if row.HasPrice {
			v.PricePerPurchaseUnit = FormatRupiahString(row.PricePerPurchaseUnit)
			v.Subtotal = FormatRupiahString(row.Subtotal)
		}
		shopping.Rows = append(shopping.Rows, v)
	}
	return header, orders, plan.Menus, requirements, shopping
}

func FormatRupiahString(raw string) string {
	r, ok := new(big.Rat).SetString(raw)
	if !ok {
		return raw
	}
	decimals := 2
	if r.IsInt() {
		decimals = 0
	}
	return formatRupiah(r, decimals)
}

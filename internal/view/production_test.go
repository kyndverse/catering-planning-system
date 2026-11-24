package view

import (
	"golang-web/internal/model"
	"golang-web/internal/timeutil"
	"testing"
)

func TestProductionPlanView(t *testing.T) {
	date, _ := timeutil.ParseDateISO("2026-12-01")
	plan := &model.ProductionPlan{Date: date, Shopping: model.ShoppingList{TotalCost: "70000", Items: []model.ShoppingItem{{Name: "Rice", Shortage: "15000.0000", HasPrice: true, PricePerPurchaseUnit: "70000", Subtotal: "70000"}, {Name: "Salt", Shortage: "1", HasPrice: false}}}}
	header, _, _, _, shopping := ProductionPlan(plan)
	if header.PrevDate != "2026-11-30" || header.NextDate != "2026-12-02" {
		t.Fatalf("header %#v", header)
	}
	if shopping.TotalCost != "Rp70.000" || shopping.Rows[0].Shortage != "15000" || shopping.Rows[1].PricePerPurchaseUnit != "" {
		t.Fatalf("shopping %#v", shopping)
	}
	if FormatQuantity("0.333333") != "0.33" {
		t.Fatal("quantity format")
	}
}

package service

import (
	"golang-web/internal/model"
	"testing"
)

func TestBuildMenuCogsNasiKotakA(t *testing.T) {
	price := func(v string) *string { return &v }
	rows := []model.MenuIngredientCostRow{
		{IngredientName: "Chicken", BaseUnit: "g", BaseQtyPerPurchaseUnit: "1000", QtyPerPortion: "150", PricePerPurchaseUnit: price("38000")},
		{IngredientName: "Chili", BaseUnit: "g", BaseQtyPerPurchaseUnit: "1000", QtyPerPortion: "30", PricePerPurchaseUnit: price("50000")},
		{IngredientName: "Cooking Oil", BaseUnit: "ml", BaseQtyPerPurchaseUnit: "2000", QtyPerPortion: "39", PricePerPurchaseUnit: price("34000")},
		{IngredientName: "Rice", BaseUnit: "g", BaseQtyPerPurchaseUnit: "5000", QtyPerPortion: "100", PricePerPurchaseUnit: price("70000")},
		{IngredientName: "Shallot", BaseUnit: "g", BaseQtyPerPurchaseUnit: "1000", QtyPerPortion: "24", PricePerPurchaseUnit: price("40000")},
		{IngredientName: "Turmeric", BaseUnit: "g", BaseQtyPerPurchaseUnit: "1000", QtyPerPortion: "2", PricePerPurchaseUnit: price("20000")},
	}
	got, err := BuildMenuCogs(model.MenuRef{Name: "Nasi Kotak A"}, rows)
	if err != nil || got.Status != "complete" || got.Missing != 0 || ratString(got.Total) != "10263" {
		t.Fatalf("%#v, %v", got, err)
	}
}
func TestBuildMenuCogsIncompleteAndInvalid(t *testing.T) {
	price := "70000"
	got, err := BuildMenuCogs(model.MenuRef{}, []model.MenuIngredientCostRow{{IngredientName: "Rice", BaseQtyPerPurchaseUnit: "5000", QtyPerPortion: "100", PricePerPurchaseUnit: &price}, {IngredientName: "Saffron", BaseQtyPerPurchaseUnit: "2", QtyPerPortion: "5"}})
	if err != nil || got.Status != "incomplete" || got.Missing != 1 || ratString(got.Total) != "1400" || got.Lines[1].Cost != nil {
		t.Fatalf("%#v %v", got, err)
	}
	if _, err = BuildMenuCogs(model.MenuRef{}, []model.MenuIngredientCostRow{{BaseQtyPerPurchaseUnit: "0", QtyPerPortion: "1"}}); err == nil {
		t.Fatal("zero base accepted")
	}
}

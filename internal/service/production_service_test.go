package service

import (
	"testing"
	"time"

	"golang-web/internal/model"
)

func price(value string) *string { return &value }

func TestBuildShoppingListNasiKotakA(t *testing.T) {
	rows := []model.ProductionRequirementRow{
		{Name: "Chicken", BaseUnit: "g", PurchaseUnit: "kg", BaseQtyPerPurchaseUnit: "1000", RequiredQty: "15000", StockQty: "0", PricePerPurchaseUnit: price("38000")},
		{Name: "Chili", BaseUnit: "g", PurchaseUnit: "kg", BaseQtyPerPurchaseUnit: "1000", RequiredQty: "3000", StockQty: "0", PricePerPurchaseUnit: price("50000")},
		{Name: "Cooking Oil", BaseUnit: "ml", PurchaseUnit: "pouch 2 L", BaseQtyPerPurchaseUnit: "2000", RequiredQty: "3900", StockQty: "1000", PricePerPurchaseUnit: price("34000")},
		{Name: "Rice", BaseUnit: "g", PurchaseUnit: "sack 5 kg", BaseQtyPerPurchaseUnit: "5000", RequiredQty: "10000", StockQty: "5000", PricePerPurchaseUnit: price("70000")},
		{Name: "Shallot", BaseUnit: "g", PurchaseUnit: "kg", BaseQtyPerPurchaseUnit: "1000", RequiredQty: "2400", StockQty: "0", PricePerPurchaseUnit: price("40000")},
		{Name: "Turmeric", BaseUnit: "g", PurchaseUnit: "kg", BaseQtyPerPurchaseUnit: "1000", RequiredQty: "200", StockQty: "0", PricePerPurchaseUnit: price("20000")},
	}
	got, err := BuildShoppingList(rows)
	if err != nil {
		t.Fatal(err)
	}
	if got.TotalCost != "998000" || got.PriceIncomplete || len(got.Items) != 6 {
		t.Fatalf("unexpected list: %#v", got)
	}
	wantBuy := []int64{15, 3, 2, 1, 3, 1}
	wantSubtotal := []string{"570000", "150000", "68000", "70000", "120000", "20000"}
	for i, item := range got.Items {
		if item.BuyQty != wantBuy[i] || item.Subtotal != wantSubtotal[i] {
			t.Errorf("item %s = %#v", item.Name, item)
		}
	}
}

func TestBuildShoppingListRoundingAndIncompletePrice(t *testing.T) {
	for _, tc := range []struct {
		required string
		buy      int64
	}{{"5000", 1}, {"5000.0001", 2}, {"4999.9999", 1}, {"10000", 2}, {"0.0001", 1}} {
		got, err := BuildShoppingList([]model.ProductionRequirementRow{{Name: "Rice", BaseQtyPerPurchaseUnit: "5000", RequiredQty: tc.required, StockQty: "0"}})
		if err != nil || len(got.Items) != 1 || got.Items[0].BuyQty != tc.buy || !got.PriceIncomplete {
			t.Errorf("%s: %#v, %v", tc.required, got, err)
		}
	}
	for _, stock := range []string{"5000", "6000"} {
		got, err := BuildShoppingList([]model.ProductionRequirementRow{{BaseQtyPerPurchaseUnit: "5000", RequiredQty: "5000", StockQty: stock}})
		if err != nil || len(got.Items) != 0 {
			t.Errorf("stock %s: %#v %v", stock, got, err)
		}
	}
	if _, err := BuildShoppingList([]model.ProductionRequirementRow{{BaseQtyPerPurchaseUnit: "0", RequiredQty: "1", StockQty: "0"}}); err == nil {
		t.Error("zero base accepted")
	}
	if _, err := BuildShoppingList([]model.ProductionRequirementRow{{BaseQtyPerPurchaseUnit: "1", RequiredQty: "invalid", StockQty: "0"}}); err == nil {
		t.Error("invalid decimal accepted")
	}
}

func TestBuildRequirements(t *testing.T) {
	got := BuildRequirements([]model.ProductionRequirementRow{{Name: "A", RequiredQty: "2", StockQty: "1"}, {Name: "B", RequiredQty: "2", StockQty: "2"}})
	if !got[0].Short || got[1].Short {
		t.Fatalf("short flags: %#v", got)
	}
}

func TestProductionAllLockedRule(t *testing.T) {
	today := time.Date(2026, 11, 24, 0, 0, 0, 0, time.UTC)
	locked, _ := time.Parse("2006-01-02", "2026-11-25")
	editable, _ := time.Parse("2006-01-02", "2026-11-26")
	if !OrderLocked(locked, today) || OrderLocked(editable, today) {
		t.Fatal("unexpected locked status")
	}
}

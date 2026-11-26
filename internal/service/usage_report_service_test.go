package service

import (
	"golang-web/internal/model"
	"testing"
)

func TestBuildUsageScenario(t *testing.T) {
	p := func(v string) *string { return &v }
	rows := []model.UsageInputRow{
		{IngredientName: "Chicken", BaseQtyPerPurchaseUnit: "1000", PlannedQty: "15000", ActualQty: "15000", HasActual: true, PricePerPurchaseUnit: p("38000")},
		{IngredientName: "Chili", BaseQtyPerPurchaseUnit: "1000", PlannedQty: "3000", ActualQty: "2500", HasActual: true, PricePerPurchaseUnit: p("50000")},
		{IngredientName: "Cooking Oil", BaseQtyPerPurchaseUnit: "2000", PlannedQty: "3900", ActualQty: "4095", HasActual: true, PricePerPurchaseUnit: p("34000")},
		{IngredientName: "Rice", BaseQtyPerPurchaseUnit: "5000", PlannedQty: "10000", ActualQty: "10600", HasActual: true, PricePerPurchaseUnit: p("70000")},
		{IngredientName: "Saffron", BaseQtyPerPurchaseUnit: "2", PlannedQty: "0", ActualQty: "5", HasActual: true},
		{IngredientName: "Shallot", BaseQtyPerPurchaseUnit: "1000", PlannedQty: "2400", ActualQty: "0", HasActual: false, PricePerPurchaseUnit: p("40000")},
		{IngredientName: "Turmeric", BaseQtyPerPurchaseUnit: "1000", PlannedQty: "200", ActualQty: "200", HasActual: true, PricePerPurchaseUnit: p("20000")}}
	lines, sum, err := BuildUsage(rows, UsageVarianceTolerance)
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"match", "under", "match", "over", "unplanned", "unrecorded", "match"}
	for i := range want {
		if lines[i].Status != want[i] {
			t.Errorf("%d %s", i, lines[i].Status)
		}
	}
	if sum.OverCount != 1 || sum.UnderCount != 1 || sum.UnrecordedCount != 1 || !sum.HasValue || !sum.PriceIncomplete || ratString(sum.TotalValue) != "-13285" {
		t.Fatalf("%#v", sum)
	}
}
func TestBuildUsageTolerance(t *testing.T) {
	for _, tc := range []struct{ actual, status string }{{"4095", "match"}, {"4096", "over"}, {"3705", "match"}, {"3704", "under"}} {
		lines, _, err := BuildUsage([]model.UsageInputRow{{BaseQtyPerPurchaseUnit: "1", PlannedQty: "3900", ActualQty: tc.actual, HasActual: true}}, UsageVarianceTolerance)
		if err != nil || lines[0].Status != tc.status {
			t.Errorf("%#v: %v", tc, err)
		}
	}
}
